package usershttp

import (
	"net/url"
	"strings"
	"testing"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
)

func TestParseOnboardingTourProgressQueryNormalizesVersionedScope(t *testing.T) {
	t.Parallel()

	query, err := parseOnboardingTourProgressQuery(url.Values{
		"tourKey":     {" workspace-getting-started "},
		"tourVersion": {" v1 "},
	})
	if err != nil {
		t.Fatalf("parse onboarding progress query: %v", err)
	}
	if query.TourKey != "workspace-getting-started" || query.TourVersion != "v1" {
		t.Fatalf("parsed scope = %#v", query)
	}
}

func TestParseOnboardingTourProgressQueryRejectsUnsafeScope(t *testing.T) {
	t.Parallel()

	for name, values := range map[string]url.Values{
		"missing key":       {"tourVersion": {"v1"}},
		"repeated key":      {"tourKey": {"one", "two"}, "tourVersion": {"v1"}},
		"nul version":       {"tourKey": {"workspace"}, "tourVersion": {"v1\x00"}},
		"oversized version": {"tourKey": {"workspace"}, "tourVersion": {strings.Repeat("v", users.MaximumOnboardingTourVersionRunes+1)}},
	} {
		values := values
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := parseOnboardingTourProgressQuery(values); err == nil {
				t.Fatal("parseOnboardingTourProgressQuery() error = nil")
			}
		})
	}
}
