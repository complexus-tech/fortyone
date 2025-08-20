package searchgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/search"
	"github.com/google/uuid"
)

// AppSearchStory represents a story in search results for the API.
type AppSearchStory struct {
	ID         uuid.UUID   `json:"id"`
	SequenceID int         `json:"sequenceId"`
	Title      string      `json:"title"`
	Parent     *uuid.UUID  `json:"parentId"`
	Objective  *uuid.UUID  `json:"objectiveId"`
	Status     *uuid.UUID  `json:"statusId"`
	Assignee   *uuid.UUID  `json:"assigneeId"`
	Reporter   *uuid.UUID  `json:"reporterId"`
	Priority   string      `json:"priority"`
	Sprint     *uuid.UUID  `json:"sprintId"`
	Team       uuid.UUID   `json:"teamId"`
	Workspace  uuid.UUID   `json:"workspaceId"`
	StartDate  *time.Time  `json:"startDate"`
	EndDate    *time.Time  `json:"endDate"`
	CreatedAt  time.Time   `json:"createdAt"`
	UpdatedAt  time.Time   `json:"updatedAt"`
	Labels     []uuid.UUID `json:"labels"`
	SubStories []uuid.UUID `json:"subStories"`
}

// AppSearchObjective represents an objective in search results for the API.
type AppSearchObjective struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	LeadUser    *uuid.UUID `json:"leadUser"`
	Team        uuid.UUID  `json:"teamId"`
	Workspace   uuid.UUID  `json:"workspaceId"`
	StartDate   *time.Time `json:"startDate"`
	EndDate     *time.Time `json:"endDate"`
	Status      uuid.UUID  `json:"statusId"`
	Priority    *string    `json:"priority"`
	Health      *string    `json:"health"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// AppSearchResponse represents the API response for a search request.
type AppSearchResponse struct {
	Stories         []AppSearchStory     `json:"stories"`
	Objectives      []AppSearchObjective `json:"objectives"`
	TotalStories    int                  `json:"totalStories"`
	TotalObjectives int                  `json:"totalObjectives"`
	Page            int                  `json:"page"`
	PageSize        int                  `json:"pageSize"`
	TotalPages      int                  `json:"totalPages"`
}

// AppSearchParams represents the query parameters for a search request.
type AppSearchParams struct {
	Type       string     `query:"type"`
	Query      string     `query:"query"`
	TeamID     *uuid.UUID `query:"teamId"`
	AssigneeID *uuid.UUID `query:"assigneeId"`
	LabelID    *uuid.UUID `query:"labelId"`
	StatusID   *uuid.UUID `query:"statusId"`
	Priority   *string    `query:"priority"`
	SortBy     string     `query:"sortBy"`
	Page       int        `query:"page"`
	PageSize   int        `query:"pageSize"`
}

// toAppSearchStories converts core stories to app stories.
func toAppSearchStories(stories []search.CoreSearchStory) []AppSearchStory {
	result := make([]AppSearchStory, len(stories))
	for i, story := range stories {
		result[i] = toAppSearchStory(story)
	}
	return result
}

// toAppSearchStory converts a core story to an app story.
func toAppSearchStory(story search.CoreSearchStory) AppSearchStory {
	var subStories []uuid.UUID = []uuid.UUID{}
	return AppSearchStory{
		ID:         story.ID,
		SequenceID: story.SequenceID,
		Title:      story.Title,
		Parent:     story.Parent,
		Objective:  story.Objective,
		Status:     story.Status,
		Assignee:   story.Assignee,
		Reporter:   story.Reporter,
		Priority:   story.Priority,
		Sprint:     story.Sprint,
		Team:       story.Team,
		Workspace:  story.Workspace,
		StartDate:  story.StartDate,
		EndDate:    story.EndDate,
		CreatedAt:  story.CreatedAt,
		UpdatedAt:  story.UpdatedAt,
		Labels:     story.Labels,
		SubStories: subStories,
	}
}

// toAppSearchObjectives converts core objectives to app objectives.
func toAppSearchObjectives(objectives []search.CoreSearchObjective) []AppSearchObjective {
	result := make([]AppSearchObjective, len(objectives))
	for i, objective := range objectives {
		result[i] = toAppSearchObjective(objective)
	}
	return result
}

// toAppSearchObjective converts a core objective to an app objective.
func toAppSearchObjective(objective search.CoreSearchObjective) AppSearchObjective {
	return AppSearchObjective{
		ID:          objective.ID,
		Name:        objective.Name,
		Description: objective.Description,
		LeadUser:    objective.LeadUser,
		Team:        objective.Team,
		Workspace:   objective.Workspace,
		StartDate:   objective.StartDate,
		EndDate:     objective.EndDate,
		Status:      objective.Status,
		Priority:    objective.Priority,
		Health:      objective.Health,
		CreatedAt:   objective.CreatedAt,
		UpdatedAt:   objective.UpdatedAt,
	}
}

// toAppSearchResponse converts a core search result to an app search response.
func toAppSearchResponse(result search.CoreSearchResult, page, pageSize int) AppSearchResponse {
	// Calculate total pages
	totalItems := result.TotalStories + result.TotalObjectives
	totalPages := totalItems / pageSize
	if totalItems%pageSize > 0 {
		totalPages++
	}

	return AppSearchResponse{
		Stories:         toAppSearchStories(result.Stories),
		Objectives:      toAppSearchObjectives(result.Objectives),
		TotalStories:    result.TotalStories,
		TotalObjectives: result.TotalObjectives,
		Page:            page,
		PageSize:        pageSize,
		TotalPages:      totalPages,
	}
}
