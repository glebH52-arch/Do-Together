package domain

import (
	"encoding/json"
	"time"
	"uuid"
)

type ProjectActivity struct {
	ID            uuid.UUID
	ProjectID     uuid.UUID
	ActorID       uuid.UUID
	ActorName     string
	Type          string
	ObjectType    string
	ObjectID      uuid.UUID
	ObjectTitle   string
	ObjectDeleted bool
	OldValues     json.RawMessage
	NewValues     json.RawMessage
	OccurredAt    time.Time
}

func PublicActivityType(value, objectType string) string {
	switch value {
	case "created":
		return "task.created"
	case "edited":
		if objectType == "comment" {
			return "comment.updated"
		}
		return "task.updated"
	case "status_changed":
		return "task.status_changed"
	case "priority_changed":
		return "task.priority_changed"
	case "assignee_changed":
		return "task.assignee_changed"
	case "due_at_changed":
		return "task.due_at_changed"
	case "comment_created":
		return "comment.created"
	case "comment_deleted":
		return "comment.deleted"
	default:
		return value
	}
}
