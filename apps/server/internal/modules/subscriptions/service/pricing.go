package subscriptions

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v82"
)

var supportedPaidPrices = map[string]SubscriptionTier{
	"pro_monthly":        TierPro,
	"pro_yearly":         TierPro,
	"business_monthly":   TierBusiness,
	"business_yearly":    TierBusiness,
	"enterprise_monthly": TierEnterprise,
	"enterprise_yearly":  TierEnterprise,
}

func supportedPaidLookupKey(value string) bool {
	_, ok := supportedPaidPrices[value]
	return ok
}

func subscriptionTierForLookupKey(value string) (SubscriptionTier, bool) {
	tier, ok := supportedPaidPrices[value]
	return tier, ok
}

func (s *Service) lookupStripePriceID(ctx context.Context, lookupKey string) (string, error) {
	params := &stripe.PriceListParams{}
	params.Context = ctx
	params.Filters.AddFilter("lookup_keys[]", "", lookupKey)
	params.Filters.AddFilter("limit", "", "1")

	prices := s.stripeClient.Prices.List(params)
	if prices.Next() {
		price := prices.Price()
		if price != nil && price.ID != "" {
			return price.ID, nil
		}
	}
	if err := prices.Err(); err != nil {
		return "", fmt.Errorf("%w: list price for lookup key %q: %v", ErrStripeOperationFailed, lookupKey, err)
	}
	return "", fmt.Errorf("no price found for lookup key: %s", lookupKey)
}
