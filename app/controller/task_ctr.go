package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/tekintian/googlefonts-tools/app/model"
	"github.com/tekintian/googlefonts-tools/app/service"
)

type TaskController struct{}

var TaskCtl = &TaskController{}

func (tc *TaskController) CreateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		URL          string `json:"url"`
		NotifyConfig string `json:"notify_config"`
	}

	if r.Method == "POST" {
		contentType := r.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/json") {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				jsonResp(w, 400, "invalid request body", nil)
				return
			}
		} else {
			req.URL = r.FormValue("url")
			req.NotifyConfig = r.FormValue("notify_config")
		}
	} else {
		req.URL = r.URL.Query().Get("url")
		req.NotifyConfig = r.URL.Query().Get("notify_config")
	}

	if req.URL == "" {
		jsonResp(w, 400, "url parameter is required", nil)
		return
	}

	if !strings.Contains(req.URL, "fonts.googleapis.com") && !strings.Contains(req.URL, "fonts.gstatic.com") {
		jsonResp(w, 400, "url must be a Google Fonts URL", nil)
		return
	}

	tm := service.DefaultTaskManager
	if tm == nil {
		jsonResp(w, 500, "task manager not initialized", nil)
		return
	}

	task, err := tm.Submit(req.URL, req.NotifyConfig)
	if err != nil {
		jsonResp(w, 500, err.Error(), nil)
		return
	}

	host := r.Host
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, host)

	resp := map[string]interface{}{
		"sign":         task.Sign,
		"status":       task.Status,
		"progress":     task.Progress,
		"font_name":    task.FontName,
		"permalink":    fmt.Sprintf("%s/d/%s", baseURL, task.Sign),
		"download_url": fmt.Sprintf("%s/d/%s/download", baseURL, task.Sign),
		"progress_url": fmt.Sprintf("%s/d/%s/progress", baseURL, task.Sign),
	}

	if task.Status == model.StatusSuccess {
		resp["zip_size"] = task.ZipSize
		resp["download_count"] = task.DownloadCount
	}

	jsonResp(w, 200, "ok", resp)
}

func (tc *TaskController) GetTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sign := extractPathParam(r.URL.Path, "/api/v1/tasks/", "")
	if sign == "" {
		jsonResp(w, 400, "sign is required", nil)
		return
	}

	tm := service.DefaultTaskManager
	task, err := tm.GetTask(sign)
	if err != nil {
		jsonResp(w, 500, err.Error(), nil)
		return
	}
	if task == nil {
		jsonResp(w, 404, "task not found", nil)
		return
	}

	jsonResp(w, 200, "ok", task)
}

func (tc *TaskController) ListTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}

	tm := service.DefaultTaskManager
	tasks, err := tm.ListTasks(offset, limit)
	if err != nil {
		jsonResp(w, 500, err.Error(), nil)
		return
	}

	jsonResp(w, 200, "ok", tasks)
}

func (tc *TaskController) TaskProgress(w http.ResponseWriter, r *http.Request) {
	sign := extractPathParam(r.URL.Path, "/api/v1/tasks/", "/progress")
	if sign == "" {
		http.Error(w, "sign is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	tm := service.DefaultTaskManager
	task, _ := tm.GetTask(sign)

	if task != nil && (task.Status == model.StatusSuccess || task.Status == model.StatusFailed) {
		fmt.Fprintf(w, "data: {\"sign\":\"%s\",\"status\":\"%s\",\"progress\":%d,\"error_msg\":\"%s\"}\n\n",
			task.Sign, task.Status, task.Progress, task.ErrorMsg)
		w.(http.Flusher).Flush()
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
			data, _ := json.Marshal(p)
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.(http.Flusher).Flush()
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

func jsonResp(w http.ResponseWriter, code int, msg string, data interface{}) {
	w.WriteHeader(code)
	resp := map[string]interface{}{
		"code": code,
		"msg":  msg,
	}
	if data != nil {
		resp["data"] = data
	}
	json.NewEncoder(w).Encode(resp)
}
