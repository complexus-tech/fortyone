package searchrepository

import (
	"errors"
	"math"
	"strconv"
	"strings"

	searchdomain "github.com/complexus-tech/projects-api/internal/modules/search/domain"
	searchsql "github.com/complexus-tech/projects-api/internal/modules/search/repository/sqlc"
	"github.com/google/uuid"
)

var errInvalidProjection = errors.New("search database projection is incomplete")

func toCoreSearchStory(row searchsql.SearchStoriesRow) (searchdomain.CoreSearchStory, error) {
	if row.SequenceID == nil || row.Priority == nil {
		return searchdomain.CoreSearchStory{}, errInvalidProjection
	}

	return searchdomain.CoreSearchStory{
		ID:                       row.ID,
		SequenceID:               int(*row.SequenceID),
		Title:                    row.Title,
		Parent:                   row.ParentID,
		Objective:                row.ObjectiveID,
		Status:                   row.StatusID,
		StatusName:               row.StatusName,
		StatusColor:              row.StatusColor,
		StatusCategory:           row.StatusCategory,
		Assignee:                 row.AssigneeID,
		AssigneeFullName:         row.AssigneeFullName,
		AssigneeUsername:         row.AssigneeUsername,
		Reporter:                 row.ReporterID,
		Priority:                 *row.Priority,
		EstimateLabel:            searchEstimateLabel(row.EstimateScheme, row.EstimateUnit),
		EstimateValue:            row.EstimateUnit,
		EstimateScheme:           row.EstimateScheme,
		Sprint:                   row.SprintID,
		KeyResult:                row.KeyResultID,
		Team:                     row.TeamID,
		TeamName:                 row.TeamName,
		TeamCode:                 row.TeamCode,
		Workspace:                row.WorkspaceID,
		StartDate:                row.StartDate,
		EndDate:                  row.EndDate,
		EstimatedDurationMinutes: int32PointerToInt(row.EstimatedDurationMinutes),
		MinimumFocusBlockMinutes: int32PointerToInt(row.MinimumFocusBlockMinutes),
		AutoSchedulingEnabled:    row.AutoSchedulingEnabled,
		AutoSchedulingLocked:     row.AutoSchedulingLocked,
		AutoSchedulingStatus:     row.AutoSchedulingStatus,
		AutoSchedulingReason:     row.AutoSchedulingReason,
		AutoSchedulingUpdatedAt:  row.AutoSchedulingUpdatedAt,
		CreatedAt:                row.CreatedAt,
		UpdatedAt:                row.UpdatedAt,
		Labels:                   append([]uuid.UUID(nil), row.LabelIds...),
	}, nil
}

func toCoreSearchObjective(row searchsql.SearchObjectivesRow) (searchdomain.CoreSearchObjective, error) {
	if row.TeamID == nil || row.WorkspaceID == nil || row.StatusID == nil {
		return searchdomain.CoreSearchObjective{}, errInvalidProjection
	}

	var health *string
	if row.Health != nil {
		value := string(*row.Health)
		health = &value
	}
	return searchdomain.CoreSearchObjective{
		ID:           row.ObjectiveID,
		Name:         row.Name,
		Description:  row.Description,
		ShortSummary: row.ShortSummary,
		LeadUser:     row.LeadUserID,
		LeadFullName: row.LeadFullName,
		LeadUsername: row.LeadUsername,
		Team:         *row.TeamID,
		TeamName:     row.TeamName,
		TeamCode:     row.TeamCode,
		Workspace:    *row.WorkspaceID,
		StartDate:    row.StartDate,
		EndDate:      row.EndDate,
		Status:       *row.StatusID,
		Priority:     row.Priority,
		Health:       health,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}, nil
}

func toCoreSimilarStory(row searchsql.FindSimilarStoriesRow) (searchdomain.CoreSimilarStory, error) {
	if row.SequenceID == nil || row.Priority == nil {
		return searchdomain.CoreSimilarStory{}, errInvalidProjection
	}
	return searchdomain.CoreSimilarStory{
		ID:         row.ID,
		SequenceID: int(*row.SequenceID),
		Title:      row.Title,
		Team:       row.TeamID,
		Status:     row.StatusID,
		Assignee:   row.AssigneeID,
		Priority:   *row.Priority,
		Confidence: row.Confidence,
	}, nil
}

func searchEstimateLabel(scheme string, value *int16) *string {
	if value == nil {
		return nil
	}
	var label string
	if strings.EqualFold(strings.TrimSpace(scheme), "points") {
		label = strconv.FormatInt(int64(*value), 10)
	} else {
		label = map[int16]string{1: "XS", 2: "S", 3: "M", 5: "L", 8: "XL"}[*value]
	}
	if label == "" {
		return nil
	}
	return &label
}

func int32PointerToInt(value *int32) *int {
	if value == nil {
		return nil
	}
	converted := int(*value)
	return &converted
}

func boundedCount(value int64) int {
	if value <= 0 {
		return 0
	}
	if value > int64(math.MaxInt) {
		return math.MaxInt
	}
	return int(value)
}
