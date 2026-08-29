package workspaces

import workspacedomain "github.com/complexus-tech/projects-api/internal/modules/workspaces/domain"

type CoreWorkspace = workspacedomain.Workspace
type CurrentMembership = workspacedomain.CurrentMembership
type DefaultStatus = workspacedomain.DefaultStatus
type CoreWorkspaceSettings = workspacedomain.WorkspaceSettings

var DefaultObjectiveStatuses = workspacedomain.DefaultObjectiveStatuses
