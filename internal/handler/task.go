package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/team-task-api/internal/middleware"
	"github.com/team-task-api/internal/model"
	"github.com/team-task-api/internal/repository"
	"github.com/team-task-api/internal/service"
)

type TaskHandler struct {
	taskService         *service.TaskService
	taskCommentHandler  *TaskCommentHandler
	taskHistoryHandler  *TaskHistoryHandler
}

func NewTaskHandler(taskService *service.TaskService, commentHandler *TaskCommentHandler, historyHandler *TaskHistoryHandler) *TaskHandler {
	return &TaskHandler{taskService: taskService, taskCommentHandler: commentHandler, taskHistoryHandler: historyHandler}
}

func (h *TaskHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Mount("/{id}/comments", h.taskCommentHandler.Routes())
	r.Mount("/{id}/history", h.taskHistoryHandler.Routes())
	return r
}

// List godoc
// @Summary      List tasks
// @Description  Get tasks with optional filters. Only returns tasks from teams the user belongs to.
// @Tags         tasks
// @Produce      json
// @Security     BearerAuth
// @Param        team_id     query   string  false  "Filter by team ID"
// @Param        status      query   string  false  "Filter by status (pending, in_progress, done)"
// @Param        assignee_id query   string  false  "Filter by assignee user ID"
// @Param        limit       query   int     false  "Limit (default 50, max 100)"
// @Param        offset      query   int     false  "Offset (default 0)"
// @Success      200  {array}   model.Task
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/v1/tasks [get]
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	filter := repository.TaskFilter{}

	if teamID := r.URL.Query().Get("team_id"); teamID != "" {
		filter.TeamIDs = []string{teamID}
	}
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
	}
	if assigneeID := r.URL.Query().Get("assignee_id"); assigneeID != "" {
		filter.AssigneeID = &assigneeID
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		if v, err := strconv.Atoi(limit); err == nil {
			filter.Limit = v
		}
	}
	if offset := r.URL.Query().Get("offset"); offset != "" {
		if v, err := strconv.Atoi(offset); err == nil {
			filter.Offset = v
		}
	}

	tasks, err := h.taskService.List(r.Context(), userID, filter)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotTeamMember):
			writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, tasks)
}

// Create godoc
// @Summary      Create a task
// @Description  Create a new task. User must be a member of the team.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateTaskRequest  true  "Task payload"
// @Success      201   {object}  model.Task
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Router       /api/v1/tasks [post]
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	task, err := h.taskService.Create(r.Context(), userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotTeamMember):
			writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, task)
}

// GetByID godoc
// @Summary      Get a task
// @Description  Get a task by ID. User must be a member of the task's team.
// @Tags         tasks
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Task ID"
// @Success      200  {object}  model.Task
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/tasks/{id} [get]
func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	task, err := h.taskService.GetByID(r.Context(), userID, taskID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotTeamMember):
			writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		case errors.Is(err, service.ErrTaskNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, task)
}

// Update godoc
// @Summary      Update a task
// @Description  Update a task by ID. User must be a member of the task's team.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string                  true  "Task ID"
// @Param        body body      model.UpdateTaskRequest  true  "Update payload"
// @Success      200  {object}  model.Task
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/tasks/{id} [put]
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	var req model.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	task, err := h.taskService.Update(r.Context(), userID, taskID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotTeamMember), errors.Is(err, service.ErrAssigneeNotMember):
			writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		case errors.Is(err, service.ErrTaskNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, task)
}

// Delete godoc
// @Summary      Delete a task
// @Description  Delete a task by ID. User must be a member of the task's team.
// @Tags         tasks
// @Security     BearerAuth
// @Param        id   path      string  true  "Task ID"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/tasks/{id} [delete]
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	err := h.taskService.Delete(r.Context(), userID, taskID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotTeamMember):
			writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		case errors.Is(err, service.ErrTaskNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
