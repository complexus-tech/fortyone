package adminrepository

import (
	"fmt"

	admindomain "github.com/complexus-tech/projects-api/internal/modules/admin/domain"
	"github.com/complexus-tech/projects-api/internal/platform/pagination"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
)

type sqlPage struct {
	limit  int32
	offset int32
}

func newSQLPage(page pagination.OffsetParams) (sqlPage, error) {
	if page.Page < 1 || page.PageSize < 1 || page.PageSize > pagination.MaximumPageSize {
		return sqlPage{}, admindomain.ErrInvalidPagination
	}
	limit, err := safecast.Int32(page.PageSize)
	if err != nil {
		return sqlPage{}, fmt.Errorf("%w: page size: %v", admindomain.ErrInvalidPagination, err)
	}
	offset, err := safecast.Int32(page.Offset())
	if err != nil {
		return sqlPage{}, fmt.Errorf("%w: page offset: %v", admindomain.ErrInvalidPagination, err)
	}
	return sqlPage{limit: limit, offset: offset}, nil
}

func paginationResult(page pagination.OffsetParams, total int64) (admindomain.Pagination, error) {
	totalCount, err := safecast.Int64(total)
	if err != nil {
		return admindomain.Pagination{}, fmt.Errorf("map admin total: %w", err)
	}
	return admindomain.Pagination{
		Total:  totalCount,
		Page:   page.Page,
		Limit:  page.PageSize,
		Offset: page.Offset(),
	}, nil
}
