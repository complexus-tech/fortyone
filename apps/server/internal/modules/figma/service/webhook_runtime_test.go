package figma

import (
	"context"
	"errors"
	"testing"
	"time"

	figmaprovider "github.com/complexus-tech/projects-api/internal/modules/figma"
	figmadomain "github.com/complexus-tech/projects-api/internal/modules/figma/domain"
	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type figmaWebhookStoreStub struct {
	Repository
	webhook    Webhook
	err        error
	currentErr error
}

func (store *figmaWebhookStoreStub) GetWebhook(context.Context, int64) (Webhook, error) {
	return store.webhook, store.err
}

func (store *figmaWebhookStoreStub) GetCurrentWebhook(
	context.Context,
	uuid.UUID,
	uuid.UUID,
	int64,
) (Webhook, error) {
	return store.webhook, store.currentErr
}

type figmaWebhookQueueStub struct {
	payloads []tasks.FigmaWebhookPayload
}

func (queue *figmaWebhookQueueStub) EnqueueFigmaWebhook(
	_ context.Context,
	payload tasks.FigmaWebhookPayload,
) error {
	queue.payloads = append(queue.payloads, payload)
	return nil
}

type figmaInboxStub struct {
	record *webhooks.Record
}

func (inbox *figmaInboxStub) Register(
	_ context.Context,
	envelope webhooks.Envelope,
	encryptedPayload string,
	expiresAt time.Time,
) (webhooks.Record, bool, error) {
	if inbox.record != nil {
		return *inbox.record, false, nil
	}
	record := webhooks.Record{
		Envelope: envelope, ID: uuid.New(), Status: webhooks.StatusPending,
		EncryptedPayload: &encryptedPayload, PayloadExpiresAt: &expiresAt,
		UpdatedAt: envelope.ReceivedAt,
	}
	inbox.record = &record
	return record, true, nil
}

func (*figmaInboxStub) MarkQueued(context.Context, uuid.UUID, time.Time) error { return nil }
func (*figmaInboxStub) ClaimRecoverable(context.Context, integrations.ProviderKey, webhooks.RecoveryPolicy, time.Time) ([]webhooks.Record, error) {
	return nil, nil
}
func (*figmaInboxStub) ReleaseRecovery(context.Context, uuid.UUID, int32, time.Time) error {
	return nil
}
func (inbox *figmaInboxStub) GetByID(context.Context, uuid.UUID) (webhooks.Record, error) {
	if inbox.record == nil {
		return webhooks.Record{}, webhooks.ErrNotFound
	}
	return *inbox.record, nil
}
func (inbox *figmaInboxStub) GetByExternalKey(context.Context, integrations.ProviderKey, string, string) (webhooks.Record, error) {
	if inbox.record == nil {
		return webhooks.Record{}, webhooks.ErrNotFound
	}
	return *inbox.record, nil
}
func (*figmaInboxStub) Start(context.Context, uuid.UUID, time.Time, time.Duration) (webhooks.Record, bool, error) {
	return webhooks.Record{}, false, errors.New("not implemented")
}
func (*figmaInboxStub) Complete(context.Context, uuid.UUID, webhooks.Status, string, time.Time) error {
	return errors.New("not implemented")
}
func (*figmaInboxStub) ExpirePayloads(context.Context, time.Time, int32) ([]uuid.UUID, error) {
	return nil, nil
}

func TestFigmaWebhookGatewayAuthenticatesEncryptsAndDeduplicates(t *testing.T) {
	t.Parallel()

	connectionID, workspaceID, generation := uuid.New(), uuid.New(), uuid.New()
	store := &figmaWebhookStoreStub{webhook: Webhook{
		ID: uuid.New(), ConnectionID: connectionID, WorkspaceID: workspaceID,
		InstallationGeneration: generation, FileKey: "file-key",
		EventType: EventFileUpdate, FigmaWebhookID: 42,
		PasscodeHash: digest("provider-passcode"), IsActive: true,
	}}
	queue := &figmaWebhookQueueStub{}
	runtime, err := NewWebhookRuntime(store, queue, Config{
		WebhookPayloadSecret: "figma-webhook-payload-secret",
	})
	require.NoError(t, err)
	catalog, err := integrations.NewRegistry(figmaprovider.ProviderDescriptor())
	require.NoError(t, err)
	runtimes, err := webhooks.NewRuntimeRegistry(catalog, runtime.Registration)
	require.NoError(t, err)
	inbox := &figmaInboxStub{}
	now := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	gateway, err := webhooks.NewGateway(inbox, runtimes, webhooks.Config{
		Now: func() time.Time { return now },
	})
	require.NoError(t, err)
	body := []byte(`{
		"event_type":"FILE_UPDATE",
		"file_key":"file-key",
		"file_name":"Product",
		"passcode":"provider-passcode",
		"timestamp":"2026-08-28T12:00:00Z",
		"webhook_id":"42"
	}`)

	first, err := gateway.Receive(context.Background(), figmaprovider.ProviderKey, webhooks.SignedRequest{
		Method: "POST", Body: body,
	})
	require.NoError(t, err)
	require.True(t, first.Created)
	require.True(t, first.Queued)
	require.NotEqual(t, uuid.Nil, first.ID)
	require.NotNil(t, inbox.record)
	require.Equal(t, workspaceID, inbox.record.WorkspaceID)
	require.Equal(t, connectionID, inbox.record.InstallationID)
	require.Equal(t, generation, inbox.record.InstallationGeneration)
	require.NotContains(t, *inbox.record.EncryptedPayload, "provider-passcode")
	require.Equal(t, []tasks.FigmaWebhookPayload{{InboxID: first.ID}}, queue.payloads)

	second, err := gateway.Receive(context.Background(), figmaprovider.ProviderKey, webhooks.SignedRequest{
		Method: "POST", Body: body,
	})
	require.NoError(t, err)
	require.False(t, second.Created)
	require.Equal(t, first.ID, second.ID)
	require.Len(t, queue.payloads, 2)
}

func TestFigmaWebhookVerifierRejectsPasscodeBeforePersistence(t *testing.T) {
	t.Parallel()

	store := &figmaWebhookStoreStub{webhook: Webhook{
		ID: uuid.New(), ConnectionID: uuid.New(), WorkspaceID: uuid.New(),
		InstallationGeneration: uuid.New(), FileKey: "file-key",
		EventType: EventFileUpdate, FigmaWebhookID: 42,
		PasscodeHash: digest("expected"), IsActive: true,
	}}
	verifier := &figmaWebhookVerifier{installations: store}
	_, err := verifier.Verify(context.Background(), webhooks.SignedRequest{Body: []byte(`{
		"event_type":"FILE_UPDATE",
		"file_key":"file-key",
		"passcode":"wrong",
		"webhook_id":42
	}`)})
	require.ErrorIs(t, err, webhooks.ErrUnauthenticated)
}

func TestFigmaWebhookVerifierIgnoresOnlyMissingInstallations(t *testing.T) {
	t.Parallel()

	request := webhooks.SignedRequest{Body: []byte(`{
		"event_type":"FILE_UPDATE",
		"file_key":"file-key",
		"passcode":"passcode",
		"webhook_id":42
	}`)}
	verifier := &figmaWebhookVerifier{installations: &figmaWebhookStoreStub{
		err: figmadomain.ErrNotFound,
	}}
	_, err := verifier.Verify(context.Background(), request)
	require.ErrorIs(t, err, webhooks.ErrDeliveryIgnored)

	databaseErr := errors.New("database unavailable")
	verifier.installations = &figmaWebhookStoreStub{err: databaseErr}
	_, err = verifier.Verify(context.Background(), request)
	require.ErrorIs(t, err, webhooks.ErrVerificationUnavailable)
	require.NotErrorIs(t, err, databaseErr)
	require.NotErrorIs(t, err, webhooks.ErrDeliveryIgnored)
}

func TestFigmaWebhookGrantCancelsMissingInstallationButSurfacesInfrastructureFailure(t *testing.T) {
	t.Parallel()

	connectionID, workspaceID, generation := uuid.New(), uuid.New(), uuid.New()
	record := webhooks.Record{Envelope: webhooks.Envelope{
		ExternalAccountID: "42", WorkspaceID: workspaceID,
		InstallationID: connectionID, InstallationGeneration: generation,
	}}
	event := WebhookEvent{WebhookID: 42}
	store := &figmaWebhookStoreStub{currentErr: figmadomain.ErrNotFound}
	service := &Service{repo: store}

	current, err := service.currentWebhookGrant(context.Background(), record, event)
	require.NoError(t, err)
	require.False(t, current)

	databaseErr := errors.New("database unavailable")
	store.currentErr = databaseErr
	current, err = service.currentWebhookGrant(context.Background(), record, event)
	require.False(t, current)
	require.ErrorIs(t, err, databaseErr)
}
