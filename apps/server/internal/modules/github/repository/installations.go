package githubrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	githubsql "github.com/complexus-tech/projects-api/internal/modules/github/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrInstallationOwnershipConflict = errors.New("github installation ownership conflict")

func (r *Repo) UpsertInstallationWithRepositories(
	ctx context.Context,
	workspaceID uuid.UUID,
	installedByUserID uuid.UUID,
	appID int64,
	installation GithubInstallationPayload,
	repositories []GithubRepositoryPayload,
) error {
	if workspaceID == uuid.Nil || installedByUserID == uuid.Nil || appID <= 0 || installation.ID <= 0 {
		return errors.New("invalid GitHub installation input")
	}

	permissionsJSON, err := json.Marshal(installation.Permissions)
	if err != nil {
		return fmt.Errorf("encode GitHub installation permissions: %w", err)
	}
	eventsJSON, err := json.Marshal(installation.Events)
	if err != nil {
		return fmt.Errorf("encode GitHub installation events: %w", err)
	}

	return r.withinTransaction(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	}, func(queries githubsql.Querier) error {
		installationID, err := queries.UpsertGitHubInstallation(ctx, githubsql.UpsertGitHubInstallationParams{
			WorkspaceID:             workspaceID,
			GithubAppID:             appID,
			GithubInstallationID:    installation.ID,
			AccountID:               installation.Account.ID,
			AccountLogin:            installation.Account.Login,
			AccountType:             installation.Account.Type,
			AccountAvatarURL:        installation.Account.AvatarURL,
			RepositorySelection:     installation.RepositorySelection,
			Permissions:             permissionsJSON,
			Events:                  eventsJSON,
			InstalledByUserID:       &installedByUserID,
			InstalledByGithubUserID: &installation.Sender.ID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInstallationOwnershipConflict
		}
		if err != nil {
			return err
		}

		repositoryIDs := make([]int64, 0, len(repositories))
		for _, repository := range repositories {
			if repository.ID <= 0 {
				return errors.New("invalid GitHub repository input")
			}
			repositoryIDs = append(repositoryIDs, repository.ID)
			if err := queries.UpsertGitHubRepository(ctx, githubsql.UpsertGitHubRepositoryParams{
				WorkspaceID:        workspaceID,
				InstallationID:     installationID,
				GithubRepositoryID: repository.ID,
				OwnerID:            repository.Owner.ID,
				OwnerLogin:         repository.Owner.Login,
				Name:               repository.Name,
				FullName:           repository.FullName,
				Description:        repository.Description,
				HtmlURL:            repository.HTMLURL,
				CloneURL:           repository.CloneURL,
				SshURL:             repository.SSHURL,
				DefaultBranch:      repository.DefaultBranch,
				IsPrivate:          repository.Private,
				IsArchived:         repository.Archived,
				IsDisabled:         repository.Disabled,
			}); err != nil {
				return err
			}
		}

		if len(repositoryIDs) == 0 {
			return queries.DeactivateAllGitHubRepositories(ctx, githubsql.DeactivateAllGitHubRepositoriesParams{
				InstallationID: installationID,
			})
		}
		return queries.DeactivateMissingGitHubRepositories(ctx, githubsql.DeactivateMissingGitHubRepositoriesParams{
			InstallationID:      installationID,
			GithubRepositoryIds: repositoryIDs,
		})
	})
}
