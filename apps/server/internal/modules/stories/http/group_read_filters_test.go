package storieshttp

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestParseStoryQueryRejectsMalformedFilters(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		wantInError string
	}{
		{name: "uuid array", query: "statusIds=not-a-uuid", wantInError: "statusIds"},
		{name: "mixed uuid array", query: "statusIds=" + uuid.NewString() + ",invalid", wantInError: "statusIds"},
		{name: "empty array member", query: "priorities=High,,Low", wantInError: "priorities"},
		{name: "estimate", query: "estimateValues=not-an-int", wantInError: "estimateValues"},
		{name: "scalar uuid", query: "objectiveId=invalid", wantInError: "objectiveId"},
		{name: "boolean", query: "includeDeleted=sometimes", wantInError: "includeDeleted"},
		{name: "date", query: "deadlineBefore=tomorrow", wantInError: "deadlineBefore"},
		{name: "integer", query: "storiesPerGroup=many", wantInError: "storiesPerGroup"},
		{name: "duplicate scalar", query: "parentId=" + uuid.NewString() + "&parentId=" + uuid.NewString(), wantInError: "parentId"},
		{name: "duplicate enum", query: "groupBy=status&groupBy=team", wantInError: "groupBy"},
		{name: "blank search", query: "titleContains=%20%20", wantInError: "titleContains"},
		{name: "zero uuid", query: "teamIds=" + uuid.Nil.String(), wantInError: "teamIds"},
		{name: "non canonical boolean", query: "includeDeleted=TRUE", wantInError: "includeDeleted"},
		{name: "blank boolean", query: "includeDeleted=", wantInError: "includeDeleted"},
		{name: "page below range", query: "page=0", wantInError: "page"},
		{name: "page above range", query: "page=1001", wantInError: "page"},
		{name: "page size above range", query: "pageSize=101", wantInError: "pageSize"},
		{name: "stories per group above range", query: "storiesPerGroup=101", wantInError: "storiesPerGroup"},
		{name: "unsupported priority", query: "priorities=Critical", wantInError: "priorities"},
		{name: "unsupported category", query: "categories=unknown", wantInError: "categories"},
		{name: "oversized scalar", query: "titleContains=" + strings.Repeat("x", maxStoryScalarQueryBytes+1), wantInError: "titleContains"},
		{name: "oversized list item", query: "categories=" + strings.Repeat("x", maxStoryListItemBytes+1), wantInError: "categories"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/stories?"+tt.query, nil)

			_, err := parseStoryQuery(request)
			if err == nil {
				t.Fatal("expected malformed filter to be rejected")
			}
			if !strings.Contains(err.Error(), tt.wantInError) {
				t.Fatalf("error %q does not identify %q", err, tt.wantInError)
			}
		})
	}
}

func TestParseStoryQueryUsesBoundedDefaultsAndDeduplicatesLists(t *testing.T) {
	statusOne := uuid.New()
	statusTwo := uuid.New()
	request := httptest.NewRequest(
		"GET",
		"/stories?statusIds="+statusOne.String()+","+statusOne.String()+"&statusIds="+statusTwo.String()+"&includeArchived=false",
		nil,
	)

	query, err := parseStoryQuery(request)
	if err != nil {
		t.Fatalf("parse story query: %v", err)
	}
	if query.GroupBy != "status" || query.OrderBy != "created" || query.OrderDirection != "desc" {
		t.Fatalf("unexpected defaults: %#v", query)
	}
	if query.Page != 1 || query.PageSize != 0 || query.StoriesPerGroup != 0 {
		t.Fatalf("unexpected pagination defaults: %#v", query)
	}
	if len(query.Filters.StatusIDs) != 2 || query.Filters.StatusIDs[0] != statusOne || query.Filters.StatusIDs[1] != statusTwo {
		t.Fatalf("deduplicated status IDs = %#v", query.Filters.StatusIDs)
	}
	if query.Filters.IncludeArchived == nil || *query.Filters.IncludeArchived {
		t.Fatalf("include archived = %v, want false", query.Filters.IncludeArchived)
	}
}

func TestParseStoryQueryCapsListCardinalityWithoutLeakingValues(t *testing.T) {
	values := make([]string, maxStoryListItems+1)
	for index := range values {
		values[index] = uuid.NewString()
	}
	sensitive := strings.Join(values, ",")
	request := httptest.NewRequest("GET", "/stories?statusIds="+sensitive, nil)

	_, err := parseStoryQuery(request)
	if err == nil {
		t.Fatal("expected excessive list cardinality to be rejected")
	}
	if strings.Contains(err.Error(), sensitive) {
		t.Fatalf("error exposes supplied values: %v", err)
	}
}

func TestParseStoryPaginationRejectsMalformedValues(t *testing.T) {
	for name, query := range map[string]string{
		"repeated page":       "page=1&page=2",
		"blank page":          "page=",
		"overflow page":       "page=999999999999999999999999999",
		"negative page size":  "pageSize=-1",
		"oversized page size": "pageSize=101",
	} {
		t.Run(name, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/stories?"+query, nil)
			if _, _, err := parseStoryPagination(request, 1, 20); err == nil {
				t.Fatal("expected invalid pagination to be rejected")
			}
		})
	}
}

func TestParseStoryQueryMapsKeyResultAndRepeatedLists(t *testing.T) {
	keyResultID := uuid.New()
	statusOne := uuid.New()
	statusTwo := uuid.New()
	request := httptest.NewRequest(
		"GET",
		"/stories?keyResultId="+keyResultID.String()+"&statusIds="+statusOne.String()+"&statusIds="+statusTwo.String()+"&orderBy=completed",
		nil,
	)

	query, err := parseStoryQuery(request)
	if err != nil {
		t.Fatalf("parse story query: %v", err)
	}
	if query.Filters.KeyResult == nil || *query.Filters.KeyResult != keyResultID {
		t.Fatalf("key result = %v, want %s", query.Filters.KeyResult, keyResultID)
	}
	if len(query.Filters.StatusIDs) != 2 || query.Filters.StatusIDs[0] != statusOne || query.Filters.StatusIDs[1] != statusTwo {
		t.Fatalf("status IDs = %v, want [%s %s]", query.Filters.StatusIDs, statusOne, statusTwo)
	}
	core := toCoreStoryQuery(query)
	if core.Filters.KeyResult == nil || *core.Filters.KeyResult != keyResultID {
		t.Fatalf("core key result = %v, want %s", core.Filters.KeyResult, keyResultID)
	}
}
