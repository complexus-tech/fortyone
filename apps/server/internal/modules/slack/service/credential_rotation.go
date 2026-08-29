package slack

import (
	"context"
	"errors"
	"fmt"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
)

type credentialRewrapRepository interface {
	ListSlackCredentialsForRewrap(ctx context.Context, after *uuid.UUID, limit int) ([]slackdomain.InstallationCredentialForRewrap, error)
	RewrapSlackCredential(ctx context.Context, record slackdomain.InstallationCredentialForRewrap, rewrapped string) (bool, error)
	ListSlackUninstallCredentialsForRewrap(ctx context.Context, after *uuid.UUID, limit int) ([]slackdomain.UninstallCredentialForRewrap, error)
	RewrapSlackUninstallCredential(ctx context.Context, record slackdomain.UninstallCredentialForRewrap, rewrapped string) (bool, error)
}

// RewrapCredentials authenticates and rewraps active Slack installation and
// retained uninstall credentials under the active vault key. Installation
// generation is preserved and included in every compare-and-swap, so a
// concurrent refresh, reinstall, disconnect, revoke, or uninstall completion
// always wins over stale maintenance.
func (p *EventProcessor) RewrapCredentials(ctx context.Context) (credentialvault.RotationReport, error) {
	if p == nil || p.codec == nil || p.codec.vault == nil {
		return credentialvault.RotationReport{}, credentialvault.ErrNotConfigured
	}
	repo, ok := p.repo.(credentialRewrapRepository)
	if !ok {
		return credentialvault.RotationReport{}, errors.New("slack repository does not support credential rewrap")
	}
	activeKey, err := p.codec.vault.ActiveKeyRef()
	if err != nil {
		return credentialvault.RotationReport{}, err
	}
	report := credentialvault.RotationReport{ActiveKey: activeKey}
	if err := p.rewrapInstallations(ctx, repo, &report); err != nil {
		return report, err
	}
	if err := p.rewrapUninstalls(ctx, repo, &report); err != nil {
		return report, err
	}
	return report, nil
}

func (p *EventProcessor) rewrapInstallations(
	ctx context.Context,
	repo credentialRewrapRepository,
	report *credentialvault.RotationReport,
) error {
	var after *uuid.UUID
	for {
		records, err := repo.ListSlackCredentialsForRewrap(
			ctx,
			after,
			credentialvault.DefaultMaintenanceBatchSize,
		)
		if err != nil {
			return fmt.Errorf("list Slack installation credentials for rewrap: %w", err)
		}
		for _, record := range records {
			report.Scanned++
			binding := slackCredentialBinding{
				WorkspaceID:       record.WorkspaceID,
				SlackTeamID:       record.SlackTeamID,
				InstallGeneration: record.InstallGeneration,
			}
			result, err := p.codec.vault.Rewrap(binding.vaultContext(), record.Credential)
			if err != nil {
				return fmt.Errorf("rewrap Slack installation credential %s: %w", record.SlackWorkspaceID, err)
			}
			if !result.Changed {
				report.Current++
				continue
			}
			replaced, err := repo.RewrapSlackCredential(ctx, record, result.Envelope)
			if err != nil {
				return fmt.Errorf("persist Slack installation credential rewrap %s: %w", record.SlackWorkspaceID, err)
			}
			if replaced {
				report.Rewrapped++
			} else {
				report.Stale++
			}
		}
		if len(records) < credentialvault.DefaultMaintenanceBatchSize {
			return nil
		}
		cursor := records[len(records)-1].SlackWorkspaceID
		after = &cursor
	}
}

func (p *EventProcessor) rewrapUninstalls(
	ctx context.Context,
	repo credentialRewrapRepository,
	report *credentialvault.RotationReport,
) error {
	var after *uuid.UUID
	for {
		records, err := repo.ListSlackUninstallCredentialsForRewrap(
			ctx,
			after,
			credentialvault.DefaultMaintenanceBatchSize,
		)
		if err != nil {
			return fmt.Errorf("list Slack uninstall credentials for rewrap: %w", err)
		}
		for _, record := range records {
			report.Scanned++
			binding := slackCredentialBinding{
				WorkspaceID:       record.WorkspaceID,
				SlackTeamID:       record.SlackTeamID,
				InstallGeneration: record.InstallGeneration,
			}
			result, err := p.codec.vault.Rewrap(binding.vaultContext(), record.Credential)
			if err != nil {
				return fmt.Errorf("rewrap Slack uninstall credential %s: %w", record.UninstallID, err)
			}
			if !result.Changed {
				report.Current++
				continue
			}
			replaced, err := repo.RewrapSlackUninstallCredential(ctx, record, result.Envelope)
			if err != nil {
				return fmt.Errorf("persist Slack uninstall credential rewrap %s: %w", record.UninstallID, err)
			}
			if replaced {
				report.Rewrapped++
			} else {
				report.Stale++
			}
		}
		if len(records) < credentialvault.DefaultMaintenanceBatchSize {
			return nil
		}
		cursor := records[len(records)-1].UninstallID
		after = &cursor
	}
}
