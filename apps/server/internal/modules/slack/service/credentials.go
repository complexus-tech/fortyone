package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
)

type slackCredential struct {
	AccessToken  string     `json:"accessToken"`
	RefreshToken string     `json:"refreshToken,omitempty"`
	ExpiresAt    *time.Time `json:"expiresAt,omitempty"`
}

type slackCredentialBinding struct {
	WorkspaceID       uuid.UUID
	SlackTeamID       string
	InstallGeneration uuid.UUID
}

func (b slackCredentialBinding) vaultContext() credentialvault.Context {
	// #nosec G101 -- these literals are public AAD domain identifiers, not credentials.
	return credentialvault.Context{
		Provider:       "slack",
		TenantID:       b.WorkspaceID.String(),
		SubjectID:      strings.TrimSpace(b.SlackTeamID),
		CredentialType: "bot-oauth",
		Generation:     b.InstallGeneration.String(),
	}
}

type credentialCodec struct {
	vault CredentialVault
}

func (s *Service) botToken(_ context.Context, installation slackdomain.Installation) (string, error) {
	payload := strings.TrimSpace(installation.BotAccessToken)
	if payload == "" {
		return "", errors.New("slack installation is missing bot token")
	}
	if s.credentials == nil {
		return "", errors.New("slack credential encryption is not configured")
	}
	if installation.CredentialVersion != credentialvault.CurrentVersion {
		return "", errors.New("slack credential requires vault migration")
	}
	credential, openedVersion, err := s.credentials.open(slackCredentialBinding{
		WorkspaceID:       installation.WorkspaceID,
		SlackTeamID:       installation.SlackTeamID,
		InstallGeneration: installation.InstallGeneration,
	}, payload)
	if err != nil {
		return "", err
	}
	if openedVersion != installation.CredentialVersion {
		return "", errors.New("slack credential envelope version mismatch")
	}
	return credential.AccessToken, nil
}

func newCredentialCodec(vault CredentialVault) (*credentialCodec, error) {
	if vault == nil {
		return nil, credentialvault.ErrNotConfigured
	}
	return &credentialCodec{vault: vault}, nil
}

func (c *credentialCodec) seal(binding slackCredentialBinding, credential slackCredential) (string, int, error) {
	if c == nil || c.vault == nil {
		return "", 0, errors.New("slack credential encryption is not configured")
	}
	credential.AccessToken = strings.TrimSpace(credential.AccessToken)
	credential.RefreshToken = strings.TrimSpace(credential.RefreshToken)
	if credential.AccessToken == "" {
		return "", 0, errors.New("slack access token is required")
	}
	// #nosec G117 -- this buffer is cleared and passed directly to envelope encryption below.
	payload, err := json.Marshal(credential)
	if err != nil {
		return "", 0, fmt.Errorf("encode slack credential: %w", err)
	}
	defer clear(payload)
	value, err := c.vault.Seal(binding.vaultContext(), payload)
	if err != nil {
		return "", 0, fmt.Errorf("encrypt slack credential: %w", err)
	}
	return value, credentialvault.CurrentVersion, nil
}

func (c *credentialCodec) open(binding slackCredentialBinding, value string) (slackCredential, int, error) {
	if c == nil || c.vault == nil {
		return slackCredential{}, 0, errors.New("slack credential encryption is not configured")
	}
	if !credentialvault.IsEnvelope(value) {
		return slackCredential{}, 0, errors.New("slack credential requires vault migration")
	}
	opened, err := c.vault.Open(binding.vaultContext(), value)
	if err != nil {
		return slackCredential{}, 0, fmt.Errorf("decrypt slack credential: %w", err)
	}
	defer opened.Destroy()
	plaintext := opened.Reveal()
	defer clear(plaintext)
	credential, err := decodeSlackCredential(plaintext)
	if err != nil {
		return slackCredential{}, 0, err
	}
	return credential, credentialvault.CurrentVersion, nil
}
func decodeSlackCredential(payload []byte) (slackCredential, error) {
	var credential slackCredential
	if err := json.Unmarshal(payload, &credential); err != nil {
		// JSON/time decoding errors can echo input values. Keep provider
		// credential plaintext out of error and logging boundaries.
		return slackCredential{}, errors.New("decode slack credential: invalid payload")
	}
	credential.AccessToken = strings.TrimSpace(credential.AccessToken)
	credential.RefreshToken = strings.TrimSpace(credential.RefreshToken)
	if credential.AccessToken == "" {
		return slackCredential{}, errors.New("slack access token is empty")
	}
	return credential, nil
}
