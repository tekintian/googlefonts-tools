package controller

import "fmt"

var AppVer = "dev"

const githubRepo = "https://github.com/tekintian/googlefonts-tools"

const footerCSS = `.footer{margin-top:20px;padding-top:12px;border-top:1px solid #eee;font-size:12px;color:#aaa;text-align:center}.footer a{color:#667eea;text-decoration:none}.footer a:hover{text-decoration:underline}`

func footerHTML() string {
	return fmt.Sprintf(`<div class="footer"><a href="%s" target="_blank">GoogleFonts Tools %s</a> · Powered by <a href="https://ai.tekin.cn/" target="_blank">Tekin</a></div>`, githubRepo, AppVer)
}

func indexHTML() string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Google Fonts Download Tools - 自托管 Web 字体下载与 CSS 生成</title>
<meta name="description" content="在线下载 Google Fonts 字体文件并生成自托管 CSS，摆脱 CDN 依赖。支持异步下载、ZIP 打包、Nginx 静态直出、协议相对 URL 引用。">
<meta name="keywords" content="Google Fonts,字体下载,自托管字体,self-host fonts,web fonts,font download,Google Fonts CDN,字体打包,woff2 download,CSS 生成">
<meta name="robots" content="index, follow">
<link rel="canonical" href="/">
<meta property="og:type" content="website">
<meta property="og:title" content="Google Fonts Download Tools">
<meta property="og:description" content="在线下载 Google Fonts 字体文件并生成自托管 CSS，摆脱 CDN 依赖">
<meta property="og:url" content="/">
<meta property="og:image" content="/assets/img/logo.png">
<link rel="icon" type="image/x-icon" href="/assets/img/favicon.ico">
<link rel="apple-touch-icon" href="/assets/img/logo.png">
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#0f0f1a;color:#e0e0e0;min-height:100vh}
.hero{background:linear-gradient(135deg,#1a1a2e 0%,#16213e 50%,#0f3460 100%);padding:60px 20px 40px;text-align:center}
.hero-icon{width:64px;height:64px;margin-bottom:16px;display:inline-block;animation:float 3s ease-in-out infinite}
@keyframes float{0%,100%{transform:translateY(0)}50%{transform:translateY(-8px)}}
.hero h1{font-size:32px;font-weight:800;background:linear-gradient(135deg,#89b4fa,#cba6f7,#f5c2e7);-webkit-background-clip:text;-webkit-text-fill-color:transparent;background-clip:text;margin-bottom:10px;letter-spacing:-0.5px}
.hero p{color:#a0a0b8;font-size:15px;max-width:480px;margin:0 auto;line-height:1.6}
.main{max-width:720px;margin:0 auto;padding:0 20px}
.input-card{background:#1e1e2e;border-radius:16px;padding:28px;margin-top:-24px;box-shadow:0 8px 32px rgba(0,0,0,.4);position:relative;z-index:1}
.input-card label{display:block;color:#cdd6f4;font-weight:600;font-size:13px;margin-bottom:10px;letter-spacing:0.5px}
.input-wrap{position:relative}
.input-wrap input{width:100%;padding:14px 16px;background:#11111b;border:2px solid #313244;border-radius:10px;color:#cdd6f4;font-size:14px;transition:border-color .3s,box-shadow .3s;outline:none;font-family:'Fira Code',Consolas,monospace}
.input-wrap input:focus{border-color:#89b4fa;box-shadow:0 0 0 3px rgba(137,180,250,.15)}
.input-wrap input::placeholder{color:#585b70}
.btn{display:flex;align-items:center;justify-content:center;gap:8px;width:100%;padding:14px;margin-top:16px;background:linear-gradient(135deg,#89b4fa,#cba6f7);color:#1e1e2e;border:none;border-radius:10px;font-size:15px;font-weight:700;cursor:pointer;transition:transform .2s,box-shadow .2s;letter-spacing:0.3px}
.btn:hover{transform:translateY(-2px);box-shadow:0 6px 20px rgba(137,180,250,.3)}
.btn:disabled{opacity:.5;cursor:not-allowed;transform:none;box-shadow:none}
.examples{margin-top:20px;padding-top:16px;border-top:1px solid #313244}
.examples-title{color:#6c7086;font-size:12px;margin-bottom:10px;letter-spacing:0.5px}
.example-list{display:flex;flex-direction:column;gap:6px}
.example-item{padding:8px 12px;background:#11111b;border-radius:8px;cursor:pointer;transition:background .2s,border-color .2s;border:1px solid transparent;font-size:12px;color:#a6adc8;font-family:'Fira Code',Consolas,monospace;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.example-item:hover{background:#181825;border-color:#45475a;color:#cdd6f4}
.features{display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:16px;margin-top:32px;padding-bottom:8px}
.feature{background:#1e1e2e;border-radius:12px;padding:20px;text-align:center;transition:transform .2s}
.feature:hover{transform:translateY(-3px)}
.feature-icon{font-size:28px;margin-bottom:10px}
.feature h3{color:#cdd6f4;font-size:14px;font-weight:600;margin-bottom:6px}
.feature p{color:#6c7086;font-size:12px;line-height:1.5}
.result{margin-top:20px;padding:20px;background:#11111b;border:1px solid #313244;border-radius:12px;display:none}
.result h3{color:#a6e3a1;font-size:15px;margin-bottom:14px;font-weight:600}
.task-info{display:flex;align-items:baseline;padding:6px 0;font-size:13px;border-bottom:1px solid #1e1e2e}
.task-info:last-child{border-bottom:none}
.task-label{color:#6c7086;min-width:72px;flex-shrink:0}
.task-value{color:#cdd6f4;flex:1;word-break:break-all}
.task-value a{color:#89b4fa;text-decoration:none}
.task-value a:hover{text-decoration:underline}
.nav-links{display:flex;justify-content:center;gap:24px;margin-top:28px;padding-bottom:32px}
.nav-links a{color:#6c7086;font-size:13px;text-decoration:none;transition:color .2s}
.nav-links a:hover{color:#89b4fa}
.footer{margin-top:0;padding:16px;text-align:center;font-size:12px;color:#45475a;border-top:1px solid #1e1e2e}
.footer a{color:#585b70;text-decoration:none}
.footer a:hover{color:#89b4fa;text-decoration:underline}
</style>
</head>
<body>
<div class="hero">
<div class="hero-icon"><img src="/assets/img/logo.png" alt="Logo" width="64" height="64"></div>
<h1>Google Fonts Download</h1>
<p>下载字体文件并生成自托管 CSS，一行代码替代 Google Fonts CDN</p>
</div>
<div class="main">
<div class="input-card">
<label>GOOGLE FONTS URL</label>
<div class="input-wrap">
<input type="text" id="url" placeholder="粘贴 Google Fonts URL..." required>
</div>
<button type="button" class="btn" id="btn" onclick="submitTask()">⬇ 下载字体</button>
<div class="examples">
<div class="examples-title">快速示例 — 点击自动填入</div>
<div class="example-list">
<div class="example-item" onclick="fillUrl(this.textContent)">https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap</div>
<div class="example-item" onclick="fillUrl(this.textContent)">https://fonts.googleapis.com/css2?family=Roboto:wght@400;700&display=swap</div>
<div class="example-item" onclick="fillUrl(this.textContent)">https://fonts.googleapis.com/css2?family=Poppins:wght@400;600&display=swap</div>
</div>
</div>
</div>
<div class="result" id="result">
<h3>✅ 任务已提交</h3>
<div class="task-info"><span class="task-label">字体</span><span class="task-value" id="fontName">-</span></div>
<div class="task-info"><span class="task-label">状态</span><span class="task-value" id="status">-</span></div>
<div class="task-info"><span class="task-label">永久链接</span><span class="task-value"><a id="permalink" href="#">-</a></span></div>
<div class="task-info" id="downloadRow" style="display:none"><span class="task-label">下载</span><span class="task-value"><a id="downloadLink" href="#">点击下载 ZIP</a></span></div>
</div>
<div class="features">
<div class="feature"><div class="feature-icon">📦</div><h3>ZIP 打包下载</h3><p>字体文件 + 自托管 CSS 一键打包</p></div>
<div class="feature"><div class="feature-icon">🏠</div><h3>自托管 CSS</h3><p>本地相对路径，摆脱 CDN 依赖</p></div>
<div class="feature"><div class="feature-icon">🔒</div><h3>协议自适应</h3><p>协议相对 URL，HTTP/HTTPS 通用</p></div>
</div>
<div class="nav-links">
<a href="/recent">📋 最近下载</a>
<a href="https://github.com/tekintian/googlefonts-tools/issues" target="_blank">📖 Issues</a>
</div>
</div>
` + footerHTML() + `
<script>
function fillUrl(url){document.getElementById('url').value=url;document.getElementById('url').focus()}
function submitTask(){
var btn=document.getElementById('btn');
var url=document.getElementById('url').value.trim();
if(!url){document.getElementById('url').focus();return}
btn.disabled=true;btn.textContent='提交中...';
fetch('/api/v1/tasks',{method:'POST',headers:{'Content-Type':'application/x-www-form-urlencoded'},body:'url='+encodeURIComponent(url)})
.then(r=>r.json())
.then(data=>{
if(data.code!==200){alert('Error: '+data.msg);btn.disabled=false;btn.innerHTML='⬇ 下载字体';return}
var d=data.data;
document.getElementById('result').style.display='block';
document.getElementById('fontName').textContent=d.font_name||'-';
document.getElementById('status').textContent=d.status;
document.getElementById('permalink').textContent=d.permalink;
document.getElementById('permalink').href=d.permalink;
if(d.status==='success'){
document.getElementById('downloadRow').style.display='flex';
document.getElementById('downloadLink').href=d.download_url;
document.getElementById('status').textContent='✅ 已完成(缓存)';
}else{
document.getElementById('downloadRow').style.display='none';
window.location.href=d.permalink;
}
btn.disabled=false;btn.innerHTML='⬇ 下载字体';
})
.catch(err=>{alert('请求失败');btn.disabled=false;btn.innerHTML='⬇ 下载字体'});
}
document.getElementById('url').addEventListener('keydown',function(e){if(e.key==='Enter'){e.preventDefault();submitTask()}});
</script>
</body>
</html>`
}

const progressHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>下载中 - %s - Google Fonts Tools</title>
<meta name="description" content="%s 字体正在下载中，请稍候...">
<meta name="robots" content="noindex, nofollow">
<link rel="icon" type="image/x-icon" href="/assets/img/favicon.ico">
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:linear-gradient(135deg,#667eea 0%%,#764ba2 100%%);min-height:100vh;display:flex;align-items:center;justify-content:center}
.container{background:#fff;border-radius:16px;padding:40px;max-width:560px;width:90%%;box-shadow:0 20px 60px rgba(0,0,0,.3);text-align:center}
h1{color:#333;font-size:22px;margin-bottom:4px}
.sign{color:#aaa;font-size:12px;margin-bottom:24px;font-family:monospace}
.progress-bar{width:100%%;height:24px;background:#e0e0e0;border-radius:12px;overflow:hidden;margin-bottom:12px}
.progress-fill{height:100%%;background:linear-gradient(90deg,#667eea,#764ba2);border-radius:12px;transition:width .3s;width:%d%%}
.progress-text{color:#555;font-size:14px;margin-bottom:8px}
.file-info{color:#888;font-size:13px}
.status-icon{font-size:48px;margin-bottom:16px}
` + footerCSS + `
</style>
</head>
<body>
<div class="container">
<div class="status-icon" id="icon">⏳</div>
<h1 id="title">下载中: %s</h1>
<p class="sign">%s</p>
<div class="progress-bar"><div class="progress-fill" id="fill"></div></div>
<p class="progress-text" id="progressText">进度: %d%%</p>
<p class="file-info" id="fileInfo">(%d / %d 字体文件)</p>
` + "%s" + `
</div>
<script>
var sign='%s';
var es=new EventSource('/d/'+sign+'/progress');
es.onmessage=function(e){
var d=JSON.parse(e.data);
document.getElementById('fill').style.width=d.progress+'%%';
document.getElementById('progressText').textContent='进度: '+d.progress+'%%';
document.getElementById('fileInfo').textContent='('+d.done_files+' / '+d.total_files+' 字体文件)';
if(d.status==='success'){
document.getElementById('icon').textContent='✅';
document.getElementById('title').textContent='下载完成!';
es.close();
setTimeout(function(){window.location.href='/d/'+sign;},1000);
}else if(d.status==='failed'){
document.getElementById('icon').textContent='❌';
document.getElementById('title').textContent='下载失败';
document.getElementById('progressText').textContent='错误: '+d.error_msg;
es.close();
}
};
es.onerror=function(){es.close()};
</script>
</body>
</html>`

const resultHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>%s - 自托管字体下载完成 | Google Fonts Tools</title>
<meta name="description" content="下载 %s 字体文件并生成自托管 CSS，可直接通过 Nginx 静态服务访问，无需依赖 Google Fonts CDN">
<meta name="keywords" content="%s,Google Fonts,自托管字体,self-host fonts,font download,woff2">
<meta name="robots" content="index, follow">
<link rel="canonical" href="/d/%s">
<meta property="og:type" content="article">
<meta property="og:title" content="%s - 自托管字体下载完成">
<meta property="og:description" content="下载 %s 字体文件并生成自托管 CSS，无需依赖 Google Fonts CDN">
<meta property="og:image" content="/assets/img/logo.png">
<link rel="icon" type="image/x-icon" href="/assets/img/favicon.ico">
<link rel="apple-touch-icon" href="/assets/img/logo.png">
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:linear-gradient(135deg,#667eea 0%%,#764ba2 100%%);min-height:100vh;display:flex;align-items:center;justify-content:center;padding:20px}
.container{background:#fff;border-radius:16px;padding:36px;max-width:620px;width:100%%;box-shadow:0 20px 60px rgba(0,0,0,.3)}
.header{text-align:center;margin-bottom:24px}
.icon{font-size:48px;margin-bottom:8px}
h1{color:#1a1a2e;font-size:20px;font-weight:700}
.info{background:#f8f9fb;border-radius:10px;padding:14px 16px;margin-bottom:20px}
.info-row{display:flex;align-items:baseline;padding:5px 0;font-size:13px;border-bottom:1px solid #eef0f4}
.info-row:last-child{border-bottom:none}
.info-label{color:#888;min-width:72px;flex-shrink:0}
.info-value{color:#333;word-break:break-all;flex:1}
.info-value a{color:#667eea;text-decoration:none}
.info-value a:hover{text-decoration:underline}
.actions{display:flex;gap:10px;margin-bottom:20px}
.btn{flex:1;padding:12px 0;background:linear-gradient(135deg,#667eea,#764ba2);color:#fff;border:none;border-radius:8px;font-size:14px;font-weight:600;cursor:pointer;text-decoration:none;text-align:center;transition:transform .2s,box-shadow .2s}
.btn:hover{transform:translateY(-2px);box-shadow:0 4px 12px rgba(102,126,234,.4)}
.btn-outline{flex:1;padding:12px 0;background:#fff;color:#667eea;border:2px solid #667eea;border-radius:8px;font-size:14px;font-weight:600;cursor:pointer;text-decoration:none;text-align:center;transition:transform .2s}
.btn-outline:hover{transform:translateY(-2px);background:#f0f0ff}
.section{margin-bottom:16px}
.section-title{font-size:13px;font-weight:700;color:#1a1a2e;margin-bottom:8px;display:flex;align-items:center;gap:6px}
.section-desc{font-size:12px;color:#888;margin-bottom:8px}
.code-box{position:relative;background:#1e1e2e;border-radius:8px;padding:14px 16px}
.code-box code{color:#cdd6f4;font-size:12px;font-family:'Fira Code',Consolas,monospace;white-space:pre-wrap;word-break:break-all}
.code-box .tag{color:#89b4fa}
.code-box .attr{color:#f9e2af}
.code-box .val{color:#a6e3a1}
.code-copy{position:absolute;top:8px;right:8px;padding:4px 10px;background:#45475a;color:#cdd6f4;border:none;border-radius:4px;font-size:11px;cursor:pointer;transition:background .2s}
.code-copy:hover{background:#667eea;color:#fff}
.code-copy.copied{background:#a6e3a1;color:#1e1e2e}
.permalink{padding:10px 12px;background:#f8f9fb;border-radius:8px;font-size:12px;color:#888}
.permalink a{color:#667eea;text-decoration:none}
.permalink small{color:#aaa}
` + footerCSS + `
</style>
</head>
<body>
<div class="container">
<div class="header">
<div class="icon">✅</div>
<h1>%s</h1>
</div>
<div class="info">
<div class="info-row"><span class="info-label">签名</span><span class="info-value">%s</span></div>
<div class="info-row"><span class="info-label">Google 源</span><span class="info-value"><a href="%s" target="_blank">%s</a></span></div>
<div class="info-row"><span class="info-label">大小</span><span class="info-value">%s</span></div>
<div class="info-row"><span class="info-label">耗时</span><span class="info-value">%s</span></div>
<div class="info-row"><span class="info-label">下载次数</span><span class="info-value">%d</span></div>
<div class="info-row"><span class="info-label">创建时间</span><span class="info-value">%s</span></div>
</div>
<div class="actions">
<a href="%s" class="btn">⬇ 下载 ZIP</a>
<a href="/" class="btn-outline">🏠 首页</a>
</div>
<div class="section">
<div class="section-title">📋 自托管引用</div>
<div class="section-desc">将此标签加入 HTML &lt;head&gt; 即可自托管字体，无需依赖 Google Fonts CDN</div>
<div class="code-box">
<button class="code-copy" onclick="copyCode(this,'linkCode')">复制</button>
<code id="linkCode">&lt;link rel="<span class="attr">stylesheet</span>" href="<span class="val">%s</span>"&gt;</code>
</div>
</div>
<div class="section">
<div class="section-title">🔗 CSS 文件地址</div>
<div class="code-box">
<button class="code-copy" onclick="copyCode(this,'cssUrl')">复制</button>
<code id="cssUrl"><a href="%s" target="_blank" style="color:#89b4fa;text-decoration:none">%s</a></code>
</div>
</div>
<div class="permalink">
永久链接: <a href="/d/%s">/d/%s</a> · <small>可分享此链接，随时下载</small>
</div>
` + "%s" + `
</div>
<script>
function copyCode(btn,id){var el=document.getElementById(id);var text=el.textContent;navigator.clipboard.writeText(text).then(()=>{btn.textContent='已复制!';btn.classList.add('copied');setTimeout(()=>{btn.textContent='复制';btn.classList.remove('copied')},2000)}).catch(()=>{var ta=document.createElement('textarea');ta.value=text;document.body.appendChild(ta);ta.select();document.execCommand('copy');document.body.removeChild(ta);btn.textContent='已复制!';btn.classList.add('copied');setTimeout(()=>{btn.textContent='复制';btn.classList.remove('copied')},2000)})}
</script>
</body>
</html>`

const errorHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>下载失败 - %s | Google Fonts Tools</title>
<meta name="robots" content="noindex, nofollow">
<link rel="icon" type="image/x-icon" href="/assets/img/favicon.ico">
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:linear-gradient(135deg,#667eea 0%%,#764ba2 100%%);min-height:100vh;display:flex;align-items:center;justify-content:center}
.container{background:#fff;border-radius:16px;padding:40px;max-width:560px;width:90%%;box-shadow:0 20px 60px rgba(0,0,0,.3);text-align:center}
.icon{font-size:64px;margin-bottom:16px}
h1{color:#333;font-size:22px;margin-bottom:20px}
.info{text-align:left;margin-bottom:20px}
.info p{color:#555;font-size:14px;padding:6px 0;border-bottom:1px solid #f0f0f0}
.info strong{color:#333}
.btn{display:inline-block;padding:12px 24px;background:linear-gradient(135deg,#667eea,#764ba2);color:#fff;border:none;border-radius:8px;font-size:14px;font-weight:600;cursor:pointer;text-decoration:none;margin:4px;transition:transform .2s}
.btn:hover{transform:translateY(-2px)}
` + footerCSS + `
</style>
</head>
<body>
<div class="container">
<div class="icon">❌</div>
<h1>%s 下载失败</h1>
<div class="info">
<p>签名: <strong>%s</strong></p>
<p>错误: <strong>%s</strong></p>
</div>
<a href="/" class="btn">🏠 返回首页</a>
` + "%s" + `
</div>
</body>
</html>`

const notFoundHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>任务未找到 | Google Fonts Tools</title>
<meta name="robots" content="noindex, nofollow">
<link rel="icon" type="image/x-icon" href="/assets/img/favicon.ico">
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:linear-gradient(135deg,#667eea 0%%,#764ba2 100%%);min-height:100vh;display:flex;align-items:center;justify-content:center}
.container{background:#fff;border-radius:16px;padding:40px;max-width:560px;width:90%%;box-shadow:0 20px 60px rgba(0,0,0,.3);text-align:center}
.icon{font-size:64px;margin-bottom:16px}
h1{color:#333;font-size:22px;margin-bottom:12px}
p{color:#888;font-size:14px;margin-bottom:20px}
.btn{display:inline-block;padding:12px 24px;background:linear-gradient(135deg,#667eea,#764ba2);color:#fff;border:none;border-radius:8px;font-size:14px;font-weight:600;cursor:pointer;text-decoration:none;transition:transform .2s}
.btn:hover{transform:translateY(-2px)}
` + footerCSS + `
</style>
</head>
<body>
<div class="container">
<div class="icon">🔍</div>
<h1>任务未找到</h1>
<p>签名: %s</p>
<a href="/" class="btn">🏠 返回首页</a>
` + "%s" + `
</div>
</body>
</html>`

var recentHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>最近下载 - Google Fonts Download Tools</title>
<meta name="description" content="查看最近下载的 Google Fonts 字体列表，包含原始 URL、下载状态和文件大小，方便溯源和复用">
<meta name="keywords" content="Google Fonts,最近下载,字体下载记录,font download history">
<meta name="robots" content="index, follow">
<link rel="canonical" href="/recent">
<meta property="og:type" content="website">
<meta property="og:title" content="最近下载 - Google Fonts Tools">
<meta property="og:description" content="查看最近下载的 Google Fonts 字体列表，方便溯源和复用">
<meta property="og:image" content="/assets/img/logo.png">
<link rel="icon" type="image/x-icon" href="/assets/img/favicon.ico">
<link rel="apple-touch-icon" href="/assets/img/logo.png">
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:linear-gradient(135deg,#667eea 0%%,#764ba2 100%%);min-height:100vh;padding:20px}
.container{background:#fff;border-radius:16px;padding:32px;max-width:900px;margin:0 auto;box-shadow:0 20px 60px rgba(0,0,0,.3)}
h1{color:#333;font-size:22px;margin-bottom:20px}
table{width:100%%;border-collapse:collapse;font-size:13px}
th{background:#f8f9fa;color:#555;font-weight:600;padding:10px 8px;text-align:left;border-bottom:2px solid #eee}
td{padding:10px 8px;border-bottom:1px solid #f0f0f0;color:#555;vertical-align:top}
td a{color:#667eea;text-decoration:none;font-weight:500}
td a:hover{text-decoration:underline}
.font-name{font-weight:600;color:#333}
.url-cell{max-width:320px;word-break:break-all;font-size:12px;color:#888}
.url-cell a{color:#667eea}
.status-success{color:#22c55e;font-weight:600}
.status-failed{color:#ef4444;font-weight:600}
.status-running{color:#f59e0b;font-weight:600}
.status-pending{color:#94a3b8;font-weight:600}
.empty{text-align:center;color:#aaa;padding:40px 0;font-size:16px}
.back{display:inline-block;margin-top:16px;padding:10px 20px;background:linear-gradient(135deg,#667eea,#764ba2);color:#fff;border:none;border-radius:8px;font-size:14px;font-weight:600;cursor:pointer;text-decoration:none;transition:transform .2s}
.back:hover{transform:translateY(-2px)}
` + footerCSS + `
</style>
</head>
<body>
<div class="container">
<h1>📋 最近下载</h1>
%s
` + footerHTML() + `
</div>
</body>
</html>`
