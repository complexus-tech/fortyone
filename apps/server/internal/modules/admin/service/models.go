package admin

import (
	"time"

	admindomain "github.com/complexus-tech/projects-api/internal/modules/admin/domain"
	"github.com/google/uuid"
)

const (
	defaultPageLimit = 20
	maxPageLimit     = 100
)

var (
	ErrForbidden              = admindomain.ErrForbidden
	ErrNotFound               = admindomain.ErrNotFound
	ErrConflict               = admindomain.ErrConflict
	ErrInvalidAdminAction     = admindomain.ErrInvalidAction
	ErrInvalidFilter          = admindomain.ErrInvalidFilter
	ErrInvalidAdminNote       = admindomain.ErrInvalidNote
	ErrReasonRequired         = admindomain.ErrReasonRequired
	ErrSelfMutation           = admindomain.ErrSelfMutation
	ErrInvalidTrialEndsOn     = admindomain.ErrInvalidTrialEndsOn
	ErrIntegrationUnavailable = admindomain.ErrIntegrationUnavailable
)

type PaginationInput struct {
	Page  int
	Limit int
}

type Pagination = admindomain.Pagination
type ListResult[T any] = admindomain.ListResult[T]
type DashboardSummary = admindomain.DashboardSummary
type UserSummary = admindomain.UserSummary
type UserMembership = admindomain.UserMembership
type UserOverview = admindomain.UserOverview
type WorkspaceSummary = admindomain.WorkspaceSummary
type WorkspaceMember = admindomain.WorkspaceMember
type WorkspaceOverview = admindomain.WorkspaceOverview
type AuditLog = admindomain.AuditLog
type AdminNote = admindomain.AdminNote

type ListWorkspacesInput struct {
	Pagination PaginationInput
	Query      string
	Status     string
}

type ListUsersInput struct {
	Pagination PaginationInput
	Query      string
}

type ListAuditLogsInput struct {
	Pagination  PaginationInput
	WorkspaceID *uuid.UUID
	TargetType  string
	Query       string
	Action      string
	ActorQuery  string
	From        *time.Time
	To          *time.Time
}

type UpdateWorkspaceTrialInput struct {
	TrialEndsOn time.Time `json:"trialEndsOn"`
	Reason      string    `json:"reason"`
}

type UpdateWorkspaceDeletedInput struct {
	Deleted bool   `json:"deleted"`
	Reason  string `json:"reason"`
}

type RequestWorkspaceSubscriptionSyncInput struct {
	Reason string `json:"reason"`
}

type UpdateUserStateInput struct {
	Patch  admindomain.UserStatePatch
	Reason string `json:"reason"`
}

type RequestUserSessionRevocationInput struct {
	Reason string `json:"reason"`
}

type ListAdminNotesInput struct {
	Pagination  PaginationInput
	TargetType  string
	TargetID    *uuid.UUID
	WorkspaceID *uuid.UUID
}

type CreateAdminNoteInput struct {
	TargetType  string     `json:"targetType"`
	TargetID    uuid.UUID  `json:"targetId"`
	WorkspaceID *uuid.UUID `json:"workspaceId"`
	Body        string     `json:"body"`
}
