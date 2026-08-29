package slack

import (
	"context"
	"strings"
	"testing"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type legacyWebhookPayloadRepositoryStub struct {
	*eventRepositoryStub
	records      []webhooks.Record
	replacements map[uuid.UUID]string
	beforeCAS    func(*legacyWebhookPayloadRepositoryStub, webhooks.Record)
}

func (repository *legacyWebhookPayloadRepositoryStub) ListLegacySlackWebhookPayloads(
	_ context.Context,
	afterID uuid.UUID,
	limit int,
) ([]webhooks.Record, error) {
	rows := make([]webhooks.Record, 0, limit)
	for _, record := range repository.records {
		if afterID != uuid.Nil && record.ID.String() <= afterID.String() {
			continue
		}
		if record.EncryptedPayload == nil ||
			strings.HasPrefix(*record.EncryptedPayload, slackWebhookPayloadEnvelopePrefix) {
			continue
		}
		rows = append(rows, record)
		if len(rows) == limit {
			break
		}
	}
	return rows, nil
}

func (repository *legacyWebhookPayloadRepositoryStub) UpgradeLegacySlackWebhookPayload(
	_ context.Context,
	record webhooks.Record,
	previousPayload, replacementPayload string,
) error {
	if repository.beforeCAS != nil {
		repository.beforeCAS(repository, record)
		repository.beforeCAS = nil
	}
	for index := range repository.records {
		candidate := &repository.records[index]
		if candidate.ID != record.ID || candidate.EncryptedPayload == nil || *candidate.EncryptedPayload != previousPayload {
			continue
		}
		candidate.EncryptedPayload = &replacementPayload
		if repository.replacements == nil {
			repository.replacements = make(map[uuid.UUID]string)
		}
		repository.replacements[record.ID] = replacementPayload
		return nil
	}
	return slackdomain.ErrNotFound
}

func TestBackfillLegacyWebhookPayloadsResealsExactReceiptIdentity(t *testing.T) {
	t.Parallel()
	base := newEventRepositoryStub()
	processor := newTestEventProcessor(t, base, newEventStoreStub(), &assistantStub{}, &accessCheckerStub{allowed: true}, &messageSenderStub{})
	cutover := mustTestLegacyCutover(t)
	record := legacyWebhookCutoverRecord(t, cutover, "Ev-cutover")
	repository := &legacyWebhookPayloadRepositoryStub{eventRepositoryStub: base, records: []webhooks.Record{record}}
	processor.repo = repository

	updated, err := processor.BackfillLegacyWebhookPayloads(context.Background(), cutover)
	require.NoError(t, err)
	require.Equal(t, 1, updated)
	replacement := repository.replacements[record.ID]
	require.True(t, strings.HasPrefix(replacement, slackWebhookPayloadEnvelopePrefix))
	body, err := processor.webhookPayloads.Open(repository.records[0], replacement)
	require.NoError(t, err)
	require.Contains(t, string(body), `"event_id":"Ev-cutover"`)
	clear(body)

	swapped := repository.records[0]
	swapped.InstallationGeneration = uuid.New()
	_, err = processor.webhookPayloads.Open(swapped, replacement)
	require.Error(t, err)
}

func TestBackfillLegacyWebhookPayloadsDoesNotOverwriteConcurrentReplacement(t *testing.T) {
	t.Parallel()
	base := newEventRepositoryStub()
	processor := newTestEventProcessor(t, base, newEventStoreStub(), &assistantStub{}, &accessCheckerStub{allowed: true}, &messageSenderStub{})
	cutover := mustTestLegacyCutover(t)
	record := legacyWebhookCutoverRecord(t, cutover, "Ev-race")
	concurrent := "slack-webhook.v2.concurrent"
	repository := &legacyWebhookPayloadRepositoryStub{
		eventRepositoryStub: base,
		records:             []webhooks.Record{record},
		beforeCAS: func(repository *legacyWebhookPayloadRepositoryStub, _ webhooks.Record) {
			repository.records[0].EncryptedPayload = &concurrent
		},
	}
	processor.repo = repository

	updated, err := processor.BackfillLegacyWebhookPayloads(context.Background(), cutover)
	require.NoError(t, err)
	require.Zero(t, updated)
	require.Equal(t, concurrent, *repository.records[0].EncryptedPayload)
}

func TestNormalWebhookProcessingRejectsLegacyReceipt(t *testing.T) {
	t.Parallel()
	repository := newEventRepositoryStub()
	store := newEventStoreStub()
	processor := newTestEventProcessor(t, repository, store, &assistantStub{}, &accessCheckerStub{allowed: true}, &messageSenderStub{})
	record := legacyWebhookCutoverRecord(t, mustTestLegacyCutover(t), "Ev-no-runtime-fallback")
	store.inboundRecords[record.DeliveryID] = record

	err := processor.ProcessWebhook(context.Background(), slackWebhookProvider, record.ID)
	require.Error(t, err)
	require.ErrorContains(t, err, "envelope is invalid")
}

func legacyWebhookCutoverRecord(t testing.TB, cutover *LegacyCutover, deliveryID string) webhooks.Record {
	t.Helper()
	record := webhooks.Record{
		ID: uuid.New(),
		Envelope: webhooks.Envelope{
			Version:                webhooks.CurrentEnvelopeVersion,
			Provider:               slackWebhookProvider,
			DeliveryID:             deliveryID,
			EventType:              "message",
			ExternalAccountID:      "T1",
			WorkspaceID:            testWorkspaceID,
			InstallationID:         testSlackWorkspaceID,
			InstallationGeneration: testInstallGeneration,
		},
		Status: webhooks.StatusPending,
	}
	body := []byte(`{"type":"event_callback","team_id":"T1","event_id":"` + deliveryID + `","event":{"type":"unsupported"}}`)
	legacy, err := cutover.box.Seal(body)
	require.NoError(t, err)
	record.EncryptedPayload = &legacy
	return record
}
