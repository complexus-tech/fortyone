package slack

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
)

func inboundReceiptMatchesInstallation(receipt inboundEventRecord, installation slackWorkspaceRecord) bool {
	if receipt.InstallGeneration == nil ||
		*receipt.InstallGeneration == uuid.Nil ||
		*receipt.InstallGeneration != installation.InstallGeneration {
		return false
	}
	if receipt.WorkspaceID != nil && *receipt.WorkspaceID != installation.WorkspaceID {
		return false
	}
	if receipt.InstallationID != nil && *receipt.InstallationID != installation.ID {
		return false
	}
	return true
}

func isSlackLifecycleEvent(kind slackEventKind) bool {
	return kind == slackEventKindUninstalled || kind == slackEventKindRevoked
}

func slackLifecycleEventIsCurrent(eventTimeUnix int64, authorizedAt time.Time) bool {
	if eventTimeUnix <= 0 || authorizedAt.IsZero() {
		return false
	}
	// Slack event_time has second precision. Events from the authorization
	// second are current; older lifecycle events belong to a prior install.
	return eventTimeUnix >= authorizedAt.UTC().Unix()
}

func installationBotTokenRevoked(installation slackWorkspaceRecord, revokedBotUserIDs []string) bool {
	if len(revokedBotUserIDs) == 0 {
		return false
	}
	if installation.BotUserID == nil || strings.TrimSpace(*installation.BotUserID) == "" {
		return true
	}
	botUserID := strings.TrimSpace(*installation.BotUserID)
	for _, revokedUserID := range revokedBotUserIDs {
		if strings.TrimSpace(revokedUserID) == botUserID {
			return true
		}
	}
	return false
}

func (p *EventProcessor) requireCurrentInstallation(ctx context.Context, workspaceID uuid.UUID, slackTeamID string, generation uuid.UUID) error {
	current, err := p.repo.GetSlackWorkspaceByTeamID(ctx, slackTeamID)
	if err != nil {
		if isSlackRepositoryNotFound(err) {
			return fmt.Errorf("%w: Slack team is no longer connected", errSlackInstallationChanged)
		}
		return err
	}
	if !current.IsActive || current.WorkspaceID != workspaceID || current.InstallGeneration != generation {
		return fmt.Errorf("%w: active generation no longer matches", errSlackInstallationChanged)
	}
	return nil
}

func (p *EventProcessor) botToken(ctx context.Context, installation slackWorkspaceRecord) (string, error) {
	if installation.CredentialVersion != credentialvault.CurrentVersion {
		return "", errors.New("slack credential requires vault migration")
	}
	credential, version, err := p.codec.open(slackCredentialBinding{
		WorkspaceID:       installation.WorkspaceID,
		SlackTeamID:       installation.SlackTeamID,
		InstallGeneration: installation.InstallGeneration,
	}, installation.BotAccessToken)
	if err != nil {
		return "", err
	}
	if version != installation.CredentialVersion {
		return "", errors.New("slack credential envelope version mismatch")
	}
	return credential.AccessToken, nil
}
