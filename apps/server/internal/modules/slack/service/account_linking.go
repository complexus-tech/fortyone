package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) CreateAccountLinkSession(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	workspaceSlug, returnURL string,
) (CoreCreateAccountLinkSession, error) {
	if workspaceID == uuid.Nil || userID == uuid.Nil || strings.TrimSpace(workspaceSlug) == "" {
		return CoreCreateAccountLinkSession{}, errors.New("workspace, user, and workspace slug are required")
	}
	if err := s.requireWorkspaceMember(ctx, workspaceID, userID); err != nil {
		return CoreCreateAccountLinkSession{}, err
	}
	installation, err := s.repo.GetSlackWorkspace(ctx, workspaceID)
	if err != nil {
		return CoreCreateAccountLinkSession{}, err
	}
	link, err := s.repo.FindSlackUserLinkByUser(ctx, workspaceID, installation.SlackTeamID, userID)
	if err != nil {
		return CoreCreateAccountLinkSession{}, err
	}
	if link != nil {
		return CoreCreateAccountLinkSession{Linked: true}, nil
	}
	if !s.canUseOAuth() {
		return CoreCreateAccountLinkSession{}, nil
	}

	returnURL = s.safeSlackAccountLinkReturnURL(workspaceSlug, returnURL)
	payload, err := json.Marshal(oauthAccountLinkNoncePayload{
		WorkspaceSlug: strings.TrimSpace(workspaceSlug),
		ReturnURL:     returnURL,
	})
	if err != nil {
		return CoreCreateAccountLinkSession{}, err
	}
	state, digest, err := s.newOpaqueNonce()
	if err != nil {
		return CoreCreateAccountLinkSession{}, err
	}
	boundUserID := userID
	boundTeamID := strings.TrimSpace(installation.SlackTeamID)
	if boundTeamID == "" {
		return CoreCreateAccountLinkSession{}, errors.New("slack installation is missing its team ID")
	}
	if err := s.nonces.CreateNonce(ctx, nonceInput{
		Provider:            slackProviderMessaging,
		Purpose:             slackNoncePurposeAccount,
		NonceHash:           digest,
		WorkspaceID:         workspaceID,
		UserID:              &boundUserID,
		ExternalWorkspaceID: boundTeamID,
		Payload:             payload,
		ExpiresAt:           s.clock.Now().UTC().Add(slackAccountLinkNonceTTL),
	}); err != nil {
		return CoreCreateAccountLinkSession{}, fmt.Errorf("create Slack account-link state: %w", err)
	}

	authURL := fmt.Sprintf(
		"https://slack.com/oauth/v2/authorize?client_id=%s&scope=%s&team=%s&state=%s&redirect_uri=%s",
		url.QueryEscape(s.cfg.ClientID),
		url.QueryEscape(slackBotOAuthScopeValue()),
		url.QueryEscape(boundTeamID),
		url.QueryEscape(state),
		url.QueryEscape(s.cfg.RedirectURL),
	)
	return CoreCreateAccountLinkSession{CanLink: true, InstallURL: authURL}, nil
}

