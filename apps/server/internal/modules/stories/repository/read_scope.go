package storiesrepository

import (
	"errors"
	"fmt"
	"strings"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
)

var (
	errReadRepositoryNotConfigured = errors.New("story read repository is not configured")
	errInvalidReadQuery            = storydomain.ErrInvalidReadQuery
)

const (
	maxMyStoriesResultCount = 1_000
	maxCategoryPage         = 10_000
	maxCategoryPageSize     = 100
)

type storyCategory string

const (
	storyCategoryBacklog   storyCategory = "backlog"
	storyCategoryUnstarted storyCategory = "unstarted"
	storyCategoryStarted   storyCategory = "started"
	storyCategoryPaused    storyCategory = "paused"
	storyCategoryCompleted storyCategory = "completed"
	storyCategoryCancelled storyCategory = "cancelled"
)

func validateReadScope(scope storydomain.ReadScope) error {
	if err := scope.Validate(); err != nil {
		return fmt.Errorf("%w: %v", errInvalidReadQuery, err)
	}
	return nil
}

func parseStoryCategory(value string) (storyCategory, error) {
	category := storyCategory(strings.ToLower(strings.TrimSpace(value)))
	switch category {
	case storyCategoryBacklog,
		storyCategoryUnstarted,
		storyCategoryStarted,
		storyCategoryPaused,
		storyCategoryCompleted,
		storyCategoryCancelled:
		return category, nil
	default:
		return "", fmt.Errorf("%w: unsupported category", errInvalidReadQuery)
	}
}

func categoryPage(page, pageSize int) (offset int32, limit int32, err error) {
	if page < 1 || page > maxCategoryPage {
		return 0, 0, fmt.Errorf("%w: page must be between 1 and %d", errInvalidReadQuery, maxCategoryPage)
	}
	if pageSize < 1 || pageSize > maxCategoryPageSize {
		return 0, 0, fmt.Errorf("%w: page size must be between 1 and %d", errInvalidReadQuery, maxCategoryPageSize)
	}
	offset, err = safecast.Int64ToInt32(int64(page-1) * int64(pageSize))
	if err != nil {
		return 0, 0, fmt.Errorf("%w: category page offset is outside the supported window", errInvalidReadQuery)
	}
	limit, err = safecast.Int64ToInt32(int64(pageSize) + 1)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: category page size is outside the supported window", errInvalidReadQuery)
	}
	return offset, limit, nil
}
