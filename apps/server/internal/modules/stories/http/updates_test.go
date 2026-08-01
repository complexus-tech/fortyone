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
