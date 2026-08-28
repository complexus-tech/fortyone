package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	"github.com/google/uuid"
)

func (s *Service) CreateInstallSession(ctx context.Context, workspaceID, userID uuid.UUID, workspaceSlug string) (CoreCreateInstallSession, error) {
	if !s.canUseOAuth() {
		return CoreCreateInstallSession{}, ErrSlackNotConfigured
	}
	if workspaceID == uuid.Nil || userID == uuid.Nil || strings.TrimSpace(workspaceSlug) == "" {
		return CoreCreateInstallSession{}, errors.New("workspace, user, and workspace slug are required")
	}
	if err := s.requireWorkspaceAdmin(ctx, workspaceID, userID); err != nil {
		return CoreCreateInstallSession{}, err
	}
	payload, err := json.Marshal(oauthInstallNoncePayload{WorkspaceSlug: strings.TrimSpace(workspaceSlug)})
	if err != nil {
		return CoreCreateInstallSession{}, err
	}
	state, digest, err := s.newOpaqueNonce()
	if err != nil {
		return CoreCreateInstallSession{}, err
	}
	boundUserID := userID
	if err := s.nonces.CreateNonce(ctx, nonceInput{
		Provider:    slackProviderMessaging,
		Purpose:     slackNoncePurposeOAuth,
		NonceHash:   digest,
		WorkspaceID: workspaceID,
		UserID:      &boundUserID,
		Payload:     payload,
		ExpiresAt:   s.clock.Now().UTC().Add(slackOAuthNonceTTL),
	}); err != nil {
		return CoreCreateInstallSession{}, fmt.Errorf("create Slack install state: %w", err)
	}

	authURL := fmt.Sprintf(
		"https://slack.com/oauth/v2/authorize?client_id=%s&scope=%s&state=%s&redirect_uri=%s",
		url.QueryEscape(s.cfg.ClientID),
		url.QueryEscape(slackBotOAuthScopeValue()),
		url.QueryEscape(state),
		url.QueryEscape(s.cfg.RedirectURL),
	)

	return CoreCreateInstallSession{InstallURL: authURL}, nil
}

func (s *Service) HandleSetup(ctx context.Context, code, state, slackError string) (string, error) {
	if !s.canUseOAuth() {
		return "", ErrSlackNotConfigured
	}
	if nonce, err := s.consumeNonce(ctx, slackNoncePurposeAccount, state, nil, nil); err == nil {
		return s.handleAccountLinkSetup(ctx, nonce, code, slackError)
	}
	return s.handleInstallSetup(ctx, code, state, slackError)
}

