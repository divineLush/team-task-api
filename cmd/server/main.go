package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/team-task-api/internal/cache"
	"github.com/team-task-api/internal/config"
	"github.com/team-task-api/internal/handler"
	"github.com/team-task-api/internal/middleware"
	"github.com/team-task-api/internal/repository"
	"github.com/team-task-api/internal/service"
	"github.com/team-task-api/pkg/database"
	"github.com/team-task-api/pkg/logger"

	_ "github.com/team-task-api/docs"
)

// @title Team Task API
// @version 1.0
// @description API for managing tasks between teams
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
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

	taskCache := cache.NewTaskCache(rdb)

	userRepo := repository.NewUserRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	teamMemberRepo := repository.NewTeamMemberRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	taskCommentRepo := repository.NewTaskCommentRepository(db)
	taskHistoryRepo := repository.NewTaskHistoryRepository(db)

	authService := service.NewAuthService(db, userRepo, cfg.Auth)
	teamService := service.NewTeamService(db, teamRepo, teamMemberRepo)
	taskHistoryService := service.NewTaskHistoryService(db, taskHistoryRepo, taskRepo, teamMemberRepo)
	taskService := service.NewTaskService(db, taskRepo, teamMemberRepo, taskHistoryService, taskCache)
	taskCommentService := service.NewTaskCommentService(db, taskCommentRepo, taskRepo, teamMemberRepo)

	authHandler := handler.NewAuthHandler(authService)
	teamHandler := handler.NewTeamHandler(teamService)
	taskCommentHandler := handler.NewTaskCommentHandler(taskCommentService)
	taskHistoryHandler := handler.NewTaskHistoryHandler(taskHistoryService)
	taskHandler := handler.NewTaskHandler(taskService, taskCommentHandler, taskHistoryHandler)

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
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg.Auth.JWTSecret))
			r.Mount("/tasks", taskHandler.Routes())
			r.Mount("/teams", teamHandler.Routes())
		})
		r.Mount("/", authHandler.Routes())
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Error("server failed", "error", err)
	}
}
