// Package messaging provides provider-neutral conversational assistance for
// messaging integrations such as Slack, Microsoft Teams, and WhatsApp.
package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	messagingdomain "github.com/complexus-tech/projects-api/internal/modules/messaging/domain"
	"github.com/google/uuid"
)

// ConversationRole is a provider-neutral role in a visible conversation.
type ConversationRole string

const (
	RoleUser      ConversationRole = "user"
	RoleAssistant ConversationRole = "assistant"
)

var (
	ErrInvalidRequest         = errors.New("invalid assistant request")
	ErrMessageTooLarge        = errors.New("assistant message is too large")
	ErrAssistantNotConfigured = errors.New("assistant is not configured")
	ErrResponseRefused        = errors.New("assistant response refused")
	ErrResponseIncomplete     = errors.New("assistant response incomplete")
	ErrResponseFailed         = errors.New("assistant response failed")
	ErrMalformedResponse      = errors.New("malformed assistant response")
	ErrMaxToolSteps           = errors.New("assistant exceeded maximum tool steps")
	ErrUnknownTool            = errors.New("unknown assistant tool")
	ErrInvalidToolArguments   = errors.New("invalid assistant tool arguments")
	ErrToolExecution          = errors.New("assistant tool execution failed")
	ErrTeamNotAccessible      = errors.New("team is not accessible to the user")
	ErrMutationNotAllowed     = errors.New("story mutations are not allowed")
	ErrStaleMutation          = errors.New("story changed after mutation confirmation was requested")
)

var (
	ErrInvalidConfirmation   = messagingdomain.ErrInvalidConfirmation
	ErrExpiredConfirmation   = messagingdomain.ErrExpiredConfirmation
	ErrCancelledConfirmation = messagingdomain.ErrCancelledConfirmation
	ErrAppliedConfirmation   = messagingdomain.ErrAppliedConfirmation
)

// Assistant answers a provider-neutral conversation request. Provider adapters
// are responsible for loading and persisting the visible conversation turns.
type Assistant interface {
	Respond(ctx context.Context, request Request) (Response, error)
}

// Request contains the authoritative FortyOne identity and the visible
// conversation leading up to the current prompt.
type Request struct {
	WorkspaceID    uuid.UUID
	UserID         uuid.UUID
	AllowedTeamIDs []uuid.UUID
	// SharedTeamIDs is the explicit subset of AllowedTeamIDs whose work may be
	// queried across assignees on the current conversation surface. A nil or
	// empty slice denies shared-team queries.
	SharedTeamIDs []uuid.UUID
	// RuntimeContext contains server-authoritative, display-safe hints that help
	// the assistant resolve conversational references. It is descriptive only;
	// authorization remains exclusively in WorkspaceID, UserID,
	// AllowedTeamIDs, and SharedTeamIDs.
	RuntimeContext *RuntimeContext
	// Guidance is authoritative workspace-admin configuration, not message
	// content. It cannot widen authorization or bypass mutation confirmation.
	Guidance       string
	AllowMutations bool
	WebsiteURL     string
	// SourceURL is a provider-adapter supplied, server-authoritative link to
	// the exact conversation source used for this request. It must never be
	// populated from model tool arguments or untrusted message content.
	SourceURL    string
	Conversation []ConversationTurn
	Prompt       string
}

// RuntimeContext is provider-neutral, display-safe context for one assistant
// request. Its values are rendered as untrusted data, never as policy,
// permissions, or user confirmation. IDs deliberately do not belong here.
type RuntimeContext struct {
	Actor       RuntimeActorContext
	Workspace   RuntimeWorkspaceContext
	LocalTime   time.Time
	Terminology RuntimeTerminologyContext
	TeamHints   []RuntimeTeamHint
	Surface     RuntimeSurfaceContext
}

// RuntimeActorContext identifies the linked FortyOne user conversationally.
// Tool execution continues to use Request.UserID as the authoritative actor.
type RuntimeActorContext struct {
	DisplayName string
	Username    string
}

// RuntimeWorkspaceContext describes the active FortyOne workspace without
// exposing internal identifiers.
type RuntimeWorkspaceContext struct {
	Name string
	Slug string
	Role string
}

