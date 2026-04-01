package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct{ conn *sql.DB }

func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	conn, err := sql.Open("sqlite", filepath.Join(dataDir, "lasso.db"))
	if err != nil {
		return nil, err
	}
	conn.Exec("PRAGMA journal_mode=WAL")
	conn.Exec("PRAGMA busy_timeout=5000")
	conn.SetMaxOpenConns(4)
	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DB) Close() error { return db.conn.Close() }

func (db *DB) migrate() error {
	_, err := db.conn.Exec(`
CREATE TABLE IF NOT EXISTS links (
    id TEXT PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    target_url TEXT NOT NULL,
    title TEXT DEFAULT '',
    expires_at TEXT DEFAULT '',
    password TEXT DEFAULT '',
    enabled INTEGER DEFAULT 1,
    clicks INTEGER DEFAULT 0,
    created_at TEXT DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_links_slug ON links(slug);

CREATE TABLE IF NOT EXISTS clicks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    link_id TEXT NOT NULL,
    referrer TEXT DEFAULT '',
    user_agent TEXT DEFAULT '',
    country TEXT DEFAULT '',
    device TEXT DEFAULT '',
    browser TEXT DEFAULT '',
    source_ip TEXT DEFAULT '',
    clicked_at TEXT DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_clicks_link ON clicks(link_id);
CREATE INDEX IF NOT EXISTS idx_clicks_time ON clicks(clicked_at);
`)
	return err
}

// --- Links ---

type Link struct {
	ID        string `json:"id"`
	Slug      string `json:"slug"`
	TargetURL string `json:"target_url"`
	Title     string `json:"title"`
	ExpiresAt string `json:"expires_at,omitempty"`
	Password  string `json:"-"`
	HasPwd    bool   `json:"password_protected"`
	Enabled   bool   `json:"enabled"`
	Clicks    int    `json:"clicks"`
	CreatedAt string `json:"created_at"`
}

func (db *DB) CreateLink(slug, targetURL, title, expiresAt, password string) (*Link, error) {
	id := "lnk_" + genID(6)
	now := time.Now().UTC().Format(time.RFC3339)
	if slug == "" {
		slug = shortCode(6)
	}
	_, err := db.conn.Exec("INSERT INTO links (id,slug,target_url,title,expires_at,password,created_at) VALUES (?,?,?,?,?,?,?)",
		id, slug, targetURL, title, expiresAt, password, now)
	if err != nil {
		return nil, err
	}
	return &Link{ID: id, Slug: slug, TargetURL: targetURL, Title: title, ExpiresAt: expiresAt,
		HasPwd: password != "", Enabled: true, CreatedAt: now}, nil
}

