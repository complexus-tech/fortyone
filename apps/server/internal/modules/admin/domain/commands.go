package admindomain

import (
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/pagination"
	platformpatch "github.com/complexus-tech/projects-api/internal/platform/patch"
	"github.com/google/uuid"
)

type ListWorkspacesQuery struct {
	ActorID uuid.UUID
	Page    pagination.OffsetParams
	Search  string
	Status  WorkspaceStatus
	Now     time.Time
}

type DashboardSummaryQuery struct {
	ActorID uuid.UUID
	Now     time.Time
}

type GetWorkspaceQuery struct {
	ActorID     uuid.UUID
	WorkspaceID uuid.UUID
}

type UpdateWorkspaceTrialCommand struct {
	ActorID     uuid.UUID
	WorkspaceID uuid.UUID
	TrialEndsOn time.Time
	Reason      string
	Now         time.Time
}

type SetWorkspaceDeletedCommand struct {
	ActorID     uuid.UUID
	WorkspaceID uuid.UUID
	Deleted     bool
	Reason      string
	Now         time.Time
}

type ListUsersQuery struct {
	ActorID uuid.UUID
	Page    pagination.OffsetParams
	Search  string
}

type GetUserQuery struct {
	ActorID uuid.UUID
	UserID  uuid.UUID
}

type UserStatePatch struct {
	IsActive   platformpatch.Field[bool]
	IsInternal platformpatch.Field[bool]
}

func (patch UserStatePatch) Empty() bool {
	return !patch.IsActive.Specified() && !patch.IsInternal.Specified()
}

type UpdateUserStateCommand struct {
	ActorID uuid.UUID
	UserID  uuid.UUID
	Patch   UserStatePatch
	Reason  string
	Now     time.Time
}

type RequestSessionRevocationCommand struct {
	ActorID uuid.UUID
	UserID  uuid.UUID
	Reason  string
}

type ListAuditLogsQuery struct {
	ActorID     uuid.UUID
	Page        pagination.OffsetParams
	WorkspaceID *uuid.UUID
	TargetType  TargetType
	Search      string
	Action      AuditAction
	ActorSearch string
	From        *time.Time
	To          *time.Time
}

type ListAdminNotesQuery struct {
	ActorID     uuid.UUID
	Page        pagination.OffsetParams
	TargetType  TargetType
	TargetID    *uuid.UUID
	WorkspaceID *uuid.UUID
}

type CreateAdminNoteCommand struct {
	ActorID     uuid.UUID
	TargetType  TargetType
	TargetID    uuid.UUID
	WorkspaceID *uuid.UUID
	Body        string
}

type BeginSubscriptionSyncCommand struct {
	ActorID     uuid.UUID
	WorkspaceID uuid.UUID
	Reason      string
}

type SubscriptionSyncAttempt struct {
	AuditID      uuid.UUID
	ActorID      uuid.UUID
	WorkspaceID  uuid.UUID
	Reason       string
	BeforeStatus *string
}

type FinishSubscriptionSyncCommand struct {
	Attempt SubscriptionSyncAttempt
	Outcome SubscriptionSyncOutcome
}

func NormalizeSearch(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func RequireReason(value string) (string, error) {
	reason := strings.TrimSpace(value)
	if reason == "" {
		return "", ErrReasonRequired
	}
	return reason, nil
}
