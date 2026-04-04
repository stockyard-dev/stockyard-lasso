package server
import "net/http"
func(s *Server)dashboard(w http.ResponseWriter,r *http.Request){w.Header().Set("Content-Type","text/html");w.Write([]byte(dashHTML))}
const dashHTML=`<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>Lasso</title>
<style>:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--mono:'JetBrains Mono',monospace}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--mono);line-height:1.5}
.hdr{padding:1rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}.hdr h1{font-size:.9rem;letter-spacing:2px}
.main{padding:1.5rem;max-width:900px;margin:0 auto}
.stats{display:grid;grid-template-columns:1fr 1fr;gap:.6rem;margin-bottom:1.2rem}
.st{background:var(--bg2);border:1px solid var(--bg3);padding:.7rem;text-align:center}.st-v{font-size:1.3rem}.st-l{font-size:.55rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-top:.1rem}
.link{background:var(--bg2);border:1px solid var(--bg3);padding:.8rem 1rem;margin-bottom:.5rem}
.link-top{display:flex;justify-content:space-between;align-items:center}
.link-slug{font-size:.82rem;color:var(--gold);cursor:pointer}.link-slug:hover{text-decoration:underline}
.link-url{font-size:.7rem;color:var(--cm);margin-top:.2rem;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;max-width:500px}
.link-meta{display:flex;gap:.8rem;font-size:.6rem;color:var(--cm);margin-top:.3rem}
.clicks{color:var(--green);font-weight:bold}
.btn{font-size:.6rem;padding:.25rem .6rem;cursor:pointer;border:1px solid var(--bg3);background:var(--bg);color:var(--cd)}.btn:hover{border-color:var(--leather);color:var(--cream)}
.btn-p{background:var(--rust);border-color:var(--rust);color:var(--bg)}
.new-form{background:var(--bg2);border:1px solid var(--bg3);padding:1rem;margin-bottom:1.5rem;display:flex;gap:.5rem;align-items:flex-end;flex-wrap:wrap}
.new-form input{padding:.4rem .6rem;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.75rem}
.new-form label{font-size:.55rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;display:block;margin-bottom:.15rem}
.empty{text-align:center;padding:3rem;color:var(--cm);font-style:italic;font-size:.75rem}
</style></head><body>
<div class="hdr"><h1>LASSO</h1></div>
<div class="main">
<div class="stats" id="stats"></div>
<div class="new-form"><div style="flex:2"><label>URL to shorten</label><input id="f-url" style="width:100%" placeholder="https://example.com/long/path"></div><div><label>Custom slug</label><input id="f-slug" style="width:100px" placeholder="optional"></div><div><label>Title</label><input id="f-title" style="width:120px" placeholder="optional"></div><button class="btn btn-p" onclick="create()">Shorten</button></div>
<div id="links"></div>
</div>
<script>
const A='/api',BASE=location.origin+'/s/';let links=[];
async function load(){const[l,s]=await Promise.all([fetch(A+'/links').then(r=>r.json()),fetch(A+'/stats').then(r=>r.json())]);links=l.links||[];
document.getElementById('stats').innerHTML='<div class="st"><div class="st-v">'+s.links+'</div><div class="st-l">Links</div></div><div class="st"><div class="st-v">'+s.clicks+'</div><div class="st-l">Total Clicks</div></div>';render();}
function render(){if(!links.length){document.getElementById('links').innerHTML='<div class="empty">No short links yet. Create one above.</div>';return;}
let h='';links.forEach(l=>{h+='<div class="link"><div class="link-top"><div><span class="link-slug" onclick="navigator.clipboard.writeText(\''+BASE+l.slug+'\')">'+BASE+l.slug+'</span>';if(l.title)h+=' — '+esc(l.title);h+='</div><div style="display:flex;gap:.3rem"><span class="clicks">'+l.clicks+' clicks</span><button class="btn" onclick="del(\''+l.id+'\')" style="font-size:.5rem;color:var(--cm)">✕</button></div></div><div class="link-url">→ '+esc(l.url)+'</div><div class="link-meta"><span>Created '+ft(l.created_at)+'</span><span>Slug: '+l.slug+'</span></div></div>';});
document.getElementById('links').innerHTML=h;}
async function create(){const url=document.getElementById('f-url').value;if(!url)return;await fetch(A+'/links',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({url,slug:document.getElementById('f-slug').value,title:document.getElementById('f-title').value})});document.getElementById('f-url').value='';document.getElementById('f-slug').value='';document.getElementById('f-title').value='';load();}
async function del(id){if(confirm('Delete?')){await fetch(A+'/links/'+id,{method:'DELETE'});load();}}
function ft(t){if(!t)return'';const d=new Date(t);return d.toLocaleDateString();}
function esc(s){if(!s)return'';const d=document.createElement('div');d.textContent=s;return d.innerHTML;}
load();
</script></body></html>`
