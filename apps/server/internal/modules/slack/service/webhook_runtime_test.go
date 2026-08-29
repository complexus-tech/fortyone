package slack

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type slackGatewayInboxStub struct {
	status        webhooks.Status
	created       bool
	registerCalls int
	queuedCalls   int
}

func (inbox *slackGatewayInboxStub) Register(
	_ context.Context,
	envelope webhooks.Envelope,
	encryptedPayload string,
	expiresAt time.Time,
) (webhooks.Record, bool, error) {
	inbox.registerCalls++
	status := inbox.status
	if status == "" {
		status = webhooks.StatusPending
	}
	return webhooks.Record{
		ID:               testInboundReceiptID,
		Envelope:         envelope,
		Status:           status,
		EncryptedPayload: &encryptedPayload,
		PayloadExpiresAt: &expiresAt,
	}, inbox.created, nil
}

func (inbox *slackGatewayInboxStub) MarkQueued(context.Context, uuid.UUID, time.Time) error {
	inbox.queuedCalls++
	return nil
}

func (*slackGatewayInboxStub) ClaimRecoverable(
	context.Context,
	integrations.ProviderKey,
	webhooks.RecoveryPolicy,
	time.Time,
) ([]webhooks.Record, error) {
	return nil, nil
}

func (*slackGatewayInboxStub) ReleaseRecovery(context.Context, uuid.UUID, int32, time.Time) error {
	return nil
}

func (*slackGatewayInboxStub) GetByID(context.Context, uuid.UUID) (webhooks.Record, error) {
	return webhooks.Record{}, webhooks.ErrNotFound
}

func (*slackGatewayInboxStub) GetByExternalKey(
	context.Context,
	integrations.ProviderKey,
	string,
	string,
) (webhooks.Record, error) {
	return webhooks.Record{}, webhooks.ErrNotFound
}

func (*slackGatewayInboxStub) Start(
	context.Context,
	uuid.UUID,
	time.Time,
	time.Duration,
) (webhooks.Record, bool, error) {
	return webhooks.Record{}, false, webhooks.ErrNotFound
}

func (*slackGatewayInboxStub) Complete(
	context.Context,
	uuid.UUID,
	webhooks.Status,
	string,
	time.Time,
) error {
	return nil
}

func (*slackGatewayInboxStub) ExpirePayloads(context.Context, time.Time, int32) ([]uuid.UUID, error) {
	return nil, nil
}

func TestSlackWebhookGatewayVerifiesSignatureAndQueuesInboxIdentity(t *testing.T) {
	t.Parallel()
	now := time.Unix(1_800_000_000, 0).UTC()
	inbox := &slackGatewayInboxStub{created: true}
	queue := &eventQueueStub{}
	gateway := newTestSlackWebhookGateway(t, now, inbox, queue)
	body := []byte(directMessageEvent("Ev-signed", "show my work"))

	receipt, err := gateway.Receive(context.Background(), slackWebhookProvider, signedSlackWebhookRequest(now, body))

	require.NoError(t, err)
	require.True(t, receipt.Created)
	require.True(t, receipt.Queued)
	require.Equal(t, 1, inbox.registerCalls)
	require.Equal(t, 1, inbox.queuedCalls)
	require.Len(t, queue.payloads, 1)
	require.Equal(t, "slack", queue.payloads[0].Provider)
	require.Equal(t, testInboundReceiptID, queue.payloads[0].InboxID)
	require.Empty(t, queue.payloads[0].ExternalWorkspaceID)
	require.Empty(t, queue.payloads[0].EventID)
}

func TestSlackWebhookGatewayRejectsReplayBeforePersistence(t *testing.T) {
	t.Parallel()
	now := time.Unix(1_800_000_000, 0).UTC()
	inbox := &slackGatewayInboxStub{created: true}
	queue := &eventQueueStub{}
	gateway := newTestSlackWebhookGateway(t, now, inbox, queue)
	body := []byte(directMessageEvent("Ev-replay", "show my work"))
	request := signedSlackWebhookRequest(now.Add(-6*time.Minute), body)

	_, err := gateway.Receive(context.Background(), slackWebhookProvider, request)

	require.ErrorIs(t, err, webhooks.ErrReplay)
	require.Zero(t, inbox.registerCalls)
	require.Empty(t, queue.payloads)
}

