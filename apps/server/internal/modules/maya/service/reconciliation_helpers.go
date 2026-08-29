package maya

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

func (s *Service) failScheduleRun(ctx context.Context, run CoreRun, runErr error) error {
	message := runErr.Error()
	_, completeErr := s.repo.CompleteRun(ctx, run.ID, RunStatusFailed, "Automatic schedule reconciliation failed.", &message)
	return errors.Join(runErr, completeErr)
}

func (s *Service) markScheduleActionsFailed(ctx context.Context, actions []CoreAction, actionErr error) error {
	var persistenceErr error
	for index := range actions {
		persistenceErr = errors.Join(
			persistenceErr,
			s.markActionFailed(ctx, &actions[index], actionErr.Error()),
		)
	}
	return persistenceErr
}

func appendUniqueUUID(values []uuid.UUID, value uuid.UUID) []uuid.UUID {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func valueOrZero(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func (s *Service) scheduleRepository() (ScheduleRepository, error) {
	repo, ok := s.repo.(ScheduleRepository)
	if !ok {
		return nil, ErrNotConfigured
	}
	return repo, nil
}

func (s *Service) scheduleCalendarService() (ScheduleCalendarService, error) {
	calendarService, ok := s.calendar.(ScheduleCalendarService)
	if !ok {
		return nil, ErrNotConfigured
	}
	return calendarService, nil
}
