package integrationrequestsadapter

import (
	"context"
	"testing"
	"time"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

func TestStoryCreatorMapsOnlyTheConsumerOwnedContract(t *testing.T) {
	t.Parallel()

	actorID, workspaceID, teamID, storyID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	statusID, reporterID, assigneeID := uuid.New(), uuid.New(), uuid.New()
	objectiveID, keyResultID, sprintID := uuid.New(), uuid.New(), uuid.New()
	description := "A provider-created request"
	estimateValue, duration, focus := int16(5), 120, 45
	startDate := time.Date(2026, time.August, 29, 8, 0, 0, 0, time.UTC)
	endDate := startDate.Add(24 * time.Hour)
	createdAt, updatedAt := startDate.Add(-time.Hour), startDate.Add(-time.Minute)
	labelIDs := []uuid.UUID{uuid.New(), uuid.New()}

	backend := &storyBackendStub{result: stories.CoreSingleStory{
		ID: storyID, SequenceID: 41, Team: teamID, TeamCode: "API", Title: "Typed boundary",
		Description: &description, Status: &statusID, Priority: "High", Assignee: &assigneeID,
		Reporter: &reporterID, EndDate: &endDate, CreatedAt: createdAt, UpdatedAt: updatedAt,
	}}
	creator := NewStoryCreator(backend)
	got, err := creator.CreateForIntegrationRequest(context.Background(), actorID, workspaceID, integrationrequests.NewStory{
		Title: "Typed boundary", Description: &description, StatusID: &statusID, ReporterID: &reporterID,
		AssigneeID: &assigneeID, TeamID: teamID, Priority: "High", EstimateValue: &estimateValue,
		EstimatedDurationMinutes: &duration, MinimumFocusBlockMinutes: &focus, ObjectiveID: &objectiveID,
		KeyResultID: &keyResultID, SprintID: &sprintID, StartDate: &startDate, EndDate: &endDate,
		LabelIDs: labelIDs, CreationKey: "integration-request:workspace:request",
	})
	if err != nil {
		t.Fatalf("CreateForIntegrationRequest() error = %v", err)
	}
	if backend.actorID != actorID || backend.workspaceID != workspaceID {
		t.Fatalf("backend scope = %s/%s, want %s/%s", backend.actorID, backend.workspaceID, actorID, workspaceID)
	}
	if backend.input.Team != teamID || backend.input.Status != &statusID || backend.input.Objective != &objectiveID || backend.input.KeyResult != &keyResultID || backend.input.Sprint != &sprintID {
		t.Fatalf("mapped story input = %#v", backend.input)
	}
	if backend.input.CreationKey == nil || *backend.input.CreationKey != "integration-request:workspace:request" {
		t.Fatalf("creation key = %#v", backend.input.CreationKey)
	}
	labelIDs[0] = uuid.New()
	if backend.input.LabelIDs[0] == labelIDs[0] {
		t.Fatal("adapter retained the caller's mutable label slice")
	}
	if got.ID != storyID || got.TeamID != teamID || got.TeamCode != "API" || got.SequenceID != 41 || got.ReporterID != &reporterID {
		t.Fatalf("mapped story result = %#v", got)
	}
}

type storyBackendStub struct {
	actorID     uuid.UUID
	workspaceID uuid.UUID
	input       stories.CoreNewStory
	result      stories.CoreSingleStory
}

func (stub *storyBackendStub) CreateExternalUserAction(
	_ context.Context,
	actorID uuid.UUID,
	input stories.CoreNewStory,
	workspaceID uuid.UUID,
) (stories.CoreSingleStory, error) {
	stub.actorID = actorID
	stub.workspaceID = workspaceID
	stub.input = input
	return stub.result, nil
}
