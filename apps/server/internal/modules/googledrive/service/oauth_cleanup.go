package googledrive

import (
	"context"
	"errors"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
)

// stageUnpersistedGrantRevocation durably retains only a sealed token after an
// OAuth exchange that cannot become a connection. The caller holds both the
// user and Google-subject provider gates. ownershipAbsenceProven controls only
// the last-resort inline revoke and must be false after an outcome-ambiguous
// connection write.
func (service *Service) stageUnpersistedGrantRevocation(
	ctx context.Context,
	state domain.OAuthState,
	providerUser ProviderUser,
	token domain.OAuthToken,
	ownershipAbsenceProven bool,
) error {
	generation := uuid.New()
	account := domain.Account{
		UserID: state.UserID, GoogleSubject: providerUser.Subject,
		CredentialVersion:      int16(credentialvault.CurrentVersion),
		InstallationGeneration: generation,
	}
	payload, err := service.sealToken(account, token)
	if err != nil {
		stageErr := fmt.Errorf("seal failed Google Drive OAuth grant for revocation: %w", err)
		if !ownershipAbsenceProven {
			return stageErr
		}
		return errors.Join(stageErr, service.revokeUnstagedGrant(ctx, state, token))
	}
	_, err = service.repo.EnqueueRevocation(ctx, domain.Revocation{
		UserID: state.UserID, GoogleSubject: providerUser.Subject,
		InstallationGeneration: generation,
		CredentialPayload:      payload,
		CredentialVersion:      int16(credentialvault.CurrentVersion),
	})
	if err == nil {
		return nil
	}
	if errors.Is(err, domain.ErrConflict) || errors.Is(err, domain.ErrAccountOwned) {
		// Ownership changed while the transaction revalidated the subject. A
		// cleanup revoke would invalidate that active owner's Google grant.
		return fmt.Errorf("stage failed Google Drive OAuth revocation: %w", err)
	}
	if !ownershipAbsenceProven {
		return fmt.Errorf("stage failed Google Drive OAuth revocation: %w", err)
	}
	return errors.Join(
		fmt.Errorf("stage failed Google Drive OAuth revocation: %w", err),
		service.revokeUnstagedGrant(ctx, state, token),
	)
}

func (service *Service) revokeUnstagedGrant(
	ctx context.Context,
	state domain.OAuthState,
	token domain.OAuthToken,
) error {
	if err := service.client.Revoke(ctx, revocationToken(token)); err != nil {
		if service.log != nil {
			service.log.Warn(
				ctx,
				"failed revoking Google Drive OAuth grant that could not be staged",
				"error", err,
				"workspace_id", state.WorkspaceID,
				"user_id", state.UserID,
			)
		}
		return fmt.Errorf("Google Drive OAuth grant revocation failed before staging: %w", err)
	}
	return nil
}
