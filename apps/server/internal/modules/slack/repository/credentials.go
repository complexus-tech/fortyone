package slackrepository

import (
	"context"
	"errors"
	"strings"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	slacksql "github.com/complexus-tech/projects-api/internal/modules/slack/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/google/uuid"
)

const slackWebhookPayloadEnvelopePrefix = "slack-webhook.v2."

type (
	LegacySlackCredentialRecord          = slackdomain.LegacyInstallationCredential
	LegacySlackUninstallCredentialRecord = slackdomain.LegacyUninstallCredential
	SlackCredentialRewrapRecord          = slackdomain.InstallationCredentialForRewrap
	SlackUninstallCredentialRewrapRecord = slackdomain.UninstallCredentialForRewrap
)

func (repository *Repo) ListSlackCredentialsForRewrap(ctx context.Context, after *uuid.UUID, limit int) ([]SlackCredentialRewrapRecord, error) {
	limit = credentialvault.MaintenanceBatchSize(limit)
	queryLimit, version, err := credentialQueryBounds(limit)
	if err != nil {
		return nil, err
	}
	rows, err := repository.queries.ListSlackCredentialsForRewrap(ctx, slacksql.ListSlackCredentialsForRewrapParams{
		AfterID: after, ResultLimit: queryLimit, CredentialKeyVersion: version,
	})
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	result := make([]SlackCredentialRewrapRecord, 0, len(rows))
	for _, row := range rows {
		credential := ""
		if row.CredentialPayload != nil {
			credential = *row.CredentialPayload
		}
		result = append(result, slackdomain.InstallationCredentialForRewrap{
			SlackWorkspaceID: row.ID, WorkspaceID: row.WorkspaceID,
			SlackTeamID: row.SlackTeamID, InstallGeneration: row.InstallationGeneration,
			Credential: credential, CredentialVersion: int(row.CredentialKeyVersion),
		})
	}
	return result, nil
}

func (repository *Repo) RewrapSlackCredential(ctx context.Context, record SlackCredentialRewrapRecord, rewrapped string) (bool, error) {
	if record.SlackWorkspaceID == uuid.Nil || record.WorkspaceID == uuid.Nil ||
		record.InstallGeneration == uuid.Nil || strings.TrimSpace(record.SlackTeamID) == "" ||
		record.CredentialVersion != credentialvault.CurrentVersion ||
		!credentialvault.IsEnvelope(record.Credential) || !credentialvault.IsEnvelope(rewrapped) {
		return false, errors.Join(slackdomain.ErrInvalidInput, errors.New("slack credential rewrap metadata is invalid"))
	}
	version, err := safecast.Int16(record.CredentialVersion)
	if err != nil {
		return false, err
	}
	affected, err := repository.queries.RewrapSlackCredential(ctx, slacksql.RewrapSlackCredentialParams{
		ReplacementCredential: rewrapped, SlackWorkspaceID: record.SlackWorkspaceID,
		WorkspaceID: record.WorkspaceID, SlackTeamID: record.SlackTeamID,
		InstallationGeneration: record.InstallGeneration, CredentialKeyVersion: version,
		PreviousCredential: record.Credential,
	})
	return affected == 1, mapDatabaseError(err)
}

func (repository *Repo) ListSlackUninstallCredentialsForRewrap(ctx context.Context, after *uuid.UUID, limit int) ([]SlackUninstallCredentialRewrapRecord, error) {
	limit = credentialvault.MaintenanceBatchSize(limit)
	queryLimit, version, err := credentialQueryBounds(limit)
	if err != nil {
		return nil, err
	}
	rows, err := repository.queries.ListSlackUninstallCredentialsForRewrap(ctx, slacksql.ListSlackUninstallCredentialsForRewrapParams{
		AfterID: after, ResultLimit: queryLimit, CredentialKeyVersion: version,
	})
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	result := make([]SlackUninstallCredentialRewrapRecord, 0, len(rows))
	for _, row := range rows {
		credential := ""
		if row.CredentialPayload != nil {
			credential = *row.CredentialPayload
		}
		result = append(result, slackdomain.UninstallCredentialForRewrap{
			UninstallID: row.ID, WorkspaceID: row.WorkspaceID,
			SlackTeamID: row.SlackTeamID, InstallGeneration: row.InstallationGeneration,
			Credential: credential, CredentialVersion: int(row.CredentialKeyVersion),
		})
	}
	return result, nil
}

