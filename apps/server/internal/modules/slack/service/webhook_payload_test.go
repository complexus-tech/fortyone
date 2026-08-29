package slack

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSlackWebhookPayloadCodecBindsEveryDurableIdentityField(t *testing.T) {
	t.Parallel()
	codec, err := newSlackWebhookPayloadCodec(testSlackWebhookPayloadSecret)
	require.NoError(t, err)
	binding := testSlackWebhookPayloadBinding()
	body := []byte(`{"type":"event_callback","team_id":"T1","event_id":"Ev-bound"}`)

	encrypted, err := codec.Seal(context.Background(), binding, body)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(encrypted, slackWebhookPayloadEnvelopePrefix))

	opened, err := codec.Open(webhookRecordForBinding(binding), encrypted)
	require.NoError(t, err)
	require.Equal(t, body, opened)

	tests := map[string]func(*webhooks.Record){
		"provider":     func(record *webhooks.Record) { record.Provider = "github" },
		"delivery":     func(record *webhooks.Record) { record.DeliveryID = "Ev-other" },
		"workspace":    func(record *webhooks.Record) { record.WorkspaceID = uuid.New() },
		"installation": func(record *webhooks.Record) { record.InstallationID = uuid.New() },
		"generation":   func(record *webhooks.Record) { record.InstallationGeneration = uuid.New() },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			record := webhookRecordForBinding(binding)
			mutate(&record)
			_, openErr := codec.Open(record, encrypted)
			require.Error(t, openErr)
		})
	}
}

func TestSlackWebhookPayloadCodecRejectsTamperWrongKeyAndLegacyFormats(t *testing.T) {
	t.Parallel()
	codec, err := newSlackWebhookPayloadCodec(testSlackWebhookPayloadSecret)
	require.NoError(t, err)
	binding := testSlackWebhookPayloadBinding()
	record := webhookRecordForBinding(binding)
	body := []byte(`{"type":"event_callback","team_id":"T1","event_id":"Ev-bound"}`)
	encrypted, err := codec.Seal(context.Background(), binding, body)
	require.NoError(t, err)

	tampered := []byte(encrypted)
	tamperIndex := len(slackWebhookPayloadEnvelopePrefix) + 8
	if tampered[tamperIndex] == 'A' {
		tampered[tamperIndex] = 'B'
	} else {
		tampered[tamperIndex] = 'A'
	}
	_, err = codec.Open(record, string(tampered))
	require.Error(t, err)

	wrongKeyCodec, err := newSlackWebhookPayloadCodec("different-dedicated-slack-payload-key")
	require.NoError(t, err)
	_, err = wrongKeyCodec.Open(record, encrypted)
	require.Error(t, err)

	legacy, err := NewLegacyCutover(testSlackCredentialSecret)
	require.NoError(t, err)
	legacyCiphertext, err := legacy.box.Seal(body)
	require.NoError(t, err)
	_, err = codec.Open(record, legacyCiphertext)
	require.Error(t, err)

	oldBoundCiphertext := sealLegacyBoundSlackWebhookPayload(t, legacy, record, body)
	_, err = codec.Open(record, oldBoundCiphertext)
	require.Error(t, err)
}

func TestSlackLegacyWebhookCutoverAuthenticatesReceiptBeforeReseal(t *testing.T) {
	t.Parallel()
	cutover, err := NewLegacyCutover(testSlackCredentialSecret)
	require.NoError(t, err)
	codec, err := newSlackWebhookPayloadCodec(testSlackWebhookPayloadSecret)
	require.NoError(t, err)
	binding := testSlackWebhookPayloadBinding()
	record := webhookRecordForBinding(binding)
	record.ExternalAccountID = "T1"
	body := []byte(`{"type":"event_callback","team_id":"T1","event_id":"Ev-bound"}`)
	legacy := sealLegacyBoundSlackWebhookPayload(t, cutover, record, body)

	opened, err := cutover.openWebhookPayload(record, legacy)
	require.NoError(t, err)
	require.Equal(t, body, opened)
	resealed, err := codec.Seal(context.Background(), slackWebhookPayloadBinding(record), opened)
	require.NoError(t, err)
	clear(opened)
	opened, err = codec.Open(record, resealed)
	require.NoError(t, err)
	require.Equal(t, body, opened)

	swapped := record
	swapped.InstallationGeneration = uuid.New()
	_, err = cutover.openWebhookPayload(swapped, legacy)
	require.ErrorContains(t, err, "does not match its durable identity")
}

func sealLegacyBoundSlackWebhookPayload(
	t testing.TB,
	cutover *LegacyCutover,
	record webhooks.Record,
	body []byte,
) string {
	t.Helper()
	encoded, err := json.Marshal(legacySlackWebhookPayloadEnvelope{
		Version:                legacySlackWebhookPayloadVersion,
		Provider:               string(record.Provider),
		DeliveryID:             record.DeliveryID,
		WorkspaceID:            record.WorkspaceID,
		InstallationID:         record.InstallationID,
		InstallationGeneration: record.InstallationGeneration,
		Body:                   body,
	})
	require.NoError(t, err)
	defer clear(encoded)
	encrypted, err := cutover.box.Seal(encoded)
	require.NoError(t, err)
	return legacySlackWebhookPayloadEnvelopePrefix + encrypted
}

func testSlackWebhookPayloadBinding() webhooks.PayloadBinding {
	return webhooks.PayloadBinding{
		Provider:               slackWebhookProvider,
		DeliveryID:             "Ev-bound",
		WorkspaceID:            uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		InstallationID:         uuid.MustParse("22222222-2222-4222-8222-222222222222"),
		InstallationGeneration: uuid.MustParse("33333333-3333-4333-8333-333333333333"),
	}
}

func webhookRecordForBinding(binding webhooks.PayloadBinding) webhooks.Record {
	return webhooks.Record{Envelope: webhooks.Envelope{
		Provider:               binding.Provider,
		DeliveryID:             binding.DeliveryID,
		WorkspaceID:            binding.WorkspaceID,
		InstallationID:         binding.InstallationID,
		InstallationGeneration: binding.InstallationGeneration,
	}}
}
