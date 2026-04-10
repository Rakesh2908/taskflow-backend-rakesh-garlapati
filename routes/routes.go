package routes

import (
	"net/http"

	"github.com/Rakesh2908/taskflow/config"
	"github.com/Rakesh2908/taskflow/middleware"
	"github.com/go-chi/chi/v5"
)

type AuthHandlers interface {
	Register(http.ResponseWriter, *http.Request)
	Login(http.ResponseWriter, *http.Request)
}

type ProjectHandlers interface {
	List(http.ResponseWriter, *http.Request)
	Create(http.ResponseWriter, *http.Request)
	GetByID(http.ResponseWriter, *http.Request)
	Stats(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
	Delete(http.ResponseWriter, *http.Request)
}

type TaskHandlers interface {
	ListForProject(http.ResponseWriter, *http.Request)
	CreateForProject(http.ResponseWriter, *http.Request)
	Update(http.ResponseWriter, *http.Request)
	Delete(http.ResponseWriter, *http.Request)
}

func NewRouter(cfg *config.Config, authH AuthHandlers, projectH ProjectHandlers, taskH TaskHandlers) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.ForceJSONContentType())

	r.Post("/auth/register", authH.Register)
	r.Post("/auth/login", authH.Login)

	r.Group(func(pr chi.Router) {
		pr.Use(middleware.JWTAuth(cfg.JWTSecret))

		pr.Route("/projects", func(rr chi.Router) {
			rr.Get("/", projectH.List)
			rr.Post("/", projectH.Create)

			rr.Route("/{id}", func(rrr chi.Router) {
				rrr.Get("/", projectH.GetByID)
				rrr.Get("/stats", projectH.Stats)
				rrr.Patch("/", projectH.Update)
				rrr.Delete("/", projectH.Delete)

				rrr.Route("/tasks", func(tr chi.Router) {
					tr.Get("/", taskH.ListForProject)
					tr.Post("/", taskH.CreateForProject)
				})
			})
		})

		pr.Route("/tasks", func(tr chi.Router) {
			tr.Route("/{id}", func(rr chi.Router) {
				rr.Patch("/", taskH.Update)
				rr.Delete("/", taskH.Delete)
			})
		})
	})

	return r
}

