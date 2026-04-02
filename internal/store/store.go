package store
import ("database/sql";"fmt";"os";"path/filepath";"time";_ "modernc.org/sqlite")
type DB struct{db *sql.DB}
type Link struct{
	ID string `json:"id"`
	Slug string `json:"slug"`
	URL string `json:"url"`
	Title string `json:"title"`
	Clicks int `json:"clicks"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
}
func Open(d string)(*DB,error){if err:=os.MkdirAll(d,0755);err!=nil{return nil,err};db,err:=sql.Open("sqlite",filepath.Join(d,"lasso.db")+"?_journal_mode=WAL&_busy_timeout=5000");if err!=nil{return nil,err}
db.Exec(`CREATE TABLE IF NOT EXISTS links(id TEXT PRIMARY KEY,slug TEXT UNIQUE NOT NULL,url TEXT NOT NULL,title TEXT DEFAULT '',clicks INTEGER DEFAULT 0,expires_at TEXT DEFAULT '',created_at TEXT DEFAULT(datetime('now')))`)
return &DB{db:db},nil}
func(d *DB)Close()error{return d.db.Close()}
func genID()string{return fmt.Sprintf("%d",time.Now().UnixNano())}
func now()string{return time.Now().UTC().Format(time.RFC3339)}
func(d *DB)Create(e *Link)error{e.ID=genID();e.CreatedAt=now();_,err:=d.db.Exec(`INSERT INTO links(id,slug,url,title,clicks,expires_at,created_at)VALUES(?,?,?,?,?,?,?)`,e.ID,e.Slug,e.URL,e.Title,e.Clicks,e.ExpiresAt,e.CreatedAt);return err}
func(d *DB)Get(id string)*Link{var e Link;if d.db.QueryRow(`SELECT id,slug,url,title,clicks,expires_at,created_at FROM links WHERE id=?`,id).Scan(&e.ID,&e.Slug,&e.URL,&e.Title,&e.Clicks,&e.ExpiresAt,&e.CreatedAt)!=nil{return nil};return &e}
func(d *DB)List()[]Link{rows,_:=d.db.Query(`SELECT id,slug,url,title,clicks,expires_at,created_at FROM links ORDER BY created_at DESC`);if rows==nil{return nil};defer rows.Close();var o []Link;for rows.Next(){var e Link;rows.Scan(&e.ID,&e.Slug,&e.URL,&e.Title,&e.Clicks,&e.ExpiresAt,&e.CreatedAt);o=append(o,e)};return o}
func(d *DB)Delete(id string)error{_,err:=d.db.Exec(`DELETE FROM links WHERE id=?`,id);return err}
func(d *DB)Count()int{var n int;d.db.QueryRow(`SELECT COUNT(*) FROM links`).Scan(&n);return n}
