package github

import (
	githubshared "github.com/complexus-tech/projects-api/internal/modules/github/shared"
	"github.com/google/uuid"
)

const (
	SyncDirectionInboundOnly            = githubshared.SyncDirectionInboundOnly
	SyncDirectionBidirectional          = githubshared.SyncDirectionBidirectional
	BranchFormatUsernameIdentifierTitle = githubshared.BranchFormatUsernameIdentifierTitle
	EventDraftPROpen                    = githubshared.EventDraftPROpen
	EventPROpen                         = githubshared.EventPROpen
	EventPRReviewActivity               = githubshared.EventPRReviewActivity
	EventPRReadyForMerge                = githubshared.EventPRReadyForMerge
	EventPRMerge                        = githubshared.EventPRMerge
	EventIssueOpen                      = githubshared.EventIssueOpen
	EventIssueReopen                    = githubshared.EventIssueReopen
	EventIssueClose                     = githubshared.EventIssueClose
)

type CoreWorkspaceSettings = githubshared.CoreWorkspaceSettings
type CoreInstallation = githubshared.CoreInstallation
type CoreRepository = githubshared.CoreRepository
type CoreIssueSyncLink = githubshared.CoreIssueSyncLink
type CoreWorkflowRule = githubshared.CoreWorkflowRule
type CoreIntegration = githubshared.CoreIntegration
type CoreCreateInstallSession = githubshared.CoreCreateInstallSession
type CoreIssueSyncLinkInput = githubshared.CoreIssueSyncLinkInput
type CoreUpdateIssueSyncLinkInput = githubshared.CoreUpdateIssueSyncLinkInput
type CoreUpdateWorkspaceSettingsInput = githubshared.CoreUpdateWorkspaceSettingsInput
type CoreTeamGitHubSettings = githubshared.CoreTeamGitHubSettings
type CoreWorkflowRuleInput = githubshared.CoreWorkflowRuleInput
type CoreUpdateTeamGitHubSettings = githubshared.CoreUpdateTeamGitHubSettings

type Config struct {
	AppID          int64
	AppSlug        string
	PrivateKeyPath string
	RedirectURL    string
	WebhookSecret  string
	WebsiteURL     string
	SecretKey      string
	GitHubUserID   uuid.UUID
}
