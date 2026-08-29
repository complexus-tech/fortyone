package github

import (
	"context"
	"strings"
)

func (s *Service) ensureIssueImportComment(ctx context.Context, repository repositoryRecord, issueNumber int, story singleStory) error {
	// Some create paths may not hydrate team_code on the freshly returned story.
	// Re-load before constructing the task key to avoid malformed values like "-418".
	if strings.TrimSpace(story.TeamCode) == "" {
		loadedStory, loadErr := s.stories.Get(ctx, story.ID, repository.WorkspaceID)
		if loadErr == nil {
			story = loadedStory
		}
	}

	storyURL, err := storyURLFromWebsite(
		s.cfg.WebsiteURL,
		repository.WorkspaceSlug,
		buildStoryReference(story.TeamCode, story.SequenceID, story.ID.String()),
	)
	if err != nil {
		return err
	}

	exists, err := s.issueHasStoryComment(
		ctx,
		repository.GitHubInstallationID,
		repository.OwnerLogin,
		repository.RepositorySlug,
		issueNumber,
		story.ID,
		storyURL,
	)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	commentBody := buildLinkedTaskComment(storyURL, story.TeamCode, story.SequenceID)
	return s.createIssueComment(
		ctx,
		repository.GitHubInstallationID,
		repository.OwnerLogin,
		repository.RepositorySlug,
		issueNumber,
		commentBody,
	)
}

func (s *Service) ensurePRLinkedComment(ctx context.Context, repository repositoryRecord, prNumber int, story storyMatch) error {
	storyURL, err := storyURLFromWebsite(
		s.cfg.WebsiteURL,
		repository.WorkspaceSlug,
		buildStoryReference(story.TeamCode, story.SequenceID, story.StoryID.String()),
	)
	if err != nil {
		return err
	}

	exists, err := s.issueHasStoryComment(
		ctx,
		repository.GitHubInstallationID,
		repository.OwnerLogin,
		repository.RepositorySlug,
		prNumber,
		story.StoryID,
		storyURL,
	)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	commentBody := buildLinkedTaskComment(storyURL, story.TeamCode, story.SequenceID)
	return s.createIssueComment(
		ctx,
		repository.GitHubInstallationID,
		repository.OwnerLogin,
		repository.RepositorySlug,
		prNumber,
		commentBody,
	)
}

func issueBodyFromStoryDescription(description *string) string {
	if description == nil {
		return ""
	}
	return stripManagedIssueLink(*description)
}

func stripManagedIssueLink(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}

	lines := strings.Split(value, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, githubIssueLinkPrefix) {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.TrimSpace(strings.Join(filtered, "\n"))
}
