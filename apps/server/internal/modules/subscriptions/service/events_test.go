package subscriptions

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	subscriptionsdomain "github.com/complexus-tech/projects-api/internal/modules/subscriptions/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/client"
)

type eventRepositoryStub struct {
	Repository

	workspaceID uuid.UUID
	snapshot    subscriptionsdomain.SubscriptionSnapshot
	cursor      subscriptionsdomain.StripeEventCursor
}

func (r *eventRepositoryStub) UpsertStripeSubscription(
	_ context.Context,
	workspaceID uuid.UUID,
	snapshot subscriptionsdomain.SubscriptionSnapshot,
	cursor subscriptionsdomain.StripeEventCursor,
) (subscriptionsdomain.SubscriptionMutation, error) {
	r.workspaceID = workspaceID
	r.snapshot = snapshot
	r.cursor = cursor
	return subscriptionsdomain.SubscriptionMutation{WorkspaceID: workspaceID, Applied: true}, nil
}

func TestSubscriptionCreatedReconcilesCurrentProviderSnapshotAtCreatedPriority(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	provider := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/v1/subscriptions/sub_created":
			_, _ = io.WriteString(response, `{
				"id":"sub_created","object":"subscription","status":"past_due",
				"customer":{"id":"cus_current","object":"customer"},
				"items":{"object":"list","data":[{
					"id":"si_current","object":"subscription_item","quantity":7,
					"current_period_end":1788000000,
					"price":{"id":"price_business_yearly","object":"price","lookup_key":"business_yearly","recurring":{"interval":"year"}}
				}]}
			}`)
		case "/v1/customers/cus_current":
			_, _ = fmt.Fprintf(response, `{"id":"cus_current","object":"customer","metadata":{"workspace_id":%q}}`, workspaceID.String())
		default:
			http.NotFound(response, request)
		}
	}))
	t.Cleanup(provider.Close)

	backends := stripe.NewBackendsWithConfig(&stripe.BackendConfig{
		URL:               stripe.String(provider.URL),
		HTTPClient:        provider.Client(),
		MaxNetworkRetries: stripe.Int64(0),
		LeveledLogger:     &stripe.LeveledLogger{Level: stripe.LevelNull},
	})
	repository := &eventRepositoryStub{}
	service := &Service{
		repo:         repository,
		stripeClient: client.New("sk_test_subscriptions", backends),
		log:          logger.NewWithText(io.Discard, slog.LevelError, "subscriptions-event-test"),
	}

	eventCreatedAt := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	event := stripe.Event{
		ID:      "evt_created",
		Created: eventCreatedAt.Unix(),
		Data: &stripe.EventData{Raw: []byte(`{
			"id":"sub_created","object":"subscription","status":"active",
			"customer":{"id":"cus_delivered","object":"customer"},
			"items":{"object":"list","data":[{"id":"si_delivered","quantity":1}]}
		}`)},
	}

	outcome, err := service.handleSubscriptionCreated(t.Context(), event)
	if err != nil {
		t.Fatalf("handle subscription.created: %v", err)
	}
	if outcome.Result != WebhookResultHandled || outcome.WorkspaceID == nil || *outcome.WorkspaceID != workspaceID {
		t.Fatalf("outcome = %+v", outcome)
	}
	if repository.workspaceID != workspaceID {
		t.Fatalf("bound workspace = %s, want %s", repository.workspaceID, workspaceID)
	}
	if repository.snapshot.StripeCustomerID != "cus_current" ||
		repository.snapshot.Status != StatusPastDue ||
		repository.snapshot.Tier != TierBusiness ||
		repository.snapshot.SeatCount != 7 {
		t.Fatalf("provider snapshot = %+v", repository.snapshot)
	}
	if repository.cursor.Priority != subscriptionsdomain.StripeEventPriorityCreated ||
		repository.cursor.EventID != event.ID ||
		!repository.cursor.CreatedAt.Equal(eventCreatedAt) {
		t.Fatalf("created cursor = %+v", repository.cursor)
	}
}