func (s *Service) handleInstallSetup(ctx context.Context, code, state, slackError string) (string, error) {
	nonce, err := s.consumeNonce(ctx, slackNoncePurposeOAuth, state, nil, nil)
	if err != nil {
		return "", fmt.Errorf("invalid or expired Slack install state: %w", err)
	}
	if strings.TrimSpace(slackError) != "" {
		return "", fmt.Errorf("slack oauth failed: %s", slackError)
	}
	if strings.TrimSpace(code) == "" {
		return "", errors.New("missing slack oauth code")
	}
	if nonce.WorkspaceID == uuid.Nil || nonce.UserID == nil || *nonce.UserID == uuid.Nil {
		return "", errors.New("invalid Slack install state binding")
	}
	if err := s.requireWorkspaceAdmin(ctx, nonce.WorkspaceID, *nonce.UserID); err != nil {
		return "", err
	}
	var noncePayload oauthInstallNoncePayload
	if err := json.Unmarshal(nonce.Payload, &noncePayload); err != nil || strings.TrimSpace(noncePayload.WorkspaceSlug) == "" {
		return "", errors.New("invalid Slack install state payload")
	}

	oauthResp, err := s.exchangeOAuthCode(ctx, code)
	if err != nil {
		return "", err
	}

	credential := slackCredential{
		AccessToken:  strings.TrimSpace(oauthResp.AccessToken),
		RefreshToken: strings.TrimSpace(oauthResp.RefreshToken),
	}
	if oauthResp.ExpiresIn > 0 {
		expiresAt := s.clock.Now().UTC().Add(time.Duration(oauthResp.ExpiresIn) * time.Second)
		credential.ExpiresAt = &expiresAt
	}
	slackTeamID := strings.TrimSpace(oauthResp.Team.ID)
	installGeneration := uuid.New()
	binding := slackCredentialBinding{
		WorkspaceID:       nonce.WorkspaceID,
		SlackTeamID:       slackTeamID,
		InstallGeneration: installGeneration,
	}
	credentialPayload, credentialVersion, err := s.credentials.seal(binding, credential)
	if err != nil {
		return "", err
	}
	slackWorkspace, err := s.repo.UpsertSlackWorkspace(ctx, nonce.WorkspaceID, *nonce.UserID, slackOAuthInstallPayload{
		SlackTeamID:       slackTeamID,
		SlackTeamName:     strings.TrimSpace(oauthResp.Team.Name),
		SlackTeamDomain:   strings.TrimSpace(oauthResp.Team.Domain),
		BotUserID:         optionalString(oauthResp.BotUserID),
		BotAccessToken:    credentialPayload,
		CredentialVersion: credentialVersion,
		InstallGeneration: installGeneration,
		SlackAppID:        optionalString(oauthResp.AppID),
		EnterpriseID:      optionalString(oauthResp.Enterprise.ID),
		AuthedUserID:      optionalString(oauthResp.AuthedUser.ID),
		Scope:             optionalString(oauthResp.Scope),
	})
	if err != nil {
		if errors.Is(err, errWorkspaceAlreadyConnected) {
			if cleanupErr := s.cleanupRejectedOAuthInstallation(
				ctx,
				nonce.WorkspaceID,
				strings.TrimSpace(oauthResp.Team.ID),
				credentialPayload,
				credentialVersion,
				installGeneration,
			); cleanupErr != nil && s.log != nil {
				s.log.Error(ctx, "failed scheduling cleanup for rejected Slack OAuth installation", "error", cleanupErr, "workspace_id", nonce.WorkspaceID, "slack_team_id", strings.TrimSpace(oauthResp.Team.ID))
			}
		}
		return "", err
	}

	if syncErr := s.syncChannelsWithToken(
		ctx,
		nonce.WorkspaceID,
		*nonce.UserID,
		slackWorkspace.ID,
		slackWorkspace.InstallGeneration,
		credential.AccessToken,
	); syncErr != nil && s.log != nil {
		s.log.Warn(
			ctx,
			"Slack connect succeeded but initial channel sync failed",
			"error", syncErr,
			"workspace_id", nonce.WorkspaceID,
			"slack_team_id", strings.TrimSpace(oauthResp.Team.ID),
		)
	}
	if linkErr := s.autoLinkWorkspaceMembers(ctx, slackWorkspace); linkErr != nil && s.log != nil {
		s.log.Warn(
			ctx,
			"Slack connect succeeded but automatic member linking failed",
			"error", linkErr,
			"workspace_id", nonce.WorkspaceID,
			"slack_team_id", strings.TrimSpace(oauthResp.Team.ID),
		)
	}

	return s.buildWorkspaceIntegrationURL(noncePayload.WorkspaceSlug), nil
}

func (s *Service) cleanupRejectedOAuthInstallation(
	ctx context.Context,
	workspaceID uuid.UUID,
	slackTeamID, encryptedCredential string,
	credentialVersion int,
	installGeneration uuid.UUID,
) error {
	_, err := s.repo.GetSlackWorkspaceByTeamID(ctx, slackTeamID)
	if err == nil {
		// The token belongs to an app installation that is now legitimately
		// bound, whether by a concurrent winner or an uncertain commit.
		return nil
	}
	if !isSlackRepositoryNotFound(err) {
		return fmt.Errorf("verify rejected Slack OAuth team ownership: %w", err)
	}
	workspaceInstallation, err := s.repo.GetSlackWorkspace(ctx, workspaceID)
	if err == nil && strings.TrimSpace(workspaceInstallation.SlackTeamID) == slackTeamID {
		return nil
	}
	if err != nil && !isSlackRepositoryNotFound(err) {
		return fmt.Errorf("verify rejected Slack OAuth workspace ownership: %w", err)
	}

	attemptCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	uninstall, err := s.repo.EnqueueSlackUninstall(attemptCtx, slackUninstallInput{
		SlackWorkspaceID:     uuid.New(),
		WorkspaceID:          workspaceID,
		InstallGeneration:    installGeneration,
		SlackTeamID:          slackTeamID,
		UninstallKind:        "orphaned_oauth",
		CredentialPayload:    encryptedCredential,
		CredentialKeyVersion: credentialVersion,
	})
	if err != nil {
		// A direct provider call after an uncertain persistence failure is not
		// safe: a concurrent OAuth callback may have installed this team between
		// the ownership checks above and the failed enqueue. Retain the error and
		// let an operator retry the guarded cleanup path.
		return err
	}
	claimed, shouldAttempt, err := s.repo.ClaimSlackUninstall(attemptCtx, uninstall.ID)
	if err != nil || !shouldAttempt {
		return err
	}
	_, err = executeSlackUninstall(
		attemptCtx,
		s.repo,
		s.repo,
		s.slackClient(),
		s.credentials,
		s.cfg.ClientID,
		s.cfg.ClientSecret,
		s.clock.Now(),
		claimed,
	)
	return err
}

