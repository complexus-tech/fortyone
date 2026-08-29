package integrationrequestshttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	"github.com/complexus-tech/projects-api/pkg/date"
	"github.com/google/uuid"
)

type AppIntegrationRequest struct {
	ID                       uuid.UUID      `json:"id"`
	WorkspaceID              uuid.UUID      `json:"workspaceId"`
	TeamID                   uuid.UUID      `json:"teamId"`
	Provider                 string         `json:"provider"`
	SourceType               string         `json:"sourceType"`
	SourceExternalID         string         `json:"sourceExternalId"`
	SourceNumber             *int           `json:"sourceNumber,omitempty"`
	SourceURL                *string        `json:"sourceUrl,omitempty"`
	Title                    string         `json:"title"`
	Description              *string        `json:"description,omitempty"`
	StatusID                 *uuid.UUID     `json:"statusId,omitempty"`
	Priority                 string         `json:"priority"`
	AssigneeID               *uuid.UUID     `json:"assigneeId,omitempty"`
	EstimateValue            *int16         `json:"estimateValue,omitempty"`
	EstimatedDurationMinutes *int           `json:"estimatedDurationMinutes,omitempty"`
	MinimumFocusBlockMinutes *int           `json:"minimumFocusBlockMinutes,omitempty"`
	ObjectiveID              *uuid.UUID     `json:"objectiveId,omitempty"`
	KeyResultID              *uuid.UUID     `json:"keyResultId,omitempty"`
	SprintID                 *uuid.UUID     `json:"sprintId,omitempty"`
	StartDate                *time.Time     `json:"startDate,omitempty"`
	EndDate                  *time.Time     `json:"endDate,omitempty"`
	LabelIDs                 []uuid.UUID    `json:"labelIds"`
	Status                   string         `json:"status"`
	Metadata                 map[string]any `json:"metadata"`
	AcceptedStoryID          *uuid.UUID     `json:"acceptedStoryId,omitempty"`
	CreatedAt                string         `json:"createdAt"`
	UpdatedAt                string         `json:"updatedAt"`
}

type AppUpdateIntegrationRequest struct {
	Title                    *string                     `json:"title,omitempty"`
	Description              appOptionalValue[string]    `json:"description"`
	StatusID                 appOptionalValue[uuid.UUID] `json:"statusId"`
	Priority                 *string                     `json:"priority,omitempty"`
	AssigneeID               appOptionalValue[uuid.UUID] `json:"assigneeId"`
	EstimateValue            appOptionalValue[int16]     `json:"estimateValue"`
	EstimatedDurationMinutes appOptionalValue[int]       `json:"estimatedDurationMinutes"`
	MinimumFocusBlockMinutes appOptionalValue[int]       `json:"minimumFocusBlockMinutes"`
	ObjectiveID              appOptionalValue[uuid.UUID] `json:"objectiveId"`
	KeyResultID              appOptionalValue[uuid.UUID] `json:"keyResultId"`
	SprintID                 appOptionalValue[uuid.UUID] `json:"sprintId"`
	StartDate                appOptionalValue[date.Date] `json:"startDate"`
	EndDate                  appOptionalValue[date.Date] `json:"endDate"`
	LabelIDs                 *[]uuid.UUID                `json:"labelIds,omitempty"`
}

type appOptionalValue[T any] struct {
	Set   bool
	Value *T
}

func (value *appOptionalValue[T]) UnmarshalJSON(data []byte) error {
	value.Set = true
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("null")) {
		value.Value = nil
		return nil
	}
	var decoded T
	if err := json.Unmarshal(data, &decoded); err != nil {
		return fmt.Errorf("decode optional value: %w", err)
	}
	value.Value = &decoded
	return nil
}

func toCoreOptionalValue[T any](value appOptionalValue[T]) integrationrequests.OptionalValue[T] {
	return integrationrequests.OptionalValue[T]{Set: value.Set, Value: value.Value}
}

func toCoreOptionalDate(value appOptionalValue[date.Date]) integrationrequests.OptionalValue[time.Time] {
	result := integrationrequests.OptionalValue[time.Time]{Set: value.Set}
	if value.Value != nil {
		result.Value = value.Value.TimePtr()
	}
	return result
}

type AppProviderThread struct {
	ID                      uuid.UUID  `json:"id"`
	IntegrationRequestID    uuid.UUID  `json:"integrationRequestId"`
	TeamID                  uuid.UUID  `json:"teamId"`
	AcceptedStoryID         *uuid.UUID `json:"acceptedStoryId,omitempty"`
	Provider                string     `json:"provider"`
	ExternalChannelID       string     `json:"externalChannelId"`
	ExternalThreadID        string     `json:"externalThreadId"`
	ExternalSourceMessageID *string    `json:"externalSourceMessageId,omitempty"`
	SourceURL               *string    `json:"sourceUrl,omitempty"`
	RequestTitle            string     `json:"requestTitle"`
	CreatedAt               time.Time  `json:"createdAt"`
	UpdatedAt               time.Time  `json:"updatedAt"`
}

