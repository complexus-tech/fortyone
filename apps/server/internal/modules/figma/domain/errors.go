package domain

import "errors"

var (
	ErrNotFound  = errors.New("figma resource was not found")
	ErrConflict  = errors.New("figma resource changed concurrently")
	ErrForbidden = errors.New("figma resource is outside the authorized workspace")
)
