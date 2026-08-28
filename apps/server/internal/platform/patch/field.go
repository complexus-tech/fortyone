// Package patch carries update intent independently from an update value.
//
// A Field has three states: omitted, set to a value, and explicitly cleared.
// Domain packages remain responsible for deciding which of those states are
// valid for each property.
package patch

// Field represents one field in a partial update.
type Field[T any] struct {
	specified bool
	value     *T
}

// Set returns a field containing an explicit value. Zero values are preserved.
func Set[T any](value T) Field[T] {
	return Field[T]{specified: true, value: &value}
}

// Clear returns an explicitly specified field with no value.
func Clear[T any]() Field[T] {
	return Field[T]{specified: true}
}

// Specified reports whether the caller supplied this field.
func (field Field[T]) Specified() bool {
	return field.specified
}

// Value returns the supplied value and whether the field was specified. A nil
// value with specified=true represents an explicit clear.
func (field Field[T]) Value() (*T, bool) {
	return field.value, field.specified
}
