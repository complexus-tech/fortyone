package integrationrequestdomain

import (
	"errors"
	"time"

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

	CommentDirectionInbound  = "inbound"
	CommentDirectionOutbound = "outbound"
)

var (
	ErrNotFound               = errors.New("integration request not found")
	ErrUnsupportedProvider    = errors.New("unsupported integration request provider")
	ErrRequestNotPending      = errors.New("integration request is not pending")
	ErrInvalidRequestProperty = errors.New("invalid integration request property")
	ErrProviderThreadNotFound = errors.New("integration request provider thread not found")
	ErrIdempotencyConflict    = errors.New("comment idempotency key was already used for different content")
)

type IntegrationRequest struct {
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

type UpsertRequestInput struct {
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

type ListRequestsFilter struct {
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

// OptionalValue distinguishes an omitted patch field from an explicitly null
// value. Value is nil only when the caller intentionally clears a nullable
// property.
type OptionalValue[T any] struct {
	Set   bool
	Value *T
}

type UpdateRequestInput struct {
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

type ProviderThread struct {
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

type Comment struct {
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

type ThreadActivity struct {
	Thread   ProviderThread
	Comments []Comment
}

type BindProviderThreadInput struct {
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

type ProviderThreadMatchInput struct {
	WorkspaceID            uuid.UUID
	UserID                 uuid.UUID
	Provider               string
	ExternalWorkspaceID    string
	InstallationGeneration uuid.UUID
	ExternalChannelID      string
	ExternalThreadID       string
}

// ProviderThreadLookupInput identifies a provider thread using only the signed
// provider envelope and current installation generation. External participants
// can use it without a linked FortyOne actor.
type ProviderThreadLookupInput struct {
	WorkspaceID            uuid.UUID
	Provider               string
	ExternalWorkspaceID    string
	InstallationGeneration uuid.UUID
	ExternalChannelID      string
	ExternalThreadID       string
}

type CreateCommentInput struct {
	WorkspaceID          uuid.UUID
	RequestID            uuid.UUID
	AuthorID             uuid.UUID
	ClientIdempotencyKey uuid.UUID
	Body                 string
}

// PreparedProviderComment is the provider-neutral envelope persisted in the
// same transaction as an outbound comment. ProviderPayload remains opaque to
// the integration request repository.
type PreparedProviderComment struct {
	ExternalRecipientUserID string
	ProviderPayload         []byte
}

type InboundProviderCommentInput struct {
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
