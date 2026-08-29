package github

import (
	"context"
	"errors"
	"fmt"
	"strings"

	githubsdk "github.com/google/go-github/v72/github"
)

// ==================== pull_request_review ====================

func (s *Service) handlePullRequestReviewEvent(ctx context.Context, repository repositoryRecord, payload webhookEnvelope) error {
	if s.stories == nil {
		return errors.New("stories service is not configured")
	}
	if payload.Action != "submitted" && payload.Action != "dismissed" {
		return nil
	}

	refs := extractStoryRefs(payload.PullRequest.Title, payload.PullRequest.Body, payload.PullRequest.Head.Ref)
	matches, err := s.repo.ResolveStoriesByRefs(ctx, repository.WorkspaceID, refs)
	if err != nil || len(matches) == 0 {
		return err
	}

	reviewState := normalizeReviewState(payload.Review.State)
	approved, changesRequested, err := countReviews(ctx, s, repository, payload.PullRequest.Number)
	if err != nil {
		return err
	}
	reviewActorID, err := s.resolveActorID(ctx, payload.Review.User.ID)
	if err != nil {
		return err
	}
	for _, story := range matches {
		summaryState := summarizeReviewState(approved, changesRequested)
		if err := s.repo.UpdateStoryLinkReviewState(ctx, story.StoryID, repository.ID, payload.PullRequest.ID, summaryState, approved, changesRequested); err != nil {
			return fmt.Errorf("update GitHub review state: %w", err)
		}

		if err := s.stories.RecordActivity(ctx, storyActivity{
			StoryID:      story.StoryID,
			Type:         "link",
			Field:        "github_review",
			CurrentValue: fmt.Sprintf("PR #%d %s by %s", payload.PullRequest.Number, reviewState, payload.Review.User.Login),
			NewValue:     payload.Review.HTMLURL,
			UserID:       reviewActorID,
			WorkspaceID:  repository.WorkspaceID,
		}); err != nil {
			return err
		}

		if err := s.moveStoryByRule(ctx, repository.WorkspaceID, story.TeamID, story.StoryID, EventPRReviewActivity, &payload.PullRequest.Base.Ref); err != nil {
			return err
		}
	}
	return nil
}

func normalizeReviewState(state string) string {
	switch strings.ToLower(state) {
	case "approved":
		return "approved"
	case "changes_requested":
		return "changes_requested"
	case "dismissed":
		return "dismissed"
	case "commented":
		return "commented"
	default:
		return state
	}
}

func countReviews(ctx context.Context, s *Service, repository repositoryRecord, prNumber int) (approved, changesRequested int, err error) {
	client, err := s.newInstallationClient(ctx, repository.GitHubInstallationID)
	if err != nil {
		return 0, 0, err
	}
	reviews, _, err := client.PullRequests.ListReviews(ctx, repository.OwnerLogin, repository.RepositorySlug, prNumber, &githubsdk.ListOptions{PerPage: 100})
	if err != nil {
		return 0, 0, err
	}
	latestByUser := map[int64]string{}
	for _, r := range reviews {
		if r.GetUser() != nil {
			latestByUser[r.GetUser().GetID()] = r.GetState()
		}
	}
	for _, state := range latestByUser {
		switch strings.ToUpper(state) {
		case "APPROVED":
			approved++
		case "CHANGES_REQUESTED":
			changesRequested++
		}
	}
	return approved, changesRequested, nil
}

func summarizeReviewState(approved, changesRequested int) string {
	if changesRequested > 0 {
		return "changes_requested"
	}
	if approved > 0 {
		return "approved"
	}
	return "pending"
}
