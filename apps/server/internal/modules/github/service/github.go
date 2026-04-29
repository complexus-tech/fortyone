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
	"net"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	githubrepository "github.com/complexus-tech/projects-api/internal/modules/github/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/internal/platform/actors"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
	githubsdk "github.com/google/go-github/v72/github"
	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
)

var refPattern = regexp.MustCompile(`\b([A-Za-z]{2,}[ -]?\d+)\b`)

const githubIssueLinkPrefix = "GitHub issue: "

type Service struct {
	log                 *logger.Logger
	repo                *githubrepository.Repo
	stories             StoryService
	avatars             AvatarResolver
	httpClient          *http.Client
	cfg                 Config
	privateKey          *rsa.PrivateKey
	privateKeyLoadError string
}

func New(log *logger.Logger, repo *githubrepository.Repo, storyService StoryService, avatarResolver AvatarResolver, cfg Config) (*Service, error) {
	var privateKey *rsa.PrivateKey
	var err error
	if cfg.AppID != 0 && strings.TrimSpace(cfg.PrivateKeyBase64) != "" {
		privateKey, err = loadPrivateKey(cfg.PrivateKeyBase64)
		if err != nil {
			log.Warn(
				context.Background(),
				"github integration disabled: failed to load private key",
				"app_id_configured",
				cfg.AppID != 0,
				"private_key_base64_present",
				strings.TrimSpace(cfg.PrivateKeyBase64) != "",
				"private_key_base64_length",
				len(cfg.PrivateKeyBase64),
				"error",
				err,
			)
		}
	}
	return &Service{
		log:     log,
		repo:    repo,
		stories: storyService,
		avatars: avatarResolver,
		cfg:     cfg,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
		privateKey:          privateKey,
		privateKeyLoadError: errorString(err),
	}, nil
}

func (s *Service) canInstall() bool {
	return s.cfg.AppID != 0 &&
		strings.TrimSpace(s.cfg.AppSlug) != "" &&
		strings.TrimSpace(s.cfg.RedirectURL) != "" &&
		s.privateKey != nil
}

func (s *Service) canUseAppAPI() bool {
	return s.cfg.AppID != 0 && s.privateKey != nil
}

func (s *Service) canVerifyWebhooks() bool {
	return strings.TrimSpace(s.cfg.WebhookSecret) != ""
}

func (s *Service) HasAnyConfig() bool {
	return s.cfg.AppID != 0 ||
		strings.TrimSpace(s.cfg.AppSlug) != "" ||
		strings.TrimSpace(s.cfg.PrivateKeyBase64) != "" ||
		strings.TrimSpace(s.cfg.RedirectURL) != "" ||
		strings.TrimSpace(s.cfg.WebhookSecret) != ""
}

