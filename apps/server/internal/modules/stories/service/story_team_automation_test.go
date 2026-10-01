package stories

import (
	"encoding/json"
	"testing"
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func TestTeamAutomationPreservesHumanAuthorizationCASAndInternalDelivery(t *testing.T) {
	workspace, team, storyID, owner := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	version := time.Now().UTC()
	repository := &typedStoryMutationRepositoryStub{story: CoreSingleStory{ID: storyID, Workspace: workspace, Team: team, Title: "Task", Priority: "Low", UpdatedAt: version, EstimateScheme: DefaultEstimateScheme}}
	service := newTypedMutationService(repository)
	if err := service.UpdateTeamAutomationIfUnchanged(storyMutationActorContext(t, workspace, owner), storyID, workspace, version, map[string]any{"priority": "High"}, "Team rule: Triage"); err != nil {
		t.Fatal(err)
	}
	command := repository.updateCommand
	if command == nil || command.Scope.Actor.Kind != platformauth.PrincipalHumanUser || command.Scope.Actor.PrincipalID != owner || !command.ExpectedUpdatedAt.Equal(version) {
		t.Fatal("automation bypassed owner identity or compare-and-swap")
	}
	var payload struct {
		Delivery string `json:"_delivery"`
	}
	if err := json.Unmarshal(command.Event.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Delivery != string(mutationEventDeliveryInternalOnly) {
		t.Fatal("automation output could trigger more rules")
	}
	publish, err := shouldPublishStoryMutationEvent(command.Event.Payload)
	if err != nil || publish {
		t.Fatal("internal rule output was eligible for outbound delivery")
	}
}