func (s *Service) SyncChannels(ctx context.Context, workspaceID, actorID uuid.UUID) error {
	if err := s.requireWorkspaceAdmin(ctx, workspaceID, actorID); err != nil {
		return err
	}
	slackWorkspace, err := s.repo.GetSlackWorkspace(ctx, workspaceID)
	if err != nil {
		return err
	}
	botToken, err := s.botToken(ctx, slackWorkspace)
	if err != nil {
		return err
	}
	if err := s.syncChannelsWithToken(
		ctx,
		workspaceID,
		actorID,
		slackWorkspace.ID,
		slackWorkspace.InstallGeneration,
		botToken,
	); err != nil {
		return err
	}
	return s.autoLinkWorkspaceMembers(ctx, slackWorkspace)
}

func (s *Service) syncChannelsWithToken(
	ctx context.Context,
	workspaceID, actorID, slackWorkspaceID, installGeneration uuid.UUID,
	botToken string,
) error {
	if workspaceID == uuid.Nil || actorID == uuid.Nil || slackWorkspaceID == uuid.Nil || installGeneration == uuid.Nil {
		return errors.New("workspace, actor, Slack installation, and generation are required")
	}
	if strings.TrimSpace(botToken) == "" {
		return errors.New("slack bot token is required")
	}
	channels, err := s.fetchChannels(ctx, botToken)
	if err != nil {
		return err
	}
	return s.repo.UpsertChannels(ctx, slackdomain.SyncChannelsCommand{
		WorkspaceID: workspaceID, ActorID: actorID,
		InstallationID: slackWorkspaceID, InstallationGeneration: installGeneration,
		Channels: channels, Now: s.clock.Now().UTC(),
	})
}

func (s *Service) DisconnectWorkspace(ctx context.Context, workspaceID, actorID uuid.UUID) error {
	if workspaceID == uuid.Nil {
		return errors.New("workspace id is required")
	}
	if err := s.requireWorkspaceAdmin(ctx, workspaceID, actorID); err != nil {
		return err
	}

	slackWorkspace, err := s.repo.GetSlackWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("load Slack installation for uninstall: %w", err)
	}
	if slackWorkspace.CredentialVersion <= 0 {
		if _, err := s.botToken(ctx, slackWorkspace); err != nil {
			return fmt.Errorf("encrypt Slack credential before disconnect: %w", err)
		}
	}

	uninstall, err := s.repo.DisconnectSlackWorkspace(ctx, slackdomain.DisconnectInstallationCommand{
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Now:         s.clock.Now().UTC(),
	})
	if err != nil {
		return err
	}
	attemptCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	claimed, shouldAttempt, err := s.repo.ClaimSlackUninstall(attemptCtx, uninstall.ID)
	if err != nil {
		if s.log != nil {
			s.log.Warn(ctx, "Slack was disconnected locally but the uninstall record could not be claimed", "error", err, "workspace_id", workspaceID, "slack_team_id", uninstall.SlackTeamID, "uninstall_id", uninstall.ID)
		}
		return nil
	}
	if !shouldAttempt {
		return nil
	}
	_, uninstallErr := executeSlackUninstall(
		attemptCtx,
		s.repo,
		s.repo,
		s.slackClient(),
		s.credentials,
		s.cfg.ClientID,
		s.cfg.ClientSecret,
		s.clock.Now(),
		claimed,
	)
	if uninstallErr != nil {
		if s.log != nil {
			s.log.Warn(ctx, "Slack was disconnected locally and remote uninstall was scheduled for retry", "error", uninstallErr, "workspace_id", workspaceID, "slack_team_id", uninstall.SlackTeamID, "uninstall_id", uninstall.ID, "attempt", claimed.AttemptCount)
		}
	}
	return nil
}
