package statesdomain

import "errors"

var (
	ErrNotFound         = errors.New("status not found")
	ErrNoFields         = errors.New("no status fields to update")
	ErrStatusHasStories = errors.New("cannot delete status with attached stories")
	ErrLastInCategory   = errors.New("cannot delete the last status in a category")
	ErrInvalidOrder     = errors.New("status order index is outside the supported range")
)
