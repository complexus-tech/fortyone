package figmarepository

import (
	"context"

	figmadomain "github.com/complexus-tech/projects-api/internal/modules/figma/domain"
	figmasql "github.com/complexus-tech/projects-api/internal/modules/figma/repository/sqlc"
	"github.com/google/uuid"
)

func (repository *Repository) SaveWebhook(
	ctx context.Context,
	webhook figmadomain.Webhook,
) error {
	rows, err := repository.queries.SaveWebhook(ctx, figmasql.SaveWebhookParams{
		ConnectionID: webhook.ConnectionID, FileKey: webhook.FileKey,
		EventType: webhook.EventType, FigmaWebhookID: webhook.FigmaWebhookID,
		PasscodeHash: webhook.PasscodeHash, UpdatedAt: repository.currentTime(),
	})
	return requireAffected(rows, err, figmadomain.ErrNotFound)
}

func (repository *Repository) GetWebhook(
	ctx context.Context,
	figmaWebhookID int64,
) (figmadomain.Webhook, error) {
	row, err := repository.queries.GetAuthorizedWebhook(
		ctx,
		figmasql.GetAuthorizedWebhookParams{FigmaWebhookID: figmaWebhookID},
	)
	if err != nil {
		return figmadomain.Webhook{}, mapDatabaseError(err)
	}
	return figmadomain.Webhook{
		ID: row.ID, ConnectionID: row.ConnectionID, WorkspaceID: row.WorkspaceID,
		InstallationGeneration: row.InstallationGeneration, FileKey: row.FileKey,
		EventType: row.EventType, FigmaWebhookID: row.FigmaWebhookID,
		PasscodeHash: row.PasscodeHash, IsActive: row.IsActive,
	}, nil
}

func (repository *Repository) GetCurrentWebhook(
	ctx context.Context,
	connectionID, installationGeneration uuid.UUID,
	figmaWebhookID int64,
) (figmadomain.Webhook, error) {
	row, err := repository.queries.GetCurrentWebhook(ctx, figmasql.GetCurrentWebhookParams{
		ConnectionID: connectionID, InstallationGeneration: installationGeneration,
		FigmaWebhookID: figmaWebhookID,
	})
	if err != nil {
		return figmadomain.Webhook{}, mapDatabaseError(err)
	}
	return figmadomain.Webhook{
		ID: row.ID, ConnectionID: row.ConnectionID, WorkspaceID: row.WorkspaceID,
		InstallationGeneration: row.InstallationGeneration, FileKey: row.FileKey,
		EventType: row.EventType, FigmaWebhookID: row.FigmaWebhookID,
		PasscodeHash: row.PasscodeHash, IsActive: row.IsActive,
	}, nil
}

func (repository *Repository) FindWebhook(
	ctx context.Context,
	connectionID uuid.UUID,
	fileKey, eventType string,
) (figmadomain.Webhook, error) {
	row, err := repository.queries.FindWebhook(ctx, figmasql.FindWebhookParams{
		ConnectionID: connectionID, FileKey: fileKey, EventType: eventType,
	})
	if err != nil {
		return figmadomain.Webhook{}, mapDatabaseError(err)
	}
	return figmadomain.Webhook{
		ID: row.ID, ConnectionID: row.ConnectionID, FileKey: row.FileKey,
		EventType: row.EventType, FigmaWebhookID: row.FigmaWebhookID,
		PasscodeHash: row.PasscodeHash, IsActive: row.IsActive,
	}, nil
}

func (repository *Repository) ListWebhooks(
	ctx context.Context,
	connectionID uuid.UUID,
) ([]figmadomain.Webhook, error) {
	rows, err := repository.queries.ListWebhooks(ctx, figmasql.ListWebhooksParams{
		ConnectionID: connectionID,
	})
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	result := make([]figmadomain.Webhook, 0, len(rows))
	for _, row := range rows {
		result = append(result, figmadomain.Webhook{
			ID: row.ID, ConnectionID: row.ConnectionID, FileKey: row.FileKey,
			EventType: row.EventType, FigmaWebhookID: row.FigmaWebhookID,
			PasscodeHash: row.PasscodeHash, IsActive: row.IsActive,
		})
	}
	return result, nil
}

func (repository *Repository) DeactivateWebhook(
	ctx context.Context,
	figmaWebhookID int64,
) error {
	_, err := repository.queries.DeactivateWebhook(ctx, figmasql.DeactivateWebhookParams{
		FigmaWebhookID: figmaWebhookID, UpdatedAt: repository.currentTime(),
	})
	return mapDatabaseError(err)
}
