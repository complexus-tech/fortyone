package teamshttp

import (
	"net/url"
	"strings"
	"testing"
)

func TestParseListTeamsQueryPreservesFiltersAndPaginationContract(t *testing.T) {
	t.Parallel()

	filter, params, err := parseListTeamsQuery(url.Values{
		"search":     {" Product "},
		"joinedOnly": {"true"},
		"page":       {"2"},
	})
	if err != nil {
		t.Fatalf("parseListTeamsQuery() error = %v", err)
	}
	if filter.Search != "Product" || !filter.JoinedOnly {
		t.Fatalf("filter = %#v", filter)
	}
	if params == nil || params.Page != 2 || params.PageSize != 15 || params.Offset() != 15 {
		t.Fatalf("pagination = %#v", params)
	}

	_, capped, err := parseListTeamsQuery(url.Values{"pageSize": {"500"}})
	if err != nil {
		t.Fatalf("parse capped page size: %v", err)
	}
	if capped == nil || capped.PageSize != 100 {
		t.Fatalf("capped pagination = %#v", capped)
	}

	_, unpaged, err := parseListTeamsQuery(url.Values{"search": {"platform"}})
	if err != nil || unpaged != nil {
		t.Fatalf("unpaged query = %#v, %v", unpaged, err)
	}
}

func TestParseListTeamsQueryRejectsUnsafeValuesWithoutEchoingThem(t *testing.T) {
	t.Parallel()

	for name, values := range map[string]url.Values{
		"repeated search":    {"search": {"first-sensitive", "second-sensitive"}},
		"oversized search":   {"search": {strings.Repeat("sensitive", 500)}},
		"invalid search":     {"search": {string([]byte{0xff})}},
		"nul search":         {"search": {"sensitive\x00search"}},
		"ambiguous boolean":  {"joinedOnly": {"TRUE"}},
		"repeated boolean":   {"joinedOnly": {"true", "false"}},
		"malformed page":     {"page": {"sensitive-page"}},
		"negative page size": {"pageSize": {"-1"}},
		"overflowing offset": {"page": {"2147483649"}, "pageSize": {"1"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, _, err := parseListTeamsQuery(values)
			if err == nil {
				t.Fatal("parseListTeamsQuery() error = nil")
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
