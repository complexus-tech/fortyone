package outboundwebhooksdomain

import "errors"

var (
	ErrInvalidEventType      = errors.New("outbound webhook event type is invalid")
	ErrInvalidSubject        = errors.New("outbound webhook event subject is invalid")
	ErrInvalidEndpoint       = errors.New("outbound webhook endpoint is invalid")
	ErrInvalidSubscription   = errors.New("outbound webhook subscription is invalid")
	ErrInvalidPayload        = errors.New("outbound webhook payload is invalid")
	ErrEndpointConflict      = errors.New("outbound webhook endpoint changed concurrently")
	ErrEndpointNotFound      = errors.New("outbound webhook endpoint was not found")
	ErrDeliveryNotFound      = errors.New("outbound webhook delivery was not found")
	ErrDeliveryLeaseLost     = errors.New("outbound webhook delivery lease was lost")
	ErrEndpointDisabled      = errors.New("outbound webhook endpoint is disabled")
	ErrEndpointOwnerInactive = errors.New("outbound webhook endpoint owner is inactive")
)
