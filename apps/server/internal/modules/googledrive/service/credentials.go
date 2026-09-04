package googledrive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	googledriveprovider "github.com/complexus-tech/projects-api/internal/modules/googledrive"
	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
)

func (service *Service) connectionToken(ctx context.Context, workspaceID, userID uuid.UUID) (domain.Account, domain.OAuthToken, error) {
	connection, err := service.repo.GetConnection(ctx, workspaceID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.Account{}, domain.OAuthToken{}, domain.ErrNotConnected
		}
		return domain.Account{}, domain.OAuthToken{}, err
	}
	return service.accountToken(ctx, workspaceID, userID, connection.Account)
}

func (service *Service) Disconnect(ctx context.Context, workspaceID, userID uuid.UUID) error {
	return service.repo.WithinProviderUserLifecycle(ctx, userID, func(lifecycleCtx context.Context) error {
		_, err := service.repo.Disconnect(lifecycleCtx, workspaceID, userID)
		return err
	})
}

func (service *Service) accountToken(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	account domain.Account,
) (domain.Account, domain.OAuthToken, error) {
	if account.UserID != userID {
		return domain.Account{}, domain.OAuthToken{}, domain.ErrForbidden
	}
	if account.RequiresReauthorization {
		return domain.Account{}, domain.OAuthToken{}, domain.ErrReauthorizationRequired
	}
	token, err := service.openToken(account)
	if err != nil {
		return domain.Account{}, domain.OAuthToken{}, err
	}
	if service.now().UTC().Before(token.Expiry.Add(-tokenRefreshSkew)) {
		return account, token, nil
	}
	refreshed, err := service.client.Refresh(ctx, token.RefreshToken)
	if err != nil {
		if isReauthorizationError(err) {
			_ = service.repo.MarkReauthorizationRequired(ctx, account, "invalid_grant")
			return domain.Account{}, domain.OAuthToken{}, domain.ErrReauthorizationRequired
		}
		if current, currentToken, currentErr := service.currentChangedToken(ctx, workspaceID, userID, account); currentErr == nil {
			return current, currentToken, nil
		}
		return domain.Account{}, domain.OAuthToken{}, err
	}
	if refreshed.RefreshToken == "" {
		refreshed.RefreshToken = token.RefreshToken
	}
	payload, err := service.sealToken(account, refreshed)
	if err != nil {
		return domain.Account{}, domain.OAuthToken{}, err
	}
	replaced, err := service.repo.CompareAndSwapCredential(ctx, account, payload, refreshed.Expiry)
	if err != nil {
		return domain.Account{}, domain.OAuthToken{}, err
	}
	if !replaced {
		return service.currentChangedToken(ctx, workspaceID, userID, account)
	}
	account.CredentialPayload = payload
	account.ExpiresAt = refreshed.Expiry
	return account, refreshed, nil
}

func (service *Service) currentChangedToken(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	previous domain.Account,
) (domain.Account, domain.OAuthToken, error) {
	connection, err := service.repo.GetConnection(ctx, workspaceID, userID)
	if err != nil {
		return domain.Account{}, domain.OAuthToken{}, err
	}
	current := connection.Account
	if current.ID != previous.ID || current.InstallationGeneration != previous.InstallationGeneration ||
		current.CredentialPayload == previous.CredentialPayload {
		return domain.Account{}, domain.OAuthToken{}, domain.ErrConflict
	}
	token, err := service.openToken(current)
	if err != nil {
		return domain.Account{}, domain.OAuthToken{}, err
	}
	return current, token, nil
}

func (service *Service) sealToken(account domain.Account, token domain.OAuthToken) (string, error) {
	if service == nil || service.config.Credentials == nil {
		return "", credentialvault.ErrNotConfigured
	}
	payload, err := marshalToken(token)
	if err != nil {
		return "", fmt.Errorf("encode Google Drive credential: %w", err)
	}
	defer clearBytes(payload)
	envelope, err := service.config.Credentials.Seal(
		googledriveprovider.CredentialContext(account.UserID, account.GoogleSubject, account.InstallationGeneration),
		payload,
	)
	if err != nil {
		return "", fmt.Errorf("seal Google Drive credential: %w", err)
	}
	return envelope, nil
}

func (service *Service) openToken(account domain.Account) (domain.OAuthToken, error) {
	if service == nil {
		return domain.OAuthToken{}, credentialvault.ErrNotConfigured
	}
	return openToken(service.config.Credentials, account)
}

func openToken(credentials CredentialVault, account domain.Account) (domain.OAuthToken, error) {
	token, err := decodeTokenEnvelope(credentials, account)
	if err != nil {
		return domain.OAuthToken{}, err
	}
	if strings.TrimSpace(token.AccessToken) == "" || strings.TrimSpace(token.RefreshToken) == "" || token.Expiry.IsZero() {
		return domain.OAuthToken{}, errors.New("decode Google Drive credential: required fields are empty")
	}
	return token, nil
}

func openRevocationToken(credentials CredentialVault, account domain.Account) (domain.OAuthToken, error) {
	token, err := decodeTokenEnvelope(credentials, account)
	if err != nil {
		return domain.OAuthToken{}, err
	}
	if strings.TrimSpace(token.AccessToken) == "" && strings.TrimSpace(token.RefreshToken) == "" {
		return domain.OAuthToken{}, errors.New("decode Google Drive revocation credential: no revocable token")
	}
	return token, nil
}

func decodeTokenEnvelope(credentials CredentialVault, account domain.Account) (domain.OAuthToken, error) {
	if credentials == nil {
		return domain.OAuthToken{}, credentialvault.ErrNotConfigured
	}
	if account.UserID == uuid.Nil || strings.TrimSpace(account.GoogleSubject) == "" ||
		account.InstallationGeneration == uuid.Nil || account.CredentialVersion != int16(credentialvault.CurrentVersion) ||
		!strings.HasPrefix(account.CredentialPayload, credentialvault.EnvelopePrefix) {
		return domain.OAuthToken{}, errors.New("Google Drive credential requires vault migration")
	}
	opened, err := credentials.Open(
		googledriveprovider.CredentialContext(account.UserID, account.GoogleSubject, account.InstallationGeneration),
		account.CredentialPayload,
	)
	if err != nil {
		return domain.OAuthToken{}, fmt.Errorf("open Google Drive credential: %w", err)
	}
	defer opened.Destroy()
	payload := opened.Reveal()
	defer clearBytes(payload)
	var token domain.OAuthToken
	if err := json.Unmarshal(payload, &token); err != nil {
		return domain.OAuthToken{}, errors.New("decode Google Drive credential: invalid payload")
	}
	return token, nil
}