func (repository *Repo) RewrapSlackUninstallCredential(ctx context.Context, record SlackUninstallCredentialRewrapRecord, rewrapped string) (bool, error) {
	if record.UninstallID == uuid.Nil || record.WorkspaceID == uuid.Nil ||
		record.InstallGeneration == uuid.Nil || strings.TrimSpace(record.SlackTeamID) == "" ||
		record.CredentialVersion != credentialvault.CurrentVersion ||
		!credentialvault.IsEnvelope(record.Credential) || !credentialvault.IsEnvelope(rewrapped) {
		return false, errors.Join(slackdomain.ErrInvalidInput, errors.New("slack uninstall credential rewrap metadata is invalid"))
	}
	version, err := safecast.Int16(record.CredentialVersion)
	if err != nil {
		return false, err
	}
	affected, err := repository.queries.RewrapSlackUninstallCredential(ctx, slacksql.RewrapSlackUninstallCredentialParams{
		ReplacementCredential: rewrapped, UninstallID: record.UninstallID,
		WorkspaceID: record.WorkspaceID, SlackTeamID: record.SlackTeamID,
		InstallationGeneration: record.InstallGeneration, CredentialKeyVersion: version,
		PreviousCredential: record.Credential,
	})
	return affected == 1, mapDatabaseError(err)
}

func (repository *Repo) UpgradeSlackCredential(ctx context.Context, record LegacySlackCredentialRecord, encrypted string, version int) error {
	if record.SlackWorkspaceID == uuid.Nil || strings.TrimSpace(record.Credential) == "" ||
		version != credentialvault.CurrentVersion || !credentialvault.IsEnvelope(encrypted) {
		return errors.Join(slackdomain.ErrInvalidInput, errors.New("slack credential upgrade metadata is invalid"))
	}
	replacementVersion, err := safecast.Int16(version)
	if err != nil {
		return err
	}
	previousVersion, err := safecast.Int16(record.CredentialVersion)
	if err != nil {
		return err
	}
	affected, err := repository.queries.UpgradeSlackCredential(ctx, slacksql.UpgradeSlackCredentialParams{
		ReplacementCredential: encrypted, ReplacementVersion: replacementVersion,
		SlackWorkspaceID: record.SlackWorkspaceID, PreviousVersion: previousVersion,
		PreviousCredential: record.Credential,
	})
	if err != nil {
		return mapDatabaseError(err)
	}
	if affected != 1 {
		return slackdomain.ErrNotFound
	}
	return nil
}

func (repository *Repo) ScrubVersionedLegacySlackCredentials(ctx context.Context, limit int) (int, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	queryLimit, err := safecast.Int32(limit)
	if err != nil {
		return 0, err
	}
	affected, err := repository.queries.ScrubVersionedLegacySlackCredentials(ctx, slacksql.ScrubVersionedLegacySlackCredentialsParams{ResultLimit: queryLimit})
	if err != nil {
		return 0, mapDatabaseError(err)
	}
	return safecast.Int64(affected)
}

func (repository *Repo) ListLegacySlackCredentials(ctx context.Context, limit int) ([]LegacySlackCredentialRecord, error) {
	queryLimit, currentVersion, err := normalizedLegacyBounds(limit)
	if err != nil {
		return nil, err
	}
	rows, err := repository.queries.ListLegacySlackCredentials(ctx, slacksql.ListLegacySlackCredentialsParams{CurrentVersion: currentVersion, ResultLimit: queryLimit})
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	result := make([]LegacySlackCredentialRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, slackdomain.LegacyInstallationCredential{
			SlackWorkspaceID: row.ID, WorkspaceID: row.WorkspaceID,
			SlackTeamID: row.SlackTeamID, InstallGeneration: row.InstallationGeneration,
			Credential: row.Credential, CredentialVersion: int(row.CredentialKeyVersion),
		})
	}
	return result, nil
}

func (repository *Repo) ListLegacySlackUninstallCredentials(ctx context.Context, limit int) ([]LegacySlackUninstallCredentialRecord, error) {
	queryLimit, currentVersion, err := normalizedLegacyBounds(limit)
	if err != nil {
		return nil, err
	}
	rows, err := repository.queries.ListLegacySlackUninstallCredentials(ctx, slacksql.ListLegacySlackUninstallCredentialsParams{CurrentVersion: currentVersion, ResultLimit: queryLimit})
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	result := make([]LegacySlackUninstallCredentialRecord, 0, len(rows))
	for _, row := range rows {
		credential := ""
		if row.CredentialPayload != nil {
			credential = *row.CredentialPayload
		}
		result = append(result, slackdomain.LegacyUninstallCredential{
			UninstallID: row.ID, WorkspaceID: row.WorkspaceID,
			SlackTeamID: row.SlackTeamID, InstallGeneration: row.InstallationGeneration,
			Credential: credential, CredentialVersion: int(row.CredentialKeyVersion),
		})
	}
	return result, nil
}

