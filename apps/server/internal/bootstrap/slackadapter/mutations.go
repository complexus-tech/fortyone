package slackadapter

import (
	"context"
	"errors"

	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	"github.com/google/uuid"
)

type MutationConfirmer struct {
	backend messaging.StoryMutationConfirmer
}

func NewMutationConfirmer(backend messaging.StoryMutationConfirmer) *MutationConfirmer {
	if backend == nil {
		return nil
	}
	return &MutationConfirmer{backend: backend}
}

func (adapter *MutationConfirmer) ConfirmStoryMutation(ctx context.Context, scope slack.StoryMutationScope, token string) (slack.StoryMutationResult, error) {
	result, err := adapter.backend.ConfirmStoryMutation(ctx, mapMutationScope(scope), token)
	return mapMutationResult(result), mapMutationError(err)
}

func (adapter *MutationConfirmer) CancelStoryMutation(ctx context.Context, scope slack.StoryMutationScope, token string) (slack.StoryMutationCancellationResult, error) {
	result, err := adapter.backend.CancelStoryMutation(ctx, mapMutationScope(scope), token)
	return slack.StoryMutationCancellationResult{Status: result.Status}, mapMutationError(err)
}

func mapMutationScope(scope slack.StoryMutationScope) messaging.ToolScope {
	return messaging.ToolScope{
		WorkspaceID: scope.WorkspaceID, UserID: scope.UserID,
		AllowedTeamIDs: append([]uuid.UUID(nil), scope.AllowedTeamIDs...),
		SharedTeamIDs:  append([]uuid.UUID(nil), scope.SharedTeamIDs...),
		AllowMutations: scope.AllowMutations, WebsiteURL: scope.WebsiteURL,
		SourceURL: scope.SourceURL, WorkspaceSlug: scope.WorkspaceSlug, Timezone: scope.Timezone,
	}
}

func mapMutationResult(result messaging.StoryMutationResult) slack.StoryMutationResult {
	mapped := slack.StoryMutationResult{
		Status: result.Status, Operation: slack.StoryMutationOperation(result.Operation),
		StoryID: result.StoryID, Reference: result.Reference, TeamID: result.TeamID, Title: result.Title,
		Priority: result.Priority, AssigneeID: result.AssigneeID,
		EstimatedDurationMinutes: result.EstimatedDurationMinutes,
		MinimumFocusBlockMinutes: result.MinimumFocusBlockMinutes,
		AutoSchedulingEnabled:    result.AutoSchedulingEnabled, AutoSchedulingLocked: result.AutoSchedulingLocked,
		AutoSchedulingStatus: result.AutoSchedulingStatus, AutoSchedulingReason: result.AutoSchedulingReason,
		AutoSchedulingUpdatedAt: result.AutoSchedulingUpdatedAt,
		CommentID:               result.CommentID, AssociationID: result.AssociationID,
	}
	if result.Items != nil {
		mapped.Items = make([]slack.StoryMutationItemResult, 0, len(result.Items))
		for _, item := range result.Items {
			mapped.Items = append(mapped.Items, slack.StoryMutationItemResult{
				Index: item.Index, Status: item.Status, StoryID: item.StoryID,
				Reference: item.Reference, TeamID: item.TeamID, Title: item.Title,
				Priority: item.Priority, AssigneeID: item.AssigneeID,
				EstimatedDurationMinutes: item.EstimatedDurationMinutes,
				MinimumFocusBlockMinutes: item.MinimumFocusBlockMinutes,
				AutoSchedulingEnabled:    item.AutoSchedulingEnabled,
				AutoSchedulingLocked:     item.AutoSchedulingLocked,
				AutoSchedulingStatus:     item.AutoSchedulingStatus,
				AutoSchedulingReason:     item.AutoSchedulingReason,
				AutoSchedulingUpdatedAt:  item.AutoSchedulingUpdatedAt,
			})
		}
	}
	return mapped
}

func mapMutationError(err error) error {
	switch {
	case errors.Is(err, messaging.ErrAppliedConfirmation):
		return errors.Join(slack.ErrStoryMutationApplied, err)
	case errors.Is(err, messaging.ErrCancelledConfirmation):
		return errors.Join(slack.ErrStoryMutationCancelled, err)
	case errors.Is(err, messaging.ErrExpiredConfirmation):
		return errors.Join(slack.ErrStoryMutationExpired, err)
	case errors.Is(err, messaging.ErrInvalidConfirmation):
		return errors.Join(slack.ErrStoryMutationInvalid, err)
	case errors.Is(err, messaging.ErrMutationNotAllowed):
		return errors.Join(slack.ErrStoryMutationNotAllowed, err)
	case errors.Is(err, messaging.ErrStaleMutation):
		return errors.Join(slack.ErrStoryMutationStale, err)
	case errors.Is(err, messaging.ErrTeamNotAccessible):
		return errors.Join(slack.ErrStoryMutationTeamRestricted, err)
	default:
		return err
	}
}

var _ slack.StoryMutationConfirmer = (*MutationConfirmer)(nil)
