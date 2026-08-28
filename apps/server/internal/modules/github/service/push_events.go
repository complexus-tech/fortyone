package github

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

func (s *Service) handlePushEvent(ctx context.Context, repository repositoryRecord, payload webhookEnvelope) error {
	if s.stories == nil {
		return errors.New("stories service is not configured")
	}

	settings, err := s.repo.GetWorkspaceSettings(ctx, repository.WorkspaceID)
	if err != nil {
		return err
	}

	refs := extractStoryRefs(payload.Ref)
	if settings.LinkCommitsByMagicWords {
		for _, commit := range payload.Commits {
			refs = append(refs, extractStoryRefs(commit.Message)...)
		}
	}
	stories, err := s.repo.ResolveStoriesByRefs(ctx, repository.WorkspaceID, refs)
	if err != nil || len(stories) == 0 {
		return err
	}
	branch := strings.TrimPrefix(payload.Ref, "refs/heads/")
	for _, story := range stories {
		newCommits := 0
		if branch != "" {
			branchLinkCreated, err := s.upsertStoryLink(ctx, repository.WorkspaceID, story.StoryID, repository.ID, "branch", 0, 0, &branch, payload.Repository.HTMLURL+"/tree/"+branch, branch, "active", map[string]any{"ref": payload.Ref})
			if err != nil {
				return err
			}
			if err := s.repo.EnsureStoryLink(ctx, story.StoryID, githubBranchStoryLinkTitle(branch), payload.Repository.HTMLURL+"/tree/"+branch); err != nil {
				return err
			}
			if branchLinkCreated {
				if err := s.recordLinkActivity(ctx, repository.WorkspaceID, story.StoryID, "github_branch", fmt.Sprintf("branch %s", branch), payload.Repository.HTMLURL+"/tree/"+branch); err != nil {
					return err
				}
			}
		}
		for _, commit := range payload.Commits {
			refName := commit.ID
			commitLinkCreated, err := s.upsertStoryLink(ctx, repository.WorkspaceID, story.StoryID, repository.ID, "commit", 0, 0, &refName, commit.URL, commit.Message, "linked", commit)
			if err != nil {
				return err
			}
			if commitLinkCreated {
				newCommits++
			}
		}
		if newCommits > 0 {
			label := "1 commit"
			if newCommits > 1 {
				label = fmt.Sprintf("%d commits", newCommits)
			}
			if err := s.recordLinkActivity(ctx, repository.WorkspaceID, story.StoryID, "github_commit", label, payload.Repository.HTMLURL+"/commits/"+branch); err != nil {
				return err
			}
		}
		// Auto-close from commit keywords is configurable at workspace level.
		if settings.CloseOnCommitKeywords && hasClosingKeyword(payload.Commits, story.TeamCode, story.SequenceID) {
			if err := s.moveStoryByRule(ctx, repository.WorkspaceID, story.TeamID, story.StoryID, EventCommitClose, nil); err != nil {
				return err
			}
		}
	}
	return nil
}

var closingKeywordPattern = regexp.MustCompile(`(?i)\b(fix|fixes|fixed|close|closes|closed|resolve|resolves|resolved)\b\s+`)

func hasClosingKeyword(commits []struct {
	ID      string `json:"id"`
	Message string `json:"message"`
	URL     string `json:"url"`
}, teamCode string, sequenceID int) bool {
	target := strings.ToUpper(fmt.Sprintf("%s-%d", teamCode, sequenceID))
	targetAlt := strings.ToUpper(fmt.Sprintf("%s %d", teamCode, sequenceID))
	for _, commit := range commits {
		upper := strings.ToUpper(commit.Message)
		if !closingKeywordPattern.MatchString(commit.Message) {
			continue
		}
		if strings.Contains(upper, target) || strings.Contains(upper, targetAlt) {
			return true
		}
	}
	return false
}
