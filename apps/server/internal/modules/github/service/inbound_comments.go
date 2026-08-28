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

// ==================== issue_comment ====================

func (s *Service) handleIssueCommentEvent(ctx context.Context, repository repositoryRecord, payload webhookEnvelope) error {
	if s.stories == nil {
		return errors.New("stories service is not configured")
	}
	if payload.Action != "created" {
		return nil
	}
	// Ignore comments from our own app to prevent loops.
	if payload.Comment.User.ID == 0 ||
		isFortyOneAuthoredCommentBody(payload.Comment.Body) ||
		isFortyOneSystemLinkedTaskComment(payload.Comment.Body) ||
		s.isAppSender(payload.Comment.User.Login) {
		return nil
	}
	if isOutbound, err := s.repo.IsOutboundGitHubComment(ctx, repository.ID, payload.Comment.ID); err != nil {
		return err
	} else if isOutbound {
		return nil
	}

	_, storyID, err := s.repo.FindStoryLink(ctx, repository.ID, "issue", payload.Issue.ID, nil)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil // Issue not linked to any story.
		}
		return fmt.Errorf("resolve story for GitHub issue comment: %w", err)
	}

	actorID, err := s.resolveActorID(ctx, payload.Comment.User.ID)
	if err != nil {
		return err
	}
	reserved, err := s.repo.ReserveInboundGitHubComment(ctx, repository.WorkspaceID, storyID, repository.ID, payload.Comment.ID, actorID)
	if err != nil {
		return err
	}
	if !reserved {
		return nil
	}

	commentBody := fmt.Sprintf("**%s** commented on GitHub issue #%d:\n\n%s", payload.Comment.User.Login, payload.Issue.Number, payload.Comment.Body)
	comment, err := s.stories.CreateCommentExternal(ctx, actorID, repository.WorkspaceID, newStoryComment{
		StoryID: storyID,
		UserID:  actorID,
		Comment: commentBody,
	})
	if err != nil {
		_ = s.repo.DeleteGitHubCommentLink(ctx, repository.ID, payload.Comment.ID)
		return err
	}
	return s.repo.CompleteInboundGitHubComment(ctx, repository.ID, payload.Comment.ID, comment.ID)
}

func (s *Service) isAppSender(senderLogin string) bool {
	login := strings.ToLower(strings.TrimSpace(senderLogin))
	if login == "" {
		return false
	}
	appSlug := strings.ToLower(strings.TrimSpace(s.cfg.AppSlug))
	if appSlug == "" {
		return false
	}

	return login == appSlug+"[bot]"
}

// ==================== outbound comment sync (bidirectional) ====================

// SyncCommentToGitHub posts a FortyOne comment to the linked GitHub issue
// when the issue sync link is configured as bidirectional.
func (s *Service) SyncCommentToGitHub(ctx context.Context, workspaceID, storyID, teamID, localCommentID uuid.UUID, authorName, content string) error {
	link, err := s.repo.FindBidirectionalIssueSyncLinkByTeamID(ctx, workspaceID, teamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil // No bidirectional link for this team.
		}
		return fmt.Errorf("resolve bidirectional GitHub issue link: %w", err)
	}

	issueLink, err := s.repo.FindIssueStoryLinkByStoryID(ctx, workspaceID, storyID, link.RepositoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil // Story has no linked GitHub issue.
		}
		return fmt.Errorf("resolve linked GitHub issue: %w", err)
	}

	client, err := s.newInstallationClient(ctx, link.GitHubInstallationID)
	if err != nil {
		return fmt.Errorf("failed to create installation client: %w", err)
	}

	body := buildFortyOneBotCommentBody(authorName, content, localCommentID)
	comment, _, err := client.Issues.CreateComment(ctx, link.OwnerLogin, link.RepositorySlug, issueLink.GitHubNumber, &githubsdk.IssueComment{
		Body: &body,
	})
	if err != nil {
		return fmt.Errorf("failed to create github issue comment: %w", err)
	}
	return s.repo.RecordOutboundGitHubComment(ctx, workspaceID, storyID, link.RepositoryID, comment.GetID(), &localCommentID, s.cfg.GitHubUserID)
}
