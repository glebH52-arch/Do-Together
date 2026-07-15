package main

import (
	"do-together/internal/handler"
	"do-together/internal/repository"
	"do-together/internal/service"
	"log"
	"net/http"
)

func main() {
	repo := repository.NewMemoryProjectRepository()

	projectService := service.NewProjectService(repo)

	projectHandler := handler.NewProjectHandler(projectService)

	router := handler.NewRouter(projectHandler)

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Println(err)
	}
}
