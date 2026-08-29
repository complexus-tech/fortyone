package slack

import (
	"context"
	"errors"
	"fmt"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
)

const legacyCredentialBatchSize = 100

type legacyCredentialRepository interface {
	ListLegacySlackCredentials(ctx context.Context, limit int) ([]slackdomain.LegacyInstallationCredential, error)
	UpgradeSlackCredential(ctx context.Context, record slackdomain.LegacyInstallationCredential, encrypted string, version int) error
	ScrubVersionedLegacySlackCredentials(ctx context.Context, limit int) (int, error)
	ListLegacySlackUninstallCredentials(ctx context.Context, limit int) ([]slackdomain.LegacyUninstallCredential, error)
	UpgradeSlackUninstallCredential(ctx context.Context, record slackdomain.LegacyUninstallCredential, encrypted string, version int) error
}

// BackfillLegacyCredentials encrypts bounded batches of pre-migration Slack
// tokens. Each replacement is an atomic compare-and-swap, and normal provider
// operations remain fail-closed until every active credential is migrated.
func (p *EventProcessor) BackfillLegacyCredentials(ctx context.Context, cutover *LegacyCutover) (int, error) {
	repo, ok := p.repo.(legacyCredentialRepository)
	if !ok {
		return 0, errors.New("slack repository does not support credential backfill")
	}
	if p.codec == nil {
		return 0, errors.New("slack credential encryption is not configured")
	}
	if cutover == nil {
		return 0, errors.New("slack legacy credential cutover is not configured")
	}

	updated := 0
	for {
		records, err := repo.ListLegacySlackCredentials(ctx, legacyCredentialBatchSize)
		if err != nil {
			return updated, fmt.Errorf("list legacy Slack credentials: %w", err)
		}
		for _, record := range records {
			credential, _, err := cutover.openCredential(record.Credential)
			if err != nil {
				return updated, fmt.Errorf("open legacy Slack credential %s: %w", record.SlackWorkspaceID, err)
			}
			binding := slackCredentialBinding{
				WorkspaceID:       record.WorkspaceID,
				SlackTeamID:       record.SlackTeamID,
				InstallGeneration: record.InstallGeneration,
			}
			encrypted, currentVersion, err := p.codec.seal(binding, credential)
			if err != nil {
				return updated, fmt.Errorf("seal legacy Slack credential %s: %w", record.SlackWorkspaceID, err)
			}
			if err := repo.UpgradeSlackCredential(ctx, record, encrypted, currentVersion); err != nil {
				if isSlackRepositoryNotFound(err) {
					continue
				}
				return updated, fmt.Errorf("upgrade legacy Slack credential %s: %w", record.SlackWorkspaceID, err)
			}
			updated++
		}
		if len(records) < legacyCredentialBatchSize {
			break
		}
	}
	for {
		records, err := repo.ListLegacySlackUninstallCredentials(ctx, legacyCredentialBatchSize)
		if err != nil {
			return updated, fmt.Errorf("list legacy Slack uninstall credentials: %w", err)
		}
		for _, record := range records {
			credential, _, err := cutover.openCredential(record.Credential)
			if err != nil {
				return updated, fmt.Errorf("open legacy Slack uninstall credential %s: %w", record.UninstallID, err)
			}
			binding := slackCredentialBinding{
				WorkspaceID:       record.WorkspaceID,
				SlackTeamID:       record.SlackTeamID,
				InstallGeneration: record.InstallGeneration,
			}
			encrypted, currentVersion, err := p.codec.seal(binding, credential)
			if err != nil {
				return updated, fmt.Errorf("seal legacy Slack uninstall credential %s: %w", record.UninstallID, err)
			}
			if err := repo.UpgradeSlackUninstallCredential(ctx, record, encrypted, currentVersion); err != nil {
				if isSlackRepositoryNotFound(err) {
					continue
				}
				return updated, fmt.Errorf("upgrade legacy Slack uninstall credential %s: %w", record.UninstallID, err)
			}
			updated++
		}
		if len(records) < legacyCredentialBatchSize {
			break
		}
	}
	for {
		scrubbed, err := repo.ScrubVersionedLegacySlackCredentials(ctx, legacyCredentialBatchSize)
		if err != nil {
			return updated, fmt.Errorf("scrub versioned legacy Slack credentials: %w", err)
		}
		updated += scrubbed
		if scrubbed < legacyCredentialBatchSize {
			break
		}
	}
	return updated, nil
}
