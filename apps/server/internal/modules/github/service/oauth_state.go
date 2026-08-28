package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/oauthstate"
	"github.com/google/uuid"
)

const (
	githubOAuthStateVersion     = 1
	githubOAuthStateProvider    = "github"
	githubInstallStatePurpose   = "app-install"
	githubUserLinkStatePurpose  = "user-link"
	githubInstallStateTTL       = 10 * time.Minute
	githubMaximumOAuthStateTTL  = 15 * time.Minute
	githubOAuthStateCachePrefix = "provider-oauth-state"
)

var (
	ErrGitHubOAuthStateNotConfigured = errors.New("GitHub OAuth state store is not configured")
	ErrGitHubOAuthStateInvalid       = errors.New("invalid or expired GitHub OAuth state")
	ErrGitHubOAuthStateBinding       = errors.New("GitHub OAuth state binding mismatch")
)

type githubOAuthStateRecord struct {
	Version     int             `json:"version"`
	Provider    string          `json:"provider"`
	Purpose     string          `json:"purpose"`
	WorkspaceID *uuid.UUID      `json:"workspace_id,omitempty"`
	UserID      uuid.UUID       `json:"user_id"`
	Payload     json.RawMessage `json:"payload"`
	ExpiresAt   time.Time       `json:"expires_at"`
}

type githubInstallStatePayload struct {
	WorkspaceSlug string `json:"workspace_slug"`
}

type githubUserLinkStatePayload struct {
	ReturnTo string `json:"return_to"`
}

func (s *Service) createInstallState(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	workspaceSlug string,
) (string, error) {
	workspaceSlug = strings.TrimSpace(workspaceSlug)
	if workspaceID == uuid.Nil || userID == uuid.Nil || workspaceSlug == "" {
		return "", fmt.Errorf("%w: workspace, user, and workspace slug are required", ErrGitHubOAuthStateBinding)
	}
	return s.issueOAuthState(
		ctx,
		githubInstallStatePurpose,
		&workspaceID,
		userID,
		githubInstallStatePayload{WorkspaceSlug: workspaceSlug},
		githubInstallStateTTL,
	)
}

func (s *Service) consumeInstallState(ctx context.Context, rawState string) (githubOAuthStateRecord, githubInstallStatePayload, error) {
	record, err := s.consumeOAuthState(ctx, githubInstallStatePurpose, rawState, nil, nil)
	if err != nil {
		return githubOAuthStateRecord{}, githubInstallStatePayload{}, err
	}
	if record.WorkspaceID == nil || *record.WorkspaceID == uuid.Nil || record.UserID == uuid.Nil {
		return githubOAuthStateRecord{}, githubInstallStatePayload{}, ErrGitHubOAuthStateBinding
	}
	var payload githubInstallStatePayload
	if err := json.Unmarshal(record.Payload, &payload); err != nil || strings.TrimSpace(payload.WorkspaceSlug) == "" {
		return githubOAuthStateRecord{}, githubInstallStatePayload{}, ErrGitHubOAuthStateBinding
	}
	payload.WorkspaceSlug = strings.TrimSpace(payload.WorkspaceSlug)
	return record, payload, nil
}

func (s *Service) createUserLinkState(ctx context.Context, userID uuid.UUID, returnTo string) (string, error) {
	if userID == uuid.Nil {
		return "", fmt.Errorf("%w: user is required", ErrGitHubOAuthStateBinding)
	}
	safeReturnTo, err := s.safeUserLinkReturnTo(returnTo)
	if err != nil {
		return "", err
	}
	return s.issueOAuthState(
		ctx,
		githubUserLinkStatePurpose,
		nil,
		userID,
		githubUserLinkStatePayload{ReturnTo: safeReturnTo},
		userLinkStateTTL,
	)
}

