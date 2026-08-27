package integrationrequests

import (
	"context"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

const (
	ProviderGitHub   = "github"
	ProviderSlack    = "slack"
	ProviderIntercom = "intercom"

	SourceTypeIssue = "issue"

	StatusPending  = "pending"
	StatusAccepted = "accepted"
	StatusDeclined = "declined"

	AcceptanceStateIdle     = "idle"
	AcceptanceStateReserved = "reserved"
)

type CoreIntegrationRequest struct {
	ID                        uuid.UUID
	WorkspaceID               uuid.UUID
	TeamID                    uuid.UUID
	Provider                  string
	SourceType                string
	SourceExternalID          string
	SourceNumber              *int
	SourceURL                 *string
	Title                     string
	Description               *string
	StatusID                  *uuid.UUID
	Priority                  string
	AssigneeID                *uuid.UUID
	EstimateValue             *int16
	EstimatedDurationMinutes  *int
	MinimumFocusBlockMinutes  *int
	ObjectiveID               *uuid.UUID
	KeyResultID               *uuid.UUID
	SprintID                  *uuid.UUID
	StartDate                 *time.Time
	EndDate                   *time.Time
	LabelIDs                  []uuid.UUID
	Status                    string
	Metadata                  map[string]any
	AcceptedStoryID           *uuid.UUID
	AcceptedByUserID          *uuid.UUID
	AcceptedAt                *time.Time
	DeclinedByUserID          *uuid.UUID
	DeclinedAt                *time.Time
	AcceptanceState           string
	AcceptanceStartedByUserID *uuid.UUID
	AcceptanceStartedAt       *time.Time
	CreatedByUserID           *uuid.UUID
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

type CoreUpsertRequestInput struct {
	WorkspaceID              uuid.UUID
	TeamID                   uuid.UUID
	Provider                 string
	SourceType               string
	SourceExternalID         string
	SourceNumber             *int
	SourceURL                *string
	Title                    string
	Description              *string
	StatusID                 *uuid.UUID
	Priority                 string
	AssigneeID               *uuid.UUID
	EstimateValue            *int16
	EstimatedDurationMinutes *int
	MinimumFocusBlockMinutes *int
	ObjectiveID              *uuid.UUID
	KeyResultID              *uuid.UUID
	SprintID                 *uuid.UUID
	StartDate                *time.Time
	EndDate                  *time.Time
	LabelIDs                 []uuid.UUID
	Metadata                 map[string]any
	CreatedByUserID          *uuid.UUID
}

type CoreListRequestsFilter struct {
	Search        string
	Status        string
	Provider      string
	Priority      string
	AssigneeID    *uuid.UUID
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	Page          int
	PageSize      int
}

type CoreBulkRequestResult struct {
	// Count and RequestIDs are retained for existing callers and contain only
	// successful mutations.
	Count          int
	RequestIDs     []uuid.UUID
	TotalCount     int
	SucceededCount int
	FailedCount    int
	Partial        bool
	Items          []CoreBulkRequestItemResult
}

type CoreBulkRequestItemResult struct {
	RequestID       uuid.UUID
	Success         bool
	Status          string
	AcceptedStoryID *uuid.UUID
	Error           string
}

// OptionalValue distinguishes an omitted patch field from an explicitly null
// value. Value is nil only when the caller intentionally clears a nullable
// property.
type OptionalValue[T any] struct {
	Set   bool
	Value *T
}

type CoreUpdateRequestInput struct {
	Title                    *string
	Description              OptionalValue[string]
	StatusID                 OptionalValue[uuid.UUID]
	Priority                 *string
	AssigneeID               OptionalValue[uuid.UUID]
	EstimateValue            OptionalValue[int16]
	EstimatedDurationMinutes OptionalValue[int]
	MinimumFocusBlockMinutes OptionalValue[int]
	ObjectiveID              OptionalValue[uuid.UUID]
	KeyResultID              OptionalValue[uuid.UUID]
	SprintID                 OptionalValue[uuid.UUID]
	StartDate                OptionalValue[time.Time]
	EndDate                  OptionalValue[time.Time]
	LabelIDs                 *[]uuid.UUID
}

const (
	CommentDirectionInbound  = "inbound"
	CommentDirectionOutbound = "outbound"
)

type CoreProviderThread struct {
	ID                      uuid.UUID
	WorkspaceID             uuid.UUID
	IntegrationRequestID    uuid.UUID
	TeamID                  uuid.UUID
	AcceptedStoryID         *uuid.UUID
	Provider                string
	ExternalWorkspaceID     string
	InstallationGeneration  *uuid.UUID
	ExternalChannelID       string
	ExternalThreadID        string
	ExternalSourceMessageID *string
	SourceURL               *string
	RequestTitle            string
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

type CoreIntegrationRequestComment struct {
	ID                     uuid.UUID
	WorkspaceID            uuid.UUID
	ThreadID               uuid.UUID
	Direction              string
	AuthorUserID           *uuid.UUID
	AuthorName             string
	AuthorAvatar           *string
	ExternalAuthorID       *string
	ExternalMessageID      *string
	ClientIdempotencyKey   *uuid.UUID
	OutboundIdempotencyKey *string
	DeliveryStatus         *string
	Body                   string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type CoreThreadActivity struct {
	Thread   CoreProviderThread
	Comments []CoreIntegrationRequestComment
}

type CoreBindProviderThreadInput struct {
	WorkspaceID             uuid.UUID
	IntegrationRequestID    uuid.UUID
	Provider                string
	ExternalWorkspaceID     string
	InstallationGeneration  *uuid.UUID
	ExternalChannelID       string
	ExternalThreadID        string
	ExternalSourceMessageID *string
	SourceURL               *string
}

type CoreProviderThreadMatchInput struct {
	WorkspaceID            uuid.UUID
	UserID                 uuid.UUID
	Provider               string
	ExternalWorkspaceID    string
	InstallationGeneration uuid.UUID
	ExternalChannelID      string
	ExternalThreadID       string
}

// CoreProviderThreadLookupInput identifies a provider thread using only the
// signed provider envelope and the current installation generation. It is
// intentionally independent of a FortyOne user because external participants
// can reply to an already-authorized Slack thread before linking an account.
type CoreProviderThreadLookupInput struct {
	WorkspaceID            uuid.UUID
	Provider               string
	ExternalWorkspaceID    string
	InstallationGeneration uuid.UUID
	ExternalChannelID      string
	ExternalThreadID       string
}

type CoreCreateCommentInput struct {
	WorkspaceID          uuid.UUID
	RequestID            uuid.UUID
	AuthorID             uuid.UUID
	ClientIdempotencyKey uuid.UUID
	Body                 string
}

// CorePreparedProviderComment is the provider-neutral envelope persisted in
// the same transaction as an outbound comment. ProviderPayload remains opaque
// to the integration request repository.
type CorePreparedProviderComment struct {
	ExternalRecipientUserID string
	ProviderPayload         []byte
}

type CoreInboundProviderCommentInput struct {
	Provider               string
	ExternalWorkspaceID    string
	InstallationGeneration uuid.UUID
	ExternalChannelID      string
	ExternalThreadID       string
	ExternalMessageID      string
	ExternalAuthorID       string
	AuthorUserID           *uuid.UUID
	Body                   string
	CreatedAt              time.Time
}

type StoryService interface {
	CreateExternalUserAction(ctx context.Context, actorID uuid.UUID, ns stories.CoreNewStory, workspaceID uuid.UUID) (stories.CoreSingleStory, error)
}

type ProviderAccepter interface {
	AcceptIntegrationRequest(ctx context.Context, request CoreIntegrationRequest, story stories.CoreSingleStory) error
}

type ProviderCommenter interface {
	PrepareIntegrationRequestComment(ctx context.Context, request CoreIntegrationRequest, thread CoreProviderThread, input CoreCreateCommentInput) (CorePreparedProviderComment, error)
	DeliverIntegrationRequestComment(ctx context.Context, request CoreIntegrationRequest, thread CoreProviderThread, comment CoreIntegrationRequestComment, prepared CorePreparedProviderComment) error
}
