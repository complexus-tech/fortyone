package linksrepository

import (
	"context"
	"errors"
	"fmt"

	linksdomain "github.com/complexus-tech/projects-api/internal/modules/links/domain"
	linksql "github.com/complexus-tech/projects-api/internal/modules/links/repository/sqlc"
	apptracing "github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (r *repo) CreateLink(ctx context.Context, actorID uuid.UUID, input linksdomain.CreateLink) (linksdomain.Link, error) {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.links.CreateLink")
	defer span.End()

	span.SetAttributes(
		attribute.String("storyId", input.StoryID.String()),
		attribute.String("workspaceId", input.WorkspaceID.String()),
		attribute.String("actorId", actorID.String()),
	)

	link, err := r.queries.CreateLinkForWorkspace(ctx, linksql.CreateLinkForWorkspaceParams{
		Title:       input.Title,
		URL:         input.URL,
		StoryID:     input.StoryID,
		WorkspaceID: input.WorkspaceID,
		ActorID:     actorID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return linksdomain.Link{}, linksdomain.ErrNotFound
		}
		r.log.Error(ctx, "error creating link", "error", err)
		return linksdomain.Link{}, fmt.Errorf("create link: %w", err)
	}

	return fromCreateLinkRow(link), nil
}

func (r *repo) UpdateLink(
	ctx context.Context,
	actorID uuid.UUID,
	linkID uuid.UUID,
	workspaceID uuid.UUID,
	input linksdomain.UpdateLink,
) error {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.links.UpdateLink")
	defer span.End()

	span.SetAttributes(
		attribute.String("linkId", linkID.String()),
		attribute.String("workspaceId", workspaceID.String()),
		attribute.String("actorId", actorID.String()),
	)

	rowsAffected, err := r.queries.UpdateLinkForWorkspace(ctx, linksql.UpdateLinkForWorkspaceParams{
		Title:       input.Title,
		URL:         input.URL,
		LinkID:      linkID,
		WorkspaceID: workspaceID,
		ActorID:     actorID,
	})
	if err != nil {
		r.log.Error(ctx, "error updating link", "error", err)
		return fmt.Errorf("update link: %w", err)
	}
	if rowsAffected == 0 {
		return linksdomain.ErrNotFound
	}

	r.log.Info(ctx, "link updated successfully", "linkId", linkID)
	return nil
}

func (r *repo) DeleteLink(ctx context.Context, actorID, linkID, workspaceID uuid.UUID) error {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.links.DeleteLink")
	defer span.End()

	span.SetAttributes(
		attribute.String("linkId", linkID.String()),
		attribute.String("workspaceId", workspaceID.String()),
		attribute.String("actorId", actorID.String()),
	)

	rowsAffected, err := r.queries.DeleteLinkForWorkspace(ctx, linksql.DeleteLinkForWorkspaceParams{
		LinkID:      linkID,
		WorkspaceID: workspaceID,
		ActorID:     actorID,
	})
	if err != nil {
		r.log.Error(ctx, "error deleting link", "linkId", linkID, "error", err)
		return fmt.Errorf("delete link: %w", err)
	}
	if rowsAffected == 0 {
		return linksdomain.ErrNotFound
	}

	r.log.Info(ctx, "link deleted successfully", "linkId", linkID)
	span.AddEvent("Link deleted.", trace.WithAttributes(attribute.String("link.id", linkID.String())))
	return nil
}
