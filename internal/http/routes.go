package httpx

import (
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
	// statik
	r.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(admin.Assets))))
	return &Server{R: r}
}
