package github

import (
	"context"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	githubrepository "github.com/complexus-tech/projects-api/internal/modules/github/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var refPattern = regexp.MustCompile(`\b([A-Za-z]{2,}[ -]?\d+)\b`)

const githubIssueLinkPrefix = "GitHub issue: "

type Service struct {
	log        *logger.Logger
	repo       *githubrepository.Repo
	stories    StoryService
	httpClient *http.Client
	cfg        Config
	privateKey *rsa.PrivateKey
}

func New(log *logger.Logger, repo *githubrepository.Repo, storyService StoryService, cfg Config) (*Service, error) {
	var privateKey *rsa.PrivateKey
	var err error
	if cfg.AppID != 0 && strings.TrimSpace(cfg.PrivateKeyPath) != "" {
		privateKey, err = loadPrivateKey(cfg.PrivateKeyPath)
		if err != nil {
			log.Warn(
				context.Background(),
				"github integration disabled: failed to load private key",
				"path",
				cfg.PrivateKeyPath,
				"error",
				err,
			)
		}
	}
	return &Service{
		log:     log,
		repo:    repo,
		stories: storyService,
		cfg:     cfg,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
		privateKey: privateKey,
	}, nil
}

func (s *Service) isConfigured() bool {
	return s.cfg.AppID != 0 &&
		strings.TrimSpace(s.cfg.AppSlug) != "" &&
		strings.TrimSpace(s.cfg.RedirectURL) != "" &&
		s.privateKey != nil
}

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
	if !s.isConfigured() {
		return CoreCreateInstallSession{}, errors.New("github integration is not configured")
	}
	state, err := s.signState(map[string]string{
		"workspace_id":   workspaceID.String(),
		"workspace_slug": workspaceSlug,
		"user_id":        userID.String(),
	})
	if err != nil {
		return CoreCreateInstallSession{}, err
	}
	installURL := fmt.Sprintf("https://github.com/apps/%s/installations/new?state=%s&redirect_url=%s", s.cfg.AppSlug, state, url.QueryEscape(s.cfg.RedirectURL))
	return CoreCreateInstallSession{InstallURL: installURL}, nil
}

func (s *Service) HandleSetup(ctx context.Context, installationExternalID int64, state string) (string, error) {
	if !s.isConfigured() {
		return "", errors.New("github integration is not configured")
	}
	values, err := s.verifyState(state)
	if err != nil {
		return "", err
	}
	workspaceID, err := uuid.Parse(values["workspace_id"])
	if err != nil {
		return "", err
	}
	userID, err := uuid.Parse(values["user_id"])
	if err != nil {
		return "", err
	}
	workspaceSlug := values["workspace_slug"]
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
	return fmt.Sprintf("%s/%s/settings/workspace/integrations/github?connected=1", strings.TrimRight(s.cfg.WebsiteURL, "/"), workspaceSlug), nil
}

