package httpx

import (
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	R *chi.Mux
}

func NewServer(admin *AdminHandlers) *Server {
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Get("/", admin.Index)
    r.Get("/new", admin.NewForm)
    r.Post("/jobs", admin.CreateJob)
    r.Post("/jobs/{id}/cancel", admin.CancelJob)
    // Maintenance: reschedule all pending jobs (past 24h)
    r.Post("/maintenance/reschedule", admin.ReschedulePending)

	staticFS, err := fs.Sub(admin.Assets, ".")
	if err != nil {
		panic(err)
	}
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	return &Server{R: r}
}
