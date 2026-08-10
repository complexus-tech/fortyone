package slackrepository

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func (r *Repo) HasSlackUserOnboardingReceipt(
	ctx context.Context,
	workspaceID uuid.UUID,
	slackTeamID, slackUserID string,
) (bool, error) {
	if workspaceID == uuid.Nil {
		return false, errors.New("workspace is required")
	}
	slackTeamID = strings.TrimSpace(slackTeamID)
	if slackTeamID == "" {
		return false, errors.New("Slack team is required")
	}
	slackUserID = strings.TrimSpace(slackUserID)
	if slackUserID == "" {
		return false, errors.New("Slack user is required")
	}
	identityDigest := slackOnboardingIdentityDigest(slackTeamID, slackUserID)

	var exists bool
	if err := r.db.GetContext(ctx, &exists, `
		SELECT EXISTS (
			SELECT 1
			FROM slack_user_onboarding_receipts
			WHERE workspace_id = $1
			  AND external_identity_digest = $2
		)
	`, workspaceID, identityDigest[:]); err != nil {
		return false, fmt.Errorf("check Slack user onboarding receipt: %w", err)
	}
	return exists, nil
}

func slackOnboardingIdentityDigest(slackTeamID, slackUserID string) [sha256.Size]byte {
	return sha256.Sum256([]byte(strings.TrimSpace(slackTeamID) + "\x1f" + strings.TrimSpace(slackUserID)))
}
