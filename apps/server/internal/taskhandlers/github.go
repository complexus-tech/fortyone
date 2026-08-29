package taskhandlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	githubprovider "github.com/complexus-tech/projects-api/internal/modules/github"
	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

func (h *handlers) HandleGitHubWebhook(ctx context.Context, task *asynq.Task) error {
	if h.githubService == nil {
		return fmt.Errorf("GitHub webhook processor is not configured: %w", asynq.SkipRetry)
	}
	var payload tasks.GitHubWebhookPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("decode GitHub webhook task: %w: %w", err, asynq.SkipRetry)
	}
	if payload.InboxID == uuid.Nil {
		return fmt.Errorf("GitHub webhook task has an invalid inbox payload: %w", asynq.SkipRetry)
	}
	if err := h.githubService.ProcessWebhook(ctx, githubprovider.ProviderKey, payload.InboxID); err != nil {
		return fmt.Errorf("process GitHub webhook inbox %s: %w", payload.InboxID, err)
	}
	return nil
}

func (h *handlers) HandleGitHubWebhookRecovery(ctx context.Context, _ *asynq.Task) error {
	if h.githubService == nil {
		return fmt.Errorf("GitHub inbox recoverer is not configured: %w", asynq.SkipRetry)
	}
	recovered, err := h.githubService.RecoverPendingWebhooks(ctx)
	if err != nil {
		return err
	}
	if recovered > 0 && h.log != nil {
		h.log.Info(ctx, "recovered pending GitHub webhooks", "recovered", recovered)
	}
	return nil
}

func (h *handlers) HandleGitHubStorySync(ctx context.Context, t *asynq.Task) error {
	if h.githubService == nil {
		return fmt.Errorf("github service is not configured: %w", asynq.SkipRetry)
	}
	if h.storySyncReader == nil {
		return fmt.Errorf("story sync reader is not configured: %w", asynq.SkipRetry)
	}

	var payload tasks.GitHubStorySyncPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		h.log.Error(ctx, "Failed to unmarshal GitHubStorySyncPayload", "error", err)
		return fmt.Errorf("unmarshal payload failed: %w: %w", err, asynq.SkipRetry)
	}

	scope, err := h.githubStorySyncScope(payload.WorkspaceID)
	if err != nil {
		return fmt.Errorf("construct GitHub story sync scope: %w: %w", err, asynq.SkipRetry)
	}
	story, err := h.storySyncReader.GetStoryForMutation(ctx, scope, payload.StoryID)
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

func (h *handlers) githubStorySyncScope(workspaceID uuid.UUID) (storydomain.MutationScope, error) {
	if h.systemUserID == uuid.Nil || workspaceID == uuid.Nil {
		return storydomain.MutationScope{}, errors.New("system actor and workspace are required")
	}
	actor, err := platformauth.NewActor(
		h.systemUserID,
		platformauth.PrincipalSystem,
		uuid.Nil,
		platformauth.MustScopeSet(platformauth.ScopeStoriesWrite),
		platformauth.UnrestrictedTeamAccess(),
	)
	if err != nil {
		return storydomain.MutationScope{}, err
	}
	actor, err = actor.WithWorkspace(workspaceID)
	if err != nil {
		return storydomain.MutationScope{}, err
	}
	scope := storydomain.MutationScope{
		Actor: actor, WorkspaceID: workspaceID, ActivityUser: &h.systemUserID,
	}
	if err := scope.Validate(); err != nil {
		return storydomain.MutationScope{}, err
	}
	return scope, nil
}
