package github

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

const (
	maxGitHubOAuthCodeBytes     = 4 << 10
	maxGitHubOAuthResponseBytes = 64 << 10
)

var (
	ErrGitHubOAuthCodeInvalid         = errors.New("github oauth code is invalid")
	ErrGitHubOAuthCodeRejected        = errors.New("github oauth code was rejected")
	ErrGitHubOAuthExchangeUnavailable = errors.New("github oauth exchange is unavailable")
	ErrGitHubOAuthNotConfigured       = errors.New("github oauth is not configured")
)

func (s *Service) exchangeOAuthCode(ctx context.Context, code string) (string, error) {
	code = strings.TrimSpace(code)
	if strings.TrimSpace(s.cfg.ClientID) == "" || strings.TrimSpace(s.cfg.ClientSecret) == "" || s.httpClient == nil {
		return "", ErrGitHubOAuthNotConfigured
	}
	if code == "" || len(code) > maxGitHubOAuthCodeBytes {
		return "", ErrGitHubOAuthCodeInvalid
	}
	values := url.Values{}
	values.Set("client_id", s.cfg.ClientID)
	values.Set("client_secret", s.cfg.ClientSecret)
	values.Set("code", code)
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://github.com/login/oauth/access_token",
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", ErrGitHubOAuthExchangeUnavailable
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError {
			return "", ErrGitHubOAuthExchangeUnavailable
		}
		return "", ErrGitHubOAuthCodeRejected
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxGitHubOAuthResponseBytes+1))
	if err != nil || len(body) > maxGitHubOAuthResponseBytes {
		return "", ErrGitHubOAuthExchangeUnavailable
	}
	defer clear(body)
	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&result); err != nil {
		return "", ErrGitHubOAuthExchangeUnavailable
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return "", ErrGitHubOAuthExchangeUnavailable
	}
	token := strings.TrimSpace(result.AccessToken)
	if result.Error != "" || token == "" {
		if result.Error != "" {
			return "", ErrGitHubOAuthCodeRejected
		}
		return "", ErrGitHubOAuthExchangeUnavailable
	}
	return token, nil
}

func (s *Service) moveStoryByRule(ctx context.Context, workspaceID, teamID, storyID uuid.UUID, eventKey string, baseBranch *string) error {
	settings, err := s.GetTeamSettings(ctx, workspaceID, teamID)
	if err != nil {
		return err
	}
	var statusID *uuid.UUID
	for _, rule := range settings.Rules {
		if rule.EventKey != eventKey || !rule.IsActive {
			continue
		}
		if rule.BaseBranchPattern == nil {
			statusID = rule.TargetStatusID
			continue
		}
		if baseBranch != nil && matchBranchPattern(*rule.BaseBranchPattern, *baseBranch) {
			statusID = rule.TargetStatusID
			break
		}
	}
	if statusID == nil {
		return nil
	}
	story, err := s.stories.Get(ctx, storyID, workspaceID)
	if err != nil {
		return err
	}
	if story.Status != nil && *story.Status == *statusID {
		return nil
	}
	return s.stories.UpdateExternalWithReason(ctx, s.cfg.GitHubUserID, storyID, workspaceID, map[string]any{
		"status_id": *statusID,
	}, githubWorkflowAutomationReason())
}

func githubIssueSyncReason() string {
	return "GitHub issue details changed, so FortyOne synced the linked story."
}

func githubAssigneeSyncReason() string {
	return "GitHub assignee changed, so FortyOne synced the linked story assignment."
}

func githubLabelSyncReason() string {
	return "GitHub labels changed, so FortyOne synced the linked story labels."
}

func githubWorkflowAutomationReason() string {
	return "A GitHub workflow automation matched a repository event and moved this story to the configured status."
}

