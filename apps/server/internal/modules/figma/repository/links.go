package figmarepository

import (
	"context"
	"encoding/json"
	"strings"

	figmadomain "github.com/complexus-tech/projects-api/internal/modules/figma/domain"
	figmasql "github.com/complexus-tech/projects-api/internal/modules/figma/repository/sqlc"
	"github.com/google/uuid"
)

func (repository *Repository) ListStoryLinks(
	ctx context.Context,
	workspaceID, storyID uuid.UUID,
) ([]figmadomain.StoryLink, error) {
	rows, err := repository.queries.ListStoryLinks(ctx, figmasql.ListStoryLinksParams{
		WorkspaceID: workspaceID, StoryID: storyID,
	})
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	return mapStoryLinks(rows), nil
}

func (repository *Repository) ListStoryHandoffStatuses(
	ctx context.Context,
	workspaceID uuid.UUID,
) ([]figmadomain.StoryHandoffStatus, error) {
	rows, err := repository.queries.ListStoryHandoffStatuses(
		ctx,
		figmasql.ListStoryHandoffStatusesParams{WorkspaceID: workspaceID},
	)
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	statuses := make([]figmadomain.StoryHandoffStatus, 0, len(rows))
	for _, row := range rows {
		statuses = append(statuses, figmadomain.StoryHandoffStatus{
			StoryID: row.StoryID,
			Status:  row.Status,
		})
	}
	return statuses, nil
}

func (repository *Repository) ListLinksByFile(
	ctx context.Context,
	workspaceID uuid.UUID,
	fileKey string,
) ([]figmadomain.StoryLink, error) {
	rows, err := repository.queries.ListLinksByFile(ctx, figmasql.ListLinksByFileParams{
		WorkspaceID: workspaceID, FileKey: fileKey,
	})
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	return mapStoryLinks(rows), nil
}

func (repository *Repository) UpsertStoryLink(
	ctx context.Context,
	link figmadomain.StoryLink,
) (figmadomain.StoryLink, error) {
	metadata := normalizedMetadata(link.Artifact.Metadata)
	var result figmadomain.StoryLink
	err := repository.withinTransaction(ctx, func(queries figmasql.Querier) error {
		now := repository.currentTime()
		externalKey := "figma:" + link.StoryID.String() + ":" + link.Artifact.FileKey + ":"
		if link.Artifact.NodeID != nil {
			externalKey += *link.Artifact.NodeID
		}
		genericLinkID, err := queries.UpsertGenericStoryLink(
			ctx,
			figmasql.UpsertGenericStoryLinkParams{
				Title: storyLinkTitle(link.Artifact), URL: link.Artifact.CanonicalURL,
				ExternalSourceKey: &externalKey, ActorID: link.CreatedByUserID,
				StoryID: link.StoryID, WorkspaceID: link.WorkspaceID,
				UpdatedAt: now,
			},
		)
		if err != nil {
			return err
		}
		row, err := queries.UpsertFigmaStoryLink(ctx, figmasql.UpsertFigmaStoryLinkParams{
			StoryLinkID: &genericLinkID, FileKey: link.Artifact.FileKey,
			NodeID: link.Artifact.NodeID, OriginalURL: link.Artifact.OriginalURL,
			CanonicalURL: link.Artifact.CanonicalURL, FileName: link.Artifact.FileName,
			NodeName: link.Artifact.NodeName, NodeType: link.Artifact.NodeType,
			ThumbnailURL: link.Artifact.ThumbnailURL, Version: link.Artifact.Version,
			LastModified: link.Artifact.LastModified, Metadata: metadata,
			ActorID: link.CreatedByUserID, StoryID: link.StoryID,
			WorkspaceID: link.WorkspaceID, UnavailableAt: link.UnavailableAt,
			UpdatedAt: now,
		})
		if err != nil {
			return err
		}
		result = mapStoryLink(row)
		return nil
	})
	return result, err
}

