package github

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) GetIntegration(ctx context.Context, workspaceID uuid.UUID) (CoreIntegration, error) {
	settings, err := s.repo.GetWorkspaceSettings(ctx, workspaceID)
	if err != nil {
		return CoreIntegration{}, err
	}
	installations, err := s.repo.ListInstallations(ctx, workspaceID)
	if err != nil {
		return CoreIntegration{}, err
	}
	repositories, err := s.repo.ListRepositories(ctx, workspaceID)
	if err != nil {
		return CoreIntegration{}, err
	}
	links, err := s.repo.ListIssueSyncLinks(ctx, workspaceID)
	if err != nil {
		return CoreIntegration{}, err
	}
	return CoreIntegration{
		Settings: settings, Installations: installations, Repositories: repositories, IssueSyncLinks: links,
	}, nil
}

func (s *Service) CreateInstallSession(ctx context.Context, workspaceID, userID uuid.UUID, workspaceSlug string) (CoreCreateInstallSession, error) {
	if err := s.requireWorkspaceAdmin(ctx, workspaceID, userID); err != nil {
		return CoreCreateInstallSession{}, err
	}
	if !s.canInstall() {
		return CoreCreateInstallSession{}, errors.New("github integration is not configured")
	}
	state, err := s.createInstallState(ctx, workspaceID, userID, workspaceSlug)
	if err != nil {
		return CoreCreateInstallSession{}, err
	}
	installURL := url.URL{
		Scheme: "https",
		Host:   "github.com",
		Path:   "/apps/" + strings.TrimSpace(s.cfg.AppSlug) + "/installations/new",
	}
	query := installURL.Query()
	query.Set("state", state)
	query.Set("redirect_url", strings.TrimSpace(s.cfg.RedirectURL))
	installURL.RawQuery = query.Encode()
	return CoreCreateInstallSession{InstallURL: installURL.String()}, nil
}

func (s *Service) HandleSetup(ctx context.Context, installationExternalID int64, state string) (string, error) {
	if !s.canInstall() {
		return "", errors.New("github integration is not configured")
	}
	stateRecord, statePayload, err := s.consumeInstallState(ctx, state)
	if err != nil {
		return "", err
	}
	workspaceID := *stateRecord.WorkspaceID
	userID := stateRecord.UserID
	if err := s.requireWorkspaceAdmin(ctx, workspaceID, userID); err != nil {
		return "", err
	}
	installation, err := s.getInstallation(ctx, installationExternalID)
	if err != nil {
		return "", err
	}
	repositories, err := s.listInstallationRepositories(ctx, installationExternalID)
	if err != nil {
		return "", err
	}
	if err := s.repo.UpsertInstallationWithRepositories(ctx, workspaceID, userID, s.cfg.AppID, installation, repositories); err != nil {
		return "", err
	}
	websiteURL, err := url.Parse(strings.TrimSpace(s.cfg.WebsiteURL))
	if err != nil || !validGitHubApplicationURL(s.cfg.WebsiteURL) {
		return "", errors.New("github website URL is not configured")
	}
	websiteURL = websiteURL.JoinPath(statePayload.WorkspaceSlug, "settings", "workspace", "integrations", "github")
	query := websiteURL.Query()
	query.Set("connected", "1")
	websiteURL.RawQuery = query.Encode()
	return websiteURL.String(), nil
}

func (s *Service) ResyncRepositories(ctx context.Context, workspaceID, actorID uuid.UUID) error {
	if err := s.requireWorkspaceAdmin(ctx, workspaceID, actorID); err != nil {
		return err
	}
	if !s.canUseAppAPI() {
		return errors.New("github integration is not configured")
	}
	integrations, err := s.repo.ListInstallations(ctx, workspaceID)
	if err != nil {
		return err
	}
	for _, installation := range integrations {
		if !installation.IsActive {
			continue
		}
		payload, err := s.getInstallation(ctx, installation.GitHubInstallationID)
		if err != nil {
			return err
		}
		repositories, err := s.listInstallationRepositories(ctx, installation.GitHubInstallationID)
		if err != nil {
			return err
		}
		if err := s.repo.UpsertInstallationWithRepositories(ctx, workspaceID, s.cfg.GitHubUserID, s.cfg.AppID, payload, repositories); err != nil {
			return err
		}
	}
	return nil
}
