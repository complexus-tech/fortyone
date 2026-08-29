package slackrepository

import (
	"fmt"
	"time"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	slacksql "github.com/complexus-tech/projects-api/internal/modules/slack/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
)

type installationFields struct {
	id                       uuid.UUID
	workspaceID              uuid.UUID
	slackTeamID              string
	slackTeamName            string
	slackTeamDomain          string
	botUserID                *string
	credentialPayload        string
	credentialKeyVersion     int16
	installationGeneration   uuid.UUID
	installationAuthorizedAt time.Time
	slackAppID               *string
	enterpriseID             *string
	authedUserID             *string
	scope                    *string
	isActive                 bool
	installedByUserID        *uuid.UUID
	revokedAt                *time.Time
	createdAt                time.Time
	updatedAt                time.Time
}

func mapInstallation(fields installationFields) slackdomain.Installation {
	return slackdomain.Installation{
		ID: fields.id, WorkspaceID: fields.workspaceID,
		SlackTeamID: fields.slackTeamID, SlackTeamName: fields.slackTeamName,
		SlackTeamDomain: fields.slackTeamDomain, BotUserID: fields.botUserID,
		BotAccessToken: fields.credentialPayload, CredentialVersion: int(fields.credentialKeyVersion),
		InstallGeneration: fields.installationGeneration, AuthorizedAt: fields.installationAuthorizedAt,
		SlackAppID: fields.slackAppID, EnterpriseID: fields.enterpriseID,
		AuthedUserID: fields.authedUserID, Scope: fields.scope, IsActive: fields.isActive,
		InstalledByUserID: fields.installedByUserID, RevokedAt: fields.revokedAt,
		CreatedAt: fields.createdAt, UpdatedAt: fields.updatedAt,
	}
}

func installationFromWorkspaceRow(row slacksql.GetSlackWorkspaceRow) slackdomain.Installation {
	return mapInstallation(installationFields{
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
}

func installationFromTeamRow(row slacksql.GetSlackWorkspaceByTeamIDRow) slackdomain.Installation {
	return mapInstallation(installationFields{
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
}

func mapUninstall(row slacksql.SlackUninstallOutbox) (slackdomain.Uninstall, error) {
	attemptCount, err := safecast.Int64(int64(row.AttemptCount))
	if err != nil {
		return slackdomain.Uninstall{}, fmt.Errorf("map Slack uninstall attempt count: %w", err)
	}
	credential := ""
	if row.CredentialPayload != nil {
		credential = *row.CredentialPayload
	}
	return slackdomain.Uninstall{
		ID: row.ID, SlackWorkspaceID: row.SlackWorkspaceID, WorkspaceID: row.WorkspaceID,
		InstallGeneration: row.InstallationGeneration, SlackTeamID: row.SlackTeamID,
		UninstallKind: slackdomain.UninstallKind(row.UninstallKind), CredentialPayload: credential,
		CredentialKeyVersion: int(row.CredentialKeyVersion), Status: slackdomain.UninstallStatus(row.Status),
		AttemptCount: attemptCount, LastError: row.LastError, NextAttemptAt: row.NextAttemptAt,
		ProcessingStartedAt: row.ProcessingStartedAt, CompletedAt: row.CompletedAt,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}, nil
}

func mapRequestLog(row slacksql.SlackRequestLog) (slackdomain.RequestLog, error) {
	responseCode, err := safecast.Int64(int64(row.ResponseCode))
	if err != nil {
		return slackdomain.RequestLog{}, fmt.Errorf("map Slack response code: %w", err)
	}
	return slackdomain.RequestLog{
		ID: row.ID, RequestType: row.RequestType, Endpoint: row.Endpoint,
		WorkspaceID: row.WorkspaceID, SlackTeamID: row.SlackTeamID,
		SlackUserID: row.SlackUserID, SlackChannel: row.SlackChannelID,
		Command: row.Command, TriggerID: row.TriggerID, RequestBody: row.RequestBody,
		Headers: row.Headers, ResponseCode: responseCode, Outcome: row.Outcome,
		ErrorMessage: row.ErrorMessage, CreatedAt: row.CreatedAt,
	}, nil
}
