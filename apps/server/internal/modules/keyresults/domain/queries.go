package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type GetQuery struct {
	Access      AccessScope
	KeyResultID uuid.UUID
}

func (query GetQuery) Validate() error {
	if err := query.Access.Validate(); err != nil {
		return err
	}
	if query.KeyResultID == uuid.Nil {
		return fmt.Errorf("%w: key result id is required", ErrInvalid)
	}
	return nil
}

type ObjectiveListQuery struct {
	Access      AccessScope
	ObjectiveID uuid.UUID
}

func (query ObjectiveListQuery) Validate() error {
	if err := query.Access.Validate(); err != nil {
		return err
	}
	if query.ObjectiveID == uuid.Nil {
		return fmt.Errorf("%w: objective id is required", ErrInvalid)
	}
	return nil
}

type Filters struct {
	ObjectiveIDs     []uuid.UUID
	TeamIDs          []uuid.UUID
	MeasurementTypes []string
	LeadIDs          []uuid.UUID
	CreatedBy        []uuid.UUID
	WorkspaceID      uuid.UUID
	CurrentUserID    uuid.UUID
	CreatedAfter     *time.Time
	CreatedBefore    *time.Time
	EndDateAfter     *time.Time
	EndDateBefore    *time.Time
	UpdatedAfter     *time.Time
	UpdatedBefore    *time.Time
	Page             int
	PageSize         int
	OrderBy          string
	OrderDirection   string
}

type PaginatedListQuery struct {
	Access  AccessScope
	Filters Filters
}

func (query PaginatedListQuery) Normalize() (PaginatedListQuery, error) {
	if err := query.Access.Validate(); err != nil {
		return PaginatedListQuery{}, err
	}
	filters := query.Filters
	filters.WorkspaceID = query.Access.WorkspaceID
	filters.CurrentUserID = query.Access.ActorID
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 {
		filters.PageSize = DefaultPageSize
	}
	if filters.PageSize > MaximumPageSize {
		filters.PageSize = MaximumPageSize
	}
	var err error
	if filters.ObjectiveIDs, err = normalizeUUIDs(filters.ObjectiveIDs); err != nil {
		return PaginatedListQuery{}, err
	}
	if filters.TeamIDs, err = normalizeUUIDs(filters.TeamIDs); err != nil {
		return PaginatedListQuery{}, err
	}
	if filters.LeadIDs, err = normalizeUUIDs(filters.LeadIDs); err != nil {
		return PaginatedListQuery{}, err
	}
	if filters.CreatedBy, err = normalizeUUIDs(filters.CreatedBy); err != nil {
		return PaginatedListQuery{}, err
	}
	for index, value := range filters.MeasurementTypes {
		measurement := MeasurementType(strings.TrimSpace(value))
		if !measurement.Valid() {
			return PaginatedListQuery{}, fmt.Errorf("%w: unsupported measurement filter", ErrInvalid)
		}
		filters.MeasurementTypes[index] = string(measurement)
	}
	filters.OrderBy = strings.TrimSpace(filters.OrderBy)
	filters.OrderDirection = strings.ToLower(strings.TrimSpace(filters.OrderDirection))
	if filters.OrderDirection != "asc" && filters.OrderDirection != "desc" {
		filters.OrderDirection = "desc"
	}
	switch filters.OrderBy {
	case "name", "created_at", "updated_at", "objective_name":
	default:
		filters.OrderBy = "created_at"
	}
	query.Filters = filters
	return query, nil
}

func (query PaginatedListQuery) SortKey() string {
	return query.Filters.OrderBy + "_" + query.Filters.OrderDirection
}
