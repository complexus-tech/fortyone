package githubhttp

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/complexus-tech/projects-api/pkg/web"
)

func TestWebhookIngressHTTPContractIsRetryableAndSanitized(t *testing.T) {
	t.Parallel()
	const sensitiveCause = "postgresql://user:password@database"
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{
			name:   "authentication failure",
			err:    fmt.Errorf("%w: %s", webhooks.ErrUnauthenticated, sensitiveCause),
			status: http.StatusUnauthorized,
		},
		{
			name:   "verification dependency outage",
			err:    fmt.Errorf("%w: %s", webhooks.ErrVerificationUnavailable, sensitiveCause),
			status: http.StatusServiceUnavailable,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			recorder := httptest.NewRecorder()
			if err := web.RespondError(context.Background(), recorder, test.err, webhooks.IngressHTTPStatus(test.err)); err != nil {
				t.Fatalf("RespondError() error = %v", err)
			}
			if recorder.Code != test.status {
				t.Fatalf("status = %d, want %d", recorder.Code, test.status)
			}
			if strings.Contains(recorder.Body.String(), sensitiveCause) {
				t.Fatalf("response disclosed verifier cause: %s", recorder.Body.String())
			}
		})
	}
}
