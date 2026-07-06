package labelshttp

import (
	"time"

	labels "github.com/complexus-tech/projects-api/internal/modules/labels/service"
	"github.com/google/uuid"
)

type AppLabel struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	TeamID      *uuid.UUID `json:"teamId"`
	WorkspaceID uuid.UUID  `json:"workspaceId"`
	Color       string     `json:"color"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type AppNewLabel struct {
	Name   string     `json:"name"`
	TeamID *uuid.UUID `json:"teamId"`
	Color  string     `json:"color"`
}

type AppUpdateLabel struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type AppFilters struct {
	Team     *uuid.UUID `json:"teamId" db:"team_id"`
	Search   string     `json:"search" db:"search"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
	Limit    int        `json:"limit"`
	Offset   int        `json:"offset"`
}

type AppPagination struct {
	Page     int  `json:"page"`
	PageSize int  `json:"pageSize"`
	HasMore  bool `json:"hasMore"`
	NextPage int  `json:"nextPage"`
}

type AppLabelsResponse struct {
	Labels     []AppLabel    `json:"labels"`
	Pagination AppPagination `json:"pagination"`
}

func toAppLabel(label labels.CoreLabel) AppLabel {
	return AppLabel{
		ID:          label.LabelID,
		Name:        label.Name,
		TeamID:      label.TeamID,
		WorkspaceID: label.WorkspaceID,
		Color:       label.Color,
		CreatedAt:   label.CreatedAt,
		UpdatedAt:   label.UpdatedAt,
	}
}

func toAppLabels(labels []labels.CoreLabel) []AppLabel {
	appLabels := make([]AppLabel, len(labels))
	for i, label := range labels {
		appLabels[i] = toAppLabel(label)
	}
	return appLabels
}

func toAppLabelsResponse(labels []labels.CoreLabel, page, pageSize int, hasMore bool) AppLabelsResponse {
	nextPage := 0
	if hasMore {
		nextPage = page + 1
	}

	return AppLabelsResponse{
		Labels: toAppLabels(labels),
		Pagination: AppPagination{
			Page:     page,
			PageSize: pageSize,
			HasMore:  hasMore,
			NextPage: nextPage,
		},
	}
}
