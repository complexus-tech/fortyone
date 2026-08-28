package objectivestatusdomain

import "errors"

var (
	ErrNotFound            = errors.New("objective status not found")
	ErrNoFields            = errors.New("no objective status fields to update")
	ErrStatusHasObjectives = errors.New("cannot delete status with attached objectives")
	ErrLastInCategory      = errors.New("cannot delete the last status in a category")
	ErrInvalidOrder        = errors.New("objective status order index is outside the supported range")
)