// RuntimeTerminologyContext carries the workspace's preferred product terms.
type RuntimeTerminologyContext struct {
	Story     RuntimeTerm
	Sprint    RuntimeTerm
	Objective RuntimeTerm
	KeyResult RuntimeTerm
}

// RuntimeTerm is a singular and plural workspace term.
type RuntimeTerm struct {
	Singular string
	Plural   string
}

// RuntimeTeamHint is an ordered, display-safe team hint. The corresponding
// authorization remains in Request.AllowedTeamIDs, with cross-assignee reads
// further constrained by Request.SharedTeamIDs.
type RuntimeTeamHint struct {
	Name string
	Code string
}

// RuntimeSurfaceKind identifies the provider-neutral conversation surface.
type RuntimeSurfaceKind string

const (
	RuntimeSurfaceDirect  RuntimeSurfaceKind = "direct"
	RuntimeSurfaceChannel RuntimeSurfaceKind = "channel"
	RuntimeSurfaceThread  RuntimeSurfaceKind = "thread"
)

// RuntimeSurfaceContext describes where the user is talking to the assistant.
// Location is a human-readable provider label such as a channel name, never a
// provider identifier. CurrentEntity is optional contextual work.
type RuntimeSurfaceContext struct {
	Provider      string
	Kind          RuntimeSurfaceKind
	Location      string
	CurrentEntity *RuntimeEntityHint
}

// RuntimeEntityHint is a display-safe reference to the work currently in
// context, for example a story or integration request connected to a thread.
type RuntimeEntityHint struct {
	Kind      string
	Reference string
	Title     string
}

// ConversationTurn is a visible user or assistant message. Provider-specific
// message identifiers intentionally live outside this package.
type ConversationTurn struct {
	Role ConversationRole
	Text string
}

// Response is the assistant's provider-neutral text response and aggregate
// model usage across every Responses API call needed to produce it.
type Response struct {
	Text         string
	Usage        Usage
	Confirmation *StoryMutationConfirmation
}

type Usage = messagingdomain.Usage

// ToolScope is the complete authorization scope available to a tool call.
type ToolScope struct {
	WorkspaceID    uuid.UUID
	UserID         uuid.UUID
	AllowedTeamIDs []uuid.UUID
	// SharedTeamIDs is the explicit subset of AllowedTeamIDs available to tools
	// that read work across assignees. Nil and empty both deny shared access.
	SharedTeamIDs  []uuid.UUID
	AllowMutations bool
	WebsiteURL     string
	// SourceURL is copied from Request.SourceURL after provider-neutral HTTPS
	// validation. Mutation tools may attach it as attribution, but cannot
	// accept or replace it through model-authored arguments.
	SourceURL     string
	WorkspaceSlug string
	Timezone      string
}

type StoryMutationOperation = messagingdomain.StoryMutationOperation

const (
	StoryMutationCreate      = messagingdomain.StoryMutationCreate
	StoryMutationCreateBatch = messagingdomain.StoryMutationCreateBatch
	StoryMutationUpdate      = messagingdomain.StoryMutationUpdate
	StoryMutationComment     = messagingdomain.StoryMutationComment
	StoryMutationRelation    = messagingdomain.StoryMutationRelation
)

// StoryMutationConfirmation is a provider-neutral, non-mutating proposal.
// Provider adapters must show this proposal to the user and call a
// StoryMutationConfirmer only after an explicit affirmative action.
type StoryMutationConfirmation struct {
	Operation StoryMutationOperation `json:"operation"`
	Token     string                 `json:"token"`
	ExpiresAt time.Time              `json:"expires_at"`
	Prompt    string                 `json:"prompt"`
	Story     StoryMutationPreview   `json:"story"`
	Stories   []StoryMutationPreview `json:"stories,omitempty"`
}

