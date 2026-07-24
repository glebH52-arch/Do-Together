package main

import (
	"context"
	"do-together/internal/auth"
	"do-together/internal/config"
	"do-together/internal/handler"
	"do-together/internal/middleware"
	"do-together/internal/repository/postgres"
	"do-together/internal/service"
	"log"
	"net/http"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	pool, err := postgres.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect to postgres: %v", err)
	}
	defer pool.Close()
	authManager := auth.NewJWTManager(cfg.JWTSecret, cfg.AccessTokenTTL, "do-together")
	authMiddleware := middleware.NewAuthMiddleware(authManager)
	userRepository := postgres.NewPostgresUserRepository(pool)
	userService := service.NewUserService(userRepository)
	userHandler := handler.NewUserHandler(userService)
	authService := service.NewAuthService(userRepository, authManager)
	authHandler := handler.NewAuthHandler(authService)
	projectRepository := postgres.NewPostgresProjectRepository(pool)
	projectService := service.NewProjectService(projectRepository)
	projectHandler := handler.NewProjectHandler(projectService)
	router := handler.NewRouter(projectHandler, userHandler, authHandler, authMiddleware)
	err = http.ListenAndServe(":8080", router)
	if err != nil {
		log.Println(err)
	}
}
