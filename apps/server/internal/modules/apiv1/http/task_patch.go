package apiv1http

import (
	"context"
	"fmt"
	"net/http"

	openapiv1 "github.com/complexus-tech/projects-api/internal/generated/openapi/v1"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
)

func (s *server) UpdateStory(ctx context.Context, request openapiv1.UpdateStoryRequestObject) (openapiv1.UpdateStoryResponseObject, error) {
	actor, problem := actorFor(ctx, request.WorkspaceId, platformauth.ScopeStoriesWrite)
	if problem == nil {
		problem = requireStoryWriter(actor)
	}
	if problem != nil {
		return mutationFailure(ctx, problem), nil
	}
	if request.Body == nil {
		return mutationFailure(ctx, statusFailure(http.StatusBadRequest)), nil
	}
	patch, err := publicStoryPatch(*request.Body)
	if err != nil {
		return mutationFailure(ctx, statusFailure(http.StatusBadRequest)), nil
	}
	if s.storyPatches == nil {
		return mutationFailure(ctx, statusFailure(http.StatusServiceUnavailable)), nil
	}
	if err := s.storyPatches.UpdatePatchIfUnchanged(ctx, request.StoryId, request.WorkspaceId, request.Body.ExpectedUpdatedAt, patch); err != nil {
		return mutationFailure(ctx, classifyFailure(err)), nil
	}
	return openapiv1.UpdateStory204Response{}, nil
}

func publicStoryPatch(body openapiv1.ComponentsResourcesUpdateStoryRequest) (stories.StoryPatch, error) {
	patch := stories.StoryPatch{
		Title: optionalPatchField(body.Title), Description: optionalPatchField(body.Description),
		StatusID: optionalPatchField(body.StatusId), AssigneeID: optionalPatchField(body.AssigneeId),
		SprintID: optionalPatchField(body.SprintId), ObjectiveID: optionalPatchField(body.ObjectiveId),
		ParentID: optionalPatchField(body.ParentId), KeyResultID: optionalPatchField(body.KeyResultId),
		StartDate: optionalPatchField(body.StartDate), EndDate: optionalPatchField(body.EndDate),
		AutoSchedulingEnabled: optionalPatchField(body.AutoSchedulingEnabled),
	}
	if body.Priority != nil {
		patch.Priority = stories.SetField(string(*body.Priority))
	}
	if body.ClearFields != nil {
		for _, field := range *body.ClearFields {
			var err error
			switch field {
			case "description":
				err = clearPublicField(&patch.Description)
			case "assigneeId":
				err = clearPublicField(&patch.AssigneeID)
			case "sprintId":
				err = clearPublicField(&patch.SprintID)
			case "objectiveId":
				err = clearPublicField(&patch.ObjectiveID)
			case "parentId":
				err = clearPublicField(&patch.ParentID)
			case "keyResultId":
				err = clearPublicField(&patch.KeyResultID)
			case "startDate":
				err = clearPublicField(&patch.StartDate)
			case "endDate":
				err = clearPublicField(&patch.EndDate)
			default:
				return stories.StoryPatch{}, fmt.Errorf("%w: unsupported clear field", stories.ErrInvalidStoryMutation)
			}
			if err != nil {
				return stories.StoryPatch{}, err
			}
		}
	}
	if body.ExpectedUpdatedAt.IsZero() || len(patch.Fields()) == 0 {
		return stories.StoryPatch{}, stories.ErrInvalidStoryMutation
	}
	return patch, patch.Validate()
}

func optionalPatchField[T any](value *T) stories.Field[T] {
	if value == nil {
		return stories.Field[T]{}
	}
	return stories.SetField(*value)
}

func clearPublicField[T any](field *stories.Field[T]) error {
	if field.Specified() {
		return fmt.Errorf("%w: a field cannot be set and cleared", stories.ErrInvalidStoryMutation)
	}
	*field = stories.ClearField[T]()
	return nil
}
