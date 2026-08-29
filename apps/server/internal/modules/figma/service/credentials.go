package figma

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	figmaprovider "github.com/complexus-tech/projects-api/internal/modules/figma"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
)

const figmaCredentialRefreshSkew = 5 * time.Minute

var ErrCredentialMigrationRequired = errors.New("figma credential requires shared-vault migration")

func (s *Service) connectionToken(ctx context.Context, workspaceID uuid.UUID) (Connection, Token, error) {
	connection, err := s.repo.GetConnection(ctx, workspaceID)
	if err != nil {
		return Connection{}, Token{}, ErrNotConnected
	}
	token, err := s.openToken(connection)
	if err != nil {
		return Connection{}, Token{}, err
	}
	if s.now().UTC().Before(token.ExpiresAt.Add(-figmaCredentialRefreshSkew)) {
		return connection, token, nil
	}

	refreshed, err := s.client.refresh(ctx, token.RefreshToken)
	if err != nil {
		// Another instance may have rotated the refresh token first. Prefer the
		// current database value when it changed and is now usable.
		if current, currentToken, currentErr := s.currentChangedToken(ctx, connection); currentErr == nil {
			return current, currentToken, nil
		}
		return Connection{}, Token{}, err
	}
	if refreshed.RefreshToken == "" {
		refreshed.RefreshToken = token.RefreshToken
	}
	token = Token{
		AccessToken: refreshed.AccessToken, RefreshToken: refreshed.RefreshToken,
		TokenType: refreshed.TokenType,
		ExpiresAt: s.now().UTC().Add(time.Duration(refreshed.ExpiresIn) * time.Second),
	}
	payload, err := s.sealToken(
		connection.WorkspaceID,
		connection.ID,
		connection.InstallationGeneration,
		token,
	)
	if err != nil {
		return Connection{}, Token{}, err
	}
	replaced, err := s.repo.UpdateConnectionCredential(
		ctx,
		connection.ID,
		connection.InstallationGeneration,
		connection.CredentialPayload,
		payload,
		token.ExpiresAt,
	)
	if err != nil {
		return Connection{}, Token{}, err
	}
	if !replaced {
		return s.currentChangedToken(ctx, connection)
	}
	connection.CredentialPayload = payload
	connection.ExpiresAt = token.ExpiresAt
	return connection, token, nil
}

func (s *Service) currentChangedToken(ctx context.Context, previous Connection) (Connection, Token, error) {
	current, err := s.repo.GetConnection(ctx, previous.WorkspaceID)
	if err != nil {
		return Connection{}, Token{}, err
	}
	if current.ID != previous.ID ||
		current.InstallationGeneration != previous.InstallationGeneration ||
		current.CredentialPayload == previous.CredentialPayload {
		return Connection{}, Token{}, errors.New("figma credential refresh did not produce a current replacement")
	}
	token, err := s.openToken(current)
	if err != nil {
		return Connection{}, Token{}, err
	}
	return current, token, nil
}

func (s *Service) sealToken(
	workspaceID, connectionID, generation uuid.UUID,
	token Token,
) (string, error) {
	if s == nil || s.secrets == nil {
		return "", credentialvault.ErrNotConfigured
	}
	payload, err := json.Marshal(token)
	if err != nil {
		return "", fmt.Errorf("encode Figma credential: %w", err)
	}
	defer clear(payload)
	envelope, err := s.secrets.Seal(
		figmaprovider.CredentialContext(workspaceID, connectionID, generation),
		payload,
	)
	if err != nil {
		return "", fmt.Errorf("seal Figma credential: %w", err)
	}
	return envelope, nil
}

func (s *Service) openToken(connection Connection) (Token, error) {
	if s == nil || s.secrets == nil {
		return Token{}, credentialvault.ErrNotConfigured
	}
	if connection.ID == uuid.Nil || connection.WorkspaceID == uuid.Nil ||
		connection.InstallationGeneration == uuid.Nil ||
		connection.CredentialVersion != int16(credentialvault.CurrentVersion) ||
		!strings.HasPrefix(connection.CredentialPayload, credentialvault.EnvelopePrefix) {
		return Token{}, ErrCredentialMigrationRequired
	}
	opened, err := s.secrets.Open(
		figmaprovider.CredentialContext(
			connection.WorkspaceID,
			connection.ID,
			connection.InstallationGeneration,
		),
		connection.CredentialPayload,
	)
	if err != nil {
		return Token{}, fmt.Errorf("open Figma credential: %w", err)
	}
	defer opened.Destroy()
	payload := opened.Reveal()
	defer clear(payload)
	var token Token
	if err := json.Unmarshal(payload, &token); err != nil {
		return Token{}, errors.New("decode Figma credential: invalid payload")
	}
	if strings.TrimSpace(token.AccessToken) == "" || token.ExpiresAt.IsZero() {
		return Token{}, errors.New("decode Figma credential: required fields are empty")
	}
	return token, nil
}
