package githubrepository

import (
	"time"

	githubshared "github.com/complexus-tech/projects-api/internal/modules/github/shared"
	"github.com/google/uuid"
)

func newCoreWorkspaceSettings(
	workspaceID uuid.UUID,
	branchFormat string,
	linkCommitsByMagicWords, syncAssignees, syncLabels, autoPopulatePRBody, closeOnCommitKeywords bool,
	createdAt, updatedAt time.Time,
) githubshared.CoreWorkspaceSettings {
	return githubshared.CoreWorkspaceSettings{
		WorkspaceID:             workspaceID,
		BranchFormat:            branchFormat,
		LinkCommitsByMagicWords: linkCommitsByMagicWords,
		SyncAssignees:           syncAssignees,
		SyncLabels:              syncLabels,
		AutoPopulatePRBody:      autoPopulatePRBody,
		CloseOnCommitKeywords:   closeOnCommitKeywords,
		CreatedAt:               createdAt,
		UpdatedAt:               updatedAt,
	}
}

type GithubInstallationPayload = githubshared.InstallationPayload
type GithubInstallationAccountPayload = githubshared.InstallationAccountPayload
type GithubInstallationSenderPayload = githubshared.InstallationSenderPayload
type GithubRepositoryPayload = githubshared.RepositoryPayload
type GithubRepositoryOwnerPayload = githubshared.RepositoryOwnerPayload
