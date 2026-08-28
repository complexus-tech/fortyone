package taskhandlers

import (
	"context"
	"errors"
	"testing"
	"time"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

func TestUserLifecycleHandlersDelegateBothPolicyPhases(t *testing.T) {
	store := &userLifecycleStoreStub{}
	handlers := NewUserLifecycleHandlers(UserLifecycleHandlerDependencies{
		Log:    testTaskLogger(),
		Store:  store,
		Mailer: guidanceMailerStub{},
	})

	require.NoError(t, handlers.HandleUserInactivityWarning(t.Context(), asynq.NewTask("test", nil)))
	require.NoError(t, handlers.HandleUserDeactivation(t.Context(), asynq.NewTask("test", nil)))
	require.Equal(t, 1, store.warningListCalls)
	require.Equal(t, 1, store.deactivationCalls)
}

func TestUserLifecycleHandlersPreserveStoreFailures(t *testing.T) {
	sentinel := errors.New("user lifecycle store unavailable")
	store := &userLifecycleStoreStub{warningListErr: sentinel}
	handlers := NewUserLifecycleHandlers(UserLifecycleHandlerDependencies{
		Log:    testTaskLogger(),
		Store:  store,
		Mailer: guidanceMailerStub{},
	})

	err := handlers.HandleUserInactivityWarning(t.Context(), asynq.NewTask("test", nil))
	require.ErrorIs(t, err, sentinel)

	store.warningListErr = nil
	store.deactivationErr = sentinel
	err = handlers.HandleUserDeactivation(t.Context(), asynq.NewTask("test", nil))
	require.ErrorIs(t, err, sentinel)
}

func TestUserLifecycleHandlersRequireCoreDependencies(t *testing.T) {
	var handlers *UserLifecycleHandlers
	err := handlers.HandleUserDeactivation(t.Context(), nil)
	require.ErrorContains(t, err, "dependencies are required")

	handlers = NewUserLifecycleHandlers(UserLifecycleHandlerDependencies{Log: testTaskLogger()})
	err = handlers.HandleUserDeactivation(t.Context(), nil)
	require.ErrorContains(t, err, "user maintenance store is required")
}

type userLifecycleStoreStub struct {
	warningListCalls  int
	deactivationCalls int
	warningListErr    error
	deactivationErr   error
}

func (store *userLifecycleStoreStub) ListUserInactivityWarningCandidates(
	context.Context,
	usersdomain.InactivityWarningQuery,
) ([]usersdomain.InactivityWarningCandidate, error) {
	store.warningListCalls++
	return nil, store.warningListErr
}

func (*userLifecycleStoreStub) GetEligibleUserInactivityWarningCandidate(
	context.Context,
	usersdomain.InactivityWarningEligibility,
) (usersdomain.InactivityWarningCandidate, bool, error) {
	return usersdomain.InactivityWarningCandidate{}, false, nil
}

func (*userLifecycleStoreStub) RecordUserInactivityWarning(
	context.Context,
	usersdomain.InactivityWarningReceipt,
) (bool, error) {
	return true, nil
}

func (store *userLifecycleStoreStub) DeactivateInactiveUsers(
	context.Context,
	time.Time,
	time.Time,
	time.Time,
	int,
) (int64, error) {
	store.deactivationCalls++
	return 0, store.deactivationErr
}

var _ UserLifecycleStore = (*userLifecycleStoreStub)(nil)
