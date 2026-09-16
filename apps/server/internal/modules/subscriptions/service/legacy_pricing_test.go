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

type legacyPricingRepository struct{ cancellationRepositoryStub }

func (legacyPricingRepository) HasActiveSubscriptionByWorkspaceID(context.Context, uuid.UUID) (bool, error) {
	return true, nil
}
func (legacyPricingRepository) GetWorkspaceUserCount(context.Context, uuid.UUID) (int, error) {
	return 6, nil
}

func TestSeatAdditionsKeepLegacyPrice(t *testing.T) {
	for _, tc := range []struct {
		key, interval string
		amount        int64
	}{
		{"pro_monthly", "month", 700}, {"pro_yearly", "year", 6720},
		{"business_monthly", "month", 1000}, {"business_yearly", "year", 9600},
	} {
		t.Run(tc.key, func(t *testing.T) {
			writes := 0
			provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/v1/subscriptions/sub_legacy", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodPost {
					writes++
					require.NoError(t, r.ParseForm())
					require.Equal(t, "si_legacy", r.Form.Get("items[0][id]"))
					require.Equal(t, "6", r.Form.Get("items[0][quantity]"))
					// Neither a price ID nor inline price data may accompany a seat change.
					for key := range r.Form {
						require.NotContains(t, key, "price")
					}
				}
				_, err := fmt.Fprintf(w, `{"id":"sub_legacy","items":{"data":[{"id":"si_legacy","quantity":6,"price":{"id":"price_legacy","lookup_key":null,"unit_amount":%d,"currency":"usd","billing_scheme":"per_unit","metadata":{"fortyone_lookup_key":%q},"recurring":{"interval":%q,"interval_count":1}}}]}}`, tc.amount, tc.key, tc.interval)
				require.NoError(t, err)
			}))
			defer provider.Close()
			subscription := CoreWorkspaceSubscription{WorkspaceID: uuid.New(), StripeSubscriptionID: stripe.String("sub_legacy"), StripeSubscriptionItemID: stripe.String("si_legacy"), SeatCount: 5}
			service := &Service{
				repo:         legacyPricingRepository{cancellationRepositoryStub{subscription: subscription}},
				stripeClient: client.New("sk_test_fake", stripe.NewBackendsWithConfig(&stripe.BackendConfig{URL: stripe.String(provider.URL), HTTPClient: provider.Client(), MaxNetworkRetries: stripe.Int64(0)})),
				log:          logger.NewWithText(io.Discard, slog.LevelError, "legacy-price-test"),
			}
			require.NoError(t, service.UpdateSubscriptionSeats(t.Context(), subscription.WorkspaceID))
			require.Equal(t, 1, writes)
			price, err := service.GetSubscriptionPrice(t.Context(), subscription)
			require.NoError(t, err)
			require.Equal(t, tc.amount, price.UnitAmount)
			require.Equal(t, tc.interval, price.Interval)
			require.Equal(t, 1, writes, "price reads must never mutate the subscription")
		})
	}
}