func (s *Service) handleAccountLinkSetup(
	ctx context.Context,
	nonce nonceRecord,
	code, slackError string,
) (string, error) {
	var noncePayload oauthAccountLinkNoncePayload
	if err := json.Unmarshal(nonce.Payload, &noncePayload); err != nil || strings.TrimSpace(noncePayload.WorkspaceSlug) == "" {
		return "", errors.New("invalid Slack account-link state payload")
	}
	if strings.TrimSpace(slackError) != "" {
		return s.slackAccountLinkRedirect(noncePayload.ReturnURL, "error"), nil
	}
	if strings.TrimSpace(code) == "" {
		return "", errors.New("missing Slack account-link code")
	}
	if nonce.WorkspaceID == uuid.Nil || nonce.UserID == nil || *nonce.UserID == uuid.Nil {
		return "", errors.New("invalid Slack account-link state binding")
	}
	if err := s.requireWorkspaceMember(ctx, nonce.WorkspaceID, *nonce.UserID); err != nil {
		return "", err
	}

	oauthResp, err := s.exchangeOAuthCode(ctx, code)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(valueOrEmpty(nonce.ExternalWorkspaceID)) == "" || strings.TrimSpace(oauthResp.Team.ID) != strings.TrimSpace(valueOrEmpty(nonce.ExternalWorkspaceID)) {
		return "", errors.New("slack account-link team does not match the connected workspace")
	}
	slackWorkspace, err := s.repo.GetSlackWorkspaceByTeamID(ctx, oauthResp.Team.ID)
	if err != nil {
		return "", err
	}
	if slackWorkspace.WorkspaceID != nonce.WorkspaceID {
		return "", errors.New("slack account-link workspace does not match the connected workspace")
	}
	slackUserID := strings.TrimSpace(oauthResp.AuthedUser.ID)
	if !validSlackUserID(slackUserID) {
		return "", errors.New("slack account-link returned an invalid user ID")
	}
	if err := s.repo.UpsertSlackUserLinks(ctx, nonce.WorkspaceID, slackWorkspace.ID, oauthResp.Team.ID, []slackUserLinkUpsert{
		{
			SlackUserID: slackUserID,
			UserID:      *nonce.UserID,
			LinkedVia:   "dashboard_oauth",
		},
	}); err != nil {
		return "", err
	}
	return s.slackAccountLinkRedirect(noncePayload.ReturnURL, "success"), nil
}

func (s *Service) LinkSlackAccount(ctx context.Context, workspaceID, userID uuid.UUID, token string) (CoreLinkSlackAccountResult, error) {
	if workspaceID == uuid.Nil {
		return CoreLinkSlackAccountResult{}, errors.New("workspace id is required")
	}
	if userID == uuid.Nil {
		return CoreLinkSlackAccountResult{}, errors.New("user id is required")
	}
	if err := s.requireWorkspaceMember(ctx, workspaceID, userID); err != nil {
		return CoreLinkSlackAccountResult{}, err
	}
	installation, err := s.repo.GetSlackWorkspace(ctx, workspaceID)
	if err != nil {
		return CoreLinkSlackAccountResult{}, err
	}
	existing, err := s.repo.FindSlackUserLinkByUser(ctx, workspaceID, installation.SlackTeamID, userID)
	if err != nil {
		return CoreLinkSlackAccountResult{}, err
	}
	if existing != nil {
		return CoreLinkSlackAccountResult{
			AlreadyLinked: true,
			SlackUserID:   existing.SlackUserID,
		}, nil
	}
	nonce, err := s.consumeNonce(ctx, slackNoncePurposeAccount, token, &workspaceID, &userID)
	if err != nil {
		return CoreLinkSlackAccountResult{}, fmt.Errorf("invalid or expired Slack link token: %w", err)
	}
	if nonce.WorkspaceID != workspaceID {
		return CoreLinkSlackAccountResult{}, errors.New("slack link token workspace mismatch")
	}
	if nonce.UserID != nil && *nonce.UserID != userID {
		return CoreLinkSlackAccountResult{}, errors.New("slack link token user mismatch")
	}

	slackTeamID := strings.TrimSpace(valueOrEmpty(nonce.ExternalWorkspaceID))
	slackUserID := strings.TrimSpace(valueOrEmpty(nonce.ExternalUserID))
	if slackTeamID == "" || slackUserID == "" {
		return CoreLinkSlackAccountResult{}, errors.New("invalid slack link token")
	}

	slackWorkspace, err := s.repo.GetSlackWorkspaceByTeamID(ctx, slackTeamID)
	if err != nil {
		return CoreLinkSlackAccountResult{}, err
	}
	if slackWorkspace.WorkspaceID != workspaceID {
		return CoreLinkSlackAccountResult{}, errors.New("slack workspace does not belong to this workspace")
	}

	if err := s.repo.UpsertSlackUserLinks(ctx, workspaceID, slackWorkspace.ID, slackTeamID, []slackUserLinkUpsert{
		{
			SlackUserID: slackUserID,
			UserID:      userID,
			LinkedVia:   "manual_link",
		},
	}); err != nil {
		return CoreLinkSlackAccountResult{}, err
	}
	s.dispatchFirstInteractionGuide(ctx, slackWorkspace, slackUserID)
	return CoreLinkSlackAccountResult{SlackUserID: slackUserID}, nil
}

