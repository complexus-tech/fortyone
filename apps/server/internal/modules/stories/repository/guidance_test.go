package storiesrepository

import (
	"context"
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type storyGuidanceQueriesStub struct {
	storyreadsql.Querier
	recipientParams storyreadsql.ListOverdueStoryGuidanceRecipientsParams
	recipients      []storyreadsql.ListOverdueStoryGuidanceRecipientsRow
	itemParams      storyreadsql.ListOverdueStoryGuidanceItemsParams
	items           []storyreadsql.ListOverdueStoryGuidanceItemsRow
}

func (stub *storyGuidanceQueriesStub) ListOverdueStoryGuidanceRecipients(
	_ context.Context,
	params storyreadsql.ListOverdueStoryGuidanceRecipientsParams,
) ([]storyreadsql.ListOverdueStoryGuidanceRecipientsRow, error) {
	stub.recipientParams = params
	return stub.recipients, nil
}

func (stub *storyGuidanceQueriesStub) ListOverdueStoryGuidanceItems(
	_ context.Context,
	params storyreadsql.ListOverdueStoryGuidanceItemsParams,
) ([]storyreadsql.ListOverdueStoryGuidanceItemsRow, error) {
	stub.itemParams = params
	return stub.items, nil
}

func TestListOverdueStoryGuidanceRecipientsMapsCursorAndDomainValues(t *testing.T) {
	t.Parallel()

	assigneeID := uuid.New()
	workspaceID := uuid.New()
	asOfInput := time.Date(2026, time.August, 29, 0, 30, 0, 0, time.FixedZone("test", 2*60*60))
	asOfDate := time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC)
	stub := &storyGuidanceQueriesStub{recipients: []storyreadsql.ListOverdueStoryGuidanceRecipientsRow{{
		AssigneeID: &assigneeID, AssigneeEmail: "assignee@example.com", AssigneeName: "Story assignee",
		WorkspaceID: workspaceID, WorkspaceName: "Product", WorkspaceSlug: "product", EmailEnabled: true,
	}}}
	repository := &repo{reads: stub}

	recipients, err := repository.ListOverdueStoryGuidanceRecipients(
		context.Background(),
		asOfInput,
		&storydomain.OverdueGuidanceCursor{AssigneeID: assigneeID, WorkspaceID: workspaceID},
		100,
	)
	require.NoError(t, err)
	require.Equal(t, asOfDate, stub.recipientParams.AsOf)
	require.Equal(t, int32(100), stub.recipientParams.ResultLimit)
	require.True(t, stub.recipientParams.HasCursor)
	require.Equal(t, assigneeID, *stub.recipientParams.AfterAssigneeID)
	require.Equal(t, workspaceID, stub.recipientParams.AfterWorkspaceID)
	require.Equal(t, []storydomain.OverdueGuidanceRecipient{{
		AssigneeID: assigneeID, AssigneeEmail: "assignee@example.com", AssigneeName: "Story assignee",
		WorkspaceID: workspaceID, WorkspaceName: "Product", WorkspaceSlug: "product", EmailEnabled: true,
	}}, recipients)
}

func TestListOverdueStoryGuidanceItemsMapsTypedRow(t *testing.T) {
	t.Parallel()

	storyID := uuid.New()
	assigneeID := uuid.New()
	workspaceID := uuid.New()
	teamID := uuid.New()
	sequenceID := int32(42)
	statusCategory := "started"
	endDate := time.Date(2026, time.August, 28, 0, 0, 0, 0, time.FixedZone("test", 2*60*60))
	asOfInput := time.Date(2026, time.August, 29, 0, 30, 0, 0, time.FixedZone("test", 2*60*60))
	asOfDate := time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC)
	stub := &storyGuidanceQueriesStub{items: []storyreadsql.ListOverdueStoryGuidanceItemsRow{{
		ID: storyID, SequenceID: &sequenceID, Title: "Prepare launch metrics", EndDate: &endDate,
		AssigneeID: &assigneeID, WorkspaceID: workspaceID, TeamID: teamID,
		AssigneeEmail: "assignee@example.com", AssigneeName: "Story assignee",
		WorkspaceName: "Product", WorkspaceSlug: "product", TeamName: "Product", TeamCode: "PRD",
		StatusName: "In progress", StatusCategory: &statusCategory, DeadlineStatus: "due_today",
	}}}
	repository := &repo{reads: stub}

	items, err := repository.ListOverdueStoryGuidanceItems(context.Background(), asOfInput, assigneeID, workspaceID)
	require.NoError(t, err)
	require.Equal(t, asOfDate, stub.itemParams.AsOf)
	require.Equal(t, assigneeID, *stub.itemParams.AssigneeID)
	require.Equal(t, workspaceID, stub.itemParams.WorkspaceID)
	require.Len(t, items, 1)
	require.Equal(t, storyID, items[0].ID)
	require.Equal(t, 42, items[0].SequenceID)
	require.Equal(t, endDate.UTC(), items[0].EndDate)
	require.True(t, items[0].EmailEnabled)
}
