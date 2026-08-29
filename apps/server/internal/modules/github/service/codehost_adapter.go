package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/complexus-tech/projects-api/internal/platform/codehost"
	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	githubsdk "github.com/google/go-github/v72/github"
)

func (s *Service) Provider() integrations.ProviderKey {
	return githubWebhookProvider
}

func (s *Service) Capabilities() codehost.Capabilities {
	return codehost.Capabilities{
		codehost.CapabilityInstallationAuth:     true,
		codehost.CapabilityRepositoryCatalog:    true,
		codehost.CapabilityWorkItemWriter:       true,
		codehost.CapabilityCommentWriter:        true,
		codehost.CapabilityWebhookNormalization: true,
	}
}

func (s *Service) Authorize(ctx context.Context, installation codehost.InstallationRef) error {
	externalID, err := validateGitHubInstallationRef(installation)
	if err != nil {
		return err
	}
	if _, err := s.getInstallationToken(ctx, externalID); err != nil {
		return classifyGitHubProviderError(err)
	}
	return nil
}

func (s *Service) ListRepositories(
	ctx context.Context,
	installation codehost.InstallationRef,
	cursor codehost.Cursor,
) (codehost.RepositoryPage, error) {
	externalID, err := validateGitHubInstallationRef(installation)
	if err != nil {
		return codehost.RepositoryPage{}, err
	}
	if err := codehost.ValidateCursor(cursor); err != nil {
		return codehost.RepositoryPage{}, err
	}
	page := 1
	if strings.TrimSpace(cursor.Value) != "" {
		page, err = strconv.Atoi(cursor.Value)
		if err != nil || page < 1 {
			return codehost.RepositoryPage{}, codehost.ErrInvalidInput
		}
	}
	client, err := s.newInstallationClient(ctx, externalID)
	if err != nil {
		return codehost.RepositoryPage{}, classifyGitHubProviderError(err)
	}
	responseRepositories, response, err := client.Apps.ListRepos(ctx, &githubsdk.ListOptions{Page: page, PerPage: cursor.Limit})
	if err != nil {
		return codehost.RepositoryPage{}, classifyGitHubProviderError(err)
	}
	repositories := make([]codehost.RepositoryRef, 0, len(responseRepositories.Repositories))
	for _, repository := range responseRepositories.Repositories {
		if repository == nil {
			continue
		}
		repositories = append(repositories, mapGitHubRepository(repository))
	}
	nextCursor := ""
	if response != nil && response.NextPage > 0 {
		nextCursor = strconv.Itoa(response.NextPage)
	}
	return codehost.RepositoryPage{Repositories: repositories, NextCursor: nextCursor}, nil
}

func (s *Service) CreateWorkItem(
	ctx context.Context,
	installation codehost.InstallationRef,
	command codehost.CreateWorkItem,
) (codehost.WorkItem, error) {
	externalID, err := validateGitHubInstallationRef(installation)
	if err != nil {
		return codehost.WorkItem{}, err
	}
	if err := codehost.ValidateCreateWorkItem(command); err != nil {
		return codehost.WorkItem{}, err
	}
	issue, err := s.createIssue(ctx, externalID, command.Repository.Owner, command.Repository.Name, command.Title, command.Body)
	if err != nil {
		return codehost.WorkItem{}, classifyGitHubProviderError(err)
	}
	state, err := codehost.ParseWorkItemState(issue.State)
	if err != nil {
		return codehost.WorkItem{}, err
	}
	return codehost.WorkItem{
		ExternalID: strconv.FormatInt(issue.ID, 10),
		Number:     int64(issue.Number),
		Repository: command.Repository,
		Title:      issue.Title,
		Body:       command.Body,
		State:      state,
		WebURL:     issue.HTMLURL,
	}, nil
}

func (s *Service) AddComment(
	ctx context.Context,
	installation codehost.InstallationRef,
	command codehost.AddComment,
) (codehost.Comment, error) {
	externalID, err := validateGitHubInstallationRef(installation)
	if err != nil {
		return codehost.Comment{}, err
	}
	if err := codehost.ValidateAddComment(command); err != nil {
		return codehost.Comment{}, err
	}
	client, err := s.newInstallationClient(ctx, externalID)
	if err != nil {
		return codehost.Comment{}, classifyGitHubProviderError(err)
	}
	comment, _, err := client.Issues.CreateComment(
		ctx,
		command.WorkItem.Repository.Owner,
		command.WorkItem.Repository.Name,
		int(command.WorkItem.Number),
		&githubsdk.IssueComment{Body: githubsdk.Ptr(command.Body)},
	)
	if err != nil {
		return codehost.Comment{}, classifyGitHubProviderError(err)
	}
	return codehost.Comment{
		ExternalID:  strconv.FormatInt(comment.GetID(), 10),
		WorkItem:    command.WorkItem,
		AuthorID:    strconv.FormatInt(comment.GetUser().GetID(), 10),
		AuthorLogin: comment.GetUser().GetLogin(),
		Body:        comment.GetBody(),
		WebURL:      comment.GetHTMLURL(),
		CreatedAt:   comment.GetCreatedAt().Time,
	}, nil
}

