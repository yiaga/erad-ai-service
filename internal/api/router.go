package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/yiaga/erad-ai-service/internal/handlers"
)

func NewRouter(jobHandler *handlers.JobHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	r.Get("/health", jobHandler.HealthCheck)

	r.Route("/jobs", func(r chi.Router) {
		r.Post("/", jobHandler.CreateJob)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", jobHandler.GetJobStatus)
			r.Get("/result", jobHandler.GetJobResult)
			r.Get("/flags", jobHandler.GetJobFlags)
		})
	})

	r.Post("/retry/{id}", jobHandler.RetryJob)

	return r
}