func (s *Service) seedDefaultWorkflowRules(ctx context.Context, workspaceID, teamID uuid.UUID) (CoreTeamGitHubSettings, error) {
	statuses, err := s.repo.ListTeamStatuses(ctx, teamID)
	if err != nil {
		return CoreTeamGitHubSettings{}, err
	}
	findCategory := func(category string) *uuid.UUID {
		for _, status := range statuses {
			if status.Category == category {
				id := status.ID
				return &id
			}
		}
		return nil
	}
	findReview := func() *uuid.UUID {
		for _, status := range statuses {
			if strings.Contains(strings.ToLower(status.Name), "review") {
				id := status.ID
				return &id
			}
		}
		return nil
	}
	rules := []CoreWorkflowRuleInput{
		{EventKey: EventDraftPROpen, TargetStatusID: nil, IsActive: true},
		{EventKey: EventPROpen, TargetStatusID: findCategory("started"), IsActive: true},
		{EventKey: EventPRReviewActivity, TargetStatusID: findReview(), IsActive: true},
		{EventKey: EventPRReadyForMerge, TargetStatusID: nil, IsActive: true},
		{EventKey: EventPRMerge, TargetStatusID: findCategory("completed"), IsActive: true},
		{EventKey: EventIssueOpen, TargetStatusID: findCategory("unstarted"), IsActive: true},
		{EventKey: EventIssueReopen, TargetStatusID: findCategory("unstarted"), IsActive: true},
		{EventKey: EventIssueClose, TargetStatusID: findCategory("completed"), IsActive: true},
		{EventKey: EventCommitClose, TargetStatusID: findCategory("completed"), IsActive: true},
	}
	return s.repo.ReplaceTeamWorkflowSettings(ctx, workspaceID, teamID, rules)
}

func extractStoryRefs(values ...string) []string {
	items := make([]string, 0)
	seen := map[string]struct{}{}
	for _, value := range values {
		for _, match := range refPattern.FindAllString(value, -1) {
			normalized := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(match), "-", ""), " ", ""))
			if normalized == "" {
				continue
			}
			if _, ok := seen[normalized]; ok {
				continue
			}
			seen[normalized] = struct{}{}
			items = append(items, normalized)
		}
	}
	return items
}

func matchBranchPattern(pattern, branch string) bool {
	if pattern == branch {
		return true
	}
	if strings.HasSuffix(pattern, "/*") {
		return strings.HasPrefix(branch, strings.TrimSuffix(pattern, "*"))
	}
	return false
}

func (s *Service) upsertStoryLink(ctx context.Context, workspaceID, storyID, repositoryID uuid.UUID, externalType string, githubID int64, githubNumber int, refName *string, url, title, state string, metadata any) (bool, error) {
	_, _, err := s.repo.FindStoryLink(ctx, repositoryID, externalType, githubID, refName)
	linkCreated := errors.Is(err, sql.ErrNoRows)
	if err != nil && !linkCreated {
		return false, err
	}
	if err := s.repo.UpsertStoryLink(ctx, workspaceID, storyID, repositoryID, externalType, githubID, githubNumber, refName, url, title, state, metadata); err != nil {
		return false, err
	}
	return linkCreated, nil
}

func (s *Service) recordLinkActivity(ctx context.Context, workspaceID, storyID uuid.UUID, field, currentValue, targetURL string) error {
	if s.stories == nil {
		return errors.New("stories service is not configured")
	}

	return s.stories.RecordActivity(ctx, storyActivity{
		StoryID:      storyID,
		Type:         "link",
		Field:        field,
		CurrentValue: currentValue,
		NewValue:     targetURL,
		UserID:       s.cfg.GitHubUserID,
		WorkspaceID:  workspaceID,
	})
}

func githubIssueStoryLinkTitle(issueNumber int) *string {
	title := fmt.Sprintf("GitHub issue #%d", issueNumber)
	return &title
}

func githubPullRequestStoryLinkTitle(prNumber int) *string {
	title := fmt.Sprintf("GitHub PR #%d", prNumber)
	return &title
}

func githubBranchStoryLinkTitle(branchName string) *string {
	title := fmt.Sprintf("GitHub branch %s", branchName)
	return &title
}
