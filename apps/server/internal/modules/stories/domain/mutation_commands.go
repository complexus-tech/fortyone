package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type MutationScope struct {
	Actor        platformauth.Actor
	WorkspaceID  uuid.UUID
	ActivityUser *uuid.UUID
}

type MutationKeyResultReference struct {
	ObjectiveID uuid.UUID
	Name        string
}

type MutationPreconditions struct {
	EstimateScheme string
	KeyResult      *MutationKeyResultReference
}

func (scope MutationScope) Validate() error {
	if scope.WorkspaceID == uuid.Nil || scope.Actor.WorkspaceID != scope.WorkspaceID {
		return fmt.Errorf("%w: workspace mismatch", ErrMutationForbidden)
	}
	if err := scope.Actor.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrMutationForbidden, err)
	}
	if !scope.Actor.Scopes.Has(platformauth.ScopeStoriesWrite) {
		return fmt.Errorf("%w: stories:write scope is required", ErrMutationForbidden)
	}
	requiresUserActivity := scope.Actor.IsUserActor() || scope.Actor.Kind == platformauth.PrincipalSystem
	if requiresUserActivity && (scope.ActivityUser == nil || *scope.ActivityUser != scope.Actor.PrincipalID) {
		return fmt.Errorf("%w: activity user must match the acting user", ErrMutationForbidden)
	}
	if !requiresUserActivity && scope.ActivityUser != nil {
		return fmt.Errorf("%w: non-user actors cannot be attributed to a user", ErrMutationForbidden)
	}
	return nil
}

type MutationEventType string

const (
	MutationEventStoryCreated MutationEventType = "story.created"
	MutationEventStoryUpdated MutationEventType = "story.updated"
	MutationEventStoryDeleted MutationEventType = "story.deleted"
)

func (eventType MutationEventType) Validate() error {
	switch eventType {
	case MutationEventStoryCreated, MutationEventStoryUpdated, MutationEventStoryDeleted:
		return nil
	default:
		return fmt.Errorf("%w: unsupported event type", ErrInvalidMutation)
	}
}

type MutationEvent struct {
	ID            uuid.UUID
	WorkspaceID   uuid.UUID
	StoryID       uuid.UUID
	Type          MutationEventType
	Actor         platformauth.Actor
	Payload       json.RawMessage
	OccurredAt    time.Time
	AttemptCount  int
	ClaimToken    uuid.UUID
	NextAttemptAt time.Time
	LastError     string
	CreatedAt     time.Time
	ClaimedAt     *time.Time
	CompletedAt   *time.Time
}

