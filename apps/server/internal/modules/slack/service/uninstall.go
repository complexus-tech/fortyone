package slack

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
)

type slackUninstallStore interface {
	CompleteSlackUninstall(ctx context.Context, id uuid.UUID, message string) error
	FailSlackUninstall(ctx context.Context, id uuid.UUID, message string, nextAttemptAt *time.Time) error
}

type activeSlackInstallationFinder interface {
	GetSlackWorkspaceByTeamID(ctx context.Context, slackTeamID string) (slackdomain.Installation, error)
}

func executeSlackUninstall(
	ctx context.Context,
	store slackUninstallStore,
	installations activeSlackInstallationFinder,
	client *slackWebClient,
	codec *credentialCodec,
	clientID, clientSecret string,
	now time.Time,
	record slackdomain.Uninstall,
) (bool, error) {
	active, err := installations.GetSlackWorkspaceByTeamID(ctx, record.SlackTeamID)
	if err == nil {
		message := "superseded by a newer Slack installation"
		if active.ID == record.SlackWorkspaceID && active.InstallGeneration == record.InstallGeneration {
			message = "superseded because the Slack installation is active"
		}
		if completeErr := store.CompleteSlackUninstall(ctx, record.ID, message); completeErr != nil {
			return false, completeErr
		}
		return true, nil
	}
	if !isSlackRepositoryNotFound(err) {
		return false, failSlackUninstall(ctx, store, now, record, fmt.Errorf("check active Slack installation: %w", err))
	}
	if strings.TrimSpace(clientID) == "" || strings.TrimSpace(clientSecret) == "" {
		return false, failSlackUninstall(ctx, store, now, record, ErrSlackNotConfigured)
	}
	if codec == nil {
		return false, failSlackUninstall(ctx, store, now, record, errors.New("slack credential encryption is not configured"))
	}
	if client == nil {
		return false, failSlackUninstall(ctx, store, now, record, errors.New("slack web client is not configured"))
	}
	if record.CredentialKeyVersion != credentialvault.CurrentVersion {
		return false, failSlackUninstall(ctx, store, now, record, errors.New("slack uninstall credential requires vault migration"))
	}
	credential, version, err := codec.open(slackCredentialBinding{
		WorkspaceID:       record.WorkspaceID,
		SlackTeamID:       record.SlackTeamID,
		InstallGeneration: record.InstallGeneration,
	}, record.CredentialPayload)
	if err != nil {
		return false, failSlackUninstall(ctx, store, now, record, err)
	}
	if version != record.CredentialKeyVersion {
		return false, failSlackUninstall(ctx, store, now, record, errors.New("slack uninstall credential version mismatch"))
	}

	err = client.appsUninstall(ctx, clientID, clientSecret, credential.AccessToken)
	if err == nil {
		if completeErr := store.CompleteSlackUninstall(ctx, record.ID, ""); completeErr != nil {
			return false, completeErr
		}
		return true, nil
	}
	if terminalSlackUninstallError(err) {
		if completeErr := store.CompleteSlackUninstall(ctx, record.ID, "Slack installation was already revoked"); completeErr != nil {
			return false, completeErr
		}
		return true, nil
	}
	return false, failSlackUninstall(ctx, store, now, record, err)
}

func failSlackUninstall(ctx context.Context, store slackUninstallStore, now time.Time, record slackdomain.Uninstall, cause error) error {
	var nextAttemptAt *time.Time
	if record.AttemptCount < slackdomain.UninstallMaxAttempts {
		delay := slackUninstallBackoff(record.AttemptCount, cause)
		next := now.UTC().Add(delay)
		nextAttemptAt = &next
	}
	if err := store.FailSlackUninstall(ctx, record.ID, truncateError(cause), nextAttemptAt); err != nil {
		return errors.Join(cause, err)
	}
	return cause
}

func slackUninstallBackoff(attempt int, err error) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	delay := time.Minute
	for i := 1; i < attempt && delay < time.Hour; i++ {
		delay *= 2
	}
	if delay > time.Hour {
		delay = time.Hour
	}
	if retryAfter, ok := SlackRetryAfter(err); ok && retryAfter > delay {
		delay = retryAfter
		if delay > time.Hour {
			delay = time.Hour
		}
	}
	return delay
}

func terminalSlackUninstallError(err error) bool {
	code, ok := SlackAPIErrorCode(err)
	if !ok {
		return false
	}
	switch code {
	case "account_inactive", "app_not_installed", "invalid_auth", "not_authed", "token_revoked":
		return true
	default:
		return false
	}
}
