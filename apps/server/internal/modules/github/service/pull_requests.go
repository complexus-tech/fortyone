package github

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

func (s *Service) handlePullRequestEvent(ctx context.Context, repository repositoryRecord, payload webhookEnvelope) error {
	if s.stories == nil {
		return errors.New("stories service is not configured")
	}

	refs := extractStoryRefs(payload.PullRequest.Title, payload.PullRequest.Body, payload.PullRequest.Head.Ref)
	stories, err := s.repo.ResolveStoriesByRefs(ctx, repository.WorkspaceID, refs)
	if err != nil || len(stories) == 0 {
		return err
	}

	eventKey, shouldMoveStory := pullRequestWorkflowEvent(payload.Action, payload.PullRequest.Draft, payload.PullRequest.Merged)

	for _, story := range stories {
		prLinkCreated, err := s.upsertStoryLink(ctx, repository.WorkspaceID, story.StoryID, repository.ID, "pull_request", payload.PullRequest.ID, payload.PullRequest.Number, nil, payload.PullRequest.HTMLURL, payload.PullRequest.Title, payload.PullRequest.State, payload.PullRequest)
		if err != nil {
			return err
		}
		if err := s.repo.EnsureStoryLink(ctx, story.StoryID, githubPullRequestStoryLinkTitle(payload.PullRequest.Number), payload.PullRequest.HTMLURL); err != nil {
			return err
		}
		if prLinkCreated {
			if err := s.recordLinkActivity(ctx, repository.WorkspaceID, story.StoryID, "github_pull_request", fmt.Sprintf("PR #%d", payload.PullRequest.Number), payload.PullRequest.HTMLURL); err != nil {
				return err
			}
			if err := s.ensurePRLinkedComment(ctx, repository, payload.PullRequest.Number, story); err != nil {
				return err
			}
			if err := s.autoPopulatePRBody(ctx, repository, payload, story); err != nil {
				return err
			}
		}
		// Sync assignee from PR to story
		if err := s.syncAssigneeFromGitHub(ctx, repository, story, payload.PullRequest.Assignee); err != nil {
			return err
		}
		// Sync labels from PR to story
		if err := s.syncLabelsFromGitHub(ctx, repository, story, payload.PullRequest.Labels); err != nil {
			return err
		}

		branchRef := payload.PullRequest.Head.Ref
		if branchRef != "" {
			branchLinkCreated, err := s.upsertStoryLink(ctx, repository.WorkspaceID, story.StoryID, repository.ID, "branch", 0, 0, &branchRef, payload.PullRequest.Head.HTMLURL, payload.PullRequest.Head.Ref, payload.PullRequest.State, payload.PullRequest.Head)
			if err != nil {
				return err
			}
			if err := s.repo.EnsureStoryLink(ctx, story.StoryID, githubBranchStoryLinkTitle(branchRef), payload.PullRequest.Head.HTMLURL); err != nil {
				return err
			}
			if branchLinkCreated {
				if err := s.recordLinkActivity(ctx, repository.WorkspaceID, story.StoryID, "github_branch", fmt.Sprintf("branch %s", branchRef), payload.PullRequest.Head.HTMLURL); err != nil {
					return err
				}
			}
		}
		if shouldMoveStory {
			if err := s.moveStoryByRule(ctx, repository.WorkspaceID, story.TeamID, story.StoryID, eventKey, &payload.PullRequest.Base.Ref); err != nil {
				return err
			}
		}
	}
	return nil
}

func pullRequestWorkflowEvent(action string, draft bool, merged bool) (string, bool) {
	switch action {
	case "opened":
		if draft {
			return EventDraftPROpen, true
		}
		return EventPROpen, true
	case "ready_for_review":
		return EventPRReadyForMerge, true
	case "closed":
		if merged {
			return EventPRMerge, true
		}
		return "", false
	default:
		return "", false
	}
}

func (s *Service) handleCreateEvent(ctx context.Context, repository repositoryRecord, payload webhookEnvelope) error {
	if s.stories == nil {
		return errors.New("stories service is not configured")
	}
	if payload.RefType != "branch" || strings.TrimSpace(payload.Ref) == "" {
		return nil
	}

	refs := extractStoryRefs(payload.Ref)
	stories, err := s.repo.ResolveStoriesByRefs(ctx, repository.WorkspaceID, refs)
	if err != nil || len(stories) == 0 {
		return err
	}

	branchURL := payload.Repository.HTMLURL + "/tree/" + payload.Ref
	for _, story := range stories {
		branchRef := payload.Ref
		branchLinkCreated, err := s.upsertStoryLink(ctx, repository.WorkspaceID, story.StoryID, repository.ID, "branch", 0, 0, &branchRef, branchURL, payload.Ref, "active", map[string]any{"ref": payload.Ref, "ref_type": payload.RefType})
		if err != nil {
			return err
		}
		if err := s.repo.EnsureStoryLink(ctx, story.StoryID, githubBranchStoryLinkTitle(branchRef), branchURL); err != nil {
			return err
		}
		if branchLinkCreated {
			if err := s.recordLinkActivity(ctx, repository.WorkspaceID, story.StoryID, "github_branch", fmt.Sprintf("branch %s", branchRef), branchURL); err != nil {
				return err
			}
		}
	}
	return nil
}
