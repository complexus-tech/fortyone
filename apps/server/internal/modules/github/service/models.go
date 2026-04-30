package github

import (
	"context"
	"time"

	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	githubshared "github.com/complexus-tech/projects-api/internal/modules/github/shared"
	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
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

type CoreStorySyncInput struct {
	StoryID     uuid.UUID
	WorkspaceID uuid.UUID
	TeamID      uuid.UUID
	Title       string
	Description *string
	StatusID    *uuid.UUID
}

type StoryService interface {
	Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (stories.CoreSingleStory, error)
	CreateExternal(ctx context.Context, actorID uuid.UUID, ns stories.CoreNewStory, workspaceID uuid.UUID) (stories.CoreSingleStory, error)
	UpdateExternal(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, updates map[string]any) error
	RecordActivity(ctx context.Context, activity stories.CoreActivity) error
	CreateComment(ctx context.Context, workspaceID uuid.UUID, cnc stories.CoreNewComment) (comments.CoreComment, error)
	CreateCommentExternal(ctx context.Context, actorID uuid.UUID, workspaceID uuid.UUID, cnc stories.CoreNewComment) (comments.CoreComment, error)
}

type RequestStore interface {
	UpsertPending(ctx context.Context, input integrationrequests.CoreUpsertRequestInput) (integrationrequests.CoreIntegrationRequest, error)
	Get(ctx context.Context, workspaceID, requestID uuid.UUID) (integrationrequests.CoreIntegrationRequest, error)
}

// AvatarResolver resolves stored avatar blob names to accessible URLs.
type AvatarResolver interface {
	ResolveProfileImageURL(ctx context.Context, avatar string, expiry time.Duration) (string, error)
}

type Config struct {
	AppID            int64
	AppSlug          string
	ClientID         string
	ClientSecret     string
	PrivateKeyBase64 string
	RedirectURL      string
	WebhookSecret    string
	WebsiteURL       string
	SecretKey        string
	GitHubUserID     uuid.UUID
}
