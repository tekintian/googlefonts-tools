package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tekintian/googlefonts-tools/app/model"
	"github.com/tekintian/googlefonts-tools/app/service"
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
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(indexHTML))
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(indexHTML))
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
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, notFoundHTML, sign)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	switch task.Status {
	case model.StatusSuccess:
		sizeStr := formatFileSize(task.ZipSize)
		var durationStr string
		if task.CompletedAt != nil {
			durationStr = task.CompletedAt.Sub(task.CreatedAt).Round(time.Second).String()
		} else {
			durationStr = "-"
		}
		fmt.Fprintf(w, resultHTML,
			task.FontName, task.FontName, task.Sign,
			sizeStr, durationStr, task.DownloadCount,
			task.CreatedAt.Format("2006-01-02 15:04:05"),
			sign, sign, sign,
		)

	case model.StatusPending, model.StatusRunning:
		fmt.Fprintf(w, progressHTML,
			task.FontName, task.FontName, sign,
			task.Progress, task.Progress,
			task.DoneFiles, task.TotalFiles, sign,
		)

	case model.StatusFailed:
		fmt.Fprintf(w, errorHTML,
			task.FontName, task.FontName, task.Sign, task.ErrorMsg,
		)
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

	if err := pc.engine.ServeZipFile(w, task.ZipPath, task.FontName); err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
	}
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

func extractPathParam(path, prefix, suffix string) string {
	s := strings.TrimPrefix(path, prefix)
	s = strings.TrimSuffix(s, suffix)
	s = strings.TrimRight(s, "/")
	parts := strings.SplitN(s, "/", 2)
	return parts[0]
}
