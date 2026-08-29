package slackrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	slacksql "github.com/complexus-tech/projects-api/internal/modules/slack/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repo) EnqueueSlackUninstall(ctx context.Context, input SlackUninstallInput) (SlackUninstallRecord, error) {
	if input.SlackWorkspaceID == uuid.Nil {
		input.SlackWorkspaceID = uuid.New()
	}
	if input.InstallGeneration == uuid.Nil {
		input.InstallGeneration = uuid.New()
	}
	if strings.TrimSpace(string(input.UninstallKind)) == "" {
		input.UninstallKind = slackdomain.UninstallDisconnect
	}
	input.SlackTeamID = strings.TrimSpace(input.SlackTeamID)
	input.CredentialPayload = strings.TrimSpace(input.CredentialPayload)
	if input.WorkspaceID == uuid.Nil || input.SlackTeamID == "" ||
		input.CredentialKeyVersion != credentialvault.CurrentVersion ||
		!credentialvault.IsEnvelope(input.CredentialPayload) {
		return SlackUninstallRecord{}, errors.Join(slackdomain.ErrInvalidInput, errors.New("slack uninstall requires workspace, team, and current vault credential"))
	}
	credentialVersion, err := safecast.Int16(input.CredentialKeyVersion)
	if err != nil {
		return SlackUninstallRecord{}, errors.Join(slackdomain.ErrInvalidInput, err)
	}
	row, err := repository.queries.EnqueueSlackUninstall(ctx, slacksql.EnqueueSlackUninstallParams{
		SlackWorkspaceID: input.SlackWorkspaceID, WorkspaceID: input.WorkspaceID,
		InstallationGeneration: input.InstallGeneration, SlackTeamID: input.SlackTeamID,
		UninstallKind: string(input.UninstallKind), CredentialPayload: input.CredentialPayload,
		CredentialKeyVersion: credentialVersion,
	})
	if err != nil {
		return SlackUninstallRecord{}, fmt.Errorf("enqueue Slack uninstall: %w", mapDatabaseError(err))
	}
	return mapUninstall(row)
}

func (repository *Repo) ClaimSlackUninstall(ctx context.Context, id uuid.UUID) (SlackUninstallRecord, bool, error) {
	maxAttempts, err := safecast.Int32(SlackUninstallMaxAttempts)
	if err != nil {
		return SlackUninstallRecord{}, false, err
	}
	var result SlackUninstallRecord
	claimed := false
	err = repository.withinTransaction(ctx, func(queries slacksql.Querier) error {
		if err := lockInstallationLifecycle(ctx, queries); err != nil {
			return err
		}
		row, err := queries.ClaimSlackUninstall(ctx, slacksql.ClaimSlackUninstallParams{
			UninstallID: id, MaxAttempts: maxAttempts,
			LeaseSeconds: int64(slackUninstallLease / time.Second),
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("claim Slack uninstall: %w", err)
		}
		result, err = mapUninstall(row)
		claimed = err == nil
		return err
	})
	return result, claimed, err
}

func (repository *Repo) ClaimRecoverableSlackUninstalls(ctx context.Context, limit int) ([]SlackUninstallRecord, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	queryLimit, err := safecast.Int32(limit)
	if err != nil {
		return nil, err
	}
	maxAttempts, err := safecast.Int32(SlackUninstallMaxAttempts)
	if err != nil {
		return nil, err
	}
	result := make([]SlackUninstallRecord, 0, limit)
	err = repository.withinTransaction(ctx, func(queries slacksql.Querier) error {
		if err := lockInstallationLifecycle(ctx, queries); err != nil {
			return err
		}
		leaseSeconds := int64(slackUninstallLease / time.Second)
		if _, err := queries.DeadLetterExhaustedSlackUninstalls(ctx, slacksql.DeadLetterExhaustedSlackUninstallsParams{
			MaxAttempts: maxAttempts, LeaseSeconds: leaseSeconds,
		}); err != nil {
			return fmt.Errorf("dead-letter exhausted Slack uninstalls: %w", err)
		}
		rows, err := queries.ClaimRecoverableSlackUninstalls(ctx, slacksql.ClaimRecoverableSlackUninstallsParams{
			MaxAttempts: maxAttempts, LeaseSeconds: leaseSeconds, ResultLimit: queryLimit,
		})
		if err != nil {
			return fmt.Errorf("claim recoverable Slack uninstalls: %w", err)
		}
		for _, row := range rows {
			mapped, err := mapUninstall(row)
			if err != nil {
				return err
			}
			result = append(result, mapped)
		}
		return nil
	})
	return result, err
}

func (repository *Repo) CompleteSlackUninstall(ctx context.Context, id uuid.UUID, message string) error {
	affected, err := repository.queries.CompleteSlackUninstall(ctx, slacksql.CompleteSlackUninstallParams{
		UninstallID: id,
		Message:     strings.TrimSpace(message),
	})
	if err != nil {
		return fmt.Errorf("complete Slack uninstall: %w", mapDatabaseError(err))
	}
	if affected != 1 {
		return slackdomain.ErrConflict
	}
	return nil
}

func (repository *Repo) FailSlackUninstall(ctx context.Context, id uuid.UUID, message string, nextAttemptAt *time.Time) error {
	affected, err := repository.queries.FailSlackUninstall(ctx, slacksql.FailSlackUninstallParams{
		UninstallID: id, Message: strings.TrimSpace(message), NextAttemptAt: nextAttemptAt,
	})
	if err != nil {
		return fmt.Errorf("fail Slack uninstall: %w", mapDatabaseError(err))
	}
	if affected != 1 {
		return slackdomain.ErrConflict
	}
	return nil
}
