package objectivestatusrepository

import (
	"errors"
	"math"
	"testing"

	objectivestatusdomain "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/domain"
)

func TestObjectiveStatusOrderIndexRejectsOverflow(t *testing.T) {
	t.Parallel()

	_, err := objectiveStatusOrderIndex(int(int64(math.MaxInt32) + 1))
	if !errors.Is(err, objectivestatusdomain.ErrInvalidOrder) {
		t.Fatalf("objectiveStatusOrderIndex() error = %v, want ErrInvalidOrder", err)
	}
}
