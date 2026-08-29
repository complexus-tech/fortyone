package stories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type storyActivityWriter interface {
	RecordStoryActivities(context.Context, storydomain.RecordActivitiesCommand) ([]storydomain.MutationActivity, error)
}

func (s *Service) recordActivities(ctx context.Context, activities []CoreActivity) error {
	if len(activities) == 0 {
		return nil
	}
	repository, typed := s.repo.(storyActivityWriter)
	if !typed {
		legacy, ok := s.repo.(legacyActivityWriter)
		if !ok {
			return errors.New("story repository does not support activity recording")
		}
		_, err := legacy.RecordActivities(ctx, activities)
		return err
	}

	workspaceID := activities[0].WorkspaceID
	actorID := activities[0].UserID
	scope, err := mutationScope(ctx, workspaceID, actorID, platformauth.PrincipalHumanUser)
	if err != nil {
		return mapStoryMutationError(err)
	}
	writes := make([]storydomain.ActivityWrite, 0, len(activities))
	for _, activity := range activities {
		if activity.WorkspaceID != workspaceID || activity.UserID != actorID {
			return fmt.Errorf("%w: activity batch scope mismatch", ErrInvalidStoryMutation)
		}
		oldValue, err := json.Marshal(activity.OldValue)
		if err != nil {
			return fmt.Errorf("encode old story activity value: %w", err)
		}
		newValue, err := json.Marshal(activity.NewValue)
		if err != nil {
			return fmt.Errorf("encode new story activity value: %w", err)
		}
		createdAt := activity.CreatedAt.UTC()
		if createdAt.IsZero() {
			createdAt = time.Now().UTC()
		}
		compact := activity.ID == uuid.Nil && activity.Type == "update" && strings.TrimSpace(activity.Field) != ""
		activityID := activity.ID
		if activityID == uuid.Nil {
			activityID = uuid.New()
		}
		writes = append(writes, storydomain.ActivityWrite{
			Compact: compact,
			Activity: storydomain.MutationActivity{
				ID: activityID, StoryID: activity.StoryID, UserID: activity.UserID,
				Type: activity.Type, Field: activity.Field, CurrentValue: activity.CurrentValue,
				OldValue: oldValue, NewValue: newValue, Reason: activity.Reason,
				WorkspaceID: workspaceID, CreatedAt: createdAt,
			},
		})
	}
	_, err = repository.RecordStoryActivities(ctx, storydomain.RecordActivitiesCommand{Scope: scope, Activities: writes})
	return mapStoryMutationError(err)
}
