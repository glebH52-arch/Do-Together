package handler

import (
	"bytes"
	"do-together/internal/middleware"
	"do-together/internal/service"
	"encoding/json"
	"net/http"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(p *service.UserService) *UserHandler {
	return &UserHandler{
		userService: p,
	}
}

func (u *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	request := createUserRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	user, err := u.userService.CreateUser(r.Context(), request.Username, request.Email, request.Password)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	response := userToResponse(user)
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(response)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(buf.Bytes())
	if err != nil {
		return
	}
}

func (u *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	id, ok := middleware.UserIDFromContext(r.Context())
	if !ok || id <= 0 {
		status := http.StatusUnauthorized
		http.Error(w, http.StatusText(status), status)
		return
	}
	user, err := u.userService.GetUser(r.Context(), id)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	response := userToResponse(user)
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(response)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(buf.Bytes())
	if err != nil {
		return
	}
}
