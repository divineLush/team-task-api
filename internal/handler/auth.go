package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/login", h.Login)
	r.Post("/register", h.Register)
	return r
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {}
