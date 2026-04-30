package integrationrequestshttp

import (
	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	"github.com/google/uuid"
)

type AppIntegrationRequest struct {
	ID               uuid.UUID      `json:"id"`
	WorkspaceID      uuid.UUID      `json:"workspaceId"`
	TeamID           uuid.UUID      `json:"teamId"`
	Provider         string         `json:"provider"`
	SourceType       string         `json:"sourceType"`
	SourceExternalID string         `json:"sourceExternalId"`
	SourceNumber     *int           `json:"sourceNumber,omitempty"`
	SourceURL        *string        `json:"sourceUrl,omitempty"`
	Title            string         `json:"title"`
	Description      *string        `json:"description,omitempty"`
	Status           string         `json:"status"`
	Metadata         map[string]any `json:"metadata"`
	AcceptedStoryID  *uuid.UUID     `json:"acceptedStoryId,omitempty"`
	CreatedAt        string         `json:"createdAt"`
	UpdatedAt        string         `json:"updatedAt"`
}

func toAppRequest(core integrationrequests.CoreIntegrationRequest) AppIntegrationRequest {
	return AppIntegrationRequest{
		ID:               core.ID,
		WorkspaceID:      core.WorkspaceID,
		TeamID:           core.TeamID,
		Provider:         core.Provider,
		SourceType:       core.SourceType,
		SourceExternalID: core.SourceExternalID,
		SourceNumber:     core.SourceNumber,
		SourceURL:        core.SourceURL,
		Title:            core.Title,
		Description:      core.Description,
		Status:           core.Status,
		Metadata:         core.Metadata,
		AcceptedStoryID:  core.AcceptedStoryID,
		CreatedAt:        core.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        core.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toAppRequests(core []integrationrequests.CoreIntegrationRequest) []AppIntegrationRequest {
	result := make([]AppIntegrationRequest, 0, len(core))
	for _, request := range core {
		result = append(result, toAppRequest(request))
	}
	return result
}
