//go:build integration

package chatrepository

import (
	"cmp"
	"slices"
	"testing"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestChatResourcesFilterAndPageInBothDirections(t *testing.T) {
	f := newFixture(t)
	baseTime := time.Date(2026, time.October, 1, 12, 0, 0, 0, time.UTC)
	files, media := []uuid.UUID{}, []uuid.UUID{}
	for index, file := range []struct{ name, mediaType string }{
		{"design.pdf", "application/pdf"},
		{"design.png", "image/png"},
		{"design.jpg", "image/jpeg"},
		{"design.mp4", "video/mp4"},
		{"design.mp3", "audio/mpeg"},
	} {
		id := addResourceAttachment(t, f, file.name, file.mediaType, baseTime.Add(time.Duration(index/2)*time.Hour))
		files = append(files, id)
		if index >= 1 && index <= 3 {
			media = append(media, id)
		}
	}
	firstLink := f.create(t, "https://fortyone.app/design", nil)
	f.create(t, "https://fortyone.app/guide", &firstLink.ID)
	query := domain.SearchQuery{Scope: f.scope, ConversationID: f.room, Kind: "all"}
	all, err := f.service.Resources(f.ctx, query)
	require.NoError(t, err)
	require.Len(t, all.Items, 7)

	for _, filter := range []struct {
		kind string
		ids  []uuid.UUID
	}{
		{kind: "all", ids: resourceIDs(all.Items)},
		{kind: "files", ids: files},
		{kind: "media", ids: media},
		{kind: "links", ids: resourceIDs(slices.DeleteFunc(slices.Clone(all.Items), func(item domain.Resource) bool { return item.Kind != "link" }))},
	} {
		for _, order := range []string{"newest", "oldest"} {
			t.Run(filter.kind+"/"+order, func(t *testing.T) {
				expected := slices.DeleteFunc(slices.Clone(all.Items), func(item domain.Resource) bool { return !slices.Contains(filter.ids, item.ID) })
				slices.SortFunc(expected, func(a, b domain.Resource) int {
					comparison := a.CreatedAt.Compare(b.CreatedAt)
					if comparison == 0 {
						comparison = cmp.Compare(a.ID.String(), b.ID.String())
					}
					if order == "newest" {
						return -comparison
					}
					return comparison
				})
				query := domain.SearchQuery{Scope: f.scope, ConversationID: f.room, Kind: filter.kind, Sort: order, Limit: 2}
				var received []uuid.UUID
				for pageNumber := 0; ; pageNumber++ {
					require.Less(t, pageNumber, 5, "cursor must advance")
					page, err := f.service.Resources(f.ctx, query)
					require.NoError(t, err)
					require.LessOrEqual(t, len(page.Items), 2)
					received = append(received, resourceIDs(page.Items)...)
					if page.NextCursor == nil {
						break
					}
					query.Cursor = *page.NextCursor
				}
				require.Equal(t, resourceIDs(expected), received)
			})
		}
	}
	query.Kind, query.Query = "media", "PNG"
	matching, err := f.service.Resources(f.ctx, query)
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{media[0]}, resourceIDs(matching.Items))
	_, err = f.service.Resources(f.ctx, domain.SearchQuery{Scope: humanScope(t, f.workspace, f.outsider), ConversationID: f.room, Kind: "media", Sort: "oldest"})
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestChatResourceCursorsAreScopedToFiltersAndOrder(t *testing.T) {
	f := newFixture(t)
	for _, name := range []string{"design-a.png", "design-b.png", "design-c.pdf"} {
		addResourceAttachment(t, f, name, "image/png", time.Now())
	}
	query := domain.SearchQuery{Scope: f.scope, ConversationID: f.room, Kind: "files", Query: "design", Limit: 1}
	first, err := f.service.Resources(f.ctx, query)
	require.NoError(t, err)
	require.NotNil(t, first.NextCursor)
	query.Cursor = *first.NextCursor
	for _, test := range []struct {
		name   string
		change func(*domain.SearchQuery)
	}{
		{"filter", func(q *domain.SearchQuery) { q.Kind = "media" }},
		{"order", func(q *domain.SearchQuery) { q.Sort = "oldest" }},
		{"search", func(q *domain.SearchQuery) { q.Query = "other" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			changed := query
			test.change(&changed)
			_, err := f.service.Resources(f.ctx, changed)
			require.ErrorIs(t, err, domain.ErrInvalid)
		})
	}
	query.Kind, query.Sort, query.Query = "file", "newest", "  design  "
	second, err := f.service.Resources(f.ctx, query)
	require.NoError(t, err, "equivalent normalized filters must accept their cursor")
	require.Len(t, second.Items, 1)
	require.NotEqual(t, first.Items[0].ID, second.Items[0].ID)
	dm, err := f.service.CreateDirect(f.ctx, f.scope, f.peer)
	require.NoError(t, err)
	query.ConversationID = dm.ID
	_, err = f.service.Resources(f.ctx, query)
	require.ErrorIs(t, err, domain.ErrInvalid)
}

func addResourceAttachment(t *testing.T, f fixture, filename, mediaType string, createdAt time.Time) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := f.pool.Exec(f.ctx, `INSERT INTO attachments(attachment_id,workspace_id,uploaded_by,filename,size,mime_type,blob_name,scan_status) VALUES($1,$2,$3,$4,128,$5,$6,'clean')`, id, f.workspace, f.author, filename, mediaType, id.String())
	require.NoError(t, err)
	require.NoError(t, f.service.LinkPendingAttachment(f.ctx, domain.AttachmentScope{ActorScope: f.scope, ConversationID: f.room}, id))
	message, err := f.service.CreateMessage(f.ctx, domain.CreateMessageCommand{Scope: f.scope, ConversationID: f.room, CreationID: uuid.New(), AttachmentIDs: []uuid.UUID{id}})
	require.NoError(t, err)
	_, err = f.pool.Exec(f.ctx, `UPDATE chat_conversation_messages SET created_at=$2 WHERE id=$1`, message.ID, createdAt)
	require.NoError(t, err)
	return id
}

func resourceIDs(resources []domain.Resource) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(resources))
	for _, resource := range resources {
		ids = append(ids, resource.ID)
	}
	return ids
}
