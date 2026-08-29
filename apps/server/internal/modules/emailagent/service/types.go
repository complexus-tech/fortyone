// Package emailagent provides the provider-neutral decision boundary for
// replies sent to Maya by email. It classifies a reply, produces safe email
// copy, and resolves model-selected references into inert mutation proposals.
// It never applies a domain mutation.
package emailagent

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Intent identifies the only outcomes the email agent may return.
type Intent string

const (
	IntentAnswer  Intent = "answer"
	IntentClarify Intent = "clarify"
	IntentPropose Intent = "propose"
	IntentCancel  Intent = "cancel"
	IntentConfirm Intent = "confirm"
	IntentRefuse  Intent = "refuse"
)

// Generator produces an untrusted structured decision from model-safe input.
// Implementations should use ModelInstructions and ResponseFormat with a
// strict structured-output API.
type Generator interface {
	Generate(ctx context.Context, request ModelRequest) (Generation, error)
}

// Summarizer incrementally compacts newly omitted durable turns into a bounded
// factual summary. Callers persist the returned summary with their thread.
type Summarizer interface {
	Summarize(ctx context.Context, request SummaryRequest) (SummaryGeneration, error)
}

// Generation includes provider usage for quota accounting and observability.
type Generation struct {
	Decision ModelDecision
	Usage    Usage
}

// Usage is aggregate model token usage for one decision.
type Usage struct {
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
	TotalTokens  int `json:"totalTokens"`
}

// SummaryRequest contains the prior durable summary and only the newly omitted
// chronological turns. SafetyIdentifier is raw caller identity; concrete
// adapters must hash it before transmission.
type SummaryRequest struct {
	SafetyIdentifier string        `json:"-"`
	PreviousSummary  string        `json:"previousSummary"`
	OmittedTurns     []HistoryTurn `json:"omittedTurns"`
}

// SummaryGeneration is the validated replacement durable summary.
type SummaryGeneration struct {
	Summary string
	Usage   Usage
}

// Request contains server-authoritative scope and the visible email
// conversation. The caller must load only targets the actor is currently
// allowed to see. Service validation is defense in depth; confirmation-time
// code must still re-read and reauthorize every target.
type Request struct {
	WorkspaceID      uuid.UUID
	ActorID          uuid.UUID
	SafetyIdentifier string
	AllowedTeamIDs   []uuid.UUID
	Subject          string
	Message          string
	Summary          string
	History          []HistoryTurn
	Facts            []GroundedFact
	Targets          []AuthorizedTarget
	Choices          []AuthorizedChoice
	PendingProposals []PendingProposal
}

// ConversationRole is a visible role in the durable email conversation.
type ConversationRole string

const (
	RoleUser      ConversationRole = "user"
	RoleAssistant ConversationRole = "assistant"
)

// HistoryTurn is one durable visible email turn. Provider message IDs remain
// outside this package because they are transport and idempotency concerns.
type HistoryTurn struct {
	Role   ConversationRole `json:"role"`
	Text   string           `json:"text"`
	SentAt time.Time        `json:"sentAt,omitempty"`
}

// HistoryWindow is the bounded chronological context sent to the generator.
type HistoryWindow struct {
	Turns        []HistoryTurn `json:"turns"`
	OmittedTurns int           `json:"omittedTurns"`
	Truncated    bool          `json:"truncated"`
}

// HistoryLimits bounds durable context before it reaches a model.
type HistoryLimits struct {
	MaxTurns        int
	MaxTotalRunes   int
	MaxTurnRunes    int
	MaxSummaryRunes int
}

// GroundedFact is server-supplied context. Text is data, never an instruction.
// The model must cite Reference in every copy block that relies on the fact.
type GroundedFact struct {
	Reference       string   `json:"reference"`
	Text            string   `json:"text"`
	ProtectedTokens []string `json:"protectedTokens"`
}

// TargetKind identifies an entity for which an action can be proposed.
type TargetKind string

const (
	TargetObjective TargetKind = "objective"
	TargetKeyResult TargetKind = "key_result"
	TargetStory     TargetKind = "story"
	TargetFeedback  TargetKind = "feedback"
)

