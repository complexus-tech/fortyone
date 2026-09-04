package workspacebootstrap

import (
	"context"
	"errors"
	"fmt"
	"testing"

	subscriptions "github.com/complexus-tech/projects-api/internal/modules/subscriptions/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceCancellationAcceptsSettledBillingStates(t *testing.T) {
	t.Parallel()

	providerErr := errors.New("provider unavailable")
	tests := []struct {
		name    string
		err     error
		wantErr error
	}{
		{name: "cancellation scheduled"},
		{name: "no subscription", err: subscriptions.ErrNoActiveSubscriptionToChange},
		{name: "already canceled", err: subscriptions.ErrSubscriptionAlreadyCanceled},
		{name: "wrapped settled state", err: fmt.Errorf("subscription: %w", subscriptions.ErrSubscriptionAlreadyCanceled)},
		{name: "provider failure", err: providerErr, wantErr: providerErr},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			workspaceID := uuid.New()
			service := &subscriptionServiceStub{err: test.err}
			manager := subscriptionManager{service: service}

			require.ErrorIs(t, manager.CancelWorkspaceSubscription(t.Context(), workspaceID), test.wantErr)
			require.Equal(t, workspaceID, service.workspaceID)
		})
	}
}

type subscriptionServiceStub struct {
	workspaceID uuid.UUID
	err         error
}

func (stub *subscriptionServiceStub) CancelSubscription(_ context.Context, workspaceID uuid.UUID) error {
	stub.workspaceID = workspaceID
	return stub.err
}

func (*subscriptionServiceStub) UpdateSubscriptionSeats(context.Context, uuid.UUID) error {
	return nil
}
