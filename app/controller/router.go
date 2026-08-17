package controller

import (
	"net/http"
	"strings"

	"github.com/tekintian/googlefonts-tools/app/service"
)

type Router struct {
	mux *http.ServeMux
}

func NewRouter() *Router {
	return &Router{mux: http.NewServeMux()}
}

func (rt *Router) Setup(engine *service.DownloadEngine) {
	InitPageController(engine)

	rt.mux.HandleFunc("/", CorsMiddleware(LoggingMiddleware(PageCtl.Index)))

	rt.mux.HandleFunc("/d/", CorsMiddleware(LoggingMiddleware(rt.handleSignRoutes)))

	rt.mux.HandleFunc("/api/v1/tasks", CorsMiddleware(LoggingMiddleware(rt.handleTasksRoot)))
	rt.mux.HandleFunc("/api/v1/tasks/", CorsMiddleware(LoggingMiddleware(rt.handleTaskAPIRoutes)))
}

func (rt *Router) handleTasksRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		TaskCtl.ListTasks(w, r)
	case http.MethodPost:
		TaskCtl.CreateTask(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (rt *Router) handleSignRoutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasSuffix(path, "/download") {
		PageCtl.SignDownload(w, r)
		return
	}
	if strings.HasSuffix(path, "/progress") {
		PageCtl.SignProgress(w, r)
		return
	}

	PageCtl.SignPage(w, r)
}

func (rt *Router) handleTaskAPIRoutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasSuffix(path, "/progress") {
		TaskCtl.TaskProgress(w, r)
		return
	}

	if path == "/api/v1/tasks/" {
		switch r.Method {
		case http.MethodGet:
			TaskCtl.ListTasks(w, r)
		case http.MethodPost:
			TaskCtl.CreateTask(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	TaskCtl.GetTask(w, r)
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rt.mux.ServeHTTP(w, r)
}
