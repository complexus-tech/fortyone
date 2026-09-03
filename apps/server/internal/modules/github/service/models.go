package github

import (
	"context"
	"time"

	githubshared "github.com/complexus-tech/projects-api/internal/modules/github/shared"
	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/google/uuid"
)

const (
	SyncDirectionInboundOnly            = githubshared.SyncDirectionInboundOnly
	SyncDirectionBidirectional          = githubshared.SyncDirectionBidirectional
	BranchFormatUsernameIdentifierTitle = githubshared.BranchFormatUsernameIdentifierTitle
	BranchFormatIdentifierTitle         = githubshared.BranchFormatIdentifierTitle
	BranchFormatIdentifierSlashTitle    = githubshared.BranchFormatIdentifierSlashTitle
	EventDraftPROpen                    = githubshared.EventDraftPROpen
	EventPROpen                         = githubshared.EventPROpen
	EventPRReviewActivity               = githubshared.EventPRReviewActivity
	EventPRReadyForMerge                = githubshared.EventPRReadyForMerge
	EventPRMerge                        = githubshared.EventPRMerge
	EventIssueOpen                      = githubshared.EventIssueOpen
	EventIssueReopen                    = githubshared.EventIssueReopen
	EventIssueClose                     = githubshared.EventIssueClose
	EventCommitClose                    = githubshared.EventCommitClose
)

type CoreWorkspaceSettings = githubshared.CoreWorkspaceSettings
type CoreInstallation = githubshared.CoreInstallation
type CoreRepository = githubshared.CoreRepository
type CoreIssueSyncLink = githubshared.CoreIssueSyncLink
type CoreWorkflowRule = githubshared.CoreWorkflowRule
type CoreIntegration = githubshared.CoreIntegration
type CoreCreateInstallSession = githubshared.CoreCreateInstallSession
type CoreCreateUserLinkSession = githubshared.CoreCreateUserLinkSession
type CoreIssueSyncLinkInput = githubshared.CoreIssueSyncLinkInput
type CoreUpdateIssueSyncLinkInput = githubshared.CoreUpdateIssueSyncLinkInput
type CoreUpdateWorkspaceSettingsInput = githubshared.CoreUpdateWorkspaceSettingsInput
type CoreTeamGitHubSettings = githubshared.CoreTeamGitHubSettings
type CoreWorkflowRuleInput = githubshared.CoreWorkflowRuleInput
type CoreUpdateTeamGitHubSettings = githubshared.CoreUpdateTeamGitHubSettings

// StoryActivity and NewStoryComment are the caller-owned contracts used by
// GitHub workflows. Bootstrap adapters translate them into the stories
// module's use-case commands.
type StoryActivity struct {
	StoryID      uuid.UUID
	UserID       uuid.UUID
	Type         string
	Field        string
	CurrentValue string
	OldValue     any
	NewValue     any
	Reason       *string
	WorkspaceID  uuid.UUID
}

type NewStoryComment struct {
	StoryID  uuid.UUID
	Parent   *uuid.UUID
	UserID   uuid.UUID
	Comment  string
	Mentions []uuid.UUID
}

// IntegrationRequest is the GitHub-owned projection needed to accept or write
// back to a GitHub issue. It intentionally omits unrelated request fields.
type IntegrationRequest struct {
	WorkspaceID      uuid.UUID
	Provider         string
	SourceType       string
	SourceExternalID string
	SourceNumber     *int
	SourceURL        *string
	Title            string
	Metadata         map[string]any
}

type UpsertIntegrationRequestInput struct {
	WorkspaceID      uuid.UUID
	TeamID           uuid.UUID
	Provider         string
	SourceType       string
	SourceExternalID string
	SourceNumber     *int
	SourceURL        *string
	Title            string
	Description      *string
	Priority         string
	Metadata         map[string]any
	CreatedByUserID  *uuid.UUID
}

type CoreStorySyncInput struct {
	StoryID     uuid.UUID
	WorkspaceID uuid.UUID
	TeamID      uuid.UUID
	Title       string
	Description *string
	StatusID    *uuid.UUID
}

type StoryService interface {
	Get(ctx context.Context, id uuid.UUID, workspaceID uuid.UUID) (storydomain.Story, error)
	UpdateExternalWithReason(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, updates map[string]any, reason string) error
	RecordActivity(ctx context.Context, activity StoryActivity) error
	CreateCommentExternal(ctx context.Context, actorID uuid.UUID, workspaceID uuid.UUID, comment NewStoryComment) (storydomain.Comment, error)
}

type RequestStore interface {
	UpsertPending(ctx context.Context, input UpsertIntegrationRequestInput) (IntegrationRequest, error)
	Get(ctx context.Context, workspaceID, requestID uuid.UUID) (IntegrationRequest, error)
}

// AvatarResolver resolves stored avatar blob names to accessible URLs.
type AvatarResolver interface {
	ResolveProfileImageURL(ctx context.Context, avatar string, expiry time.Duration) (string, error)
}

// WorkspaceRoleReader provides an authoritative, workspace-scoped membership
// lookup for privileged integration operations.
type WorkspaceRoleReader interface {
	GetWorkspaceRole(ctx context.Context, workspaceID, actorID uuid.UUID) (authorization.WorkspaceRole, error)
}

// OAuthStateStore is the shared, replica-safe cache capability used for
// short-lived GitHub callback state. Implementations must make Take atomic.
type OAuthStateStore interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Take(ctx context.Context, key string, destination any) error
}

// CredentialVault is the narrow recoverable-secret capability consumed by the
// GitHub adapter. Provider SDK types and key material never cross this port.
type CredentialVault interface {
	Seal(binding credentialvault.Context, plaintext []byte) (string, error)
	Open(binding credentialvault.Context, envelope string) (credentialvault.Secret, error)
	Rewrap(binding credentialvault.Context, envelope string) (credentialvault.RewrapResult, error)
	ActiveKeyRef() (credentialvault.KeyRef, error)
}

type Config struct {
	AppID                int64
	AppSlug              string
	ClientID             string
	ClientSecret         string
	PrivateKeyBase64     string
	RedirectURL          string
	WebhookSecret        string
	WebsiteURL           string
	WebhookPayloadSecret string
	GitHubUserID         uuid.UUID
	CredentialVault      CredentialVault
	OAuthStateStore      OAuthStateStore
	WebhookGateway       *webhooks.Gateway
	WebhookInbox         webhooks.Inbox
	WebhookPayloads      WebhookPayloadOpener
	WebhookDispatcher    webhooks.Dispatcher
}
