package slackrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	slacksql "github.com/complexus-tech/projects-api/internal/modules/slack/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repo) UpsertSlackWorkspace(
	ctx context.Context,
	workspaceID, installedByUserID uuid.UUID,
	payload OAuthInstallPayload,
) (SlackWorkspaceRecord, error) {
	payload.SlackTeamID = strings.TrimSpace(payload.SlackTeamID)
	if workspaceID == uuid.Nil || installedByUserID == uuid.Nil || payload.SlackTeamID == "" || payload.InstallGeneration == uuid.Nil {
		return SlackWorkspaceRecord{}, errors.Join(slackdomain.ErrInvalidInput, errors.New("workspace, actor, Slack team, and generation are required"))
	}
	if payload.CredentialVersion != credentialvault.CurrentVersion || !credentialvault.IsEnvelope(payload.BotAccessToken) {
		return SlackWorkspaceRecord{}, errors.Join(slackdomain.ErrInvalidInput, errors.New("slack credential must use the current credential vault envelope"))
	}
	credentialVersion, err := safecast.Int16(payload.CredentialVersion)
	if err != nil {
		return SlackWorkspaceRecord{}, errors.Join(slackdomain.ErrInvalidInput, err)
	}
	var result SlackWorkspaceRecord
	err = repository.withinTransaction(ctx, func(queries slacksql.Querier) error {
		if err := lockInstallationLifecycle(ctx, queries); err != nil {
			return err
		}
		if _, err := queries.LockSlackWorkspaceAdmin(ctx, slacksql.LockSlackWorkspaceAdminParams{
			WorkspaceID: workspaceID,
			ActorID:     installedByUserID,
		}); errors.Is(err, pgx.ErrNoRows) {
			return slackdomain.ErrForbidden
		} else if err != nil {
			return fmt.Errorf("authorize Slack installation actor: %w", err)
		}

		active, err := queries.ListActiveInstallationsForUpdate(ctx, slacksql.ListActiveInstallationsForUpdateParams{
			WorkspaceID: workspaceID,
			SlackTeamID: payload.SlackTeamID,
		})
		if err != nil {
			return fmt.Errorf("lock active Slack installations: %w", err)
		}
		refreshing := false
		previousGeneration := uuid.Nil
		for _, installation := range active {
			switch {
			case installation.WorkspaceID == workspaceID && installation.SlackTeamID == payload.SlackTeamID:
				refreshing = true
				previousGeneration = installation.InstallationGeneration
			case installation.SlackTeamID == payload.SlackTeamID:
				return errors.Join(ErrActiveInstallationConflict, ErrSlackTeamAlreadyConnected)
			case installation.WorkspaceID == workspaceID:
				return errors.Join(ErrActiveInstallationConflict, ErrWorkspaceAlreadyConnected)
			}
		}
		uninstallState, err := queries.GetSlackUninstallState(ctx, slacksql.GetSlackUninstallStateParams{SlackTeamID: payload.SlackTeamID})
		if err != nil {
			return fmt.Errorf("read Slack uninstall state: %w", err)
		}
		if uninstallState.Processing {
			return ErrUninstallInProgress
		}
		if uninstallState.ResolutionRequired {
			return ErrUninstallResolutionRequired
		}
		if _, err := queries.CompleteSupersededSlackUninstalls(ctx, slacksql.CompleteSupersededSlackUninstallsParams{SlackTeamID: payload.SlackTeamID}); err != nil {
			return fmt.Errorf("complete superseded Slack uninstalls: %w", err)
		}
		if _, err := queries.DeleteInactiveSlackInstallations(ctx, slacksql.DeleteInactiveSlackInstallationsParams{
			WorkspaceID: workspaceID,
			SlackTeamID: payload.SlackTeamID,
		}); err != nil {
			return fmt.Errorf("delete inactive Slack installations: %w", err)
		}
		if refreshing {
			if err := cancelSlackMessaging(ctx, queries, payload.SlackTeamID, "Slack installation refreshed"); err != nil {
				return err
			}
		}
		row, err := queries.UpsertSlackInstallation(ctx, slacksql.UpsertSlackInstallationParams{
			WorkspaceID: workspaceID, SlackTeamID: payload.SlackTeamID,
			SlackTeamName:   strings.TrimSpace(payload.SlackTeamName),
			SlackTeamDomain: strings.TrimSpace(payload.SlackTeamDomain),
			BotUserID:       payload.BotUserID, CredentialPayload: payload.BotAccessToken,
			CredentialKeyVersion: credentialVersion, InstallationGeneration: payload.InstallGeneration,
			SlackAppID: payload.SlackAppID, EnterpriseID: payload.EnterpriseID,
			AuthedUserID: payload.AuthedUserID, Scope: payload.Scope,
			InstalledByUserID: installedByUserID,
		})
		if err != nil {
			return fmt.Errorf("upsert Slack installation: %w", err)
		}
		if refreshing && previousGeneration != row.InstallationGeneration {
			if _, err := queries.RebindSlackRequestThreads(ctx, slacksql.RebindSlackRequestThreadsParams{
				WorkspaceID: workspaceID, SlackTeamID: payload.SlackTeamID,
				PreviousGeneration: previousGeneration, CurrentGeneration: row.InstallationGeneration,
			}); err != nil {
				return fmt.Errorf("rebind Slack request threads: %w", err)
			}
		}
		if payload.AuthedUserID != nil && strings.TrimSpace(*payload.AuthedUserID) != "" {
			affected, err := queries.UpsertSlackOAuthInstallerLink(ctx, slacksql.UpsertSlackOAuthInstallerLinkParams{
				WorkspaceID: workspaceID, SlackWorkspaceID: row.ID, SlackTeamID: payload.SlackTeamID,
				SlackUserID: strings.TrimSpace(*payload.AuthedUserID), ActorID: installedByUserID,
			})
			if err != nil {
				return fmt.Errorf("link Slack OAuth installer: %w", err)
			}
			if affected != 1 {
				return slackdomain.ErrForbidden
			}
		}
		result = mapInstallation(installationFields{
			id: row.ID, workspaceID: row.WorkspaceID, slackTeamID: row.SlackTeamID,
			slackTeamName: row.SlackTeamName, slackTeamDomain: row.SlackTeamDomain,
			botUserID: row.BotUserID, credentialPayload: row.CredentialPayload,
			credentialKeyVersion:     row.CredentialKeyVersion,
			installationGeneration:   row.InstallationGeneration,
			installationAuthorizedAt: row.InstallationAuthorizedAt,
			slackAppID:               row.SlackAppID, enterpriseID: row.EnterpriseID,
			authedUserID: row.AuthedUserID, scope: row.Scope, isActive: row.IsActive,
			installedByUserID: row.InstalledByUserID, revokedAt: row.RevokedAt,
			createdAt: row.CreatedAt, updatedAt: row.UpdatedAt,
		})
		return nil
	})
	return result, err
}

