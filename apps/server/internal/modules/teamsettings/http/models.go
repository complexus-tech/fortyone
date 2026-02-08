package teamsettingshttp

import (
	"time"

	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/service"
)

type AppTeamSprintSettings struct {
	AutoCreateSprints            bool      `json:"autoCreateSprints"`
	UpcomingSprintsCount         int       `json:"upcomingSprintsCount"`
	SprintDurationWeeks          int       `json:"sprintDurationWeeks"`
	SprintStartDay               string    `json:"sprintStartDay"`
	MoveIncompleteStoriesEnabled bool      `json:"moveIncompleteStoriesEnabled"`
	CreatedAt                    time.Time `json:"createdAt"`
	UpdatedAt                    time.Time `json:"updatedAt"`
}

type AppTeamStoryAutomationSettings struct {
	AutoCloseInactiveEnabled bool      `json:"autoCloseInactiveEnabled"`
	AutoCloseInactiveMonths  int       `json:"autoCloseInactiveMonths"`
	AutoArchiveEnabled       bool      `json:"autoArchiveEnabled"`
	AutoArchiveMonths        int       `json:"autoArchiveMonths"`
	CreatedAt                time.Time `json:"createdAt"`
	UpdatedAt                time.Time `json:"updatedAt"`
}

type AppTeamEstimationSettings struct {
	Scheme    string    `json:"scheme"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type AppTeamSettings struct {
	SprintSettings          AppTeamSprintSettings          `json:"sprintSettings"`
	StoryAutomationSettings AppTeamStoryAutomationSettings `json:"storyAutomationSettings"`
	EstimationSettings      AppTeamEstimationSettings      `json:"estimationSettings"`
}

type AppUpdateTeamSprintSettings struct {
	AutoCreateSprints            *bool   `json:"autoCreateSprints,omitempty"`
	UpcomingSprintsCount         *int    `json:"upcomingSprintsCount,omitempty"`
	SprintDurationWeeks          *int    `json:"sprintDurationWeeks,omitempty"`
	SprintStartDay               *string `json:"sprintStartDay,omitempty"`
	MoveIncompleteStoriesEnabled *bool   `json:"moveIncompleteStoriesEnabled,omitempty"`
}

type AppUpdateTeamStoryAutomationSettings struct {
	AutoCloseInactiveEnabled *bool `json:"autoCloseInactiveEnabled,omitempty"`
	AutoCloseInactiveMonths  *int  `json:"autoCloseInactiveMonths,omitempty"`
	AutoArchiveEnabled       *bool `json:"autoArchiveEnabled,omitempty"`
	AutoArchiveMonths        *int  `json:"autoArchiveMonths,omitempty"`
}

type AppUpdateTeamEstimationSettings struct {
	Scheme *string `json:"scheme,omitempty"`
}

// Conversion functions
func toAppTeamSprintSettings(settings teamsettings.CoreTeamSprintSettings) AppTeamSprintSettings {
	return AppTeamSprintSettings{
		AutoCreateSprints:            settings.AutoCreateSprints,
		UpcomingSprintsCount:         settings.UpcomingSprintsCount,
		SprintDurationWeeks:          settings.SprintDurationWeeks,
		SprintStartDay:               settings.SprintStartDay,
		MoveIncompleteStoriesEnabled: settings.MoveIncompleteStoriesEnabled,
		CreatedAt:                    settings.CreatedAt,
		UpdatedAt:                    settings.UpdatedAt,
	}
}

func toAppTeamStoryAutomationSettings(settings teamsettings.CoreTeamStoryAutomationSettings) AppTeamStoryAutomationSettings {
	return AppTeamStoryAutomationSettings{
		AutoCloseInactiveEnabled: settings.AutoCloseInactiveEnabled,
		AutoCloseInactiveMonths:  settings.AutoCloseInactiveMonths,
		AutoArchiveEnabled:       settings.AutoArchiveEnabled,
		AutoArchiveMonths:        settings.AutoArchiveMonths,
		CreatedAt:                settings.CreatedAt,
		UpdatedAt:                settings.UpdatedAt,
	}
}

func toAppTeamSettings(settings teamsettings.CoreTeamSettings) AppTeamSettings {
	return AppTeamSettings{
		SprintSettings:          toAppTeamSprintSettings(settings.SprintSettings),
		StoryAutomationSettings: toAppTeamStoryAutomationSettings(settings.StoryAutomationSettings),
		EstimationSettings:      toAppTeamEstimationSettings(settings.EstimationSettings),
	}
}

func toCoreUpdateTeamSprintSettings(app AppUpdateTeamSprintSettings) teamsettings.CoreUpdateTeamSprintSettings {
	return teamsettings.CoreUpdateTeamSprintSettings{
		AutoCreateSprints:            app.AutoCreateSprints,
		UpcomingSprintsCount:         app.UpcomingSprintsCount,
		SprintDurationWeeks:          app.SprintDurationWeeks,
		SprintStartDay:               app.SprintStartDay,
		MoveIncompleteStoriesEnabled: app.MoveIncompleteStoriesEnabled,
	}
}

func toCoreUpdateTeamStoryAutomationSettings(app AppUpdateTeamStoryAutomationSettings) teamsettings.CoreUpdateTeamStoryAutomationSettings {
	return teamsettings.CoreUpdateTeamStoryAutomationSettings{
		AutoCloseInactiveEnabled: app.AutoCloseInactiveEnabled,
		AutoCloseInactiveMonths:  app.AutoCloseInactiveMonths,
		AutoArchiveEnabled:       app.AutoArchiveEnabled,
		AutoArchiveMonths:        app.AutoArchiveMonths,
	}
}

func toAppTeamEstimationSettings(settings teamsettings.CoreTeamEstimationSettings) AppTeamEstimationSettings {
	return AppTeamEstimationSettings{
		Scheme:    settings.Scheme,
		CreatedAt: settings.CreatedAt,
		UpdatedAt: settings.UpdatedAt,
	}
}

func toCoreUpdateTeamEstimationSettings(app AppUpdateTeamEstimationSettings) teamsettings.CoreUpdateTeamEstimationSettings {
	return teamsettings.CoreUpdateTeamEstimationSettings{
		Scheme: app.Scheme,
	}
}
