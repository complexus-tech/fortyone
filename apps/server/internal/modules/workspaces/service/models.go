package workspaces

import workspacedomain "github.com/complexus-tech/projects-api/internal/modules/workspaces/domain"

type CoreWorkspace = workspacedomain.Workspace
type CurrentMembership = workspacedomain.CurrentMembership
type DefaultStatus = workspacedomain.DefaultStatus
type CoreWorkspaceSettings = workspacedomain.WorkspaceSettings
type CreationOptions = workspacedomain.CreationOptions
type WorkType = workspacedomain.WorkType

const (
	WorkTypeProduct    = workspacedomain.WorkTypeProduct
	WorkTypeMarketing  = workspacedomain.WorkTypeMarketing
	WorkTypeOperations = workspacedomain.WorkTypeOperations
	WorkTypePersonal   = workspacedomain.WorkTypePersonal
	WorkTypeGeneral    = workspacedomain.WorkTypeGeneral
)

var DefaultObjectiveStatuses = workspacedomain.DefaultObjectiveStatuses
