package domain

import "errors"

var (
	ErrInvalid          = errors.New("invalid key result")
	ErrForbidden        = errors.New("key result access is forbidden")
	ErrNotFound         = errors.New("key result not found")
	ErrInvalidReference = errors.New("invalid key result reference")
	ErrVersionConflict  = errors.New("key result changed since it was reviewed")
)
