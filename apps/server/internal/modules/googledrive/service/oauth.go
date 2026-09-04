package googledrive

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	"github.com/complexus-tech/projects-api/internal/platform/workspaceurl"
	"github.com/google/uuid"
)

func (service *Service) createOAuthState(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	workspaceSlug string,
	returnURL *string,
) (string, string, error) {
	state, err := randomURLToken(32)
	if err != nil {
		return "", "", fmt.Errorf("generate Google Drive OAuth state: %w", err)
	}
	verifier, err := randomURLToken(64)
	if err != nil {
		return "", "", fmt.Errorf("generate Google Drive PKCE verifier: %w", err)
	}
	digest := sha256.Sum256([]byte(state))
	err = service.repo.SaveOAuthState(ctx, domain.OAuthState{
		StateHash:   base64.RawURLEncoding.EncodeToString(digest[:]),
		WorkspaceID: workspaceID, UserID: userID, WorkspaceSlug: workspaceSlug,
		ReturnURL: returnURL, CodeVerifier: verifier,
		ExpiresAt: service.now().UTC().Add(oauthStateTTL),
	})
	if err != nil {
		return "", "", err
	}
	return state, verifier, nil
}

func (service *Service) consumeOAuthState(ctx context.Context, state string) (domain.OAuthState, error) {
	state = strings.TrimSpace(state)
	if state == "" {
		return domain.OAuthState{}, errors.New("Google Drive OAuth state is missing")
	}
	digest := sha256.Sum256([]byte(state))
	return service.repo.ConsumeOAuthState(
		ctx,
		base64.RawURLEncoding.EncodeToString(digest[:]),
		service.now().UTC(),
	)
}

func (service *Service) oauthReturnURL(state domain.OAuthState) string {
	if state.ReturnURL != nil {
		if validated := service.validatedReturnURL(state.WorkspaceSlug, *state.ReturnURL); validated != nil {
			return *validated
		}
	}
	return workspaceurl.Build(
		service.config.WebsiteURL,
		state.WorkspaceSlug,
		"settings",
		"account",
		"google-drive",
	)
}

func (service *Service) validatedReturnURL(workspaceSlug, raw string) *string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	workspaceBase := workspaceurl.Build(service.config.WebsiteURL, workspaceSlug)
	expected, err := url.Parse(workspaceBase)
	if err != nil || expected.Host == "" {
		return nil
	}
	requested, err := url.Parse(raw)
	if err != nil || requested.User != nil || requested.Fragment != "" {
		return nil
	}
	if !requested.IsAbs() {
		if !strings.HasPrefix(requested.Path, "/") {
			return nil
		}
		requested.Scheme = expected.Scheme
		requested.Host = expected.Host
	}
	if requested.Scheme != expected.Scheme || !strings.EqualFold(requested.Host, expected.Host) {
		return nil
	}
	requested.Fragment = ""
	validated := requested.String()
	return &validated
}
