package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stockyard-dev/stockyard-lasso/internal/store"
)

type Server struct {
	db     *store.DB
	mux    *http.ServeMux
	port   int
	limits Limits
}

func New(db *store.DB, port int, limits Limits) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), port: port, limits: limits}
	s.routes()
	return s
}

func (s *Server) routes() {
	// Links CRUD
	s.mux.HandleFunc("POST /api/links", s.handleCreateLink)
	s.mux.HandleFunc("GET /api/links", s.handleListLinks)
	s.mux.HandleFunc("GET /api/links/{id}", s.handleGetLink)
	s.mux.HandleFunc("PUT /api/links/{id}", s.handleUpdateLink)
	s.mux.HandleFunc("DELETE /api/links/{id}", s.handleDeleteLink)

	// Click analytics
	s.mux.HandleFunc("GET /api/links/{id}/clicks", s.handleListClicks)
	s.mux.HandleFunc("GET /api/links/{id}/stats", s.handleLinkStats)

	// Status
	s.mux.HandleFunc("GET /api/status", s.handleStatus)
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /ui", s.handleUI)
	s.mux.HandleFunc("GET /api/version", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"product": "stockyard-lasso", "version": "0.1.0"})
	})

	// Redirect handler — catch-all for short URLs
	s.mux.HandleFunc("GET /{slug}", s.handleRedirect)
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("[lasso] listening on %s", addr)
	return http.ListenAndServe(addr, s.mux)
}

// --- Redirect (the hot path) ---

func (s *Server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	// Skip known paths
	if slug == "api" || slug == "ui" || slug == "health" || slug == "" {
		http.NotFound(w, r)
		return
	}

	link, err := s.db.GetLinkBySlug(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if !link.Enabled {
		http.Error(w, "This link has been disabled", 410)
		return
	}

	// Check expiry
	if link.ExpiresAt != "" {
		exp, err := time.Parse(time.RFC3339, link.ExpiresAt)
		if err == nil && time.Now().After(exp) {
			http.Error(w, "This link has expired", 410)
			return
		}
	}

	// Track click
	ip := r.RemoteAddr
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		ip = strings.Split(fwd, ",")[0]
	}
	ua := r.UserAgent()
	device, browser := parseUA(ua)
	referrer := r.Referer()

	s.db.IncrementClicks(link.ID)
	go s.db.RecordClick(link.ID, referrer, ua, device, browser, ip)

	http.Redirect(w, r, link.TargetURL, http.StatusFound)
}

// --- Link CRUD ---

func (s *Server) handleCreateLink(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL       string `json:"url"`
		Slug      string `json:"slug"`
		Title     string `json:"title"`
		ExpiresAt string `json:"expires_at"`
		Password  string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	if req.URL == "" {
		writeJSON(w, 400, map[string]string{"error": "url is required"})
		return
	}

	if s.limits.MaxLinks > 0 {
		total := s.db.TotalLinks()
		if LimitReached(s.limits.MaxLinks, total) {
			writeJSON(w, 402, map[string]string{
				"error":   fmt.Sprintf("free tier limit: %d links max — upgrade to Pro", s.limits.MaxLinks),
				"upgrade": "https://stockyard.dev/lasso/",
			})
			return
		}
	}

	if req.Password != "" && !s.limits.PasswordLinks {
		writeJSON(w, 402, map[string]string{
			"error":   "password-protected links require Pro — upgrade at https://stockyard.dev/lasso/",
			"upgrade": "https://stockyard.dev/lasso/",
		})
		return
	}

	link, err := s.db.CreateLink(req.Slug, req.URL, req.Title, req.ExpiresAt, req.Password)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	shortURL := fmt.Sprintf("http://localhost:%d/%s", s.port, link.Slug)
	writeJSON(w, 201, map[string]any{"link": link, "short_url": shortURL})
}

func (s *Server) handleListLinks(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil {
			limit = n
		}
	}
	links, err := s.db.ListLinks(limit)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if links == nil {
		links = []store.Link{}
	}
	writeJSON(w, 200, map[string]any{"links": links, "count": len(links)})
}

func (s *Server) handleGetLink(w http.ResponseWriter, r *http.Request) {
	link, err := s.db.GetLink(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "link not found"})
		return
	}
	link.Password = ""
	shortURL := fmt.Sprintf("http://localhost:%d/%s", s.port, link.Slug)
	writeJSON(w, 200, map[string]any{"link": link, "short_url": shortURL})
}

func (s *Server) handleUpdateLink(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := s.db.GetLink(id); err != nil {
		writeJSON(w, 404, map[string]string{"error": "link not found"})
		return
	}
	var req struct {
		URL     *string `json:"url"`
		Title   *string `json:"title"`
		Enabled *bool   `json:"enabled"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	link, err := s.db.UpdateLink(id, req.URL, req.Title, req.Enabled)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	link.Password = ""
	writeJSON(w, 200, map[string]any{"link": link})
}

func (s *Server) handleDeleteLink(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := s.db.GetLink(id); err != nil {
		writeJSON(w, 404, map[string]string{"error": "link not found"})
		return
	}
	s.db.DeleteLink(id)
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

// --- Click analytics ---

func (s *Server) handleListClicks(w http.ResponseWriter, r *http.Request) {
	clicks, err := s.db.ListClicks(r.PathValue("id"), 100)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if clicks == nil {
		clicks = []store.Click{}
	}
	writeJSON(w, 200, map[string]any{"clicks": clicks, "count": len(clicks)})
}

func (s *Server) handleLinkStats(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	link, err := s.db.GetLink(id)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "link not found"})
		return
	}
	daily := s.db.ClicksByDay(id, 30)
	refs := s.db.TopReferrers(id, 10)
	if daily == nil {
		daily = []store.TopEntry{}
	}
	if refs == nil {
		refs = []store.TopEntry{}
	}
	writeJSON(w, 200, map[string]any{
		"slug": link.Slug, "total_clicks": link.Clicks,
		"daily": daily, "referrers": refs,
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, s.db.Stats())
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

// --- UA parsing ---

func parseUA(ua string) (device, browser string) {
	ua = strings.ToLower(ua)
	device = "desktop"
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") {
		device = "mobile"
	} else if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		device = "tablet"
	}
	switch {
	case strings.Contains(ua, "firefox"):
		browser = "Firefox"
	case strings.Contains(ua, "edg/"):
		browser = "Edge"
	case strings.Contains(ua, "chrome"):
		browser = "Chrome"
	case strings.Contains(ua, "safari"):
		browser = "Safari"
	default:
		browser = "Other"
	}
	return
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
