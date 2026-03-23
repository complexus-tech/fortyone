package githubhttp

import (
	"time"

	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
	"github.com/google/uuid"
)

type AppWorkspaceSettings struct {
	BranchFormat            string    `json:"branchFormat"`
	LinkCommitsByMagicWords bool      `json:"linkCommitsByMagicWords"`
	SyncAssignees           bool      `json:"syncAssignees"`
	SyncLabels              bool      `json:"syncLabels"`
	AutoPopulatePRBody      bool      `json:"autoPopulatePrBody"`
	CloseOnCommitKeywords   bool      `json:"closeOnCommitKeywords"`
	CreatedAt               time.Time `json:"createdAt"`
	UpdatedAt               time.Time `json:"updatedAt"`
}

type AppInstallation struct {
	ID                   uuid.UUID `json:"id"`
	GitHubInstallationID int64     `json:"githubInstallationId"`
	AccountLogin         string    `json:"accountLogin"`
	AccountType          string    `json:"accountType"`
	AccountAvatarURL     *string   `json:"accountAvatarUrl,omitempty"`
	RepositorySelection  string    `json:"repositorySelection"`
	IsActive             bool      `json:"isActive"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

type AppRepository struct {
	ID                 uuid.UUID  `json:"id"`
	InstallationID     uuid.UUID  `json:"installationId"`
	GitHubRepositoryID int64      `json:"githubRepositoryId"`
	OwnerLogin         string     `json:"ownerLogin"`
	Name               string     `json:"name"`
	FullName           string     `json:"fullName"`
	Description        *string    `json:"description,omitempty"`
	HTMLURL            string     `json:"htmlUrl"`
	DefaultBranch      string     `json:"defaultBranch"`
	IsPrivate          bool       `json:"isPrivate"`
	IsArchived         bool       `json:"isArchived"`
	IsDisabled         bool       `json:"isDisabled"`
	IsActive           bool       `json:"isActive"`
	LastSyncedAt       *time.Time `json:"lastSyncedAt,omitempty"`
}

type AppIssueSyncLink struct {
	ID             uuid.UUID `json:"id"`
	RepositoryID   uuid.UUID `json:"repositoryId"`
	RepositoryName string    `json:"repositoryName"`
	TeamID         uuid.UUID `json:"teamId"`
	TeamName       string    `json:"teamName"`
	TeamColor      string    `json:"teamColor"`
	SyncDirection  string    `json:"syncDirection"`
	IsActive       bool      `json:"isActive"`
}

type AppWorkflowRule struct {
	ID                uuid.UUID  `json:"id"`
	EventKey          string     `json:"eventKey"`
	TargetStatusID    *uuid.UUID `json:"targetStatusId,omitempty"`
	BaseBranchPattern *string    `json:"baseBranchPattern,omitempty"`
	IsActive          bool       `json:"isActive"`
}

type AppIntegration struct {
	Settings       AppWorkspaceSettings `json:"settings"`
	Installations  []AppInstallation    `json:"installations"`
	Repositories   []AppRepository      `json:"repositories"`
	IssueSyncLinks []AppIssueSyncLink   `json:"issueSyncLinks"`
}

type AppCreateInstallSession struct {
	InstallURL string `json:"installUrl"`
}

type AppCreateIssueSyncLinkRequest struct {
	RepositoryID  uuid.UUID `json:"repositoryId"`
	TeamID        uuid.UUID `json:"teamId"`
	SyncDirection string    `json:"syncDirection"`
}

type AppUpdateIssueSyncLinkRequest struct {
	SyncDirection *string `json:"syncDirection,omitempty"`
	IsActive      *bool   `json:"isActive,omitempty"`
}

type AppUpdateWorkspaceSettingsRequest struct {
	BranchFormat            *string `json:"branchFormat,omitempty"`
	LinkCommitsByMagicWords *bool   `json:"linkCommitsByMagicWords,omitempty"`
	SyncAssignees           *bool   `json:"syncAssignees,omitempty"`
	SyncLabels              *bool   `json:"syncLabels,omitempty"`
	AutoPopulatePRBody      *bool   `json:"autoPopulatePrBody,omitempty"`
	CloseOnCommitKeywords   *bool   `json:"closeOnCommitKeywords,omitempty"`
}

type AppLinkGitHubUserRequest struct {
	Code string `json:"code" validate:"required"`
}

type AppPostGitHubCommentRequest struct {
	Body string `json:"body" validate:"required"`
}

type AppTeamGitHubSettings struct {
	TeamID uuid.UUID         `json:"teamId"`
	Rules  []AppWorkflowRule `json:"rules"`
}

type AppUpdateTeamGitHubSettingsRequest struct {
	Rules []AppUpdateWorkflowRule `json:"rules"`
}

type AppUpdateWorkflowRule struct {
	EventKey          string     `json:"eventKey"`
	TargetStatusID    *uuid.UUID `json:"targetStatusId,omitempty"`
	BaseBranchPattern *string    `json:"baseBranchPattern,omitempty"`
	IsActive          bool       `json:"isActive"`
}

func toAppIntegration(core github.CoreIntegration) AppIntegration {
	installations := make([]AppInstallation, 0, len(core.Installations))
	for _, item := range core.Installations {
		installations = append(installations, AppInstallation{
			ID: item.ID, GitHubInstallationID: item.GitHubInstallationID, AccountLogin: item.AccountLogin,
			AccountType: item.AccountType, AccountAvatarURL: item.AccountAvatarURL,
			RepositorySelection: item.RepositorySelection, IsActive: item.IsActive,
			CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
		})
	}
	repositories := make([]AppRepository, 0, len(core.Repositories))
	for _, item := range core.Repositories {
		repositories = append(repositories, AppRepository{
			ID: item.ID, InstallationID: item.InstallationID, GitHubRepositoryID: item.GitHubRepositoryID,
			OwnerLogin: item.OwnerLogin, Name: item.Name, FullName: item.FullName, Description: item.Description,
			HTMLURL: item.HTMLURL, DefaultBranch: item.DefaultBranch, IsPrivate: item.IsPrivate,
			IsArchived: item.IsArchived, IsDisabled: item.IsDisabled, IsActive: item.IsActive,
			LastSyncedAt: item.LastSyncedAt,
		})
	}
	links := make([]AppIssueSyncLink, 0, len(core.IssueSyncLinks))
	for _, item := range core.IssueSyncLinks {
		links = append(links, AppIssueSyncLink{
			ID: item.ID, RepositoryID: item.RepositoryID, RepositoryName: item.RepositoryName, TeamID: item.TeamID,
			TeamName: item.TeamName, TeamColor: item.TeamColor, SyncDirection: item.SyncDirection, IsActive: item.IsActive,
		})
	}
	return AppIntegration{
		Settings: AppWorkspaceSettings{
			BranchFormat:            core.Settings.BranchFormat,
			LinkCommitsByMagicWords: core.Settings.LinkCommitsByMagicWords,
			SyncAssignees:           core.Settings.SyncAssignees,
			SyncLabels:              core.Settings.SyncLabels,
			AutoPopulatePRBody:      core.Settings.AutoPopulatePRBody,
			CloseOnCommitKeywords:   core.Settings.CloseOnCommitKeywords,
			CreatedAt:               core.Settings.CreatedAt,
			UpdatedAt:               core.Settings.UpdatedAt,
		},
		Installations:  installations,
		Repositories:   repositories,
		IssueSyncLinks: links,
	}
}

func toCoreIssueSyncLinkInput(input AppCreateIssueSyncLinkRequest) github.CoreIssueSyncLinkInput {
	return github.CoreIssueSyncLinkInput{
		RepositoryID:  input.RepositoryID,
		TeamID:        input.TeamID,
		SyncDirection: input.SyncDirection,
	}
}

func toCoreIssueSyncLinkUpdate(input AppUpdateIssueSyncLinkRequest) github.CoreUpdateIssueSyncLinkInput {
	return github.CoreUpdateIssueSyncLinkInput{
		SyncDirection: input.SyncDirection,
		IsActive:      input.IsActive,
	}
}

func toAppIssueSyncLink(core github.CoreIssueSyncLink) AppIssueSyncLink {
	return AppIssueSyncLink{
		ID: core.ID, RepositoryID: core.RepositoryID, RepositoryName: core.RepositoryName, TeamID: core.TeamID,
		TeamName: core.TeamName, TeamColor: core.TeamColor, SyncDirection: core.SyncDirection, IsActive: core.IsActive,
	}
}

func toCoreWorkspaceSettingsUpdate(input AppUpdateWorkspaceSettingsRequest) github.CoreUpdateWorkspaceSettingsInput {
	return github.CoreUpdateWorkspaceSettingsInput{
		BranchFormat:            input.BranchFormat,
		LinkCommitsByMagicWords: input.LinkCommitsByMagicWords,
		SyncAssignees:           input.SyncAssignees,
		SyncLabels:              input.SyncLabels,
		AutoPopulatePRBody:      input.AutoPopulatePRBody,
		CloseOnCommitKeywords:   input.CloseOnCommitKeywords,
	}
}

func toAppWorkspaceSettings(core github.CoreWorkspaceSettings) AppWorkspaceSettings {
	return AppWorkspaceSettings{
		BranchFormat:            core.BranchFormat,
		LinkCommitsByMagicWords: core.LinkCommitsByMagicWords,
		SyncAssignees:           core.SyncAssignees,
		SyncLabels:              core.SyncLabels,
		AutoPopulatePRBody:      core.AutoPopulatePRBody,
		CloseOnCommitKeywords:   core.CloseOnCommitKeywords,
		CreatedAt:               core.CreatedAt,
		UpdatedAt:               core.UpdatedAt,
	}
}

func toAppTeamSettings(core github.CoreTeamGitHubSettings) AppTeamGitHubSettings {
	rules := make([]AppWorkflowRule, 0, len(core.Rules))
	for _, rule := range core.Rules {
		rules = append(rules, AppWorkflowRule{
			ID: rule.ID, EventKey: rule.EventKey, TargetStatusID: rule.TargetStatusID,
			BaseBranchPattern: rule.BaseBranchPattern, IsActive: rule.IsActive,
		})
	}
	return AppTeamGitHubSettings{TeamID: core.TeamID, Rules: rules}
}

func toCoreTeamSettingsUpdate(input AppUpdateTeamGitHubSettingsRequest) github.CoreUpdateTeamGitHubSettings {
	rules := make([]github.CoreWorkflowRuleInput, 0, len(input.Rules))
	for _, rule := range input.Rules {
		rules = append(rules, github.CoreWorkflowRuleInput{
			EventKey:          rule.EventKey,
			TargetStatusID:    rule.TargetStatusID,
			BaseBranchPattern: rule.BaseBranchPattern,
			IsActive:          rule.IsActive,
		})
	}
	return github.CoreUpdateTeamGitHubSettings{Rules: rules}
}