func (s *Service) ResyncRepositories(ctx context.Context, workspaceID uuid.UUID) error {
	if !s.isConfigured() {
		return errors.New("github integration is not configured")
	}
	integrations, err := s.repo.ListInstallations(ctx, workspaceID)
	if err != nil {
		return err
	}
	for _, installation := range integrations {
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

func (s *Service) UpdateWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID, input CoreUpdateWorkspaceSettingsInput) (CoreWorkspaceSettings, error) {
	if input.BranchFormat != nil && strings.TrimSpace(*input.BranchFormat) == "" {
		return CoreWorkspaceSettings{}, errors.New("branch format is required")
	}
	return s.repo.UpdateWorkspaceSettings(ctx, workspaceID, input)
}

func (s *Service) CreateIssueSyncLink(ctx context.Context, workspaceID, userID uuid.UUID, input CoreIssueSyncLinkInput) (CoreIssueSyncLink, error) {
	if !slices.Contains([]string{SyncDirectionInboundOnly, SyncDirectionBidirectional}, input.SyncDirection) {
		return CoreIssueSyncLink{}, errors.New("invalid sync direction")
	}
	link, err := s.repo.CreateIssueSyncLink(ctx, workspaceID, userID, input)
	if err != nil {
		return CoreIssueSyncLink{}, err
	}
	settings, err := s.repo.GetTeamWorkflowSettings(ctx, workspaceID, input.TeamID)
	if err != nil {
		return CoreIssueSyncLink{}, err
	}
	if len(settings.Rules) == 0 {
		if _, err := s.seedDefaultWorkflowRules(ctx, workspaceID, input.TeamID); err != nil {
			return CoreIssueSyncLink{}, err
		}
	}
	return link, nil
}

func (s *Service) UpdateIssueSyncLink(ctx context.Context, workspaceID, linkID uuid.UUID, input CoreUpdateIssueSyncLinkInput) (CoreIssueSyncLink, error) {
	if input.SyncDirection != nil && !slices.Contains([]string{SyncDirectionInboundOnly, SyncDirectionBidirectional}, *input.SyncDirection) {
		return CoreIssueSyncLink{}, errors.New("invalid sync direction")
	}
	return s.repo.UpdateIssueSyncLink(ctx, workspaceID, linkID, input)
}

func (s *Service) DeleteIssueSyncLink(ctx context.Context, workspaceID, linkID uuid.UUID) error {
	return s.repo.DeleteIssueSyncLink(ctx, workspaceID, linkID)
}

func (s *Service) GetTeamSettings(ctx context.Context, workspaceID, teamID uuid.UUID) (CoreTeamGitHubSettings, error) {
	settings, err := s.repo.GetTeamWorkflowSettings(ctx, workspaceID, teamID)
	if err != nil {
		return CoreTeamGitHubSettings{}, err
	}
	if len(settings.Rules) == 0 {
		return s.seedDefaultWorkflowRules(ctx, workspaceID, teamID)
	}
	return settings, nil
}

func (s *Service) UpdateTeamSettings(ctx context.Context, workspaceID, teamID uuid.UUID, input CoreUpdateTeamGitHubSettings) (CoreTeamGitHubSettings, error) {
	if len(input.Rules) == 0 {
		return CoreTeamGitHubSettings{}, errors.New("at least one rule is required")
	}
	return s.repo.ReplaceTeamWorkflowSettings(ctx, workspaceID, teamID, input.Rules)
}

func (s *Service) SyncStoryFromFortyOne(ctx context.Context, input CoreStorySyncInput) error {
	if !s.isConfigured() {
		return nil
	}

	link, err := s.repo.FindBidirectionalIssueSyncLinkByTeamID(ctx, input.WorkspaceID, input.TeamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	desiredState := "open"
	if input.StatusID != nil {
		category, err := s.repo.GetStatusCategory(ctx, *input.StatusID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if category == "completed" {
			desiredState = "closed"
		}
	}

	issueBody := issueBodyFromStoryDescription(input.Description)
	existingLink, err := s.repo.FindIssueStoryLinkByStoryID(ctx, input.StoryID, link.RepositoryID)
	switch {
	case err == nil:
		issue, err := s.updateIssue(ctx, link.GitHubInstallationID, link.OwnerLogin, link.RepositorySlug, existingLink.GitHubNumber, input.Title, issueBody, desiredState)
		if err != nil {
			return err
		}
		if err := s.repo.UpsertStoryLink(ctx, input.WorkspaceID, input.StoryID, link.RepositoryID, "issue", issue.ID, issue.Number, nil, issue.HTMLURL, issue.Title, issue.State, issue); err != nil {
			return err
		}
		return s.repo.EnsureStoryLink(ctx, input.StoryID, githubIssueStoryLinkTitle(issue.Number), issue.HTMLURL)
	case !errors.Is(err, sql.ErrNoRows):
		return err
	}

	issue, err := s.createIssue(ctx, link.GitHubInstallationID, link.OwnerLogin, link.RepositorySlug, input.Title, issueBody)
	if err != nil {
		return err
	}
	if desiredState == "closed" {
		issue, err = s.updateIssue(ctx, link.GitHubInstallationID, link.OwnerLogin, link.RepositorySlug, issue.Number, input.Title, issueBody, desiredState)
		if err != nil {
			return err
		}
	}
	if err := s.repo.UpsertStoryLink(ctx, input.WorkspaceID, input.StoryID, link.RepositoryID, "issue", issue.ID, issue.Number, nil, issue.HTMLURL, issue.Title, issue.State, issue); err != nil {
		return err
	}
	return s.repo.EnsureStoryLink(ctx, input.StoryID, githubIssueStoryLinkTitle(issue.Number), issue.HTMLURL)
}

func (s *Service) HandleWebhook(ctx context.Context, deliveryID, eventName, signature string, body []byte) error {
	if !s.isConfigured() {
		return errors.New("github integration is not configured")
	}
	if err := s.verifyWebhookSignature(body, signature); err != nil {
		return err
	}

	var payload webhookEnvelope
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}

	recorded, err := s.repo.RecordWebhookEvent(ctx, deliveryID, eventName, payload.Action, payload.installationID(), payload.repositoryID(), payload.senderID(), body)
	if err != nil || !recorded {
		return err
	}

	if err := s.processWebhook(ctx, eventName, payload); err != nil {
		_ = s.repo.MarkWebhookFailed(ctx, deliveryID, err.Error())
		return err
	}
	return s.repo.MarkWebhookProcessed(ctx, deliveryID)
}

func (s *Service) processWebhook(ctx context.Context, eventName string, payload webhookEnvelope) error {
	repository, err := s.repo.FindRepositoryByExternalID(ctx, payload.Repository.ID)
	if err != nil {
		return nil
	}

	switch eventName {
	case "issues":
		return s.handleIssueEvent(ctx, repository, payload)
	case "pull_request":
		return s.handlePullRequestEvent(ctx, repository, payload)
	case "push":
		return s.handlePushEvent(ctx, repository, payload)
	default:
		return nil
	}
}

func (s *Service) handleIssueEvent(ctx context.Context, repository githubrepository.RepoByExternalRow, payload webhookEnvelope) error {
	if s.stories == nil {
		return errors.New("stories service is not configured")
	}

	link, err := s.repo.FindIssueSyncLinkByRepositoryID(ctx, repository.ID)
	if err != nil {
		return nil
	}

	issueDescription := githubString(payload.Issue.Body)
	_, storyID, err := s.repo.FindStoryLink(ctx, repository.ID, "issue", payload.Issue.ID, nil)
	switch {
	case err == nil:
		story, getErr := s.stories.Get(ctx, storyID, repository.WorkspaceID)
		if getErr != nil {
			return getErr
		}
		updates := map[string]any{
			"title":       payload.Issue.Title,
			"description": issueDescription,
		}
		if updateErr := s.stories.UpdateExternal(ctx, s.cfg.GitHubUserID, story.ID, repository.WorkspaceID, updates); updateErr != nil {
			return updateErr
		}
	case errors.Is(err, sql.ErrNoRows):
		statusID, statusErr := s.repo.FindFirstStatusByCategory(ctx, link.TeamID, "unstarted")
		if statusErr != nil {
			return statusErr
		}
		if statusID == nil {
			return errors.New("team has no unstarted status configured")
		}

		story, createErr := s.stories.CreateExternal(ctx, s.cfg.GitHubUserID, stories.CoreNewStory{
			Title:       payload.Issue.Title,
			Description: issueDescription,
			Status:      statusID,
			Reporter:    &s.cfg.GitHubUserID,
			Team:        link.TeamID,
			Priority:    "No Priority",
		}, repository.WorkspaceID)
		if createErr != nil {
			return createErr
		}
		storyID = story.ID
	default:
		return err
	}

	linkCreated, err := s.upsertStoryLink(ctx, repository.WorkspaceID, storyID, repository.ID, "issue", payload.Issue.ID, payload.Issue.Number, nil, payload.Issue.HTMLURL, payload.Issue.Title, payload.Issue.State, payload.Issue)
	if err != nil {
		return err
	}
	if err := s.repo.EnsureStoryLink(ctx, storyID, githubIssueStoryLinkTitle(payload.Issue.Number), payload.Issue.HTMLURL); err != nil {
		return err
	}
	if linkCreated {
		if err := s.recordLinkActivity(ctx, repository.WorkspaceID, storyID, "github_issue", fmt.Sprintf("issue %d", payload.Issue.Number), payload.Issue.HTMLURL); err != nil {
			return err
		}
	}
	switch payload.Action {
	case "opened":
		return s.moveStoryByRule(ctx, repository.WorkspaceID, link.TeamID, storyID, EventIssueOpen, nil)
	case "reopened":
		return s.moveStoryByRule(ctx, repository.WorkspaceID, link.TeamID, storyID, EventIssueReopen, nil)
	case "closed":
		return s.moveStoryByRule(ctx, repository.WorkspaceID, link.TeamID, storyID, EventIssueClose, nil)
	case "edited":
		return nil
	default:
		return nil
	}
}

func (s *Service) handlePullRequestEvent(ctx context.Context, repository githubrepository.RepoByExternalRow, payload webhookEnvelope) error {
	if s.stories == nil {
		return errors.New("stories service is not configured")
	}

	refs := extractStoryRefs(payload.PullRequest.Title, payload.PullRequest.Body, payload.PullRequest.Head.Ref)
	stories, err := s.repo.ResolveStoriesByRefs(ctx, repository.WorkspaceID, refs)
	if err != nil || len(stories) == 0 {
		return err
	}

	var eventKey string
	switch {
	case payload.Action == "closed" && payload.PullRequest.Merged:
		eventKey = EventPRMerge
	case payload.Action == "ready_for_review":
		eventKey = EventPRReadyForMerge
	case payload.PullRequest.Draft:
		eventKey = EventDraftPROpen
	default:
		eventKey = EventPROpen
	}

	for _, story := range stories {
		prLinkCreated, err := s.upsertStoryLink(ctx, repository.WorkspaceID, story.StoryID, repository.ID, "pull_request", payload.PullRequest.ID, payload.PullRequest.Number, nil, payload.PullRequest.HTMLURL, payload.PullRequest.Title, payload.PullRequest.State, payload.PullRequest)
		if err != nil {
			return err
		}
		if prLinkCreated {
			if err := s.recordLinkActivity(ctx, repository.WorkspaceID, story.StoryID, "github_pull_request", fmt.Sprintf("PR #%d", payload.PullRequest.Number), payload.PullRequest.HTMLURL); err != nil {
				return err
			}
		}
		branchRef := payload.PullRequest.Head.Ref
		if branchRef != "" {
			branchLinkCreated, err := s.upsertStoryLink(ctx, repository.WorkspaceID, story.StoryID, repository.ID, "branch", 0, 0, &branchRef, payload.PullRequest.Head.HTMLURL, payload.PullRequest.Head.Ref, payload.PullRequest.State, payload.PullRequest.Head)
			if err != nil {
				return err
			}
			if branchLinkCreated {
				if err := s.recordLinkActivity(ctx, repository.WorkspaceID, story.StoryID, "github_branch", fmt.Sprintf("branch %s", branchRef), payload.PullRequest.Head.HTMLURL); err != nil {
					return err
				}
			}
		}
		if err := s.moveStoryByRule(ctx, repository.WorkspaceID, story.TeamID, story.StoryID, eventKey, &payload.PullRequest.Base.Ref); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) handlePushEvent(ctx context.Context, repository githubrepository.RepoByExternalRow, payload webhookEnvelope) error {
	if s.stories == nil {
		return errors.New("stories service is not configured")
	}

	refs := extractStoryRefs(payload.Ref)
	for _, commit := range payload.Commits {
		refs = append(refs, extractStoryRefs(commit.Message)...)
	}
	stories, err := s.repo.ResolveStoriesByRefs(ctx, repository.WorkspaceID, refs)
	if err != nil || len(stories) == 0 {
		return err
	}
	branch := strings.TrimPrefix(payload.Ref, "refs/heads/")
	for _, story := range stories {
		newCommits := 0
		if branch != "" {
			branchLinkCreated, err := s.upsertStoryLink(ctx, repository.WorkspaceID, story.StoryID, repository.ID, "branch", 0, 0, &branch, payload.Repository.HTMLURL+"/tree/"+branch, branch, "active", map[string]any{"ref": payload.Ref})
			if err != nil {
				return err
			}
			if branchLinkCreated {
				if err := s.recordLinkActivity(ctx, repository.WorkspaceID, story.StoryID, "github_branch", fmt.Sprintf("branch %s", branch), payload.Repository.HTMLURL+"/tree/"+branch); err != nil {
					return err
				}
			}
		}
		for _, commit := range payload.Commits {
			refName := commit.ID
			commitLinkCreated, err := s.upsertStoryLink(ctx, repository.WorkspaceID, story.StoryID, repository.ID, "commit", 0, 0, &refName, commit.URL, commit.Message, "linked", commit)
			if err != nil {
				return err
			}
			if commitLinkCreated {
				newCommits++
			}
		}
		if newCommits > 0 {
			label := "1 commit"
			if newCommits > 1 {
				label = fmt.Sprintf("%d commits", newCommits)
			}
			if err := s.recordLinkActivity(ctx, repository.WorkspaceID, story.StoryID, "github_commit", label, payload.Repository.HTMLURL+"/commits/"+branch); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) moveStoryByRule(ctx context.Context, workspaceID, teamID, storyID uuid.UUID, eventKey string, baseBranch *string) error {
	settings, err := s.GetTeamSettings(ctx, workspaceID, teamID)
	if err != nil {
		return err
	}
	var statusID *uuid.UUID
	for _, rule := range settings.Rules {
		if rule.EventKey != eventKey || !rule.IsActive {
			continue
		}
		if rule.BaseBranchPattern == nil {
			statusID = rule.TargetStatusID
			continue
		}
		if baseBranch != nil && matchBranchPattern(*rule.BaseBranchPattern, *baseBranch) {
			statusID = rule.TargetStatusID
			break
		}
	}
	if statusID == nil {
		return nil
	}
	story, err := s.stories.Get(ctx, storyID, workspaceID)
	if err != nil {
		return err
	}
	if story.Status != nil && *story.Status == *statusID {
		return nil
	}
	return s.stories.UpdateExternal(ctx, s.cfg.GitHubUserID, storyID, workspaceID, map[string]any{
		"status_id": *statusID,
	})
}

func (s *Service) seedDefaultWorkflowRules(ctx context.Context, workspaceID, teamID uuid.UUID) (CoreTeamGitHubSettings, error) {
	statuses, err := s.repo.ListTeamStatuses(ctx, teamID)
	if err != nil {
		return CoreTeamGitHubSettings{}, err
	}
	findCategory := func(category string) *uuid.UUID {
		for _, status := range statuses {
			if status.Category == category {
				id := status.ID
				return &id
			}
		}
		return nil
	}
	findReview := func() *uuid.UUID {
		for _, status := range statuses {
			if strings.Contains(strings.ToLower(status.Name), "review") {
				id := status.ID
				return &id
			}
		}
		return nil
	}
	rules := []CoreWorkflowRuleInput{
		{EventKey: EventDraftPROpen, TargetStatusID: nil, IsActive: true},
		{EventKey: EventPROpen, TargetStatusID: findCategory("started"), IsActive: true},
		{EventKey: EventPRReviewActivity, TargetStatusID: findReview(), IsActive: true},
		{EventKey: EventPRReadyForMerge, TargetStatusID: nil, IsActive: true},
		{EventKey: EventPRMerge, TargetStatusID: findCategory("completed"), IsActive: true},
		{EventKey: EventIssueOpen, TargetStatusID: findCategory("unstarted"), IsActive: true},
		{EventKey: EventIssueReopen, TargetStatusID: findCategory("unstarted"), IsActive: true},
		{EventKey: EventIssueClose, TargetStatusID: findCategory("completed"), IsActive: true},
	}
	return s.repo.ReplaceTeamWorkflowSettings(ctx, workspaceID, teamID, rules)
}

func extractStoryRefs(values ...string) []string {
	items := make([]string, 0)
	seen := map[string]struct{}{}
	for _, value := range values {
		for _, match := range refPattern.FindAllString(value, -1) {
			normalized := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(match), "-", ""), " ", ""))
			if normalized == "" {
				continue
			}
			if _, ok := seen[normalized]; ok {
				continue
			}
			seen[normalized] = struct{}{}
			items = append(items, normalized)
		}
	}
	return items
}

func matchBranchPattern(pattern, branch string) bool {
	if pattern == branch {
		return true
	}
	if strings.HasSuffix(pattern, "/*") {
		return strings.HasPrefix(branch, strings.TrimSuffix(pattern, "*"))
	}
	return false
}

func (s *Service) upsertStoryLink(ctx context.Context, workspaceID, storyID, repositoryID uuid.UUID, externalType string, githubID int64, githubNumber int, refName *string, url, title, state string, metadata any) (bool, error) {
	_, _, err := s.repo.FindStoryLink(ctx, repositoryID, externalType, githubID, refName)
	linkCreated := errors.Is(err, sql.ErrNoRows)
	if err != nil && !linkCreated {
		return false, err
	}
	if err := s.repo.UpsertStoryLink(ctx, workspaceID, storyID, repositoryID, externalType, githubID, githubNumber, refName, url, title, state, metadata); err != nil {
		return false, err
	}
	return linkCreated, nil
}

func (s *Service) recordLinkActivity(ctx context.Context, workspaceID, storyID uuid.UUID, field, currentValue, targetURL string) error {
	if s.stories == nil {
		return errors.New("stories service is not configured")
	}

	return s.stories.RecordActivity(ctx, stories.CoreActivity{
		StoryID:      storyID,
		Type:         "link",
		Field:        field,
		CurrentValue: currentValue,
		NewValue:     targetURL,
		UserID:       s.cfg.GitHubUserID,
		WorkspaceID:  workspaceID,
	})
}

func githubIssueStoryLinkTitle(issueNumber int) *string {
	title := fmt.Sprintf("GitHub issue #%d", issueNumber)
	return &title
}

func githubString(value string) *string {
	return &value
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(bytes)
	if block == nil {
		return nil, errors.New("invalid github private key")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("github private key is not RSA")
	}
	return rsaKey, nil
}

func (s *Service) createAppJWT() (string, error) {
	if s.privateKey == nil || s.cfg.AppID == 0 {
		return "", errors.New("github app credentials are not configured")
	}
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": now.Add(-time.Minute).Unix(),
		"exp": now.Add(9 * time.Minute).Unix(),
		"iss": s.cfg.AppID,
	})
	return token.SignedString(s.privateKey)
}

