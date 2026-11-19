package handler

import (
	"bytes"
	"do-together/internal/domain"
	"do-together/internal/middleware"
	"do-together/internal/service"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"uuid"

	"github.com/gorilla/mux"
)

type ProjectActivityHandler struct {
	service *service.ProjectActivityService
}

func NewProjectActivityHandler(service *service.ProjectActivityService) *ProjectActivityHandler {
	return &ProjectActivityHandler{service: service}
}

type activityActorResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type activityObjectResponse struct {
	Type    string    `json:"type"`
	ID      uuid.UUID `json:"id"`
	Title   string    `json:"title"`
	Deleted bool      `json:"deleted"`
}

type activityDataResponse struct {
	Old json.RawMessage `json:"old,omitempty"`
	New json.RawMessage `json:"new,omitempty"`
}

type projectActivityResponse struct {
	ID         uuid.UUID              `json:"id"`
	Type       string                 `json:"type"`
	OccurredAt time.Time              `json:"occurred_at"`
	Actor      activityActorResponse  `json:"actor"`
	Object     activityObjectResponse `json:"object"`
	Data       activityDataResponse   `json:"data"`
}

type projectActivityPageResponse struct {
	Items      []projectActivityResponse `json:"items"`
	NextCursor *string                   `json:"next_cursor"`
}

func activityToResponse(item *domain.ProjectActivity) projectActivityResponse {
	return projectActivityResponse{
		ID: item.ID, Type: item.Type, OccurredAt: item.OccurredAt,
		Actor: activityActorResponse{ID: item.ActorID, Name: item.ActorName},
		Object: activityObjectResponse{
			Type: item.ObjectType, ID: item.ObjectID, Title: item.ObjectTitle, Deleted: item.ObjectDeleted,
		},
		Data: activityDataResponse{Old: sanitizeActivityJSON(item.OldValues), New: sanitizeActivityJSON(item.NewValues)},
	}
}

func sanitizeActivityJSON(value json.RawMessage) json.RawMessage {
	if len(value) == 0 || bytes.Equal(value, []byte("null")) {
		return nil
	}
	var decoded any
	if err := json.Unmarshal(value, &decoded); err != nil {
		return nil
	}
	removeSensitiveActivityFields(decoded)
	result, err := json.Marshal(decoded)
	if err != nil || bytes.Equal(result, []byte("{}")) {
		return nil
	}
	return result
}

func removeSensitiveActivityFields(value any) {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			switch key {
			case "description":
				// Expose only the change marker, never description text or snapshots.
				if marker, ok := child.(map[string]any); ok && marker["changed"] == true {
					typed[key] = map[string]any{"changed": true}
				} else {
					delete(typed, key)
				}
			case "body", "email", "code", "token", "password", "password_hash":
				delete(typed, key)
			default:
				removeSensitiveActivityFields(child)
			}
		}
	case []any:
		for _, child := range typed {
			removeSensitiveActivityFields(child)
		}
	}
}

func (h *ProjectActivityHandler) List(w http.ResponseWriter, r *http.Request) {
	actorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || actorID == uuid.Nil() {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	projectID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil || projectID == uuid.Nil() {
		http.Error(w, "invalid project id", http.StatusBadRequest)
		return
	}
	limit := service.DefaultActivityLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil {
			http.Error(w, service.ErrInvalidActivityLimit.Error(), http.StatusBadRequest)
			return
		}
	}
	page, err := h.service.List(r.Context(), actorID, projectID, r.URL.Query().Get("cursor"), limit)
	if err != nil {
		status := statusFromError(err)
		message := http.StatusText(status)
		if status == http.StatusBadRequest {
			message = err.Error()
		}
		http.Error(w, message, status)
		return
	}
	response := projectActivityPageResponse{Items: make([]projectActivityResponse, 0, len(page.Items)), NextCursor: page.NextCursor}
	for _, item := range page.Items {
		response.Items = append(response.Items, activityToResponse(item))
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
