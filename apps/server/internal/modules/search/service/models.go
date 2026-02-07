package search

import (
	"time"

	"github.com/google/uuid"
)

// CoreSearchStory represents a story in search results
type CoreSearchStory struct {
	ID         uuid.UUID
	SequenceID int
	Title      string
	Parent     *uuid.UUID
	Objective  *uuid.UUID
	Status     *uuid.UUID
	Assignee   *uuid.UUID
	Reporter   *uuid.UUID
	Priority   string
	Sprint     *uuid.UUID
	KeyResult  *uuid.UUID
	Team       uuid.UUID
	Workspace  uuid.UUID
	StartDate  *time.Time
	EndDate    *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Labels     []uuid.UUID
}

// CoreSearchObjective represents an objective in search results
type CoreSearchObjective struct {
	ID          uuid.UUID
	Name        string
	Description *string
	LeadUser    *uuid.UUID
	Team        uuid.UUID
	Workspace   uuid.UUID
	StartDate   *time.Time
	EndDate     *time.Time
	Status      uuid.UUID
	Priority    *string
	Health      *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CoreSearchResult represents the combined search results
type CoreSearchResult struct {
	Stories         []CoreSearchStory
	Objectives      []CoreSearchObjective
	TotalStories    int
	TotalObjectives int
}

// SearchType defines the type of content to search for
type SearchType string

const (
	// SearchTypeAll searches all content types
	SearchTypeAll SearchType = "all"
	// SearchTypeStories searches only stories
	SearchTypeStories SearchType = "stories"
	// SearchTypeObjectives searches only objectives
	SearchTypeObjectives SearchType = "objectives"
)

// SortOption defines how search results should be sorted
type SortOption string

const (
	// SortByRelevance sorts by search relevance (default)
	SortByRelevance SortOption = "relevance"
	// SortByUpdated sorts by last updated time
	SortByUpdated SortOption = "updated"
	// SortByCreated sorts by creation time
	SortByCreated SortOption = "created"
)

// SearchParams represents the parameters for a search query
type SearchParams struct {
	Type       SearchType
	Query      string
	TeamID     *uuid.UUID
	AssigneeID *uuid.UUID
	LabelID    *uuid.UUID
	StatusID   *uuid.UUID
	Priority   *string
	SortBy     SortOption
	Page       int
	PageSize   int
}
