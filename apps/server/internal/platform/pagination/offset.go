package pagination

import (
	"errors"
	"math"
	"net/url"

	"github.com/complexus-tech/projects-api/pkg/web"
)

const (
	DefaultMenuPageSize          = 15
	MaximumPageSize              = 100
	maximumIntegerParameterBytes = 20
)

var ErrInvalidOffsetQueryConfig = errors.New("invalid offset query configuration")

type OffsetParams struct {
	Page     int
	PageSize int
}

// OffsetQueryConfig defines an endpoint's explicit offset-pagination contract.
// MaximumOffset should match the persistence adapter's representable range.
type OffsetQueryConfig struct {
	DefaultPageSize int
	MaximumPageSize int
	MaximumOffset   int
}

func OffsetRequested(query url.Values) bool {
	return query.Has("page") || query.Has("pageSize")
}

// ParseOffsetQuery parses the strict offset-pagination contract used by
// externally reachable APIs. Explicit malformed or repeated values fail closed,
// while a valid page size above the documented maximum retains the established
// cap behavior.
func ParseOffsetQuery(query url.Values, config OffsetQueryConfig) (OffsetParams, error) {
	if config.DefaultPageSize < 1 || config.MaximumPageSize < config.DefaultPageSize || config.MaximumOffset < 0 {
		return OffsetParams{}, ErrInvalidOffsetQueryConfig
	}

	page, present, err := web.OptionalIntegerQueryParameter(
		query, "page", maximumIntegerParameterBytes, 1, math.MaxInt,
	)
	if err != nil {
		return OffsetParams{}, err
	}
	if !present {
		page = 1
	}
	pageSize, present, err := web.OptionalIntegerQueryParameter(
		query, "pageSize", maximumIntegerParameterBytes, 1, math.MaxInt,
	)
	if err != nil {
		return OffsetParams{}, err
	}
	if !present {
		pageSize = config.DefaultPageSize
	}
	if pageSize > config.MaximumPageSize {
		pageSize = config.MaximumPageSize
	}
	if page-1 > config.MaximumOffset/pageSize {
		return OffsetParams{}, &web.QueryParameterError{Name: "page", Cause: web.ErrInvalidQueryParameter}
	}
	return OffsetParams{Page: page, PageSize: pageSize}, nil
}

func (p OffsetParams) Offset() int {
	if p.Page <= 1 || p.PageSize <= 0 {
		return 0
	}
	if p.Page-1 > math.MaxInt/p.PageSize {
		return math.MaxInt
	}
	return (p.Page - 1) * p.PageSize
}
