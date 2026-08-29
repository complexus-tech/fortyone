package webhooks

import (
	"context"
	"errors"
	"net/http"
)

var (
	ErrNotConfigured           = errors.New("webhook gateway is not configured")
	ErrRuntimeNotFound         = errors.New("webhook provider runtime is not configured")
	ErrInvalidRequest          = errors.New("webhook request is invalid")
	ErrPayloadTooLarge         = errors.New("webhook payload exceeds the configured limit")
	ErrHeadersTooLarge         = errors.New("webhook headers exceed the configured limit")
	ErrVerificationFailed      = errors.New("webhook verification failed")
	ErrVerificationUnavailable = errors.New("webhook verification dependency is unavailable")
	ErrUnauthenticated         = errors.New("webhook authentication failed")
	ErrReplay                  = errors.New("webhook delivery is outside the accepted replay window")
	ErrDeliveryIgnored         = errors.New("webhook delivery is intentionally ignored")
	ErrInvalidDelivery         = errors.New("verified webhook delivery is invalid")
	ErrPayloadProtection       = errors.New("webhook payload protection failed")
	ErrDeliveryConflict        = errors.New("webhook delivery identity conflicts with an existing receipt")
	ErrDispatchUnavailable     = errors.New("webhook delivery dispatch is unavailable")
	ErrNotFound                = errors.New("webhook delivery was not found")
	ErrLeaseBusy               = errors.New("webhook delivery is already being processed")
	ErrInvalidState            = errors.New("webhook delivery state is invalid")
)

// IngressHTTPStatus is the provider-neutral HTTP contract for durable webhook
// ingress. It keeps authentication failures distinct from transient verifier
// dependencies so providers retry outages instead of treating them as bad
// credentials.
func IngressHTTPStatus(err error) int {
	switch {
	case errors.Is(err, ErrUnauthenticated), errors.Is(err, ErrVerificationFailed), errors.Is(err, ErrReplay):
		return http.StatusUnauthorized
	case errors.Is(err, ErrPayloadTooLarge):
		return http.StatusRequestEntityTooLarge
	case errors.Is(err, ErrHeadersTooLarge):
		return http.StatusRequestHeaderFieldsTooLarge
	case errors.Is(err, ErrDeliveryConflict):
		return http.StatusConflict
	case errors.Is(err, ErrVerificationUnavailable),
		errors.Is(err, ErrDispatchUnavailable),
		errors.Is(err, ErrPayloadProtection),
		errors.Is(err, ErrNotConfigured),
		errors.Is(err, ErrRuntimeNotFound),
		errors.Is(err, context.Canceled),
		errors.Is(err, context.DeadlineExceeded):
		return http.StatusServiceUnavailable
	default:
		return http.StatusBadRequest
	}
}