func (s *Service) DisconnectSlackAccount(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error) {
	if workspaceID == uuid.Nil || userID == uuid.Nil {
		return false, errors.New("workspace and user are required")
	}
	if err := s.requireWorkspaceMember(ctx, workspaceID, userID); err != nil {
		return false, err
	}
	installation, err := s.repo.GetSlackWorkspace(ctx, workspaceID)
	if err != nil {
		return false, err
	}
	link, err := s.repo.FindSlackUserLinkByUser(ctx, workspaceID, installation.SlackTeamID, userID)
	if err != nil || link == nil {
		return false, err
	}
	return s.repo.DeleteSlackUserLink(ctx, workspaceID, installation.SlackTeamID, link.SlackUserID, userID)
}

func (s *Service) disconnectSlackAccountBySource(ctx context.Context, source requestSourceContext) (bool, error) {
	installation, err := s.repo.GetSlackWorkspaceByTeamID(ctx, strings.TrimSpace(source.SlackTeamID))
	if err != nil {
		if isSlackRepositoryNotFound(err) {
			return false, nil
		}
		return false, err
	}
	userID, err := s.repo.FindLinkedUserIDBySlackUser(
		ctx,
		installation.WorkspaceID,
		installation.SlackTeamID,
		strings.TrimSpace(source.SlackUserID),
	)
	if err != nil || userID == nil {
		return false, err
	}
	return s.repo.DeleteSlackUserLink(
		ctx,
		installation.WorkspaceID,
		installation.SlackTeamID,
		strings.TrimSpace(source.SlackUserID),
		*userID,
	)
}

func (s *Service) buildSlackUserLinkURL(ctx context.Context, workspaceID uuid.UUID, slackTeamID, slackUserID string) (string, error) {
	if workspaceID == uuid.Nil {
		return "", errors.New("workspace id is required")
	}
	if s.nonces == nil {
		return "", errors.New("slack account-link nonce store is not configured")
	}
	slackTeamID = strings.TrimSpace(slackTeamID)
	slackUserID = strings.TrimSpace(slackUserID)
	if slackTeamID == "" || slackUserID == "" {
		return "", nil
	}

	workspace, err := s.repo.FindWorkspaceByID(ctx, workspaceID)
	if err != nil {
		return "", err
	}
	token, digest, err := s.newOpaqueNonce()
	if err != nil {
		return "", err
	}
	if err := s.nonces.CreateNonce(ctx, nonceInput{
		Provider:            slackProviderMessaging,
		Purpose:             slackNoncePurposeAccount,
		NonceHash:           digest,
		WorkspaceID:         workspaceID,
		ExternalWorkspaceID: slackTeamID,
		ExternalUserID:      slackUserID,
		ExpiresAt:           s.clock.Now().UTC().Add(slackAccountLinkNonceTTL),
	}); err != nil {
		return "", fmt.Errorf("create Slack account-link token: %w", err)
	}

	baseLink := s.buildAccountIntegrationURL(workspace.Slug)
	linkURL, err := url.Parse(baseLink)
	if err != nil {
		return "", nil
	}
	query := linkURL.Query()
	query.Set("slack_link_token", token)
	linkURL.RawQuery = query.Encode()
	return linkURL.String(), nil
}