func (event MutationEvent) Validate() error {
	if event.ID == uuid.Nil || event.WorkspaceID == uuid.Nil || event.StoryID == uuid.Nil || event.OccurredAt.IsZero() {
		return fmt.Errorf("%w: event identity is incomplete", ErrInvalidMutation)
	}
	if err := event.Type.Validate(); err != nil {
		return err
	}
	if event.Actor.WorkspaceID != event.WorkspaceID {
		return fmt.Errorf("%w: event actor workspace mismatch", ErrInvalidMutation)
	}
	if err := event.Actor.Validate(); err != nil {
		return fmt.Errorf("%w: invalid event actor", ErrInvalidMutation)
	}
	if len(event.Payload) < 2 || len(event.Payload) > 256<<10 || !json.Valid(event.Payload) {
		return fmt.Errorf("%w: event payload must be a bounded JSON object", ErrInvalidMutation)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(event.Payload, &object); err != nil || object == nil {
		return fmt.Errorf("%w: event payload must be an object", ErrInvalidMutation)
	}
	return nil
}

type MutationActivity struct {
	ID           uuid.UUID
	StoryID      uuid.UUID
	UserID       uuid.UUID
	Type         string
	Field        string
	CurrentValue string
	OldValue     json.RawMessage
	NewValue     json.RawMessage
	Reason       *string
	WorkspaceID  uuid.UUID
	CreatedAt    time.Time
}

func (activity MutationActivity) Validate(scope MutationScope, storyID uuid.UUID) error {
	if activity.ID == uuid.Nil || activity.StoryID != storyID || activity.WorkspaceID != scope.WorkspaceID ||
		activity.UserID == uuid.Nil || activity.CreatedAt.IsZero() {
		return fmt.Errorf("%w: activity identity and scope are required", ErrInvalidMutation)
	}
	if scope.ActivityUser == nil || activity.UserID != *scope.ActivityUser {
		return fmt.Errorf("%w: activity user does not match the mutation actor", ErrInvalidMutation)
	}
	if strings.TrimSpace(activity.Type) == "" || strings.TrimSpace(activity.Field) == "" {
		return fmt.Errorf("%w: activity type and field are required", ErrInvalidMutation)
	}
	if !json.Valid(activity.OldValue) || !json.Valid(activity.NewValue) {
		return fmt.Errorf("%w: activity values must be valid JSON", ErrInvalidMutation)
	}
	return nil
}

type CreateStoryCommand struct {
	Scope    MutationScope
	Story    Story
	LabelIDs []uuid.UUID
	Event    MutationEvent
	Activity *MutationActivity
}

func (command CreateStoryCommand) Validate() error {
	if err := command.Scope.Validate(); err != nil {
		return err
	}
	if command.Story.ID == uuid.Nil || command.Story.Workspace != command.Scope.WorkspaceID || command.Story.Team == uuid.Nil {
		return fmt.Errorf("%w: story identity and scope are required", ErrInvalidMutation)
	}
	if strings.TrimSpace(command.Story.Title) == "" {
		return fmt.Errorf("%w: title cannot be blank", ErrInvalidMutation)
	}
	if command.Event.StoryID != command.Story.ID || command.Event.WorkspaceID != command.Scope.WorkspaceID || command.Event.Type != MutationEventStoryCreated {
		return fmt.Errorf("%w: create event scope does not match story", ErrInvalidMutation)
	}
	if command.Activity != nil {
		if err := command.Activity.Validate(command.Scope, command.Story.ID); err != nil {
			return err
		}
	}
	return command.Event.Validate()
}

type CreateStoryResult struct {
	Story   Story
	Created bool
}

type UpdateStoryCommand struct {
	Scope             MutationScope
	StoryID           uuid.UUID
	ExpectedUpdatedAt time.Time
	Patch             StoryPatch
	Event             MutationEvent
	Activities        []MutationActivity
	ReferencedMedia   []uuid.UUID
	ReconcileMedia    bool
}

func (command UpdateStoryCommand) Validate() error {
	if err := command.Scope.Validate(); err != nil {
		return err
	}
	if command.StoryID == uuid.Nil || command.ExpectedUpdatedAt.IsZero() {
		return fmt.Errorf("%w: story identity and expected version are required", ErrInvalidMutation)
	}
	if command.Patch.Empty() {
		if !command.ReconcileMedia {
			return fmt.Errorf("%w: at least one field is required", ErrInvalidMutation)
		}
	} else if err := command.Patch.Validate(); err != nil {
		return err
	}
	if command.Event.StoryID != command.StoryID || command.Event.WorkspaceID != command.Scope.WorkspaceID || command.Event.Type != MutationEventStoryUpdated {
		return fmt.Errorf("%w: event scope does not match mutation", ErrInvalidMutation)
	}
	if err := command.Event.Validate(); err != nil {
		return err
	}
	for _, activity := range command.Activities {
		if err := activity.Validate(command.Scope, command.StoryID); err != nil {
			return err
		}
	}
	return nil
}

type UpdateStoryResult struct {
	UpdatedAt             time.Time
	OrphanedAttachmentIDs []uuid.UUID
}

type DeleteStoryCommand struct {
	Scope             MutationScope
	StoryID           uuid.UUID
	ExpectedUpdatedAt time.Time
	Event             MutationEvent
	Activity          *MutationActivity
}

func (command DeleteStoryCommand) Validate() error {
	if err := command.Scope.Validate(); err != nil {
		return err
	}
	if command.StoryID == uuid.Nil || command.ExpectedUpdatedAt.IsZero() {
		return fmt.Errorf("%w: story identity and expected version are required", ErrInvalidMutation)
	}
	if command.Event.StoryID != command.StoryID || command.Event.WorkspaceID != command.Scope.WorkspaceID || command.Event.Type != MutationEventStoryDeleted {
		return fmt.Errorf("%w: delete event scope does not match story", ErrInvalidMutation)
	}
	if command.Activity != nil {
		if err := command.Activity.Validate(command.Scope, command.StoryID); err != nil {
			return err
		}
	}
	return command.Event.Validate()
}

type DeleteStoryResult struct {
	Deleted   bool
	DeletedAt *time.Time
}
