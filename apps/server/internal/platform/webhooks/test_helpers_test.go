package webhooks

import (
	"context"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/google/uuid"
)

const testProvider integrations.ProviderKey = "slack"

var testNow = time.Date(2026, time.August, 28, 10, 30, 0, 0, time.UTC)

type verifierFunc func(context.Context, SignedRequest) (VerifiedDelivery, error)

func (function verifierFunc) Verify(ctx context.Context, request SignedRequest) (VerifiedDelivery, error) {
	return function(ctx, request)
}

type protectorFunc func(context.Context, PayloadBinding, []byte) (string, error)

func (function protectorFunc) Seal(ctx context.Context, binding PayloadBinding, payload []byte) (string, error) {
	return function(ctx, binding, payload)
}

type dispatcherFunc func(context.Context, Task) error

func (function dispatcherFunc) Enqueue(ctx context.Context, task Task) error {
	return function(ctx, task)
}

type inboxStub struct {
	register      func(context.Context, Envelope, string, time.Time) (Record, bool, error)
	markQueued    func(context.Context, uuid.UUID, time.Time) error
	claim         func(context.Context, integrations.ProviderKey, RecoveryPolicy, time.Time) ([]Record, error)
	release       func(context.Context, uuid.UUID, int32, time.Time) error
	getByID       func(context.Context, uuid.UUID) (Record, error)
	getByExternal func(context.Context, integrations.ProviderKey, string, string) (Record, error)
	start         func(context.Context, uuid.UUID, time.Time, time.Duration) (Record, bool, error)
	complete      func(context.Context, uuid.UUID, Status, string, time.Time) error
	expire        func(context.Context, time.Time, int32) ([]uuid.UUID, error)
}

func (stub *inboxStub) Register(ctx context.Context, envelope Envelope, payload string, expires time.Time) (Record, bool, error) {
	return stub.register(ctx, envelope, payload, expires)
}

func (stub *inboxStub) MarkQueued(ctx context.Context, id uuid.UUID, at time.Time) error {
	return stub.markQueued(ctx, id, at)
}

func (stub *inboxStub) ClaimRecoverable(ctx context.Context, provider integrations.ProviderKey, policy RecoveryPolicy, now time.Time) ([]Record, error) {
	return stub.claim(ctx, provider, policy, now)
}

func (stub *inboxStub) ReleaseRecovery(ctx context.Context, id uuid.UUID, generation int32, at time.Time) error {
	return stub.release(ctx, id, generation, at)
}

func (stub *inboxStub) GetByID(ctx context.Context, id uuid.UUID) (Record, error) {
	return stub.getByID(ctx, id)
}

func (stub *inboxStub) GetByExternalKey(ctx context.Context, provider integrations.ProviderKey, accountID, deliveryID string) (Record, error) {
	return stub.getByExternal(ctx, provider, accountID, deliveryID)
}

func (stub *inboxStub) Start(ctx context.Context, id uuid.UUID, now time.Time, lease time.Duration) (Record, bool, error) {
	return stub.start(ctx, id, now, lease)
}

func (stub *inboxStub) Complete(ctx context.Context, id uuid.UUID, status Status, code string, at time.Time) error {
	return stub.complete(ctx, id, status, code, at)
}

func (stub *inboxStub) ExpirePayloads(ctx context.Context, now time.Time, limit int32) ([]uuid.UUID, error) {
	return stub.expire(ctx, now, limit)
}

func newGatewayFixture(
	t *testing.T,
	inbox Inbox,
	verifier WebhookVerifier,
	protector PayloadProtector,
	dispatcher Dispatcher,
) *Gateway {
	t.Helper()
	catalog, err := integrations.NewRegistry(testProviderDescriptor())
	if err != nil {
		t.Fatalf("create provider catalog: %v", err)
	}
	runtimes, err := NewRuntimeRegistry(catalog, RuntimeRegistration{
		Provider:   testProvider,
		Verifier:   verifier,
		Protector:  protector,
		Dispatcher: dispatcher,
	})
	if err != nil {
		t.Fatalf("create webhook runtime registry: %v", err)
	}
	gateway, err := NewGateway(inbox, runtimes, Config{Now: func() time.Time { return testNow }})
	if err != nil {
		t.Fatalf("create webhook gateway: %v", err)
	}
	return gateway
}

func testProviderDescriptor() integrations.Descriptor {
	return integrations.Descriptor{
		Key:         testProvider,
		DisplayName: "Slack",
		Family:      integrations.FamilyMessaging,
		Capabilities: []integrations.Capability{{
			Key: integrations.CapabilityWebhookVerification, MajorVersion: 1,
		}},
		AuthStrategies:  []integrations.AuthStrategy{integrations.AuthStrategyAppInstallation},
		OperatorRunbook: "docs/integrations/providers.md#slack",
	}
}

func testVerifiedDelivery() VerifiedDelivery {
	return VerifiedDelivery{
		DeliveryID:             "Ev-123",
		EventType:              "message.im",
		ExternalAccountID:      "T-123",
		WorkspaceID:            uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		InstallationID:         uuid.MustParse("20000000-0000-4000-8000-000000000002"),
		InstallationGeneration: uuid.MustParse("30000000-0000-4000-8000-000000000003"),
		TraceID:                "4bf92f3577b34da6a3ce929d0e0e4736",
	}
}

func recordFor(envelope Envelope, status Status) Record {
	return Record{
		ID:       uuid.MustParse("40000000-0000-4000-8000-000000000004"),
		Envelope: envelope,
		Status:   status,
	}
}
