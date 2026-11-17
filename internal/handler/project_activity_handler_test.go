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
