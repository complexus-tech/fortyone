package figmarepository

import (
	"context"
	"time"

	figmadomain "github.com/complexus-tech/projects-api/internal/modules/figma/domain"
	figmasql "github.com/complexus-tech/projects-api/internal/modules/figma/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
)

func (repository *Repository) UpsertConnection(
	ctx context.Context,
	connection figmadomain.Connection,
) (figmadomain.Connection, error) {
	var result figmadomain.Connection
	err := repository.withinTransaction(ctx, func(queries figmasql.Querier) error {
		now := repository.currentTime()
		if _, err := queries.DeactivateConnectionWebhooks(ctx, figmasql.DeactivateConnectionWebhooksParams{
			WorkspaceID: connection.WorkspaceID, UpdatedAt: now,
		}); err != nil {
			return err
		}
		disconnectedAt := now
		if _, err := queries.DeactivateWorkspaceConnection(ctx, figmasql.DeactivateWorkspaceConnectionParams{
			WorkspaceID: connection.WorkspaceID, DisconnectedAt: &disconnectedAt,
		}); err != nil {
			return err
		}
		row, err := queries.CreateConnection(ctx, figmasql.CreateConnectionParams{
			ID: connection.ID, WorkspaceID: connection.WorkspaceID,
			FigmaUserID: connection.FigmaUserID, FigmaEmail: connection.Email,
			FigmaHandle: connection.Handle, TokenPayload: connection.CredentialPayload,
			CredentialKeyVersion:   connection.CredentialVersion,
			InstallationGeneration: connection.InstallationGeneration,
			Scopes:                 connection.Scopes, ExpiresAt: connection.ExpiresAt.UTC(),
			ConnectedByUserID: connection.ConnectedByUserID,
		})
		if err != nil {
			return err
		}
		result = mapCreatedConnection(row)
		return nil
	})
	return result, err
}

