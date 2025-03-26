package events

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	StoryUpdated       EventType = "story.updated"
	StoryCommented     EventType = "story.commented"
	StoryDuplicated    EventType = "story.duplicated"
	ObjectiveUpdated   EventType = "objective.updated"
	KeyResultUpdated   EventType = "keyresult.updated"
	EmailVerification  EventType = "email.verification"
	InvitationEmail    EventType = "invitation.email"
	InvitationAccepted EventType = "invitation.accepted"
)

// Event is the base event structure
type Event struct {
	Type      EventType `json:"type"`
	Payload   any       `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
	ActorID   uuid.UUID `json:"actor_id"`
}

// StoryUpdatedPayload contains data for story update events
type StoryUpdatedPayload struct {
	StoryID     uuid.UUID      `json:"story_id"`
	WorkspaceID uuid.UUID      `json:"workspace_id"`
	Updates     map[string]any `json:"updates"`
	AssigneeID  *uuid.UUID     `json:"assignee_id,omitempty"`
}

// StoryCommentedPayload contains data for story comment events
type StoryCommentedPayload struct {
	StoryID     uuid.UUID  `json:"story_id"`
	WorkspaceID uuid.UUID  `json:"workspace_id"`
	CommentID   uuid.UUID  `json:"comment_id"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
}

// ObjectiveUpdatedPayload contains data for objective update events
type ObjectiveUpdatedPayload struct {
	ObjectiveID uuid.UUID      `json:"objective_id"`
	WorkspaceID uuid.UUID      `json:"workspace_id"`
	Updates     map[string]any `json:"updates"`
	LeadID      *uuid.UUID     `json:"lead_id,omitempty"`
}

// KeyResultUpdatedPayload contains data for key result update events
type KeyResultUpdatedPayload struct {
	KeyResultID uuid.UUID      `json:"key_result_id"`
	ObjectiveID uuid.UUID      `json:"objective_id"`
	WorkspaceID uuid.UUID      `json:"workspace_id"`
	Updates     map[string]any `json:"updates"`
}

// EmailVerificationPayload contains data for email verification events
type EmailVerificationPayload struct {
	Email     string `json:"email"`
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
}

// InvitationEmailPayload contains data for invitation email events
type InvitationEmailPayload struct {
	InviterName   string    `json:"inviter_name"`
	Email         string    `json:"email"`
	Token         string    `json:"token"`
	Role          string    `json:"role"`
	ExpiresAt     time.Time `json:"expires_at"`
	WorkspaceID   uuid.UUID `json:"workspace_id"`
	WorkspaceName string    `json:"workspace_name"`
}

// InvitationAcceptedPayload contains data for invitation acceptance notification events
type InvitationAcceptedPayload struct {
	InviterEmail  string    `json:"inviter_email"`
	InviterName   string    `json:"inviter_name"`
	InviteeName   string    `json:"invitee_name"`
	InviteeEmail  string    `json:"invitee_email"`
	Role          string    `json:"role"`
	WorkspaceID   uuid.UUID `json:"workspace_id"`
	WorkspaceName string    `json:"workspace_name"`
	WorkspaceSlug string    `json:"workspace_slug"`
}

// StoryDuplicatedPayload contains data for story duplication events
type StoryDuplicatedPayload struct {
	StoryID         uuid.UUID `json:"story_id"`
	OriginalStoryID uuid.UUID `json:"original_story_id"`
	WorkspaceID     uuid.UUID `json:"workspace_id"`
}