// AuthorizedTarget binds a model-visible reference to server-authoritative
// identity and version data. IDs and versions are deliberately omitted from
// ModelRequest.
type AuthorizedTarget struct {
	Reference         string
	Kind              TargetKind
	DisplayName       string
	CurrentState      string
	ID                uuid.UUID
	TeamID            uuid.UUID
	ExpectedUpdatedAt time.Time
}

// ModelTarget is the display-safe target shape visible to a generator.
type ModelTarget struct {
	Reference    string     `json:"reference"`
	Kind         TargetKind `json:"kind"`
	DisplayName  string     `json:"displayName"`
	CurrentState string     `json:"currentState"`
}

// ChoiceKind identifies a server-authoritative value that can be selected for
// a story action.
type ChoiceKind string

const (
	ChoiceStoryStatus   ChoiceKind = "story_status"
	ChoiceStoryAssignee ChoiceKind = "story_assignee"
)

// AuthorizedChoice binds a model-visible reference to an ID. TeamID scopes a
// status or member choice to the target story's team.
type AuthorizedChoice struct {
	Reference   string
	Kind        ChoiceKind
	DisplayName string
	ID          uuid.UUID
	TeamID      uuid.UUID
}

// ModelChoice is the display-safe choice shape visible to a generator.
type ModelChoice struct {
	Reference   string     `json:"reference"`
	Kind        ChoiceKind `json:"kind"`
	DisplayName string     `json:"displayName"`
}

// PendingProposal identifies durable state owned by the caller. Its ID is
// never included in model input or visible email copy.
type PendingProposal struct {
	ID      uuid.UUID
	Summary string
}

// PendingProposalPreview is the only pending-state data visible to a model.
type PendingProposalPreview struct {
	Summary string `json:"summary"`
}

// ModelRequest is the bounded, display-safe input passed to a Generator.
type ModelRequest struct {
	SafetyIdentifier string                  `json:"-"`
	Subject          string                  `json:"subject"`
	Message          string                  `json:"message"`
	Summary          string                  `json:"summary"`
	SummaryTruncated bool                    `json:"summaryTruncated"`
	History          HistoryWindow           `json:"history"`
	Facts            []GroundedFact          `json:"facts"`
	Targets          []ModelTarget           `json:"targets"`
	Choices          []ModelChoice           `json:"choices"`
	PendingProposal  *PendingProposalPreview `json:"pendingProposal"`
}

// ModelDecision is untrusted structured model output. Confirm and cancel are
// intentionally invalid here; only ParseControlCommand can create them.
type ModelDecision struct {
	Intent   Intent               `json:"intent"`
	Copy     DraftEmailCopy       `json:"copy"`
	Proposal *ModelActionProposal `json:"proposal"`
}

// GroundedSubject is a subject plus the supplied references that support it.
type GroundedSubject struct {
	Text       string   `json:"text"`
	References []string `json:"references"`
}

// DraftEmailCopy is model-produced copy. Plain text is derived from Blocks so
// the text and HTML alternatives cannot drift.
type DraftEmailCopy struct {
	Subject GroundedSubject `json:"subject"`
	Blocks  []CopyBlock     `json:"blocks"`
}

// EmailCopy contains transport-ready subject and plain text plus safe content
// primitives. RenderHTML escapes every model-authored value.
type EmailCopy struct {
	Subject   string      `json:"subject"`
	PlainText string      `json:"plainText"`
	Blocks    []CopyBlock `json:"blocks"`
}

// CopyBlockKind identifies an allowed email content primitive.
type CopyBlockKind string

const (
	CopyBlockParagraph  CopyBlockKind = "paragraph"
	CopyBlockBulletList CopyBlockKind = "bullet_list"
	CopyBlockCallout    CopyBlockKind = "callout"
)

// CopyBlock contains text only, never raw HTML or a model-authored URL.
type CopyBlock struct {
	Kind       CopyBlockKind `json:"kind"`
	Text       string        `json:"text"`
	Items      []string      `json:"items"`
	References []string      `json:"references"`
}

