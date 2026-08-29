package reports

import (
	"context"
	"fmt"
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

const (
	maxReportFilterIDs = 100
	maxReportDateRange = 366 * 24 * time.Hour
)

func normalizeReportFilters(ctx context.Context, workspaceID uuid.UUID, filters ReportFilters, requireDates bool) (ReportFilters, error) {
	actorID, err := reportActorID(ctx, workspaceID)
	if err != nil {
		return ReportFilters{}, err
	}
	filters.ActorID = actorID

	filters.TeamIDs, err = normalizeFilterIDs(filters.TeamIDs)
	if err != nil {
		return ReportFilters{}, err
	}
	filters.AssigneeIDs, err = normalizeFilterIDs(filters.AssigneeIDs)
	if err != nil {
		return ReportFilters{}, err
	}
	filters.SprintIDs, err = normalizeFilterIDs(filters.SprintIDs)
	if err != nil {
		return ReportFilters{}, err
	}
	filters.ObjectiveIDs, err = normalizeFilterIDs(filters.ObjectiveIDs)
	if err != nil {
		return ReportFilters{}, err
	}

	if requireDates && (filters.StartDate == nil || filters.EndDate == nil) {
		return ReportFilters{}, ErrInvalidReportFilters
	}
	if err := validateDateRange(filters.StartDate, filters.EndDate); err != nil {
		return ReportFilters{}, err
	}

	return filters, nil
}

func normalizeStoryStatsFilters(ctx context.Context, workspaceID uuid.UUID, filters StoryStatsFilters) (StoryStatsFilters, error) {
	actorID, err := reportActorID(ctx, workspaceID)
	if err != nil {
		return StoryStatsFilters{}, err
	}
	if err := validateRequiredDateRange(filters.StartDate, filters.EndDate); err != nil {
		return StoryStatsFilters{}, err
	}
	filters.ActorID = actorID

	return filters, nil
}

func normalizeStatsFilters(ctx context.Context, workspaceID uuid.UUID, filters StatsFilters) (StatsFilters, error) {
	actorID, err := reportActorID(ctx, workspaceID)
	if err != nil {
		return StatsFilters{}, err
	}
	if err := validateRequiredDateRange(filters.StartDate, filters.EndDate); err != nil {
		return StatsFilters{}, err
	}
	filters.ActorID = actorID

	return filters, nil
}

func authorizeUserReport(ctx context.Context, actorID uuid.UUID, workspaceID uuid.UUID, startDate *time.Time, endDate *time.Time) error {
	contextActorID, err := reportActorID(ctx, workspaceID)
	if err != nil {
		return err
	}
	if actorID == uuid.Nil || actorID != contextActorID {
		return ErrReportsAccessDenied
	}
	return validateDateRange(startDate, endDate)
}

func reportActorID(ctx context.Context, workspaceID uuid.UUID) (uuid.UUID, error) {
	if workspaceID == uuid.Nil {
		return uuid.Nil, ErrReportsAccessDenied
	}
	actorID, err := platformauth.GetUserID(ctx)
	if err != nil || actorID == uuid.Nil {
		return uuid.Nil, ErrReportsAccessDenied
	}

	return actorID, nil
}

func normalizeFilterIDs(values []uuid.UUID) ([]uuid.UUID, error) {
	if len(values) > maxReportFilterIDs {
		return nil, fmt.Errorf("too many report filter identifiers: %w", ErrInvalidReportFilters)
	}
	seen := make(map[uuid.UUID]struct{}, len(values))
	result := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if value == uuid.Nil {
			return nil, ErrInvalidReportFilters
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}

	return result, nil
}

func validateRequiredDateRange(startDate time.Time, endDate time.Time) error {
	if startDate.IsZero() || endDate.IsZero() {
		return ErrInvalidReportFilters
	}
	return validateDateRange(&startDate, &endDate)
}

func validateDateRange(startDate *time.Time, endDate *time.Time) error {
	if startDate == nil && endDate == nil {
		return nil
	}
	if startDate == nil || endDate == nil || startDate.IsZero() || endDate.IsZero() {
		return ErrInvalidReportFilters
	}
	if endDate.Before(*startDate) || endDate.Sub(*startDate) > maxReportDateRange {
		return ErrInvalidReportFilters
	}

	return nil
}
