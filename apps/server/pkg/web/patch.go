package web

import (
	"bytes"
	"encoding/json"
)

// PatchField preserves the three states of a JSON partial update: omitted, a
// concrete value (including a zero value), and explicit null. Handlers map this
// transport type into their module's domain patch type.
type PatchField[T any] struct {
	specified bool
	value     *T
}

// UnmarshalJSON records field presence before decoding its optional value.
func (field *PatchField[T]) UnmarshalJSON(data []byte) error {
	field.specified = true
	field.value = nil
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return nil
	}

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	field.value = &value
	return nil
}

// Specified reports whether the JSON object contained this property.
func (field PatchField[T]) Specified() bool {
	return field.specified
}

// Value returns the decoded value and whether the property was present. A nil
// value with specified=true represents JSON null.
func (field PatchField[T]) Value() (*T, bool) {
	return field.value, field.specified
}
