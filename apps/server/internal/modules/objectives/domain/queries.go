package domain

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const (
	DefaultPageSize = 20
	MaximumPageSize = 100
)

type ListQuery struct {
	WorkspaceID uuid.UUID
	ActorID     uuid.UUID
	ObjectiveID *uuid.UUID
	TeamID      *uuid.UUID
	Search      string
	Limit       int
	Offset      int
}

func (query ListQuery) Normalize() (ListQuery, error) {
	if query.WorkspaceID == uuid.Nil || query.ActorID == uuid.Nil {
		return ListQuery{}, fmt.Errorf("%w: workspace and actor are required", ErrInvalid)
	}
	query.Search = strings.TrimSpace(query.Search)
	if len([]rune(query.Search)) > 200 {
		return ListQuery{}, fmt.Errorf("%w: search cannot exceed 200 characters", ErrInvalid)
	}
	if query.Offset < 0 || query.Limit < 0 {
		return ListQuery{}, fmt.Errorf("%w: pagination cannot be negative", ErrInvalid)
	}
	if query.Limit > MaximumPageSize+1 {
		query.Limit = MaximumPageSize + 1
	}
	return query, nil
}

type GetQuery struct {
	ObjectiveID uuid.UUID
	WorkspaceID uuid.UUID
	ActorID     uuid.UUID
	Internal    bool
}

func (query GetQuery) Validate() error {
	if query.ObjectiveID == uuid.Nil || query.WorkspaceID == uuid.Nil || (!query.Internal && query.ActorID == uuid.Nil) {
		return fmt.Errorf("%w: objective, workspace, and actor are required", ErrInvalid)
	}
	return nil
}

type AnalyticsQuery struct {
	ObjectiveID uuid.UUID
	WorkspaceID uuid.UUID
	ActorID     uuid.UUID
}

func (query AnalyticsQuery) Validate() error {
	if query.ObjectiveID == uuid.Nil || query.WorkspaceID == uuid.Nil || query.ActorID == uuid.Nil {
		return fmt.Errorf("%w: objective, workspace, and actor are required", ErrInvalid)
	}
	return nil
}
