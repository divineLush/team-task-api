package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type TeamHandler struct{}

func NewTeamHandler() *TeamHandler {
	return &TeamHandler{}
}

func (h *TeamHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

func (h *TeamHandler) List(w http.ResponseWriter, r *http.Request) {}

func (h *TeamHandler) Create(w http.ResponseWriter, r *http.Request) {}

func (h *TeamHandler) GetByID(w http.ResponseWriter, r *http.Request) {}

func (h *TeamHandler) Update(w http.ResponseWriter, r *http.Request) {}

func (h *TeamHandler) Delete(w http.ResponseWriter, r *http.Request) {}
