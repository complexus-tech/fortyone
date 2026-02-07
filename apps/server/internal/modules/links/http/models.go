package linkshttp

import (
	"time"

	links "github.com/complexus-tech/projects-api/internal/modules/links/service"
	"github.com/google/uuid"
)

type Link struct {
	LinkID    uuid.UUID `json:"id"`
	Title     *string   `json:"title"`
	URL       string    `json:"url"`
	StoryID   uuid.UUID `json:"storyId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type NewLink struct {
	Title   *string   `json:"title"`
	URL     string    `json:"url"`
	StoryID uuid.UUID `json:"storyId" validate:"required"`
}

type UpdateLink struct {
	Title *string `json:"title"`
	URL   *string `json:"url"`
}

func toLink(cl links.CoreLink) Link {
	return Link{
		LinkID:    cl.LinkID,
		Title:     cl.Title,
		URL:       cl.URL,
		StoryID:   cl.StoryID,
		CreatedAt: cl.CreatedAt,
		UpdatedAt: cl.UpdatedAt,
	}
}

func ToLinks(cls []links.CoreLink) []Link {
	links := make([]Link, len(cls))
	for i, cl := range cls {
		links[i] = toLink(cl)
	}
	return links
}

func toCoreNewLink(nl NewLink) links.CoreNewLink {
	return links.CoreNewLink{
		Title:   nl.Title,
		URL:     nl.URL,
		StoryID: nl.StoryID,
	}
}

func toCoreUpdateLink(ul UpdateLink) links.CoreUpdateLink {
	return links.CoreUpdateLink{
		Title: ul.Title,
		URL:   ul.URL,
	}
}
