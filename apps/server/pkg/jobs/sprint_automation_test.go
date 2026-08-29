package jobs

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	teamsettingsdomain "github.com/complexus-tech/projects-api/internal/modules/teamsettings/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSprintAutomationUsesOneUTCClock(t *testing.T) {
	t.Parallel()

	location := time.FixedZone("CAT", 2*60*60)
	asOf := time.Date(2026, time.August, 28, 10, 15, 0, 0, location)
	team := sprintAutomationRef(1, 1)
	store := &sprintAutomationStoreStub{
		creationPages: [][]teamsettingsdomain.SprintAutomationTeamRef{{team}},
		creationOutcomes: map[uuid.UUID][]sprintAutomationRunOutcome{
			team.TeamID: {{result: teamsettingsdomain.SprintAutomationRunResult{Created: 2, Rescheduled: 1}}},
		},
	}

	err := processSprintAutoCreationAt(context.Background(), store, sprintAutomationTestLogger(), asOf)

	require.NoError(t, err)
	require.Equal(t, []teamsettingsdomain.SprintAutomationQuery{{BatchSize: sprintAutomationBatchSize}}, store.creationQueries)
	require.Equal(t, []sprintAutomationRunCall{{team: team, asOf: asOf.UTC()}}, store.creationCalls)
}

func TestSprintAutomationUsesStableKeysetPages(t *testing.T) {
	t.Parallel()

	firstPage := make([]teamsettingsdomain.SprintAutomationTeamRef, sprintAutomationBatchSize)
	for index := range firstPage {
		firstPage[index] = sprintAutomationRef(1, index+1)
	}
	last := firstPage[len(firstPage)-1]
	finalTeam := sprintAutomationRef(1, sprintAutomationBatchSize+1)
	store := &sprintAutomationStoreStub{
		creationPages: [][]teamsettingsdomain.SprintAutomationTeamRef{firstPage, {finalTeam}},
	}

	err := processSprintAutoCreationAt(
		context.Background(),
		store,
		sprintAutomationTestLogger(),
		time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC),
	)

	require.NoError(t, err)
	require.Len(t, store.creationQueries, 2)
	require.Equal(t, teamsettingsdomain.SprintAutomationCursor{}, store.creationQueries[0].Cursor)
	require.Equal(t, teamsettingsdomain.SprintAutomationCursor{
		WorkspaceID: last.WorkspaceID,
		TeamID:      last.TeamID,
		Valid:       true,
	}, store.creationQueries[1].Cursor)
	require.Equal(t, sprintAutomationBatchSize+1, len(store.creationCalls))
}

func TestSprintAutomationRetriesOnlyConcurrentTransactions(t *testing.T) {
	t.Parallel()

	team := sprintAutomationRef(1, 1)
	store := &sprintAutomationStoreStub{
		creationPages: [][]teamsettingsdomain.SprintAutomationTeamRef{{team}},
		creationOutcomes: map[uuid.UUID][]sprintAutomationRunOutcome{
			team.TeamID: {
				{err: teamsettingsdomain.ErrConcurrentUpdate},
				{err: teamsettingsdomain.ErrConcurrentUpdate},
				{result: teamsettingsdomain.SprintAutomationRunResult{Created: 1}},
			},
		},
	}
	asOf := time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC)

	err := processSprintAutoCreationAt(context.Background(), store, sprintAutomationTestLogger(), asOf)

	require.NoError(t, err)
	require.Len(t, store.creationCalls, sprintAutomationTeamMaxAttempts)
	for _, call := range store.creationCalls {
		require.Equal(t, asOf, call.asOf)
	}
}

func TestSprintAutomationStopsAtBoundedBacklog(t *testing.T) {
	store := &sprintAutomationStoreStub{endlessCreation: true}

	err := processSprintAutoCreationAt(
		context.Background(),
		store,
		sprintAutomationTestLogger(),
		time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC),
	)

	require.ErrorIs(t, err, errSprintAutomationBacklogRemaining)
	require.Len(t, store.creationQueries, sprintAutomationMaxBatches)
	require.Len(t, store.creationCalls, sprintAutomationBatchSize*sprintAutomationMaxBatches)
}

func TestSprintAutomationInactivityAdvancesPastRecentTeams(t *testing.T) {
	t.Parallel()

	firstPage := make([]teamsettingsdomain.SprintAutomationTeamRef, sprintAutomationBatchSize)
	for index := range firstPage {
		firstPage[index] = sprintAutomationRef(1, index+1)
	}
	last := firstPage[len(firstPage)-1]
	inactiveTeam := sprintAutomationRef(1, sprintAutomationBatchSize+1)
	store := &sprintAutomationStoreStub{
		inactivityPages: [][]teamsettingsdomain.SprintAutomationTeamRef{firstPage, {inactiveTeam}},
		disableOutcomes: map[uuid.UUID]sprintAutomationDisableOutcome{
			inactiveTeam.TeamID: {disabled: true},
		},
	}
	location := time.FixedZone("CAT", 2*60*60)
	asOf := time.Date(2026, time.August, 28, 10, 15, 0, 0, location)

	err := disableAutomationForInactiveTeamsAt(context.Background(), store, sprintAutomationTestLogger(), asOf)

	require.NoError(t, err)
	require.Len(t, store.inactivityQueries, 2)
	require.Equal(t, teamsettingsdomain.SprintAutomationCursor{
		WorkspaceID: last.WorkspaceID,
		TeamID:      last.TeamID,
		Valid:       true,
	}, store.inactivityQueries[1].Cursor)
	require.Len(t, store.disableCalls, sprintAutomationBatchSize+1)
	lastCall := store.disableCalls[len(store.disableCalls)-1]
	require.Equal(t, inactiveTeam.TeamID, lastCall.TeamID)
	require.Equal(t, asOf.UTC(), lastCall.DisabledAt)
	require.Equal(t, asOf.UTC().AddDate(0, 0, -sprintAutomationInactivityDays), lastCall.ActivityBefore)
	require.Equal(t, asOf.UTC().AddDate(0, 0, -sprintAutomationSettingsGraceDays), lastCall.SettingsUpdatedBefore)
	require.Equal(t, sprintAutomationDisabledReason, lastCall.Reason)
}