// Decision is the trusted output of Service.Decide. Proposal is inert and a
// Command merely identifies pending state; neither applies a mutation.
type Decision struct {
	Intent         Intent          `json:"intent"`
	Source         DecisionSource  `json:"source"`
	FallbackReason FallbackReason  `json:"fallbackReason,omitempty"`
	Copy           *EmailCopy      `json:"copy,omitempty"`
	Proposal       *ActionProposal `json:"proposal,omitempty"`
	Command        *ControlCommand `json:"command,omitempty"`
	Usage          Usage           `json:"usage"`
	fallbackCause  error
}

// DecisionSource makes model, deterministic control, and degraded fallback
// paths observable without exposing provider errors in serialized output.
type DecisionSource string

const (
	DecisionSourceModel    DecisionSource = "model"
	DecisionSourceControl  DecisionSource = "control"
	DecisionSourceFallback DecisionSource = "fallback"
)

// FallbackReason is a stable metric label for a safe deterministic fallback.
type FallbackReason string

const (
	FallbackGeneratorUnavailable FallbackReason = "generator_unavailable"
	FallbackGeneratorFailed      FallbackReason = "generator_failed"
	FallbackInvalidOutput        FallbackReason = "invalid_output"
)

// FallbackCause returns the internal generator or validation error that
// triggered a degraded decision. It is never serialized.
func (decision Decision) FallbackCause() error {
	return decision.fallbackCause
}

// ActionKind identifies the proposed domain mutation.
type ActionKind string

const (
	ActionObjectiveUpdate ActionKind = "update_objective"
	ActionKeyResultUpdate ActionKind = "update_key_result"
	ActionStoryUpdate     ActionKind = "update_story"
	ActionFeedbackStatus  ActionKind = "update_feedback_status"
)

// ModelActionProposal is the model-visible, reference-only action union.
// Exactly one action payload must match Kind.
type ModelActionProposal struct {
	Kind      ActionKind                 `json:"kind"`
	Summary   string                     `json:"summary"`
	Objective *ModelObjectiveAction      `json:"objective"`
	KeyResult *ModelKeyResultAction      `json:"keyResult"`
	Story     *ModelStoryAction          `json:"story"`
	Feedback  *ModelFeedbackStatusAction `json:"feedback"`
}

// ObjectiveHealth mirrors the objective service's persisted values without
// coupling the decision package to a concrete service implementation.
type ObjectiveHealth string

const (
	ObjectiveHealthAtRisk   ObjectiveHealth = "At Risk"
	ObjectiveHealthOnTrack  ObjectiveHealth = "On Track"
	ObjectiveHealthOffTrack ObjectiveHealth = "Off Track"
)

// ModelObjectiveAction proposes a required health change and an optional
// check-in comment. Standalone check-ins are not supported in v1.
type ModelObjectiveAction struct {
	TargetReference string           `json:"targetReference"`
	Health          *ObjectiveHealth `json:"health"`
	CheckIn         *string          `json:"checkIn"`
}

// ModelKeyResultAction proposes a required current-value change and an optional
// check-in comment. Standalone check-ins are not supported in v1.
type ModelKeyResultAction struct {
	TargetReference string   `json:"targetReference"`
	CurrentValue    *float64 `json:"currentValue"`
	CheckIn         *string  `json:"checkIn"`
}

// DateOperation describes setting or clearing a story due date.
type DateOperation string

const (
	DateSet   DateOperation = "set"
	DateClear DateOperation = "clear"
)

// ModelDateChange uses a date-only ISO value to avoid transport timezone
// ambiguity. Date is empty when Operation is clear.
type ModelDateChange struct {
	Operation DateOperation `json:"operation"`
	Date      string        `json:"date"`
}

// ModelStatusChange selects one supplied story-status choice.
type ModelStatusChange struct {
	ChoiceReference string `json:"choiceReference"`
}

// AssigneeOperation describes assigning or unassigning a story.
type AssigneeOperation string

const (
	AssigneeAssign   AssigneeOperation = "assign"
	AssigneeUnassign AssigneeOperation = "unassign"
)

// ModelAssigneeChange selects a supplied assignee for assign and uses an empty
// ChoiceReference for unassign.
type ModelAssigneeChange struct {
	Operation       AssigneeOperation `json:"operation"`
	ChoiceReference string            `json:"choiceReference"`
}