func (s *Service) consumeUserLinkState(ctx context.Context, rawState string, userID uuid.UUID) (string, error) {
	if userID == uuid.Nil {
		return "", ErrGitHubOAuthStateBinding
	}
	record, err := s.consumeOAuthState(ctx, githubUserLinkStatePurpose, rawState, nil, &userID)
	if err != nil {
		return "", err
	}
	if record.WorkspaceID != nil {
		return "", ErrGitHubOAuthStateBinding
	}
	var payload githubUserLinkStatePayload
	if err := json.Unmarshal(record.Payload, &payload); err != nil {
		return "", ErrGitHubOAuthStateBinding
	}
	return s.safeUserLinkReturnTo(payload.ReturnTo)
}

func (s *Service) issueOAuthState(
	ctx context.Context,
	purpose string,
	workspaceID *uuid.UUID,
	userID uuid.UUID,
	payload any,
	ttl time.Duration,
) (string, error) {
	if s.oauthStates == nil {
		return "", ErrGitHubOAuthStateNotConfigured
	}
	if ttl <= 0 || ttl > githubMaximumOAuthStateTTL {
		return "", fmt.Errorf("%w: invalid state lifetime", ErrGitHubOAuthStateInvalid)
	}
	token, err := oauthstate.NewRandom()
	if err != nil {
		return "", err
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode GitHub OAuth state payload: %w", err)
	}
	now := s.currentTime().UTC()
	record := githubOAuthStateRecord{
		Version:     githubOAuthStateVersion,
		Provider:    githubOAuthStateProvider,
		Purpose:     purpose,
		WorkspaceID: workspaceID,
		UserID:      userID,
		Payload:     payloadJSON,
		ExpiresAt:   now.Add(ttl),
	}
	if err := s.oauthStates.Set(ctx, githubOAuthStateCacheKey(purpose, token.Digest()), record, ttl); err != nil {
		return "", fmt.Errorf("store GitHub OAuth state: %w", err)
	}
	return token.String(), nil
}

func (s *Service) consumeOAuthState(
	ctx context.Context,
	purpose, rawState string,
	expectedWorkspaceID, expectedUserID *uuid.UUID,
) (githubOAuthStateRecord, error) {
	if s.oauthStates == nil {
		return githubOAuthStateRecord{}, ErrGitHubOAuthStateNotConfigured
	}
	token, err := oauthstate.Parse(rawState)
	if err != nil {
		return githubOAuthStateRecord{}, errors.Join(ErrGitHubOAuthStateInvalid, err)
	}
	var record githubOAuthStateRecord
	if err := s.oauthStates.Take(ctx, githubOAuthStateCacheKey(purpose, token.Digest()), &record); err != nil {
		// Cache errors are intentionally collapsed at this boundary. They may
		// describe infrastructure, but must never cause a callback to fail open
		// or echo a state-derived key to an API client.
		return githubOAuthStateRecord{}, ErrGitHubOAuthStateInvalid
	}
	if record.Version != githubOAuthStateVersion ||
		record.Provider != githubOAuthStateProvider ||
		record.Purpose != purpose ||
		record.UserID == uuid.Nil ||
		!s.currentTime().UTC().Before(record.ExpiresAt) {
		return githubOAuthStateRecord{}, ErrGitHubOAuthStateInvalid
	}
	if expectedWorkspaceID != nil && (record.WorkspaceID == nil || *record.WorkspaceID != *expectedWorkspaceID) {
		return githubOAuthStateRecord{}, ErrGitHubOAuthStateBinding
	}
	if expectedUserID != nil && record.UserID != *expectedUserID {
		return githubOAuthStateRecord{}, ErrGitHubOAuthStateBinding
	}
	return record, nil
}

func githubOAuthStateCacheKey(purpose string, digest []byte) string {
	return strings.Join([]string{
		githubOAuthStateCachePrefix,
		"v1",
		githubOAuthStateProvider,
		purpose,
		base64.RawURLEncoding.EncodeToString(digest),
	}, ":")
}

func (s *Service) currentTime() time.Time {
	if s.now == nil {
		return time.Now()
	}
	return s.now()
}
