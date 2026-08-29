// Package slackadapter contains composition-only bridges between Slack's
// caller-owned capability contracts and sibling-module use cases.
package slackadapter

import (
	"context"
	"errors"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	"github.com/google/uuid"
)

type RequestBackend interface {
	UpsertPending(context.Context, integrationrequests.CoreUpsertRequestInput) (integrationrequests.CoreIntegrationRequest, error)
	GetForUser(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (integrationrequests.CoreIntegrationRequest, error)
	BindProviderThread(context.Context, integrationrequests.CoreBindProviderThreadInput) (integrationrequests.CoreProviderThread, error)
	HasAuthorizedProviderThread(context.Context, integrationrequests.CoreProviderThreadMatchInput) (bool, error)
	HasCurrentProviderThread(context.Context, integrationrequests.CoreProviderThreadLookupInput) (bool, error)
	FindProviderThread(context.Context, uuid.UUID, uuid.UUID, string) (integrationrequests.CoreProviderThread, error)
	IngestInboundProviderComment(context.Context, integrationrequests.CoreInboundProviderCommentInput) (bool, error)
}

type RequestStore struct {
	backend RequestBackend
}

func NewRequestStore(backend RequestBackend) *RequestStore {
	if backend == nil {
		return nil
	}
	return &RequestStore{backend: backend}
}

func (adapter *RequestStore) UpsertPending(ctx context.Context, input slack.UpsertIntegrationRequestInput) (slack.IntegrationRequest, error) {
	request, err := adapter.backend.UpsertPending(ctx, integrationrequests.CoreUpsertRequestInput{
		WorkspaceID: input.WorkspaceID, TeamID: input.TeamID, Provider: input.Provider,
		SourceType: input.SourceType, SourceExternalID: input.SourceExternalID,
		SourceNumber: input.SourceNumber, SourceURL: input.SourceURL, Title: input.Title,
		Description: input.Description, StatusID: input.StatusID, Priority: input.Priority,
		AssigneeID: input.AssigneeID, EstimateValue: input.EstimateValue,
		EstimatedDurationMinutes: input.EstimatedDurationMinutes,
		MinimumFocusBlockMinutes: input.MinimumFocusBlockMinutes,
		ObjectiveID:              input.ObjectiveID, KeyResultID: input.KeyResultID, SprintID: input.SprintID,
		StartDate: input.StartDate, EndDate: input.EndDate,
		LabelIDs: append([]uuid.UUID(nil), input.LabelIDs...), Metadata: cloneAnyMap(input.Metadata),
		CreatedByUserID: input.CreatedByUserID,
	})
	return mapIntegrationRequest(request), mapRequestError(err)
}

func (adapter *RequestStore) GetForUser(ctx context.Context, workspaceID, requestID, userID uuid.UUID) (slack.IntegrationRequest, error) {
	request, err := adapter.backend.GetForUser(ctx, workspaceID, requestID, userID)
	return mapIntegrationRequest(request), mapRequestError(err)
}

func (adapter *RequestStore) BindProviderThread(ctx context.Context, input slack.BindProviderThreadInput) (slack.ProviderThread, error) {
	thread, err := adapter.backend.BindProviderThread(ctx, integrationrequests.CoreBindProviderThreadInput{
		WorkspaceID: input.WorkspaceID, IntegrationRequestID: input.IntegrationRequestID,
		Provider: input.Provider, ExternalWorkspaceID: input.ExternalWorkspaceID,
		InstallationGeneration: input.InstallationGeneration, ExternalChannelID: input.ExternalChannelID,
		ExternalThreadID: input.ExternalThreadID, ExternalSourceMessageID: input.ExternalSourceMessageID,
		SourceURL: input.SourceURL,
	})
	return mapProviderThread(thread), mapRequestError(err)
}

func (adapter *RequestStore) HasAuthorizedProviderThread(ctx context.Context, input slack.ProviderThreadMatchInput) (bool, error) {
	matched, err := adapter.backend.HasAuthorizedProviderThread(ctx, integrationrequests.CoreProviderThreadMatchInput{
		WorkspaceID: input.WorkspaceID, UserID: input.UserID, Provider: input.Provider,
		ExternalWorkspaceID: input.ExternalWorkspaceID, InstallationGeneration: input.InstallationGeneration,
		ExternalChannelID: input.ExternalChannelID, ExternalThreadID: input.ExternalThreadID,
	})
	return matched, mapRequestError(err)
}

func (adapter *RequestStore) HasCurrentProviderThread(ctx context.Context, input slack.ProviderThreadLookupInput) (bool, error) {
	matched, err := adapter.backend.HasCurrentProviderThread(ctx, integrationrequests.CoreProviderThreadLookupInput{
		WorkspaceID: input.WorkspaceID, Provider: input.Provider,
		ExternalWorkspaceID: input.ExternalWorkspaceID, InstallationGeneration: input.InstallationGeneration,
		ExternalChannelID: input.ExternalChannelID, ExternalThreadID: input.ExternalThreadID,
	})
	return matched, mapRequestError(err)
}

func (adapter *RequestStore) FindProviderThread(ctx context.Context, workspaceID, requestID uuid.UUID, provider string) (slack.ProviderThread, error) {
	thread, err := adapter.backend.FindProviderThread(ctx, workspaceID, requestID, provider)
	return mapProviderThread(thread), mapRequestError(err)
}

func (adapter *RequestStore) IngestInboundProviderComment(ctx context.Context, input slack.InboundProviderCommentInput) (bool, error) {
	handled, err := adapter.backend.IngestInboundProviderComment(ctx, integrationrequests.CoreInboundProviderCommentInput{
		Provider: input.Provider, ExternalWorkspaceID: input.ExternalWorkspaceID,
		InstallationGeneration: input.InstallationGeneration, ExternalChannelID: input.ExternalChannelID,
		ExternalThreadID: input.ExternalThreadID, ExternalMessageID: input.ExternalMessageID,
		ExternalAuthorID: input.ExternalAuthorID, AuthorUserID: input.AuthorUserID,
		Body: input.Body, CreatedAt: input.CreatedAt,
	})
	return handled, mapRequestError(err)
}

func mapIntegrationRequest(request integrationrequests.CoreIntegrationRequest) slack.IntegrationRequest {
	return slack.IntegrationRequest{
		ID: request.ID, WorkspaceID: request.WorkspaceID, TeamID: request.TeamID,
		Provider: request.Provider, SourceType: request.SourceType, SourceExternalID: request.SourceExternalID,
		SourceNumber: request.SourceNumber, SourceURL: request.SourceURL, Title: request.Title,
		Description: request.Description, StatusID: request.StatusID, Priority: request.Priority,
		AssigneeID: request.AssigneeID, EstimateValue: request.EstimateValue,
		EstimatedDurationMinutes: request.EstimatedDurationMinutes,
		MinimumFocusBlockMinutes: request.MinimumFocusBlockMinutes,
		ObjectiveID:              request.ObjectiveID, KeyResultID: request.KeyResultID, SprintID: request.SprintID,
		StartDate: request.StartDate, EndDate: request.EndDate,
		LabelIDs: append([]uuid.UUID(nil), request.LabelIDs...), Status: request.Status,
		Metadata: cloneAnyMap(request.Metadata), AcceptedStoryID: request.AcceptedStoryID,
		CreatedByUserID: request.CreatedByUserID, CreatedAt: request.CreatedAt, UpdatedAt: request.UpdatedAt,
	}
}

func mapProviderThread(thread integrationrequests.CoreProviderThread) slack.ProviderThread {
	return slack.ProviderThread{
		ID: thread.ID, WorkspaceID: thread.WorkspaceID, IntegrationRequestID: thread.IntegrationRequestID,
		TeamID: thread.TeamID, AcceptedStoryID: thread.AcceptedStoryID, Provider: thread.Provider,
		ExternalWorkspaceID: thread.ExternalWorkspaceID, InstallationGeneration: thread.InstallationGeneration,
		ExternalChannelID: thread.ExternalChannelID, ExternalThreadID: thread.ExternalThreadID,
		ExternalSourceMessageID: thread.ExternalSourceMessageID, SourceURL: thread.SourceURL,
		RequestTitle: thread.RequestTitle, CreatedAt: thread.CreatedAt, UpdatedAt: thread.UpdatedAt,
	}
}

func mapRequestError(err error) error {
	if errors.Is(err, integrationrequests.ErrProviderThreadNotFound) {
		return errors.Join(slack.ErrProviderThreadNotFound, err)
	}
	return err
}

func cloneAnyMap(values map[string]any) map[string]any {
	if values == nil {
		return nil
	}
	cloned := make(map[string]any, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

var (
	_ slack.RequestStore    = (*RequestStore)(nil)
	_ slack.SlackThreadSync = (*RequestStore)(nil)
)
