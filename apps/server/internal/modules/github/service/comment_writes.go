package github

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	githubsdk "github.com/google/go-github/v72/github"
	"github.com/google/uuid"
)

var (
	ErrGitHubAppNotConfigured   = errors.New("github app api is not configured")
	ErrNoLinkedGitHubIssues     = errors.New("no linked github issues found")
	ErrGitHubCommentKeyConflict = errors.New("github comment idempotency key was already used for different content")
)

// PostCommentToGitHub posts a comment to all linked GitHub issues for a story.
// If the user has a linked GitHub account with a stored token, the comment is
// posted as the user directly. Otherwise it falls back to the installation bot.
func (s *Service) PostCommentToGitHub(ctx context.Context, workspaceID, storyID, userID uuid.UUID, localCommentID *uuid.UUID, authorName, body string) error {
	if !s.canUseAppAPI() {
		return ErrGitHubAppNotConfigured
	}

	issues, err := s.repo.GetStoryLinkedIssues(ctx, workspaceID, storyID)
	if err != nil {
		return fmt.Errorf("failed to get linked issues: %w", err)
	}
	if len(issues) == 0 {
		return ErrNoLinkedGitHubIssues
	}

	// Try to use the user's own GitHub token so the comment appears as them.
	userToken, tokenErr := s.userGitHubToken(ctx, userID)
	useUserToken := tokenErr == nil && strings.TrimSpace(userToken) != ""
	if tokenErr != nil && !errors.Is(tokenErr, sql.ErrNoRows) {
		return fmt.Errorf("load GitHub user credential for comment: %w", tokenErr)
	}
	var userClient *githubsdk.Client
	if useUserToken {
		userClient = githubsdk.NewClient(s.httpClient).WithAuthToken(userToken)
	}

	commentMarkerID := uuid.New()
	if localCommentID != nil {
		commentMarkerID = *localCommentID
	}
	userCommentBody := buildFortyOneUserCommentBody(body, commentMarkerID)
	fallbackCommentBody := buildFortyOneBotCommentBody(authorName, body, commentMarkerID)

	for _, issue := range issues {
		installationClient, err := s.newInstallationClient(ctx, issue.GitHubInstallationID)
		if err != nil {
			return fmt.Errorf("create GitHub installation client: %w", err)
		}
		existingCommentID, existingBody, exists, err := s.findIssueCommentMarker(
			ctx,
			installationClient,
			issue.OwnerLogin,
			issue.RepositorySlug,
			issue.GitHubNumber,
			commentMarkerID,
		)
		if err != nil {
			return err
		}
		if exists {
			if !matchesGitHubCommentIntent(existingBody, body) {
				return ErrGitHubCommentKeyConflict
			}
			if err := s.repo.RecordOutboundGitHubComment(ctx, workspaceID, storyID, issue.RepositoryID, existingCommentID, localCommentID, userID); err != nil {
				return fmt.Errorf("record existing outbound GitHub comment: %w", err)
			}
			continue
		}

		var comment *githubsdk.IssueComment
		if useUserToken {
			comment, _, err = userClient.Issues.CreateComment(ctx, issue.OwnerLogin, issue.RepositorySlug, issue.GitHubNumber, &githubsdk.IssueComment{
				Body: &userCommentBody,
			})
			if err == nil {
				if recordErr := s.repo.RecordOutboundGitHubComment(ctx, workspaceID, storyID, issue.RepositoryID, comment.GetID(), localCommentID, userID); recordErr != nil {
					return fmt.Errorf("record outbound GitHub user comment: %w", recordErr)
				}
				continue
			}

			// If the linked user token is stale/revoked or lacks scope, transparently
			// fall back to installation auth so comment posting still succeeds.
			if !isGitHubAuthOrPermissionError(err) {
				return err
			}

			s.log.Warn(ctx, "github user token failed for comment; falling back to app installation",
				"issue_number", issue.GitHubNumber,
				"owner", issue.OwnerLogin,
				"repo", issue.RepositorySlug,
				"outcome", "authentication_or_permission_denied",
			)
			// Stop retrying the same bad user token for remaining linked issues.
			useUserToken = false
		}

		comment, _, err = installationClient.Issues.CreateComment(ctx, issue.OwnerLogin, issue.RepositorySlug, issue.GitHubNumber, &githubsdk.IssueComment{
			Body: &fallbackCommentBody,
		})
		if err != nil {
			return err
		}
		if recordErr := s.repo.RecordOutboundGitHubComment(ctx, workspaceID, storyID, issue.RepositoryID, comment.GetID(), localCommentID, userID); recordErr != nil {
			return fmt.Errorf("record outbound GitHub app comment: %w", recordErr)
		}
	}
	return nil
}

