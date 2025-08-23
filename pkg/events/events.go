package events

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	StoryCreated       EventType = "story.created"
	StoryUpdated       EventType = "story.updated"
	StoryDuplicated    EventType = "story.duplicated"
	CommentCreated     EventType = "comment.created"
	CommentReplied     EventType = "comment.replied"
	UserMentioned      EventType = "user.mentioned"
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

// StoryCreatedPayload contains data for story creation events
type StoryCreatedPayload struct {
	StoryID     uuid.UUID  `json:"story_id"`
	WorkspaceID uuid.UUID  `json:"workspace_id"`
	Title       string     `json:"title"`
	AssigneeID  *uuid.UUID `json:"assignee_id,omitempty"`
	ReporterID  uuid.UUID  `json:"reporter_id"`
}

// StoryUpdatedPayload contains data for story update events
type StoryUpdatedPayload struct {
	StoryID     uuid.UUID      `json:"story_id"`
	WorkspaceID uuid.UUID      `json:"workspace_id"`
	Updates     map[string]any `json:"updates"`
	AssigneeID  *uuid.UUID     `json:"assignee_id,omitempty"`
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

// CommentCreatedPayload contains data for comment creation events
type CommentCreatedPayload struct {
	CommentID   uuid.UUID   `json:"comment_id"`
	StoryID     uuid.UUID   `json:"story_id"`
	StoryTitle  string      `json:"story_title"`
	AssigneeID  *uuid.UUID  `json:"assignee_id"`
	WorkspaceID uuid.UUID   `json:"workspace_id"`
	Content     string      `json:"content"`
	Mentions    []uuid.UUID `json:"mentions"`
}

// CommentRepliedPayload contains data for comment reply events
type CommentRepliedPayload struct {
	CommentID       uuid.UUID   `json:"comment_id"`
	ParentCommentID uuid.UUID   `json:"parent_comment_id"`
	ParentAuthorID  uuid.UUID   `json:"parent_author_id"`
	StoryID         uuid.UUID   `json:"story_id"`
	StoryTitle      string      `json:"story_title"`
	WorkspaceID     uuid.UUID   `json:"workspace_id"`
	Content         string      `json:"content"`
	Mentions        []uuid.UUID `json:"mentions"`
}

// UserMentionedPayload contains data for user mention events
type UserMentionedPayload struct {
	CommentID     uuid.UUID `json:"comment_id"`
	StoryID       uuid.UUID `json:"story_id"`
	StoryTitle    string    `json:"story_title"`
	WorkspaceID   uuid.UUID `json:"workspace_id"`
	MentionedUser uuid.UUID `json:"mentioned_user"`
	Content       string    `json:"content"`
}
