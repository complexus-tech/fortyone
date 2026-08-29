package subscriptionsdomain

import "errors"

var (
	ErrSubscriptionNotFound       = errors.New("subscription not found")
	ErrProviderIdentityConflict   = errors.New("stripe identity is already bound to another workspace")
	ErrWebhookEventClaimLost      = errors.New("stripe webhook event claim was lost")
	ErrWebhookEventTypeConflict   = errors.New("stripe webhook event type conflicts with the recorded delivery")
	ErrInvalidStripeEventIdentity = errors.New("invalid Stripe event identity")
	ErrInvalidSeatCount           = errors.New("subscription seat count is outside the supported range")
)
