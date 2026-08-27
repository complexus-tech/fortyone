package searchrepository

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	"github.com/google/uuid"
)

// dbStory represents the database model for a story in search results.
type dbStory struct {
	ID                       uuid.UUID        `db:"id"`
	SequenceID               int              `db:"sequence_id"`
	Title                    string           `db:"title"`
	Parent                   *uuid.UUID       `db:"parent_id"`
	Objective                *uuid.UUID       `db:"objective_id"`
	Status                   *uuid.UUID       `db:"status_id"`
	StatusName               *string          `db:"status_name"`
	StatusColor              *string          `db:"status_color"`
	StatusCategory           *string          `db:"status_category"`
	Assignee                 *uuid.UUID       `db:"assignee_id"`
	AssigneeFullName         *string          `db:"assignee_full_name"`
	AssigneeUsername         *string          `db:"assignee_username"`
	Reporter                 *uuid.UUID       `db:"reporter_id"`
	Priority                 string           `db:"priority"`
	EstimateValue            *int16           `db:"estimate_unit"`
	EstimateScheme           string           `db:"estimate_scheme"`
	Sprint                   *uuid.UUID       `db:"sprint_id"`
	KeyResult                *uuid.UUID       `db:"key_result_id"`
	Team                     uuid.UUID        `db:"team_id"`
	TeamName                 string           `db:"team_name"`
	TeamCode                 string           `db:"team_code"`
	Workspace                uuid.UUID        `db:"workspace_id"`
	StartDate                *time.Time       `db:"start_date"`
	EndDate                  *time.Time       `db:"end_date"`
	EstimatedDurationMinutes *int             `db:"estimated_duration_minutes"`
	MinimumFocusBlockMinutes *int             `db:"minimum_focus_block_minutes"`
	AutoSchedulingEnabled    bool             `db:"auto_scheduling_enabled"`
	AutoSchedulingLocked     bool             `db:"auto_scheduling_locked"`
	AutoSchedulingStatus     string           `db:"auto_scheduling_status"`
	AutoSchedulingReason     *string          `db:"auto_scheduling_reason"`
	AutoSchedulingUpdatedAt  *time.Time       `db:"auto_scheduling_updated_at"`
	CreatedAt                time.Time        `db:"created_at"`
	UpdatedAt                time.Time        `db:"updated_at"`
	Labels                   *json.RawMessage `db:"labels"`
}

