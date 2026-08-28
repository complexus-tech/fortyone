package usershttp

import (
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestParseUserListQueryProducesTypedBoundedFilter(t *testing.T) {
	t.Parallel()

	teamID := uuid.New()
	query, err := parseUserListQuery(url.Values{
		"teamId":   {teamID.String()},
		"search":   {" Product "},
		"page":     {"2"},
		"pageSize": {"25"},
	})
	if err != nil {
		t.Fatalf("parseUserListQuery() error = %v", err)
	}
	if query.Filter.TeamID == nil || *query.Filter.TeamID != teamID || query.Filter.Search != "Product" {
		t.Fatalf("filter = %#v", query.Filter)
	}
	if query.Page == nil || query.Page.Page != 2 || query.Page.PageSize != 25 || query.Page.Offset() != 25 {
		t.Fatalf("pagination = %#v", query.Page)
	}

	capped, err := parseUserListQuery(url.Values{"pageSize": {"500"}})
	if err != nil || capped.Page == nil || capped.Page.PageSize != 100 {
		t.Fatalf("capped query = %#v, %v", capped, err)
	}

	unpaged, err := parseUserListQuery(url.Values{"search": {"product"}})
	if err != nil || unpaged.Page != nil {
		t.Fatalf("unpaged query = %#v, %v", unpaged, err)
	}
}

func TestParseUserListQueryRejectsUnsafeValuesWithoutEchoingThem(t *testing.T) {
	t.Parallel()

	for name, values := range map[string]url.Values{
		"repeated team":      {"teamId": {uuid.NewString(), uuid.NewString()}},
		"malformed team":     {"teamId": {"sensitive-team"}},
		"zero team":          {"teamId": {uuid.Nil.String()}},
		"repeated search":    {"search": {"first-sensitive", "second-sensitive"}},
		"oversized search":   {"search": {strings.Repeat("sensitive", 500)}},
		"invalid search":     {"search": {string([]byte{0xff})}},
		"nul search":         {"search": {"sensitive\x00search"}},
		"malformed page":     {"page": {"sensitive-page"}},
		"repeated page size": {"pageSize": {"10", "20"}},
		"negative page size": {"pageSize": {"-1"}},
		"overflowing offset": {"page": {"2147483649"}, "pageSize": {"1"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := parseUserListQuery(values)
			if err == nil {
				t.Fatal("parseUserListQuery() error = nil")
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
