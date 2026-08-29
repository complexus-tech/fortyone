package workspaceshttp

import (
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/web"
)

func TestParseSlugAvailabilityQueryIsBoundedAndUnambiguous(t *testing.T) {
	t.Parallel()

	got, err := parseSlugAvailabilityQuery(url.Values{"slug": {"roadmap-team"}})
	if err != nil || got != "roadmap-team" {
		t.Fatalf("parseSlugAvailabilityQuery() = %q, %v", got, err)
	}

	for name, test := range map[string]struct {
		values url.Values
		cause  error
	}{
		"missing":   {values: url.Values{}},
		"short":     {values: url.Values{"slug": {"ab"}}},
		"uppercase": {values: url.Values{"slug": {"Sensitive-Slug"}}},
		"nul":       {values: url.Values{"slug": {"abc\x00def"}}, cause: web.ErrInvalidQueryParameter},
		"oversized": {
			values: url.Values{"slug": {strings.Repeat("x", maximumWorkspaceSlugLength+1)}}, cause: web.ErrQueryParameterTooLong,
		},
		"repeated": {
			values: url.Values{"slug": {"first", "second"}}, cause: web.ErrRepeatedQueryParameter,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := parseSlugAvailabilityQuery(test.values)
			if err == nil {
				t.Fatal("expected invalid slug query")
			}
			if test.cause != nil && !errors.Is(err, test.cause) {
				t.Fatalf("error = %v, want cause %v", err, test.cause)
			}
			if strings.Contains(err.Error(), "Sensitive-Slug") || strings.Contains(err.Error(), "abc\x00def") {
				t.Fatalf("error exposes supplied value: %v", err)
			}
		})
	}
}
