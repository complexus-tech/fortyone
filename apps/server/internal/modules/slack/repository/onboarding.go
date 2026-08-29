package slackrepository

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	slacksql "github.com/complexus-tech/projects-api/internal/modules/slack/repository/sqlc"
	"github.com/google/uuid"
)

func (repository *Repo) HasSlackUserOnboardingReceipt(
	ctx context.Context,
	workspaceID uuid.UUID,
	slackTeamID, slackUserID string,
) (bool, error) {
	if workspaceID == uuid.Nil {
		return false, errors.Join(slackdomain.ErrInvalidInput, errors.New("workspace is required"))
	}
	slackTeamID = strings.TrimSpace(slackTeamID)
	if slackTeamID == "" {
		return false, errors.Join(slackdomain.ErrInvalidInput, errors.New("slack team is required"))
	}
	slackUserID = strings.TrimSpace(slackUserID)
	if slackUserID == "" {
		return false, errors.Join(slackdomain.ErrInvalidInput, errors.New("slack user is required"))
	}
	identityDigest := slackOnboardingIdentityDigest(slackTeamID, slackUserID)
	exists, err := repository.queries.HasSlackUserOnboardingReceipt(ctx, slacksql.HasSlackUserOnboardingReceiptParams{
		WorkspaceID:    workspaceID,
		IdentityDigest: identityDigest[:],
	})
	if err != nil {
		return false, fmt.Errorf("check Slack user onboarding receipt: %w", mapDatabaseError(err))
	}
	return exists, nil
}

func slackOnboardingIdentityDigest(slackTeamID, slackUserID string) [sha256.Size]byte {
	return sha256.Sum256([]byte(strings.TrimSpace(slackTeamID) + "\x1f" + strings.TrimSpace(slackUserID)))
}
