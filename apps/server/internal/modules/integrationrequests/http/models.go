package integrationrequestshttp

import (
	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	"github.com/complexus-tech/projects-api/pkg/date"
	"github.com/google/uuid"
	"time"
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
	StatusID         *uuid.UUID     `json:"statusId,omitempty"`
	Priority         string         `json:"priority"`
	AssigneeID       *uuid.UUID     `json:"assigneeId,omitempty"`
	EstimateValue    *int16         `json:"estimateValue,omitempty"`
	ObjectiveID      *uuid.UUID     `json:"objectiveId,omitempty"`
	KeyResultID      *uuid.UUID     `json:"keyResultId,omitempty"`
	SprintID         *uuid.UUID     `json:"sprintId,omitempty"`
	StartDate        *time.Time     `json:"startDate,omitempty"`
	EndDate          *time.Time     `json:"endDate,omitempty"`
	Status           string         `json:"status"`
	Metadata         map[string]any `json:"metadata"`
	AcceptedStoryID  *uuid.UUID     `json:"acceptedStoryId,omitempty"`
	CreatedAt        string         `json:"createdAt"`
	UpdatedAt        string         `json:"updatedAt"`
}

type AppUpdateIntegrationRequest struct {
	Title         *string    `json:"title,omitempty"`
	Description   *string    `json:"description,omitempty"`
	StatusID      *uuid.UUID `json:"statusId,omitempty"`
	Priority      *string    `json:"priority,omitempty"`
	AssigneeID    *uuid.UUID `json:"assigneeId,omitempty"`
	EstimateValue *int16     `json:"estimateValue,omitempty"`
	ObjectiveID   *uuid.UUID `json:"objectiveId,omitempty"`
	KeyResultID   *uuid.UUID `json:"keyResultId,omitempty"`
	SprintID      *uuid.UUID `json:"sprintId,omitempty"`
	StartDate     *date.Date `json:"startDate,omitempty"`
	EndDate       *date.Date `json:"endDate,omitempty"`
}

type AppBulkRequestResult struct {
	Count      int         `json:"count"`
	RequestIDs []uuid.UUID `json:"requestIds"`
}

type AppPagination struct {
	Page       int  `json:"page"`
	PageSize   int  `json:"pageSize"`
	TotalCount int  `json:"totalCount"`
	HasMore    bool `json:"hasMore"`
	NextPage   int  `json:"nextPage"`
}

type AppIntegrationRequestsResponse struct {
	Requests   []AppIntegrationRequest `json:"requests"`
	Pagination AppPagination           `json:"pagination"`
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
		StatusID:         core.StatusID,
		Priority:         core.Priority,
		AssigneeID:       core.AssigneeID,
		EstimateValue:    core.EstimateValue,
		ObjectiveID:      core.ObjectiveID,
		KeyResultID:      core.KeyResultID,
		SprintID:         core.SprintID,
		StartDate:        core.StartDate,
		EndDate:          core.EndDate,
		Status:           core.Status,
		Metadata:         core.Metadata,
		AcceptedStoryID:  core.AcceptedStoryID,
		CreatedAt:        core.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        core.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toAppBulkRequestResult(core integrationrequests.CoreBulkRequestResult) AppBulkRequestResult {
	return AppBulkRequestResult{
		Count:      core.Count,
		RequestIDs: core.RequestIDs,
	}
}

func toAppRequests(core []integrationrequests.CoreIntegrationRequest) []AppIntegrationRequest {
	result := make([]AppIntegrationRequest, 0, len(core))
	for _, request := range core {
		result = append(result, toAppRequest(request))
	}
	return result
}

func toAppRequestsResponse(core []integrationrequests.CoreIntegrationRequest, page, pageSize, totalCount int, hasMore bool) AppIntegrationRequestsResponse {
	return AppIntegrationRequestsResponse{
		Requests: toAppRequests(core),
		Pagination: AppPagination{
			Page:       page,
			PageSize:   pageSize,
			TotalCount: totalCount,
			HasMore:    hasMore,
			NextPage:   page + 1,
		},
	}
}