func (repository *Repository) GetConnection(
	ctx context.Context,
	workspaceID uuid.UUID,
) (figmadomain.Connection, error) {
	row, err := repository.queries.GetActiveConnection(ctx, figmasql.GetActiveConnectionParams{
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return figmadomain.Connection{}, mapDatabaseError(err)
	}
	return mapActiveConnection(row), nil
}

func (repository *Repository) UpdateConnectionCredential(
	ctx context.Context,
	connectionID, installationGeneration uuid.UUID,
	previousPayload, nextPayload string,
	expiresAt time.Time,
) (bool, error) {
	rows, err := repository.queries.CompareAndSwapConnectionCredential(
		ctx,
		figmasql.CompareAndSwapConnectionCredentialParams{
			ID: connectionID, InstallationGeneration: installationGeneration,
			PreviousPayload: previousPayload, NextPayload: nextPayload,
			CredentialKeyVersion: int16(credentialvault.CurrentVersion),
			ExpiresAt:            expiresAt.UTC(), UpdatedAt: repository.currentTime(),
		},
	)
	if err != nil {
		return false, mapDatabaseError(err)
	}
	return rows == 1, nil
}

func (repository *Repository) Disconnect(ctx context.Context, workspaceID uuid.UUID) error {
	return repository.withinTransaction(ctx, func(queries figmasql.Querier) error {
		now := repository.currentTime()
		if _, err := queries.DeactivateConnectionWebhooks(ctx, figmasql.DeactivateConnectionWebhooksParams{
			WorkspaceID: workspaceID, UpdatedAt: now,
		}); err != nil {
			return err
		}
		disconnectedAt := now
		_, err := queries.DeactivateWorkspaceConnection(ctx, figmasql.DeactivateWorkspaceConnectionParams{
			WorkspaceID: workspaceID, DisconnectedAt: &disconnectedAt,
		})
		return err
	})
}

func (repository *Repository) ListLegacyCredentials(
	ctx context.Context,
	after *uuid.UUID,
	limit int32,
) ([]figmadomain.LegacyCredential, error) {
	rows, err := repository.queries.ListLegacyConnections(ctx, figmasql.ListLegacyConnectionsParams{
		AfterID: after, PageLimit: limit,
	})
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	records := make([]figmadomain.LegacyCredential, 0, len(rows))
	for _, row := range rows {
		records = append(records, figmadomain.LegacyCredential{
			ID: row.ID, WorkspaceID: row.WorkspaceID, Payload: row.TokenPayload,
			InstallationGeneration: row.InstallationGeneration,
		})
	}
	return records, nil
}

func (repository *Repository) UpgradeLegacyCredential(
	ctx context.Context,
	record figmadomain.LegacyCredential,
	nextPayload string,
) (bool, error) {
	rows, err := repository.queries.UpgradeLegacyConnectionCredential(
		ctx,
		figmasql.UpgradeLegacyConnectionCredentialParams{
			NextPayload:          nextPayload,
			CredentialKeyVersion: int16(credentialvault.CurrentVersion),
			UpdatedAt:            repository.currentTime(), ID: record.ID,
			InstallationGeneration: record.InstallationGeneration,
			PreviousPayload:        record.Payload,
		},
	)
	if err != nil {
		return false, mapDatabaseError(err)
	}
	return rows == 1, nil
}

func (repository *Repository) ListCredentialsForRewrap(
	ctx context.Context,
	after *uuid.UUID,
	limit int32,
) ([]figmadomain.Credential, error) {
	rows, err := repository.queries.ListConnectionsForCredentialRewrap(
		ctx,
		figmasql.ListConnectionsForCredentialRewrapParams{
			CredentialKeyVersion: int16(credentialvault.CurrentVersion),
			AfterID:              after, PageLimit: limit,
		},
	)
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	records := make([]figmadomain.Credential, 0, len(rows))
	for _, row := range rows {
		records = append(records, figmadomain.Credential{
			ID: row.ID, WorkspaceID: row.WorkspaceID, Payload: row.TokenPayload,
			CredentialVersion:      row.CredentialKeyVersion,
			InstallationGeneration: row.InstallationGeneration,
		})
	}
	return records, nil
}

func (repository *Repository) RewrapCredential(
	ctx context.Context,
	record figmadomain.Credential,
	nextPayload string,
) (bool, error) {
	rows, err := repository.queries.CompareAndSwapRewrappedCredential(
		ctx,
		figmasql.CompareAndSwapRewrappedCredentialParams{
			NextPayload: nextPayload, UpdatedAt: repository.currentTime(), ID: record.ID,
			InstallationGeneration: record.InstallationGeneration,
			CredentialKeyVersion:   record.CredentialVersion,
			PreviousPayload:        record.Payload,
		},
	)
	if err != nil {
		return false, mapDatabaseError(err)
	}
	return rows == 1, nil
}

func mapCreatedConnection(row figmasql.CreateConnectionRow) figmadomain.Connection {
	return figmadomain.Connection{
		ID: row.ID, WorkspaceID: row.WorkspaceID, FigmaUserID: row.FigmaUserID,
		Email: row.FigmaEmail, Handle: row.FigmaHandle, CredentialPayload: row.TokenPayload,
		CredentialVersion:      row.CredentialKeyVersion,
		InstallationGeneration: row.InstallationGeneration, Scopes: row.Scopes,
		ExpiresAt: row.ExpiresAt, ConnectedByUserID: row.ConnectedByUserID,
		IsActive: row.IsActive, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func mapActiveConnection(row figmasql.GetActiveConnectionRow) figmadomain.Connection {
	return figmadomain.Connection{
		ID: row.ID, WorkspaceID: row.WorkspaceID, FigmaUserID: row.FigmaUserID,
		Email: row.FigmaEmail, Handle: row.FigmaHandle, CredentialPayload: row.TokenPayload,
		CredentialVersion:      row.CredentialKeyVersion,
		InstallationGeneration: row.InstallationGeneration, Scopes: row.Scopes,
		ExpiresAt: row.ExpiresAt, ConnectedByUserID: row.ConnectedByUserID,
		IsActive: row.IsActive, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}
