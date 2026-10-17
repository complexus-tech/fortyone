package chatservice

import (
	"context"
	"strings"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type resourceRepository struct {
	Repository
	query domain.SearchQuery
}

func (r *resourceRepository) Resources(_ context.Context, query domain.SearchQuery) (domain.ResourcePage, error) {
	r.query = query
	return domain.ResourcePage{}, nil
}

func TestResourcesNormalizeFiltersAndSort(t *testing.T) {
	for _, test := range []struct {
		name, kind, sort, expectedKind, expectedSort string
	}{
		{name: "defaults", expectedKind: "all", expectedSort: "newest"},
		{name: "files alias", kind: "files", expectedKind: "file", expectedSort: "newest"},
		{name: "links alias", kind: "links", expectedKind: "link", expectedSort: "newest"},
		{name: "oldest media", kind: "media", sort: "oldest", expectedKind: "media", expectedSort: "oldest"},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := &resourceRepository{}
			_, err := New(repo).Resources(t.Context(), domain.SearchQuery{
				Scope: testScope(t), ConversationID: uuid.New(), Kind: test.kind, Sort: test.sort, Query: "  design  ",
			})
			require.NoError(t, err)
			require.Equal(t, test.expectedKind, repo.query.Kind)
			require.Equal(t, test.expectedSort, repo.query.Sort)
			require.Equal(t, "design", repo.query.Query)
			require.Equal(t, 50, repo.query.Limit)
		})
	}
}

func TestResourcesRejectUnsupportedFiltersAndSort(t *testing.T) {
	for _, test := range []struct{ name, kind, sort, query string }{
		{name: "unknown filter", kind: "audio"},
		{name: "unknown sort", sort: "name"},
		{name: "long search", query: strings.Repeat("é", 201)},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := &resourceRepository{}
			_, err := New(repo).Resources(t.Context(), domain.SearchQuery{
				Scope: testScope(t), ConversationID: uuid.New(), Kind: test.kind, Sort: test.sort, Query: test.query,
			})
			require.ErrorIs(t, err, domain.ErrInvalid)
			require.Zero(t, repo.query)
		})
	}
}
