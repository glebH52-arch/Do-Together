package main

import (
	"context"
	"do-together/internal/auth"
	"do-together/internal/config"
	"do-together/internal/handler"
	"do-together/internal/middleware"
	"do-together/internal/repository/postgres"
	"do-together/internal/repository/redis"
	redisrepo "do-together/internal/repository/redis"
	"do-together/internal/service"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}

}
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	pool, err := postgres.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to postgres: %v", err)
	}
	defer pool.Close()
	redisClient, err := redisrepo.NewRedisClient(context.Background(), cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		return fmt.Errorf("connect to redis: %v", err)
	}
	defer redisClient.Close()
	refreshSessionRepository := redis.NewRefreshSessionRepository(redisClient)
	authManager := auth.NewJWTManager(cfg.JWTSecret, cfg.AccessTokenTTL, "do-together")
	authMiddleware := middleware.NewAuthMiddleware(authManager)
	userRepository := postgres.NewPostgresUserRepository(pool)
	userService := service.NewUserService(userRepository)
	userHandler := handler.NewUserHandler(userService)
	authService := service.NewAuthService(
		userRepository,
		authManager,
		refreshSessionRepository,
		cfg.RefreshTokenIdleTTL,
		cfg.RefreshTokenAbsoluteTTL,
	)
	authHandler := handler.NewAuthHandler(authService)
	projectRepository := postgres.NewPostgresProjectRepository(pool)
	projectService := service.NewProjectService(projectRepository)
	projectHandler := handler.NewProjectHandler(projectService)
	projectMemberRepository := postgres.NewPostgresProjectMemberRepository(pool)
	projectMemberService := service.NewProjectMemberService(projectMemberRepository)
	projectMemberHandler := handler.NewProjectMemberHandler(projectMemberService)
	inviteRepository := postgres.NewPostgresInviteRepository(pool)
	inviteService := service.NewInviteService(inviteRepository)
	inviteHandler := handler.NewInviteHandler(inviteService)
	taskRepository := postgres.NewTaskRepository(pool)
	taskService := service.NewTaskService(taskRepository)
	taskHandler := handler.NewTaskHandler(taskService)
	healthHandler := handler.NewHealthHandler(pool, redisClient, 2*time.Second)
	router := handler.NewRouter(projectHandler, userHandler, authHandler, authMiddleware, projectMemberHandler, inviteHandler, taskHandler, healthHandler)
	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       3 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		log.Println("shutdown signal received")

	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server error: %v", err)
		}
	}
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %v", err)
	}
	return nil
}
