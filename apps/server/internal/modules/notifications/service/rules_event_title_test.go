package notifications

import (
	"context"
	"errors"
	"testing"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type eventTitleStories struct {
	scheduleRulesStories
	actorID, storyID, workspaceID uuid.UUID
	calls                         int
	err                           error
}

func (s *eventTitleStories) GetEventStoryTitle(_ context.Context, actorID, storyID, workspaceID uuid.UUID) (string, error) {
	s.actorID, s.storyID, s.workspaceID = actorID, storyID, workspaceID
	s.calls++
	return "Task title", s.err
}

func TestStoryUpdateResolvesTitleOnceWithExplicitEventActor(t *testing.T) {
	t.Parallel()
	source := &eventTitleStories{}
	rules := NewRules(nil, source, nil, nil)
	actorID, oldAssigneeID, newAssigneeID := uuid.New(), uuid.New(), uuid.New()
	payload := events.StoryUpdatedPayload{
		WorkspaceID: uuid.New(), StoryID: uuid.New(), AssigneeID: &oldAssigneeID,
		Updates: map[string]any{"assignee_id": newAssigneeID.String()},
	}
	batch, err := rules.ProcessStoryUpdate(context.Background(), payload, actorID)
	require.NoError(t, err)
	require.Len(t, batch, 2)
	require.Equal(t, 1, source.calls)
	require.Equal(t, actorID, source.actorID)
	require.Equal(t, payload.StoryID, source.storyID)
	require.Equal(t, payload.WorkspaceID, source.workspaceID)
	for _, notification := range batch {
		require.Equal(t, "Task title", notification.Title)
		require.NoError(t, notification.Validate())
	}
}

func TestStoryUpdateTitleFailuresRemainRetryable(t *testing.T) {
	t.Parallel()
	unavailable := errors.New("database unavailable")
	source := &eventTitleStories{err: unavailable}
	rules := NewRules(nil, source, nil, nil)
	payload := events.StoryUpdatedPayload{
		StoryID: uuid.New(), WorkspaceID: uuid.New(), Updates: map[string]any{"assignee_id": uuid.New().String()},
	}
	batch, err := rules.ProcessStoryUpdate(context.Background(), payload, uuid.New())
	require.ErrorIs(t, err, unavailable)
	require.Empty(t, batch)
	source.err = storydomain.ErrNotFound
	batch, err = rules.ProcessStoryUpdate(context.Background(), payload, uuid.New())
	require.NoError(t, err, "deleted or revoked resources must not create stale notifications")
	require.Empty(t, batch)
}
