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
	redisClient, err := redisrepo.NewRedisClient(context.Background(), cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("connect to redis: %v", err)
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
	router := handler.NewRouter(projectHandler, userHandler, authHandler, authMiddleware, projectMemberHandler, inviteHandler, taskHandler)
	err = http.ListenAndServe(":8080", router)
	if err != nil {
		log.Println(err)
	}
}
