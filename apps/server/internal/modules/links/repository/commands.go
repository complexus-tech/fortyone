package linksrepository

import (
	"context"
	"fmt"

	links "github.com/complexus-tech/projects-api/internal/modules/links/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (r *repo) CreateLink(ctx context.Context, cnl links.CoreNewLink) (links.CoreLink, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.links.CreateLink")
	defer span.End()

	span.SetAttributes(
		attribute.String("storyId", cnl.StoryID.String()),
		attribute.String("url", cnl.URL),
	)

	var link DbLink
	query := `
		INSERT INTO
			story_links (title, url, story_id)
		VALUES
			(:title,:url,:story_id)
		RETURNING
			link_id,
			title,
			url,
			story_id,
			created_at,
			updated_at
	`

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, "error preparing statement", err)
		return links.CoreLink{}, err
	}

	if err := stmt.GetContext(ctx, &link, toDbNewLink(cnl)); err != nil {
		r.log.Error(ctx, "error getting link", err)
		return links.CoreLink{}, err
	}

	return toCoreLink(link), nil
}

func (r *repo) UpdateLink(ctx context.Context, linkID uuid.UUID, cul links.CoreUpdateLink) error {
	ctx, span := web.AddSpan(ctx, "business.repository.links.UpdateLink")
	defer span.End()

	span.SetAttributes(attribute.String("linkId", linkID.String()))

	query := `
		UPDATE story_links
		SET
			title = COALESCE(:title, title),
			url = COALESCE(:url, url),
			updated_at = NOW()
		WHERE link_id = :link_id
	`

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, "error preparing statement", err)
		return err
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, toDbUpdateLink(cul, linkID)); err != nil {
		r.log.Error(ctx, "error updating link", err)
		return err
	}

	r.log.Info(ctx, "link updated successfully", "linkId", linkID)

	return nil
}

func (r *repo) DeleteLink(ctx context.Context, linkID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.links.DeleteLink")
	defer span.End()

	span.SetAttributes(attribute.String("linkId", linkID.String()))

	query := `
		DELETE FROM story_links
		WHERE link_id = :link_id
	`
	params := map[string]interface{}{
		"link_id": linkID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, "error preparing statement", err)
		return err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("Deleting link #%s", linkID), "linkId", linkID)
	if _, err := stmt.ExecContext(ctx, params); err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to delete link: %s", err), "linkId", linkID)
		return err
	}

	r.log.Info(ctx, fmt.Sprintf("Link #%s deleted successfully", linkID), "linkId", linkID)
	span.AddEvent("Link deleted.", trace.WithAttributes(attribute.String("link.id", linkID.String())))

	return nil
}