// StoryMutationPreview contains only display-safe details for a proposed
// story mutation. AssigneeAction is one of "unchanged", "me", "named", or
// "unassigned".
type StoryMutationPreview struct {
	StoryID                  *uuid.UUID `json:"story_id,omitempty"`
	Reference                string     `json:"reference,omitempty"`
	TeamID                   uuid.UUID  `json:"team_id"`
	TeamName                 string     `json:"team_name"`
	TeamCode                 string     `json:"team_code"`
	Title                    string     `json:"title"`
	Description              string     `json:"description,omitempty"`
	SourceURL                string     `json:"source_url,omitempty"`
	Priority                 *string    `json:"priority,omitempty"`
	AssigneeID               *uuid.UUID `json:"assignee_id,omitempty"`
	AssigneeName             string     `json:"assignee_name,omitempty"`
	AssigneeAction           string     `json:"assignee_action"`
	EstimatedDurationMinutes *int       `json:"estimated_duration_minutes,omitempty"`
	MinimumFocusBlockMinutes *int       `json:"minimum_focus_block_minutes,omitempty"`
	AutoSchedulingEnabled    *bool      `json:"auto_scheduling_enabled,omitempty"`
	AutoSchedulingLocked     *bool      `json:"auto_scheduling_locked,omitempty"`
	ChangedFields            []string   `json:"changed_fields,omitempty"`
}

type StoryMutationResult = messagingdomain.StoryMutationResult
type StoryMutationItemResult = messagingdomain.StoryMutationItemResult
type StoryMutationCancellationResult = messagingdomain.StoryMutationCancellationResult
type StoryMutationConfirmationStatus = messagingdomain.StoryMutationConfirmationStatus

const (
	StoryMutationConfirmationPending   = messagingdomain.StoryMutationConfirmationPending
	StoryMutationConfirmationApplied   = messagingdomain.StoryMutationConfirmationApplied
	StoryMutationConfirmationCancelled = messagingdomain.StoryMutationConfirmationCancelled
	StoryMutationConfirmationExpired   = messagingdomain.StoryMutationConfirmationExpired
)

type StoryMutationConfirmationStateInput = messagingdomain.StoryMutationConfirmationStateInput
type StoryMutationConfirmationRecord = messagingdomain.StoryMutationConfirmationRecord
type StoryMutationConfirmationBinding = messagingdomain.StoryMutationConfirmationBinding

// StoryMutationConfirmationStore arbitrates confirmation outcomes in durable
// storage. Apply must atomically consume pending consent before invoking apply,
// serialize concurrent applies, persist batch progress on errors, and return a
// completed persisted result on retries.
type StoryMutationConfirmationStore interface {
	RegisterStoryMutationConfirmation(ctx context.Context, input StoryMutationConfirmationStateInput) error
	LoadStoryMutationConfirmation(
		ctx context.Context,
		binding StoryMutationConfirmationBinding,
	) (StoryMutationConfirmationRecord, error)
	ApplyStoryMutationConfirmation(
		ctx context.Context,
		binding StoryMutationConfirmationBinding,
		now time.Time,
		apply func(context.Context) (StoryMutationResult, error),
	) (result StoryMutationResult, duplicate bool, err error)
	CancelStoryMutationConfirmation(
		ctx context.Context,
		binding StoryMutationConfirmationBinding,
		now time.Time,
	) (StoryMutationCancellationResult, error)
}

// StoryMutationConfirmer applies a previously proposed mutation. The caller
// must supply the same authoritative identity and current channel team scope
// used by the provider at confirmation time.
type StoryMutationConfirmer interface {
	ConfirmStoryMutation(ctx context.Context, scope ToolScope, token string) (StoryMutationResult, error)
	CancelStoryMutation(ctx context.Context, scope ToolScope, token string) (StoryMutationCancellationResult, error)
}

// ToolDefinition is an OpenAI Responses function tool definition. Parameters
// must be a strict JSON schema.
type ToolDefinition struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
	Strict      bool           `json:"strict"`
}

// ToolCall is a normalized Responses API function call.
type ToolCall struct {
	ID        string
	Name      string
	Arguments json.RawMessage
}

// ToolExecutor exposes a fixed tool catalog and executes calls within an
// authoritative FortyOne workspace/user/team scope. Mutating tool calls return
// proposals; they never apply writes directly.
type ToolExecutor interface {
	Definitions() []ToolDefinition
	Execute(ctx context.Context, scope ToolScope, call ToolCall) (json.RawMessage, error)
}
