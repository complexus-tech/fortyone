package storieshttp

import (
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/google/uuid"
)

// AppStoryLink is the story endpoint's stable representation of an attached
// external link. The adapter owns this wire shape instead of importing another
// module's HTTP package.
type AppStoryLink struct {
	ID        uuid.UUID `json:"id"`
	Title     *string   `json:"title"`
	URL       string    `json:"url"`
	StoryID   uuid.UUID `json:"storyId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func toAppStoryLinks(coreLinks []storydomain.StoryLink) []AppStoryLink {
	result := make([]AppStoryLink, len(coreLinks))
	for index, link := range coreLinks {
		result[index] = AppStoryLink{
			ID:        link.ID,
			Title:     link.Title,
			URL:       link.URL,
			StoryID:   link.StoryID,
			CreatedAt: link.CreatedAt,
			UpdatedAt: link.UpdatedAt,
		}
	}
	return result
}
