package reportshttp

import (
	"net/url"
	"strings"
	"time"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const (
	defaultReportWindowDays = 60
	maximumReportFilterIDs  = 100
	maximumReportDateBytes  = 64
)

func parseReportFilters(query url.Values, now time.Time) (reports.ReportFilters, error) {
	filters, err := parseReportFilterIDs(query)
	if err != nil {
		return reports.ReportFilters{}, err
	}

	now = now.UTC()
	defaultStartDate := now.AddDate(0, 0, -defaultReportWindowDays)
	filters.StartDate = &defaultStartDate
	filters.EndDate = &now

	startDate, present, err := optionalReportDate(query, "startDate", false)
	if err != nil {
		return reports.ReportFilters{}, err
	}
	if present {
		filters.StartDate = startDate
	}
	endDate, present, err := optionalReportDate(query, "endDate", false)
	if err != nil {
		return reports.ReportFilters{}, err
	}
	if present {
		filters.EndDate = endDate
	}

	return filters, nil
}

func parseWorkloadAnalysisFilters(query url.Values) (reports.ReportFilters, error) {
	filters, err := parseReportFilterIDs(query)
	if err != nil {
		return reports.ReportFilters{}, err
	}
	startDate, present, err := optionalReportDate(query, "startDate", true)
	if err != nil {
		return reports.ReportFilters{}, err
	}
	if present {
		filters.StartDate = startDate
	}
	endDate, present, err := optionalReportDate(query, "endDate", true)
	if err != nil {
		return reports.ReportFilters{}, err
	}
	if present {
		filters.EndDate = endDate
	}

	return filters, nil
}

func parseReportFilterIDs(query url.Values) (reports.ReportFilters, error) {
	filters := reports.ReportFilters{}
	destinations := []struct {
		key         string
		destination *[]uuid.UUID
	}{
		{key: "teamIds", destination: &filters.TeamIDs},
		{key: "assigneeIds", destination: &filters.AssigneeIDs},
		{key: "sprintIds", destination: &filters.SprintIDs},
		{key: "objectiveIds", destination: &filters.ObjectiveIDs},
	}

	for _, field := range destinations {
		ids, err := web.OptionalCommaSeparatedUUIDQueryParameter(query, field.key, maximumReportFilterIDs)
		if err != nil {
			return reports.ReportFilters{}, err
		}
		*field.destination = ids
	}

	return filters, nil
}

func parseStatsFilters(query url.Values) (reports.StatsFilters, error) {
	filters := reports.StatsFilters{}
	destinations := []struct {
		key         string
		destination **uuid.UUID
	}{
		{key: "teamId", destination: &filters.TeamID},
		{key: "sprintId", destination: &filters.SprintID},
		{key: "objectiveId", destination: &filters.ObjectiveID},
	}

	for _, field := range destinations {
		value, present, err := web.OptionalQueryParameter(query, field.key, web.DefaultMaxQueryParameterBytes)
		if err != nil {
			return reports.StatsFilters{}, err
		}
		if !present {
			continue
		}
		if strings.TrimSpace(value) == "" {
			return reports.StatsFilters{}, ErrInvalidFilterID
		}
		id, err := uuid.Parse(strings.TrimSpace(value))
		if err != nil || id == uuid.Nil {
			return reports.StatsFilters{}, ErrInvalidFilterID
		}
		*field.destination = &id
	}

	return filters, nil
}

func optionalReportDate(query url.Values, name string, emptyIsAbsent bool) (*time.Time, bool, error) {
	value, present, err := web.OptionalQueryParameter(query, name, maximumReportDateBytes)
	if err != nil {
		return nil, false, err
	}
	if !present {
		return nil, false, nil
	}
	value = strings.TrimSpace(value)
	if value == "" && emptyIsAbsent {
		return nil, false, nil
	}
	parsed, err := parseReportDate(value)
	if err != nil {
		return nil, false, ErrInvalidDate
	}
	return &parsed, true, nil
}

func parseReportDate(value string) (time.Time, error) {
	if parsedDate, err := time.Parse(time.RFC3339, value); err == nil {
		return parsedDate, nil
	}

	return time.Parse("2006-01-02", value)
}
