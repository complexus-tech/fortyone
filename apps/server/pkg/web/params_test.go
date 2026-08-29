package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestUUIDPathParameter(t *testing.T) {
	t.Parallel()

	validID := uuid.New()
	tests := []struct {
		name  string
		value string
		want  uuid.UUID
		cause error
	}{
		{name: "valid", value: validID.String(), want: validID},
		{name: "trimmed", value: "  " + validID.String() + "  ", want: validID},
		{name: "missing", cause: ErrMissingPathParameter},
		{name: "malformed", value: "raw-secret-value", cause: ErrInvalidPathParameter},
		{name: "zero", value: uuid.Nil.String(), cause: ErrInvalidPathParameter},
		{name: "bounded", value: strings.Repeat("x", DefaultMaxPathParameterBytes+1), cause: ErrPathParameterTooLong},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.SetPathValue("id", test.value)
			got, err := UUIDPathParameter(request, "id")
			if test.cause == nil {
				if err != nil || got != test.want {
					t.Fatalf("UUIDPathParameter = %s, %v; want %s, nil", got, err, test.want)
				}
				return
			}
			if !errors.Is(err, test.cause) {
				t.Fatalf("UUIDPathParameter error = %v, want %v", err, test.cause)
			}
			if test.value != "" && strings.Contains(err.Error(), strings.TrimSpace(test.value)) {
				t.Fatal("parameter error exposed the raw path value")
			}
		})
	}
}

func FuzzUUIDPathParameterDoesNotExposeInput(f *testing.F) {
	f.Add("not-a-uuid")
	f.Add(uuid.NewString())
	f.Fuzz(func(t *testing.T, input string) {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.SetPathValue("id", input)
		_, err := UUIDPathParameter(request, "id")
		if err == nil {
			return
		}
		switch {
		case errors.Is(err, ErrMissingPathParameter):
		case errors.Is(err, ErrInvalidPathParameter):
		case errors.Is(err, ErrPathParameterTooLong):
		default:
			t.Fatalf("unexpected parameter error: %v", err)
		}
		if !strings.HasPrefix(err.Error(), "id: path parameter is ") {
			t.Fatalf("parameter error contains unexpected detail: %v", err)
		}
	})
}
