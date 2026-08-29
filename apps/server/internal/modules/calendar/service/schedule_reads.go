package calendar

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func (s *Service) ListSchedule(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) (CoreSchedule, error) {
	if s.repo == nil {
		return CoreSchedule{}, ErrCalendarNotConfigured
	}
	if err := validateScheduleRange(startAt, endAt); err != nil {
		return CoreSchedule{}, err
	}
	busyWindows, err := s.repo.ListBusyWindows(ctx, workspaceID, userID, startAt, endAt)
	if err != nil {
		return CoreSchedule{}, err
	}
	blocks, err := s.repo.ListScheduleBlocks(ctx, workspaceID, userID, startAt, endAt)
	if err != nil {
		return CoreSchedule{}, err
	}
	return CoreSchedule{
		StartAt:     startAt.UTC(),
		EndAt:       endAt.UTC(),
		Timezone:    s.scheduleTimezone(ctx, workspaceID, userID),
		BusyWindows: busyWindows,
		Blocks:      blocks,
	}, nil
}

// ListSchedulingAvailability is an internal planning view. It includes the
// user's blocks across workspaces but the repository redacts other-workspace
// story details so account-wide collision protection cannot leak content.
func (s *Service) ListSchedulingAvailability(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) (CoreSchedule, error) {
	if s.repo == nil {
		return CoreSchedule{}, ErrCalendarNotConfigured
	}
	if err := validateScheduleRange(startAt, endAt); err != nil {
		return CoreSchedule{}, err
	}
	busyWindows, err := s.repo.ListBusyWindows(ctx, workspaceID, userID, startAt, endAt)
	if err != nil {
		return CoreSchedule{}, err
	}
	scheduleRepo, err := s.scheduleReconciliationRepository()
	if err != nil {
		return CoreSchedule{}, err
	}
	blocks, err := scheduleRepo.ListSchedulingBlocksForUser(ctx, workspaceID, userID, startAt, endAt)
	if err != nil {
		return CoreSchedule{}, err
	}
	return CoreSchedule{
		StartAt: startAt.UTC(), EndAt: endAt.UTC(), Timezone: s.scheduleTimezone(ctx, workspaceID, userID),
		BusyWindows: busyWindows, Blocks: blocks,
	}, nil
}

func (s *Service) ListManualSchedulePreference(ctx context.Context, workspaceID, userID uuid.UUID) (CoreSchedulePreference, error) {
	feedbackRepo, ok := s.repo.(ScheduleFeedbackRepository)
	if !ok {
		return CoreSchedulePreference{}, nil
	}
	events, err := feedbackRepo.ListManualScheduleRescheduleEvents(ctx, workspaceID, userID, time.Now().UTC().Add(-90*24*time.Hour))
	if err != nil {
		return CoreSchedulePreference{}, err
	}
	if len(events) == 0 {
		return CoreSchedulePreference{}, nil
	}

	now := time.Now().UTC()
	var weightedStart float64
	var totalWeight float64
	for _, event := range events {
		location, locationErr := time.LoadLocation(fallbackTimezone(event.Timezone))
		if locationErr != nil {
			location = time.UTC
		}
		localStart := event.NextStartAt.In(location)
		minutes := localStart.Hour()*60 + localStart.Minute()
		ageDays := now.Sub(event.CreatedAt.UTC()).Hours() / 24
		if ageDays < 0 {
			ageDays = 0
		}
		weight := 1 / (1 + ageDays/30)
		weightedStart += float64(minutes) * weight
		totalWeight += weight
	}
	if totalWeight == 0 {
		return CoreSchedulePreference{}, nil
	}
	preferredStartMinute := int(weightedStart/totalWeight + 0.5)
	return CoreSchedulePreference{
		PreferredStartMinute: &preferredStartMinute,
		SampleCount:          len(events),
		Confidence:           minFloat(totalWeight/3, 1),
	}, nil
}

func minFloat(left, right float64) float64 {
	if left < right {
		return left
	}
	return right
}

func (s *Service) scheduleTimezone(ctx context.Context, workspaceID, userID uuid.UUID) string {
	connection, err := s.repo.GetActiveConnection(ctx, workspaceID, userID, ProviderGoogle)
	if err != nil {
		connection, err = s.repo.GetActiveConnection(ctx, workspaceID, userID, ProviderMicrosoft)
		if err != nil {
			return "UTC"
		}
	}
	return fallbackTimezone(connection.Timezone)
}

func (s *Service) ListCalendarView(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) (CoreCalendarView, error) {
	if s.repo == nil {
		return CoreCalendarView{}, ErrCalendarNotConfigured
	}
	if err := validateScheduleRange(startAt, endAt); err != nil {
		return CoreCalendarView{}, err
	}
	events, err := s.repo.ListCalendarEvents(ctx, workspaceID, userID, startAt, endAt)
	if err != nil {
		return CoreCalendarView{}, err
	}
	busyWindows, err := s.repo.ListBusyWindows(ctx, workspaceID, userID, startAt, endAt)
	if err != nil {
		return CoreCalendarView{}, err
	}
	scheduleRepo, err := s.scheduleReconciliationRepository()
	if err != nil {
		return CoreCalendarView{}, err
	}
	blocks, err := scheduleRepo.ListSchedulingBlocksForUser(ctx, workspaceID, userID, startAt, endAt)
	if err != nil {
		return CoreCalendarView{}, err
	}
	scheduleIssues := []CoreScheduleIssue{}
	if issueRepo, ok := s.repo.(ScheduleIssueRepository); ok {
		scheduleIssues, err = issueRepo.ListScheduleIssues(ctx, workspaceID, userID)
		if err != nil {
			return CoreCalendarView{}, err
		}
	}
	return CoreCalendarView{
		StartAt:        startAt.UTC(),
		EndAt:          endAt.UTC(),
		Events:         events,
		BusyWindows:    busyWindows,
		Blocks:         blocks,
		ScheduleIssues: scheduleIssues,
	}, nil
}
