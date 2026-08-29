package searchrepository

import (
	"errors"
	"fmt"
	"math"

	"github.com/complexus-tech/projects-api/internal/platform/safecast"
)

func databasePage(page, pageSize int) (int32, int32, error) {
	if page < 1 || pageSize < 1 {
		return 0, 0, errors.New("search page and page size must be positive")
	}
	zeroBasedPage, databasePageSize := int64(page-1), int64(pageSize)
	if zeroBasedPage > math.MaxInt64/databasePageSize {
		return 0, 0, fmt.Errorf("validate search page offset: %w", safecast.ErrOutOfRange)
	}
	offset, err := safecast.Int64ToInt32(zeroBasedPage * databasePageSize)
	if err != nil {
		return 0, 0, fmt.Errorf("validate search page offset: %w", err)
	}
	limit, err := safecast.Int32(pageSize)
	if err != nil {
		return 0, 0, fmt.Errorf("validate search page size: %w", err)
	}
	return offset, limit, nil
}
