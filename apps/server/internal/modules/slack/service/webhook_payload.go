package slack

import (
	"context"

	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
)

// v2 is the Slack cutover from APP_AUTH_SECRET_KEY to the dedicated
// APP_SLACK_WEBHOOK_PAYLOAD_SECRET. The shared codec has its own authenticated
// envelope version; this prefix identifies the provider key domain used to
// create it so normal processing can never silently open a legacy receipt.
const slackWebhookPayloadEnvelopePrefix = "slack-webhook.v2."

type slackWebhookPayloadCodec interface {
	Seal(ctx context.Context, binding webhooks.PayloadBinding, payload []byte) (string, error)
	Open(record webhooks.Record, value string) ([]byte, error)
}

func newSlackWebhookPayloadCodec(secret string) (slackWebhookPayloadCodec, error) {
	return webhooks.NewBoundPayloadCodec(
		slackWebhookProvider,
		slackWebhookPayloadEnvelopePrefix,
		secret,
	)
}

func slackWebhookPayloadBinding(record webhooks.Record) webhooks.PayloadBinding {
	return webhooks.PayloadBinding{
		Provider:               record.Provider,
		DeliveryID:             record.DeliveryID,
		WorkspaceID:            record.WorkspaceID,
		InstallationID:         record.InstallationID,
		InstallationGeneration: record.InstallationGeneration,
	}
}
