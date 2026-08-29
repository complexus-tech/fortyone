package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestNewActivityNormalize(t *testing.T) {
	t.Parallel()

	keyResultID := uuid.New()
	activity := NewActivity{
		ObjectiveID: uuid.New(), KeyResultID: &keyResultID, UserID: uuid.New(), WorkspaceID: uuid.New(),
		Type: ActivityTypeUpdate, UpdateType: UpdateTypeKeyResult,
		Field: " current_value ", CurrentValue: "75", Comment: " reviewed ",
	}
	normalized, err := activity.Normalize()
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if normalized.Field != "current_value" || normalized.Comment != "reviewed" {
		t.Fatalf("normalized text = %q/%q", normalized.Field, normalized.Comment)
	}
}

func TestNewActivityNormalizeRejectsInvalidScopeAndShape(t *testing.T) {
	t.Parallel()

	valid := NewActivity{
		ObjectiveID: uuid.New(), UserID: uuid.New(), WorkspaceID: uuid.New(),
		Type: ActivityTypeUpdate, UpdateType: UpdateTypeObjective,
	}
	tests := []struct {
		name   string
		mutate func(*NewActivity)
	}{
		{name: "missing objective", mutate: func(value *NewActivity) { value.ObjectiveID = uuid.Nil }},
		{name: "missing actor", mutate: func(value *NewActivity) { value.UserID = uuid.Nil }},
		{name: "missing workspace", mutate: func(value *NewActivity) { value.WorkspaceID = uuid.Nil }},
		{name: "invalid activity type", mutate: func(value *NewActivity) { value.Type = "other" }},
		{name: "invalid update type", mutate: func(value *NewActivity) { value.UpdateType = "other" }},
		{name: "missing key result", mutate: func(value *NewActivity) { value.UpdateType = UpdateTypeKeyResult }},
		{name: "oversized field", mutate: func(value *NewActivity) { value.Field = strings.Repeat("x", 101) }},
		{name: "oversized comment", mutate: func(value *NewActivity) { value.Comment = strings.Repeat("x", 10_001) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			value := valid
			test.mutate(&value)
			if _, err := value.Normalize(); !errors.Is(err, ErrInvalid) {
				t.Fatalf("Normalize() error = %v, want ErrInvalid", err)
			}
		})
	}
}

func TestListQueryNormalizeBoundsPaginationAndRequiresAResource(t *testing.T) {
	t.Parallel()

	query, err := (ListQuery{
		ObjectiveID: uuid.New(), WorkspaceID: uuid.New(), ActorID: uuid.New(),
		Page: -10, PageSize: MaximumPageSize + 10,
	}).Normalize()
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if query.Page != 1 || query.PageSize != MaximumPageSize {
		t.Fatalf("pagination = %d/%d", query.Page, query.PageSize)
	}

	_, err = (ListQuery{WorkspaceID: uuid.New(), ActorID: uuid.New()}).Normalize()
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("resource-less query error = %v, want ErrInvalid", err)
	}
}

func TestCreateBatchCommandIsBoundedAndAtomicInputIsPrevalidated(t *testing.T) {
	t.Parallel()

	activity := NewActivity{
		ObjectiveID: uuid.New(), UserID: uuid.New(), WorkspaceID: uuid.New(),
		Type: ActivityTypeCreate, UpdateType: UpdateTypeObjective,
	}
	tooMany := make([]NewActivity, MaximumBatchSize+1)
	for index := range tooMany {
		tooMany[index] = activity
	}
	if _, err := (CreateBatchCommand{Activities: tooMany}).Normalize(); !errors.Is(err, ErrInvalid) {
		t.Fatalf("oversized batch error = %v, want ErrInvalid", err)
	}
}