type AppIntegrationRequestComment struct {
	ID                uuid.UUID  `json:"id"`
	ThreadID          uuid.UUID  `json:"threadId"`
	Direction         string     `json:"direction"`
	AuthorUserID      *uuid.UUID `json:"authorUserId,omitempty"`
	AuthorName        string     `json:"authorName"`
	AuthorAvatar      *string    `json:"authorAvatar,omitempty"`
	ExternalAuthorID  *string    `json:"externalAuthorId,omitempty"`
	ExternalMessageID *string    `json:"externalMessageId,omitempty"`
	DeliveryStatus    *string    `json:"deliveryStatus,omitempty"`
	Body              string     `json:"body"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

type AppThreadActivity struct {
	Thread   AppProviderThread              `json:"thread"`
	Comments []AppIntegrationRequestComment `json:"comments"`
}

type AppCreateIntegrationRequestComment struct {
	Body           string    `json:"body"`
	IdempotencyKey uuid.UUID `json:"idempotencyKey"`
}

type AppBulkRequestResult struct {
	Count          int                        `json:"count"`
	RequestIDs     []uuid.UUID                `json:"requestIds"`
	TotalCount     int                        `json:"totalCount"`
	SucceededCount int                        `json:"succeededCount"`
	FailedCount    int                        `json:"failedCount"`
	Partial        bool                       `json:"partial"`
	Items          []AppBulkRequestItemResult `json:"items"`
}

type AppBulkRequestItemResult struct {
	RequestID       uuid.UUID  `json:"requestId"`
	Success         bool       `json:"success"`
	Status          string     `json:"status"`
	AcceptedStoryID *uuid.UUID `json:"acceptedStoryId,omitempty"`
	Error           string     `json:"error,omitempty"`
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
		ID:                       core.ID,
		WorkspaceID:              core.WorkspaceID,
		TeamID:                   core.TeamID,
		Provider:                 core.Provider,
		SourceType:               core.SourceType,
		SourceExternalID:         core.SourceExternalID,
		SourceNumber:             core.SourceNumber,
		SourceURL:                core.SourceURL,
		Title:                    core.Title,
		Description:              core.Description,
		StatusID:                 core.StatusID,
		Priority:                 core.Priority,
		AssigneeID:               core.AssigneeID,
		EstimateValue:            core.EstimateValue,
		EstimatedDurationMinutes: core.EstimatedDurationMinutes,
		MinimumFocusBlockMinutes: core.MinimumFocusBlockMinutes,
		ObjectiveID:              core.ObjectiveID,
		KeyResultID:              core.KeyResultID,
		SprintID:                 core.SprintID,
		StartDate:                core.StartDate,
		EndDate:                  core.EndDate,
		LabelIDs:                 core.LabelIDs,
		Status:                   core.Status,
		Metadata:                 core.Metadata,
		AcceptedStoryID:          core.AcceptedStoryID,
		CreatedAt:                core.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:                core.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toAppProviderThread(core integrationrequests.CoreProviderThread) AppProviderThread {
	return AppProviderThread{
		ID: core.ID, IntegrationRequestID: core.IntegrationRequestID, TeamID: core.TeamID,
		AcceptedStoryID: core.AcceptedStoryID, Provider: core.Provider,
		ExternalChannelID: core.ExternalChannelID, ExternalThreadID: core.ExternalThreadID,
		ExternalSourceMessageID: core.ExternalSourceMessageID, SourceURL: core.SourceURL,
		RequestTitle: core.RequestTitle, CreatedAt: core.CreatedAt, UpdatedAt: core.UpdatedAt,
	}
}

func toAppComment(core integrationrequests.CoreIntegrationRequestComment) AppIntegrationRequestComment {
	return AppIntegrationRequestComment{
		ID: core.ID, ThreadID: core.ThreadID, Direction: core.Direction,
		AuthorUserID: core.AuthorUserID, AuthorName: core.AuthorName, AuthorAvatar: core.AuthorAvatar,
		ExternalAuthorID: core.ExternalAuthorID, ExternalMessageID: core.ExternalMessageID,
		DeliveryStatus: core.DeliveryStatus, Body: core.Body, CreatedAt: core.CreatedAt, UpdatedAt: core.UpdatedAt,
	}
}

func toAppThreadActivity(core integrationrequests.CoreThreadActivity) AppThreadActivity {
	comments := make([]AppIntegrationRequestComment, 0, len(core.Comments))
	for _, comment := range core.Comments {
		comments = append(comments, toAppComment(comment))
	}
	return AppThreadActivity{Thread: toAppProviderThread(core.Thread), Comments: comments}
}

func toAppBulkRequestResult(core integrationrequests.CoreBulkRequestResult) AppBulkRequestResult {
	items := make([]AppBulkRequestItemResult, 0, len(core.Items))
	for _, item := range core.Items {
		items = append(items, AppBulkRequestItemResult{
			RequestID:       item.RequestID,
			Success:         item.Success,
			Status:          item.Status,
			AcceptedStoryID: item.AcceptedStoryID,
			Error:           item.Error,
		})
	}

	return AppBulkRequestResult{
		Count:          core.Count,
		RequestIDs:     core.RequestIDs,
		TotalCount:     core.TotalCount,
		SucceededCount: core.SucceededCount,
		FailedCount:    core.FailedCount,
		Partial:        core.Partial,
		Items:          items,
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
