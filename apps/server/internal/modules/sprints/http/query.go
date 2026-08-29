package sprintshttp

import (
	"math"
	"net/url"
	"strings"
	"unicode/utf8"

	sprintdomain "github.com/complexus-tech/projects-api/internal/modules/sprints/domain"
	"github.com/complexus-tech/projects-api/internal/platform/pagination"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const (
	maximumSprintSearchBytes           = 4 * sprintdomain.MaximumSearchLength
	maximumSprintIntegerParameterBytes = 20
	maximumSprintOffset                = math.MaxInt32
)

type sprintListQuery struct {
	Query sprintdomain.ListQuery
	Page  *pagination.OffsetParams
}

func parseSprintListQuery(values url.Values, workspaceID, actorID uuid.UUID) (sprintListQuery, error) {
	objectiveID, err := web.OptionalUUIDQueryParameter(values, "objectiveId")
	if err != nil {
		return sprintListQuery{}, err
	}
	teamID, err := web.OptionalUUIDQueryParameter(values, "teamId")
	if err != nil {
		return sprintListQuery{}, err
	}
	search, _, err := web.OptionalQueryParameter(values, "search", maximumSprintSearchBytes)
	if err != nil {
		return sprintListQuery{}, err
	}
	if !utf8.ValidString(search) || strings.ContainsRune(search, '\x00') {
		return sprintListQuery{}, invalidSprintQueryParameter("search")
	}

	filter := sprintdomain.ListFilter{ObjectiveID: objectiveID, TeamID: teamID, Search: search}
	page, hasPage, err := web.OptionalIntegerQueryParameter(
		values, "page", maximumSprintIntegerParameterBytes, 1, maximumSprintOffset,
	)
	if err != nil {
		return sprintListQuery{}, err
	}
	pageSize, hasPageSize, err := web.OptionalIntegerQueryParameter(
		values, "pageSize", maximumSprintIntegerParameterBytes, 1, pagination.MaximumPageSize,
	)
	if err != nil {
		return sprintListQuery{}, err
	}

	var pageParams *pagination.OffsetParams
	if hasPage || hasPageSize {
		if !hasPage {
			page = 1
		}
		if !hasPageSize {
			pageSize = pagination.DefaultMenuPageSize
		}
		if page-1 > maximumSprintOffset/pageSize {
			return sprintListQuery{}, invalidSprintQueryParameter("page")
		}
		filter.Limit = pageSize + 1
		filter.Offset = (page - 1) * pageSize
		pageParams = &pagination.OffsetParams{Page: page, PageSize: pageSize}
	}

	query, err := (sprintdomain.ListQuery{
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Filter:      filter,
	}).Normalize()
	if err != nil {
		return sprintListQuery{}, err
	}
	return sprintListQuery{Query: query, Page: pageParams}, nil
}

func invalidSprintQueryParameter(name string) error {
	return &web.QueryParameterError{Name: name, Cause: web.ErrInvalidQueryParameter}
}
