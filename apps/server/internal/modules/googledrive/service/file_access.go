package googledrive

import (
	"context"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	"github.com/google/uuid"
)

// handleReferenceProviderError invalidates only the requesting actor's grant
// when Google no longer lets that account see the file. Google can return 404
// for both deletion and access loss, so an API error alone must not mark shared
// file metadata globally deleted.
func (service *Service) handleReferenceProviderError(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	reference domain.FileReference,
	account domain.Account,
	err error,
) error {
	var apiError *APIError
	if !errors.As(err, &apiError) {
		return err
	}
	if apiError.StatusCode == http.StatusUnauthorized {
		return service.handleProviderAuthorizationError(ctx, account, err)
	}
	if apiError.StatusCode != http.StatusForbidden && apiError.StatusCode != http.StatusNotFound {
		return err
	}
	if apiError.StatusCode == http.StatusForbidden && !apiError.isFilePermissionLoss() {
		return err
	}
	if reference.GrantGeneration == nil {
		return err
	}
	if deleteErr := service.repo.DeleteReferenceGrant(
		ctx,
		workspaceID,
		userID,
		account.ID,
		reference.ID,
		*reference.GrantGeneration,
	); deleteErr != nil {
		service.logProviderError(ctx, "invalidate Google Drive file grant", deleteErr, workspaceID, userID)
		return deleteErr
	}
	return err
}

func (service *Service) markReferenceUnavailable(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	referenceID uuid.UUID,
) error {
	if err := service.repo.MarkReferenceUnavailable(ctx, workspaceID, referenceID); err != nil {
		service.logProviderError(ctx, "mark Google Drive file unavailable", err, workspaceID, userID)
		return err
	}
	return domain.ErrNotFound
}
