package controller

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/tekintian/googlefonts-tools/app/model"
	"github.com/tekintian/googlefonts-tools/app/service"
	"github.com/tekintian/googlefonts-tools/app/templates"
	"github.com/tekintian/googlefonts-tools/utils"
)

type PageController struct {
	engine *service.DownloadEngine
}

var PageCtl *PageController

func InitPageController(engine *service.DownloadEngine) {
	PageCtl = &PageController{engine: engine}
}

func (pc *PageController) Index(w http.ResponseWriter, r *http.Request) {
	urlParam := r.URL.Query().Get("url")
	if urlParam != "" {
		if !strings.Contains(urlParam, "fonts.googleapis.com") && !strings.Contains(urlParam, "fonts.gstatic.com") {
			templates.Render(w, "index.html", pc.buildIndexData())
			return
		}
		tm := service.DefaultTaskManager
		if tm != nil {
			task, err := tm.Submit(urlParam, "")
			if err == nil && task != nil {
				http.Redirect(w, r, "/d/"+task.Sign, http.StatusFound)
				return
			}
		}
	}
	templates.Render(w, "index.html", pc.buildIndexData())
}

func (pc *PageController) buildIndexData() templates.IndexData {
	data := templates.IndexData{AppVer: templates.AppVer}
	tm := service.DefaultTaskManager
	if tm == nil {
		return data
	}
	tasks, err := tm.ListSuccessTasks(0, 5)
	if err != nil || len(tasks) == 0 {
		return data
	}
	data.RecentItems = make([]templates.RecentItem, 0, len(tasks))
	for _, t := range tasks {
		data.RecentItems = append(data.RecentItems, templates.RecentItem{
			FontName:    t.FontName,
			OriginalURL: t.URL,
			Sign:        t.Sign,
			CreatedAt:   t.CreatedAt.Format("01-02 15:04"),
		})
	}
	return data
}

func (pc *PageController) SignPage(w http.ResponseWriter, r *http.Request) {
	sign := extractPathParam(r.URL.Path, "/d/", "")
	if sign == "" || strings.Contains(sign, "/") {
		pc.Index(w, r)
		return
	}

	tm := service.DefaultTaskManager
	task, _ := tm.GetTask(sign)

	if task == nil {
		templates.Render(w, "not_found.html", templates.NotFoundData{
			AppVer: templates.AppVer,
			Sign:   sign,
		})
		return
	}

	switch task.Status {
	case model.StatusSuccess:
		sizeStr := formatFileSize(task.ZipSize)
		var durationStr string
		if task.CompletedAt != nil {
			durationStr = task.CompletedAt.Sub(task.CreatedAt).Round(time.Second).String()
		} else {
			durationStr = "-"
		}
		cssPath := fmt.Sprintf("/c/%s/%s/%s.css", task.FontName, utils.ShortSign(task.Sign), task.FontName)
		zipURL := fmt.Sprintf("/c/d/%s_%s.zip", task.FontName, utils.ShortSign(task.Sign))
		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		cssLinkHref := fmt.Sprintf("//%s%s", r.Host, cssPath)
		cssFullURL := fmt.Sprintf("%s://%s%s", scheme, r.Host, cssPath)
		templates.Render(w, "result.html", templates.ResultData{
			AppVer:        templates.AppVer,
			FontName:      task.FontName,
			Sign:          task.Sign,
			URL:           task.URL,
			Size:          sizeStr,
			Duration:      durationStr,
			DownloadCount: task.DownloadCount,
			CreatedAt:     task.CreatedAt.Format("2006-01-02 15:04:05"),
			ZipURL:        zipURL,
			CSSLinkHref:   cssLinkHref,
			CSSFullURL:    cssFullURL,
		})

	case model.StatusPending, model.StatusRunning:
		templates.Render(w, "progress.html", templates.ProgressData{
			AppVer:     templates.AppVer,
			FontName:   task.FontName,
			Sign:       sign,
			Progress:   task.Progress,
			DoneFiles:  task.DoneFiles,
			TotalFiles: task.TotalFiles,
		})

	case model.StatusFailed:
		templates.Render(w, "error.html", templates.ErrorData{
			AppVer:   templates.AppVer,
			FontName: task.FontName,
			Sign:     task.Sign,
			ErrorMsg: task.ErrorMsg,
		})
	}
}

