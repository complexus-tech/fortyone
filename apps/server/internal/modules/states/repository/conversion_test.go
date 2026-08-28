package statesrepository

import (
	"errors"
	"math"
	"testing"

	statesdomain "github.com/complexus-tech/projects-api/internal/modules/states/domain"
)

func TestStateOrderIndexRejectsOverflow(t *testing.T) {
	t.Parallel()

	_, err := stateOrderIndex(int(int64(math.MaxInt32) + 1))
	if !errors.Is(err, statesdomain.ErrInvalidOrder) {
		t.Fatalf("stateOrderIndex() error = %v, want ErrInvalidOrder", err)
	}
}
