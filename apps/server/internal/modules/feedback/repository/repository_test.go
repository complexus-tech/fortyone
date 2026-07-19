package feedbackrepository

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestItemSelectQuerySupportsNamedPaginationParameters(t *testing.T) {
	query := fmt.Sprintf(
		"%s WHERE fi.portal_id = :portal_id ORDER BY fi.created_at DESC LIMIT :limit OFFSET :offset",
		itemSelectQuery(),
	)

	boundQuery, args, err := sqlx.BindNamed(sqlx.DOLLAR, query, map[string]any{
		"portal_id": uuid.New(),
		"limit":     21,
		"offset":    0,
	})

	require.NoError(t, err)
	require.Len(t, args, 3)
	require.False(t, strings.Contains(boundQuery, ":"), "bound query contains an uncompiled named parameter")
}
