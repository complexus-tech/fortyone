package storieshttp

import (
	"fmt"
	"net/http"
	"strings"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func parseListStoryFilters(r *http.Request) (stories.CoreStoryFilters, error) {
	var filters stories.CoreStoryFilters
	var err error
	if filters.Parent, err = parseStrictOptionalUUID(r, "parentId"); err != nil {
		return stories.CoreStoryFilters{}, err
	}
	if filters.Objective, err = parseStrictOptionalUUID(r, "objectiveId"); err != nil {
		return stories.CoreStoryFilters{}, err
	}
	if filters.KeyResult, err = parseStrictOptionalUUID(r, "keyResultId"); err != nil {
		return stories.CoreStoryFilters{}, err
	}
	for _, filter := range []struct {
		key    string
		values *[]uuid.UUID
	}{
		{key: "statusId", values: &filters.StatusIDs},
		{key: "assigneeId", values: &filters.AssigneeIDs},
		{key: "sprintId", values: &filters.SprintIDs},
		{key: "teamId", values: &filters.TeamIDs},
		{key: "reporterId", values: &filters.ReporterIDs},
	} {
		value, parseErr := parseStrictOptionalUUID(r, filter.key)
		if parseErr != nil {
			return stories.CoreStoryFilters{}, parseErr
		}
		if value != nil {
			*filter.values = []uuid.UUID{*value}
		}
	}
	if filters.Epic, err = parseStrictOptionalUUID(r, "epicId"); err != nil {
		return stories.CoreStoryFilters{}, err
	}
	if filters.Epic != nil {
		return stories.CoreStoryFilters{}, fmt.Errorf("epicId is not supported by the story schema")
	}
	value, present, err := web.OptionalQueryParameter(r.URL.Query(), "priority", 32)
	if err != nil {
		return stories.CoreStoryFilters{}, err
	}
	value = strings.TrimSpace(value)
	if present && value == "" {
		return stories.CoreStoryFilters{}, fmt.Errorf("priority must not be blank")
	}
	if present {
		if !isValidPriority(value) {
			return stories.CoreStoryFilters{}, fmt.Errorf("invalid priority value")
		}
		filters.Priorities = []string{value}
	}
	filters.ShowSubStories, err = parseStrictOptionalBool(r, "showSubStories")
	if err != nil {
		return stories.CoreStoryFilters{}, err
	}
	return filters, nil
}

func isValidPriority(value string) bool {
	switch value {
	case "Urgent", "High", "Medium", "Low", "No Priority":
		return true
	default:
		return false
	}
}
