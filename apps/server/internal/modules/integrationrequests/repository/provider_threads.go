package integrationrequestsrepository

import (
	"context"
	"fmt"
	"strings"

	integrationrequestdomain "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/domain"
	integrationrequestssql "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/repository/sqlc"
	"github.com/google/uuid"
)

func (r *Repo) BindProviderThread(ctx context.Context, input integrationrequestdomain.BindProviderThreadInput) (integrationrequestdomain.ProviderThread, error) {
	if err := r.configured(); err != nil {
		return integrationrequestdomain.ProviderThread{}, err
	}
	row, err := r.queries.BindIntegrationRequestProviderThread(ctx, integrationrequestssql.BindIntegrationRequestProviderThreadParams{
		WorkspaceID: input.WorkspaceID, RequestID: input.IntegrationRequestID,
		Provider: strings.TrimSpace(input.Provider), ExternalWorkspaceID: strings.TrimSpace(input.ExternalWorkspaceID),
		InstallationGeneration: input.InstallationGeneration, ExternalChannelID: strings.TrimSpace(input.ExternalChannelID),
		ExternalThreadID:        strings.TrimSpace(input.ExternalThreadID),
		ExternalSourceMessageID: optionalString(input.ExternalSourceMessageID), SourceURL: input.SourceURL,
	})
	if err != nil {
		return integrationrequestdomain.ProviderThread{}, mapNotFound("bind integration request provider thread", err)
	}
	return providerThreadFromRecord(providerThreadRecord(row)), nil
}

func (r *Repo) HasAuthorizedProviderThread(ctx context.Context, input integrationrequestdomain.ProviderThreadMatchInput) (bool, error) {
	if err := r.configured(); err != nil {
		return false, err
	}
	exists, err := r.queries.HasAuthorizedIntegrationRequestProviderThread(ctx, integrationrequestssql.HasAuthorizedIntegrationRequestProviderThreadParams{
		WorkspaceID: input.WorkspaceID, Provider: strings.TrimSpace(input.Provider),
		ExternalWorkspaceID:    strings.TrimSpace(input.ExternalWorkspaceID),
		InstallationGeneration: uuidPointer(input.InstallationGeneration),
		ExternalChannelID:      strings.TrimSpace(input.ExternalChannelID), ExternalThreadID: strings.TrimSpace(input.ExternalThreadID),
		ActorID: input.UserID,
	})
	if err != nil {
		return false, fmt.Errorf("check authorized integration request provider thread: %w", err)
	}
	return exists, nil
}

func (r *Repo) HasCurrentProviderThread(ctx context.Context, input integrationrequestdomain.ProviderThreadLookupInput) (bool, error) {
	if err := r.configured(); err != nil {
		return false, err
	}
	exists, err := r.queries.HasCurrentIntegrationRequestProviderThread(ctx, integrationrequestssql.HasCurrentIntegrationRequestProviderThreadParams{
		WorkspaceID: input.WorkspaceID, Provider: strings.TrimSpace(input.Provider),
		ExternalWorkspaceID:    strings.TrimSpace(input.ExternalWorkspaceID),
		InstallationGeneration: uuidPointer(input.InstallationGeneration),
		ExternalChannelID:      strings.TrimSpace(input.ExternalChannelID), ExternalThreadID: strings.TrimSpace(input.ExternalThreadID),
	})
	if err != nil {
		return false, fmt.Errorf("check current integration request provider thread: %w", err)
	}
	return exists, nil
}

func (r *Repo) FindProviderThread(ctx context.Context, workspaceID, requestID uuid.UUID, provider string) (integrationrequestdomain.ProviderThread, error) {
	if err := r.configured(); err != nil {
		return integrationrequestdomain.ProviderThread{}, err
	}
	row, err := r.queries.FindIntegrationRequestProviderThread(ctx, integrationrequestssql.FindIntegrationRequestProviderThreadParams{
		WorkspaceID: workspaceID, RequestID: requestID, Provider: strings.TrimSpace(provider),
	})
	if err != nil {
		return integrationrequestdomain.ProviderThread{}, mapProviderThreadNotFound("find integration request provider thread", err)
	}
	return providerThreadFromRecord(providerThreadRecord(row)), nil
}

func (r *Repo) GetThreadActivityForRequest(ctx context.Context, workspaceID, requestID, userID uuid.UUID) (integrationrequestdomain.ThreadActivity, error) {
	if err := r.configured(); err != nil {
		return integrationrequestdomain.ThreadActivity{}, err
	}
	row, err := r.queries.GetAuthorizedIntegrationRequestProviderThread(ctx, integrationrequestssql.GetAuthorizedIntegrationRequestProviderThreadParams{
		WorkspaceID: workspaceID, RequestID: requestID, ActorID: userID,
	})
	if err != nil {
		return integrationrequestdomain.ThreadActivity{}, mapProviderThreadNotFound("get authorized integration request provider thread", err)
	}
	thread := providerThreadFromRecord(providerThreadRecord(row))
	comments, err := r.listThreadComments(ctx, thread.ID)
	if err != nil {
		return integrationrequestdomain.ThreadActivity{}, err
	}
	return integrationrequestdomain.ThreadActivity{Thread: thread, Comments: comments}, nil
}

func (r *Repo) ListProviderThreadsForStory(ctx context.Context, workspaceID, storyID, userID uuid.UUID) ([]integrationrequestdomain.ProviderThread, error) {
	if err := r.configured(); err != nil {
		return nil, err
	}
	rows, err := r.queries.ListAuthorizedIntegrationRequestProviderThreadsForStory(ctx, integrationrequestssql.ListAuthorizedIntegrationRequestProviderThreadsForStoryParams{
		WorkspaceID: workspaceID, StoryID: uuidPointer(storyID), ActorID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("list authorized integration request provider threads for story: %w", err)
	}
	result := make([]integrationrequestdomain.ProviderThread, 0, len(rows))
	for _, row := range rows {
		result = append(result, providerThreadFromRecord(providerThreadRecord(row)))
	}
	return result, nil
}
