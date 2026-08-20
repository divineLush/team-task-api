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

// List godoc
// @Summary      List tasks
// @Description  Get all tasks for a team
// @Tags         tasks
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   model.Task
// @Router       /api/v1/tasks [get]
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {}

// Create godoc
// @Summary      Create a task
// @Description  Create a new task
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateTaskRequest  true  "Task payload"
// @Success      201   {object}  model.Task
// @Failure      400   {object}  map[string]string
// @Router       /api/v1/tasks [post]
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {}

// GetByID godoc
// @Summary      Get a task
// @Description  Get a task by ID
// @Tags         tasks
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Task ID"
// @Success      200  {object}  model.Task
// @Router       /api/v1/tasks/{id} [get]
func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {}

// Update godoc
// @Summary      Update a task
// @Description  Update a task by ID
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string                  true  "Task ID"
// @Param        body body      model.UpdateTaskRequest  true  "Update payload"
// @Success      200  {object}  model.Task
// @Failure      400  {object}  map[string]string
// @Router       /api/v1/tasks/{id} [put]
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {}

// Delete godoc
// @Summary      Delete a task
// @Description  Delete a task by ID
// @Tags         tasks
// @Security     BearerAuth
// @Param        id   path      string  true  "Task ID"
// @Success      204
// @Router       /api/v1/tasks/{id} [delete]
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {}
