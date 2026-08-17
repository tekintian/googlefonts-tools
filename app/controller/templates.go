package controller

import "fmt"

var AppVer = "dev"

const githubRepo = "https://github.com/tekintian/googlefonts-tools"

const footerCSS = `.footer{margin-top:20px;padding-top:12px;border-top:1px solid #eee;font-size:12px;color:#aaa;text-align:center}.footer a{color:#667eea;text-decoration:none}.footer a:hover{text-decoration:underline}`

func footerHTML() string {
	return fmt.Sprintf(`<div class="footer"><a href="%s" target="_blank">GoogleFonts Tools v%s</a> · Powered by <a href="https://ai.tekin.cn/" target="_blank">Tekin</a></div>`, githubRepo, AppVer)
}

func indexHTML() string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Google Fonts Download Tools</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);min-height:100vh;display:flex;align-items:center;justify-content:center}
.container{background:#fff;border-radius:16px;padding:40px;max-width:640px;width:90%;box-shadow:0 20px 60px rgba(0,0,0,.3)}
h1{color:#333;font-size:24px;margin-bottom:8px}
.subtitle{color:#888;font-size:14px;margin-bottom:24px}
label{display:block;color:#555;font-weight:600;margin-bottom:8px;font-size:14px}
input[type=text]{width:100%;padding:12px 16px;border:2px solid #e0e0e0;border-radius:8px;font-size:14px;transition:border-color .3s;outline:none}
input[type=text]:focus{border-color:#667eea}
.btn{display:inline-block;padding:12px 24px;background:linear-gradient(135deg,#667eea,#764ba2);color:#fff;border:none;border-radius:8px;font-size:16px;font-weight:600;cursor:pointer;width:100%;margin-top:16px;transition:transform .2s}
.btn:hover{transform:translateY(-2px)}
.btn:disabled{opacity:.6;cursor:not-allowed;transform:none}
.result{margin-top:20px;padding:16px;background:#f8f9fa;border-radius:8px;display:none}
.result h3{color:#333;margin-bottom:12px;font-size:16px}
.result a{color:#667eea;text-decoration:none;font-weight:600;word-break:break-all}
.result a:hover{text-decoration:underline}
.task-info{margin-top:8px;color:#666;font-size:13px}
.recent{margin-top:24px;border-top:1px solid #eee;padding-top:16px}
.recent h3{color:#555;font-size:14px;margin-bottom:8px}
.recent-list{list-style:none}
.recent-list li{padding:6px 0;font-size:13px}
.recent-list a{color:#667eea;text-decoration:none}
.recent-list a:hover{text-decoration:underline}
` + footerCSS + `
</style>
</head>
<body>
<div class="container">
<h1>🔤 Google Fonts Download</h1>
<p class="subtitle">输入 Google Fonts URL，异步下载字体文件并打包为 ZIP</p>
<form id="form" onsubmit="return submitTask(event)">
<label>Google Fonts URL</label>
<input type="text" id="url" placeholder="https://fonts.googleapis.com/css?family=Open+Sans:400,700&display=swap" required>
<button type="submit" class="btn" id="btn">提交下载任务</button>
</form>
<div class="result" id="result">
<h3>✅ 任务已提交</h3>
<p class="task-info">字体: <strong id="fontName">-</strong></p>
<p class="task-info">状态: <strong id="status">-</strong></p>
<p class="task-info">永久链接: <a id="permalink" href="#">-</a></p>
<p class="task-info" id="downloadRow" style="display:none">下载链接: <a id="downloadLink" href="#">点击下载 ZIP</a></p>
</div>
` + footerHTML() + `
</div>
<script>
function submitTask(e){
e.preventDefault();
var btn=document.getElementById('btn');
btn.disabled=true;
btn.textContent='提交中...';
var url=document.getElementById('url').value;
fetch('/api/v1/tasks',{method:'POST',headers:{'Content-Type':'application/x-www-form-urlencoded'},body:'url='+encodeURIComponent(url)})
.then(r=>r.json())
.then(data=>{
if(data.code!==200){alert('Error: '+data.msg);btn.disabled=false;btn.textContent='提交下载任务';return}
var d=data.data;
document.getElementById('result').style.display='block';
document.getElementById('fontName').textContent=d.font_name||'-';
document.getElementById('status').textContent=d.status;
document.getElementById('permalink').textContent=d.permalink;
document.getElementById('permalink').href=d.permalink;
if(d.status==='success'){
document.getElementById('downloadRow').style.display='block';
document.getElementById('downloadLink').href=d.download_url;
document.getElementById('status').textContent='✅ 已完成(缓存)';
}else{
document.getElementById('downloadRow').style.display='none';
window.location.href=d.permalink;
}
btn.disabled=false;
btn.textContent='提交下载任务';
})
.catch(err=>{alert('Request failed');btn.disabled=false;btn.textContent='提交下载任务'});
}
</script>
</body>
</html>`
}

const progressHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>下载中 - %s</title>
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
<title>下载完成 - %s</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:linear-gradient(135deg,#667eea 0%%,#764ba2 100%%);min-height:100vh;display:flex;align-items:center;justify-content:center}
.container{background:#fff;border-radius:16px;padding:40px;max-width:560px;width:90%%;box-shadow:0 20px 60px rgba(0,0,0,.3);text-align:center}
.icon{font-size:64px;margin-bottom:16px}
h1{color:#333;font-size:22px;margin-bottom:20px}
.info{text-align:left;margin-bottom:20px}
.info p{color:#555;font-size:14px;padding:6px 0;border-bottom:1px solid #f0f0f0}
.info strong{color:#333}
.btn{display:inline-block;padding:14px 32px;background:linear-gradient(135deg,#667eea,#764ba2);color:#fff;border:none;border-radius:8px;font-size:16px;font-weight:600;cursor:pointer;text-decoration:none;transition:transform .2s;margin:4px}
.btn:hover{transform:translateY(-2px)}
.btn-outline{display:inline-block;padding:12px 24px;background:#fff;color:#667eea;border:2px solid #667eea;border-radius:8px;font-size:14px;font-weight:600;cursor:pointer;text-decoration:none;transition:transform .2s;margin:4px}
.btn-outline:hover{transform:translateY(-2px);background:#f0f0ff}
.permalink{margin-top:20px;padding:12px;background:#f8f9fa;border-radius:8px;font-size:12px;color:#888;word-break:break-all}
.permalink a{color:#667eea;text-decoration:none}
` + footerCSS + `
</style>
</head>
<body>
<div class="container">
<div class="icon">✅</div>
<h1>%s 下载完成</h1>
<div class="info">
<p>签名: <strong>%s</strong></p>
<p>文件大小: <strong>%s</strong></p>
<p>耗时: <strong>%s</strong></p>
<p>下载次数: <strong>%d</strong></p>
<p>创建时间: <strong>%s</strong></p>
</div>
<a href="/d/%s/download" class="btn">⬇ 下载 ZIP 文件</a>
<a href="/" class="btn-outline">🏠 返回首页</a>
<div class="permalink">
永久链接: <a href="/d/%s">/d/%s</a><br>
<small>可分享此链接，随时下载</small>
</div>
` + "%s" + `
</div>
</body>
</html>`

const errorHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>下载失败 - %s</title>
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
<title>任务未找到</title>
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