func TestSprintAutomationRejectsMissingDependencies(t *testing.T) {
	t.Parallel()

	asOf := time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC)
	log := sprintAutomationTestLogger()
	require.ErrorContains(t, processSprintAutoCreationAt(context.Background(), nil, log, asOf), "store")
	require.ErrorContains(t, processSprintAutoCreationAt(context.Background(), &sprintAutomationStoreStub{}, nil, asOf), "logger")
	require.ErrorContains(t, processSprintAutoCreationAt(context.Background(), &sprintAutomationStoreStub{}, log, time.Time{}), "as-of")
	require.ErrorContains(t, disableAutomationForInactiveTeamsAt(context.Background(), nil, log, asOf), "store")
}

type sprintAutomationRunOutcome struct {
	result teamsettingsdomain.SprintAutomationRunResult
	err    error
}

type sprintAutomationDisableOutcome struct {
	disabled bool
	err      error
}

type sprintAutomationRunCall struct {
	team teamsettingsdomain.SprintAutomationTeamRef
	asOf time.Time
}

type sprintAutomationStoreStub struct {
	creationPages    [][]teamsettingsdomain.SprintAutomationTeamRef
	creationQueries  []teamsettingsdomain.SprintAutomationQuery
	creationOutcomes map[uuid.UUID][]sprintAutomationRunOutcome
	creationCalls    []sprintAutomationRunCall
	endlessCreation  bool

	inactivityPages   [][]teamsettingsdomain.SprintAutomationTeamRef
	inactivityQueries []teamsettingsdomain.SprintAutomationInactivityQuery
	disableOutcomes   map[uuid.UUID]sprintAutomationDisableOutcome
	disableCalls      []teamsettingsdomain.SprintAutomationInactivityEligibility
}

func (stub *sprintAutomationStoreStub) ListSprintAutomationTeams(
	_ context.Context,
	query teamsettingsdomain.SprintAutomationQuery,
) ([]teamsettingsdomain.SprintAutomationTeamRef, error) {
	stub.creationQueries = append(stub.creationQueries, query)
	if stub.endlessCreation {
		page := make([]teamsettingsdomain.SprintAutomationTeamRef, sprintAutomationBatchSize)
		workspaceNumber := len(stub.creationQueries)
		for index := range page {
			page[index] = sprintAutomationRef(workspaceNumber, index+1)
		}
		return page, nil
	}
	if len(stub.creationPages) == 0 {
		return nil, nil
	}
	page := stub.creationPages[0]
	stub.creationPages = stub.creationPages[1:]
	return page, nil
}

func (stub *sprintAutomationStoreStub) RunSprintAutomationForTeam(
	_ context.Context,
	team teamsettingsdomain.SprintAutomationTeamRef,
	asOf time.Time,
) (teamsettingsdomain.SprintAutomationRunResult, error) {
	stub.creationCalls = append(stub.creationCalls, sprintAutomationRunCall{team: team, asOf: asOf})
	outcomes := stub.creationOutcomes[team.TeamID]
	if len(outcomes) == 0 {
		return teamsettingsdomain.SprintAutomationRunResult{}, nil
	}
	outcome := outcomes[0]
	stub.creationOutcomes[team.TeamID] = outcomes[1:]
	return outcome.result, outcome.err
}

func (stub *sprintAutomationStoreStub) ListSprintAutomationInactivityCandidates(
	_ context.Context,
	query teamsettingsdomain.SprintAutomationInactivityQuery,
) ([]teamsettingsdomain.SprintAutomationTeamRef, error) {
	stub.inactivityQueries = append(stub.inactivityQueries, query)
	if len(stub.inactivityPages) == 0 {
		return nil, nil
	}
	page := stub.inactivityPages[0]
	stub.inactivityPages = stub.inactivityPages[1:]
	return page, nil
}

func (stub *sprintAutomationStoreStub) DisableSprintAutomationIfInactive(
	_ context.Context,
	eligibility teamsettingsdomain.SprintAutomationInactivityEligibility,
) (bool, error) {
	stub.disableCalls = append(stub.disableCalls, eligibility)
	outcome := stub.disableOutcomes[eligibility.TeamID]
	return outcome.disabled, outcome.err
}

func sprintAutomationRef(workspaceNumber, teamNumber int) teamsettingsdomain.SprintAutomationTeamRef {
	return teamsettingsdomain.SprintAutomationTeamRef{
		WorkspaceID: uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0000-%012d", workspaceNumber)),
		TeamID:      uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0001-%012d", teamNumber)),
	}
}

func sprintAutomationTestLogger() *logger.Logger {
	return logger.NewWithText(io.Discard, slog.LevelDebug, "sprint-automation-test")
}

var _ SprintAutomationStore = (*sprintAutomationStoreStub)(nil)