func (s *Service) PostRequestCommentToGitHub(ctx context.Context, workspaceID, requestID, userID uuid.UUID, localCommentID *uuid.UUID, authorName, body string) error {
	if !s.canUseAppAPI() {
		return ErrGitHubAppNotConfigured
	}
	repository, issueNumber, err := s.requestGitHubIssue(ctx, workspaceID, requestID)
	if err != nil {
		return err
	}

	commentMarkerID := uuid.New()
	if localCommentID != nil {
		commentMarkerID = *localCommentID
	}
	userCommentBody := buildFortyOneUserCommentBody(body, commentMarkerID)
	fallbackCommentBody := buildFortyOneBotCommentBody(authorName, body, commentMarkerID)
	installationClient, err := s.newInstallationClient(ctx, repository.GitHubInstallationID)
	if err != nil {
		return err
	}
	_, existingBody, exists, err := s.findIssueCommentMarker(ctx, installationClient, repository.OwnerLogin, repository.RepositorySlug, issueNumber, commentMarkerID)
	if err != nil {
		return err
	}
	if exists {
		if !matchesGitHubCommentIntent(existingBody, body) {
			return ErrGitHubCommentKeyConflict
		}
		return nil
	}

	userToken, tokenErr := s.userGitHubToken(ctx, userID)
	if tokenErr != nil && !errors.Is(tokenErr, sql.ErrNoRows) {
		return fmt.Errorf("load GitHub user credential for comment: %w", tokenErr)
	}
	if tokenErr == nil && strings.TrimSpace(userToken) != "" {
		userClient := githubsdk.NewClient(s.httpClient).WithAuthToken(userToken)
		if _, _, err := userClient.Issues.CreateComment(ctx, repository.OwnerLogin, repository.RepositorySlug, issueNumber, &githubsdk.IssueComment{
			Body: &userCommentBody,
		}); err == nil {
			return nil
		} else if !isGitHubAuthOrPermissionError(err) {
			return err
		}
	}

	_, _, err = installationClient.Issues.CreateComment(ctx, repository.OwnerLogin, repository.RepositorySlug, issueNumber, &githubsdk.IssueComment{Body: &fallbackCommentBody})
	return err
}

func (s *Service) findIssueCommentMarker(
	ctx context.Context,
	client *githubsdk.Client,
	owner, repository string,
	issueNumber int,
	commentID uuid.UUID,
) (int64, string, bool, error) {
	if client == nil || commentID == uuid.Nil {
		return 0, "", false, errors.New("GitHub comment marker lookup is not configured")
	}
	marker := fortyOneCommentMarker(commentID)
	options := &githubsdk.IssueListCommentsOptions{ListOptions: githubsdk.ListOptions{PerPage: 100}}
	seen := 0
	for {
		comments, response, err := client.Issues.ListComments(ctx, owner, repository, issueNumber, options)
		if err != nil {
			return 0, "", false, err
		}
		for _, comment := range comments {
			if comment == nil {
				continue
			}
			seen++
			if seen > maxGitHubCommentsPerRead {
				return 0, "", false, errors.New("GitHub comments exceed the idempotency lookup limit")
			}
			if strings.Contains(comment.GetBody(), marker) {
				return comment.GetID(), comment.GetBody(), true, nil
			}
		}
		if response == nil || response.NextPage == 0 {
			return 0, "", false, nil
		}
		options.Page = response.NextPage
	}
}

