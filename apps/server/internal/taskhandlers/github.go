package taskhandlers

import (
	"context"
	"encoding/json"
	"fmt"

	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/hibiken/asynq"
)

func (h *handlers) HandleGitHubStorySync(ctx context.Context, t *asynq.Task) error {
	if h.githubService == nil {
		return fmt.Errorf("github service is not configured: %w", asynq.SkipRetry)
	}

	var payload tasks.GitHubStorySyncPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		h.log.Error(ctx, "Failed to unmarshal GitHubStorySyncPayload", "error", err)
		return fmt.Errorf("unmarshal payload failed: %w: %w", err, asynq.SkipRetry)
	}

	repo := storiesrepository.New(h.log, h.db)
	story, err := repo.Get(ctx, payload.StoryID, payload.WorkspaceID)
	if err != nil {
		h.log.Error(ctx, "Failed to load story for GitHub sync", "error", err, "story_id", payload.StoryID)
		return err
	}

	if err := h.githubService.SyncStoryFromFortyOne(ctx, github.CoreStorySyncInput{
		StoryID:     story.ID,
		WorkspaceID: payload.WorkspaceID,
		TeamID:      story.Team,
		Title:       story.Title,
		Description: story.Description,
		StatusID:    story.Status,
	}); err != nil {
		h.log.Error(ctx, "Failed to sync story to GitHub", "error", err, "story_id", payload.StoryID)
		return err
	}

	return nil
}
