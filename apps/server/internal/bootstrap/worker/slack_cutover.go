package workerbootstrap

import (
	"context"
	"errors"

	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
)

type slackLegacyBackfiller interface {
	BackfillLegacyCredentials(context.Context, *slack.LegacyCutover) (int, error)
	BackfillLegacyWebhookPayloads(context.Context, *slack.LegacyCutover) (int, error)
}

type slackLegacyBackfillResult struct {
	Credentials     int
	WebhookPayloads int
}

// backfillLegacySlackData preserves the security-sensitive startup order. A
// worker must not consume a durable Slack receipt until provider credentials
// and all retryable webhook bodies use their dedicated encryption domains.
func backfillLegacySlackData(
	ctx context.Context,
	backfiller slackLegacyBackfiller,
	cutover *slack.LegacyCutover,
) (slackLegacyBackfillResult, error) {
	if backfiller == nil || cutover == nil {
		return slackLegacyBackfillResult{}, errors.New("slack legacy backfiller and cutover are required")
	}
	credentials, err := backfiller.BackfillLegacyCredentials(ctx, cutover)
	if err != nil {
		return slackLegacyBackfillResult{}, err
	}
	payloads, err := backfiller.BackfillLegacyWebhookPayloads(ctx, cutover)
	if err != nil {
		return slackLegacyBackfillResult{Credentials: credentials}, err
	}
	return slackLegacyBackfillResult{Credentials: credentials, WebhookPayloads: payloads}, nil
}
