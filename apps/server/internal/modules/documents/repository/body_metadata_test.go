package documentsrepository

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBodyMetadataUsesStructuredAtomsAndDeduplicates(t *testing.T) {
	user, story := uuid.New(), uuid.New()
	mention := fmt.Sprintf(`<span data-type="mention" data-mention-kind="user" data-id="%s">@User</span>`, user)
	reference := fmt.Sprintf(`<span data-type="document-reference" data-entity-type="story" data-id="%s">Story</span>`, story)
	refs, mentions, err := documentBodyMetadata("<p>" + mention + mention + reference + reference + "</p><pre>" + mention + "</pre><span data-type=mention data-id=invalid>bad</span>")
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{user}, mentions)
	require.Equal(t, []bodyReference{{EntityType: "story", ID: story}}, refs)
	refs, mentions, err = documentBodyMetadata("<p>@User and plain link text</p>")
	require.NoError(t, err)
	require.Empty(t, refs)
	require.Empty(t, mentions)
}
