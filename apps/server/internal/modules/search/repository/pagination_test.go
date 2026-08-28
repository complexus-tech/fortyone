package searchrepository

import (
	"errors"
	"math"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/stretchr/testify/require"
)

func TestDatabasePageRejectsInvalidAndOverflowingOffsets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		page     int
		pageSize int
		wantErr  error
	}{
		{name: "first page", page: 1, pageSize: 20},
		{name: "later page", page: 3, pageSize: 25},
		{name: "missing page", page: 0, pageSize: 20, wantErr: errors.New("invalid")},
		{name: "missing size", page: 1, pageSize: 0, wantErr: errors.New("invalid")},
		{name: "offset overflow", page: math.MaxInt, pageSize: math.MaxInt, wantErr: safecast.ErrOutOfRange},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			offset, limit, err := databasePage(test.page, test.pageSize)
			if test.wantErr != nil {
				require.Error(t, err)
				if errors.Is(test.wantErr, safecast.ErrOutOfRange) {
					require.ErrorIs(t, err, safecast.ErrOutOfRange)
				}
				return
			}
			require.NoError(t, err)
			require.Equal(t, int32((test.page-1)*test.pageSize), offset)
			require.Equal(t, int32(test.pageSize), limit)
		})
	}
}
