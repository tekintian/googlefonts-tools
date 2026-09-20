package controller

import (
	"net/http"
	"strings"

	"github.com/tekintian/googlefonts-tools/app/service"
	"github.com/tekintian/googlefonts-tools/utils"
)

type Router struct {
	mux *http.ServeMux
}

func NewRouter() *Router {
	return &Router{mux: http.NewServeMux()}
}

func (rt *Router) Setup(engine *service.DownloadEngine) {
	InitPageController(engine)

	contentDir := utils.GetPwd() + "/" + service.ContentDir
	fsHandler := http.StripPrefix("/c/", http.FileServer(http.Dir(contentDir)))
	cacheHandler := StaticCacheMiddleware(fsHandler)
	rt.mux.Handle("/c/", cacheHandler)

	rt.mux.HandleFunc("/", CorsMiddleware(LoggingMiddleware(PageCtl.Index)))

	rt.mux.HandleFunc("/recent", CorsMiddleware(LoggingMiddleware(PageCtl.Recent)))

	rt.mux.HandleFunc("/d/", CorsMiddleware(LoggingMiddleware(rt.handleSignRoutes)))

	rt.mux.HandleFunc("/api/v1/tasks", CorsMiddleware(LoggingMiddleware(rt.handleTasksRoot)))
	rt.mux.HandleFunc("/api/v1/tasks/", CorsMiddleware(LoggingMiddleware(rt.handleTaskAPIRoutes)))
}

func StaticCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
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
