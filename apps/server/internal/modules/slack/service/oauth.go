package slack

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/complexus-tech/projects-api/internal/platform/oauthstate"
	"github.com/google/uuid"
)

func (s *Service) newOpaqueNonce() (string, []byte, error) {
	token, err := oauthstate.New(s.random)
	if err != nil {
		return "", nil, fmt.Errorf("generate Slack nonce: %w", err)
	}
	return token.String(), token.Digest(), nil
}

func (s *Service) consumeNonce(ctx context.Context, purpose, rawToken string, workspaceID, userID *uuid.UUID) (nonceRecord, error) {
	if s.nonces == nil {
		return nonceRecord{}, errors.New("slack nonce store is not configured")
	}
	token, err := oauthstate.Parse(rawToken)
	if err != nil {
		return nonceRecord{}, errors.Join(errors.New("invalid Slack nonce"), err)
	}
	record, err := s.nonces.ConsumeNonce(ctx, nonceConsumeInput{
		Provider:    slackProviderMessaging,
		Purpose:     purpose,
		NonceHash:   token.Digest(),
		WorkspaceID: workspaceID,
		UserID:      userID,
		Now:         s.clock.Now().UTC(),
	})
	if err != nil {
		return nonceRecord{}, err
	}
	return record, nil
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (s *Service) exchangeOAuthCode(ctx context.Context, code string) (oauthAccessResponse, error) {
	var response oauthAccessResponse
	if err := s.slackClient().oauthV2Access(ctx, s.cfg.ClientID, s.cfg.ClientSecret, s.cfg.RedirectURL, code, &response); err != nil {
		return oauthAccessResponse{}, err
	}
	if strings.TrimSpace(response.AccessToken) == "" {
		return oauthAccessResponse{}, errors.New("slack oauth returned empty access token")
	}
	if strings.TrimSpace(response.Team.ID) == "" {
		return oauthAccessResponse{}, errors.New("slack oauth returned empty team id")
	}
	if strings.TrimSpace(response.Team.Domain) == "" {
		response.Team.Domain = response.Team.ID
	}
	if strings.TrimSpace(response.Team.Name) == "" {
		response.Team.Name = response.Team.Domain
	}
	return response, nil
}

func (s *Service) canUseOAuth() bool {
	return strings.TrimSpace(s.cfg.ClientID) != "" &&
		strings.TrimSpace(s.cfg.ClientSecret) != "" &&
		strings.TrimSpace(s.cfg.RedirectURL) != "" &&
		s.credentials != nil &&
		s.nonces != nil
}

type oauthAccessResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	BotUserID    string `json:"bot_user_id"`
	AppID        string `json:"app_id"`
	Scope        string `json:"scope"`
	Team         struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Domain string `json:"domain"`
	} `json:"team"`
	Enterprise struct {
		ID string `json:"id"`
	} `json:"enterprise"`
	AuthedUser struct {
		ID string `json:"id"`
	} `json:"authed_user"`
}

type oauthInstallNoncePayload struct {
	WorkspaceSlug string `json:"workspace_slug"`
}

type oauthAccountLinkNoncePayload struct {
	WorkspaceSlug string `json:"workspace_slug"`
	ReturnURL     string `json:"return_url"`
}
