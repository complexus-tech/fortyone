package domain

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidEstimatedDuration    = errors.New("estimated duration minutes must be greater than zero")
	ErrInvalidMinimumFocusBlock    = errors.New("minimum focus block minutes must be greater than zero")
	ErrEstimatedDurationTooLarge   = errors.New("estimated duration minutes must not exceed 2400")
	ErrMinimumFocusBlockTooLarge   = errors.New("minimum focus block minutes must not exceed 2400")
	ErrFocusBlockRequiresDuration  = errors.New("minimum focus block minutes require estimated duration minutes")
	ErrFocusBlockExceedsDuration   = errors.New("minimum focus block minutes must not exceed estimated duration minutes")
	ErrInvalidAutoSchedulingStatus = errors.New("invalid auto-scheduling status")
	ErrLockedAutoSchedulingOff     = errors.New("auto-scheduling cannot be locked while disabled")
	ErrAutoSchedulingLockEmpty     = errors.New("auto-scheduling cannot be locked before Maya has scheduled work")
	ErrAutoSchedulingOwnerLocked   = errors.New("unlock auto-scheduling before changing the assignee or schedule")
)

const (
	MaximumEstimatedDurationMinutes = 40 * 60

	AutoSchedulingStatusOff        = "off"
	AutoSchedulingStatusNeedsOwner = "needs_owner"
	AutoSchedulingStatusNeedsTime  = "needs_time"
	AutoSchedulingStatusPlanning   = "planning"
	AutoSchedulingStatusScheduled  = "scheduled"
	AutoSchedulingStatusAtRisk     = "at_risk"
	AutoSchedulingStatusCannotFit  = "cannot_fit"
	AutoSchedulingStatusLocked     = "locked"
)

// ValidateScheduling validates the shared story scheduling value contract.
// Both fields are optional, but a focus block is meaningful only alongside an
// estimated duration and cannot exceed that duration.
func ValidateScheduling(estimatedDurationMinutes, minimumFocusBlockMinutes *int) error {
	if estimatedDurationMinutes != nil && *estimatedDurationMinutes <= 0 {
		return ErrInvalidEstimatedDuration
	}
	if estimatedDurationMinutes != nil && *estimatedDurationMinutes > MaximumEstimatedDurationMinutes {
		return ErrEstimatedDurationTooLarge
	}
	if minimumFocusBlockMinutes != nil && *minimumFocusBlockMinutes <= 0 {
		return ErrInvalidMinimumFocusBlock
	}
	if minimumFocusBlockMinutes != nil && *minimumFocusBlockMinutes > MaximumEstimatedDurationMinutes {
		return ErrMinimumFocusBlockTooLarge
	}
	if estimatedDurationMinutes == nil && minimumFocusBlockMinutes != nil {
		return ErrFocusBlockRequiresDuration
	}
	if estimatedDurationMinutes != nil && minimumFocusBlockMinutes != nil && *minimumFocusBlockMinutes > *estimatedDurationMinutes {
		return ErrFocusBlockExceedsDuration
	}
	return nil
}

// ValidateAutoSchedulingStatus keeps story writes aligned with the
// stories_auto_scheduling_status_check database constraint.
func ValidateAutoSchedulingStatus(status string) error {
	switch status {
	case AutoSchedulingStatusOff,
		AutoSchedulingStatusNeedsOwner,
		AutoSchedulingStatusNeedsTime,
		AutoSchedulingStatusPlanning,
		AutoSchedulingStatusScheduled,
		AutoSchedulingStatusAtRisk,
		AutoSchedulingStatusCannotFit,
		AutoSchedulingStatusLocked:
		return nil
	default:
		return fmt.Errorf("%w: %q", ErrInvalidAutoSchedulingStatus, status)
	}
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
