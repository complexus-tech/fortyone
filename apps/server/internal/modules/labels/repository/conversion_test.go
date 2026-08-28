package labelsrepository

import (
	"errors"
	"math"
	"testing"

	labelsdomain "github.com/complexus-tech/projects-api/internal/modules/labels/domain"
)

func TestLabelPageBoundsRejectOverflow(t *testing.T) {
	t.Parallel()

	overflow := int(int64(math.MaxInt32) + 1)
	for _, test := range []struct {
		name   string
		limit  int
		offset int
	}{
		{name: "limit", limit: overflow},
		{name: "scalability bound", limit: 102},
		{name: "offset", limit: 1, offset: overflow},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := labelPageBounds(test.limit, test.offset)
			if !errors.Is(err, labelsdomain.ErrInvalidPagination) {
				t.Fatalf("labelPageBounds() error = %v, want ErrInvalidPagination", err)
			}
		})
	}
}
