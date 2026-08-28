package slack

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	ProviderSlack = "slack"

	IntegrationRequestStatusPending  = "pending"
	IntegrationRequestStatusAccepted = "accepted"
	IntegrationRequestStatusDeclined = "declined"

	ConversationAudienceActor   = "actor"
	ConversationAudienceChannel = "channel"

	AssistantRoleUser      AssistantConversationRole = "user"
	AssistantRoleAssistant AssistantConversationRole = "assistant"

	AssistantSurfaceDirect  AssistantSurfaceKind = "direct"
	AssistantSurfaceChannel AssistantSurfaceKind = "channel"
	AssistantSurfaceThread  AssistantSurfaceKind = "thread"

	StoryMutationCreate      StoryMutationOperation = "create_story"
	StoryMutationCreateBatch StoryMutationOperation = "create_stories"
)

// IntegrationRequest is the provider-neutral request projection needed by
// Slack. The integration-requests module is translated into this contract at
// the composition root.
type IntegrationRequest struct {
	ID                       uuid.UUID
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
	Status                   string
	Metadata                 map[string]any
	AcceptedStoryID          *uuid.UUID
	CreatedByUserID          *uuid.UUID
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

type UpsertIntegrationRequestInput struct {
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

type ProviderThreadLookupInput struct {
	WorkspaceID            uuid.UUID
	Provider               string
	ExternalWorkspaceID    string
	InstallationGeneration uuid.UUID
	ExternalChannelID      string
	ExternalThreadID       string
}

type CreateIntegrationRequestCommentInput struct {
	WorkspaceID          uuid.UUID
	RequestID            uuid.UUID
	AuthorID             uuid.UUID
	ClientIdempotencyKey uuid.UUID
	Body                 string
}

type PreparedProviderComment struct {
	ExternalRecipientUserID string
	ProviderPayload         []byte
}

type IntegrationRequestComment struct {
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

type Story struct {
	ID              uuid.UUID
	SequenceID      int
	Title           string
	TeamCode        string
	Description     *string
	DescriptionHTML *string
	Status          *uuid.UUID
	Assignee        *uuid.UUID
	Reporter        *uuid.UUID
	Priority        string
	Team            uuid.UUID
	Workspace       uuid.UUID
	EndDate         *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CreatedNow      bool
}

type NewStory struct {
	Title                    string
	EstimateValue            *int16
	EstimatedDurationMinutes *int
	MinimumFocusBlockMinutes *int
	AutoSchedulingEnabled    bool
	AutoSchedulingLocked     bool
	Description              *string
	DescriptionHTML          *string
	Objective                *uuid.UUID
	Status                   *uuid.UUID
	Assignee                 *uuid.UUID
	Reporter                 *uuid.UUID
	Priority                 string
	Sprint                   *uuid.UUID
	KeyResult                *uuid.UUID
	LabelIDs                 []uuid.UUID
	StartDate                *time.Time
	EndDate                  *time.Time
	Team                     uuid.UUID
	CreationKey              *string
}

type Objective struct {
	ID               uuid.UUID
	Name             string
	Description      *string
	LeadUser         *uuid.UUID
	Team             uuid.UUID
	Workspace        uuid.UUID
	StartDate        *time.Time
	EndDate          *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Health           *string
	TotalStories     int
	CompletedStories int
}

type Sprint struct {
	ID               uuid.UUID
	Name             string
	Goal             *string
	TeamID           uuid.UUID
	WorkspaceID      uuid.UUID
	StartDate        time.Time
	EndDate          time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	TotalStories     int
	CompletedStories int
}

type ConversationInput struct {
	Provider            string
	WorkspaceID         uuid.UUID
	ExternalWorkspaceID string
	ExternalChannelID   string
	ExternalThreadID    string
	UserID              uuid.UUID
	AudienceScope       string
	AudienceFingerprint string
}

type ConversationRecord struct {
	ID        uuid.UUID
	UpdatedAt time.Time
}

type MessageRecord struct {
	ExternalMessageID *string
	Role              string
	Content           string
	CreatedAt         time.Time
}

type NonceInput struct {
	Provider            string
	Purpose             string
	NonceHash           []byte
	WorkspaceID         uuid.UUID
	UserID              *uuid.UUID
	ExternalWorkspaceID string
	ExternalUserID      string
	Payload             json.RawMessage
	ExpiresAt           time.Time
}

type NonceConsumeInput struct {
	Provider    string
	Purpose     string
	NonceHash   []byte
	WorkspaceID *uuid.UUID
	UserID      *uuid.UUID
	Now         time.Time
}

type NonceRecord struct {
	ID                  uuid.UUID
	Provider            string
	Purpose             string
	WorkspaceID         uuid.UUID
	UserID              *uuid.UUID
	ExternalWorkspaceID *string
	ExternalUserID      *string
	Payload             json.RawMessage
	ExpiresAt           time.Time
	ConsumedAt          *time.Time
}

type OutboundDeliveryInput struct {
	Provider                string
	WorkspaceID             uuid.UUID
	UserID                  *uuid.UUID
	InstallGeneration       *uuid.UUID
	ExternalWorkspaceID     string
	ExternalRecipientUserID string
	InboundEventID          *uuid.UUID
	IdempotencyKey          string
	ExternalChannelID       string
	ExternalThreadID        string
	Content                 string
	ProviderPayload         []byte
	Purpose                 string
	ExpiresAt               *time.Time
}

type OutboundDeliveryRecord struct {
	ID                      uuid.UUID
	WorkspaceID             uuid.UUID
	UserID                  *uuid.UUID
	InstallGeneration       *uuid.UUID
	ExternalWorkspaceID     string
	ExternalRecipientUserID *string
	InboundEventID          *uuid.UUID
	IdempotencyKey          string
	ExternalChannelID       string
	ExternalThreadID        *string
	ExternalMessageID       *string
	Content                 *string
	ProviderPayload         []byte
	Status                  string
	AttemptCount            int
	Purpose                 string
	ExpiresAt               *time.Time
}

type AssistantUsage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

type DailyUsageRecordInput struct {
	InboundEventID      uuid.UUID
	WorkspaceID         uuid.UUID
	Provider            string
	ExternalWorkspaceID string
	ExternalEventID     string
	AttemptCount        int
	Usage               AssistantUsage
}

type DailyUsageSnapshot struct {
	InputTokens  int64
	OutputTokens int64
	TotalTokens  int64
	RequestCount int64
	Limit        int64
	Remaining    int64
	Allowed      bool
}

type AssistantConversationRole string

type AssistantConversationTurn struct {
	Role AssistantConversationRole
	Text string
}

type AssistantSurfaceKind string

type AssistantSurfaceContext struct {
	Provider      string
	Kind          AssistantSurfaceKind
	Location      string
	CurrentEntity *AssistantEntityHint
}

type AssistantEntityHint struct {
	Kind      string
	Reference string
	Title     string
}

type AssistantRuntimeContext struct {
	Actor       AssistantActorContext
	Workspace   AssistantWorkspaceContext
	LocalTime   time.Time
	Terminology AssistantTerminologyContext
	TeamHints   []AssistantTeamHint
	Surface     AssistantSurfaceContext
}

type AssistantActorContext struct {
	DisplayName string
	Username    string
}

type AssistantWorkspaceContext struct {
	Name string
	Slug string
	Role string
}

type AssistantTerminologyContext struct {
	Story     AssistantTerm
	Sprint    AssistantTerm
	Objective AssistantTerm
	KeyResult AssistantTerm
}

type AssistantTerm struct {
	Singular string
	Plural   string
}

type AssistantTeamHint struct {
	Name string
	Code string
}

type AssistantRequest struct {
	WorkspaceID    uuid.UUID
	UserID         uuid.UUID
	AllowedTeamIDs []uuid.UUID
	SharedTeamIDs  []uuid.UUID
	RuntimeContext *AssistantRuntimeContext
	Guidance       string
	AllowMutations bool
	WebsiteURL     string
	SourceURL      string
	Conversation   []AssistantConversationTurn
	Prompt         string
}

type StoryMutationOperation string

type StoryMutationConfirmation struct {
	Operation StoryMutationOperation
	Token     string
	ExpiresAt time.Time
	Prompt    string
}

type AssistantResponse struct {
	Text         string
	Usage        AssistantUsage
	Confirmation *StoryMutationConfirmation
}

type AssistantAdmissionInput struct {
	Provider            string
	WorkspaceID         uuid.UUID
	UserID              uuid.UUID
	ExternalWorkspaceID string
	ExternalEventID     string
}

type AssistantAdmissionDecision struct {
	Allowed        bool
	Duplicate      bool
	LimitedScope   string
	RetryAfter     time.Duration
	UserCount      int64
	WorkspaceCount int64
}

type StoryMutationScope struct {
	WorkspaceID    uuid.UUID
	UserID         uuid.UUID
	AllowedTeamIDs []uuid.UUID
	SharedTeamIDs  []uuid.UUID
	AllowMutations bool
	WebsiteURL     string
	SourceURL      string
	WorkspaceSlug  string
	Timezone       string
}

type StoryMutationResult struct {
	Status                   string
	Operation                StoryMutationOperation
	StoryID                  uuid.UUID
	Reference                string
	TeamID                   uuid.UUID
	Title                    string
	Priority                 string
	AssigneeID               *uuid.UUID
	EstimatedDurationMinutes *int
	MinimumFocusBlockMinutes *int
	AutoSchedulingEnabled    bool
	AutoSchedulingLocked     bool
	AutoSchedulingStatus     string
	AutoSchedulingReason     *string
	AutoSchedulingUpdatedAt  *time.Time
	CommentID                *uuid.UUID
	AssociationID            *uuid.UUID
	Items                    []StoryMutationItemResult
}

type StoryMutationItemResult struct {
	Index                    int
	Status                   string
	StoryID                  uuid.UUID
	Reference                string
	TeamID                   uuid.UUID
	Title                    string
	Priority                 string
	AssigneeID               *uuid.UUID
	EstimatedDurationMinutes *int
	MinimumFocusBlockMinutes *int
	AutoSchedulingEnabled    bool
	AutoSchedulingLocked     bool
	AutoSchedulingStatus     string
	AutoSchedulingReason     *string
	AutoSchedulingUpdatedAt  *time.Time
}

type StoryMutationCancellationResult struct {
	Status string
}

// AssistantAPIError is the provider-neutral diagnostic projection returned by
// a composition adapter. Slack logs useful metadata without importing the
// assistant provider implementation or SDK error types.
type AssistantAPIError struct {
	StatusCode int
	Code       string
	Message    string
	RequestID  string
	Permanent  bool
}

func (e *AssistantAPIError) Error() string {
	if e == nil {
		return "assistant provider error"
	}
	if e.Message != "" {
		return e.Message
	}
	return "assistant provider error"
}
