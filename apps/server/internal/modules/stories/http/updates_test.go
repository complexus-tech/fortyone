package storieshttp

import (
	"encoding/json"
	"testing"
	"time"

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

func TestParseStoryPatchRejectsUnknownAndMalformedFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data map[string]json.RawMessage
	}{
		{name: "unknown", data: map[string]json.RawMessage{"reporterId": json.RawMessage(`"` + uuid.NewString() + `"`)}},
		{name: "null title", data: map[string]json.RawMessage{"title": json.RawMessage(`null`)}},
		{name: "malformed UUID", data: map[string]json.RawMessage{"assigneeId": json.RawMessage(`"not-a-uuid"`)}},
		{name: "fractional duration", data: map[string]json.RawMessage{"estimatedDurationMinutes": json.RawMessage(`12.5`)}},
		{name: "empty update", data: map[string]json.RawMessage{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := parseStoryPatch(test.data); err == nil {
				t.Fatal("parseStoryPatch() error = nil")
			}
		})
	}
}

func TestParseStoryPatchPreservesNullableValuesAndDates(t *testing.T) {
	t.Parallel()

	assigneeID := uuid.New()
	patch, err := parseStoryPatch(map[string]json.RawMessage{
		"description": json.RawMessage(`null`),
		"assigneeId":  json.RawMessage(`"` + assigneeID.String() + `"`),
		"startDate":   json.RawMessage(`"2026-08-28"`),
	})
	if err != nil {
		t.Fatalf("parseStoryPatch() error = %v", err)
	}
	if value, specified := patch.Description.Value(); !specified || value != nil {
		t.Fatalf("description = value %#v specified %v, want explicit null", value, specified)
	}
	if value, specified := patch.AssigneeID.Value(); !specified || value == nil || *value != assigneeID {
		t.Fatalf("assignee = value %#v specified %v", value, specified)
	}
	wantDate := time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC)
	if value, specified := patch.StartDate.Value(); !specified || value == nil || !value.Equal(wantDate) {
		t.Fatalf("start date = value %#v specified %v, want %v", value, specified, wantDate)
	}
}

func TestGetUpdatesPreservesFalseAutoSchedulingPreferences(t *testing.T) {
	t.Parallel()

	updates, err := getUpdates(map[string]json.RawMessage{
		"autoSchedulingEnabled": json.RawMessage(`false`),
		"autoSchedulingLocked":  json.RawMessage(`false`),
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
	for _, field := range []string{"autoSchedulingStatus", "autoSchedulingReason", "autoSchedulingUpdatedAt"} {
		if _, err := getUpdates(map[string]json.RawMessage{field: json.RawMessage(`"spoofed"`)}); err == nil {
			t.Fatalf("system-owned field %q was accepted", field)
		}
	}
}
