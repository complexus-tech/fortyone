package githubrepository

import (
	"context"
	"errors"

	githubsql "github.com/complexus-tech/projects-api/internal/modules/github/repository/sqlc"
	githubshared "github.com/complexus-tech/projects-api/internal/modules/github/shared"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repo) GetAuthorizedWebhookInstallation(
	ctx context.Context,
	externalInstallationID, externalRepositoryID int64,
) (githubshared.WebhookInstallation, error) {
	if r == nil || r.queries == nil {
		return githubshared.WebhookInstallation{}, errors.New("GitHub native repository is not configured")
	}
	row, err := r.queries.GetAuthorizedWebhookInstallation(ctx, githubsql.GetAuthorizedWebhookInstallationParams{
		GithubInstallationID: externalInstallationID,
		GithubRepositoryID:   externalRepositoryID,
	})
	if err != nil {
		return githubshared.WebhookInstallation{}, mapWebhookInstallationError(err)
	}
	return mapAuthorizedWebhookInstallation(row), nil
}

func (r *Repo) GetCurrentWebhookInstallation(
	ctx context.Context,
	installationID, installationGeneration uuid.UUID,
	externalRepositoryID int64,
) (githubshared.WebhookInstallation, error) {
	if r == nil || r.queries == nil {
		return githubshared.WebhookInstallation{}, errors.New("GitHub native repository is not configured")
	}
	row, err := r.queries.GetCurrentWebhookInstallation(ctx, githubsql.GetCurrentWebhookInstallationParams{
		InstallationID:         installationID,
		InstallationGeneration: installationGeneration,
		GithubRepositoryID:     externalRepositoryID,
	})
	if err != nil {
		return githubshared.WebhookInstallation{}, mapWebhookInstallationError(err)
	}
	return githubshared.WebhookInstallation{
		ID:                     row.ID,
		WorkspaceID:            row.WorkspaceID,
		ExternalInstallationID: row.GithubInstallationID,
		InstallationGeneration: row.InstallationGeneration,
		RepositoryID:           row.RepositoryID,
		ExternalRepositoryID:   row.GithubRepositoryID,
	}, nil
}

func mapAuthorizedWebhookInstallation(row githubsql.GetAuthorizedWebhookInstallationRow) githubshared.WebhookInstallation {
	return githubshared.WebhookInstallation{
		ID:                     row.ID,
		WorkspaceID:            row.WorkspaceID,
		ExternalInstallationID: row.GithubInstallationID,
		InstallationGeneration: row.InstallationGeneration,
		RepositoryID:           row.RepositoryID,
		ExternalRepositoryID:   row.GithubRepositoryID,
	}
}

func mapWebhookInstallationError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return githubshared.ErrWebhookInstallationNotFound
	}
	return err
}
