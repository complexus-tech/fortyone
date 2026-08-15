package storieshttp

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestGetUpdatesPreservesNullableStrategyRelationships(t *testing.T) {
	keyResultID := uuid.New()
	requestData := map[string]json.RawMessage{
		"objectiveId": json.RawMessage("null"),
		"keyResultId": json.RawMessage(`"` + keyResultID.String() + `"`),
	}

	updates, err := getUpdates(requestData)
	if err != nil {
		t.Fatalf("expected updates to decode, got %v", err)
	}

	objectiveID, ok := updates["objective_id"].(*uuid.UUID)
	if !ok || objectiveID != nil {
		t.Fatalf("expected a nil objective pointer, got %#v", updates["objective_id"])
	}

	decodedKeyResultID, ok := updates["key_result_id"].(*uuid.UUID)
	if !ok || decodedKeyResultID == nil || *decodedKeyResultID != keyResultID {
		t.Fatalf("expected key result %s, got %#v", keyResultID, updates["key_result_id"])
	}
}

func TestGetUpdatesPreservesFalseAutoSchedulingPreferences(t *testing.T) {
	t.Parallel()

	updates, err := getUpdates(map[string]json.RawMessage{
		"autoSchedulingEnabled": json.RawMessage(`false`),
		"autoSchedulingLocked":  json.RawMessage(`false`),
		// System-owned scheduling state is deliberately not an accepted update.
		"autoSchedulingStatus":    json.RawMessage(`"scheduled"`),
		"autoSchedulingReason":    json.RawMessage(`"spoofed"`),
		"autoSchedulingUpdatedAt": json.RawMessage(`"2026-08-15T09:30:00Z"`),
	})
	if err != nil {
		t.Fatalf("get updates: %v", err)
	}

	if enabled, ok := updates["auto_scheduling_enabled"].(bool); !ok || enabled {
		t.Fatalf("auto_scheduling_enabled = %#v, want false", updates["auto_scheduling_enabled"])
	}
	if locked, ok := updates["auto_scheduling_locked"].(bool); !ok || locked {
		t.Fatalf("auto_scheduling_locked = %#v, want false", updates["auto_scheduling_locked"])
	}
	for _, field := range []string{"auto_scheduling_status", "auto_scheduling_reason", "auto_scheduling_updated_at"} {
		if _, accepted := updates[field]; accepted {
			t.Fatalf("system-owned field %q was accepted", field)
		}
	}
}
