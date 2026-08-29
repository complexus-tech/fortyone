package domain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestPaginatedListQueryNormalize(t *testing.T) {
	t.Parallel()

	workspaceID, actorID, teamID := uuid.New(), uuid.New(), uuid.New()
	query, err := (PaginatedListQuery{
		Access: AccessScope{WorkspaceID: workspaceID, ActorID: actorID, AllTeams: true},
		Filters: Filters{
			WorkspaceID: uuid.New(), CurrentUserID: uuid.New(), TeamIDs: []uuid.UUID{teamID, teamID},
			MeasurementTypes: []string{"number"}, Page: -1, PageSize: MaximumPageSize + 1,
			OrderBy: "unsafe_column", OrderDirection: "sideways",
		},
	}).Normalize()
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if query.Filters.WorkspaceID != workspaceID || query.Filters.CurrentUserID != actorID {
		t.Fatalf("tenant/actor filters = %s/%s", query.Filters.WorkspaceID, query.Filters.CurrentUserID)
	}
	if query.Filters.Page != 1 || query.Filters.PageSize != MaximumPageSize {
		t.Fatalf("pagination = %d/%d", query.Filters.Page, query.Filters.PageSize)
	}
	if query.Filters.OrderBy != "created_at" || query.Filters.OrderDirection != "desc" || query.SortKey() != "created_at_desc" {
		t.Fatalf("sort = %q/%q", query.Filters.OrderBy, query.Filters.OrderDirection)
	}
	if len(query.Filters.TeamIDs) != 1 || query.Filters.TeamIDs[0] != teamID {
		t.Fatalf("team IDs = %#v", query.Filters.TeamIDs)
	}
}

func TestPaginatedListQueryRejectsInvalidTypedFilters(t *testing.T) {
	t.Parallel()

	access := AccessScope{WorkspaceID: uuid.New(), ActorID: uuid.New(), AllTeams: true}
	for _, filters := range []Filters{
		{TeamIDs: []uuid.UUID{uuid.Nil}},
		{MeasurementTypes: []string{"currency"}},
	} {
		if _, err := (PaginatedListQuery{Access: access, Filters: filters}).Normalize(); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Normalize(%#v) error = %v, want ErrInvalid", filters, err)
		}
	}
}
