package emailreply

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AgentIntent string

const (
	AgentIntentAnswer  AgentIntent = "answer"
	AgentIntentClarify AgentIntent = "clarify"
	AgentIntentPropose AgentIntent = "propose"
	AgentIntentCancel  AgentIntent = "cancel"
	AgentIntentConfirm AgentIntent = "confirm"
	AgentIntentRefuse  AgentIntent = "refuse"
)

type AgentDecisionSource string

const AgentDecisionSourceFallback AgentDecisionSource = "fallback"

type ConversationRole string

const (
	ConversationRoleUser      ConversationRole = "user"
	ConversationRoleAssistant ConversationRole = "assistant"
)

type HistoryTurn struct {
	Role   ConversationRole `json:"role"`
	Text   string           `json:"text"`
	SentAt time.Time        `json:"sentAt,omitempty"`
}

type GroundedFact struct {
	Reference       string   `json:"reference"`
	Text            string   `json:"text"`
	ProtectedTokens []string `json:"protectedTokens"`
}

type TargetKind string

const (
	TargetObjective TargetKind = "objective"
	TargetKeyResult TargetKind = "key_result"
	TargetStory     TargetKind = "story"
	TargetFeedback  TargetKind = "feedback"
)

type AuthorizedTarget struct {
	Reference         string
	Kind              TargetKind
	DisplayName       string
	CurrentState      string
	ID                uuid.UUID
	TeamID            uuid.UUID
	ExpectedUpdatedAt time.Time
}

type ChoiceKind string

const (
	ChoiceStoryStatus   ChoiceKind = "story_status"
	ChoiceStoryAssignee ChoiceKind = "story_assignee"
)

type AuthorizedChoice struct {
	Reference   string
	Kind        ChoiceKind
	DisplayName string
	ID          uuid.UUID
	TeamID      uuid.UUID
}

type PendingProposal struct {
	ID      uuid.UUID
	Summary string
}

type AgentRequest struct {
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

type CopyBlockKind string

const (
	CopyBlockParagraph  CopyBlockKind = "paragraph"
	CopyBlockBulletList CopyBlockKind = "bullet_list"
	CopyBlockCallout    CopyBlockKind = "callout"
)

type CopyBlock struct {
	Kind       CopyBlockKind `json:"kind"`
	Text       string        `json:"text"`
	Items      []string      `json:"items"`
	References []string      `json:"references"`
}

type EmailCopy struct {
	Subject   string      `json:"subject"`
	PlainText string      `json:"plainText"`
	Blocks    []CopyBlock `json:"blocks"`
}

type ActionKind string

const (
	ActionObjectiveUpdate ActionKind = "update_objective"
	ActionKeyResultUpdate ActionKind = "update_key_result"
	ActionStoryUpdate     ActionKind = "update_story"
	ActionFeedbackStatus  ActionKind = "update_feedback_status"
)

type TargetSnapshot struct {
	ID                uuid.UUID `json:"id"`
	TeamID            uuid.UUID `json:"teamId"`
	DisplayName       string    `json:"displayName"`
	ExpectedUpdatedAt time.Time `json:"expectedUpdatedAt"`
}

type ObjectiveHealth string

const (
	ObjectiveHealthAtRisk  ObjectiveHealth = "At Risk"
	ObjectiveHealthOnTrack ObjectiveHealth = "On Track"
)

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

type DateOperation string

const (
	DateSet   DateOperation = "set"
	DateClear DateOperation = "clear"
)

type DateChange struct {
	Operation DateOperation `json:"operation"`
	Date      string        `json:"date,omitempty"`
}

type StatusChange struct {
	StatusID   uuid.UUID `json:"statusId"`
	StatusName string    `json:"statusName"`
}

type AssigneeOperation string

const (
	AssigneeAssign   AssigneeOperation = "assign"
	AssigneeUnassign AssigneeOperation = "unassign"
)

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

type FeedbackStatus string

type FeedbackStatusAction struct {
	Target TargetSnapshot `json:"target"`
	Status FeedbackStatus `json:"status"`
}

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

type ControlCommand struct {
	ProposalID uuid.UUID
}

type AgentDecision struct {
	Intent   AgentIntent
	Source   AgentDecisionSource
	Copy     *EmailCopy
	Proposal *ActionProposal
	Command  *ControlCommand
}

type SummaryRequest struct {
	SafetyIdentifier string
	PreviousSummary  string
	OmittedTurns     []HistoryTurn
}

type SummaryGeneration struct {
	Summary string
}

type DecisionPort interface {
	Decide(context.Context, AgentRequest) (AgentDecision, error)
}

type SummaryPort interface {
	Summarize(context.Context, SummaryRequest) (SummaryGeneration, error)
}

type CopyRenderer interface {
	RenderHTML(EmailCopy) (string, error)
}

type ControlKind string

const (
	ControlConfirm ControlKind = "confirm"
	ControlCancel  ControlKind = "cancel"
)

func parseControlCommand(currentReply string) (ControlKind, bool) {
	normalized := strings.TrimSpace(currentReply)
	switch {
	case strings.EqualFold(normalized, "CONFIRM"):
		return ControlConfirm, true
	case strings.EqualFold(normalized, "CANCEL"):
		return ControlCancel, true
	default:
		return "", false
	}
}

func renderPlainText(blocks []CopyBlock) string {
	sections := make([]string, 0, len(blocks))
	for _, block := range blocks {
		var section strings.Builder
		if block.Text != "" {
			section.WriteString(block.Text)
		}
		if block.Kind == CopyBlockBulletList {
			for _, item := range block.Items {
				if section.Len() > 0 {
					section.WriteByte('\n')
				}
				section.WriteString("- ")
				section.WriteString(item)
			}
		}
		if section.Len() > 0 {
			sections = append(sections, section.String())
		}
	}
	return strings.Join(sections, "\n\n")
}
