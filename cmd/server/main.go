package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

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

	log := logger.New(cfg.LogLevel)

	db, err := database.NewMySQL(cfg.DB)
	if err != nil {
		log.Error("mysql connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	rdb, err := database.NewRedis(cfg.RedisCfg)
	if err != nil {
		log.Error("redis connection failed", "error", err)
		os.Exit(1)
	}

	txm := database.NewTxManager(db)

	taskCache := cache.NewTaskCache(rdb)

	userRepo := repository.NewUserRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	teamMemberRepo := repository.NewTeamMemberRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	taskCommentRepo := repository.NewTaskCommentRepository(db)
	taskHistoryRepo := repository.NewTaskHistoryRepository(db)

	authService := service.NewAuthService(txm, userRepo, cfg.Auth)
	teamService := service.NewTeamService(txm, teamRepo, teamMemberRepo)
	taskHistoryService := service.NewTaskHistoryService(txm, taskHistoryRepo, taskRepo)
	taskService := service.NewTaskService(txm, taskRepo, teamMemberRepo, taskHistoryService, taskCache, log)
	taskCommentService := service.NewTaskCommentService(txm, taskCommentRepo, taskRepo)

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
	r.Use(middleware.BodyLimit)

	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg.Auth.JWTSecret))
			r.Mount("/tasks", taskHandler.Routes())
			r.Mount("/teams", teamHandler.Routes())
		})
		r.Mount("/", authHandler.Routes())
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		mysqlErr := db.PingContext(r.Context())
		redisErr := rdb.Ping(r.Context()).Err()

		if mysqlErr != nil || redisErr != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			errs := map[string]string{}
			if mysqlErr != nil {
				errs["mysql"] = mysqlErr.Error()
			}
			if redisErr != nil {
				errs["redis"] = redisErr.Error()
			}
			json.NewEncoder(w).Encode(map[string]any{"status": "unavailable", "errors": errs})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
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
