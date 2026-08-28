package domain

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const (
	DefaultListLimit    = 200
	MaximumListLimit    = 500
	MaximumSearchLength = 255
)

// ListFilter is the complete, typed sprint-list query surface.
type ListFilter struct {
	SprintID    *uuid.UUID
	ObjectiveID *uuid.UUID
	TeamID      *uuid.UUID
	Search      string
	Limit       int
	Offset      int
}

func (filter ListFilter) Normalize() (ListFilter, error) {
	for _, value := range []*uuid.UUID{filter.SprintID, filter.ObjectiveID, filter.TeamID} {
		if value != nil && *value == uuid.Nil {
			return ListFilter{}, fmt.Errorf("%w: filters cannot contain a zero id", ErrInvalid)
		}
	}
	filter.Search = strings.TrimSpace(filter.Search)
	if len([]rune(filter.Search)) > MaximumSearchLength {
		return ListFilter{}, fmt.Errorf("%w: search cannot exceed %d characters", ErrInvalid, MaximumSearchLength)
	}
	if filter.Offset < 0 {
		return ListFilter{}, fmt.Errorf("%w: offset cannot be negative", ErrInvalid)
	}
	if filter.Limit == 0 {
		filter.Limit = DefaultListLimit
	}
	if filter.Limit < 1 || filter.Limit > MaximumListLimit {
		return ListFilter{}, fmt.Errorf("%w: limit must be between 1 and %d", ErrInvalid, MaximumListLimit)
	}
	return filter, nil
}

// ListQuery binds a finite filter to a workspace member.
type ListQuery struct {
	WorkspaceID uuid.UUID
	ActorID     uuid.UUID
	Filter      ListFilter
}

func (query ListQuery) Normalize() (ListQuery, error) {
	if query.WorkspaceID == uuid.Nil || query.ActorID == uuid.Nil {
		return ListQuery{}, fmt.Errorf("%w: workspace and actor are required", ErrInvalid)
	}
	filter, err := query.Filter.Normalize()
	if err != nil {
		return ListQuery{}, err
	}
	query.Filter = filter
	return query, nil
}
