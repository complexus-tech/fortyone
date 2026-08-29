// Package githubadapter contains composition-only bridges between the GitHub
// module's caller-owned ports and existing sibling-module use cases.
package githubadapter

import (
	"context"

	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

type StoryBackend interface {
	Get(ctx context.Context, id, workspaceID uuid.UUID) (stories.CoreSingleStory, error)
	UpdateExternalWithReason(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, updates map[string]any, reason string) error
	RecordActivity(ctx context.Context, activity stories.CoreActivity) error
	CreateCommentExternal(ctx context.Context, actorID, workspaceID uuid.UUID, comment stories.CoreNewComment) (stories.CoreComment, error)
}

type storyService struct {
	backend StoryBackend
}

func NewStoryService(backend StoryBackend) github.StoryService {
	if backend == nil {
		return nil
	}
	return storyService{backend: backend}
}

func (adapter storyService) Get(ctx context.Context, id, workspaceID uuid.UUID) (stories.CoreSingleStory, error) {
	return adapter.backend.Get(ctx, id, workspaceID)
}

func (adapter storyService) UpdateExternalWithReason(
	ctx context.Context,
	actorID, storyID, workspaceID uuid.UUID,
	updates map[string]any,
	reason string,
) error {
	return adapter.backend.UpdateExternalWithReason(ctx, actorID, storyID, workspaceID, updates, reason)
}

func (adapter storyService) RecordActivity(ctx context.Context, activity github.StoryActivity) error {
	return adapter.backend.RecordActivity(ctx, stories.CoreActivity{
		StoryID:      activity.StoryID,
		UserID:       activity.UserID,
		Type:         activity.Type,
		Field:        activity.Field,
		CurrentValue: activity.CurrentValue,
		OldValue:     activity.OldValue,
		NewValue:     activity.NewValue,
		Reason:       activity.Reason,
		WorkspaceID:  activity.WorkspaceID,
	})
}

func (adapter storyService) CreateCommentExternal(
	ctx context.Context,
	actorID, workspaceID uuid.UUID,
	comment github.NewStoryComment,
) (stories.CoreComment, error) {
	return adapter.backend.CreateCommentExternal(ctx, actorID, workspaceID, stories.CoreNewComment{
		StoryID:  comment.StoryID,
		Parent:   comment.Parent,
		UserID:   comment.UserID,
		Comment:  comment.Comment,
		Mentions: append([]uuid.UUID(nil), comment.Mentions...),
	})
}

type RequestBackend interface {
	UpsertPending(ctx context.Context, input integrationrequests.CoreUpsertRequestInput) (integrationrequests.CoreIntegrationRequest, error)
	Get(ctx context.Context, workspaceID, requestID uuid.UUID) (integrationrequests.CoreIntegrationRequest, error)
}

type requestStore struct {
	backend RequestBackend
}

func NewRequestStore(backend RequestBackend) github.RequestStore {
	if backend == nil {
		return nil
	}
	return requestStore{backend: backend}
}

func (adapter requestStore) UpsertPending(
	ctx context.Context,
	input github.UpsertIntegrationRequestInput,
) (github.IntegrationRequest, error) {
	request, err := adapter.backend.UpsertPending(ctx, integrationrequests.CoreUpsertRequestInput{
		WorkspaceID:      input.WorkspaceID,
		TeamID:           input.TeamID,
		Provider:         input.Provider,
		SourceType:       input.SourceType,
		SourceExternalID: input.SourceExternalID,
		SourceNumber:     input.SourceNumber,
		SourceURL:        input.SourceURL,
		Title:            input.Title,
		Description:      input.Description,
		Priority:         input.Priority,
		Metadata:         cloneMetadata(input.Metadata),
		CreatedByUserID:  input.CreatedByUserID,
	})
	if err != nil {
		return github.IntegrationRequest{}, err
	}
	return mapIntegrationRequest(request), nil
}

func (adapter requestStore) Get(
	ctx context.Context,
	workspaceID, requestID uuid.UUID,
) (github.IntegrationRequest, error) {
	request, err := adapter.backend.Get(ctx, workspaceID, requestID)
	if err != nil {
		return github.IntegrationRequest{}, err
	}
	return mapIntegrationRequest(request), nil
}

// ProviderAccepter adapts the integration-requests callback contract at the
// composition root instead of making GitHub business logic import that module.
type ProviderAccepter struct {
	Service *github.Service
}

func (adapter ProviderAccepter) AcceptIntegrationRequest(
	ctx context.Context,
	request integrationrequests.CoreIntegrationRequest,
	story integrationrequests.Story,
) error {
	return adapter.Service.AcceptIntegrationRequest(ctx, mapIntegrationRequest(request), stories.CoreSingleStory{
		ID:          story.ID,
		SequenceID:  story.SequenceID,
		Team:        story.TeamID,
		TeamCode:    story.TeamCode,
		Title:       story.Title,
		Description: story.Description,
		Status:      story.StatusID,
		Priority:    story.Priority,
		Assignee:    story.AssigneeID,
		Reporter:    story.ReporterID,
		EndDate:     story.EndDate,
		CreatedAt:   story.CreatedAt,
		UpdatedAt:   story.UpdatedAt,
	})
}

func mapIntegrationRequest(request integrationrequests.CoreIntegrationRequest) github.IntegrationRequest {
	return github.IntegrationRequest{
		WorkspaceID:      request.WorkspaceID,
		Provider:         request.Provider,
		SourceType:       request.SourceType,
		SourceExternalID: request.SourceExternalID,
		SourceNumber:     request.SourceNumber,
		SourceURL:        request.SourceURL,
		Title:            request.Title,
		Metadata:         cloneMetadata(request.Metadata),
	}
}

func cloneMetadata(metadata map[string]any) map[string]any {
	if metadata == nil {
		return nil
	}
	cloned := make(map[string]any, len(metadata))
	for key, value := range metadata {
		cloned[key] = value
	}
	return cloned
}

var (
	_ github.StoryService                  = storyService{}
	_ github.RequestStore                  = requestStore{}
	_ integrationrequests.ProviderAccepter = ProviderAccepter{}
)
