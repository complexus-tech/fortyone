package usersdomain

import (
	"time"

	"github.com/google/uuid"
)

// InactivityWarningCursor is the stable keyset position used while paging
// account-lifecycle warning candidates. LastLoginAt and UserID mirror the
// repository's total ordering.
type InactivityWarningCursor struct {
	LastLoginAt time.Time
	UserID      uuid.UUID
	Valid       bool
}

// InactivityWarningQuery describes one bounded page of accounts that have
// crossed the application-provided inactivity cutoff.
type InactivityWarningQuery struct {
	InactiveBefore time.Time
	Cursor         InactivityWarningCursor
	BatchSize      int
}

// InactivityWarningCandidate contains only the account-owned fields required
// to address an inactivity warning.
type InactivityWarningCandidate struct {
	UserID      uuid.UUID
	Email       string
	FullName    string
	LastLoginAt time.Time
}

// InactivityWarningEligibility identifies an account whose eligibility must be
// rechecked immediately before delivery.
type InactivityWarningEligibility struct {
	UserID         uuid.UUID
	InactiveBefore time.Time
}

// InactivityWarningReceipt records the application-clock instant used after
// the mail transport accepts an inactivity warning.
type InactivityWarningReceipt struct {
	UserID         uuid.UUID
	InactiveBefore time.Time
	WarningSentAt  time.Time
}
