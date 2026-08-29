package activitiesrepository

import (
	"errors"
	"math"
	"testing"

	activitiesdomain "github.com/complexus-tech/projects-api/internal/modules/activities/domain"
)

func TestActivityLimitRejectsOverflow(t *testing.T) {
	t.Parallel()

	_, err := activityLimit(int(int64(math.MaxInt32) + 1))
	if !errors.Is(err, activitiesdomain.ErrInvalidLimit) {
		t.Fatalf("activityLimit() error = %v, want ErrInvalidLimit", err)
	}
}
