package linksrepository

import (
	"time"

	linksdomain "github.com/complexus-tech/projects-api/internal/modules/links/domain"
	linksql "github.com/complexus-tech/projects-api/internal/modules/links/repository/sqlc"
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

func toCoreLink(link DbLink) linksdomain.Link {
	return linksdomain.Link{
		LinkID:    link.ID,
		Title:     link.Title,
		URL:       link.URL,
		StoryID:   link.StoryID,
		CreatedAt: link.CreatedAt,
		UpdatedAt: link.UpdatedAt,
	}
}

func ToCoreLinks(lnks []DbLink) []linksdomain.Link {
	coreLinks := make([]linksdomain.Link, len(lnks))
	for i, link := range lnks {
		coreLinks[i] = toCoreLink(link)
	}
	return coreLinks
}

func fromCreateLinkRow(link linksql.CreateLinkForWorkspaceRow) linksdomain.Link {
	return linksdomain.Link{
		LinkID:    link.LinkID,
		Title:     link.Title,
		URL:       link.URL,
		StoryID:   link.StoryID,
		CreatedAt: link.CreatedAt,
		UpdatedAt: link.UpdatedAt,
	}
}
