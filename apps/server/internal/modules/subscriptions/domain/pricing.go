package subscriptionsdomain

// StripePriceLookupKeyMetadata preserves catalog identity after a lookup key is
// transferred to a replacement Stripe Price. Existing subscriptions keep the old Price.
const StripePriceLookupKeyMetadata = "fortyone_lookup_key"
