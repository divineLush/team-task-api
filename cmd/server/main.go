package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/team-task-api/internal/config"
	"github.com/team-task-api/internal/handler"
	"github.com/team-task-api/internal/middleware"
	"github.com/team-task-api/internal/repository"
	"github.com/team-task-api/internal/service"
	"github.com/team-task-api/pkg/database"
	"github.com/team-task-api/pkg/logger"
)

func main() {
	cfg := config.Load()

	log := logger.New("info")

	db, err := database.NewMySQL(cfg.DB)
	if err != nil {
		log.Error("mysql connection failed", "error", err)
	}
	defer db.Close()

	rdb, err := database.NewRedis(cfg.RedisCfg)
	if err != nil {
		log.Error("redis connection failed", "error", err)
	}
	_ = rdb

	userRepo := repository.NewUserRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	teamMemberRepo := repository.NewTeamMemberRepository(db)

	authService := service.NewAuthService(db, userRepo, cfg.Auth)
	teamService := service.NewTeamService(db, teamRepo, teamMemberRepo)

	authHandler := handler.NewAuthHandler(authService)
	teamHandler := handler.NewTeamHandler(teamService)

	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.NewRateLimiter(100, 200).Limit)
	r.Use(middleware.Logger(log))
	r.Use(chimw.Recoverer)
	r.Use(chimw.Compress(5))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/tasks", handler.NewTaskHandler().Routes())
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg.Auth.JWTSecret))
			r.Mount("/teams", teamHandler.Routes())
		})
		r.Mount("/", authHandler.Routes())
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Error("server failed", "error", err)
	}
}
