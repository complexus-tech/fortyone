package notifications

import "errors"

var (
	ErrInvalid   = errors.New("invalid notification request")
	ErrForbidden = errors.New("notification access is forbidden")
	ErrNotFound  = errors.New("notification not found")
	ErrConflict  = errors.New("notification request conflicts with existing state")
)