func (repository *Repo) UpgradeSlackUninstallCredential(ctx context.Context, record LegacySlackUninstallCredentialRecord, encrypted string, version int) error {
	if record.UninstallID == uuid.Nil || strings.TrimSpace(record.Credential) == "" ||
		version != credentialvault.CurrentVersion || !credentialvault.IsEnvelope(encrypted) {
		return errors.Join(slackdomain.ErrInvalidInput, errors.New("slack uninstall credential upgrade metadata is invalid"))
	}
	replacementVersion, err := safecast.Int16(version)
	if err != nil {
		return err
	}
	previousVersion, err := safecast.Int16(record.CredentialVersion)
	if err != nil {
		return err
	}
	affected, err := repository.queries.UpgradeSlackUninstallCredential(ctx, slacksql.UpgradeSlackUninstallCredentialParams{
		ReplacementCredential: encrypted, ReplacementVersion: replacementVersion,
		UninstallID: record.UninstallID, PreviousVersion: previousVersion,
		PreviousCredential: record.Credential,
	})
	if err != nil {
		return mapDatabaseError(err)
	}
	if affected != 1 {
		return slackdomain.ErrNotFound
	}
	return nil
}

func (repository *Repo) ListLegacySlackWebhookPayloads(ctx context.Context, afterID uuid.UUID, limit int) ([]webhooks.Record, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	queryLimit, err := safecast.Int32(limit)
	if err != nil {
		return nil, err
	}
	rows, err := repository.queries.ListLegacySlackWebhookPayloads(ctx, slacksql.ListLegacySlackWebhookPayloadsParams{
		CurrentPrefix: slackWebhookPayloadEnvelopePrefix, AfterID: afterID, ResultLimit: queryLimit,
	})
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	result := make([]webhooks.Record, 0, len(rows))
	for _, row := range rows {
		if row.WorkspaceID == nil || row.InstallationID == nil || row.InstallationGeneration == nil {
			return nil, errors.Join(slackdomain.ErrConflict, errors.New("legacy Slack webhook row is missing durable binding"))
		}
		result = append(result, webhooks.Record{
			Envelope: webhooks.Envelope{
				Provider: integrations.ProviderKey(row.Provider), DeliveryID: row.ExternalEventID,
				EventType: row.EventType, ExternalAccountID: row.ExternalWorkspaceID,
				WorkspaceID: *row.WorkspaceID, InstallationID: *row.InstallationID,
				InstallationGeneration: *row.InstallationGeneration,
				TraceID:                row.TraceID, ReceivedAt: row.ReceivedAt,
			},
			ID: row.ID, Status: webhooks.Status(row.Status), AttemptCount: row.AttemptCount,
			RecoveryGeneration: row.RecoveryGeneration, RecoveryEnqueuedAt: row.RecoveryEnqueuedAt,
			ProcessedAt: row.ProcessedAt, UpdatedAt: row.UpdatedAt,
			EncryptedPayload: row.PayloadEncrypted, PayloadExpiresAt: row.PayloadExpiresAt,
		})
	}
	return result, nil
}

func (repository *Repo) UpgradeLegacySlackWebhookPayload(
	ctx context.Context,
	record webhooks.Record,
	previousPayload, replacementPayload string,
) error {
	affected, err := repository.queries.UpgradeLegacySlackWebhookPayload(ctx, slacksql.UpgradeLegacySlackWebhookPayloadParams{
		ReplacementPayload: replacementPayload, EventID: record.ID,
		DeliveryID: record.DeliveryID, ExternalAccountID: record.ExternalAccountID,
		WorkspaceID: record.WorkspaceID, InstallationID: record.InstallationID,
		InstallationGeneration: record.InstallationGeneration, PreviousPayload: previousPayload,
	})
	if err != nil {
		return mapDatabaseError(err)
	}
	if affected != 1 {
		return slackdomain.ErrNotFound
	}
	return nil
}

func credentialQueryBounds(limit int) (int32, int16, error) {
	queryLimit, err := safecast.Int32(limit)
	if err != nil {
		return 0, 0, err
	}
	version, err := safecast.Int16(credentialvault.CurrentVersion)
	return queryLimit, version, err
}

func normalizedLegacyBounds(limit int) (int32, int16, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return credentialQueryBounds(limit)
}
