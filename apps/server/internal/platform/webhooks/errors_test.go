package webhooks

import (
	"errors"
	"net/http"
	"testing"
)

func TestIngressHTTPStatusSeparatesAuthenticationFromAvailability(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "invalid signature", err: ErrUnauthenticated, status: http.StatusUnauthorized},
		{name: "safe verification failure", err: ErrVerificationFailed, status: http.StatusUnauthorized},
		{name: "authorization repository unavailable", err: ErrVerificationUnavailable, status: http.StatusServiceUnavailable},
		{name: "wrapped authorization repository unavailable", err: errors.Join(errors.New("database unavailable"), ErrVerificationUnavailable), status: http.StatusServiceUnavailable},
		{name: "dispatch unavailable", err: ErrDispatchUnavailable, status: http.StatusServiceUnavailable},
		{name: "payload too large", err: ErrPayloadTooLarge, status: http.StatusRequestEntityTooLarge},
		{name: "identity conflict", err: ErrDeliveryConflict, status: http.StatusConflict},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if status := IngressHTTPStatus(test.err); status != test.status {
				t.Fatalf("IngressHTTPStatus() = %d, want %d", status, test.status)
			}
		})
	}
}
