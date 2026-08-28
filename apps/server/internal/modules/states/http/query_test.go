package stateshttp

import (
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestParseListTeamIDUsesOneTypedOptionalValue(t *testing.T) {
	t.Parallel()

	want := uuid.New()
	got, err := parseListTeamID(url.Values{"teamId": {want.String()}})
	if err != nil || got == nil || *got != want {
		t.Fatalf("parseListTeamID() = %v, %v; want %s", got, err, want)
	}
	if got, err := parseListTeamID(url.Values{}); err != nil || got != nil {
		t.Fatalf("missing team ID = %v, %v", got, err)
	}
	if got, err := parseListTeamID(url.Values{"teamId": {""}}); err != nil || got != nil {
		t.Fatalf("blank team ID = %v, %v", got, err)
	}
}

func TestParseListTeamIDRejectsAmbiguousAndMalformedValuesSafely(t *testing.T) {
	t.Parallel()

	for name, values := range map[string]url.Values{
		"repeated":  {"teamId": {uuid.NewString(), uuid.NewString()}},
		"malformed": {"teamId": {"sensitive-team-id"}},
		"zero":      {"teamId": {uuid.Nil.String()}},
		"oversized": {"teamId": {strings.Repeat("sensitive", 500)}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := parseListTeamID(values)
			if err == nil {
				t.Fatal("parseListTeamID() error = nil")
			}
			for _, supplied := range values {
				for _, value := range supplied {
					if value != "" && strings.Contains(err.Error(), value) {
						t.Fatalf("error %q exposes query value", err)
					}
				}
			}
		})
	}
}
