package subscriptions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/client"
)

type paymentAlertRepositoryStub struct {
	Repository
	invoice CoreSubscriptionInvoice
	err     error
}

func (s *paymentAlertRepositoryStub) UpsertStripeInvoice(_ context.Context, _ string, invoice CoreSubscriptionInvoice) error {
	s.invoice = invoice
	return s.err
}

func TestInternalPaymentAlertsRequireSuccessfulLiveNonzeroCollection(t *testing.T) {
	workspaceID := uuid.New()
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/customers/cus_alert", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"id":"cus_alert","object":"customer","name":"Buyer","email":"buyer@example.test","metadata":{"workspace_id":%q}}`, workspaceID.String())
	}))
	defer provider.Close()
	backends := stripe.NewBackendsWithConfig(&stripe.BackendConfig{URL: stripe.String(provider.URL), HTTPClient: provider.Client(), MaxNetworkRetries: stripe.Int64(0), LeveledLogger: &stripe.LeveledLogger{Level: stripe.LevelNull}})
	for _, tc := range []struct {
		name, eventType      string
		live                 bool
		amount               int64
		wantAlert, failStore bool
	}{
		{"initial payment", "invoice.payment_succeeded", true, 4900, true, false},
		{"renewal", "invoice.payment_succeeded", true, 9800, true, false},
		{"manual settlement", "invoice.paid", true, 4900, false, false},
		{"free invoice", "invoice.payment_succeeded", true, 0, false, false},
		{"test payment", "invoice.payment_succeeded", false, 4900, false, false},
		{"outbox write failure", "invoice.payment_succeeded", true, 4900, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &paymentAlertRepositoryStub{}
			if tc.failStore {
				repo.err = errors.New("database unavailable")
			}
			service := &Service{repo: repo, stripeClient: client.New("test-key", backends)}
			payload, err := json.Marshal(map[string]any{"id": "in_alert", "object": "invoice", "customer": "cus_alert", "amount_paid": tc.amount, "currency": "usd", "status": "paid", "created": 1789293600, "status_transitions": map[string]any{"paid_at": 1789293700}})
			require.NoError(t, err)
			event := stripe.Event{ID: "evt_alert", Type: stripe.EventType(tc.eventType), Livemode: tc.live, Data: &stripe.EventData{Raw: payload}}
			_, err = (serviceWebhookEventProcessor{service: service}).ProcessWebhookEvent(t.Context(), event)
			if tc.failStore {
				require.ErrorIs(t, err, repo.err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, workspaceID, repo.invoice.WorkspaceID)
			if tc.wantAlert {
				require.NotNil(t, repo.invoice.CollectedPayment)
				require.Equal(t, tc.amount, repo.invoice.CollectedPayment.AmountMinor)
				require.Equal(t, "usd", repo.invoice.CollectedPayment.Currency)
				require.Equal(t, "buyer@example.test", repo.invoice.CollectedPayment.Email)
			} else {
				require.Nil(t, repo.invoice.CollectedPayment)
			}
		})
	}
}
