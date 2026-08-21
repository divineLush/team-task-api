package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/team-task-api/internal/middleware"
	"github.com/team-task-api/internal/model"
	"github.com/team-task-api/internal/service"
)

type TeamHandler struct {
	teamService *service.TeamService
}

func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

func (h *TeamHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/invite", h.Invite)
	return r
}

// List godoc
// @Summary      List user's teams
// @Description  Get all teams the authenticated user belongs to
// @Tags         teams
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   model.Team
// @Failure      401  {object}  map[string]string
// @Router       /api/v1/teams [get]
func (h *TeamHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	teams, err := h.teamService.ListByUser(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, teams)
}

// Create godoc
// @Summary      Create a team
// @Description  Create a new team. The authenticated user becomes the owner.
// @Tags         teams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      model.CreateTeamRequest  true  "Team payload"
// @Success      201   {object}  model.Team
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Router       /api/v1/teams [post]
func (h *TeamHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "team name is required"})
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	team, err := h.teamService.Create(r.Context(), userID, &req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusCreated, team)
}

// GetByID godoc
// @Summary      Get a team
// @Description  Get a team by ID. User must be a member of the team.
// @Tags         teams
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Team ID"
// @Success      200  {object}  model.Team
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/teams/{id} [get]
func (h *TeamHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "id")

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	team, err := h.teamService.GetByID(r.Context(), userID, teamID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotMember):
			writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		case errors.Is(err, service.ErrTeamNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, team)
}

// Update godoc
// @Summary      Update a team
// @Description  Update a team by ID. Caller must be owner or admin.
// @Tags         teams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string                  true  "Team ID"
// @Param        body body      model.UpdateTeamRequest  true  "Update payload"
// @Success      200  {object}  model.Team
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/teams/{id} [put]
func (h *TeamHandler) Update(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "id")

	var req model.UpdateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	team, err := h.teamService.Update(r.Context(), userID, teamID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotMember), errors.Is(err, service.ErrForbidden):
			writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		case errors.Is(err, service.ErrTeamNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, team)
}

// Delete godoc
// @Summary      Delete a team
// @Description  Delete a team by ID. Caller must be the team owner.
// @Tags         teams
// @Security     BearerAuth
// @Param        id   path      string  true  "Team ID"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/teams/{id} [delete]
func (h *TeamHandler) Delete(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "id")

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	err := h.teamService.Delete(r.Context(), userID, teamID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotMember), errors.Is(err, service.ErrForbidden):
			writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		case errors.Is(err, service.ErrTeamNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Invite godoc
// @Summary      Invite user to team
// @Description  Add a user to the team. Caller must be owner or admin.
// @Tags         teams
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string                true  "Team ID"
// @Param        body body      model.InviteRequest   true  "Invite payload"
// @Success      201   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Router       /api/v1/teams/{id}/invite [post]
func (h *TeamHandler) Invite(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "id")

	var req model.InviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	req.UserID = strings.TrimSpace(req.UserID)
	if req.UserID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "user_id is required"})
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	err := h.teamService.Invite(r.Context(), userID, teamID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotMember), errors.Is(err, service.ErrForbidden):
			writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		case errors.Is(err, service.ErrInvalidRole):
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		case errors.Is(err, service.ErrAlreadyMember):
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "user invited"})
}