func (db *DB) ListLinks(limit int) ([]Link, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := db.conn.Query("SELECT id,slug,target_url,title,expires_at,password,enabled,clicks,created_at FROM links ORDER BY created_at DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLinks(rows)
}

func (db *DB) GetLinkBySlug(slug string) (*Link, error) {
	var l Link
	var en int
	err := db.conn.QueryRow("SELECT id,slug,target_url,title,expires_at,password,enabled,clicks,created_at FROM links WHERE slug=?", slug).
		Scan(&l.ID, &l.Slug, &l.TargetURL, &l.Title, &l.ExpiresAt, &l.Password, &en, &l.Clicks, &l.CreatedAt)
	if err != nil {
		return nil, err
	}
	l.Enabled = en == 1
	l.HasPwd = l.Password != ""
	return &l, nil
}

func (db *DB) GetLink(id string) (*Link, error) {
	var l Link
	var en int
	err := db.conn.QueryRow("SELECT id,slug,target_url,title,expires_at,password,enabled,clicks,created_at FROM links WHERE id=?", id).
		Scan(&l.ID, &l.Slug, &l.TargetURL, &l.Title, &l.ExpiresAt, &l.Password, &en, &l.Clicks, &l.CreatedAt)
	if err != nil {
		return nil, err
	}
	l.Enabled = en == 1
	l.HasPwd = l.Password != ""
	return &l, nil
}

func (db *DB) UpdateLink(id string, targetURL, title *string, enabled *bool) (*Link, error) {
	if targetURL != nil {
		db.conn.Exec("UPDATE links SET target_url=? WHERE id=?", *targetURL, id)
	}
	if title != nil {
		db.conn.Exec("UPDATE links SET title=? WHERE id=?", *title, id)
	}
	if enabled != nil {
		en := 0
		if *enabled {
			en = 1
		}
		db.conn.Exec("UPDATE links SET enabled=? WHERE id=?", en, id)
	}
	return db.GetLink(id)
}

func (db *DB) DeleteLink(id string) error {
	db.conn.Exec("DELETE FROM clicks WHERE link_id=?", id)
	_, err := db.conn.Exec("DELETE FROM links WHERE id=?", id)
	return err
}

func (db *DB) IncrementClicks(id string) {
	db.conn.Exec("UPDATE links SET clicks=clicks+1 WHERE id=?", id)
}

func (db *DB) TotalLinks() int {
	var count int
	db.conn.QueryRow("SELECT COUNT(*) FROM links").Scan(&count)
	return count
}

// --- Clicks ---

type Click struct {
	ID        int    `json:"id"`
	LinkID    string `json:"link_id"`
	Referrer  string `json:"referrer"`
	UserAgent string `json:"user_agent"`
	Device    string `json:"device"`
	Browser   string `json:"browser"`
	SourceIP  string `json:"source_ip"`
	ClickedAt string `json:"clicked_at"`
}

func (db *DB) RecordClick(linkID, referrer, ua, device, browser, ip string) {
	db.conn.Exec("INSERT INTO clicks (link_id,referrer,user_agent,device,browser,source_ip) VALUES (?,?,?,?,?,?)",
		linkID, referrer, ua, device, browser, ip)
}

func (db *DB) ListClicks(linkID string, limit int) ([]Click, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := db.conn.Query("SELECT id,link_id,referrer,user_agent,device,browser,source_ip,clicked_at FROM clicks WHERE link_id=? ORDER BY clicked_at DESC LIMIT ?", linkID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Click
	for rows.Next() {
		var c Click
		rows.Scan(&c.ID, &c.LinkID, &c.Referrer, &c.UserAgent, &c.Device, &c.Browser, &c.SourceIP, &c.ClickedAt)
		out = append(out, c)
	}
	return out, rows.Err()
}

type TopEntry struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func (db *DB) ClicksByDay(linkID string, days int) []TopEntry {
	cutoff := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	rows, _ := db.conn.Query("SELECT date(clicked_at), COUNT(*) FROM clicks WHERE link_id=? AND date(clicked_at)>=? GROUP BY date(clicked_at) ORDER BY date(clicked_at)", linkID, cutoff)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var out []TopEntry
	for rows.Next() {
		var e TopEntry
		rows.Scan(&e.Name, &e.Count)
		out = append(out, e)
	}
	return out
}

func (db *DB) TopReferrers(linkID string, limit int) []TopEntry {
	rows, _ := db.conn.Query("SELECT referrer, COUNT(*) as c FROM clicks WHERE link_id=? AND referrer!='' GROUP BY referrer ORDER BY c DESC LIMIT ?", linkID, limit)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var out []TopEntry
	for rows.Next() {
		var e TopEntry
		rows.Scan(&e.Name, &e.Count)
		out = append(out, e)
	}
	return out
}

// --- Stats ---

func (db *DB) Stats() map[string]any {
	var links, clicks int
	db.conn.QueryRow("SELECT COUNT(*) FROM links").Scan(&links)
	db.conn.QueryRow("SELECT COUNT(*) FROM clicks").Scan(&clicks)
	return map[string]any{"links": links, "clicks": clicks}
}

func (db *DB) Cleanup(days int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -days).Format("2006-01-02 15:04:05")
	res, err := db.conn.Exec("DELETE FROM clicks WHERE clicked_at < ?", cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// --- Helpers ---

func scanLinks(rows *sql.Rows) ([]Link, error) {
	var out []Link
	for rows.Next() {
		var l Link
		var en int
		rows.Scan(&l.ID, &l.Slug, &l.TargetURL, &l.Title, &l.ExpiresAt, &l.Password, &en, &l.Clicks, &l.CreatedAt)
		l.Enabled = en == 1
		l.HasPwd = l.Password != ""
		l.Password = "" // never expose
		out = append(out, l)
	}
	return out, rows.Err()
}

const charset = "abcdefghijkmnpqrstuvwxyz23456789" // no confusing chars

func shortCode(n int) string {
	b := make([]byte, n)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[num.Int64()]
	}
	return string(b)
}

func genID(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
