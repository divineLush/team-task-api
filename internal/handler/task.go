package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type TaskHandler struct{}

func NewTaskHandler() *TaskHandler {
	return &TaskHandler{}
}

func (h *TaskHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {}