func (s *Service) NormalizeWebhook(
	_ context.Context,
	deliveryID, eventType string,
	body []byte,
) (codehost.NormalizedEvent, error) {
	if strings.TrimSpace(deliveryID) == "" || len(body) == 0 {
		return codehost.NormalizedEvent{}, codehost.ErrInvalidInput
	}
	if !supportedGitHubWebhookEvent(eventType) {
		return codehost.NormalizedEvent{}, codehost.ErrCapabilityUnsupported
	}
	var payload webhookEnvelope
	if err := json.Unmarshal(body, &payload); err != nil {
		return codehost.NormalizedEvent{}, codehost.ErrInvalidInput
	}
	repository := codehost.RepositoryRef{
		ExternalID:    strconv.FormatInt(payload.Repository.ID, 10),
		Owner:         payload.Repository.Owner.Login,
		Name:          payload.Repository.Name,
		FullName:      payload.Repository.FullName,
		WebURL:        payload.Repository.HTMLURL,
		DefaultBranch: payload.Repository.DefaultBranch,
		Private:       payload.Repository.Private,
	}
	if err := codehost.ValidateRepository(repository); err != nil {
		return codehost.NormalizedEvent{}, err
	}
	event := codehost.NormalizedEvent{
		Provider:             githubWebhookProvider,
		DeliveryID:           strings.TrimSpace(deliveryID),
		Action:               strings.TrimSpace(payload.Action),
		ExternalRepositoryID: strconv.FormatInt(payload.Repository.ID, 10),
		ExternalActorID:      strconv.FormatInt(payload.Sender.ID, 10),
	}
	switch eventType {
	case "issues":
		if payload.Issue.ID <= 0 || payload.Issue.Number <= 0 {
			return codehost.NormalizedEvent{}, codehost.ErrInvalidInput
		}
		state, err := codehost.ParseWorkItemState(payload.Issue.State)
		if err != nil {
			return codehost.NormalizedEvent{}, err
		}
		event.Kind = codehost.EventWorkItemChanged
		event.WorkItem = &codehost.WorkItem{
			ExternalID: strconv.FormatInt(payload.Issue.ID, 10), Number: int64(payload.Issue.Number),
			Repository: repository, Title: payload.Issue.Title, Body: payload.Issue.Body,
			State: state, WebURL: payload.Issue.HTMLURL,
		}
	case "issue_comment":
		if payload.Issue.ID <= 0 || payload.Issue.Number <= 0 || payload.Comment.ID <= 0 {
			return codehost.NormalizedEvent{}, codehost.ErrInvalidInput
		}
		state, err := codehost.ParseWorkItemState(payload.Issue.State)
		if err != nil {
			return codehost.NormalizedEvent{}, err
		}
		event.Kind = codehost.EventCommentCreated
		event.Comment = &codehost.Comment{
			ExternalID: strconv.FormatInt(payload.Comment.ID, 10),
			WorkItem: codehost.WorkItem{
				ExternalID: strconv.FormatInt(payload.Issue.ID, 10), Number: int64(payload.Issue.Number),
				Repository: repository, Title: payload.Issue.Title, Body: payload.Issue.Body,
				State: state, WebURL: payload.Issue.HTMLURL,
			},
			AuthorID: strconv.FormatInt(payload.Comment.User.ID, 10), AuthorLogin: payload.Comment.User.Login,
			Body: payload.Comment.Body, WebURL: payload.Comment.HTMLURL, CreatedAt: payload.Comment.CreatedAt,
		}
	case "push":
		event.Kind = codehost.EventPush
	case "pull_request", "pull_request_review", "check_run":
		event.Kind = codehost.EventMergeRequest
	case "create":
		event.Kind = codehost.EventPush
	default:
		return codehost.NormalizedEvent{}, codehost.ErrCapabilityUnsupported
	}
	return event, nil
}

func validateGitHubInstallationRef(installation codehost.InstallationRef) (int64, error) {
	if err := codehost.ValidateInstallation(installation); err != nil || installation.Provider != githubWebhookProvider {
		return 0, codehost.ErrInvalidInput
	}
	externalID, err := strconv.ParseInt(installation.ExternalInstallationID, 10, 64)
	if err != nil || externalID <= 0 {
		return 0, codehost.ErrInvalidInput
	}
	return externalID, nil
}

func mapGitHubRepository(repository *githubsdk.Repository) codehost.RepositoryRef {
	return codehost.RepositoryRef{
		ExternalID:    strconv.FormatInt(repository.GetID(), 10),
		Owner:         repository.GetOwner().GetLogin(),
		Name:          repository.GetName(),
		FullName:      repository.GetFullName(),
		WebURL:        repository.GetHTMLURL(),
		DefaultBranch: repository.GetDefaultBranch(),
		Private:       repository.GetPrivate(),
		Archived:      repository.GetArchived(),
	}
}

func classifyGitHubProviderError(err error) error {
	var rateLimitError *githubsdk.RateLimitError
	if errors.As(err, &rateLimitError) {
		return codehost.ErrRateLimited
	}
	var abuseRateLimitError *githubsdk.AbuseRateLimitError
	if errors.As(err, &abuseRateLimitError) {
		return codehost.ErrRateLimited
	}
	var providerError *githubsdk.ErrorResponse
	if !errors.As(err, &providerError) || providerError.Response == nil {
		return err
	}
	switch providerError.Response.StatusCode {
	case 401:
		return fmt.Errorf("%w: GitHub rejected the credential", codehost.ErrAuthentication)
	case 403:
		return fmt.Errorf("%w: GitHub denied the installation grant", codehost.ErrGrantRevoked)
	case 404:
		return codehost.ErrNotFound
	case 429:
		return codehost.ErrRateLimited
	default:
		return err
	}
}

var (
	_ codehost.Adapter                   = (*Service)(nil)
	_ codehost.InstallationAuthenticator = (*Service)(nil)
	_ codehost.RepositoryCatalog         = (*Service)(nil)
	_ codehost.WorkItemWriter            = (*Service)(nil)
	_ codehost.CommentWriter             = (*Service)(nil)
	_ codehost.WebhookNormalizer         = (*Service)(nil)
)
