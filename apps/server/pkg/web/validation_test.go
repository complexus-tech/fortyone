package web

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateStructUsesJSONPathsAndNeverReflectsValues(t *testing.T) {
	t.Parallel()

	type invite struct {
		Email string `json:"email" validate:"required,email"`
	}
	type request struct {
		Invitations []invite `json:"invitations" validate:"required,dive"`
		Role        string   `json:"role" validate:"oneof=member admin"`
	}

	const rejectedEmail = "not-an-email-sensitive@example"
	err := ValidateStruct(request{
		Invitations: []invite{{Email: rejectedEmail}},
		Role:        "owner-secret-value",
	})
	var validationError *ValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("ValidateStruct() error = %v, want ValidationError", err)
	}
	if len(validationError.Violations) != 2 {
		t.Fatalf("violations = %#v, want 2", validationError.Violations)
	}
	if validationError.Violations[0].Field != "invitations[0].email" {
		t.Fatalf("first field = %q, want invitations[0].email", validationError.Violations[0].Field)
	}
	for _, violation := range validationError.Violations {
		if containsAny(violation.Message, rejectedEmail, "owner-secret-value") {
			t.Fatalf("violation reflects rejected value: %#v", violation)
		}
	}
}

func TestValidateStructAcceptsValidTaggedRequest(t *testing.T) {
	t.Parallel()

	input := struct {
		Name string `json:"name" validate:"required,min=3,max=10"`
	}{Name: "Ada"}
	if err := ValidateStruct(input); err != nil {
		t.Fatalf("ValidateStruct() error = %v", err)
	}
}

func containsAny(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if candidate != "" && strings.Contains(value, candidate) {
			return true
		}
	}
	return false
}
