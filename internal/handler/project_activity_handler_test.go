package handler

import (
	"encoding/json"
	"testing"
)

func TestSanitizeActivityJSON(t *testing.T) {
	input := json.RawMessage(`{
		"title":"Release",
		"description":"secret details",
		"body":"deleted comment",
		"nested":{"email":"invitee@example.com","role":"member"}
	}`)
	result := sanitizeActivityJSON(input)
	var decoded map[string]any
	if err := json.Unmarshal(result, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["title"] != "Release" || decoded["description"] != nil || decoded["body"] != nil {
		t.Fatalf("unexpected sanitized value: %s", result)
	}
	nested := decoded["nested"].(map[string]any)
	if nested["email"] != nil || nested["role"] != "member" {
		t.Fatalf("unexpected nested value: %s", result)
	}
}

func TestSanitizeActivityDescriptionMarker(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input string
		want  string
	}{
		{"marker", `{"changes":{"description":{"changed":true}}}`, `{"changes":{"description":{"changed":true}}}`},
		{"marker strips snapshots", `{"changes":{"description":{"changed":true,"old":"secret","new":"private","token":"secret"}}}`, `{"changes":{"description":{"changed":true}}}`},
		{"text remains hidden", `{"changes":{"description":{"old":"secret","new":"private"}}}`, `{"changes":{}}`},
		{"false marker hidden", `{"changes":{"description":{"changed":false}}}`, `{"changes":{}}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeActivityJSON(json.RawMessage(tt.input))
			if string(got) != tt.want {
				t.Fatalf("got %s, want %s", got, tt.want)
			}
		})
	}
}
