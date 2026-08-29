package storiesrepository

import (
	"errors"
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/google/uuid"
)

func TestNormalizeStoryFiltersRejectsUnsafeInputs(t *testing.T) {
	trueValue := true
	now := time.Now().UTC()
	later := now.Add(time.Hour)
	tooManyIDs := make([]uuid.UUID, maxStoryFilterValues+1)
	for index := range tooManyIDs {
		tooManyIDs[index] = uuid.New()
	}

	tests := []struct {
		name    string
		filters storydomain.StoryFilters
	}{
		{name: "unsupported epic", filters: storydomain.StoryFilters{Epic: uuidTestPointer(uuid.New())}},
		{name: "zero id", filters: storydomain.StoryFilters{StatusIDs: []uuid.UUID{uuid.Nil}}},
		{name: "too many values", filters: storydomain.StoryFilters{TeamIDs: tooManyIDs}},
		{name: "invalid priority", filters: storydomain.StoryFilters{Priorities: []string{"Critical"}}},
		{name: "invalid category", filters: storydomain.StoryFilters{Categories: []string{"open"}}},
		{name: "invalid estimate", filters: storydomain.StoryFilters{EstimateValues: []int16{13}}},
		{name: "contradictory assignee", filters: storydomain.StoryFilters{HasAssignee: &trueValue, HasNoAssignee: &trueValue}},
		{name: "contradictory completion", filters: storydomain.StoryFilters{IsCompleted: &trueValue, IsNotCompleted: &trueValue}},
		{name: "reversed range", filters: storydomain.StoryFilters{CreatedAfter: &later, CreatedBefore: &now}},
		{name: "negative limit", filters: storydomain.StoryFilters{Limit: -1}},
		{name: "oversized offset", filters: storydomain.StoryFilters{Offset: maxGroupedPage*maxGroupedPageSize + 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := normalizeStoryFilters(tt.filters)
			if !errors.Is(err, storydomain.ErrInvalidReadQuery) {
				t.Fatalf("error = %v, want ErrInvalidReadQuery", err)
			}
		})
	}
}

func TestValidateStoryQueryAndGroupPageBounds(t *testing.T) {
	valid := storydomain.StoryQuery{
		GroupBy:         storydomain.StoryGroupStatus,
		OrderBy:         storydomain.StoryOrderCreated,
		OrderDirection:  storydomain.SortDescending,
		StoriesPerGroup: 15,
		Page:            1,
		PageSize:        20,
	}
	if _, err := validateInitialStoryGroupQuery(valid); err != nil {
		t.Fatalf("valid query rejected: %v", err)
	}
	pageOnly := valid
	pageOnly.StoriesPerGroup = 0
	if pageOnly, err := validateStoryQuery(pageOnly); err != nil {
		t.Fatalf("valid group page query rejected: %v", err)
	} else if _, _, err := validateGroupPage(pageOnly, uuid.NewString()); err != nil {
		t.Fatalf("valid group page bounds rejected: %v", err)
	}

	tests := []struct {
		name  string
		query storydomain.StoryQuery
		key   string
	}{
		{name: "zero page", query: storydomain.StoryQuery{Page: 0, PageSize: 20, GroupBy: storydomain.StoryGroupStatus}, key: uuid.NewString()},
		{name: "oversized page", query: storydomain.StoryQuery{Page: maxGroupedPage + 1, PageSize: 20, GroupBy: storydomain.StoryGroupStatus}, key: uuid.NewString()},
		{name: "oversized page size", query: storydomain.StoryQuery{Page: 1, PageSize: maxGroupedPageSize + 1, GroupBy: storydomain.StoryGroupStatus}, key: uuid.NewString()},
		{name: "invalid status key", query: storydomain.StoryQuery{Page: 1, PageSize: 20, GroupBy: storydomain.StoryGroupStatus}, key: "not-a-uuid"},
		{name: "invalid none key", query: storydomain.StoryQuery{Page: 1, PageSize: 20, GroupBy: storydomain.StoryGroupNone}, key: "all"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := validateGroupPage(tt.query, tt.key)
			if !errors.Is(err, storydomain.ErrInvalidReadQuery) {
				t.Fatalf("error = %v, want ErrInvalidReadQuery", err)
			}
		})
	}
}

func TestCompatibilityReadValidationUsesDomainError(t *testing.T) {
	_, categoryErr := parseStoryCategory("unknown")
	_, _, pageErr := categoryPage(0, 20)
	scopeErr := validateReadScope(storydomain.ReadScope{})
	for name, err := range map[string]error{
		"category": categoryErr,
		"page":     pageErr,
		"scope":    scopeErr,
	} {
		if !errors.Is(err, storydomain.ErrInvalidReadQuery) {
			t.Fatalf("%s error = %v, want ErrInvalidReadQuery", name, err)
		}
	}
}

func TestGroupCatalogParamsRequestsOneLookAheadRow(t *testing.T) {
	params := groupCatalogParams(storydomain.ReadScope{
		ActorID:     uuid.New(),
		WorkspaceID: uuid.New(),
	}, storydomain.StoryQuery{GroupBy: storydomain.StoryGroupStatus})

	if params.CatalogLimit != int32(maxStoryGroupCatalog+1) {
		t.Fatalf("catalog limit = %d, want %d", params.CatalogLimit, maxStoryGroupCatalog+1)
	}
}

func uuidTestPointer(value uuid.UUID) *uuid.UUID {
	return &value
}
