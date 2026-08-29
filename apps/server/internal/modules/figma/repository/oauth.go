package figmarepository

import (
	"context"
	"time"

	figmadomain "github.com/complexus-tech/projects-api/internal/modules/figma/domain"
	figmasql "github.com/complexus-tech/projects-api/internal/modules/figma/repository/sqlc"
)

func (repository *Repository) SaveOAuthState(
	ctx context.Context,
	state figmadomain.OAuthState,
) error {
	rows, err := repository.queries.SaveOAuthState(ctx, figmasql.SaveOAuthStateParams{
		StateHash: state.StateHash, WorkspaceID: state.WorkspaceID,
		UserID: state.UserID, WorkspaceSlug: state.WorkspaceSlug,
		CodeVerifier: state.CodeVerifier, ExpiresAt: state.ExpiresAt.UTC(),
	})
	return requireAffected(rows, err, figmadomain.ErrForbidden)
}

func (repository *Repository) ConsumeOAuthState(
	ctx context.Context,
	stateHash string,
	now time.Time,
) (figmadomain.OAuthState, error) {
	consumedAt := now.UTC()
	row, err := repository.queries.ConsumeOAuthState(ctx, figmasql.ConsumeOAuthStateParams{
		StateHash: stateHash, ConsumedAt: &consumedAt,
	})
	if err != nil {
		return figmadomain.OAuthState{}, mapDatabaseError(err)
	}
	return figmadomain.OAuthState{
		StateHash: row.StateHash, WorkspaceID: row.WorkspaceID,
		UserID: row.UserID, WorkspaceSlug: row.WorkspaceSlug,
		CodeVerifier: row.CodeVerifier, ExpiresAt: row.ExpiresAt,
	}, nil
}
