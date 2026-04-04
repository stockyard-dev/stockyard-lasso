package server
import ("encoding/json";"net/http";"github.com/stockyard-dev/stockyard-lasso/internal/store")
type Server struct{db *store.DB;mux *http.ServeMux;limits Limits}
func New(db *store.DB,limits Limits)*Server{s:=&Server{db:db,mux:http.NewServeMux(),limits:limits}
s.mux.HandleFunc("GET /api/links",s.listLinks)
s.mux.HandleFunc("POST /api/links",s.createLink)
s.mux.HandleFunc("GET /api/links/{id}",s.getLink)
s.mux.HandleFunc("DELETE /api/links/{id}",s.deleteLink)
s.mux.HandleFunc("GET /api/links/{id}/clicks",s.linkClicks)
s.mux.HandleFunc("GET /api/stats",s.stats)
s.mux.HandleFunc("GET /api/health",s.health)
s.mux.HandleFunc("GET /api/tier",func(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"tier":s.limits.Tier,"upgrade_url":"https://stockyard.dev/lasso/"})})
s.mux.HandleFunc("GET /s/{slug}",s.redirect)
s.mux.HandleFunc("GET /ui",s.dashboard);s.mux.HandleFunc("GET /ui/",s.dashboard);s.mux.HandleFunc("GET /",s.root)
return s}
func(s *Server)ServeHTTP(w http.ResponseWriter,r *http.Request){s.mux.ServeHTTP(w,r)}
func wj(w http.ResponseWriter,c int,v any){w.Header().Set("Content-Type","application/json");w.WriteHeader(c);json.NewEncoder(w).Encode(v)}
func we(w http.ResponseWriter,c int,m string){wj(w,c,map[string]string{"error":m})}
func(s *Server)root(w http.ResponseWriter,r *http.Request){if r.URL.Path!="/"{http.NotFound(w,r);return};http.Redirect(w,r,"/ui",302)}
func(s *Server)listLinks(w http.ResponseWriter,r *http.Request){links:=s.db.List();if links==nil{links=[]store.Link{}};wj(w,200,map[string]any{"links":links})}
func(s *Server)createLink(w http.ResponseWriter,r *http.Request){
if s.limits.MaxItems>0&&s.db.Count()>=s.limits.MaxItems{we(w,402,"Free tier limit reached");return}
var l store.Link;json.NewDecoder(r.Body).Decode(&l);if l.URL==""{we(w,400,"url required");return}
s.db.Create(&l);wj(w,201,s.db.Get(l.ID))}
func(s *Server)getLink(w http.ResponseWriter,r *http.Request){l:=s.db.Get(r.PathValue("id"));if l==nil{we(w,404,"not found");return};wj(w,200,l)}
func(s *Server)deleteLink(w http.ResponseWriter,r *http.Request){s.db.Delete(r.PathValue("id"));wj(w,200,map[string]string{"status":"deleted"})}
func(s *Server)linkClicks(w http.ResponseWriter,r *http.Request){clicks:=s.db.ClicksForLink(r.PathValue("id"));if clicks==nil{clicks=[]store.Click{}};wj(w,200,map[string]any{"clicks":clicks})}
func(s *Server)redirect(w http.ResponseWriter,r *http.Request){slug:=r.PathValue("slug");l:=s.db.GetBySlug(slug);if l==nil{http.NotFound(w,r);return};s.db.RecordClick(slug,r.Referer(),r.UserAgent(),"");http.Redirect(w,r,l.URL,302)}
func(s *Server)stats(w http.ResponseWriter,r *http.Request){wj(w,200,s.db.Stats())}
func(s *Server)health(w http.ResponseWriter,r *http.Request){st:=s.db.Stats();wj(w,200,map[string]any{"service":"lasso","status":"ok","links":st["links"],"clicks":st["clicks"]})}
