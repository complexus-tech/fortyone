package documentsrepository

import (
	"strings"
	"testing"

	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestBuildDocumentListQueryAppliesLimitOnlyWhenRequested(t *testing.T) {
	t.Parallel()

	baseInput := documents.CoreListInput{
		WorkspaceID: uuid.New(),
		UserID:      uuid.New(),
	}

	query, params := buildDocumentListQuery(baseInput)
	require.NotContains(t, strings.ToUpper(query), "LIMIT")
	require.NotContains(t, params, "limit")

	limit := 8
	baseInput.Limit = &limit
	query, params = buildDocumentListQuery(baseInput)
	require.Contains(t, query, "LIMIT :limit")
	require.Equal(t, limit, params["limit"])
}
