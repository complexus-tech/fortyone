package outboundwebhooksservice

import (
	"bytes"
	"errors"
	"net/http"
	"net/netip"
	"strings"
	"testing"
	"time"

	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/internal/platform/safehttp"
	"github.com/google/uuid"
)

func TestDispatcherSendsExactBodyWithRotationOverlapSignatures(t *testing.T) {
	t.Parallel()
	workspaceID, endpointID := uuid.New(), uuid.New()
	secrets := newTestSecretManager(t)
	_, previousEnvelope, err := secrets.Generate(workspaceID, endpointID, 1)
	if err != nil {
		t.Fatalf("generate previous secret: %v", err)
	}
	_, currentEnvelope, err := secrets.Generate(workspaceID, endpointID, 2)
	if err != nil {
		t.Fatalf("generate current secret: %v", err)
	}
	startedAt := time.Unix(1_700_000_000, 0).UTC()
	previousGeneration := 1
	overlapExpiresAt := startedAt.Add(time.Hour)
	body := []byte("{\"id\":\"exact\\r\\nbody\"}")
	repository := &deliveryRepositoryStub{delivery: outboundwebhooksdomain.ClaimedDelivery{
		ID: uuid.New(), WorkspaceID: workspaceID, EndpointID: endpointID, EndpointURL: "https://hooks.example.com/receive",
		SigningSecretEnvelope: currentEnvelope, SecretGeneration: 2,
		PreviousSecretEnvelope: &previousEnvelope, PreviousSecretGeneration: &previousGeneration,
		PreviousSecretExpiresAt: &overlapExpiresAt, PayloadBody: body, AttemptNumber: 1,
	}}
	httpClient := &httpClientStub{result: safehttp.Result{
		StatusCode: 202, ResolvedIP: netip.MustParseAddr("203.0.113.10"), Duration: 250 * time.Millisecond,
	}}
	dispatcher, err := newDispatcher(repository, secrets, httpClient,
		&testClock{values: []time.Time{startedAt, startedAt.Add(time.Second)}},
		&testIDs{values: []uuid.UUID{uuid.New(), uuid.New()}}, 30*time.Second, 8, 20)
	if err != nil {
		t.Fatalf("create dispatcher: %v", err)
	}
	worked, err := dispatcher.DispatchOne(t.Context())
	if err != nil || !worked {
		t.Fatalf("DispatchOne() = %v, %v", worked, err)
	}
	if !bytes.Equal(httpClient.body, body) {
		t.Fatalf("sent body = %q, want exact %q", httpClient.body, body)
	}
	if signatures := strings.Fields(httpClient.headers.Get("Webhook-Signature")); len(signatures) != 2 {
		t.Fatalf("Webhook-Signature = %q, want current and previous signatures", httpClient.headers.Get("Webhook-Signature"))
	}
	if len(repository.completed) != 1 || repository.completed[0].Outcome != outboundwebhooksdomain.AttemptSucceeded ||
		repository.completed[0].CountEndpointFailure || repository.completed[0].DisableEndpoint {
		t.Fatalf("completed attempt = %+v", repository.completed)
	}
}

func TestDispatcherTreatsVaultAvailabilityAsOperationalRetry(t *testing.T) {
	t.Parallel()
	for _, vaultErr := range []error{credentialvault.ErrNotConfigured, credentialvault.ErrUnknownKey} {
		t.Run(vaultErr.Error(), func(t *testing.T) {
			secrets, err := newSecretManager(failingSecretVault{err: vaultErr}, bytes.NewReader(make([]byte, 64)))
			if err != nil {
				t.Fatalf("create secret manager: %v", err)
			}
			startedAt := time.Unix(1_700_000_000, 0).UTC()
			repository := &deliveryRepositoryStub{delivery: testClaimedDelivery(1)}
			dispatcher, err := newDispatcher(repository, secrets, &httpClientStub{},
				&testClock{values: []time.Time{startedAt, startedAt.Add(time.Second)}},
				&testIDs{values: []uuid.UUID{uuid.New(), uuid.New()}}, 30*time.Second, 8, 20)
			if err != nil {
				t.Fatalf("create dispatcher: %v", err)
			}
			worked, err := dispatcher.DispatchOne(t.Context())
			if err != nil || !worked {
				t.Fatalf("DispatchOne() = %v, %v", worked, err)
			}
			attempt := repository.completed[0]
			if attempt.Outcome != outboundwebhooksdomain.AttemptRetryScheduled || attempt.NextAttemptAt == nil ||
				attempt.CountEndpointFailure || attempt.DisableEndpoint || attempt.ErrorCode != "secret_vault_unavailable" {
				t.Fatalf("operational vault attempt = %+v", attempt)
			}
		})
	}
}

