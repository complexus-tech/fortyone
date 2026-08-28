package sprints

import (
	sprintdomain "github.com/complexus-tech/projects-api/internal/modules/sprints/domain"
	platformpatch "github.com/complexus-tech/projects-api/internal/platform/patch"
)

// Compatibility aliases keep established in-process consumers source-level
// names while the sprint domain package owns the underlying concepts.
type CoreSprint = sprintdomain.Sprint
type CoreNewSprint = sprintdomain.NewSprint
type CoreSprintAnalytics = sprintdomain.Analytics
type CoreSprintOverview = sprintdomain.Overview
type CoreStoryBreakdown = sprintdomain.StoryBreakdown
type CoreBurndownDataPoint = sprintdomain.BurndownDataPoint
type CoreTeamMemberAllocation = sprintdomain.TeamMemberAllocation
type SprintPatch = sprintdomain.Patch

var (
	ErrInvalid          = sprintdomain.ErrInvalid
	ErrForbidden        = sprintdomain.ErrForbidden
	ErrNotFound         = sprintdomain.ErrNotFound
	ErrVersionConflict  = sprintdomain.ErrVersionConflict
	ErrInvalidReference = sprintdomain.ErrInvalidReference
)

func SetField[T any](value T) platformpatch.Field[T] {
	return platformpatch.Set(value)
}

func ClearField[T any]() platformpatch.Field[T] {
	return platformpatch.Clear[T]()
}
