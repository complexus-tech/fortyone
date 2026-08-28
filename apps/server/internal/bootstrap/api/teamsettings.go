package api

import (
	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/service"
	"github.com/complexus-tech/projects-api/pkg/tasks"
)

type teamSettingsAutomationScheduler struct {
	tasks *tasks.Service
}

var _ teamsettings.AutomationScheduler = teamSettingsAutomationScheduler{}

func newTeamSettingsAutomationScheduler(taskService *tasks.Service) teamsettings.AutomationScheduler {
	if taskService == nil {
		return nil
	}
	return teamSettingsAutomationScheduler{tasks: taskService}
}

func (s teamSettingsAutomationScheduler) ScheduleSprintCreation() error {
	_, err := s.tasks.EnqueueSprintAutoCreation()
	return err
}

func (s teamSettingsAutomationScheduler) ScheduleStoryAutoClose() error {
	_, err := s.tasks.EnqueueStoryAutoClose()
	return err
}

func (s teamSettingsAutomationScheduler) ScheduleStoryAutoArchive() error {
	_, err := s.tasks.EnqueueStoryAutoArchive()
	return err
}
