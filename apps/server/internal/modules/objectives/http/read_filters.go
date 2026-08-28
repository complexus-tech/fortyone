package objectiveshttp

import (
	"math"
	"net/url"

	"github.com/complexus-tech/projects-api/internal/platform/pagination"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const (
	maximumObjectiveSearchBytes = 4 * 200
	maximumObjectiveSearchRunes = 200
	maximumObjectiveTeamIDBytes = 64
	maximumObjectiveQueryOffset = math.MaxInt32
)

type objectiveListFilters struct {
	TeamID *uuid.UUID
	Search string
}

func parseObjectiveListFilters(query url.Values) (objectiveListFilters, error) {
	search, _, err := web.OptionalTextQueryParameter(
		query,
		"search",
		maximumObjectiveSearchBytes,
		maximumObjectiveSearchRunes,
	)
	if err != nil {
		return objectiveListFilters{}, err
	}
	teamID, err := parseObjectiveTeamID(query)
	if err != nil {
		return objectiveListFilters{}, err
	}
	return objectiveListFilters{TeamID: teamID, Search: search}, nil
}

func parseObjectiveTeamID(query url.Values) (*uuid.UUID, error) {
	value, present, err := web.OptionalTextQueryParameter(
		query,
		"teamId",
		maximumObjectiveTeamIDBytes,
		maximumObjectiveTeamIDBytes,
	)
	if err != nil {
		return nil, err
	}
	if !present {
		return nil, nil
	}
	teamID, err := uuid.Parse(value)
	if err != nil || teamID == uuid.Nil {
		return nil, &web.QueryParameterError{Name: "teamId", Cause: web.ErrInvalidQueryParameter}
	}
	return &teamID, nil
}

func parseObjectiveListPagination(query url.Values) (*pagination.OffsetParams, error) {
	if !pagination.OffsetRequested(query) {
		return nil, nil
	}
	params, err := pagination.ParseOffsetQuery(query, pagination.OffsetQueryConfig{
		DefaultPageSize: pagination.DefaultMenuPageSize,
		MaximumPageSize: pagination.MaximumPageSize,
		MaximumOffset:   maximumObjectiveQueryOffset,
	})
	if err != nil {
		return nil, err
	}
	return &params, nil
}

func parseObjectiveActivityPagination(query url.Values) (pagination.OffsetParams, error) {
	return pagination.ParseOffsetQuery(query, pagination.OffsetQueryConfig{
		DefaultPageSize: 20,
		MaximumPageSize: pagination.MaximumPageSize,
		MaximumOffset:   maximumObjectiveQueryOffset,
	})
}
