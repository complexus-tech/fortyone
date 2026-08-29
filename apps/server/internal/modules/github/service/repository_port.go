package github

import (
	"context"

	githubshared "github.com/complexus-tech/projects-api/internal/modules/github/shared"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
)

// Repository is the composition-root contract for GitHub persistence. The
// embedded capability ports keep business logic independent from SQLx, SQLC,
// and concrete repository adapters while preserving a single bootstrap seam.
type Repository interface {
	WorkspaceAuthorizationStore
	IntegrationSettingsStore
	InstallationStore
	StoryLinkStore
	CommentLinkStore
	UserIdentityStore
	CredentialStore
	WebhookInstallationRepository
}

type WorkspaceAuthorizationStore interface {
	GetWorkspaceRole(ctx context.Context, workspaceID, actorID uuid.UUID) (authorization.WorkspaceRole, error)
}

type IntegrationSettingsStore interface {
	GetWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID) (githubshared.CoreWorkspaceSettings, error)
	GetWorkspaceSettingsByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) (githubshared.CoreWorkspaceSettings, error)
	UpdateWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID, updates githubshared.CoreUpdateWorkspaceSettingsInput) (githubshared.CoreWorkspaceSettings, error)
	ListInstallations(ctx context.Context, workspaceID uuid.UUID) ([]githubshared.CoreInstallation, error)
	ListRepositories(ctx context.Context, workspaceID uuid.UUID) ([]githubshared.CoreRepository, error)
	ListIssueSyncLinks(ctx context.Context, workspaceID uuid.UUID) ([]githubshared.CoreIssueSyncLink, error)
	CreateIssueSyncLink(ctx context.Context, workspaceID, userID uuid.UUID, input githubshared.CoreIssueSyncLinkInput) (githubshared.CoreIssueSyncLink, error)
	UpdateIssueSyncLink(ctx context.Context, workspaceID, linkID uuid.UUID, input githubshared.CoreUpdateIssueSyncLinkInput) (githubshared.CoreIssueSyncLink, error)
	DeleteIssueSyncLink(ctx context.Context, workspaceID, linkID uuid.UUID) error
	GetTeamWorkflowSettings(ctx context.Context, workspaceID, teamID uuid.UUID) (githubshared.CoreTeamGitHubSettings, error)
	ReplaceTeamWorkflowSettings(ctx context.Context, workspaceID, teamID uuid.UUID, rules []githubshared.CoreWorkflowRuleInput) (githubshared.CoreTeamGitHubSettings, error)
	ListTeamStatuses(ctx context.Context, teamID uuid.UUID) ([]githubshared.TeamStatus, error)
}

type InstallationStore interface {
	UpsertInstallationWithRepositories(
		ctx context.Context,
		workspaceID, installedByUserID uuid.UUID,
		appID int64,
		installation githubshared.InstallationPayload,
		repositories []githubshared.RepositoryPayload,
	) error
	FindRepositoryByExternalID(ctx context.Context, externalRepositoryID int64) (githubshared.RepositoryRecord, error)
	FindRepositoryByID(ctx context.Context, workspaceID, repositoryID uuid.UUID) (githubshared.RepositoryRecord, error)
	FindIssueSyncLinkByRepositoryID(ctx context.Context, repositoryID uuid.UUID) (githubshared.IssueSyncLinkRecord, error)
	FindBidirectionalIssueSyncLinkByTeamID(ctx context.Context, workspaceID, teamID uuid.UUID) (githubshared.BidirectionalIssueSyncLink, error)
}

