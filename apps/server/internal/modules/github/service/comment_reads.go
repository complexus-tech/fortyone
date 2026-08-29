package github

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/actors"
	githubsdk "github.com/google/go-github/v72/github"
	"github.com/google/uuid"
)

const maxGitHubCommentsPerRead = 1000

// GetStoryGitHubComments fetches comments from all linked GitHub issues for a story.
// It resolves GitHub users to FortyOne users when possible, and for bot comments
// posted via FortyOne it extracts the real author and strips the attribution prefix.
func (s *Service) GetStoryGitHubComments(ctx context.Context, workspaceID, storyID uuid.UUID) ([]GitHubComment, error) {
	if !s.canUseAppAPI() {
		return []GitHubComment{}, nil
	}

	issues, err := s.repo.GetStoryLinkedIssues(ctx, workspaceID, storyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get linked issues: %w", err)
	}
	if len(issues) == 0 {
		return []GitHubComment{}, nil
	}

	// Collect raw comments and unique GitHub user IDs.
	type rawComment struct {
		comment      *githubsdk.IssueComment
		gitHubUserID int64
		isAppAuthor  bool
	}
	var rawComments []rawComment

	for _, issue := range issues {
		client, err := s.newInstallationClient(ctx, issue.GitHubInstallationID)
		if err != nil {
			return nil, fmt.Errorf("create GitHub installation client for comments: %w", err)
		}

		opts := &githubsdk.IssueListCommentsOptions{
			Sort:        githubsdk.Ptr("created"),
			Direction:   githubsdk.Ptr("asc"),
			ListOptions: githubsdk.ListOptions{PerPage: 100},
		}
		appBotLogin := strings.ToLower(strings.TrimSpace(s.cfg.AppSlug))
		if appBotLogin != "" {
			appBotLogin += "[bot]"
		}
		for {
			comments, response, err := client.Issues.ListComments(ctx, issue.OwnerLogin, issue.RepositorySlug, issue.GitHubNumber, opts)
			if err != nil {
				return nil, fmt.Errorf("fetch GitHub comments for issue %d: %w", issue.GitHubNumber, err)
			}

			for _, c := range comments {
				if c == nil {
					continue
				}
				if len(rawComments) >= maxGitHubCommentsPerRead {
					return nil, errors.New("linked GitHub comments exceed the read limit")
				}
				var ghUID int64
				ghLogin := ""
				if c.User != nil {
					ghUID = c.User.GetID()
					ghLogin = strings.ToLower(strings.TrimSpace(c.User.GetLogin()))
				}
				rawComments = append(rawComments, rawComment{
					comment:      c,
					gitHubUserID: ghUID,
					isAppAuthor:  appBotLogin != "" && ghLogin == appBotLogin,
				})
			}
			if response == nil || response.NextPage == 0 {
				break
			}
			opts.Page = response.NextPage
		}
	}

	// Batch-resolve GitHub user IDs → FortyOne users.
	uniqueIDs := make(map[int64]struct{})
	for _, rc := range rawComments {
		if rc.gitHubUserID != 0 {
			uniqueIDs[rc.gitHubUserID] = struct{}{}
		}
	}
	idSlice := make([]int64, 0, len(uniqueIDs))
	for id := range uniqueIDs {
		idSlice = append(idSlice, id)
	}
	userMap, err := s.repo.ResolveFortyOneUsersByGitHubIDs(ctx, idSlice)
	if err != nil {
		return nil, fmt.Errorf("resolve FortyOne users for GitHub comments: %w", err)
	}
	if userMap == nil {
		userMap = map[int64]fortyOneUser{}
	}
	systemGitHubLogin := "github"
	systemGitHubAvatar := ""
	if githubActorEmail, ok := actors.EmailForKey(actors.KeyGitHub); ok {
		if systemUser, err := s.repo.ResolveFortyOneUserByEmail(ctx, githubActorEmail); err == nil {
			if username := strings.TrimSpace(systemUser.Username); username != "" {
				systemGitHubLogin = username
			}
			systemGitHubAvatar = s.resolveAvatarURL(ctx, systemUser.AvatarURL)
		} else if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("resolve GitHub system actor: %w", err)
		}
	}

	// Build result, resolving authors.
	allComments := make([]GitHubComment, 0, len(rawComments))
	for _, rc := range rawComments {
		c := rc.comment
		gc := GitHubComment{
			ID:        c.GetID(),
			Body:      c.GetBody(),
			CreatedAt: c.GetCreatedAt().Format(time.RFC3339),
			UpdatedAt: c.GetUpdatedAt().Format(time.RFC3339),
			HTMLURL:   c.GetHTMLURL(),
		}
		rawGitHubLogin := ""
		if c.User != nil {
			rawGitHubLogin = c.User.GetLogin()
			gc.UserLogin = rawGitHubLogin
			gc.UserAvatar = c.User.GetAvatarURL()
		}

		// Check if this GitHub user is a linked FortyOne user.
		if fu, ok := userMap[rc.gitHubUserID]; ok {
			gc.UserLogin = fu.Username
			gc.UserAvatar = s.resolveAvatarURL(ctx, fu.AvatarURL)
		}

		commentedViaFortyOne := false
		// For app and bot comments, extract the real author when the attribution marker exists.
		if rc.isAppAuthor || strings.HasSuffix(rawGitHubLogin, "[bot]") {
			if match := fortyOneCommentPattern.FindStringSubmatch(gc.Body); match != nil {
				commentedViaFortyOne = true
				authorName := match[1]
				gc.Body = stripFortyOneCommentMarker(match[2])

				// Try to resolve the author name to a FortyOne user.
				if fu, err := s.repo.ResolveFortyOneUserByFullName(ctx, authorName); err == nil {
					gc.UserLogin = fu.Username
					gc.UserAvatar = s.resolveAvatarURL(ctx, fu.AvatarURL)
				} else if errors.Is(err, sql.ErrNoRows) {
					gc.UserLogin = authorName
					gc.UserAvatar = ""
				} else {
					return nil, fmt.Errorf("resolve FortyOne author for GitHub comment: %w", err)
				}
			}
		}
		gc.Body = stripFortyOneCommentMarker(gc.Body)
		// For app-authored/system comments without a resolvable author, expose a
		// stable GitHub label instead of the app account login (e.g. fortyone-app[bot]).
		if !commentedViaFortyOne && (rc.isAppAuthor ||
			isFortyOneBotAuthorLogin(rawGitHubLogin, s.cfg.AppSlug) ||
			isFortyOneSystemLinkedTaskComment(gc.Body)) {
			gc.UserLogin = systemGitHubLogin
			gc.UserAvatar = systemGitHubAvatar
		}

		allComments = append(allComments, gc)
	}

	return allComments, nil
}

