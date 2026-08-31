package calendar_test

import (
	calendarrepository "github.com/complexus-tech/projects-api/internal/modules/calendar/repository"
	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
)

var _ calendar.ScheduleReconciliationRepository = (*calendarrepository.Repo)(nil)