// ModelStoryAction proposes one or more bounded story field changes.
type ModelStoryAction struct {
	TargetReference string               `json:"targetReference"`
	DueDate         *ModelDateChange     `json:"dueDate"`
	Status          *ModelStatusChange   `json:"status"`
	Assignee        *ModelAssigneeChange `json:"assignee"`
}

// FeedbackStatus mirrors the feedback service's supported item states.
type FeedbackStatus string

const (
	FeedbackStatusPending    FeedbackStatus = "pending"
	FeedbackStatusReviewing  FeedbackStatus = "reviewing"
	FeedbackStatusPlanned    FeedbackStatus = "planned"
	FeedbackStatusInProgress FeedbackStatus = "in_progress"
	FeedbackStatusCompleted  FeedbackStatus = "completed"
	FeedbackStatusClosed     FeedbackStatus = "closed"
)

// ModelFeedbackStatusAction proposes one status change for one supplied item.
// Multi-item writes are deliberately split into separate confirmations in v1.
type ModelFeedbackStatusAction struct {
	TargetReference string         `json:"targetReference"`
	Status          FeedbackStatus `json:"status"`
}

// ActionProposal is durable, actor/workspace-bound, and inert. A future
// confirmer must re-read, reauthorize, and compare ExpectedUpdatedAt before
// invoking domain services.
type ActionProposal struct {
	WorkspaceID uuid.UUID             `json:"workspaceId"`
	ActorID     uuid.UUID             `json:"actorId"`
	Kind        ActionKind            `json:"kind"`
	Summary     string                `json:"summary"`
	Objective   *ObjectiveAction      `json:"objective,omitempty"`
	KeyResult   *KeyResultAction      `json:"keyResult,omitempty"`
	Story       *StoryAction          `json:"story,omitempty"`
	Feedback    *FeedbackStatusAction `json:"feedback,omitempty"`
}

// TargetSnapshot is the authoritative target identity captured at proposal
// time. DisplayName is safe for confirmation copy; IDs are never model-made.
type TargetSnapshot struct {
	ID                uuid.UUID `json:"id"`
	TeamID            uuid.UUID `json:"teamId"`
	DisplayName       string    `json:"displayName"`
	ExpectedUpdatedAt time.Time `json:"expectedUpdatedAt"`
}

type ObjectiveAction struct {
	Target  TargetSnapshot   `json:"target"`
	Health  *ObjectiveHealth `json:"health,omitempty"`
	CheckIn *string          `json:"checkIn,omitempty"`
}

type KeyResultAction struct {
	Target       TargetSnapshot `json:"target"`
	CurrentValue *float64       `json:"currentValue,omitempty"`
	CheckIn      *string        `json:"checkIn,omitempty"`
}

type DateChange struct {
	Operation DateOperation `json:"operation"`
	Date      string        `json:"date,omitempty"`
}

type StatusChange struct {
	StatusID   uuid.UUID `json:"statusId"`
	StatusName string    `json:"statusName"`
}

type AssigneeChange struct {
	Operation    AssigneeOperation `json:"operation"`
	AssigneeID   *uuid.UUID        `json:"assigneeId,omitempty"`
	AssigneeName string            `json:"assigneeName,omitempty"`
}

type StoryAction struct {
	Target   TargetSnapshot  `json:"target"`
	DueDate  *DateChange     `json:"dueDate,omitempty"`
	Status   *StatusChange   `json:"status,omitempty"`
	Assignee *AssigneeChange `json:"assignee,omitempty"`
}

type FeedbackStatusAction struct {
	Target TargetSnapshot `json:"target"`
	Status FeedbackStatus `json:"status"`
}

// ControlKind is an exact explicit command parsed without model judgment.
type ControlKind string

const (
	ControlConfirm ControlKind = "confirm"
	ControlCancel  ControlKind = "cancel"
)

// ControlCommand binds an exact reply to one durable pending proposal. The ID
// is intentionally unexported and excluded from JSON so it cannot leak into a
// user-visible payload; ProposalID is for trusted server wiring only.
type ControlCommand struct {
	Kind       ControlKind `json:"kind"`
	proposalID uuid.UUID
}

// ProposalID returns the internal pending-proposal ID selected by the service.
func (command ControlCommand) ProposalID() uuid.UUID {
	return command.proposalID
}
