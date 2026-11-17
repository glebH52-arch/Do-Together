package handler

import (
	"bytes"
	"do-together/internal/domain"
	"do-together/internal/middleware"
	"do-together/internal/service"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"uuid"

	"github.com/gorilla/mux"
)

type UserHandler struct {
	userService   *service.UserService
	avatarStorage *service.AvatarStorage
	avatarBaseURL string
}

func NewUserHandler(p *service.UserService, avatarStorage *service.AvatarStorage, avatarBaseURL string) *UserHandler {
	return &UserHandler{
		userService:   p,
		avatarStorage: avatarStorage,
		avatarBaseURL: strings.TrimRight(avatarBaseURL, "/"),
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
	response := userToResponse(user, u.avatarURL(user))
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
	if !ok || id == uuid.Nil() {
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
	response := userToResponse(user, u.avatarURL(user))
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

func (u *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	id, ok := middleware.UserIDFromContext(r.Context())
	if !ok || id == uuid.Nil() {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	var request struct {
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	user, err := u.userService.UpdateDisplayName(r.Context(), id, request.DisplayName)
	if err != nil {
		u.writeError(w, err)
		return
	}
	u.writeUser(w, http.StatusOK, user)
}

func (u *UserHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	id, ok := middleware.UserIDFromContext(r.Context())
	if !ok || id == uuid.Nil() {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, service.MaxAvatarBytes+(1<<20))
	file, _, err := r.FormFile("avatar")
	if err != nil {
		http.Error(w, "avatar file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()
	filename, err := u.avatarStorage.Save(file)
	if err != nil {
		u.writeError(w, err)
		return
	}
	user, previous, err := u.userService.SetAvatar(r.Context(), id, filename, time.Now())
	if err != nil {
		_ = u.avatarStorage.Delete(filename)
		u.writeError(w, err)
		return
	}
	if previous != nil {
		_ = u.avatarStorage.Delete(*previous)
	}
	u.writeUser(w, http.StatusOK, user)
}

func (u *UserHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	id, ok := middleware.UserIDFromContext(r.Context())
	if !ok || id == uuid.Nil() {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	user, previous, err := u.userService.RemoveAvatar(r.Context(), id, time.Now())
	if err != nil {
		u.writeError(w, err)
		return
	}
	if previous != nil {
		_ = u.avatarStorage.Delete(*previous)
	}
	u.writeUser(w, http.StatusOK, user)
}

func (u *UserHandler) ServeAvatar(w http.ResponseWriter, r *http.Request) {
	filename := mux.Vars(r)["filename"]
	path, ok := u.avatarStorage.Path(filename)
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeFile(w, r, path)
}

func (u *UserHandler) avatarURL(user *domain.User) *string {
	if user.AvatarFilename == nil {
		return nil
	}
	value := u.avatarBaseURL + "/avatars/" + url.PathEscape(*user.AvatarFilename)
	if user.AvatarUpdatedAt != nil {
		value += "?v=" + fmt.Sprint(user.AvatarUpdatedAt.UnixNano())
	}
	return &value
}

func (u *UserHandler) writeUser(w http.ResponseWriter, status int, user *domain.User) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(userToResponse(user, u.avatarURL(user)))
}

func (u *UserHandler) writeError(w http.ResponseWriter, err error) {
	status := statusFromError(err)
	message := http.StatusText(status)
	if status == http.StatusBadRequest {
		message = err.Error()
	}
	http.Error(w, message, status)
}
