package github

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	githubshared "github.com/complexus-tech/projects-api/internal/modules/github/shared"
	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

var refPattern = regexp.MustCompile(`\b([A-Za-z]{2,}[ -]?\d+)\b`)
var githubAppSlugPattern = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]{0,98}[A-Za-z0-9])?$`)

const githubIssueLinkPrefix = "GitHub issue: "
const (
	githubSyncSourceFortyOne = "fortyone"
	githubSyncSourceGitHub   = "github"
)

// These aliases isolate pre-existing concrete module contracts at one
// compatibility seam while the service is decomposed by capability. New code
// should use caller-owned ports or the provider-neutral codehost contracts.
type (
	StoryGitHubLink                  = githubshared.StoryGitHubLink
	repositoryRecord                 = githubshared.RepositoryRecord
	storyMatch                       = githubshared.StoryMatch
	fortyOneUser                     = githubshared.FortyOneUser
	githubInstallationPayload        = githubshared.InstallationPayload
	githubInstallationAccountPayload = githubshared.InstallationAccountPayload
	githubRepositoryPayload          = githubshared.RepositoryPayload
	githubRepositoryOwnerPayload     = githubshared.RepositoryOwnerPayload
	integrationRequest               = IntegrationRequest
	upsertIntegrationRequestInput    = UpsertIntegrationRequestInput
	singleStory                      = storydomain.Story
	storyActivity                    = StoryActivity
	newStoryComment                  = NewStoryComment
)

const (
	providerGitHub         = "github"
	requestSourceTypeIssue = "issue"
)

var githubPriorityLabels = map[string]string{
	"blocker":         "Urgent",
	"critical":        "Urgent",
	"p0":              "Urgent",
	"priority p0":     "Urgent",
	"priority urgent": "Urgent",
	"urgent":          "Urgent",
	"high":            "High",
	"p1":              "High",
	"priority high":   "High",
	"priority p1":     "High",
	"medium":          "Medium",
	"p2":              "Medium",
	"priority medium": "Medium",
	"priority p2":     "Medium",
	"low":             "Low",
	"p3":              "Low",
	"priority low":    "Low",
	"priority p3":     "Low",
}

type Service struct {
	log                  *logger.Logger
	repo                 Repository
	stories              StoryService
	requests             RequestStore
	avatars              AvatarResolver
	httpClient           *http.Client
	cfg                  Config
	privateKey           *rsa.PrivateKey
	privateKeyLoadError  string
	workspaceRoles       WorkspaceRoleReader
	credentials          CredentialVault
	oauthStates          OAuthStateStore
	webhookGateway       *webhooks.Gateway
	webhookInbox         webhooks.Inbox
	webhookPayloads      WebhookPayloadOpener
	webhookInstallations WebhookInstallationRepository
	now                  func() time.Time
}

func New(log *logger.Logger, repo Repository, storyService StoryService, requestSink RequestStore, avatarResolver AvatarResolver, cfg Config) (*Service, error) {
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
		log:      log,
		repo:     repo,
		stories:  storyService,
		requests: requestSink,
		avatars:  avatarResolver,
		cfg:      cfg,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		privateKey:           privateKey,
		privateKeyLoadError:  errorString(err),
		workspaceRoles:       repo,
		credentials:          cfg.CredentialVault,
		oauthStates:          cfg.OAuthStateStore,
		webhookGateway:       cfg.WebhookGateway,
		webhookInbox:         cfg.WebhookInbox,
		webhookPayloads:      cfg.WebhookPayloads,
		webhookInstallations: repo,
		now:                  time.Now,
	}, nil
}

func (s *Service) requireWorkspaceAdmin(ctx context.Context, workspaceID, actorID uuid.UUID) error {
	role, err := s.workspaceRoles.GetWorkspaceRole(ctx, workspaceID, actorID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return authorization.ErrWorkspaceAdminRequired
		}
		return fmt.Errorf("authorize github workspace administration: %w", err)
	}
	return authorization.RequireWorkspaceAdmin(role)
}

func githubPriorityFromLabelNames(labels []string) string {
	priority, ok := githubPriorityUpdateFromLabelNames(labels)
	if !ok {
		return "No Priority"
	}
	return priority
}

func githubPriorityUpdateFromLabelNames(labels []string) (string, bool) {
	priority := "No Priority"
	found := false
	priorityRank := map[string]int{
		"No Priority": 0,
		"Low":         1,
		"Medium":      2,
		"High":        3,
		"Urgent":      4,
	}

	for _, label := range labels {
		candidate, ok := githubPriorityLabels[normalizeGitHubPriorityLabel(label)]
		if !ok {
			continue
		}
		found = true
		if priorityRank[candidate] > priorityRank[priority] {
			priority = candidate
		}
	}

	return priority, found
}

func normalizeGitHubPriorityLabel(label string) string {
	normalized := strings.ToLower(strings.TrimSpace(label))
	normalized = strings.TrimPrefix(normalized, "priority:")
	normalized = strings.TrimPrefix(normalized, "priority/")
	normalized = strings.TrimPrefix(normalized, "priority-")
	normalized = strings.TrimPrefix(normalized, "prio:")
	normalized = strings.TrimPrefix(normalized, "prio/")
	normalized = strings.TrimPrefix(normalized, "prio-")
	normalized = strings.NewReplacer("-", " ", "_", " ", "/", " ", ":", " ").Replace(normalized)
	return strings.Join(strings.Fields(normalized), " ")
}

func (s *Service) canInstall() bool {
	return s.cfg.AppID > 0 &&
		githubAppSlugPattern.MatchString(strings.TrimSpace(s.cfg.AppSlug)) &&
		validGitHubApplicationURL(s.cfg.RedirectURL) &&
		s.privateKey != nil
}

func (s *Service) canUseAppAPI() bool {
	return s.cfg.AppID > 0 && s.privateKey != nil
}

func validGitHubApplicationURL(value string) bool {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || !parsed.IsAbs() || parsed.Hostname() == "" || parsed.User != nil || parsed.Fragment != "" {
		return false
	}
	if parsed.Scheme == "https" {
		return true
	}
	return parsed.Scheme == "http" && isLocalWebsiteHost(parsed.Hostname())
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
		if strings.TrimSpace(s.cfg.WebhookSecret) != "" &&
			(s.webhookGateway == nil || s.webhookInbox == nil || s.webhookPayloads == nil) {
			return errors.New("github worker webhook runtime is not configured")
		}
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
