package usershttp

import (
	"net/url"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/microsoft"
)

func TestMicrosoftFailureURLPreservesCallbackAndReportsCancellation(t *testing.T) {
	rawURL := microsoftFailureURL(
		"https://app.fortyone.app/auth-callback?mobileApp=true",
		"access_denied",
	)
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse failure URL: %v", err)
	}
	if parsed.Path != "/" {
		t.Errorf("path = %q, want root login path", parsed.Path)
	}
	if parsed.Query().Get("mobileApp") != "true" {
		t.Errorf("mobile callback query was not preserved: %q", parsed.RawQuery)
	}
	if parsed.Query().Get("error") != "Microsoft sign-in was cancelled." {
		t.Errorf("error message = %q", parsed.Query().Get("error"))
	}
}

func TestBuildMicrosoftFullNameFallbacks(t *testing.T) {
	tests := []struct {
		name     string
		identity microsoft.Identity
		email    string
		want     string
	}{
		{name: "display name", identity: microsoft.Identity{FullName: "Ada Lovelace"}, want: "Ada Lovelace"},
		{name: "given and family name", identity: microsoft.Identity{FirstName: "Ada", LastName: "Lovelace"}, want: "Ada Lovelace"},
		{name: "preferred username", identity: microsoft.Identity{PreferredUsername: "ada.l"}, want: "ada.l"},
		{name: "email local part", email: "ada@example.com", want: "ada"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildMicrosoftFullName(tt.identity, tt.email); got != tt.want {
				t.Errorf("buildMicrosoftFullName() = %q, want %q", got, tt.want)
			}
		})
	}
}
