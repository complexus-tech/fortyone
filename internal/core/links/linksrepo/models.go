package linksrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/links"
	"github.com/google/uuid"
)

type DbLink struct {
	ID        uuid.UUID `db:"link_id"`
	Title     *string   `db:"title"`
	URL       string    `db:"url"`
	StoryID   uuid.UUID `db:"story_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type dbNewLink struct {
	Title   *string   `db:"title"`
	URL     string    `db:"url"`
	StoryID uuid.UUID `db:"story_id"`
}

type dbUpdateLink struct {
	LinkID uuid.UUID `db:"link_id"`
	Title  *string   `db:"title"`
	URL    *string   `db:"url"`
}

func toCoreLink(link DbLink) links.CoreLink {
	return links.CoreLink{
		LinkID:    link.ID,
		Title:     link.Title,
		URL:       link.URL,
		StoryID:   link.StoryID,
		CreatedAt: link.CreatedAt,
		UpdatedAt: link.UpdatedAt,
	}
}

func ToCoreLinks(lnks []DbLink) []links.CoreLink {
	coreLinks := make([]links.CoreLink, len(lnks))
	for i, link := range lnks {
		coreLinks[i] = toCoreLink(link)
	}
	return coreLinks
}

func toDbNewLink(cnl links.CoreNewLink) dbNewLink {
	return dbNewLink{
		Title:   cnl.Title,
		URL:     cnl.URL,
		StoryID: cnl.StoryID,
	}
}

func toDbUpdateLink(cul links.CoreUpdateLink, linkID uuid.UUID) dbUpdateLink {
	return dbUpdateLink{
		LinkID: linkID,
		Title:  cul.Title,
		URL:    cul.URL,
	}
}