type StoryLinkStore interface {
	ResolveStoriesByRefs(ctx context.Context, workspaceID uuid.UUID, refs []string) ([]githubshared.StoryMatch, error)
	FindStoryLink(ctx context.Context, repositoryID uuid.UUID, externalType string, externalID int64, refName *string) (uuid.UUID, uuid.UUID, error)
	FindIssueStoryLinkByStoryID(ctx context.Context, workspaceID, storyID, repositoryID uuid.UUID) (githubshared.IssueStoryLink, error)
	FindIssueStoryLinkByExternalID(ctx context.Context, repositoryID uuid.UUID, externalID int64) (githubshared.IssueStoryLink, error)
	FindStoryLinksByPRNumber(ctx context.Context, repositoryID uuid.UUID, pullRequestNumber int) ([]githubshared.StoryMatch, error)
	GetStoryLinkedIssues(ctx context.Context, workspaceID, storyID uuid.UUID) ([]githubshared.StoryIssueWithInstallation, error)
	GetStoryGitHubLinks(ctx context.Context, workspaceID, storyID uuid.UUID) ([]githubshared.StoryGitHubLink, error)
	UpsertStoryLink(ctx context.Context, workspaceID, storyID, repositoryID uuid.UUID, externalType string, externalID int64, externalNumber int, refName *string, url, title, state string, metadata any) error
	UpdateStoryLinkSyncState(ctx context.Context, linkID uuid.UUID, source, syncHash string) error
	UpdateStoryLinkReviewState(ctx context.Context, storyID, repositoryID uuid.UUID, pullRequestExternalID int64, reviewState string, approved, changesRequested int) error
	UpdateStoryLinkCheckState(ctx context.Context, storyID, repositoryID uuid.UUID, pullRequestExternalID int64, checkState string) error
	DeleteStoryGitHubLink(ctx context.Context, workspaceID, linkID uuid.UUID) error
	EnsureStoryLink(ctx context.Context, storyID uuid.UUID, title *string, url string) error
	GetStatusCategory(ctx context.Context, statusID uuid.UUID) (string, error)
	ResolveOrCreateLabelsByName(ctx context.Context, workspaceID, teamID uuid.UUID, names []string) ([]uuid.UUID, error)
}

type CommentLinkStore interface {
	RecordOutboundGitHubComment(ctx context.Context, workspaceID, storyID, repositoryID uuid.UUID, externalCommentID int64, localCommentID *uuid.UUID, createdByUserID uuid.UUID) error
	ReserveInboundGitHubComment(ctx context.Context, workspaceID, storyID, repositoryID uuid.UUID, externalCommentID int64, createdByUserID uuid.UUID) (bool, error)
	CompleteInboundGitHubComment(ctx context.Context, repositoryID uuid.UUID, externalCommentID int64, localCommentID uuid.UUID) error
	DeleteGitHubCommentLink(ctx context.Context, repositoryID uuid.UUID, externalCommentID int64) error
	IsOutboundGitHubComment(ctx context.Context, repositoryID uuid.UUID, externalCommentID int64) (bool, error)
}

type UserIdentityStore interface {
	ResolveUserByGitHubID(ctx context.Context, externalUserID int64) (uuid.UUID, error)
	ResolveFortyOneUsersByGitHubIDs(ctx context.Context, externalUserIDs []int64) (map[int64]githubshared.FortyOneUser, error)
	ResolveFortyOneUserByFullName(ctx context.Context, fullName string) (githubshared.FortyOneUser, error)
	ResolveFortyOneUserByEmail(ctx context.Context, email string) (githubshared.FortyOneUser, error)
}

type CredentialStore interface {
	LinkGitHubUser(ctx context.Context, userID uuid.UUID, externalUserID int64, username, payload string, envelopeVersion int, generation uuid.UUID) error
	UnlinkGitHubUser(ctx context.Context, userID uuid.UUID) error
	GetUserGitHubCredential(ctx context.Context, userID uuid.UUID) (githubshared.CredentialRecord, error)
	ListGitHubUserCredentialsForRewrap(ctx context.Context, after *uuid.UUID, limit int) ([]githubshared.CredentialRecord, error)
	RewrapGitHubUserCredential(ctx context.Context, record githubshared.CredentialRecord, rewrapped string) (bool, error)
	ListLegacyGitHubUserCredentials(ctx context.Context, limit int) ([]githubshared.LegacyCredentialRecord, error)
	UpgradeLegacyGitHubUserCredential(ctx context.Context, userID uuid.UUID, expectedPlaintext, encrypted string, envelopeVersion int, generation uuid.UUID) error
}
