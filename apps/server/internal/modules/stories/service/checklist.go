package stories

import (
	"context"
	"fmt"
	"strings"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

type checklistRepository interface {
	ListChecklist(context.Context, storydomain.ReadScope, uuid.UUID) ([]storydomain.ChecklistItem, error)
	ApplyChecklistMutation(context.Context, storydomain.ChecklistCommand) (storydomain.ChecklistMutationResult, error)
}

func (s *Service) ListChecklist(ctx context.Context, workspaceID, storyID uuid.UUID) ([]storydomain.ChecklistItem, error) {
	scope, err := readScopeFromContext(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	repository, ok := s.repo.(checklistRepository)
	if !ok {
		return nil, fmt.Errorf("story repository does not support checklists")
	}
	return repository.ListChecklist(ctx, scope, storyID)
}

// MutateChecklist derives identity and event metadata from the authenticated
// context. Callers supply only the owning story and requested item change.
func (s *Service) MutateChecklist(ctx context.Context, workspaceID uuid.UUID, command storydomain.ChecklistCommand) (storydomain.ChecklistMutationResult, uuid.UUID, error) {
	scope, err := mutationScope(ctx, workspaceID, uuid.Nil, auth.PrincipalHumanUser)
	if err != nil {
		return storydomain.ChecklistMutationResult{}, uuid.Nil, mapStoryMutationError(err)
	}
	repository, ok := s.repo.(checklistRepository)
	if !ok {
		return storydomain.ChecklistMutationResult{}, uuid.Nil, fmt.Errorf("story repository does not support checklists")
	}
	command.Scope = scope
	if command.Action == storydomain.ChecklistAdd {
		command.ItemID = uuid.New()
	}
	if command.Title != nil {
		title := strings.TrimSpace(*command.Title)
		command.Title = &title
	}
	now := time.Now().UTC()
	command.Event, err = newStoryMutationEvent(scope, command.StoryID, storydomain.MutationEventStoryUpdated, storyUpdatedIntegrationPayload{
		StoryID: command.StoryID, WorkspaceID: workspaceID, Changes: map[string]any{"checklist_changed": true},
	}, now)
	if err != nil {
		return storydomain.ChecklistMutationResult{}, uuid.Nil, err
	}
	if err := command.Validate(); err != nil {
		return storydomain.ChecklistMutationResult{}, uuid.Nil, mapStoryMutationError(err)
	}
	result, err := repository.ApplyChecklistMutation(ctx, command)
	if err != nil {
		return storydomain.ChecklistMutationResult{}, uuid.Nil, mapStoryMutationError(err)
	}
	if result.Changed && s.publisher != nil {
		s.publishChecklistEvent(ctx, events.Event{Type: events.StoryUpdated, ActorID: scope.Actor.PrincipalID, Timestamp: now, Payload: events.StoryUpdatedPayload{
			StoryID: command.StoryID, WorkspaceID: workspaceID, Updates: map[string]any{"checklist_changed": true},
		}})
		if created := result.CreatedStory; created != nil {
			s.publishChecklistEvent(ctx, events.Event{Type: events.StoryCreated, ActorID: scope.Actor.PrincipalID, Timestamp: now, Payload: events.StoryCreatedPayload{
				StoryID: created.ID, WorkspaceID: workspaceID, Title: created.Title, ReporterID: scope.Actor.PrincipalID,
			}})
		}
	}
	if result.CreatedStory != nil {
		s.enqueueGitHubStorySync(ctx, result.CreatedStory.ID, workspaceID)
	}
	return result, command.ItemID, nil
}

func (s *Service) publishChecklistEvent(ctx context.Context, event events.Event) {
	if err := s.publisher.Publish(context.WithoutCancel(ctx), event); err != nil && s.log != nil {
		s.log.Error(ctx, "failed to publish checklist event", "error", err, "event_type", event.Type)
	}
}
