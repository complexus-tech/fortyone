package epics

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestListFailsExplicitlyWithoutDurableEpicModel(t *testing.T) {
	t.Parallel()

	if err := New().List(t.Context(), uuid.New()); !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("List() error = %v, want ErrNotImplemented", err)
	}
}
