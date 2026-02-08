package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// HumanizeJSONDecodeError converts low-level JSON decoder errors into
// user-friendly validation messages.
func HumanizeJSONDecodeError(err error) error {
	if err == nil {
		return nil
	}

	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		fieldName := strings.TrimSpace(typeErr.Field)
		if fieldName == "" {
			return errors.New("request body has an invalid value")
		}
		return fmt.Errorf("%s must be %s", fieldName, humanReadableJSONType(typeErr.Type))
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return errors.New("request body contains invalid JSON")
	}

	if errors.Is(err, io.EOF) {
		return errors.New("request body is required")
	}

	errStr := err.Error()
	const unknownFieldPrefix = "json: unknown field "
	if strings.HasPrefix(errStr, unknownFieldPrefix) {
		fieldName := strings.Trim(errStr[len(unknownFieldPrefix):], "\"")
		if fieldName == "" {
			return errors.New("request body contains unknown fields")
		}
		return fmt.Errorf("%s is not a valid field", fieldName)
	}

	return errors.New("invalid request payload")
}

// HumanizeJSONFieldDecodeError converts field-level unmarshal errors into
// user-friendly messages when decoding partial update payloads.
func HumanizeJSONFieldDecodeError(err error, fieldName string, fieldType reflect.Type) error {
	if err == nil {
		return nil
	}

	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		return fmt.Errorf("%s must be %s", fieldName, humanReadableJSONType(fieldType))
	}

	return fmt.Errorf("invalid value for %s", fieldName)
}

func humanReadableJSONType(t reflect.Type) string {
	if t == nil {
		return "a valid value"
	}

	nullable := false
	for t.Kind() == reflect.Ptr {
		nullable = true
		t = t.Elem()
	}

	expected := "a valid value"
	switch t.Kind() {
	case reflect.String:
		expected = "a string"
	case reflect.Bool:
		expected = "true or false"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		expected = "an integer"
	case reflect.Slice, reflect.Array:
		expected = "an array"
	case reflect.Struct:
		if t.PkgPath() == "github.com/google/uuid" && t.Name() == "UUID" {
			expected = "a UUID string"
		}
	}

	if nullable {
		expected += " or null"
	}

	return expected
}
