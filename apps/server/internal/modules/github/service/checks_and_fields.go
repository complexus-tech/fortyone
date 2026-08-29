package github

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	githubsdk "github.com/google/go-github/v72/github"
	"github.com/google/uuid"
)

// ==================== check_run ====================

func (s *Service) handleCheckRunEvent(ctx context.Context, repository repositoryRecord, payload webhookEnvelope) error {
	if payload.Action != "completed" {
		return nil
	}
	for _, pr := range payload.CheckRun.PullRequests {
		matches, err := s.repo.FindStoryLinksByPRNumber(ctx, repository.ID, pr.Number)
		if errors.Is(err, sql.ErrNoRows) || len(matches) == 0 {
			continue
		}
		if err != nil {
			return fmt.Errorf("resolve stories for GitHub check run: %w", err)
		}
		for _, story := range matches {
			if err := s.repo.UpdateStoryLinkCheckState(ctx, story.StoryID, repository.ID, pr.ID, payload.CheckRun.Conclusion); err != nil {
				return fmt.Errorf("update GitHub check state: %w", err)
			}
		}
	}
	return nil
}

// ==================== auto-populate PR body ====================

func (s *Service) autoPopulatePRBody(ctx context.Context, repository repositoryRecord, payload webhookEnvelope, story storyMatch) error {
	settings, err := s.repo.GetWorkspaceSettingsByWorkspaceID(ctx, repository.WorkspaceID)
	if err != nil {
		return fmt.Errorf("load GitHub workspace settings: %w", err)
	}
	if !settings.AutoPopulatePRBody {
		return nil
	}
	storyURL, err := storyURLFromWebsite(
		s.cfg.WebsiteURL,
		repository.WorkspaceSlug,
		buildStoryReference(story.TeamCode, story.SequenceID, story.StoryID.String()),
	)
	if err != nil {
		return err
	}
	taskKey := fmt.Sprintf("%s-%d", story.TeamCode, story.SequenceID)
	marker := fmt.Sprintf("<!-- fortyone:%s -->", story.StoryID.String())
	if strings.Contains(payload.PullRequest.Body, marker) {
		return nil
	}
	footer := fmt.Sprintf("\n\n%s\n---\nLinked to [%s](%s)", marker, taskKey, storyURL)
	newBody := payload.PullRequest.Body + footer
	client, clientErr := s.newInstallationClient(ctx, repository.GitHubInstallationID)
	if clientErr != nil {
		return clientErr
	}
	_, _, err = client.PullRequests.Edit(ctx, repository.OwnerLogin, repository.RepositorySlug, payload.PullRequest.Number, &githubsdk.PullRequest{
		Body: &newBody,
	})
	return err
}

// ==================== assignee sync ====================

func (s *Service) syncAssigneeFromGitHub(ctx context.Context, repository repositoryRecord, story storyMatch, assignee *struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
}) error {
	settings, err := s.repo.GetWorkspaceSettingsByWorkspaceID(ctx, repository.WorkspaceID)
	if err != nil {
		return fmt.Errorf("load GitHub workspace settings: %w", err)
	}
	if !settings.SyncAssignees || assignee == nil || assignee.ID == 0 {
		return nil
	}
	userID, err := s.repo.ResolveUserByGitHubID(ctx, assignee.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("resolve GitHub assignee: %w", err)
	}
	fullStory, err := s.stories.Get(ctx, story.StoryID, repository.WorkspaceID)
	if err != nil {
		return err
	}
	if fullStory.Assignee != nil && *fullStory.Assignee == userID {
		return nil
	}
	return s.stories.UpdateExternalWithReason(ctx, s.cfg.GitHubUserID, story.StoryID, repository.WorkspaceID, map[string]any{
		"assignee_id": userID,
	}, githubAssigneeSyncReason())
}

// ==================== label sync ====================

func (s *Service) syncLabelsFromGitHub(ctx context.Context, repository repositoryRecord, story storyMatch, labels []struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}) error {
	settings, err := s.repo.GetWorkspaceSettingsByWorkspaceID(ctx, repository.WorkspaceID)
	if err != nil {
		return fmt.Errorf("load GitHub workspace settings: %w", err)
	}
	if !settings.SyncLabels || len(labels) == 0 {
		return nil
	}
	labelNames := make([]string, 0, len(labels))
	for _, l := range labels {
		labelNames = append(labelNames, l.Name)
	}
	resolvedIDs, err := s.repo.ResolveOrCreateLabelsByName(ctx, repository.WorkspaceID, story.TeamID, labelNames)
	if err != nil {
		return fmt.Errorf("resolve GitHub labels: %w", err)
	}
	if len(resolvedIDs) == 0 {
		return nil
	}
	return s.stories.UpdateExternalWithReason(ctx, s.cfg.GitHubUserID, story.StoryID, repository.WorkspaceID, map[string]any{
		"labels": resolvedIDs,
	}, githubLabelSyncReason())
}

// ==================== user resolution ====================

func (s *Service) resolveActorID(ctx context.Context, githubUserID int64) (uuid.UUID, error) {
	if githubUserID == 0 {
		return s.cfg.GitHubUserID, nil
	}
	userID, err := s.repo.ResolveUserByGitHubID(ctx, githubUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return s.cfg.GitHubUserID, nil
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("resolve FortyOne actor for GitHub user: %w", err)
	}
	return userID, nil
}
