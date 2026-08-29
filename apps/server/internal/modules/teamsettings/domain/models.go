package teamsettingsdomain

import (
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
)

type SprintSettings struct {
	TeamID                       uuid.UUID
	WorkspaceID                  uuid.UUID
	AutoCreateSprints            bool
	UpcomingSprintsCount         int
	SprintDurationWeeks          int
	SprintStartDay               string
	WorkingDays                  []int
	MoveIncompleteStoriesEnabled bool
	LastAutoSprintNumber         int
	NextAutoSprintNumber         int
	AutoCreateDisabledAt         *time.Time
	AutoCreateDisabledReason     *string
	CreatedAt                    time.Time
	UpdatedAt                    time.Time
}

type StoryAutomationSettings struct {
	TeamID                   uuid.UUID
	WorkspaceID              uuid.UUID
	AutoCloseInactiveEnabled bool
	AutoCloseInactiveMonths  int
	AutoArchiveEnabled       bool
	AutoArchiveMonths        int
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

type EstimationSettings struct {
	TeamID      uuid.UUID
	WorkspaceID uuid.UUID
	Scheme      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Settings struct {
	SprintSettings          SprintSettings
	StoryAutomationSettings StoryAutomationSettings
	EstimationSettings      EstimationSettings
}

// Core aliases preserve the existing application vocabulary while the domain
// package remains the dependency boundary shared by services and adapters.
type CoreTeamSprintSettings = SprintSettings
type CoreTeamStoryAutomationSettings = StoryAutomationSettings
type CoreTeamEstimationSettings = EstimationSettings
type CoreTeamSettings = Settings

// PatchField carries update intent separately from its value. This keeps zero
// values distinct from omitted fields while allowing sqlc to generate static,
// fully typed update statements.
type PatchField[T any] struct {
	Value   T
	Present bool
}

func PatchFromPointer[T any](value *T) PatchField[T] {
	if value == nil {
		return PatchField[T]{}
	}
	return PatchField[T]{Value: *value, Present: true}
}

type SprintSettingsUpdate struct {
	AutoCreateSprints            PatchField[bool]
	UpcomingSprintsCount         PatchField[int]
	SprintDurationWeeks          PatchField[int]
	SprintStartDay               PatchField[string]
	WorkingDays                  PatchField[[]int]
	MoveIncompleteStoriesEnabled PatchField[bool]
	NextAutoSprintNumber         PatchField[int]
}

func (u SprintSettingsUpdate) Empty() bool {
	return !u.AutoCreateSprints.Present &&
		!u.UpcomingSprintsCount.Present &&
		!u.SprintDurationWeeks.Present &&
		!u.SprintStartDay.Present &&
		!u.WorkingDays.Present &&
		!u.MoveIncompleteStoriesEnabled.Present &&
		!u.NextAutoSprintNumber.Present
}

type StoryAutomationSettingsUpdate struct {
	AutoCloseInactiveEnabled PatchField[bool]
	AutoCloseInactiveMonths  PatchField[int]
	AutoArchiveEnabled       PatchField[bool]
	AutoArchiveMonths        PatchField[int]
}

func (u StoryAutomationSettingsUpdate) Empty() bool {
	return !u.AutoCloseInactiveEnabled.Present &&
		!u.AutoCloseInactiveMonths.Present &&
		!u.AutoArchiveEnabled.Present &&
		!u.AutoArchiveMonths.Present
}

type EstimationSettingsUpdate struct {
	Scheme PatchField[string]
}

type CoreUpdateTeamSprintSettings = SprintSettingsUpdate
type CoreUpdateTeamStoryAutomationSettings = StoryAutomationSettingsUpdate
type CoreUpdateTeamEstimationSettings = EstimationSettingsUpdate

func (u EstimationSettingsUpdate) Empty() bool {
	return !u.Scheme.Present
}

// Access contains every caller-controlled input used by the service policy.
// Repository predicates independently repeat tenant and team ownership checks.
type Access struct {
	Actor         platformauth.Actor
	WorkspaceRole authorization.WorkspaceRole
	WorkspaceID   uuid.UUID
	TeamID        uuid.UUID
}

type AuditActor struct {
	Type string
	ID   *uuid.UUID
}

func UserAuditActor(actor platformauth.Actor) AuditActor {
	actorID := actor.PrincipalID
	return AuditActor{Type: string(actor.Kind), ID: &actorID}
}

func SystemAuditActor() AuditActor {
	return AuditActor{Type: string(platformauth.PrincipalSystem)}
}
