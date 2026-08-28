package storiesrepository

import (
	"math"
	"testing"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/stretchr/testify/require"
)

func TestCommentPageBoundsAreBoundedBeforeConversion(t *testing.T) {
	t.Parallel()

	offset, limit, err := commentPageBounds(2, maximumCommentPageSize)
	require.NoError(t, err)
	require.Equal(t, int32(maximumCommentPageSize), offset)
	require.Equal(t, int32(maximumCommentPageSize+1), limit)

	_, _, err = commentPageBounds(math.MaxInt, maximumCommentPageSize)
	require.ErrorIs(t, err, storydomain.ErrInvalidReadQuery)
	_, _, err = commentPageBounds(1, maximumCommentPageSize+1)
	require.ErrorIs(t, err, storydomain.ErrInvalidReadQuery)
}
