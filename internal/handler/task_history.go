package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/team-task-api/internal/middleware"
	"github.com/team-task-api/internal/service"
)

type TaskHistoryHandler struct {
	historyService *service.TaskHistoryService
}

func NewTaskHistoryHandler(historyService *service.TaskHistoryService) *TaskHistoryHandler {
	return &TaskHistoryHandler{historyService: historyService}
}

func (h *TaskHistoryHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	return r
}

// List godoc
// @Summary      List task history
// @Description  Get the change history for a task. User must be a member of the task's team.
// @Tags         task-history
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Task ID"
// @Success      200  {array}   model.TaskHistory
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/tasks/{id}/history [get]
func (h *TaskHistoryHandler) List(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "task id is required"})
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	history, err := h.historyService.List(r.Context(), userID, taskID)
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

	writeJSON(w, http.StatusOK, history)
}