func TestSlackWebhookGatewayRejectsInvalidSignatureBeforePersistence(t *testing.T) {
	t.Parallel()
	now := time.Unix(1_800_000_000, 0).UTC()
	inbox := &slackGatewayInboxStub{created: true}
	queue := &eventQueueStub{}
	gateway := newTestSlackWebhookGateway(t, now, inbox, queue)
	request := signedSlackWebhookRequest(now, []byte(directMessageEvent("Ev-invalid-signature", "show my work")))
	request.Headers["X-Slack-Signature"] = []string{"v0=invalid"}

	_, err := gateway.Receive(context.Background(), slackWebhookProvider, request)

	require.ErrorIs(t, err, webhooks.ErrUnauthenticated)
	require.Zero(t, inbox.registerCalls)
	require.Empty(t, queue.payloads)
}

func TestSlackWebhookGatewayDoesNotDispatchTerminalDuplicate(t *testing.T) {
	t.Parallel()
	now := time.Unix(1_800_000_000, 0).UTC()
	inbox := &slackGatewayInboxStub{status: webhooks.StatusCompleted, created: false}
	queue := &eventQueueStub{}
	gateway := newTestSlackWebhookGateway(t, now, inbox, queue)

	receipt, err := gateway.Receive(
		context.Background(),
		slackWebhookProvider,
		signedSlackWebhookRequest(now, []byte(directMessageEvent("Ev-duplicate", "show my work"))),
	)

	require.NoError(t, err)
	require.Equal(t, webhooks.StatusCompleted, receipt.Status)
	require.False(t, receipt.Queued)
	require.Empty(t, queue.payloads)
	require.Zero(t, inbox.queuedCalls)
}

func TestSlackWebhookGatewayLeavesReceiptRecoverableWhenDispatchFails(t *testing.T) {
	t.Parallel()
	now := time.Unix(1_800_000_000, 0).UTC()
	inbox := &slackGatewayInboxStub{created: true}
	queue := &eventQueueStub{errors: []error{errors.New("queue unavailable")}}
	gateway := newTestSlackWebhookGateway(t, now, inbox, queue)

	_, err := gateway.Receive(
		context.Background(),
		slackWebhookProvider,
		signedSlackWebhookRequest(now, []byte(directMessageEvent("Ev-dispatch", "show my work"))),
	)

	require.ErrorIs(t, err, webhooks.ErrDispatchUnavailable)
	require.Equal(t, 1, inbox.registerCalls)
	require.Zero(t, inbox.queuedCalls)
}

func newTestSlackWebhookGateway(
	t *testing.T,
	now time.Time,
	inbox webhooks.Inbox,
	queue EventQueue,
) *webhooks.Gateway {
	t.Helper()
	repository := newEventRepositoryStub()
	repository.installation.ID = testSlackWorkspaceID
	runtime, err := NewWebhookRuntime(repository, queue, Config{
		SigningSecret:        "slack-signing-secret",
		WebhookPayloadSecret: testSlackWebhookPayloadSecret,
	})
	require.NoError(t, err)
	catalog, err := integrations.NewRegistry(integrations.Descriptor{
		Key:         slackWebhookProvider,
		DisplayName: "Slack",
		Family:      integrations.FamilyMessaging,
		Capabilities: []integrations.Capability{
			{Key: integrations.CapabilityWebhookVerification, MajorVersion: 1},
		},
		AuthStrategies:  []integrations.AuthStrategy{integrations.AuthStrategyOAuthInstall},
		OperatorRunbook: "docs/integrations/providers.md#slack",
	})
	require.NoError(t, err)
	runtimes, err := webhooks.NewRuntimeRegistry(catalog, runtime)
	require.NoError(t, err)
	gateway, err := webhooks.NewGateway(inbox, runtimes, webhooks.Config{Now: func() time.Time { return now }})
	require.NoError(t, err)
	return gateway
}

func signedSlackWebhookRequest(timestamp time.Time, body []byte) webhooks.SignedRequest {
	unixTimestamp := strconv.FormatInt(timestamp.UTC().Unix(), 10)
	return webhooks.SignedRequest{
		Method: "POST",
		Headers: webhooks.Headers{
			"X-Slack-Request-Timestamp": {unixTimestamp},
			"X-Slack-Signature":         {slackSignature("slack-signing-secret", unixTimestamp, body)},
		},
		Body: body,
	}
}

var _ webhooks.Inbox = (*slackGatewayInboxStub)(nil)
