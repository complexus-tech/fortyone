package storieshttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestBulkUpdateDoesNotInvalidateUncachedStoryReads(t *testing.T) {
	t.Parallel()

	for _, partialFailure := range []bool{false, true} {
		name := "all stories updated"
		if partialFailure {
			name = "partial update"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			storyIDs := []uuid.UUID{uuid.New(), uuid.New()}
			statusID := uuid.New()
			repository := &bulkStatusUpdateRepository{
				previousStatusID: uuid.New(),
				targetStatusID:   statusID,
				updatedStatuses:  make(map[uuid.UUID]uuid.UUID),
			}
			if partialFailure {
				repository.failedStoryID = storyIDs[1]
			}
			cache := &bulkUpdateCacheSpy{}
			log := logger.NewWithText(io.Discard, slog.LevelError, "bulk-update-test")
			handler := &Handlers{
				stories: stories.New(log, repository, nil, nil),
				cache:   cache,
			}
			body, err := json.Marshal(map[string]any{
				"storyIds": storyIDs,
				"updates":  map[string]any{"statusId": statusID},
			})
			require.NoError(t, err)

			response := invokeStoryMutationHandler(t, handler.BulkUpdate, string(body))

			require.Equal(t, http.StatusOK, response.Code, response.Body.String())
			require.Zero(t, cache.calls.Load(), "uncached story reads must not add Redis work to a bulk update")
			var result struct {
				Data AppBulkUpdateResult `json:"data"`
			}
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
			require.Equal(t, len(storyIDs), result.Data.TotalCount)
			require.Equal(t, partialFailure, result.Data.Partial)
			require.Len(t, result.Data.Items, len(storyIDs))
			expectedUpdates := map[uuid.UUID]uuid.UUID{storyIDs[0]: statusID}
			expectedItems := []AppBulkUpdateItemResult{{StoryID: storyIDs[0], Success: true}}
			if partialFailure {
				require.Equal(t, 1, result.Data.SucceededCount)
				require.Equal(t, 1, result.Data.FailedCount)
				expectedItems = append(expectedItems, AppBulkUpdateItemResult{
					StoryID: storyIDs[1], Error: "story is no longer editable",
				})
			} else {
				require.Equal(t, 2, result.Data.SucceededCount)
				require.Zero(t, result.Data.FailedCount)
				expectedUpdates[storyIDs[1]] = statusID
				expectedItems = append(expectedItems, AppBulkUpdateItemResult{StoryID: storyIDs[1], Success: true})
			}
			require.Equal(t, expectedItems, result.Data.Items)
			require.Equal(t, expectedUpdates, repository.updatedStatuses)
		})
	}
}

type bulkUpdateCacheSpy struct {
	calls atomic.Int32
}

func (cache *bulkUpdateCacheSpy) Delete(context.Context, string) {
	cache.calls.Add(1)
}

func (cache *bulkUpdateCacheSpy) DeleteByPattern(context.Context, string) {
	cache.calls.Add(1)
}

type bulkStatusUpdateRepository struct {
	mu               sync.Mutex
	previousStatusID uuid.UUID
	targetStatusID   uuid.UUID
	failedStoryID    uuid.UUID
	updatedStatuses  map[uuid.UUID]uuid.UUID
}

func (repository *bulkStatusUpdateRepository) Get(_ context.Context, storyID, workspaceID uuid.UUID) (stories.CoreSingleStory, error) {
	return stories.CoreSingleStory{
		ID:                   storyID,
		Workspace:            workspaceID,
		Status:               &repository.previousStatusID,
		AutoSchedulingStatus: stories.AutoSchedulingStatusOff,
	}, nil
}

func (repository *bulkStatusUpdateRepository) GetStatusCategory(_ context.Context, statusID string) (string, error) {
	if statusID == repository.targetStatusID.String() {
		return "backlog", nil
	}
	return "unstarted", nil
}

func (repository *bulkStatusUpdateRepository) Update(_ context.Context, storyID, _ uuid.UUID, updates map[string]any) error {
	if storyID == repository.failedStoryID {
		return errors.New("story is no longer editable")
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	repository.updatedStatuses[storyID] = updates["status_id"].(uuid.UUID)
	return nil
}

func (*bulkStatusUpdateRepository) RecordActivities(_ context.Context, activities []stories.CoreActivity) ([]stories.CoreActivity, error) {
	return activities, nil
}