func storyLinkTitle(artifact figmadomain.Artifact) *string {
	if artifact.NodeName != nil && strings.TrimSpace(*artifact.NodeName) != "" {
		value := strings.TrimSpace(*artifact.NodeName)
		return &value
	}
	if value := strings.TrimSpace(artifact.FileName); value != "" {
		return &value
	}
	return nil
}

func (repository *Repository) UpdateStoryLink(
	ctx context.Context,
	link figmadomain.StoryLink,
) error {
	rows, err := repository.queries.UpdateFigmaStoryLink(ctx, figmasql.UpdateFigmaStoryLinkParams{
		ID: link.ID, WorkspaceID: link.WorkspaceID, FileName: link.Artifact.FileName,
		NodeName: link.Artifact.NodeName, NodeType: link.Artifact.NodeType,
		ThumbnailURL: link.Artifact.ThumbnailURL, Version: link.Artifact.Version,
		LastModified: link.Artifact.LastModified, DevStatus: link.DevStatus,
		DevResourceID: link.DevResourceID, Metadata: normalizedMetadata(link.Artifact.Metadata),
		UnavailableAt: link.UnavailableAt, UpdatedAt: repository.currentTime(),
	})
	return requireAffected(rows, err, figmadomain.ErrNotFound)
}

func (repository *Repository) GetStoryLink(
	ctx context.Context,
	workspaceID, linkID uuid.UUID,
) (figmadomain.StoryLink, error) {
	row, err := repository.queries.GetFigmaStoryLink(ctx, figmasql.GetFigmaStoryLinkParams{
		WorkspaceID: workspaceID, ID: linkID,
	})
	if err != nil {
		return figmadomain.StoryLink{}, mapDatabaseError(err)
	}
	return mapStoryLink(row), nil
}

func (repository *Repository) DeleteStoryLink(
	ctx context.Context,
	workspaceID, linkID uuid.UUID,
) (figmadomain.StoryLink, error) {
	var result figmadomain.StoryLink
	err := repository.withinTransaction(ctx, func(queries figmasql.Querier) error {
		row, err := queries.DeleteFigmaStoryLink(ctx, figmasql.DeleteFigmaStoryLinkParams{
			ID: linkID, WorkspaceID: workspaceID,
		})
		if err != nil {
			return err
		}
		if row.StoryLinkID != nil {
			rows, err := queries.DeleteGenericStoryLink(ctx, figmasql.DeleteGenericStoryLinkParams{
				LinkID: *row.StoryLinkID,
			})
			if err := requireAffected(rows, err, figmadomain.ErrConflict); err != nil {
				return err
			}
		}
		result = mapStoryLink(row)
		return nil
	})
	return result, err
}

func normalizedMetadata(metadata json.RawMessage) []byte {
	if len(metadata) == 0 || !json.Valid(metadata) {
		return []byte(`{}`)
	}
	return append([]byte(nil), metadata...)
}

func mapStoryLinks(rows []figmasql.StoryFigmaLink) []figmadomain.StoryLink {
	links := make([]figmadomain.StoryLink, 0, len(rows))
	for _, row := range rows {
		links = append(links, mapStoryLink(row))
	}
	return links
}

func mapStoryLink(row figmasql.StoryFigmaLink) figmadomain.StoryLink {
	return figmadomain.StoryLink{
		ID: row.ID, WorkspaceID: row.WorkspaceID, StoryID: row.StoryID,
		CreatedByUserID: row.CreatedByUserID, StoryLinkID: row.StoryLinkID,
		Artifact: figmadomain.Artifact{
			FileKey: row.FileKey, NodeID: row.NodeID, OriginalURL: row.OriginalURL,
			CanonicalURL: row.CanonicalURL, FileName: row.FileName,
			NodeName: row.NodeName, NodeType: row.NodeType,
			ThumbnailURL: row.ThumbnailURL, Version: row.Version,
			LastModified: row.LastModified, Metadata: append([]byte(nil), row.Metadata...),
		},
		DevStatus: row.DevStatus, DevResourceID: row.DevResourceID,
		LastSyncedAt: row.LastSyncedAt, UnavailableAt: row.UnavailableAt,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}
