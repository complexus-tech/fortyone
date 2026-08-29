package tasks

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCalendarWorkspaceScheduleBatchContract(t *testing.T) {
	if CalendarWorkspaceScheduleBatchDelay != 15*time.Minute {
		t.Fatalf("workspace schedule batch delay = %s, want 15m", CalendarWorkspaceScheduleBatchDelay)
	}

	workspaceID := uuid.New()
	body, err := json.Marshal(CalendarWorkspaceScheduleBatchPayload{WorkspaceID: workspaceID})
	if err != nil {
		t.Fatalf("marshal workspace schedule batch payload: %v", err)
	}
	var decoded CalendarWorkspaceScheduleBatchPayload
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("decode workspace schedule batch payload: %v", err)
	}
	if decoded.WorkspaceID != workspaceID {
		t.Fatalf("workspace ID = %s, want %s", decoded.WorkspaceID, workspaceID)
	}
}
