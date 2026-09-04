package subscriptions

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/client"
)

func TestCancelSubscriptionHandlesProviderStates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		status            stripe.SubscriptionStatus
		cancelAtPeriodEnd bool
		updateFails       bool
		wantErr           error
		wantUpdates       int
	}{
		{name: "active", status: stripe.SubscriptionStatusActive, wantUpdates: 1},
		{name: "past due", status: stripe.SubscriptionStatusPastDue, wantUpdates: 1},
		{
			name: "already scheduled", status: stripe.SubscriptionStatusActive,
			cancelAtPeriodEnd: true, wantErr: ErrSubscriptionAlreadyCanceled,
		},
		{name: "canceled", status: stripe.SubscriptionStatusCanceled, wantErr: ErrSubscriptionAlreadyCanceled},
		{name: "expired incomplete", status: stripe.SubscriptionStatusIncompleteExpired, wantErr: ErrSubscriptionAlreadyCanceled},
		{
			name: "provider failure", status: stripe.SubscriptionStatusActive,
			updateFails: true, wantErr: ErrStripeOperationFailed, wantUpdates: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			workspaceID := uuid.New()
			updates := 0
			provider := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				response.Header().Set("Content-Type", "application/json")
				if request.URL.Path != "/v1/subscriptions/sub_workspace" {
					http.NotFound(response, request)
					return
				}
				if request.Method == http.MethodPost {
					updates++
					if err := request.ParseForm(); err != nil {
						t.Errorf("parse provider update: %v", err)
					}
					if request.PostForm.Get("cancel_at_period_end") != "true" {
						t.Errorf("cancellation must remain scheduled for period end: %v", request.PostForm)
					}
					if request.Header.Get("Idempotency-Key") != "subscription-cancel-at-period-end:sub_workspace" {
						t.Errorf("cancellation has no stable idempotency key")
					}
					if test.updateFails {
						response.WriteHeader(http.StatusBadGateway)
						_, _ = io.WriteString(response, `{"error":{"type":"api_error","message":"temporary provider failure"}}`)
						return
					}
				}
				_, _ = fmt.Fprintf(response, `{"id":"sub_workspace","object":"subscription","status":%q,"cancel_at_period_end":%t}`, test.status, test.cancelAtPeriodEnd)
			}))
			t.Cleanup(provider.Close)
			backends := stripe.NewBackendsWithConfig(&stripe.BackendConfig{
				URL: stripe.String(provider.URL), HTTPClient: provider.Client(),
				MaxNetworkRetries: stripe.Int64(0), LeveledLogger: &stripe.LeveledLogger{Level: stripe.LevelNull},
			})
			service := &Service{
				repo: cancellationRepositoryStub{subscription: CoreWorkspaceSubscription{
					WorkspaceID: workspaceID, StripeSubscriptionID: stripe.String("sub_workspace"),
				}},
				stripeClient: client.New("test-stripe-key", backends),
				log:          logger.NewWithText(io.Discard, slog.LevelError, "subscription-cancellation-test"),
			}

			require.ErrorIs(t, service.CancelSubscription(t.Context(), workspaceID), test.wantErr)
			require.Equal(t, test.wantUpdates, updates)
		})
	}
}

func TestCancelSubscriptionWithoutProviderBinding(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name         string
		subscription CoreWorkspaceSubscription
		err          error
	}{
		{name: "no subscription", err: ErrSubscriptionNotFound},
		{name: "no provider identity"},
		{name: "empty provider identity", subscription: CoreWorkspaceSubscription{StripeSubscriptionID: stripe.String("")}},
	} {
		t.Run(test.name, func(t *testing.T) {
			service := &Service{repo: cancellationRepositoryStub{subscription: test.subscription, err: test.err}}
			require.ErrorIs(t, service.CancelSubscription(t.Context(), uuid.New()), ErrNoActiveSubscriptionToChange)
		})
	}
}

type cancellationRepositoryStub struct {
	Repository
	subscription CoreWorkspaceSubscription
	err          error
}

func (stub cancellationRepositoryStub) GetSubscriptionByWorkspaceID(context.Context, uuid.UUID) (CoreWorkspaceSubscription, error) {
	return stub.subscription, stub.err
}