func TestDispatcherDisablesEndpointForUnauthenticSecretEnvelope(t *testing.T) {
	t.Parallel()
	secrets, err := newSecretManager(failingSecretVault{err: credentialvault.ErrAuthentication}, bytes.NewReader(make([]byte, 64)))
	if err != nil {
		t.Fatalf("create secret manager: %v", err)
	}
	startedAt := time.Unix(1_700_000_000, 0).UTC()
	repository := &deliveryRepositoryStub{delivery: testClaimedDelivery(1)}
	dispatcher, err := newDispatcher(repository, secrets, &httpClientStub{},
		&testClock{values: []time.Time{startedAt, startedAt.Add(time.Second)}},
		&testIDs{values: []uuid.UUID{uuid.New(), uuid.New()}}, 30*time.Second, 8, 20)
	if err != nil {
		t.Fatalf("create dispatcher: %v", err)
	}
	if _, err := dispatcher.DispatchOne(t.Context()); err != nil {
		t.Fatalf("DispatchOne() error = %v", err)
	}
	attempt := repository.completed[0]
	if attempt.Outcome != outboundwebhooksdomain.AttemptFailed || !attempt.DisableEndpoint ||
		!attempt.CountEndpointFailure || attempt.ErrorCode != "secret_invalid" {
		t.Fatalf("invalid secret attempt = %+v", attempt)
	}
}

func TestClassifyDeliveryPolicy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		result  safehttp.Result
		err     error
		attempt int
		outcome outboundwebhooksdomain.AttemptOutcome
		code    string
		disable bool
	}{
		{name: "success", result: safehttp.Result{StatusCode: 204}, attempt: 1, outcome: outboundwebhooksdomain.AttemptSucceeded},
		{name: "rate limited", result: safehttp.Result{StatusCode: http.StatusTooManyRequests}, attempt: 1, outcome: outboundwebhooksdomain.AttemptRetryScheduled, code: "http_429"},
		{name: "server error", result: safehttp.Result{StatusCode: 503}, attempt: 1, outcome: outboundwebhooksdomain.AttemptRetryScheduled, code: "http_5xx"},
		{name: "gone", result: safehttp.Result{StatusCode: http.StatusGone}, attempt: 1, outcome: outboundwebhooksdomain.AttemptFailed, code: "http_410", disable: true},
		{name: "bad request", result: safehttp.Result{StatusCode: 400}, attempt: 1, outcome: outboundwebhooksdomain.AttemptFailed, code: "http_400"},
		{name: "unsafe address", err: safehttp.ErrUnsafeAddress, attempt: 1, outcome: outboundwebhooksdomain.AttemptFailed, code: "unsafe_endpoint", disable: true},
		{name: "network", err: errors.New("connection reset"), attempt: 1, outcome: outboundwebhooksdomain.AttemptRetryScheduled, code: "network_error"},
		{name: "exhausted", result: safehttp.Result{StatusCode: 503}, attempt: 8, outcome: outboundwebhooksdomain.AttemptFailed, code: "attempts_exhausted"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := classifyDelivery(test.result, test.err, test.attempt, 8)
			if got.outcome != test.outcome || got.errorCode != test.code || got.disableEndpoint != test.disable {
				t.Fatalf("classifyDelivery() = %+v", got)
			}
			if got.outcome != outboundwebhooksdomain.AttemptSucceeded && !got.countEndpointFailure {
				t.Fatal("destination failure did not count against endpoint health")
			}
		})
	}
}

func TestDispatcherIdleQueueIsNotWork(t *testing.T) {
	t.Parallel()
	repository := &deliveryRepositoryStub{claimErr: outboundwebhooksdomain.ErrDeliveryNotFound}
	dispatcher, err := newDispatcher(repository, newTestSecretManager(t), &httpClientStub{},
		&testClock{values: []time.Time{time.Now()}}, &testIDs{values: []uuid.UUID{uuid.New()}}, 30*time.Second, 8, 20)
	if err != nil {
		t.Fatalf("create dispatcher: %v", err)
	}
	worked, err := dispatcher.DispatchOne(t.Context())
	if err != nil || worked || len(repository.completed) != 0 {
		t.Fatalf("DispatchOne() = %v, %v, attempts=%d", worked, err, len(repository.completed))
	}
}

func testClaimedDelivery(attempt int) outboundwebhooksdomain.ClaimedDelivery {
	return outboundwebhooksdomain.ClaimedDelivery{
		ID: uuid.New(), WorkspaceID: uuid.New(), EndpointID: uuid.New(), EndpointURL: "https://hooks.example.com/receive",
		SigningSecretEnvelope: "vault-envelope", SecretGeneration: 1, PayloadBody: []byte("{}"), AttemptNumber: attempt,
	}
}
