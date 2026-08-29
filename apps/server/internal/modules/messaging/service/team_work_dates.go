package messaging

import (
	"fmt"
	"strings"
	"time"
)

type teamWorkDateRange struct {
	StartDate       string
	EndDate         string
	CompletedAfter  time.Time
	CompletedBefore time.Time
	DueAfter        time.Time
	DueBefore       time.Time
}

type teamWorkGroupMetadata struct {
	Total      int
	Loaded     int
	TotalExact bool
	Truncated  bool
}

func teamWorkDateRangeForMode(mode string, startDate, endDate *string, timezone string, now time.Time) (*teamWorkDateRange, error) {
	if mode != teamWorkModeCompleted && mode != teamWorkModeDue {
		return nil, nil
	}
	location := time.UTC
	if strings.TrimSpace(timezone) != "" {
		loaded, err := time.LoadLocation(strings.TrimSpace(timezone))
		if err != nil {
			return nil, fmt.Errorf("invalid team work timezone %q: %w", timezone, err)
		}
		location = loaded
	}

	today := now.In(location)
	start := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, location)
	end := start
	if startDate != nil {
		parsed, err := parseTeamWorkDate(*startDate, location)
		if err != nil {
			return nil, err
		}
		start = parsed
		end = parsed
	}
	if endDate != nil {
		parsed, err := parseTeamWorkDate(*endDate, location)
		if err != nil {
			return nil, err
		}
		end = parsed
		if startDate == nil {
			start = parsed
		}
	}
	if end.Before(start) {
		return nil, fmt.Errorf("%w: end_date must be on or after start_date", ErrInvalidToolArguments)
	}
	if end.AddDate(0, 0, -maxCompletedTaskDays).After(start) {
		return nil, fmt.Errorf("%w: team work date range cannot exceed %d days", ErrInvalidToolArguments, maxCompletedTaskDays)
	}

	endExclusive := end.AddDate(0, 0, 1)
	return &teamWorkDateRange{
		StartDate:       start.Format("2006-01-02"),
		EndDate:         end.Format("2006-01-02"),
		CompletedAfter:  start.UTC(),
		CompletedBefore: endExclusive.Add(-time.Nanosecond).UTC(),
		DueAfter:        time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC),
		DueBefore:       time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC),
	}, nil
}

func parseTeamWorkDate(value string, location *time.Location) (time.Time, error) {
	parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(value), location)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: team work dates must use YYYY-MM-DD: %v", ErrInvalidToolArguments, err)
	}
	return parsed, nil
}

func teamWorkCategories(mode string) []string {
	switch mode {
	case teamWorkModeInProgress:
		return []string{"started"}
	case teamWorkModeCompleted:
		return []string{"completed"}
	case teamWorkModeActive, teamWorkModeDue:
		return activeTeamWorkCategories
	default:
		return nil
	}
}