func (s *Service) getInstallationToken(ctx context.Context, installationID int64) (string, error) {
	appJWT, err := s.createAppJWT()
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", installationID), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+appJWT)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("github installation token request failed: %s", body)
	}
	var payload struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	return payload.Token, nil
}

func (s *Service) getInstallation(ctx context.Context, installationID int64) (githubrepository.GithubInstallationPayload, error) {
	appJWT, err := s.createAppJWT()
	if err != nil {
		return githubrepository.GithubInstallationPayload{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://api.github.com/app/installations/%d", installationID), nil)
	if err != nil {
		return githubrepository.GithubInstallationPayload{}, err
	}
	req.Header.Set("Authorization", "Bearer "+appJWT)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return githubrepository.GithubInstallationPayload{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		return githubrepository.GithubInstallationPayload{}, fmt.Errorf("github installation lookup failed: %s", body)
	}
	var payload githubrepository.GithubInstallationPayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return githubrepository.GithubInstallationPayload{}, err
	}
	return payload, nil
}

func (s *Service) listInstallationRepositories(ctx context.Context, installationID int64) ([]githubrepository.GithubRepositoryPayload, error) {
	token, err := s.getInstallationToken(ctx, installationID)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/installation/repositories?per_page=100", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github repository listing failed: %s", body)
	}
	var payload struct {
		Repositories []githubrepository.GithubRepositoryPayload `json:"repositories"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return payload.Repositories, nil
}

type githubIssuePayload struct {
	ID      int64  `json:"id"`
	Number  int    `json:"number"`
	HTMLURL string `json:"html_url"`
	Title   string `json:"title"`
	State   string `json:"state"`
}

func (s *Service) createIssue(ctx context.Context, installationID int64, owner, repository, title, body string) (githubIssuePayload, error) {
	requestBody := map[string]any{
		"title": title,
		"body":  body,
	}
	var issue githubIssuePayload
	if err := s.doInstallationRequest(
		ctx,
		installationID,
		http.MethodPost,
		fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", owner, repository),
		requestBody,
		&issue,
	); err != nil {
		return githubIssuePayload{}, err
	}
	return issue, nil
}

func (s *Service) updateIssue(ctx context.Context, installationID int64, owner, repository string, number int, title, body, state string) (githubIssuePayload, error) {
	requestBody := map[string]any{
		"title": title,
		"body":  body,
		"state": state,
	}
	var issue githubIssuePayload
	if err := s.doInstallationRequest(
		ctx,
		installationID,
		http.MethodPatch,
		fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", owner, repository, number),
		requestBody,
		&issue,
	); err != nil {
		return githubIssuePayload{}, err
	}
	return issue, nil
}

func (s *Service) doInstallationRequest(ctx context.Context, installationID int64, method, requestURL string, payload any, target any) error {
	token, err := s.getInstallationToken(ctx, installationID)
	if err != nil {
		return err
	}

	var body io.Reader
	if payload != nil {
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = strings.NewReader(string(payloadBytes))
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("github request failed: %s", responseBody)
	}
	if target == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func (s *Service) verifyWebhookSignature(body []byte, signature string) error {
	if strings.TrimSpace(s.cfg.WebhookSecret) == "" {
		return nil
	}
	if signature == "" {
		return errors.New("missing github webhook signature")
	}
	mac := hmac.New(sha256.New, []byte(s.cfg.WebhookSecret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return errors.New("invalid github webhook signature")
	}
	return nil
}

func issueBodyFromStoryDescription(description *string) string {
	if description == nil {
		return ""
	}
	return stripManagedIssueLink(*description)
}

func ensureIssueLinkInDescription(description *string, issueURL string) *string {
	base := ""
	if description != nil {
		base = stripManagedIssueLink(*description)
	}
	footer := githubIssueLinkPrefix + issueURL
	if strings.TrimSpace(base) == "" {
		value := footer
		return &value
	}
	value := strings.TrimSpace(base) + "\n\n" + footer
	return &value
}

func stripManagedIssueLink(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}

	lines := strings.Split(value, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, githubIssueLinkPrefix) {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.TrimSpace(strings.Join(filtered, "\n"))
}

func (s *Service) signState(values map[string]string) (string, error) {
	payloadBytes, err := json.Marshal(values)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	mac := hmac.New(sha256.New, []byte(s.cfg.SecretKey))
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	return payload + "." + sig, nil
}

func (s *Service) verifyState(value string) (map[string]string, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return nil, errors.New("invalid github setup state")
	}
	mac := hmac.New(sha256.New, []byte(s.cfg.SecretKey))
	mac.Write([]byte(parts[0]))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[1])) {
		return nil, errors.New("invalid github setup state signature")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	var values map[string]string
	if err := json.Unmarshal(payloadBytes, &values); err != nil {
		return nil, err
	}
	return values, nil
}

type webhookEnvelope struct {
	Action     string `json:"action"`
	Repository struct {
		ID      int64  `json:"id"`
		HTMLURL string `json:"html_url"`
	} `json:"repository"`
	Installation struct {
		ID int64 `json:"id"`
	} `json:"installation"`
	Sender struct {
		ID int64 `json:"id"`
	} `json:"sender"`
	Issue struct {
		ID      int64  `json:"id"`
		Number  int    `json:"number"`
		Title   string `json:"title"`
		Body    string `json:"body"`
		State   string `json:"state"`
		HTMLURL string `json:"html_url"`
	} `json:"issue"`
	PullRequest struct {
		ID      int64  `json:"id"`
		Number  int    `json:"number"`
		Title   string `json:"title"`
		Body    string `json:"body"`
		State   string `json:"state"`
		HTMLURL string `json:"html_url"`
		Draft   bool   `json:"draft"`
		Merged  bool   `json:"merged"`
		Head    struct {
			Ref     string `json:"ref"`
			HTMLURL string `json:"html_url"`
		} `json:"head"`
		Base struct {
			Ref string `json:"ref"`
		} `json:"base"`
	} `json:"pull_request"`
	Ref     string `json:"ref"`
	Commits []struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		URL     string `json:"url"`
	} `json:"commits"`
}

func (w webhookEnvelope) installationID() *int64 {
	if w.Installation.ID == 0 {
		return nil
	}
	return &w.Installation.ID
}

func (w webhookEnvelope) repositoryID() *int64 {
	if w.Repository.ID == 0 {
		return nil
	}
	return &w.Repository.ID
}

func (w webhookEnvelope) senderID() *int64 {
	if w.Sender.ID == 0 {
		return nil
	}
	return &w.Sender.ID
}
