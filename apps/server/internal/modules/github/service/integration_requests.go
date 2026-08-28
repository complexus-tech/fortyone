package github

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

func (s *Service) upsertIssueIntegrationRequest(ctx context.Context, repository repositoryRecord, teamID uuid.UUID, payload webhookEnvelope) error {
	if s.requests == nil {
		return errors.New("integration request service is not configured")
	}
	description := githubString(payload.Issue.Body)
	sourceURL := payload.Issue.HTMLURL
	sourceNumber := payload.Issue.Number
	createdBy, err := s.resolveActorID(ctx, payload.Sender.ID)
	if err != nil {
		return err
	}
	labelNames := make([]string, 0, len(payload.Issue.Labels))
	for _, label := range payload.Issue.Labels {
		labelNames = append(labelNames, label.Name)
	}
	priority := githubPriorityFromLabelNames(labelNames)
	_, err = s.requests.UpsertPending(ctx, upsertIntegrationRequestInput{
		WorkspaceID:      repository.WorkspaceID,
		TeamID:           teamID,
		Provider:         providerGitHub,
		SourceType:       requestSourceTypeIssue,
		SourceExternalID: strconv.FormatInt(payload.Issue.ID, 10),
		SourceNumber:     &sourceNumber,
		SourceURL:        &sourceURL,
		Title:            payload.Issue.Title,
		Description:      description,
		Priority:         priority,
		CreatedByUserID:  &createdBy,
		Metadata: map[string]any{
			"repository_id":          repository.ID.String(),
			"repository_external_id": payload.Repository.ID,
			"repository_full_name":   repository.FullName,
			"issue_id":               payload.Issue.ID,
			"issue_number":           payload.Issue.Number,
			"issue_state":            payload.Issue.State,
			"priority":               priority,
			"priority_labels":        labelNames,
			"sender_id":              payload.Sender.ID,
			"sender_login":           payload.Sender.Login,
		},
	})
	return err
}

func (s *Service) AcceptIntegrationRequest(ctx context.Context, request integrationRequest, story singleStory) error {
	if request.Provider != providerGitHub || request.SourceType != requestSourceTypeIssue {
		return nil
	}
	repositoryID, err := metadataUUID(request.Metadata, "repository_id")
	if err != nil {
		return err
	}
	githubID, err := strconv.ParseInt(request.SourceExternalID, 10, 64)
	if err != nil {
		return err
	}
	githubNumber := 0
	if request.SourceNumber != nil {
		githubNumber = *request.SourceNumber
	}
	sourceURL := ""
	if request.SourceURL != nil {
		sourceURL = *request.SourceURL
	}
	state := metadataString(request.Metadata, "issue_state")
	if state == "" {
		state = "open"
	}
	metadata := map[string]any{}
	for key, value := range request.Metadata {
		metadata[key] = value
	}
	if err := s.repo.UpsertStoryLink(ctx, request.WorkspaceID, story.ID, repositoryID, "issue", githubID, githubNumber, nil, sourceURL, request.Title, state, metadata); err != nil {
		return err
	}
	if sourceURL != "" {
		if err := s.repo.EnsureStoryLink(ctx, story.ID, githubIssueStoryLinkTitle(githubNumber), sourceURL); err != nil {
			return err
		}
	}
	repository, err := s.repo.FindRepositoryByID(ctx, request.WorkspaceID, repositoryID)
	if err != nil {
		return err
	}
	if err := s.ensureIssueImportComment(ctx, repository, githubNumber, story); err != nil {
		return err
	}
	return s.recordLinkActivity(ctx, request.WorkspaceID, story.ID, "github_issue", fmt.Sprintf("issue %d", githubNumber), sourceURL)
}

func metadataUUID(metadata map[string]any, key string) (uuid.UUID, error) {
	value, ok := metadata[key]
	if !ok {
		return uuid.Nil, fmt.Errorf("metadata %q is required", key)
	}
	text, ok := value.(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("metadata %q must be a string", key)
	}
	return uuid.Parse(text)
}

func metadataString(metadata map[string]any, key string) string {
	value, ok := metadata[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return text
}
