package teamsettings

import (
	teamsettingsdomain "github.com/complexus-tech/projects-api/internal/modules/teamsettings/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
)

type CoreTeamSprintSettings = teamsettingsdomain.SprintSettings
type CoreTeamStoryAutomationSettings = teamsettingsdomain.StoryAutomationSettings
type CoreTeamEstimationSettings = teamsettingsdomain.EstimationSettings
type CoreTeamSettings = teamsettingsdomain.Settings

type PatchField[T any] = teamsettingsdomain.PatchField[T]

func PatchFromPointer[T any](value *T) PatchField[T] {
	return teamsettingsdomain.PatchFromPointer(value)
}

type CoreUpdateTeamSprintSettings = teamsettingsdomain.SprintSettingsUpdate
type CoreUpdateTeamStoryAutomationSettings = teamsettingsdomain.StoryAutomationSettingsUpdate
type CoreUpdateTeamEstimationSettings = teamsettingsdomain.EstimationSettingsUpdate
type Access = teamsettingsdomain.Access
type AuditActor = teamsettingsdomain.AuditActor

func UserAuditActor(actor platformauth.Actor) AuditActor {
	return teamsettingsdomain.UserAuditActor(actor)
}

func SystemAuditActor() AuditActor {
	return teamsettingsdomain.SystemAuditActor()
}