func (pc *PageController) SignDownload(w http.ResponseWriter, r *http.Request) {
	sign := extractPathParam(r.URL.Path, "/d/", "/download")
	if sign == "" {
		http.Error(w, "invalid sign", http.StatusBadRequest)
		return
	}

	tm := service.DefaultTaskManager
	task, err := tm.GetTask(sign)
	if err != nil || task == nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	if task.Status != model.StatusSuccess {
		http.Error(w, "task not completed yet", http.StatusNotFound)
		return
	}

	go tm.IncrementDownloadCount(sign)

	zipURL := fmt.Sprintf("/c/d/%s_%s.zip", task.FontName, utils.ShortSign(sign))
	http.Redirect(w, r, zipURL, http.StatusFound)
}

func (pc *PageController) SignProgress(w http.ResponseWriter, r *http.Request) {
	sign := extractPathParam(r.URL.Path, "/d/", "/progress")
	if sign == "" {
		http.Error(w, "invalid sign", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	tm := service.DefaultTaskManager
	task, _ := tm.GetTask(sign)

	if task != nil && (task.Status == model.StatusSuccess || task.Status == model.StatusFailed) {
		writeSSE(w, model.TaskProgress{
			Sign: task.Sign, Status: task.Status, Progress: task.Progress,
			DoneFiles: task.DoneFiles, TotalFiles: task.TotalFiles, ErrorMsg: task.ErrorMsg,
		})
		return
	}

	ch := tm.Subscribe(sign)
	defer tm.Unsubscribe(sign, ch)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case p, ok := <-ch:
			if !ok {
				return
			}
			writeSSE(w, p)
			if p.Status == model.StatusSuccess || p.Status == model.StatusFailed {
				return
			}
		case <-ticker.C:
			fmt.Fprintf(w, ": keepalive\n\n")
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func writeSSE(w http.ResponseWriter, p model.TaskProgress) {
	data, _ := json.Marshal(p)
	fmt.Fprintf(w, "data: %s\n\n", data)
	w.(http.Flusher).Flush()
}

func formatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
}

func (pc *PageController) Recent(w http.ResponseWriter, r *http.Request) {
	tm := service.DefaultTaskManager
	tasks, _ := tm.ListSuccessTasks(0, 50)

	if len(tasks) == 0 {
		templates.Render(w, "recent.html", templates.RecentData{
			AppVer: templates.AppVer,
			Rows:   template.HTML(`<p class="empty">暂无下载记录</p><a href="/" class="back">🏠 返回首页</a>`),
		})
		return
	}

	rows := `<table><tr><th>字体</th><th>Google URL</th><th>大小</th><th>时间</th><th>操作</th></tr>`
	for _, t := range tasks {
		sizeStr := "-"
		if t.ZipSize > 0 {
			sizeStr = formatFileSize(t.ZipSize)
		}

		rows += fmt.Sprintf(
			`<tr><td class="font-name">%s</td><td class="url-cell"><a href="%s" target="_blank">%s</a></td><td>%s</td><td>%s</td><td><a href="/d/%s">查看</a></td></tr>`,
			t.FontName, t.URL, t.URL, sizeStr,
			t.CreatedAt.Format("01-02 15:04"), t.Sign,
		)
	}
	rows += `</table><a href="/" class="back">🏠 返回首页</a>`

	templates.Render(w, "recent.html", templates.RecentData{
		AppVer: templates.AppVer,
		Rows:   template.HTML(rows),
	})
}

func extractPathParam(path, prefix, suffix string) string {
	s := strings.TrimPrefix(path, prefix)
	s = strings.TrimSuffix(s, suffix)
	s = strings.TrimRight(s, "/")
	parts := strings.SplitN(s, "/", 2)
	return parts[0]
}
