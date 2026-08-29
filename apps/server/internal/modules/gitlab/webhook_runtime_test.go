package gitlab

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/codehost"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
)

type resolverStub struct {
	installation codehost.InstallationRef
	calls        int
}

func (resolver *resolverStub) ResolveGitLabWebhookInstallation(
	_ context.Context,
	_, _ string,
) (codehost.InstallationRef, error) {
	resolver.calls++
	return resolver.installation, nil
}

type queueStub struct{ tasks []WebhookTask }

func (queue *queueStub) EnqueueGitLabWebhook(_ context.Context, task WebhookTask) error {
	queue.tasks = append(queue.tasks, task)
	return nil
}

func TestGitLabWebhookVerifierAuthenticatesExactBytesBeforeParsing(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	resolver := &resolverStub{installation: testInstallation()}
	queue := &queueStub{}
	runtime := newTestWebhookRuntime(t, resolver, queue, now)
	verifier := runtime.Registration.Verifier
	body := []byte(`{"project":{"id":42}}`)
	request := signedGitLabRequest(body, now, "delivery-1")
	request.Body = append(request.Body, ' ')
	if _, err := verifier.Verify(t.Context(), request); err != webhooks.ErrUnauthenticated {
		t.Fatalf("Verify(tampered) error = %v, want ErrUnauthenticated", err)
	}
	if resolver.calls != 0 {
		t.Fatalf("resolver calls before signature verification = %d", resolver.calls)
	}

	request = signedGitLabRequest(body, now, "delivery-1")
	verified, err := verifier.Verify(t.Context(), request)
	if err != nil {
		t.Fatalf("Verify(valid) error = %v", err)
	}
	if verified.DeliveryID != "delivery-1" || verified.EventType != "Issue Hook" ||
		verified.InstallationGeneration != resolver.installation.Generation || resolver.calls != 1 {
		t.Fatalf("Verify(valid) = %#v, resolver calls = %d", verified, resolver.calls)
	}
}

func TestGitLabWebhookVerifierRejectsReplayAndQueueCarriesOnlyIdentity(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	resolver := &resolverStub{installation: testInstallation()}
	queue := &queueStub{}
	runtime := newTestWebhookRuntime(t, resolver, queue, now)
	request := signedGitLabRequest([]byte(`{"project":{"id":42}}`), now.Add(-10*time.Minute), "delivery-old")
	if _, err := runtime.Registration.Verifier.Verify(t.Context(), request); err != webhooks.ErrReplay {
		t.Fatalf("Verify(replay) error = %v, want ErrReplay", err)
	}

	task := webhooks.Task{Provider: ProviderKey, InboxID: testInstallation().InstallationID}
	if err := runtime.Registration.Dispatcher.Enqueue(t.Context(), task); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}
	if len(queue.tasks) != 1 || queue.tasks[0].InboxID != task.InboxID {
		t.Fatalf("queued tasks = %#v", queue.tasks)
	}
}

func TestGitLabWebhookVerifierBindsConfiguredInstanceBeforeResolution(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	resolver := &resolverStub{installation: testInstallation()}
	runtime := newTestWebhookRuntime(t, resolver, &queueStub{}, now)
	request := signedGitLabRequest([]byte(`{"project":{"id":42}}`), now, "delivery-instance")
	request.Headers["X-Gitlab-Instance"] = []string{"https://other-gitlab.example"}
	if _, err := runtime.Registration.Verifier.Verify(t.Context(), request); err != webhooks.ErrInvalidDelivery {
		t.Fatalf("Verify(instance mismatch) error = %v, want ErrInvalidDelivery", err)
	}
	if resolver.calls != 0 {
		t.Fatalf("resolver calls for mismatched instance = %d", resolver.calls)
	}
}

func newTestWebhookRuntime(
	t *testing.T,
	resolver WebhookInstallationResolver,
	queue WebhookQueue,
	now time.Time,
) WebhookRuntime {
	t.Helper()
	runtime, err := NewWebhookRuntime(resolver, queue, Config{
		BaseURL:              "https://gitlab.example",
		WebhookSigningToken:  "whsec_" + base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")),
		WebhookPayloadSecret: "abcdef0123456789abcdef0123456789",
		Now:                  func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("NewWebhookRuntime() error = %v", err)
	}
	return runtime
}

func signedGitLabRequest(body []byte, sentAt time.Time, deliveryID string) webhooks.SignedRequest {
	timestamp := fmt.Sprintf("%d", sentAt.Unix())
	key := []byte("0123456789abcdef0123456789abcdef")
	mac := hmac.New(sha256.New, key)
	_, _ = fmt.Fprintf(mac, "%s.%s.", deliveryID, timestamp)
	_, _ = mac.Write(body)
	signature := "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return webhooks.SignedRequest{
		Method: "POST",
		Headers: webhooks.Headers{
			"webhook-id":          {deliveryID},
			"webhook-timestamp":   {timestamp},
			"webhook-signature":   {signature},
			"Idempotency-Key":     {deliveryID},
			"X-Gitlab-Event":      {"Issue Hook"},
			"X-Gitlab-Instance":   {"https://gitlab.example"},
			"X-Gitlab-Event-UUID": {"44444444-4444-4444-8444-444444444444"},
		},
		Body: body,
	}
}
