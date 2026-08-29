// Package searchdomain defines the storage- and transport-independent search
// contract shared by the search use case and its persistence adapter.
package searchdomain

import (
	"time"

	"github.com/google/uuid"
)

type CoreSearchStory struct {
	ID                       uuid.UUID
	SequenceID               int
	Title                    string
	Parent                   *uuid.UUID
	Objective                *uuid.UUID
	Status                   *uuid.UUID
	StatusName               *string
	StatusColor              *string
	StatusCategory           *string
	Assignee                 *uuid.UUID
	AssigneeFullName         *string
	AssigneeUsername         *string
	Reporter                 *uuid.UUID
	Priority                 string
	EstimateLabel            *string
	EstimateValue            *int16
	EstimateScheme           string
	Sprint                   *uuid.UUID
	KeyResult                *uuid.UUID
	Team                     uuid.UUID
	TeamName                 string
	TeamCode                 string
	Workspace                uuid.UUID
	StartDate                *time.Time
	EndDate                  *time.Time
	EstimatedDurationMinutes *int
	MinimumFocusBlockMinutes *int
	AutoSchedulingEnabled    bool
	AutoSchedulingLocked     bool
	AutoSchedulingStatus     string
	AutoSchedulingReason     *string
	AutoSchedulingUpdatedAt  *time.Time
	CreatedAt                time.Time
	UpdatedAt                time.Time
	Labels                   []uuid.UUID
}

type CoreSearchObjective struct {
	ID           uuid.UUID
	Name         string
	Description  *string
	ShortSummary *string
	LeadUser     *uuid.UUID
	LeadFullName *string
	LeadUsername *string
	Team         uuid.UUID
	TeamName     string
	TeamCode     string
	Workspace    uuid.UUID
	StartDate    *time.Time
	EndDate      *time.Time
	Status       uuid.UUID
	Priority     *string
	Health       *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CoreSearchResult struct {
	Stories         []CoreSearchStory
	Objectives      []CoreSearchObjective
	TotalStories    int
	TotalObjectives int
}

type CoreSimilarStory struct {
	ID         uuid.UUID
	SequenceID int
	Title      string
	Team       uuid.UUID
	Status     *uuid.UUID
	Assignee   *uuid.UUID
	Priority   string
	Confidence float64
}

type SearchType string

const (
	SearchTypeAll        SearchType = "all"
	SearchTypeStories    SearchType = "stories"
	SearchTypeObjectives SearchType = "objectives"
)

type SortOption string

const (
	SortByRelevance SortOption = "relevance"
	SortByUpdated   SortOption = "updated"
	SortByCreated   SortOption = "created"
)

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
