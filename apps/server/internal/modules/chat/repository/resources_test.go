package chatrepository

import (
	"strings"
	"testing"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestResourceCursorSupportsMaximumUnicodeSearch(t *testing.T) {
	query := domain.SearchQuery{ConversationID: uuid.New(), Kind: "media", Sort: "oldest", Query: strings.Repeat("🌍", 200)}
	id, createdAt := uuid.New(), time.Now().UTC()
	encoded := encodeCursor(cursor{ConversationID: query.ConversationID, Time: &createdAt, ID: &id, ResourceScope: resourceCursorScope(query)})
	decoded, err := decodeCursor(encoded)
	require.NoError(t, err)
	require.Equal(t, resourceCursorScope(query), decoded.ResourceScope)
	require.Less(t, len(encoded), 1024)
}
