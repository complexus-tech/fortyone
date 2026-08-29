package figma

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/oauthstate"
	"github.com/google/uuid"
)

const figmaOAuthStateTTL = 10 * time.Minute

var (
	ErrFigmaOAuthStateInvalid = errors.New("invalid or expired Figma OAuth state")
	ErrFigmaOAuthStateBinding = errors.New("figma OAuth state binding mismatch")
)

func (s *Service) createOAuthState(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	workspaceSlug string,
) (state string, verifier string, err error) {
	workspaceSlug = strings.TrimSpace(workspaceSlug)
	if workspaceID == uuid.Nil || userID == uuid.Nil || workspaceSlug == "" {
		return "", "", fmt.Errorf("%w: workspace, user, and workspace slug are required", ErrFigmaOAuthStateBinding)
	}
	token, err := oauthstate.NewRandom()
	if err != nil {
		return "", "", err
	}
	verifier, err = randomValue(48)
	if err != nil {
		return "", "", err
	}
	state = token.String()
	if err := s.repo.SaveOAuthState(ctx, OAuthState{
		StateHash:     digest(state),
		WorkspaceID:   workspaceID,
		UserID:        userID,
		WorkspaceSlug: workspaceSlug,
		CodeVerifier:  verifier,
		ExpiresAt:     s.now().UTC().Add(figmaOAuthStateTTL),
	}); err != nil {
		return "", "", fmt.Errorf("store Figma OAuth state: %w", err)
	}
	return state, verifier, nil
}

func (s *Service) consumeOAuthState(ctx context.Context, rawState string) (OAuthState, error) {
	token, err := oauthstate.Parse(rawState)
	if err != nil {
		return OAuthState{}, errors.Join(ErrFigmaOAuthStateInvalid, err)
	}
	now := s.now().UTC()
	stored, err := s.repo.ConsumeOAuthState(ctx, digest(token.String()), now)
	if err != nil {
		return OAuthState{}, ErrFigmaOAuthStateInvalid
	}
	if stored.WorkspaceID == uuid.Nil ||
		stored.UserID == uuid.Nil ||
		strings.TrimSpace(stored.WorkspaceSlug) == "" ||
		strings.TrimSpace(stored.CodeVerifier) == "" ||
		!now.Before(stored.ExpiresAt) {
		return OAuthState{}, ErrFigmaOAuthStateBinding
	}
	stored.WorkspaceSlug = strings.TrimSpace(stored.WorkspaceSlug)
	return stored, nil
}
