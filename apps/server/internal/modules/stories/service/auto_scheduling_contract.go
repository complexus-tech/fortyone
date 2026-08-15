package stories

import (
	"errors"
	"fmt"
)

const (
	AutoSchedulingStatusOff        = "off"
	AutoSchedulingStatusNeedsOwner = "needs_owner"
	AutoSchedulingStatusNeedsTime  = "needs_time"
	AutoSchedulingStatusPlanning   = "planning"
	AutoSchedulingStatusScheduled  = "scheduled"
	AutoSchedulingStatusAtRisk     = "at_risk"
	AutoSchedulingStatusCannotFit  = "cannot_fit"
	AutoSchedulingStatusLocked     = "locked"
)

var (
	ErrInvalidAutoSchedulingStatus = errors.New("invalid auto-scheduling status")
	ErrLockedAutoSchedulingOff     = errors.New("auto-scheduling cannot be locked while disabled")
	ErrAutoSchedulingLockEmpty     = errors.New("auto-scheduling cannot be locked before Maya has scheduled work")
	ErrAutoSchedulingOwnerLocked   = errors.New("unlock auto-scheduling before changing the assignee or schedule")
)

var validAutoSchedulingStatuses = map[string]struct{}{
	AutoSchedulingStatusOff:        {},
	AutoSchedulingStatusNeedsOwner: {},
	AutoSchedulingStatusNeedsTime:  {},
	AutoSchedulingStatusPlanning:   {},
	AutoSchedulingStatusScheduled:  {},
	AutoSchedulingStatusAtRisk:     {},
	AutoSchedulingStatusCannotFit:  {},
	AutoSchedulingStatusLocked:     {},
}

// ValidateAutoSchedulingStatus keeps application writes aligned with the
// stories_auto_scheduling_status_check database constraint.
func ValidateAutoSchedulingStatus(status string) error {
	if _, ok := validAutoSchedulingStatuses[status]; !ok {
		return fmt.Errorf("%w: %q", ErrInvalidAutoSchedulingStatus, status)
	}
	return nil
}

// ValidateStoryAutoSchedulingContract validates invariants that span multiple
// auto-scheduling fields. Callers should normalize an omitted create status to
// AutoSchedulingStatusOff before invoking it.
func ValidateStoryAutoSchedulingContract(enabled, locked bool, status string) error {
	if locked && !enabled {
		return ErrLockedAutoSchedulingOff
	}
	return ValidateAutoSchedulingStatus(status)
}
