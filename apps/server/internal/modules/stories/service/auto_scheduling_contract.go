package stories

import storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"

const (
	AutoSchedulingStatusOff        = storydomain.AutoSchedulingStatusOff
	AutoSchedulingStatusNeedsOwner = storydomain.AutoSchedulingStatusNeedsOwner
	AutoSchedulingStatusNeedsTime  = storydomain.AutoSchedulingStatusNeedsTime
	AutoSchedulingStatusPlanning   = storydomain.AutoSchedulingStatusPlanning
	AutoSchedulingStatusScheduled  = storydomain.AutoSchedulingStatusScheduled
	AutoSchedulingStatusAtRisk     = storydomain.AutoSchedulingStatusAtRisk
	AutoSchedulingStatusCannotFit  = storydomain.AutoSchedulingStatusCannotFit
	AutoSchedulingStatusLocked     = storydomain.AutoSchedulingStatusLocked
)

var (
	ErrInvalidAutoSchedulingStatus = storydomain.ErrInvalidAutoSchedulingStatus
	ErrLockedAutoSchedulingOff     = storydomain.ErrLockedAutoSchedulingOff
	ErrAutoSchedulingLockEmpty     = storydomain.ErrAutoSchedulingLockEmpty
	ErrAutoSchedulingOwnerLocked   = storydomain.ErrAutoSchedulingOwnerLocked
)

// ValidateAutoSchedulingStatus keeps application writes aligned with the
// stories_auto_scheduling_status_check database constraint.
func ValidateAutoSchedulingStatus(status string) error {
	return storydomain.ValidateAutoSchedulingStatus(status)
}

// ValidateStoryAutoSchedulingContract validates invariants that span multiple
// auto-scheduling fields. Callers should normalize an omitted create status to
// AutoSchedulingStatusOff before invoking it.
func ValidateStoryAutoSchedulingContract(enabled, locked bool, status string) error {
	return storydomain.ValidateStoryAutoSchedulingContract(enabled, locked, status)
}
