package handler

import (
	"bytes"
	"do-together/internal/service"
	"encoding/json"
	"net/http"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (a *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	request := loginRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	accessToken, refreshToken, expiresIn, err := a.authService.Login(r.Context(), request.Email, request.Password)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	response := authToResponse(accessToken, refreshToken, expiresIn)
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

func (a *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	request := refreshRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	accessToken, newRefreshToken, expiresIn, err := a.authService.Refresh(r.Context(), request.RefreshToken)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	response := authToResponse(accessToken, newRefreshToken, expiresIn)
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

func (a *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	request := refreshRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	err = a.authService.Logout(r.Context(), request.RefreshToken)
	if err != nil {
		status := statusFromError(err)
		http.Error(w, http.StatusText(status), status)
		return
	}
	w.WriteHeader(http.StatusNoContent)

}
