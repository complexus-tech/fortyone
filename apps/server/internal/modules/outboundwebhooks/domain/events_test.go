package outboundwebhooksdomain

import (
	"encoding/json"
	"testing"
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func TestEventCatalogReturnsCopy(t *testing.T) {
	t.Parallel()
	first := EventCatalog()
	first[0] = "tampered"
	second := EventCatalog()
	if second[0] != EventStoryCreated {
		t.Fatalf("EventCatalog() shared mutable state: %v", second)
	}
}

func TestPublishEventRequiresObjectPayloadAndMatchingActorWorkspace(t *testing.T) {
	t.Parallel()
	workspaceID := uuid.New()
	actor, err := platformauth.NewHumanActor(uuid.New()).WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("create actor: %v", err)
	}
	base := PublishEvent{
		ID: uuid.New(), WorkspaceID: workspaceID, Type: EventCommentCreated, SubjectID: uuid.New(),
		Actor: actor, Payload: json.RawMessage(`{"body":"hello"}`), OccurredAt: time.Now().UTC(),
	}
	if err := base.Validate(); err != nil {
		t.Fatalf("valid event rejected: %v", err)
	}
	base.Payload = json.RawMessage(`[]`)
	if err := base.Validate(); err == nil {
		t.Fatal("array payload accepted")
	}
	base.Payload = json.RawMessage(`{}`)
	base.WorkspaceID = uuid.New()
	if err := base.Validate(); err == nil {
		t.Fatal("actor workspace mismatch accepted")
	}
}
