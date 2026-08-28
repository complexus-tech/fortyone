package web

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	playgroundvalidator "github.com/go-playground/validator/v10"
)

var requestStructValidator = newRequestStructValidator()

// FieldViolation is a stable, value-free description of one invalid request
// field. It deliberately excludes the rejected value so secrets and personal
// data cannot be reflected into responses or logs.
type FieldViolation struct {
	Field   string `json:"field"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

// ValidationError reports all structural request violations in one pass.
// Use errors.As rather than parsing Error's stable public summary.
type ValidationError struct {
	Violations []FieldViolation
}

func (e *ValidationError) Error() string {
	return "request validation failed"
}

// ValidateStruct applies the API's single tag-based validation policy. Domain
// invariants still belong in a request's explicit Validate method or service.
func ValidateStruct(value any) error {
	if value == nil {
		return errors.New("request validation requires a non-nil value")
	}

	err := requestStructValidator.Struct(value)
	if err == nil {
		return nil
	}
	var validationErrors playgroundvalidator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return errors.New("request validation could not be performed")
	}

	violations := make([]FieldViolation, 0, len(validationErrors))
	for _, validationError := range validationErrors {
		field := requestFieldPath(validationError.Namespace())
		violations = append(violations, FieldViolation{
			Field:   field,
			Rule:    validationError.Tag(),
			Message: validationMessage(field, validationError.Tag(), validationError.Param()),
		})
	}
	return &ValidationError{Violations: violations}
}

func newRequestStructValidator() *playgroundvalidator.Validate {
	validate := playgroundvalidator.New(playgroundvalidator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		switch name {
		case "", "-":
			return field.Name
		default:
			return name
		}
	})
	return validate
}

func requestFieldPath(namespace string) string {
	_, fieldPath, found := strings.Cut(namespace, ".")
	if !found || strings.TrimSpace(fieldPath) == "" {
		return "request"
	}
	return fieldPath
}

func validationMessage(field, rule, parameter string) string {
	switch rule {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "url":
		return field + " must be a valid URL"
	case "oneof":
		return field + " must be one of the allowed values"
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, parameter)
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, parameter)
	case "len":
		return fmt.Sprintf("%s must have length %s", field, parameter)
	case "hexcolor":
		return field + " must be a valid hexadecimal color"
	default:
		return field + " is invalid"
	}
}
