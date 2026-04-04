package store
import ("database/sql";"crypto/rand";"encoding/hex";"fmt";"os";"path/filepath";"time";_ "modernc.org/sqlite")
type DB struct{db *sql.DB}
type Link struct {
	ID string `json:"id"`
	Slug string `json:"slug"`
	URL string `json:"url"`
	Title string `json:"title"`
	Clicks int `json:"clicks"`
	CreatedAt string `json:"created_at"`
}
type Click struct {
	ID string `json:"id"`
	LinkID string `json:"link_id"`
	Referrer string `json:"referrer"`
	UserAgent string `json:"user_agent"`
	Country string `json:"country"`
	CreatedAt string `json:"created_at"`
}
func Open(d string)(*DB,error){if err:=os.MkdirAll(d,0755);err!=nil{return nil,err};db,err:=sql.Open("sqlite",filepath.Join(d,"lasso.db")+"?_journal_mode=WAL&_busy_timeout=5000");if err!=nil{return nil,err}
db.Exec(`CREATE TABLE IF NOT EXISTS links(id TEXT PRIMARY KEY,slug TEXT UNIQUE NOT NULL,url TEXT NOT NULL,title TEXT DEFAULT '',clicks INTEGER DEFAULT 0,created_at TEXT DEFAULT(datetime('now')))`)
db.Exec(`CREATE TABLE IF NOT EXISTS clicks(id TEXT PRIMARY KEY,link_id TEXT NOT NULL,referrer TEXT DEFAULT '',user_agent TEXT DEFAULT '',country TEXT DEFAULT '',created_at TEXT DEFAULT(datetime('now')))`)
db.Exec(`CREATE INDEX IF NOT EXISTS idx_clicks_link ON clicks(link_id)`)
return &DB{db:db},nil}
func(d *DB)Close()error{return d.db.Close()}
func genID()string{return fmt.Sprintf("%d",time.Now().UnixNano())}
func now()string{return time.Now().UTC().Format(time.RFC3339)}
func genSlug()string{b:=make([]byte,4);rand.Read(b);return hex.EncodeToString(b)}
func(d *DB)Create(l *Link)error{l.ID=genID();l.CreatedAt=now();if l.Slug==""{l.Slug=genSlug()};_,err:=d.db.Exec(`INSERT INTO links(id,slug,url,title,created_at)VALUES(?,?,?,?,?)`,l.ID,l.Slug,l.URL,l.Title,l.CreatedAt);return err}
func(d *DB)GetBySlug(slug string)*Link{var l Link;if d.db.QueryRow(`SELECT id,slug,url,title,clicks,created_at FROM links WHERE slug=?`,slug).Scan(&l.ID,&l.Slug,&l.URL,&l.Title,&l.Clicks,&l.CreatedAt)!=nil{return nil};return &l}
func(d *DB)Get(id string)*Link{var l Link;if d.db.QueryRow(`SELECT id,slug,url,title,clicks,created_at FROM links WHERE id=?`,id).Scan(&l.ID,&l.Slug,&l.URL,&l.Title,&l.Clicks,&l.CreatedAt)!=nil{return nil};return &l}
func(d *DB)List()[]Link{rows,_:=d.db.Query(`SELECT id,slug,url,title,clicks,created_at FROM links ORDER BY created_at DESC`);if rows==nil{return nil};defer rows.Close()
var o []Link;for rows.Next(){var l Link;rows.Scan(&l.ID,&l.Slug,&l.URL,&l.Title,&l.Clicks,&l.CreatedAt);o=append(o,l)};return o}
func(d *DB)Delete(id string)error{d.db.Exec(`DELETE FROM clicks WHERE link_id=?`,id);_,err:=d.db.Exec(`DELETE FROM links WHERE id=?`,id);return err}
func(d *DB)RecordClick(slug,referrer,ua,country string){l:=d.GetBySlug(slug);if l==nil{return};d.db.Exec(`INSERT INTO clicks(id,link_id,referrer,user_agent,country,created_at)VALUES(?,?,?,?,?,?)`,genID(),l.ID,referrer,ua,country,now());d.db.Exec(`UPDATE links SET clicks=clicks+1 WHERE slug=?`,slug)}
func(d *DB)ClicksForLink(id string)[]Click{rows,_:=d.db.Query(`SELECT id,link_id,referrer,user_agent,country,created_at FROM clicks WHERE link_id=? ORDER BY created_at DESC LIMIT 50`,id);if rows==nil{return nil};defer rows.Close()
var o []Click;for rows.Next(){var c Click;rows.Scan(&c.ID,&c.LinkID,&c.Referrer,&c.UserAgent,&c.Country,&c.CreatedAt);o=append(o,c)};return o}
func(d *DB)Count()int{var n int;d.db.QueryRow(`SELECT COUNT(*) FROM links`).Scan(&n);return n}
func(d *DB)Stats()map[string]any{var links,clicks int;d.db.QueryRow(`SELECT COUNT(*) FROM links`).Scan(&links);d.db.QueryRow(`SELECT COUNT(*) FROM clicks`).Scan(&clicks);return map[string]any{"links":links,"clicks":clicks}}
