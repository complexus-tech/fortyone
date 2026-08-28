package activitieshttp

import (
	"errors"
	"net/url"
	"strings"
	"time"

	activities "github.com/complexus-tech/projects-api/internal/modules/activities/service"
	"github.com/complexus-tech/projects-api/pkg/date"
	"github.com/complexus-tech/projects-api/pkg/web"
)

const (
	defaultActivityLimit        = 10
	maximumActivityLimit        = 100
	maximumActivityLimitBytes   = 16
	maximumActivityDateBytes    = 64
	defaultActivityLookbackDays = 30
)

type activityListQuery struct {
	Limit   int
	Filters activities.ActivityFilters
}

func parseActivityListQuery(values url.Values, now time.Time) (activityListQuery, error) {
	limit := defaultActivityLimit
	parsedLimit, present, err := web.OptionalIntegerQueryParameter(
		values,
		"limit",
		maximumActivityLimitBytes,
		1,
		maximumActivityLimit,
	)
	if err != nil {
		if errors.Is(err, web.ErrInvalidQueryParameter) {
			return activityListQuery{}, ErrInvalidLimit
		}
		return activityListQuery{}, err
	}
	if present {
		limit = parsedLimit
	}

	startDate, endDate, err := parseActivityDateRange(values, now)
	if err != nil {
		return activityListQuery{}, err
	}
	return activityListQuery{
		Limit: limit,
		Filters: activities.ActivityFilters{
			StartDate: startDate,
			EndDate:   endDate,
		},
	}, nil
}

func parseActivityDateRange(values url.Values, now time.Time) (time.Time, time.Time, error) {
	today := now.UTC()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	startDate := today.AddDate(0, 0, -defaultActivityLookbackDays)
	endDate := date.EndOfDay(today)

	parsedStartDate, present, err := parseOptionalActivityDate(values, "startDate")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if present {
		startDate = parsedStartDate
	}

	parsedEndDate, present, err := parseOptionalActivityDate(values, "endDate")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if present {
		endDate = date.EndOfDay(parsedEndDate)
	}
	if endDate.Before(startDate) {
		return time.Time{}, time.Time{}, ErrInvalidDate
	}
	return startDate, endDate, nil
}

func parseOptionalActivityDate(values url.Values, name string) (time.Time, bool, error) {
	raw, present, err := web.OptionalQueryParameter(values, name, maximumActivityDateBytes)
	if err != nil {
		return time.Time{}, false, err
	}
	raw = strings.TrimSpace(raw)
	if !present || raw == "" {
		return time.Time{}, false, nil
	}
	parsed, err := date.ParseDateOnly(raw)
	if err != nil {
		return time.Time{}, false, ErrInvalidDate
	}
	return parsed, true, nil
}