func (s *Service) ValidateWorkerConfiguration() error {
	if !s.HasAnyConfig() {
		return nil
	}
	if s.canUseAppAPI() {
		return nil
	}
	return fmt.Errorf(
		"github worker configuration invalid: app_id_configured=%t app_slug_configured=%t private_key_present=%t private_key_loaded=%t private_key_load_error=%q redirect_url_configured=%t webhook_secret_configured=%t",
		s.cfg.AppID != 0,
		strings.TrimSpace(s.cfg.AppSlug) != "",
		strings.TrimSpace(s.cfg.PrivateKeyBase64) != "",
		s.privateKey != nil,
		s.privateKeyLoadError,
		strings.TrimSpace(s.cfg.RedirectURL) != "",
		strings.TrimSpace(s.cfg.WebhookSecret) != "",
	)
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
	if !s.canInstall() {
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
	if !s.canInstall() {
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
	if !s.canUseAppAPI() {
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
	if input.BranchFormat != nil && !slices.Contains([]string{
		BranchFormatUsernameIdentifierTitle,
		BranchFormatIdentifierTitle,
		BranchFormatIdentifierSlashTitle,
	}, *input.BranchFormat) {
		return CoreWorkspaceSettings{}, errors.New("invalid branch format")
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

func (s *Service) GetStoryGitHubLinks(ctx context.Context, workspaceID, storyID uuid.UUID) ([]githubrepository.StoryGitHubLink, error) {
	return s.repo.GetStoryGitHubLinks(ctx, workspaceID, storyID)
}

func (s *Service) DeleteStoryGitHubLink(ctx context.Context, workspaceID, linkID uuid.UUID) error {
	return s.repo.DeleteStoryGitHubLink(ctx, workspaceID, linkID)
}

const avatarURLExpiry = 24 * time.Hour

// resolveAvatarURL converts a stored avatar blob name to a signed URL.
// If the avatar is already a full URL or the resolver is not configured, it returns as-is.
func (s *Service) resolveAvatarURL(ctx context.Context, avatar *string) string {
	if avatar == nil || *avatar == "" {
		return ""
	}
	if s.avatars == nil {
		return *avatar
	}
	resolved, err := s.avatars.ResolveProfileImageURL(ctx, *avatar, avatarURLExpiry)
	if err != nil {
		return ""
	}
	return resolved
}

// GitHubComment represents a single comment fetched from the GitHub API.
type GitHubComment struct {
	ID         int64  `json:"id"`
	Body       string `json:"body"`
	UserLogin  string `json:"userLogin"`
	UserAvatar string `json:"userAvatar"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
	HTMLURL    string `json:"htmlUrl"`
}

// Regex to parse bot attribution: **Name** commented via/on FortyOne:\n\nbody
var fortyOneCommentPattern = regexp.MustCompile(`(?s)\A\*\*(.+?)\*\* commented (?:via|on) FortyOne:\s*\n\n?(.*)`)
var fortyOneCommentMarkerPattern = regexp.MustCompile(`(?m)\n*\s*<!--\s*fortyone:comment:[0-9a-fA-F-]{36}\s*-->\s*`)

// GetStoryGitHubComments fetches comments from all linked GitHub issues for a story.
// It resolves GitHub users to FortyOne users when possible, and for bot comments
// posted via FortyOne it extracts the real author and strips the attribution prefix.
func (s *Service) GetStoryGitHubComments(ctx context.Context, workspaceID, storyID uuid.UUID) ([]GitHubComment, error) {
	if !s.canUseAppAPI() {
		return []GitHubComment{}, nil
	}

	issues, err := s.repo.GetStoryLinkedIssues(ctx, workspaceID, storyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get linked issues: %w", err)
	}
	if len(issues) == 0 {
		return []GitHubComment{}, nil
	}

	// Collect raw comments and unique GitHub user IDs.
	type rawComment struct {
		comment      *githubsdk.IssueComment
		gitHubUserID int64
		isAppAuthor  bool
	}
	var rawComments []rawComment

	for _, issue := range issues {
		client, err := s.newInstallationClient(ctx, issue.GitHubInstallationID)
		if err != nil {
			s.log.Warn(ctx, "failed to create installation client for fetching comments", "error", err)
			continue
		}

		opts := &githubsdk.IssueListCommentsOptions{
			Sort:        githubsdk.String("created"),
			Direction:   githubsdk.String("asc"),
			ListOptions: githubsdk.ListOptions{PerPage: 100},
		}
		comments, _, err := client.Issues.ListComments(ctx, issue.OwnerLogin, issue.RepositorySlug, issue.GitHubNumber, opts)
		if err != nil {
			s.log.Warn(ctx, "failed to fetch github comments", "issue", issue.GitHubNumber, "error", err)
			continue
		}

		var appOwnerID int64
		appBotLogin := ""
		app, _, err := client.Apps.Get(ctx, "")
		if err == nil {
			if app.GetOwner() != nil {
				appOwnerID = app.GetOwner().GetID()
			}
			appSlug := strings.TrimSpace(app.GetSlug())
			if appSlug == "" {
				appSlug = strings.TrimSpace(s.cfg.AppSlug)
			}
			if appSlug != "" {
				appBotLogin = strings.ToLower(appSlug + "[bot]")
			}
		}

		for _, c := range comments {
			if c == nil {
				continue
			}
			var ghUID int64
			ghLogin := ""
			if c.User != nil {
				ghUID = c.User.GetID()
				ghLogin = strings.ToLower(strings.TrimSpace(c.User.GetLogin()))
			}
			isAppAuthor := appOwnerID != 0 && ghUID != 0 && ghUID == appOwnerID
			if !isAppAuthor && appBotLogin != "" && ghLogin != "" {
				isAppAuthor = ghLogin == appBotLogin
			}
			rawComments = append(rawComments, rawComment{
				comment:      c,
				gitHubUserID: ghUID,
				isAppAuthor:  isAppAuthor,
			})
		}
	}

	// Batch-resolve GitHub user IDs → FortyOne users.
	uniqueIDs := make(map[int64]struct{})
	for _, rc := range rawComments {
		if rc.gitHubUserID != 0 {
			uniqueIDs[rc.gitHubUserID] = struct{}{}
		}
	}
	idSlice := make([]int64, 0, len(uniqueIDs))
	for id := range uniqueIDs {
		idSlice = append(idSlice, id)
	}
	userMap, _ := s.repo.ResolveFortyOneUsersByGitHubIDs(ctx, idSlice)
	if userMap == nil {
		userMap = map[int64]githubrepository.FortyOneUser{}
	}
	systemGitHubLogin := "github"
	systemGitHubAvatar := ""
	if githubActorEmail, ok := actors.EmailForKey(actors.KeyGitHub); ok {
		if systemUser, err := s.repo.ResolveFortyOneUserByEmail(ctx, githubActorEmail); err == nil {
			if username := strings.TrimSpace(systemUser.Username); username != "" {
				systemGitHubLogin = username
			}
			systemGitHubAvatar = s.resolveAvatarURL(ctx, systemUser.AvatarURL)
		}
	}

	// Build result, resolving authors.
	allComments := make([]GitHubComment, 0, len(rawComments))
	for _, rc := range rawComments {
		c := rc.comment
		gc := GitHubComment{
			ID:        c.GetID(),
			Body:      c.GetBody(),
			CreatedAt: c.GetCreatedAt().Format(time.RFC3339),
			UpdatedAt: c.GetUpdatedAt().Format(time.RFC3339),
			HTMLURL:   c.GetHTMLURL(),
		}
		rawGitHubLogin := ""
		if c.User != nil {
			rawGitHubLogin = c.User.GetLogin()
			gc.UserLogin = rawGitHubLogin
			gc.UserAvatar = c.User.GetAvatarURL()
		}

		// Check if this GitHub user is a linked FortyOne user.
		if fu, ok := userMap[rc.gitHubUserID]; ok {
			gc.UserLogin = fu.Username
			gc.UserAvatar = s.resolveAvatarURL(ctx, fu.AvatarURL)
		}

		commentedViaFortyOne := false
		// For app and bot comments, extract the real author when the attribution marker exists.
		if rc.isAppAuthor || strings.HasSuffix(rawGitHubLogin, "[bot]") {
			if match := fortyOneCommentPattern.FindStringSubmatch(gc.Body); match != nil {
				commentedViaFortyOne = true
				authorName := match[1]
				gc.Body = stripFortyOneCommentMarker(match[2])

				// Try to resolve the author name to a FortyOne user.
				if fu, err := s.repo.ResolveFortyOneUserByFullName(ctx, authorName); err == nil {
					gc.UserLogin = fu.Username
					gc.UserAvatar = s.resolveAvatarURL(ctx, fu.AvatarURL)
				} else {
					gc.UserLogin = authorName
					gc.UserAvatar = ""
				}
			}
		}
		gc.Body = stripFortyOneCommentMarker(gc.Body)
		// For app-authored/system comments without a resolvable author, expose a
		// stable GitHub label instead of the app account login (e.g. fortyone-app[bot]).
		if !commentedViaFortyOne && (rc.isAppAuthor ||
			isFortyOneBotAuthorLogin(rawGitHubLogin, s.cfg.AppSlug) ||
			isFortyOneSystemLinkedTaskComment(gc.Body)) {
			gc.UserLogin = systemGitHubLogin
			gc.UserAvatar = systemGitHubAvatar
		}

		allComments = append(allComments, gc)
	}

	return allComments, nil
}

// PostCommentToGitHub posts a comment to all linked GitHub issues for a story.
// If the user has a linked GitHub account with a stored token, the comment is
// posted as the user directly. Otherwise it falls back to the installation bot.
func (s *Service) PostCommentToGitHub(ctx context.Context, workspaceID, storyID, userID uuid.UUID, localCommentID *uuid.UUID, authorName, body string) error {
	if !s.canUseAppAPI() {
		return errors.New("github app api is not configured")
	}

	issues, err := s.repo.GetStoryLinkedIssues(ctx, workspaceID, storyID)
	if err != nil {
		return fmt.Errorf("failed to get linked issues: %w", err)
	}
	if len(issues) == 0 {
		return errors.New("no linked github issues found for this story")
	}

	// Try to use the user's own GitHub token so the comment appears as them.
	userToken, tokenErr := s.repo.GetUserGitHubToken(ctx, userID)
	useUserToken := tokenErr == nil && strings.TrimSpace(userToken) != ""
	userClient := githubsdk.NewClient(s.httpClient).WithAuthToken(userToken)

	commentMarkerID := uuid.New()
	if localCommentID != nil {
		commentMarkerID = *localCommentID
	}
	userCommentBody := buildFortyOneUserCommentBody(body, commentMarkerID)
	fallbackCommentBody := buildFortyOneBotCommentBody(authorName, body, commentMarkerID)

	var lastErr error
	for _, issue := range issues {
		if useUserToken {
			comment, _, err := userClient.Issues.CreateComment(ctx, issue.OwnerLogin, issue.RepositorySlug, issue.GitHubNumber, &githubsdk.IssueComment{
				Body: &userCommentBody,
			})
			if err == nil {
				if recordErr := s.repo.RecordOutboundGitHubComment(ctx, workspaceID, storyID, issue.RepositoryID, comment.GetID(), localCommentID, userID); recordErr != nil {
					s.log.Warn(ctx, "failed to record outbound github user comment", "error", recordErr, "github_comment_id", comment.GetID())
				}
				continue
			}

			// If the linked user token is stale/revoked or lacks scope, transparently
			// fall back to installation auth so comment posting still succeeds.
			if !isGitHubAuthOrPermissionError(err) {
				lastErr = err
				continue
			}

			s.log.Warn(ctx, "github user token failed for comment; falling back to app installation",
				"issue_number", issue.GitHubNumber,
				"owner", issue.OwnerLogin,
				"repo", issue.RepositorySlug,
				"error", err,
			)
			// Stop retrying the same bad user token for remaining linked issues.
			useUserToken = false
		}

		client, clientErr := s.newInstallationClient(ctx, issue.GitHubInstallationID)
		if clientErr != nil {
			lastErr = clientErr
			continue
		}
		comment, _, err := client.Issues.CreateComment(ctx, issue.OwnerLogin, issue.RepositorySlug, issue.GitHubNumber, &githubsdk.IssueComment{
			Body: &fallbackCommentBody,
		})
		if err != nil {
			lastErr = err
			continue
		}
		if recordErr := s.repo.RecordOutboundGitHubComment(ctx, workspaceID, storyID, issue.RepositoryID, comment.GetID(), localCommentID, userID); recordErr != nil {
			s.log.Warn(ctx, "failed to record outbound github app comment", "error", recordErr, "github_comment_id", comment.GetID())
		}
	}
	return lastErr
}

func isGitHubAuthOrPermissionError(err error) bool {
	if err == nil {
		return false
	}
	var ghErr *githubsdk.ErrorResponse
	if errors.As(err, &ghErr) && ghErr.Response != nil {
		return ghErr.Response.StatusCode == http.StatusUnauthorized || ghErr.Response.StatusCode == http.StatusForbidden
	}
	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, "bad credentials") ||
		strings.Contains(errText, "requires authentication") ||
		strings.Contains(errText, "resource not accessible by personal access token")
}

func isFortyOneBotAuthorLogin(login, configuredSlug string) bool {
	normalizedLogin := strings.ToLower(strings.TrimSpace(login))
	if !strings.HasSuffix(normalizedLogin, "[bot]") {
		return false
	}

	configured := strings.ToLower(strings.TrimSpace(configuredSlug))
	if configured != "" && normalizedLogin == configured+"[bot]" {
		return true
	}

	// Keep this strict enough to avoid relabeling unrelated bots.
	return strings.HasPrefix(normalizedLogin, "fortyone-")
}

func isFortyOneSystemLinkedTaskComment(body string) bool {
	return strings.HasPrefix(strings.TrimSpace(body), "Linked to FortyOne task [")
}

func isFortyOneAuthoredCommentBody(body string) bool {
	trimmed := strings.TrimSpace(body)
	return fortyOneCommentPattern.MatchString(trimmed) || fortyOneCommentMarkerPattern.MatchString(trimmed)
}

func buildFortyOneUserCommentBody(body string, commentID uuid.UUID) string {
	return fmt.Sprintf("%s\n\n<!-- fortyone:comment:%s -->", strings.TrimSpace(body), commentID.String())
}

func buildFortyOneBotCommentBody(authorName, body string, commentID uuid.UUID) string {
	return fmt.Sprintf("**%s** commented via FortyOne:\n\n%s", authorName, buildFortyOneUserCommentBody(body, commentID))
}

func stripFortyOneCommentMarker(body string) string {
	return strings.TrimSpace(fortyOneCommentMarkerPattern.ReplaceAllString(body, ""))
}

func (s *Service) SyncStoryFromFortyOne(ctx context.Context, input CoreStorySyncInput) error {
	if !s.canUseAppAPI() {
		s.log.Warn(
			ctx,
			"skipping github story sync because github app api is not configured",
			append([]any{"story_id", input.StoryID}, s.appAPIConfigDiagnostics()...)...,
		)
		return nil
	}

	link, err := s.repo.FindBidirectionalIssueSyncLinkByTeamID(ctx, input.WorkspaceID, input.TeamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.log.Info(ctx, "skipping github story sync because team has no bidirectional github issue link", "story_id", input.StoryID, "team_id", input.TeamID)
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
	existingLink, err := s.repo.FindIssueStoryLinkByStoryID(ctx, input.WorkspaceID, input.StoryID, link.RepositoryID)
	switch {
	case err == nil:
		s.log.Info(ctx, "syncing story state to github issue", "story_id", input.StoryID, "github_issue_number", existingLink.GitHubNumber, "desired_state", desiredState)
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

	s.log.Info(ctx, "creating github issue for story during bidirectional sync", "story_id", input.StoryID, "desired_state", desiredState)
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
	if !s.canVerifyWebhooks() {
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
	case "pull_request_review":
		return s.handlePullRequestReviewEvent(ctx, repository, payload)
	case "issue_comment":
		return s.handleIssueCommentEvent(ctx, repository, payload)
	case "check_run":
		return s.handleCheckRunEvent(ctx, repository, payload)
	case "create":
		return s.handleCreateEvent(ctx, repository, payload)
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
	// Ignore issue events generated by our own GitHub App actor to avoid
	// status bounce-back loops for outbound FortyOne -> GitHub sync.
	if s.isAppSender(ctx, repository, payload.Sender.Login) {
		return nil
	}

	link, err := s.repo.FindIssueSyncLinkByRepositoryID(ctx, repository.ID)
	if err != nil {
		return nil
	}

	issueDescription := githubString(payload.Issue.Body)
	var story stories.CoreSingleStory
	var getErr error
	var createErr error
	_, storyID, err := s.repo.FindStoryLink(ctx, repository.ID, "issue", payload.Issue.ID, nil)
	switch {
	case err == nil:
		story, getErr = s.stories.Get(ctx, storyID, repository.WorkspaceID)
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

		story, createErr = s.stories.CreateExternal(ctx, s.cfg.GitHubUserID, stories.CoreNewStory{
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
	if payload.Action == "opened" {
		if err := s.ensureIssueImportComment(ctx, repository, payload.Issue.Number, story); err != nil {
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

	eventKey, shouldMoveStory := pullRequestWorkflowEvent(payload.Action, payload.PullRequest.Draft, payload.PullRequest.Merged)

	for _, story := range stories {
		prLinkCreated, err := s.upsertStoryLink(ctx, repository.WorkspaceID, story.StoryID, repository.ID, "pull_request", payload.PullRequest.ID, payload.PullRequest.Number, nil, payload.PullRequest.HTMLURL, payload.PullRequest.Title, payload.PullRequest.State, payload.PullRequest)
		if err != nil {
			return err
		}
		if err := s.repo.EnsureStoryLink(ctx, story.StoryID, githubPullRequestStoryLinkTitle(payload.PullRequest.Number), payload.PullRequest.HTMLURL); err != nil {
			return err
		}
		if prLinkCreated {
			if err := s.recordLinkActivity(ctx, repository.WorkspaceID, story.StoryID, "github_pull_request", fmt.Sprintf("PR #%d", payload.PullRequest.Number), payload.PullRequest.HTMLURL); err != nil {
				return err
			}
			if err := s.ensurePRLinkedComment(ctx, repository, payload.PullRequest.Number, story); err != nil {
				return err
			}
			s.autoPopulatePRBody(ctx, repository, payload, story)
		}
		// Sync assignee from PR to story
		s.syncAssigneeFromGitHub(ctx, repository, story, payload.PullRequest.Assignee)
		// Sync labels from PR to story
		s.syncLabelsFromGitHub(ctx, repository, story, payload.PullRequest.Labels)

		branchRef := payload.PullRequest.Head.Ref
		if branchRef != "" {
			branchLinkCreated, err := s.upsertStoryLink(ctx, repository.WorkspaceID, story.StoryID, repository.ID, "branch", 0, 0, &branchRef, payload.PullRequest.Head.HTMLURL, payload.PullRequest.Head.Ref, payload.PullRequest.State, payload.PullRequest.Head)
			if err != nil {
				return err
			}
			if err := s.repo.EnsureStoryLink(ctx, story.StoryID, githubBranchStoryLinkTitle(branchRef), payload.PullRequest.Head.HTMLURL); err != nil {
				return err
			}
			if branchLinkCreated {
				if err := s.recordLinkActivity(ctx, repository.WorkspaceID, story.StoryID, "github_branch", fmt.Sprintf("branch %s", branchRef), payload.PullRequest.Head.HTMLURL); err != nil {
					return err
				}
			}
		}
		if shouldMoveStory {
			if err := s.moveStoryByRule(ctx, repository.WorkspaceID, story.TeamID, story.StoryID, eventKey, &payload.PullRequest.Base.Ref); err != nil {
				return err
			}
		}
	}
	return nil
}

func pullRequestWorkflowEvent(action string, draft bool, merged bool) (string, bool) {
	switch action {
	case "opened":
		if draft {
			return EventDraftPROpen, true
		}
		return EventPROpen, true
	case "ready_for_review":
		return EventPRReadyForMerge, true
	case "closed":
		if merged {
			return EventPRMerge, true
		}
		return "", false
	default:
		return "", false
	}
}

func (s *Service) handleCreateEvent(ctx context.Context, repository githubrepository.RepoByExternalRow, payload webhookEnvelope) error {
	if s.stories == nil {
		return errors.New("stories service is not configured")
	}
	if payload.RefType != "branch" || strings.TrimSpace(payload.Ref) == "" {
		return nil
	}

	refs := extractStoryRefs(payload.Ref)
	stories, err := s.repo.ResolveStoriesByRefs(ctx, repository.WorkspaceID, refs)
	if err != nil || len(stories) == 0 {
		return err
	}

	branchURL := payload.Repository.HTMLURL + "/tree/" + payload.Ref
	for _, story := range stories {
		branchRef := payload.Ref
		branchLinkCreated, err := s.upsertStoryLink(ctx, repository.WorkspaceID, story.StoryID, repository.ID, "branch", 0, 0, &branchRef, branchURL, payload.Ref, "active", map[string]any{"ref": payload.Ref, "ref_type": payload.RefType})
		if err != nil {
			return err
		}
		if err := s.repo.EnsureStoryLink(ctx, story.StoryID, githubBranchStoryLinkTitle(branchRef), branchURL); err != nil {
			return err
		}
		if branchLinkCreated {
			if err := s.recordLinkActivity(ctx, repository.WorkspaceID, story.StoryID, "github_branch", fmt.Sprintf("branch %s", branchRef), branchURL); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) handlePushEvent(ctx context.Context, repository githubrepository.RepoByExternalRow, payload webhookEnvelope) error {
	if s.stories == nil {
		return errors.New("stories service is not configured")
	}

	settings, err := s.repo.GetWorkspaceSettings(ctx, repository.WorkspaceID)
	if err != nil {
		return err
	}

	refs := extractStoryRefs(payload.Ref)
	if settings.LinkCommitsByMagicWords {
		for _, commit := range payload.Commits {
			refs = append(refs, extractStoryRefs(commit.Message)...)
		}
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
			if err := s.repo.EnsureStoryLink(ctx, story.StoryID, githubBranchStoryLinkTitle(branch), payload.Repository.HTMLURL+"/tree/"+branch); err != nil {
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
		// Auto-close from commit keywords is configurable at workspace level.
		if settings.CloseOnCommitKeywords && hasClosingKeyword(payload.Commits, story.TeamCode, story.SequenceID) {
			if err := s.moveStoryByRule(ctx, repository.WorkspaceID, story.TeamID, story.StoryID, EventCommitClose, nil); err != nil {
				return err
			}
		}
	}
	return nil
}

var closingKeywordPattern = regexp.MustCompile(`(?i)\b(fix|fixes|fixed|close|closes|closed|resolve|resolves|resolved)\b\s+`)

func hasClosingKeyword(commits []struct {
	ID      string `json:"id"`
	Message string `json:"message"`
	URL     string `json:"url"`
}, teamCode string, sequenceID int) bool {
	target := strings.ToUpper(fmt.Sprintf("%s-%d", teamCode, sequenceID))
	targetAlt := strings.ToUpper(fmt.Sprintf("%s %d", teamCode, sequenceID))
	for _, commit := range commits {
		upper := strings.ToUpper(commit.Message)
		if !closingKeywordPattern.MatchString(commit.Message) {
			continue
		}
		if strings.Contains(upper, target) || strings.Contains(upper, targetAlt) {
			return true
		}
	}
	return false
}

// ==================== pull_request_review ====================

func (s *Service) handlePullRequestReviewEvent(ctx context.Context, repository githubrepository.RepoByExternalRow, payload webhookEnvelope) error {
	if s.stories == nil {
		return errors.New("stories service is not configured")
	}
	if payload.Action != "submitted" && payload.Action != "dismissed" {
		return nil
	}

	refs := extractStoryRefs(payload.PullRequest.Title, payload.PullRequest.Body, payload.PullRequest.Head.Ref)
	matches, err := s.repo.ResolveStoriesByRefs(ctx, repository.WorkspaceID, refs)
	if err != nil || len(matches) == 0 {
		return err
	}

	reviewState := normalizeReviewState(payload.Review.State)
	for _, story := range matches {
		approved, changesRequested := countReviews(ctx, s, repository, payload.PullRequest.Number)
		summaryState := summarizeReviewState(approved, changesRequested)
		if err := s.repo.UpdateStoryLinkReviewState(ctx, story.StoryID, repository.ID, payload.PullRequest.ID, summaryState, approved, changesRequested); err != nil {
			s.log.Warn(ctx, "failed to update review state on story link", "error", err)
		}

		if err := s.stories.RecordActivity(ctx, stories.CoreActivity{
			StoryID:      story.StoryID,
			Type:         "link",
			Field:        "github_review",
			CurrentValue: fmt.Sprintf("PR #%d %s by %s", payload.PullRequest.Number, reviewState, payload.Review.User.Login),
			NewValue:     payload.Review.HTMLURL,
			UserID:       s.resolveActorID(ctx, payload.Review.User.ID),
			WorkspaceID:  repository.WorkspaceID,
		}); err != nil {
			return err
		}

		if err := s.moveStoryByRule(ctx, repository.WorkspaceID, story.TeamID, story.StoryID, EventPRReviewActivity, &payload.PullRequest.Base.Ref); err != nil {
			return err
		}
	}
	return nil
}

func normalizeReviewState(state string) string {
	switch strings.ToLower(state) {
	case "approved":
		return "approved"
	case "changes_requested":
		return "changes_requested"
	case "dismissed":
		return "dismissed"
	case "commented":
		return "commented"
	default:
		return state
	}
}

func countReviews(ctx context.Context, s *Service, repository githubrepository.RepoByExternalRow, prNumber int) (approved, changesRequested int) {
	client, err := s.newInstallationClient(ctx, repository.GitHubInstallationID)
	if err != nil {
		return 0, 0
	}
	reviews, _, err := client.PullRequests.ListReviews(ctx, repository.OwnerLogin, repository.RepositorySlug, prNumber, &githubsdk.ListOptions{PerPage: 100})
	if err != nil {
		return 0, 0
	}
	latestByUser := map[int64]string{}
	for _, r := range reviews {
		if r.GetUser() != nil {
			latestByUser[r.GetUser().GetID()] = r.GetState()
		}
	}
	for _, state := range latestByUser {
		switch strings.ToUpper(state) {
		case "APPROVED":
			approved++
		case "CHANGES_REQUESTED":
			changesRequested++
		}
	}
	return approved, changesRequested
}

func summarizeReviewState(approved, changesRequested int) string {
	if changesRequested > 0 {
		return "changes_requested"
	}
	if approved > 0 {
		return "approved"
	}
	return "pending"
}

// ==================== issue_comment ====================

func (s *Service) handleIssueCommentEvent(ctx context.Context, repository githubrepository.RepoByExternalRow, payload webhookEnvelope) error {
	if s.stories == nil {
		return errors.New("stories service is not configured")
	}
	if payload.Action != "created" {
		return nil
	}
	// Ignore comments from our own app to prevent loops.
	if payload.Comment.User.ID == 0 ||
		isFortyOneAuthoredCommentBody(payload.Comment.Body) ||
		isFortyOneSystemLinkedTaskComment(payload.Comment.Body) ||
		s.isAppComment(ctx, repository, payload.Comment.User.ID) {
		return nil
	}
	if isOutbound, err := s.repo.IsOutboundGitHubComment(ctx, repository.ID, payload.Comment.ID); err != nil {
		return err
	} else if isOutbound {
		return nil
	}

	_, storyID, err := s.repo.FindStoryLink(ctx, repository.ID, "issue", payload.Issue.ID, nil)
	if err != nil {
		return nil // Issue not linked to any story
	}

	actorID := s.resolveActorID(ctx, payload.Comment.User.ID)
	reserved, err := s.repo.ReserveInboundGitHubComment(ctx, repository.WorkspaceID, storyID, repository.ID, payload.Comment.ID, actorID)
	if err != nil {
		return err
	}
	if !reserved {
		return nil
	}

	commentBody := fmt.Sprintf("**%s** commented on GitHub issue #%d:\n\n%s", payload.Comment.User.Login, payload.Issue.Number, payload.Comment.Body)
	comment, err := s.stories.CreateCommentExternal(ctx, actorID, repository.WorkspaceID, stories.CoreNewComment{
		StoryID: storyID,
		UserID:  actorID,
		Comment: commentBody,
	})
	if err != nil {
		_ = s.repo.DeleteGitHubCommentLink(ctx, repository.ID, payload.Comment.ID)
		return err
	}
	return s.repo.CompleteInboundGitHubComment(ctx, repository.ID, payload.Comment.ID, comment.ID)
}

func (s *Service) isAppComment(ctx context.Context, repository githubrepository.RepoByExternalRow, userID int64) bool {
	client, err := s.newInstallationClient(ctx, repository.GitHubInstallationID)
	if err != nil {
		return false
	}
	app, _, err := client.Apps.Get(ctx, "")
	if err != nil || app.GetOwner() == nil {
		return false
	}
	return app.GetOwner().GetID() == userID
}

func (s *Service) isAppSender(ctx context.Context, repository githubrepository.RepoByExternalRow, senderLogin string) bool {
	login := strings.ToLower(strings.TrimSpace(senderLogin))
	if login == "" {
		return false
	}

	client, err := s.newInstallationClient(ctx, repository.GitHubInstallationID)
	if err != nil {
		return false
	}
	app, _, err := client.Apps.Get(ctx, "")
	if err != nil {
		return false
	}

	appSlug := strings.ToLower(strings.TrimSpace(app.GetSlug()))
	if appSlug == "" {
		appSlug = strings.ToLower(strings.TrimSpace(s.cfg.AppSlug))
	}
	if appSlug == "" {
		return false
	}

	return login == appSlug+"[bot]"
}

// ==================== outbound comment sync (bidirectional) ====================

// SyncCommentToGitHub posts a FortyOne comment to the linked GitHub issue
// when the issue sync link is configured as bidirectional.
func (s *Service) SyncCommentToGitHub(ctx context.Context, workspaceID, storyID, teamID, localCommentID uuid.UUID, authorName, content string) error {
	link, err := s.repo.FindBidirectionalIssueSyncLinkByTeamID(ctx, workspaceID, teamID)
	if err != nil {
		return nil // No bidirectional link for this team — nothing to do
	}

	issueLink, err := s.repo.FindIssueStoryLinkByStoryID(ctx, workspaceID, storyID, link.RepositoryID)
	if err != nil {
		return nil // Story has no linked GitHub issue
	}

	client, err := s.newInstallationClient(ctx, link.GitHubInstallationID)
	if err != nil {
		return fmt.Errorf("failed to create installation client: %w", err)
	}

	body := buildFortyOneBotCommentBody(authorName, content, localCommentID)
	comment, _, err := client.Issues.CreateComment(ctx, link.OwnerLogin, link.RepositorySlug, issueLink.GitHubNumber, &githubsdk.IssueComment{
		Body: &body,
	})
	if err != nil {
		return fmt.Errorf("failed to create github issue comment: %w", err)
	}
	return s.repo.RecordOutboundGitHubComment(ctx, workspaceID, storyID, link.RepositoryID, comment.GetID(), &localCommentID, s.cfg.GitHubUserID)
}

// ==================== check_run ====================

func (s *Service) handleCheckRunEvent(ctx context.Context, repository githubrepository.RepoByExternalRow, payload webhookEnvelope) error {
	if payload.Action != "completed" {
		return nil
	}
	for _, pr := range payload.CheckRun.PullRequests {
		matches, err := s.repo.FindStoryLinksByPRNumber(ctx, repository.ID, pr.Number)
		if err != nil || len(matches) == 0 {
			continue
		}
		for _, story := range matches {
			if err := s.repo.UpdateStoryLinkCheckState(ctx, story.StoryID, repository.ID, pr.ID, payload.CheckRun.Conclusion); err != nil {
				s.log.Warn(ctx, "failed to update check state on story link", "error", err)
			}
		}
	}
	return nil
}

// ==================== auto-populate PR body ====================

func (s *Service) autoPopulatePRBody(ctx context.Context, repository githubrepository.RepoByExternalRow, payload webhookEnvelope, story githubrepository.StoryMatch) {
	settings, err := s.repo.GetWorkspaceSettingsByWorkspaceID(ctx, repository.WorkspaceID)
	if err != nil || !settings.AutoPopulatePRBody {
		return
	}
	storyURL, err := storyURLFromWebsite(s.cfg.WebsiteURL, repository.WorkspaceSlug, story.StoryID, story.Title)
	if err != nil {
		return
	}
	taskKey := fmt.Sprintf("%s-%d", story.TeamCode, story.SequenceID)
	marker := fmt.Sprintf("<!-- fortyone:%s -->", story.StoryID.String())
	if strings.Contains(payload.PullRequest.Body, marker) {
		return
	}
	footer := fmt.Sprintf("\n\n%s\n---\nLinked to [%s](%s)", marker, taskKey, storyURL)
	newBody := payload.PullRequest.Body + footer
	client, clientErr := s.newInstallationClient(ctx, repository.GitHubInstallationID)
	if clientErr != nil {
		return
	}
	_, _, _ = client.PullRequests.Edit(ctx, repository.OwnerLogin, repository.RepositorySlug, payload.PullRequest.Number, &githubsdk.PullRequest{
		Body: &newBody,
	})
}

// ==================== assignee sync ====================

func (s *Service) syncAssigneeFromGitHub(ctx context.Context, repository githubrepository.RepoByExternalRow, story githubrepository.StoryMatch, assignee *struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
}) {
	settings, err := s.repo.GetWorkspaceSettingsByWorkspaceID(ctx, repository.WorkspaceID)
	if err != nil || !settings.SyncAssignees || assignee == nil || assignee.ID == 0 {
		return
	}
	userID, err := s.repo.ResolveUserByGitHubID(ctx, assignee.ID)
	if err != nil {
		return
	}
	fullStory, err := s.stories.Get(ctx, story.StoryID, repository.WorkspaceID)
	if err != nil || (fullStory.Assignee != nil && *fullStory.Assignee == userID) {
		return
	}
	_ = s.stories.UpdateExternal(ctx, s.cfg.GitHubUserID, story.StoryID, repository.WorkspaceID, map[string]any{
		"assignee_id": userID,
	})
}

// ==================== label sync ====================

func (s *Service) syncLabelsFromGitHub(ctx context.Context, repository githubrepository.RepoByExternalRow, story githubrepository.StoryMatch, labels []struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}) {
	settings, err := s.repo.GetWorkspaceSettingsByWorkspaceID(ctx, repository.WorkspaceID)
	if err != nil || !settings.SyncLabels || len(labels) == 0 {
		return
	}
	labelNames := make([]string, 0, len(labels))
	for _, l := range labels {
		labelNames = append(labelNames, l.Name)
	}
	resolvedIDs, err := s.repo.ResolveOrCreateLabelsByName(ctx, repository.WorkspaceID, story.TeamID, labelNames)
	if err != nil || len(resolvedIDs) == 0 {
		return
	}
	_ = s.stories.UpdateExternal(ctx, s.cfg.GitHubUserID, story.StoryID, repository.WorkspaceID, map[string]any{
		"labels": resolvedIDs,
	})
}

// ==================== user resolution ====================

func (s *Service) resolveActorID(ctx context.Context, githubUserID int64) uuid.UUID {
	if githubUserID == 0 {
		return s.cfg.GitHubUserID
	}
	userID, err := s.repo.ResolveUserByGitHubID(ctx, githubUserID)
	if err != nil {
		return s.cfg.GitHubUserID
	}
	return userID
}

// ==================== GitHub user linking ====================

const userLinkStateTTL = 15 * time.Minute

func (s *Service) CreateUserLinkSession(ctx context.Context, userID uuid.UUID, returnTo string) (CoreCreateUserLinkSession, error) {
	state, err := s.createUserLinkState(userID, returnTo, time.Now().Add(userLinkStateTTL))
	if err != nil {
		return CoreCreateUserLinkSession{}, err
	}
	return CoreCreateUserLinkSession{State: state}, nil
}

func (s *Service) LinkGitHubUser(ctx context.Context, userID uuid.UUID, code, state string) error {
	if _, err := s.verifyUserLinkState(state, userID, time.Now()); err != nil {
		return err
	}
	token, err := s.exchangeOAuthCode(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to exchange github oauth code: %w", err)
	}
	ghClient := githubsdk.NewClient(s.httpClient).WithAuthToken(token)
	user, _, err := ghClient.Users.Get(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to get github user: %w", err)
	}
	return s.repo.LinkGitHubUser(ctx, userID, user.GetID(), user.GetLogin(), token)
}

func (s *Service) UnlinkGitHubUser(ctx context.Context, userID uuid.UUID) error {
	return s.repo.UnlinkGitHubUser(ctx, userID)
}

func (s *Service) exchangeOAuthCode(ctx context.Context, code string) (string, error) {
	values := url.Values{}
	values.Set("client_id", s.cfg.ClientID)
	values.Set("client_secret", s.cfg.ClientSecret)
	values.Set("code", code)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://github.com/login/oauth/access_token?"+values.Encode(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Error != "" {
		return "", fmt.Errorf("github oauth error: %s", result.Error)
	}
	return result.AccessToken, nil
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
		{EventKey: EventCommitClose, TargetStatusID: findCategory("completed"), IsActive: true},
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

func githubPullRequestStoryLinkTitle(prNumber int) *string {
	title := fmt.Sprintf("GitHub PR #%d", prNumber)
	return &title
}

func githubBranchStoryLinkTitle(branchName string) *string {
	title := fmt.Sprintf("GitHub branch %s", branchName)
	return &title
}

func (s *Service) appAPIConfigDiagnostics() []any {
	return []any{
		"app_id_configured", s.cfg.AppID != 0,
		"app_slug_configured", strings.TrimSpace(s.cfg.AppSlug) != "",
		"private_key_base64_present", strings.TrimSpace(s.cfg.PrivateKeyBase64) != "",
		"private_key_loaded", s.privateKey != nil,
		"private_key_load_error", s.privateKeyLoadError,
		"redirect_url_configured", strings.TrimSpace(s.cfg.RedirectURL) != "",
		"webhook_secret_configured", strings.TrimSpace(s.cfg.WebhookSecret) != "",
	}
}

func buildLinkedTaskComment(storyURL, teamCode string, sequenceID int) string {
	normalizedTeamCode := strings.ToUpper(strings.TrimSpace(teamCode))
	taskKey := fmt.Sprintf("#%d", sequenceID)
	if normalizedTeamCode != "" {
		taskKey = fmt.Sprintf("%s-%d", normalizedTeamCode, sequenceID)
	}
	return fmt.Sprintf(
		"Linked to FortyOne task [%s](%s).",
		taskKey,
		storyURL,
	)
}

func githubString(value string) *string {
	return &value
}

func storyCommentMarker(storyID uuid.UUID) string {
	return fmt.Sprintf("`%s`", storyID.String())
}

func storyURLFromWebsite(websiteURL, workspaceSlug string, storyID uuid.UUID, storyTitle string) (string, error) {
	baseURL, err := url.Parse(strings.TrimRight(websiteURL, "/"))
	if err != nil {
		return "", err
	}

	if workspaceSlug == "" {
		return "", errors.New("workspace slug is required")
	}

	baseURL.Path = path.Join("/", "story", storyID.String(), slugifyStoryTitle(storyTitle))

	host := baseURL.Hostname()
	if host == "" {
		return "", errors.New("website host is required")
	}

	if isLocalWebsiteHost(host) {
		baseURL.Path = path.Join("/", workspaceSlug, "story", storyID.String(), slugifyStoryTitle(storyTitle))
		return baseURL.String(), nil
	}

	if !strings.HasPrefix(host, workspaceSlug+".") {
		if port := baseURL.Port(); port != "" {
			baseURL.Host = fmt.Sprintf("%s.%s:%s", workspaceSlug, host, port)
		} else {
			baseURL.Host = fmt.Sprintf("%s.%s", workspaceSlug, host)
		}
	}

	return baseURL.String(), nil
}

func isLocalWebsiteHost(host string) bool {
	return strings.EqualFold(host, "localhost") || strings.EqualFold(host, "0.0.0.0") || net.ParseIP(host) != nil
}

func slugifyStoryTitle(title string) string {
	normalized := norm.NFD.String(strings.TrimSpace(strings.ToLower(title)))
	var b strings.Builder
	lastHyphen := false

	for _, r := range normalized {
		switch {
		case unicode.Is(unicode.Mn, r):
			continue
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastHyphen = false
		case r == '&':
			if b.Len() > 0 && !lastHyphen {
				b.WriteByte('-')
			}
			b.WriteString("and")
			lastHyphen = false
		default:
			if b.Len() == 0 || lastHyphen {
				continue
			}
			b.WriteByte('-')
			lastHyphen = true
		}
	}

	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		return "story"
	}
	return slug
}

func loadPrivateKey(privateKeyBase64 string) (*rsa.PrivateKey, error) {
	pemBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(privateKeyBase64))
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode private key: %w", err)
	}
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("invalid github private key: no PEM block found after base64 decoding")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("github private key is not RSA")
	}
	return rsaKey, nil
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
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

func (s *Service) newAppClient() (*githubsdk.Client, error) {
	appJWT, err := s.createAppJWT()
	if err != nil {
		return nil, err
	}
	return githubsdk.NewClient(s.httpClient).WithAuthToken(appJWT), nil
}

func (s *Service) newInstallationClient(ctx context.Context, installationID int64) (*githubsdk.Client, error) {
	token, err := s.getInstallationToken(ctx, installationID)
	if err != nil {
		return nil, err
	}
	return githubsdk.NewClient(s.httpClient).WithAuthToken(token), nil
}

func (s *Service) getInstallationToken(ctx context.Context, installationID int64) (string, error) {
	client, err := s.newAppClient()
	if err != nil {
		return "", err
	}
	token, _, err := client.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		return "", err
	}
	return token.GetToken(), nil
}

func (s *Service) getInstallation(ctx context.Context, installationID int64) (githubrepository.GithubInstallationPayload, error) {
	client, err := s.newAppClient()
	if err != nil {
		return githubrepository.GithubInstallationPayload{}, err
	}
	installation, _, err := client.Apps.GetInstallation(ctx, installationID)
	if err != nil {
		return githubrepository.GithubInstallationPayload{}, err
	}
	return toInstallationPayload(installation), nil
}

func (s *Service) listInstallationRepositories(ctx context.Context, installationID int64) ([]githubrepository.GithubRepositoryPayload, error) {
	client, err := s.newInstallationClient(ctx, installationID)
	if err != nil {
		return nil, err
	}

	options := &githubsdk.ListOptions{PerPage: 100}
	items := make([]githubrepository.GithubRepositoryPayload, 0)
	for {
		repos, response, err := client.Apps.ListRepos(ctx, options)
		if err != nil {
			return nil, err
		}
		for _, repository := range repos.Repositories {
			if repository == nil {
				continue
			}
			items = append(items, toRepositoryPayload(repository))
		}
		if response == nil || response.NextPage == 0 {
			break
		}
		options.Page = response.NextPage
	}
	return items, nil
}

type githubIssuePayload struct {
	ID      int64  `json:"id"`
	Number  int    `json:"number"`
	HTMLURL string `json:"html_url"`
	Title   string `json:"title"`
	State   string `json:"state"`
}

func toInstallationPayload(installation *githubsdk.Installation) githubrepository.GithubInstallationPayload {
	if installation == nil {
		return githubrepository.GithubInstallationPayload{}
	}

	permissions := map[string]string{}
	if installation.Permissions != nil {
		bytes, err := json.Marshal(installation.Permissions)
		if err == nil {
			_ = json.Unmarshal(bytes, &permissions)
		}
	}

	var avatarURL *string
	account := installation.GetAccount()
	accountID := int64(0)
	accountLogin := ""
	accountType := ""
	if account != nil && account.AvatarURL != nil {
		avatarURL = account.AvatarURL
	}
	if account != nil {
		accountID = account.GetID()
		accountLogin = account.GetLogin()
		accountType = account.GetType()
	}

	return githubrepository.GithubInstallationPayload{
		ID: installation.GetID(),
		Account: githubrepository.GithubInstallationAccountPayload{
			ID:        accountID,
			Login:     accountLogin,
			Type:      accountType,
			AvatarURL: avatarURL,
		},
		RepositorySelection: installation.GetRepositorySelection(),
		Permissions:         permissions,
		Events:              installation.Events,
	}
}

func toRepositoryPayload(repository *githubsdk.Repository) githubrepository.GithubRepositoryPayload {
	if repository == nil {
		return githubrepository.GithubRepositoryPayload{}
	}

	var description *string
	if repository.Description != nil {
		description = repository.Description
	}

	owner := repository.GetOwner()
	ownerID := int64(0)
	ownerLogin := ""
	if owner != nil {
		ownerID = owner.GetID()
		ownerLogin = owner.GetLogin()
	}

	return githubrepository.GithubRepositoryPayload{
		ID:            repository.GetID(),
		Name:          repository.GetName(),
		FullName:      repository.GetFullName(),
		Description:   description,
		HTMLURL:       repository.GetHTMLURL(),
		CloneURL:      repository.GetCloneURL(),
		SSHURL:        repository.GetSSHURL(),
		DefaultBranch: repository.GetDefaultBranch(),
		Private:       repository.GetPrivate(),
		Archived:      repository.GetArchived(),
		Disabled:      repository.GetDisabled(),
		Owner: githubrepository.GithubRepositoryOwnerPayload{
			ID:    ownerID,
			Login: ownerLogin,
		},
	}
}

func toIssuePayload(issue *githubsdk.Issue) githubIssuePayload {
	if issue == nil {
		return githubIssuePayload{}
	}
	return githubIssuePayload{
		ID:      issue.GetID(),
		Number:  issue.GetNumber(),
		HTMLURL: issue.GetHTMLURL(),
		Title:   issue.GetTitle(),
		State:   issue.GetState(),
	}
}

func (s *Service) createIssue(ctx context.Context, installationID int64, owner, repository, title, body string) (githubIssuePayload, error) {
	client, err := s.newInstallationClient(ctx, installationID)
	if err != nil {
		return githubIssuePayload{}, err
	}
	request := &githubsdk.IssueRequest{
		Title: &title,
		Body:  &body,
	}
	issue, _, err := client.Issues.Create(ctx, owner, repository, request)
	if err != nil {
		return githubIssuePayload{}, err
	}
	return toIssuePayload(issue), nil
}

func (s *Service) createIssueComment(ctx context.Context, installationID int64, owner, repository string, number int, body string) error {
	client, err := s.newInstallationClient(ctx, installationID)
	if err != nil {
		return err
	}
	comment := &githubsdk.IssueComment{Body: &body}
	_, _, err = client.Issues.CreateComment(ctx, owner, repository, number, comment)
	return err
}

func (s *Service) issueHasStoryComment(ctx context.Context, installationID int64, owner, repository string, number int, storyID uuid.UUID, storyURL string) (bool, error) {
	client, err := s.newInstallationClient(ctx, installationID)
	if err != nil {
		return false, err
	}

	options := &githubsdk.IssueListCommentsOptions{
		ListOptions: githubsdk.ListOptions{PerPage: 100},
	}

	for {
		comments, response, err := client.Issues.ListComments(ctx, owner, repository, number, options)
		if err != nil {
			return false, err
		}

		for _, comment := range comments {
			if comment == nil || comment.Body == nil {
				continue
			}
			body := *comment.Body
			if strings.Contains(body, storyURL) || strings.Contains(body, storyCommentMarker(storyID)) {
				return true, nil
			}
		}

		if response == nil || response.NextPage == 0 {
			return false, nil
		}
		options.Page = response.NextPage
	}
}

func (s *Service) updateIssue(ctx context.Context, installationID int64, owner, repository string, number int, title, body, state string) (githubIssuePayload, error) {
	client, err := s.newInstallationClient(ctx, installationID)
	if err != nil {
		return githubIssuePayload{}, err
	}
	request := &githubsdk.IssueRequest{
		Title: &title,
		Body:  &body,
		State: &state,
	}
	issue, _, err := client.Issues.Edit(ctx, owner, repository, number, request)
	if err != nil {
		return githubIssuePayload{}, err
	}
	return toIssuePayload(issue), nil
}

func (s *Service) ensureIssueImportComment(ctx context.Context, repository githubrepository.RepoByExternalRow, issueNumber int, story stories.CoreSingleStory) error {
	storyURL, err := storyURLFromWebsite(s.cfg.WebsiteURL, repository.WorkspaceSlug, story.ID, story.Title)
	if err != nil {
		return err
	}

	// Some create paths may not hydrate team_code on the freshly returned story.
	// Re-load before constructing the task key to avoid malformed values like "-418".
	if strings.TrimSpace(story.TeamCode) == "" {
		loadedStory, loadErr := s.stories.Get(ctx, story.ID, repository.WorkspaceID)
		if loadErr == nil {
			story = loadedStory
		}
	}

	exists, err := s.issueHasStoryComment(
		ctx,
		repository.GitHubInstallationID,
		repository.OwnerLogin,
		repository.RepositorySlug,
		issueNumber,
		story.ID,
		storyURL,
	)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	commentBody := buildLinkedTaskComment(storyURL, story.TeamCode, story.SequenceID)
	return s.createIssueComment(
		ctx,
		repository.GitHubInstallationID,
		repository.OwnerLogin,
		repository.RepositorySlug,
		issueNumber,
		commentBody,
	)
}

func (s *Service) ensurePRLinkedComment(ctx context.Context, repository githubrepository.RepoByExternalRow, prNumber int, story githubrepository.StoryMatch) error {
	storyURL, err := storyURLFromWebsite(s.cfg.WebsiteURL, repository.WorkspaceSlug, story.StoryID, story.Title)
	if err != nil {
		return err
	}

	exists, err := s.issueHasStoryComment(
		ctx,
		repository.GitHubInstallationID,
		repository.OwnerLogin,
		repository.RepositorySlug,
		prNumber,
		story.StoryID,
		storyURL,
	)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	commentBody := buildLinkedTaskComment(storyURL, story.TeamCode, story.SequenceID)
	return s.createIssueComment(
		ctx,
		repository.GitHubInstallationID,
		repository.OwnerLogin,
		repository.RepositorySlug,
		prNumber,
		commentBody,
	)
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

func (s *Service) createUserLinkState(userID uuid.UUID, returnTo string, expiresAt time.Time) (string, error) {
	safeReturnTo, err := s.safeUserLinkReturnTo(returnTo)
	if err != nil {
		return "", err
	}
	return s.signState(map[string]string{
		"kind":      "github_user_link",
		"user_id":   userID.String(),
		"return_to": safeReturnTo,
		"expires":   strconv.FormatInt(expiresAt.Unix(), 10),
	})
}

func (s *Service) verifyUserLinkState(state string, userID uuid.UUID, now time.Time) (string, error) {
	values, err := s.verifyState(state)
	if err != nil {
		return "", err
	}
	if values["kind"] != "github_user_link" {
		return "", errors.New("invalid github user link state")
	}
	if values["user_id"] != userID.String() {
		return "", errors.New("github user link state does not match the authenticated user")
	}
	expires, err := strconv.ParseInt(values["expires"], 10, 64)
	if err != nil {
		return "", errors.New("invalid github user link state expiry")
	}
	if !now.Before(time.Unix(expires, 0)) {
		return "", errors.New("github user link state has expired")
	}
	return s.safeUserLinkReturnTo(values["return_to"])
}

func (s *Service) safeUserLinkReturnTo(returnTo string) (string, error) {
	if strings.TrimSpace(returnTo) == "" {
		return "", errors.New("return path is required")
	}
	parsed, err := url.Parse(returnTo)
	if err != nil {
		return "", err
	}
	if parsed.IsAbs() || parsed.Host != "" {
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return "", errors.New("return URL scheme is not allowed")
		}
		if !s.isAllowedUserLinkReturnHost(parsed.Hostname()) {
			return "", errors.New("return URL host is not allowed")
		}
		return parsed.String(), nil
	}
	if !strings.HasPrefix(parsed.Path, "/") || strings.HasPrefix(parsed.Path, "//") {
		return "", errors.New("return path must be a relative application path")
	}
	if parsed.Path == "" {
		parsed.Path = "/"
	}
	return parsed.RequestURI(), nil
}

func (s *Service) isAllowedUserLinkReturnHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return false
	}
	if isLocalWebsiteHost(host) {
		return true
	}
	website, err := url.Parse(s.cfg.WebsiteURL)
	if err != nil {
		return false
	}
	configuredHost := strings.ToLower(strings.TrimSpace(website.Hostname()))
	if host == configuredHost {
		return true
	}
	labels := strings.Split(configuredHost, ".")
	if len(labels) < 2 {
		return false
	}
	rootDomain := strings.Join(labels[len(labels)-2:], ".")
	return host == rootDomain || strings.HasSuffix(host, "."+rootDomain)
}

type webhookEnvelope struct {
	Action     string `json:"action"`
	RefType    string `json:"ref_type"`
	Repository struct {
		ID      int64  `json:"id"`
		HTMLURL string `json:"html_url"`
	} `json:"repository"`
	Installation struct {
		ID int64 `json:"id"`
	} `json:"installation"`
	Sender struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
	} `json:"sender"`
	Issue struct {
		ID      int64  `json:"id"`
		Number  int    `json:"number"`
		Title   string `json:"title"`
		Body    string `json:"body"`
		State   string `json:"state"`
		HTMLURL string `json:"html_url"`
		User    struct {
			ID    int64  `json:"id"`
			Login string `json:"login"`
		} `json:"user"`
		Assignee *struct {
			ID    int64  `json:"id"`
			Login string `json:"login"`
		} `json:"assignee"`
		Labels []struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"labels"`
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
		User struct {
			ID    int64  `json:"id"`
			Login string `json:"login"`
		} `json:"user"`
		Assignee *struct {
			ID    int64  `json:"id"`
			Login string `json:"login"`
		} `json:"assignee"`
		Labels []struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"labels"`
	} `json:"pull_request"`
	Review struct {
		ID    int64  `json:"id"`
		State string `json:"state"`
		Body  string `json:"body"`
		User  struct {
			ID    int64  `json:"id"`
			Login string `json:"login"`
		} `json:"user"`
		HTMLURL string `json:"html_url"`
	} `json:"review"`
	Comment struct {
		ID      int64  `json:"id"`
		Body    string `json:"body"`
		HTMLURL string `json:"html_url"`
		User    struct {
			ID    int64  `json:"id"`
			Login string `json:"login"`
		} `json:"user"`
	} `json:"comment"`
	CheckRun struct {
		ID           int64  `json:"id"`
		Name         string `json:"name"`
		Status       string `json:"status"`
		Conclusion   string `json:"conclusion"`
		HTMLURL      string `json:"html_url"`
		PullRequests []struct {
			ID     int64 `json:"id"`
			Number int   `json:"number"`
		} `json:"pull_requests"`
	} `json:"check_run"`
	Label struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"label"`
	Assignee *struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
	} `json:"assignee"`
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
