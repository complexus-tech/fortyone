package storieshttp

import (
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestToAppStoryLinksPreservesWireFields(t *testing.T) {
	t.Parallel()

	linkID := uuid.New()
	storyID := uuid.New()
	title := "Design"
	createdAt := time.Date(2026, time.August, 28, 10, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)

	actual := toAppStoryLinks([]storydomain.StoryLink{{
		ID: linkID, Title: &title, URL: "https://example.com/design",
		StoryID: storyID, CreatedAt: createdAt, UpdatedAt: updatedAt,
	}})

	require.Equal(t, []AppStoryLink{{
		ID: linkID, Title: &title, URL: "https://example.com/design",
		StoryID: storyID, CreatedAt: createdAt, UpdatedAt: updatedAt,
	}}, actual)
}