func (s *Service) GetRequestGitHubComments(ctx context.Context, workspaceID, requestID uuid.UUID) ([]GitHubComment, error) {
	if !s.canUseAppAPI() {
		return []GitHubComment{}, nil
	}
	repository, issueNumber, err := s.requestGitHubIssue(ctx, workspaceID, requestID)
	if err != nil {
		return nil, err
	}
	client, err := s.newInstallationClient(ctx, repository.GitHubInstallationID)
	if err != nil {
		return nil, err
	}
	opts := &githubsdk.IssueListCommentsOptions{
		Sort:        githubsdk.Ptr("created"),
		Direction:   githubsdk.Ptr("asc"),
		ListOptions: githubsdk.ListOptions{PerPage: 100},
	}
	result := make([]GitHubComment, 0)
	for {
		comments, response, err := client.Issues.ListComments(ctx, repository.OwnerLogin, repository.RepositorySlug, issueNumber, opts)
		if err != nil {
			return nil, err
		}
		for _, comment := range comments {
			if comment == nil {
				continue
			}
			if len(result) >= maxGitHubCommentsPerRead {
				return nil, errors.New("GitHub comments exceed the read limit")
			}
			user := comment.GetUser()
			avatar := user.GetAvatarURL()
			result = append(result, GitHubComment{
				ID:         comment.GetID(),
				Body:       stripFortyOneCommentMarker(stripFortyOneBotAttribution(comment.GetBody())),
				UserLogin:  user.GetLogin(),
				UserAvatar: avatar,
				CreatedAt:  comment.GetCreatedAt().Format(time.RFC3339),
				UpdatedAt:  comment.GetUpdatedAt().Format(time.RFC3339),
				HTMLURL:    comment.GetHTMLURL(),
			})
		}
		if response == nil || response.NextPage == 0 {
			break
		}
		opts.Page = response.NextPage
	}
	return result, nil
}
