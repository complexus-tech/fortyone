package admin

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	defaultPageLimit = 20
	maxPageLimit     = 100
)

var (
	ErrForbidden          = errors.New("admin access requires an internal user")
	ErrInvalidAdminAction = errors.New("invalid admin action")
	ErrInvalidAdminNote   = errors.New("admin note body is required")
	ErrReasonRequired     = errors.New("admin action reason is required")
	ErrSelfMutation       = errors.New("admin users cannot change their own access state")
	ErrInvalidTrialEndsOn = errors.New("trial end must extend access into the future")
	ErrNotFound           = errors.New("admin resource not found")
)

type PaginationInput struct {
	Page  int
	Limit int
}

type Pagination struct {
	Total  int `json:"total"`
	Page   int `json:"page"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type ListResult[T any] struct {
	Items      []T        `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type DashboardSummary struct {
	TotalWorkspaces      int `json:"totalWorkspaces"`
	ActiveTrials         int `json:"activeTrials"`
	ExpiredTrials        int `json:"expiredTrials"`
	PaidWorkspaces       int `json:"paidWorkspaces"`
	DeletedWorkspaces    int `json:"deletedWorkspaces"`
	TotalUsers           int `json:"totalUsers"`
	InternalUsers        int `json:"internalUsers"`
	ActiveSubscriptions  int `json:"activeSubscriptions"`
	SlackInstallations   int `json:"slackInstallations"`
	GitHubInstallations  int `json:"githubInstallations"`
	RecentAdminAuditLogs int `json:"recentAdminAuditLogs"`
}

type UserSummary struct {
	ID                  uuid.UUID  `json:"id"`
	Username            string     `json:"username"`
	Email               string     `json:"email"`
	FullName            string     `json:"fullName"`
	AvatarURL           string     `json:"avatarUrl"`
	IsActive            bool       `json:"isActive"`
	IsSystem            bool       `json:"isSystem"`
	IsInternal          bool       `json:"isInternal"`
	LastLoginAt         *time.Time `json:"lastLoginAt"`
	LastUsedWorkspaceID *uuid.UUID `json:"lastUsedWorkspaceId"`
	LastUsedWorkspace   *string    `json:"lastUsedWorkspace"`
	GitHubUsername      *string    `json:"githubUsername"`
	WorkspaceCount      int        `json:"workspaceCount"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

type UserMembership struct {
	WorkspaceID   uuid.UUID `json:"workspaceId"`
	WorkspaceName string    `json:"workspaceName"`
	WorkspaceSlug string    `json:"workspaceSlug"`
	Role          string    `json:"role"`
	JoinedAt      time.Time `json:"joinedAt"`
}

type UserOverview struct {
	User        UserSummary      `json:"user"`
	Memberships []UserMembership `json:"memberships"`
}

type WorkspaceSummary struct {
	ID                   uuid.UUID  `json:"id"`
	Name                 string     `json:"name"`
	Slug                 string     `json:"slug"`
	AvatarURL            *string    `json:"avatarUrl"`
	Color                string     `json:"color"`
	TeamSize             string     `json:"teamSize"`
	TrialEndsOn          *time.Time `json:"trialEndsOn"`
	DeletedAt            *time.Time `json:"deletedAt"`
	LastAccessedAt       *time.Time `json:"lastAccessedAt"`
	CreatedByUserID      *uuid.UUID `json:"createdByUserId"`
	CreatedByEmail       *string    `json:"createdByEmail"`
	CreatedByName        *string    `json:"createdByName"`
	MemberCount          int        `json:"memberCount"`
	TeamCount            int        `json:"teamCount"`
	StoryCount           int        `json:"storyCount"`
	SubscriptionTier     *string    `json:"subscriptionTier"`
	SubscriptionStatus   *string    `json:"subscriptionStatus"`
	SubscriptionSeats    *int       `json:"subscriptionSeats"`
	StripeCustomerID     *string    `json:"stripeCustomerId"`
	StripeSubscriptionID *string    `json:"stripeSubscriptionId"`
	SlackInstalled       bool       `json:"slackInstalled"`
	GitHubInstalled      bool       `json:"githubInstalled"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

type WorkspaceMember struct {
	UserID     uuid.UUID `json:"userId"`
	Email      string    `json:"email"`
	FullName   string    `json:"fullName"`
	Role       string    `json:"role"`
	IsInternal bool      `json:"isInternal"`
	JoinedAt   time.Time `json:"joinedAt"`
}

type WorkspaceOverview struct {
	Workspace WorkspaceSummary  `json:"workspace"`
	Members   []WorkspaceMember `json:"members"`
}

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
	WorkspaceID uuid.UUID `json:"-"`
	TrialEndsOn time.Time `json:"trialEndsOn"`
	Reason      string    `json:"reason"`
}

type UpdateWorkspaceDeletedInput struct {
	WorkspaceID uuid.UUID `json:"-"`
	ActorUserID uuid.UUID `json:"-"`
	Deleted     bool      `json:"deleted"`
	Reason      string    `json:"reason"`
}

type RequestWorkspaceSubscriptionSyncInput struct {
	WorkspaceID uuid.UUID `json:"-"`
	Reason      string    `json:"reason"`
}

type UpdateUserStateInput struct {
	UserID     uuid.UUID `json:"-"`
	IsActive   *bool     `json:"isActive"`
	IsInternal *bool     `json:"isInternal"`
	Reason     string    `json:"reason"`
}

type RequestUserSessionRevocationInput struct {
	UserID uuid.UUID `json:"-"`
	Reason string    `json:"reason"`
}

type ListAdminNotesInput struct {
	Pagination  PaginationInput
	TargetType  string
	TargetID    *uuid.UUID
	WorkspaceID *uuid.UUID
}

type CreateAdminNoteInput struct {
	TargetType      string     `json:"targetType"`
	TargetID        uuid.UUID  `json:"targetId"`
	WorkspaceID     *uuid.UUID `json:"workspaceId"`
	Body            string     `json:"body"`
	CreatedByUserID uuid.UUID  `json:"-"`
}

type AuditEntryInput struct {
	ActorUserID uuid.UUID
	TargetType  string
	TargetID    *uuid.UUID
	WorkspaceID *uuid.UUID
	Action      string
	FieldName   string
	OldValue    any
	NewValue    any
	Reason      string
	Metadata    map[string]any
}

type AuditLog struct {
	ID            uuid.UUID  `json:"id"`
	ActorUserID   uuid.UUID  `json:"actorUserId"`
	ActorEmail    string     `json:"actorEmail"`
	ActorName     string     `json:"actorName"`
	TargetType    string     `json:"targetType"`
	TargetID      *uuid.UUID `json:"targetId"`
	WorkspaceID   *uuid.UUID `json:"workspaceId"`
	WorkspaceName *string    `json:"workspaceName"`
	WorkspaceSlug *string    `json:"workspaceSlug"`
	Action        string     `json:"action"`
	FieldName     string     `json:"fieldName"`
	OldValue      any        `json:"oldValue"`
	NewValue      any        `json:"newValue"`
	Reason        string     `json:"reason"`
	Metadata      any        `json:"metadata"`
	CreatedAt     time.Time  `json:"createdAt"`
}

type AdminNote struct {
	ID              uuid.UUID  `json:"id"`
	TargetType      string     `json:"targetType"`
	TargetID        uuid.UUID  `json:"targetId"`
	WorkspaceID     *uuid.UUID `json:"workspaceId"`
	Body            string     `json:"body"`
	CreatedByUserID uuid.UUID  `json:"createdByUserId"`
	CreatedByName   string     `json:"createdByName"`
	CreatedByEmail  string     `json:"createdByEmail"`
	CreatedAt       time.Time  `json:"createdAt"`
}