func matchesGitHubCommentIntent(existingBody, requestedBody string) bool {
	existing := stripFortyOneCommentMarker(stripFortyOneBotAttribution(existingBody))
	return strings.TrimSpace(existing) == strings.TrimSpace(requestedBody)
}

func (s *Service) requestGitHubIssue(ctx context.Context, workspaceID, requestID uuid.UUID) (repositoryRecord, int, error) {
	if s.requests == nil {
		return repositoryRecord{}, 0, errors.New("integration request service is not configured")
	}
	request, err := s.requests.Get(ctx, workspaceID, requestID)
	if err != nil {
		return repositoryRecord{}, 0, err
	}
	if request.Provider != providerGitHub || request.SourceType != requestSourceTypeIssue {
		return repositoryRecord{}, 0, errors.New("request is not linked to a github issue")
	}
	repositoryID, err := metadataUUID(request.Metadata, "repository_id")
	if err != nil {
		return repositoryRecord{}, 0, err
	}
	repository, err := s.repo.FindRepositoryByID(ctx, workspaceID, repositoryID)
	if err != nil {
		return repositoryRecord{}, 0, err
	}
	if request.SourceNumber == nil || *request.SourceNumber == 0 {
		return repositoryRecord{}, 0, errors.New("github issue number is required")
	}
	return repository, *request.SourceNumber, nil
}

func isGitHubAuthOrPermissionError(err error) bool {
	if err == nil {
		return false
	}
	var ghErr *githubsdk.ErrorResponse
	if errors.As(err, &ghErr) && ghErr.Response != nil {
		return ghErr.Response.StatusCode == http.StatusUnauthorized || ghErr.Response.StatusCode == http.StatusForbidden
	}
	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, "bad credentials") ||
		strings.Contains(errText, "requires authentication") ||
		strings.Contains(errText, "resource not accessible by personal access token")
}

func isFortyOneBotAuthorLogin(login, configuredSlug string) bool {
	normalizedLogin := strings.ToLower(strings.TrimSpace(login))
	if !strings.HasSuffix(normalizedLogin, "[bot]") {
		return false
	}

	configured := strings.ToLower(strings.TrimSpace(configuredSlug))
	if configured != "" && normalizedLogin == configured+"[bot]" {
		return true
	}

	// Keep this strict enough to avoid relabeling unrelated bots.
	return strings.HasPrefix(normalizedLogin, "fortyone-")
}

func isFortyOneSystemLinkedTaskComment(body string) bool {
	return strings.HasPrefix(strings.TrimSpace(body), "Linked to FortyOne task [")
}

func isFortyOneAuthoredCommentBody(body string) bool {
	trimmed := strings.TrimSpace(body)
	return fortyOneCommentPattern.MatchString(trimmed) || fortyOneCommentMarkerPattern.MatchString(trimmed)
}

func buildFortyOneUserCommentBody(body string, commentID uuid.UUID) string {
	return fmt.Sprintf("%s\n\n%s", strings.TrimSpace(body), fortyOneCommentMarker(commentID))
}

func fortyOneCommentMarker(commentID uuid.UUID) string {
	return fmt.Sprintf("<!-- fortyone:comment:%s -->", commentID.String())
}

func buildFortyOneBotCommentBody(authorName, body string, commentID uuid.UUID) string {
	return fmt.Sprintf("**%s** commented via FortyOne:\n\n%s", authorName, buildFortyOneUserCommentBody(body, commentID))
}

func stripFortyOneCommentMarker(body string) string {
	return strings.TrimSpace(fortyOneCommentMarkerPattern.ReplaceAllString(body, ""))
}

func stripFortyOneBotAttribution(body string) string {
	if match := fortyOneCommentPattern.FindStringSubmatch(strings.TrimSpace(body)); match != nil {
		return match[2]
	}
	return body
}
