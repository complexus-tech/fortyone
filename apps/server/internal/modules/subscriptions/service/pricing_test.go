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
