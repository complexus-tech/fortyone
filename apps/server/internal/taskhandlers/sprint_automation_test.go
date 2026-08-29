package taskhandlers

import (
	"context"
	"errors"
	"testing"
	"time"

	teamsettingsdomain "github.com/complexus-tech/projects-api/internal/modules/teamsettings/domain"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

func TestSprintAutomationHandlersDelegateBothPolicyPhases(t *testing.T) {
	store := &sprintAutomationStoreStub{
		team: teamsettingsdomain.SprintAutomationTeamRef{
			WorkspaceID: uuid.New(),
			TeamID:      uuid.New(),
		},
	}
	handlers := NewSprintAutomationHandlers(SprintAutomationHandlerDependencies{
		Log:   testTaskLogger(),
		Store: store,
	})

	require.NoError(t, handlers.HandleSprintAutoCreation(t.Context(), asynq.NewTask("test", nil)))
	require.NoError(t, handlers.HandleDisableInactiveAutomation(t.Context(), asynq.NewTask("test", nil)))
	require.Equal(t, 1, store.runTeamCalls)
	require.Equal(t, 1, store.disableCalls)
}

func TestSprintAutomationHandlersPreserveStoreFailures(t *testing.T) {
	sentinel := errors.New("sprint automation store unavailable")
	store := &sprintAutomationStoreStub{creationListErr: sentinel}
	handlers := NewSprintAutomationHandlers(SprintAutomationHandlerDependencies{
		Log:   testTaskLogger(),
		Store: store,
	})

	err := handlers.HandleSprintAutoCreation(t.Context(), asynq.NewTask("test", nil))
	require.ErrorIs(t, err, sentinel)

	store.creationListErr = nil
	store.inactivityListErr = sentinel
	err = handlers.HandleDisableInactiveAutomation(t.Context(), asynq.NewTask("test", nil))
	require.ErrorIs(t, err, sentinel)
}

func TestSprintAutomationHandlersRequireCoreDependencies(t *testing.T) {
	var handlers *SprintAutomationHandlers
	err := handlers.HandleSprintAutoCreation(t.Context(), nil)
	require.ErrorContains(t, err, "dependencies are required")

	handlers = NewSprintAutomationHandlers(SprintAutomationHandlerDependencies{Log: testTaskLogger()})
	err = handlers.HandleSprintAutoCreation(t.Context(), nil)
	require.ErrorContains(t, err, "sprint automation store is required")
}

type sprintAutomationStoreStub struct {
	team               teamsettingsdomain.SprintAutomationTeamRef
	creationPageServed bool
	inactivityServed   bool
	runTeamCalls       int
	disableCalls       int
	creationListErr    error
	inactivityListErr  error
}

func (store *sprintAutomationStoreStub) ListSprintAutomationTeams(
	context.Context,
	teamsettingsdomain.SprintAutomationQuery,
) ([]teamsettingsdomain.SprintAutomationTeamRef, error) {
	if store.creationListErr != nil {
		return nil, store.creationListErr
	}
	if store.creationPageServed || store.team.TeamID == uuid.Nil {
		return nil, nil
	}
	store.creationPageServed = true
	return []teamsettingsdomain.SprintAutomationTeamRef{store.team}, nil
}

func (store *sprintAutomationStoreStub) RunSprintAutomationForTeam(
	context.Context,
	teamsettingsdomain.SprintAutomationTeamRef,
	time.Time,
) (teamsettingsdomain.SprintAutomationRunResult, error) {
	store.runTeamCalls++
	return teamsettingsdomain.SprintAutomationRunResult{}, nil
}

func (store *sprintAutomationStoreStub) ListSprintAutomationInactivityCandidates(
	context.Context,
	teamsettingsdomain.SprintAutomationInactivityQuery,
) ([]teamsettingsdomain.SprintAutomationTeamRef, error) {
	if store.inactivityListErr != nil {
		return nil, store.inactivityListErr
	}
	if store.inactivityServed || store.team.TeamID == uuid.Nil {
		return nil, nil
	}
	store.inactivityServed = true
	return []teamsettingsdomain.SprintAutomationTeamRef{store.team}, nil
}

func (store *sprintAutomationStoreStub) DisableSprintAutomationIfInactive(
	context.Context,
	teamsettingsdomain.SprintAutomationInactivityEligibility,
) (bool, error) {
	store.disableCalls++
	return true, nil
}