func (repository *Repo) GetSlackWorkspace(ctx context.Context, workspaceID uuid.UUID) (SlackWorkspaceRecord, error) {
	row, err := repository.queries.GetSlackWorkspace(ctx, slacksql.GetSlackWorkspaceParams{WorkspaceID: workspaceID})
	if err != nil {
		return SlackWorkspaceRecord{}, mapDatabaseError(err)
	}
	return installationFromWorkspaceRow(row), nil
}

func (repository *Repo) GetSlackWorkspaceByTeamID(ctx context.Context, slackTeamID string) (SlackWorkspaceRecord, error) {
	row, err := repository.queries.GetSlackWorkspaceByTeamID(ctx, slacksql.GetSlackWorkspaceByTeamIDParams{SlackTeamID: strings.TrimSpace(slackTeamID)})
	if err != nil {
		return SlackWorkspaceRecord{}, mapDatabaseError(err)
	}
	return installationFromTeamRow(row), nil
}

func (repository *Repo) DisconnectSlackWorkspace(
	ctx context.Context,
	command slackdomain.DisconnectInstallationCommand,
) (SlackUninstallRecord, error) {
	if err := command.Validate(); err != nil {
		return SlackUninstallRecord{}, err
	}
	var result SlackUninstallRecord
	err := repository.withinTransaction(ctx, func(queries slacksql.Querier) error {
		if err := lockInstallationLifecycle(ctx, queries); err != nil {
			return err
		}
		if _, err := queries.LockSlackWorkspaceAdmin(ctx, slacksql.LockSlackWorkspaceAdminParams{
			WorkspaceID: command.WorkspaceID, ActorID: command.ActorID,
		}); errors.Is(err, pgx.ErrNoRows) {
			return slackdomain.ErrForbidden
		} else if err != nil {
			return err
		}
		installation, err := queries.LockSlackInstallationForDisconnect(ctx, slacksql.LockSlackInstallationForDisconnectParams{WorkspaceID: command.WorkspaceID})
		if err != nil {
			return err
		}
		if int(installation.CredentialKeyVersion) != credentialvault.CurrentVersion || !credentialvault.IsEnvelope(installation.CredentialPayload) {
			return errors.Join(slackdomain.ErrConflict, errors.New("slack installation credential must use the current vault envelope before disconnect"))
		}
		row, err := queries.EnqueueSlackUninstall(ctx, slacksql.EnqueueSlackUninstallParams{
			SlackWorkspaceID: installation.ID, WorkspaceID: installation.WorkspaceID,
			InstallationGeneration: installation.InstallationGeneration,
			SlackTeamID:            installation.SlackTeamID, UninstallKind: string(slackdomain.UninstallDisconnect),
			CredentialPayload:    installation.CredentialPayload,
			CredentialKeyVersion: installation.CredentialKeyVersion,
		})
		if err != nil {
			return err
		}
		if err := cancelSlackMessaging(ctx, queries, installation.SlackTeamID, "Slack installation disconnected"); err != nil {
			return err
		}
		if _, err := queries.DeleteSlackUserLinksByWorkspace(ctx, slacksql.DeleteSlackUserLinksByWorkspaceParams{WorkspaceID: command.WorkspaceID}); err != nil {
			return err
		}
		if affected, err := queries.DeleteSlackInstallationByID(ctx, slacksql.DeleteSlackInstallationByIDParams{SlackWorkspaceID: installation.ID}); err != nil {
			return err
		} else if affected != 1 {
			return slackdomain.ErrConflict
		}
		result, err = mapUninstall(row)
		return err
	})
	return result, err
}

func lockInstallationLifecycle(ctx context.Context, queries slacksql.Querier) error {
	return queries.LockSlackInstallationLifecycle(ctx, slacksql.LockSlackInstallationLifecycleParams{LockKey: SlackInstallationLifecycleAdvisoryKey})
}

func cancelSlackMessaging(ctx context.Context, queries slacksql.Querier, slackTeamID, reason string) error {
	if _, err := queries.CancelSlackInboundEvents(ctx, slacksql.CancelSlackInboundEventsParams{SlackTeamID: slackTeamID, Reason: reason}); err != nil {
		return fmt.Errorf("cancel Slack inbound events: %w", err)
	}
	if _, err := queries.CancelSlackOutboundDeliveries(ctx, slacksql.CancelSlackOutboundDeliveriesParams{SlackTeamID: slackTeamID, Reason: reason}); err != nil {
		return fmt.Errorf("cancel Slack outbound deliveries: %w", err)
	}
	return nil
}
