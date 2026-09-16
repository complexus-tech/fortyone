package subscriptions

import (
	"errors"
	"testing"

	"github.com/stripe/stripe-go/v82"
)

func TestSubscriptionTierCatalogFailsClosed(t *testing.T) {
	t.Parallel()

	tier, ok := subscriptionTierForLookupKey("business_yearly")
	if !ok || tier != TierBusiness {
		t.Fatalf("business_yearly tier = %q, %t", tier, ok)
	}
	if _, ok := subscriptionTierForLookupKey("internal_unpublished_price"); ok {
		t.Fatal("unpublished price unexpectedly mapped to a subscription tier")
	}

	_, _, _, _, _, err := stripeSubscriptionDetails(&stripe.Subscription{
		ID: "sub_unknown_price",
		Items: &stripe.SubscriptionItemList{Data: []*stripe.SubscriptionItem{{
			ID:       "si_unknown_price",
			Quantity: 1,
			Price:    &stripe.Price{ID: "price_unknown", LookupKey: "internal_unpublished_price"},
		}}},
	})
	if !errors.Is(err, ErrInvalidSubscription) {
		t.Fatalf("unknown provider price error = %v", err)
	}
}

func TestHistoricalPriceRetainsTierAfterLookupTransfer(t *testing.T) {
	for _, tc := range []struct {
		name, lookup, metadata string
		want                   SubscriptionTier
		valid                  bool
	}{
		{"historical pro", "", "pro_monthly", TierPro, true},
		{"historical business", "", "business_yearly", TierBusiness, true},
		{"unknown metadata", "", "other", TierFree, false},
		{"unknown lookup does not fall back", "unknown", "pro_monthly", TierFree, false},
		{"current lookup takes priority", "business_yearly", "pro_monthly", TierBusiness, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, tier, _, _, err := stripeSubscriptionDetails(&stripe.Subscription{Items: &stripe.SubscriptionItemList{Data: []*stripe.SubscriptionItem{{ID: "si_existing", Quantity: 3, Price: &stripe.Price{ID: "price_existing", LookupKey: tc.lookup, Metadata: map[string]string{"fortyone_lookup_key": tc.metadata}}}}}})
			if (err == nil) != tc.valid || tier != tc.want {
				t.Fatalf("tier=%s error=%v", tier, err)
			}
		})
	}
}
