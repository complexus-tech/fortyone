package objectivesrepository

import (
	"context"
	"errors"
	"math"
	"testing"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	"github.com/google/uuid"
)

func TestKeyResultSequenceArithmeticRejectsOverflow(t *testing.T) {
	t.Parallel()

	count, first, err := keyResultSequenceRange(12, 3)
	if err != nil || count != 3 || first != 10 {
		t.Fatalf("keyResultSequenceRange() = (%d, %d, %v), want (3, 10, nil)", count, first, err)
	}
	sequence, err := keyResultSequenceAt(first, 2)
	if err != nil || sequence != 12 {
		t.Fatalf("keyResultSequenceAt() = (%d, %v), want (12, nil)", sequence, err)
	}

	tests := []struct {
		name  string
		check func() error
	}{
		{name: "zero count", check: func() error { _, _, err := keyResultSequenceRange(1, 0); return err }},
		{name: "range underflow", check: func() error { _, _, err := keyResultSequenceRange(1, 2); return err }},
		{name: "addition overflow", check: func() error { _, err := keyResultSequenceAt(math.MaxInt32, 1); return err }},
		{name: "negative index", check: func() error { _, err := keyResultSequenceAt(1, -1); return err }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := test.check(); !errors.Is(err, objectivesdomain.ErrInvalid) {
				t.Fatalf("error = %v, want ErrInvalid", err)
			}
		})
	}
}

func TestRepositoryRejectsPaginationAndOrderOverflowBeforeSQL(t *testing.T) {
	t.Parallel()

	repository := &Repository{}
	_, err := repository.List(context.Background(), objectivesdomain.ListQuery{
		WorkspaceID: uuid.New(), ActorID: uuid.New(), Offset: math.MaxInt,
	})
	if !errors.Is(err, objectivesdomain.ErrInvalid) {
		t.Fatalf("overflowing list offset error = %v, want ErrInvalid", err)
	}

	query := objectivesdomain.StrategyQuery{WorkspaceID: uuid.New(), ActorID: uuid.New()}
	_, err = repository.CreateStrategicPillar(context.Background(), query, objectivesdomain.NewStrategicPillar{
		Name: "Overflow", OrderIndex: math.MaxInt,
	})
	if !errors.Is(err, objectivesdomain.ErrInvalid) {
		t.Fatalf("overflowing create order error = %v, want ErrInvalid", err)
	}
	_, err = repository.UpdateStrategicPillar(context.Background(), query, uuid.New(), objectivesdomain.UpdateStrategicPillar{
		OrderIndex: objectivesdomain.SetField(math.MaxInt),
	})
	if !errors.Is(err, objectivesdomain.ErrInvalid) {
		t.Fatalf("overflowing update order error = %v, want ErrInvalid", err)
	}
}
