package githubshared

import (
	"time"

	"github.com/google/uuid"
)

const (
	SyncDirectionInboundOnly            = "inbound_only"
	SyncDirectionBidirectional          = "bidirectional"
	BranchFormatUsernameIdentifierTitle = "username/identifier-title"
	BranchFormatIdentifierTitle         = "identifier-title"
	BranchFormatIdentifierSlashTitle    = "identifier/title"
	EventDraftPROpen                    = "draft_pr_open"
	EventPROpen                         = "pr_open"
	EventPRReviewActivity               = "pr_review_activity"
	EventPRReadyForMerge                = "pr_ready_for_merge"
	EventPRMerge                        = "pr_merge"
	EventIssueOpen                      = "issue_open"
	EventIssueReopen                    = "issue_reopen"
	EventIssueClose                     = "issue_close"
)

type CoreWorkspaceSettings struct {
	WorkspaceID             uuid.UUID
	BranchFormat            string
	LinkCommitsByMagicWords bool
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

type CoreInstallation struct {
	ID                   uuid.UUID
	GitHubInstallationID int64
	AccountID            int64
	AccountLogin         string
	AccountType          string
	AccountAvatarURL     *string
	RepositorySelection  string
	IsActive             bool
	SuspendedAt          *time.Time
	DisconnectedAt       *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type CoreRepository struct {
	ID                 uuid.UUID
	InstallationID     uuid.UUID
	GitHubRepositoryID int64
	OwnerLogin         string
	Name               string
	FullName           string
	Description        *string
	HTMLURL            string
	DefaultBranch      string
	IsPrivate          bool
	IsArchived         bool
	IsDisabled         bool
	IsActive           bool
	LastSyncedAt       *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type CoreIssueSyncLink struct {
	ID             uuid.UUID
	RepositoryID   uuid.UUID
	RepositoryName string
	TeamID         uuid.UUID
	TeamName       string
	TeamColor      string
	SyncDirection  string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CoreWorkflowRule struct {
	ID                uuid.UUID
	EventKey          string
	TargetStatusID    *uuid.UUID
	BaseBranchPattern *string
	IsActive          bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type CoreIntegration struct {
	Settings       CoreWorkspaceSettings
	Installations  []CoreInstallation
	Repositories   []CoreRepository
	IssueSyncLinks []CoreIssueSyncLink
}

type CoreCreateInstallSession struct {
	InstallURL string
}

type CoreIssueSyncLinkInput struct {
	RepositoryID  uuid.UUID
	TeamID        uuid.UUID
	SyncDirection string
}

type CoreUpdateIssueSyncLinkInput struct {
	SyncDirection *string
	IsActive      *bool
}

type CoreUpdateWorkspaceSettingsInput struct {
	BranchFormat            *string
	LinkCommitsByMagicWords *bool
}

type CoreTeamGitHubSettings struct {
	TeamID uuid.UUID
	Rules  []CoreWorkflowRule
}

type CoreWorkflowRuleInput struct {
	EventKey          string
	TargetStatusID    *uuid.UUID
	BaseBranchPattern *string
	IsActive          bool
}

type CoreUpdateTeamGitHubSettings struct {
	Rules []CoreWorkflowRuleInput
}
