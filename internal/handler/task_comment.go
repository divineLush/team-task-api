package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/team-task-api/internal/middleware"
	"github.com/team-task-api/internal/model"
	"github.com/team-task-api/internal/service"
)

type TaskCommentHandler struct {
	commentService *service.TaskCommentService
}

func NewTaskCommentHandler(commentService *service.TaskCommentService) *TaskCommentHandler {
	return &TaskCommentHandler{commentService: commentService}
}

func (h *TaskCommentHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Delete("/{commentID}", h.Delete)
	return r
}

// List godoc
// @Summary      List task comments
// @Description  Get all comments for a task. User must be a member of the task's team.
// @Tags         task-comments
// @Produce      json
// @Security     BearerAuth
// @Param        id        path   string  true  "Task ID"
// @Success      200       {array}  model.TaskComment
// @Failure      401       {object}  map[string]string
// @Failure      403       {object}  map[string]string
// @Failure      404       {object}  map[string]string
// @Router       /api/v1/tasks/{id}/comments [get]
func (h *TaskCommentHandler) List(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	comments, err := h.commentService.List(r.Context(), userID, taskID)
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

	writeJSON(w, http.StatusOK, comments)
}

// Create godoc
// @Summary      Create a task comment
// @Description  Add a comment to a task. User must be a member of the task's team.
// @Tags         task-comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                  true  "Task ID"
// @Param        body  body      model.CreateCommentRequest  true  "Comment payload"
// @Success      201   {object}  model.TaskComment
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Router       /api/v1/tasks/{id}/comments [post]
func (h *TaskCommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	var req model.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	comment, err := h.commentService.Create(r.Context(), userID, taskID, &req)
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

	writeJSON(w, http.StatusCreated, comment)
}

// Delete godoc
// @Summary      Delete a task comment
// @Description  Delete a comment from a task. User must be a member of the task's team.
// @Tags         task-comments
// @Security     BearerAuth
// @Param        id         path   string  true  "Task ID"
// @Param        commentID  path   string  true  "Comment ID"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/tasks/{id}/comments/{commentID} [delete]
func (h *TaskCommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	commentID := chi.URLParam(r, "commentID")

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	err := h.commentService.Delete(r.Context(), userID, taskID, commentID)
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
