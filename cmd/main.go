package main

import (
	"context"
	"do-together/internal/handler"
	"do-together/internal/repository"
	"do-together/internal/repository/postgres"
	"do-together/internal/service"
	"log"
	"net/http"
	"os"
)

func main() {

	connectionString := os.Getenv("connectionString")
	if connectionString == "" {
		log.Fatal("connectionString is not set")
	}
	pool, err := postgres.NewPool(context.Background(), connectionString)
	if err != nil {
		log.Fatalf("connect to postgres: %v", err)
	}
	defer pool.Close()
	userRepository := postgres.NewPostgresUserRepository(pool)
	userService := service.NewUserService(userRepository)
	userHandler := handler.NewUserHandler(userService)

	repo := repository.NewMemoryProjectRepository()

	projectService := service.NewProjectService(repo)

	projectHandler := handler.NewProjectHandler(projectService)

	router := handler.NewRouter(projectHandler, userHandler)

	err = http.ListenAndServe(":8080", router)
	if err != nil {
		log.Println(err)
	}
}
