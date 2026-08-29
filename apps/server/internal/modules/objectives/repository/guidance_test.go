package objectivesrepository

import (
	"context"
	"testing"
	"time"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	objectivessql "github.com/complexus-tech/projects-api/internal/modules/objectives/repository/sqlc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type objectiveGuidanceQueriesStub struct {
	objectivessql.Querier
	recipientParams objectivessql.ListOverdueObjectiveGuidanceRecipientsParams
	recipients      []objectivessql.ListOverdueObjectiveGuidanceRecipientsRow
	itemParams      objectivessql.ListOverdueObjectiveGuidanceItemsParams
	items           []objectivessql.ListOverdueObjectiveGuidanceItemsRow
}

func (stub *objectiveGuidanceQueriesStub) ListOverdueObjectiveGuidanceRecipients(
	_ context.Context,
	params objectivessql.ListOverdueObjectiveGuidanceRecipientsParams,
) ([]objectivessql.ListOverdueObjectiveGuidanceRecipientsRow, error) {
	stub.recipientParams = params
	return stub.recipients, nil
}

func (stub *objectiveGuidanceQueriesStub) ListOverdueObjectiveGuidanceItems(
	_ context.Context,
	params objectivessql.ListOverdueObjectiveGuidanceItemsParams,
) ([]objectivessql.ListOverdueObjectiveGuidanceItemsRow, error) {
	stub.itemParams = params
	return stub.items, nil
}

func TestListOverdueObjectiveGuidanceRecipientsMapsCursorAndDomainValues(t *testing.T) {
	t.Parallel()

	leadUserID := uuid.New()
	workspaceID := uuid.New()
	asOfInput := time.Date(2026, time.August, 29, 0, 30, 0, 0, time.FixedZone("test", 2*60*60))
	asOfDate := time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC)
	stub := &objectiveGuidanceQueriesStub{recipients: []objectivessql.ListOverdueObjectiveGuidanceRecipientsRow{{
		LeadUserID:    &leadUserID,
		LeadEmail:     "lead@example.com",
		LeadName:      "Objective lead",
		WorkspaceID:   workspaceID,
		WorkspaceName: "Product",
		WorkspaceSlug: "product",
		EmailEnabled:  true,
	}}}
	repository := newWithQueries(stub)

	recipients, err := repository.ListOverdueObjectiveGuidanceRecipients(
		context.Background(),
		asOfInput,
		&objectivesdomain.OverdueGuidanceCursor{LeadUserID: leadUserID, WorkspaceID: workspaceID},
		100,
	)
	require.NoError(t, err)
	require.Equal(t, asOfDate, stub.recipientParams.AsOf)
	require.Equal(t, int32(100), stub.recipientParams.ResultLimit)
	require.True(t, stub.recipientParams.HasCursor)
	require.Equal(t, leadUserID, *stub.recipientParams.AfterLeadUserID)
	require.Equal(t, workspaceID, *stub.recipientParams.AfterWorkspaceID)
	require.Equal(t, []objectivesdomain.OverdueGuidanceRecipient{{
		LeadUserID:    leadUserID,
		LeadEmail:     "lead@example.com",
		LeadName:      "Objective lead",
		WorkspaceID:   workspaceID,
		WorkspaceName: "Product",
		WorkspaceSlug: "product",
		EmailEnabled:  true,
	}}, recipients)
}

func TestListOverdueObjectiveGuidanceItemsMapsTypedRow(t *testing.T) {
	t.Parallel()

	objectiveID := uuid.New()
	leadUserID := uuid.New()
	workspaceID := uuid.New()
	teamID := uuid.New()
	endDate := time.Date(2026, time.August, 28, 0, 0, 0, 0, time.FixedZone("test", 2*60*60))
	asOfInput := time.Date(2026, time.August, 29, 0, 30, 0, 0, time.FixedZone("test", 2*60*60))
	asOfDate := time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC)
	stub := &objectiveGuidanceQueriesStub{items: []objectivessql.ListOverdueObjectiveGuidanceItemsRow{{
		ObjectiveID: objectiveID, Name: "Protect launch readiness", EndDate: &endDate,
		LeadUserID: &leadUserID, WorkspaceID: &workspaceID, TeamID: &teamID,
		LeadEmail: "lead@example.com", LeadName: "Objective lead",
		WorkspaceName: "Product", WorkspaceSlug: "product",
		DeadlineStatus: "due_today", DaysDifference: 0, KeyResults: []byte(`[{"id":"test"}]`),
	}}}
	repository := newWithQueries(stub)

	items, err := repository.ListOverdueObjectiveGuidanceItems(context.Background(), asOfInput, leadUserID, workspaceID)
	require.NoError(t, err)
	require.Equal(t, asOfDate, stub.itemParams.AsOf)
	require.Equal(t, leadUserID, *stub.itemParams.LeadUserID)
	require.Equal(t, workspaceID, *stub.itemParams.WorkspaceID)
	require.Len(t, items, 1)
	require.Equal(t, objectiveID, items[0].ID)
	require.Equal(t, endDate.UTC(), items[0].EndDate)
	require.Equal(t, `[{"id":"test"}]`, items[0].KeyResults)
}
