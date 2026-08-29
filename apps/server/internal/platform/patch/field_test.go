package patch_test

import (
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/patch"
)

func TestFieldPreservesUpdateIntent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		field     patch.Field[int]
		specified bool
		value     *int
	}{
		{name: "omitted"},
		{name: "zero", field: patch.Set(0), specified: true, value: intPointer(0)},
		{name: "clear", field: patch.Clear[int](), specified: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			value, specified := test.field.Value()
			if specified != test.specified {
				t.Fatalf("specified = %t, want %t", specified, test.specified)
			}
			if value == nil && test.value == nil {
				return
			}
			if value == nil || test.value == nil || *value != *test.value {
				t.Fatalf("value = %v, want %v", value, test.value)
			}
		})
	}
}

func intPointer(value int) *int { return &value }