// dbObjective represents the database model for an objective in search results.
type dbObjective struct {
	ID           uuid.UUID  `db:"objective_id"`
	Name         string     `db:"name"`
	Description  *string    `db:"description"`
	ShortSummary *string    `db:"short_summary"`
	LeadUser     *uuid.UUID `db:"lead_user_id"`
	LeadFullName *string    `db:"lead_full_name"`
	LeadUsername *string    `db:"lead_username"`
	Team         uuid.UUID  `db:"team_id"`
	TeamName     string     `db:"team_name"`
	TeamCode     string     `db:"team_code"`
	Workspace    uuid.UUID  `db:"workspace_id"`
	StartDate    *time.Time `db:"start_date"`
	EndDate      *time.Time `db:"end_date"`
	Status       uuid.UUID  `db:"status_id"`
	Priority     *string    `db:"priority"`
	Health       *string    `db:"health"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}

type dbSimilarStory struct {
	ID         uuid.UUID  `db:"id"`
	SequenceID int        `db:"sequence_id"`
	Title      string     `db:"title"`
	Team       uuid.UUID  `db:"team_id"`
	Status     *uuid.UUID `db:"status_id"`
	Assignee   *uuid.UUID `db:"assignee_id"`
	Priority   string     `db:"priority"`
	Confidence float64    `db:"confidence"`
}

// toCoreSearchStory converts a dbStory to a CoreSearchStory.
func toCoreSearchStory(story dbStory) search.CoreSearchStory {
	var labels []uuid.UUID

	if story.Labels != nil {
		_ = json.Unmarshal(*story.Labels, &labels)
	}

	return search.CoreSearchStory{
		ID:                       story.ID,
		SequenceID:               story.SequenceID,
		Title:                    story.Title,
		Parent:                   story.Parent,
		Objective:                story.Objective,
		Status:                   story.Status,
		StatusName:               story.StatusName,
		StatusColor:              story.StatusColor,
		StatusCategory:           story.StatusCategory,
		Assignee:                 story.Assignee,
		AssigneeFullName:         story.AssigneeFullName,
		AssigneeUsername:         story.AssigneeUsername,
		Reporter:                 story.Reporter,
		Priority:                 story.Priority,
		EstimateLabel:            searchEstimateLabel(story.EstimateScheme, story.EstimateValue),
		EstimateValue:            story.EstimateValue,
		EstimateScheme:           story.EstimateScheme,
		Sprint:                   story.Sprint,
		KeyResult:                story.KeyResult,
		Team:                     story.Team,
		TeamName:                 story.TeamName,
		TeamCode:                 story.TeamCode,
		Workspace:                story.Workspace,
		StartDate:                story.StartDate,
		EndDate:                  story.EndDate,
		EstimatedDurationMinutes: story.EstimatedDurationMinutes,
		MinimumFocusBlockMinutes: story.MinimumFocusBlockMinutes,
		AutoSchedulingEnabled:    story.AutoSchedulingEnabled,
		AutoSchedulingLocked:     story.AutoSchedulingLocked,
		AutoSchedulingStatus:     story.AutoSchedulingStatus,
		AutoSchedulingReason:     story.AutoSchedulingReason,
		AutoSchedulingUpdatedAt:  story.AutoSchedulingUpdatedAt,
		CreatedAt:                story.CreatedAt,
		UpdatedAt:                story.UpdatedAt,
		Labels:                   labels,
	}
}

func searchEstimateLabel(scheme string, value *int16) *string {
	if value == nil {
		return nil
	}
	label := ""
	switch strings.TrimSpace(strings.ToLower(scheme)) {
	case "points":
		label = strconv.FormatInt(int64(*value), 10)
	default:
		label = map[int16]string{1: "XS", 2: "S", 3: "M", 5: "L", 8: "XL"}[*value]
	}
	if label == "" {
		return nil
	}
	return &label
}

// toCoreSearchStories converts multiple dbStories to CoreSearchStories.
func toCoreSearchStories(stories []dbStory) []search.CoreSearchStory {
	result := make([]search.CoreSearchStory, len(stories))
	for i, story := range stories {
		result[i] = toCoreSearchStory(story)
	}
	return result
}

// toCoreSearchObjective converts a dbObjective to a CoreSearchObjective.
func toCoreSearchObjective(objective dbObjective) search.CoreSearchObjective {
	return search.CoreSearchObjective{
		ID:           objective.ID,
		Name:         objective.Name,
		Description:  objective.Description,
		ShortSummary: objective.ShortSummary,
		LeadUser:     objective.LeadUser,
		LeadFullName: objective.LeadFullName,
		LeadUsername: objective.LeadUsername,
		Team:         objective.Team,
		TeamName:     objective.TeamName,
		TeamCode:     objective.TeamCode,
		Workspace:    objective.Workspace,
		StartDate:    objective.StartDate,
		EndDate:      objective.EndDate,
		Status:       objective.Status,
		Priority:     objective.Priority,
		Health:       objective.Health,
		CreatedAt:    objective.CreatedAt,
		UpdatedAt:    objective.UpdatedAt,
	}
}

// toCoreSearchObjectives converts multiple dbObjectives to CoreSearchObjectives.
func toCoreSearchObjectives(objectives []dbObjective) []search.CoreSearchObjective {
	result := make([]search.CoreSearchObjective, len(objectives))
	for i, objective := range objectives {
		result[i] = toCoreSearchObjective(objective)
	}
	return result
}

func toCoreSimilarStories(stories []dbSimilarStory) []search.CoreSimilarStory {
	result := make([]search.CoreSimilarStory, len(stories))
	for index, story := range stories {
		result[index] = search.CoreSimilarStory{
			ID:         story.ID,
			SequenceID: story.SequenceID,
			Title:      story.Title,
			Team:       story.Team,
			Status:     story.Status,
			Assignee:   story.Assignee,
			Priority:   story.Priority,
			Confidence: story.Confidence,
		}
	}
	return result
}
