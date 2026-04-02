package store
import ("database/sql";"fmt";"os";"path/filepath";"time";_ "modernc.org/sqlite")
type DB struct{db *sql.DB}
type Capture struct {
	ID string `json:"id"`
	Name string `json:"name"`
	URL string `json:"url"`
	Method string `json:"method"`
	Headers string `json:"headers"`
	Body string `json:"body"`
	ResponseCode int `json:"response_code"`
	Status string `json:"status"`
	CreatedAt string `json:"created_at"`
}
func Open(d string)(*DB,error){if err:=os.MkdirAll(d,0755);err!=nil{return nil,err};db,err:=sql.Open("sqlite",filepath.Join(d,"lasso.db")+"?_journal_mode=WAL&_busy_timeout=5000");if err!=nil{return nil,err}
db.Exec(`CREATE TABLE IF NOT EXISTS captures(id TEXT PRIMARY KEY,name TEXT NOT NULL,url TEXT DEFAULT '',method TEXT DEFAULT 'GET',headers TEXT DEFAULT '{}',body TEXT DEFAULT '',response_code INTEGER DEFAULT 0,status TEXT DEFAULT 'captured',created_at TEXT DEFAULT(datetime('now')))`)
return &DB{db:db},nil}
func(d *DB)Close()error{return d.db.Close()}
func genID()string{return fmt.Sprintf("%d",time.Now().UnixNano())}
func now()string{return time.Now().UTC().Format(time.RFC3339)}
func(d *DB)Create(e *Capture)error{e.ID=genID();e.CreatedAt=now();_,err:=d.db.Exec(`INSERT INTO captures(id,name,url,method,headers,body,response_code,status,created_at)VALUES(?,?,?,?,?,?,?,?,?)`,e.ID,e.Name,e.URL,e.Method,e.Headers,e.Body,e.ResponseCode,e.Status,e.CreatedAt);return err}
func(d *DB)Get(id string)*Capture{var e Capture;if d.db.QueryRow(`SELECT id,name,url,method,headers,body,response_code,status,created_at FROM captures WHERE id=?`,id).Scan(&e.ID,&e.Name,&e.URL,&e.Method,&e.Headers,&e.Body,&e.ResponseCode,&e.Status,&e.CreatedAt)!=nil{return nil};return &e}
func(d *DB)List()[]Capture{rows,_:=d.db.Query(`SELECT id,name,url,method,headers,body,response_code,status,created_at FROM captures ORDER BY created_at DESC`);if rows==nil{return nil};defer rows.Close();var o []Capture;for rows.Next(){var e Capture;rows.Scan(&e.ID,&e.Name,&e.URL,&e.Method,&e.Headers,&e.Body,&e.ResponseCode,&e.Status,&e.CreatedAt);o=append(o,e)};return o}
func(d *DB)Update(e *Capture)error{_,err:=d.db.Exec(`UPDATE captures SET name=?,url=?,method=?,headers=?,body=?,response_code=?,status=? WHERE id=?`,e.Name,e.URL,e.Method,e.Headers,e.Body,e.ResponseCode,e.Status,e.ID);return err}
func(d *DB)Delete(id string)error{_,err:=d.db.Exec(`DELETE FROM captures WHERE id=?`,id);return err}
func(d *DB)Count()int{var n int;d.db.QueryRow(`SELECT COUNT(*) FROM captures`).Scan(&n);return n}

func(d *DB)Search(q string, filters map[string]string)[]Capture{
    where:="1=1"
    args:=[]any{}
    if q!=""{
        where+=" AND (name LIKE ? OR body LIKE ?)"
        args=append(args,"%"+q+"%");args=append(args,"%"+q+"%");
    }
    if v,ok:=filters["status"];ok&&v!=""{where+=" AND status=?";args=append(args,v)}
    rows,_:=d.db.Query(`SELECT id,name,url,method,headers,body,response_code,status,created_at FROM captures WHERE `+where+` ORDER BY created_at DESC`,args...)
    if rows==nil{return nil};defer rows.Close()
    var o []Capture;for rows.Next(){var e Capture;rows.Scan(&e.ID,&e.Name,&e.URL,&e.Method,&e.Headers,&e.Body,&e.ResponseCode,&e.Status,&e.CreatedAt);o=append(o,e)};return o
}

func(d *DB)Stats()map[string]any{
    m:=map[string]any{"total":d.Count()}
    rows,_:=d.db.Query(`SELECT status,COUNT(*) FROM captures GROUP BY status`)
    if rows!=nil{defer rows.Close();by:=map[string]int{};for rows.Next(){var s string;var c int;rows.Scan(&s,&c);by[s]=c};m["by_status"]=by}
    return m
}
