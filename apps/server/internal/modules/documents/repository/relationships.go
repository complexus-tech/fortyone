package documentsrepository

import (
	"context"
	"fmt"

	documentdomain "github.com/complexus-tech/projects-api/internal/modules/documents/domain"
	documentssql "github.com/complexus-tech/projects-api/internal/modules/documents/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) AddRelationship(
	ctx context.Context,
	input documentdomain.RelationshipInput,
) (documentdomain.RelatedWork, error) {
	if err := repository.configured(); err != nil {
		return documentdomain.RelatedWork{}, err
	}
	var related documentdomain.RelatedWork
	err := repository.withinSerializable(ctx, func(queries *documentssql.Queries) error {
		var err error
		related, err = getRelatedWork(ctx, queries, input.WorkspaceID, input.UserID, input.EntityType, input.EntityID)
		if err != nil {
			return err
		}
		_, err = queries.InsertEditableDocumentRelationship(ctx, documentssql.InsertEditableDocumentRelationshipParams{
			EntityType: string(input.EntityType), EntityID: input.EntityID, ActorID: input.UserID,
			DocumentID: input.DocumentID, WorkspaceID: input.WorkspaceID,
		})
		if err != nil {
			return mapNotFound("insert editable document relationship", err)
		}
		return nil
	})
	if err != nil {
		return documentdomain.RelatedWork{}, err
	}
	return related, nil
}

func (repository *Repository) RemoveRelationship(
	ctx context.Context,
	input documentdomain.RelationshipInput,
) error {
	if err := repository.configured(); err != nil {
		return err
	}
	return repository.withinSerializable(ctx, func(queries *documentssql.Queries) error {
		if _, err := getRelatedWork(ctx, queries, input.WorkspaceID, input.UserID, input.EntityType, input.EntityID); err != nil {
			return err
		}
		_, err := queries.DeleteEditableDocumentRelationship(ctx, documentssql.DeleteEditableDocumentRelationshipParams{
			DocumentID: input.DocumentID, WorkspaceID: input.WorkspaceID,
			EntityType: string(input.EntityType), EntityID: input.EntityID, ActorID: input.UserID,
		})
		if err != nil {
			return mapNotFound("delete editable document relationship", err)
		}
		return nil
	})
}

func (repository *Repository) ListRelatedDocuments(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	entityType documentdomain.RelationshipType,
	entityID uuid.UUID,
) ([]documentdomain.Summary, error) {
	if err := repository.configured(); err != nil {
		return nil, err
	}
	var result []documentdomain.Summary
	err := repository.transactor.WithinTransaction(
		ctx,
		pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly},
		func(tx pgx.Tx) error {
			queries := documentssql.New(tx)
			if _, err := getRelatedWork(ctx, queries, workspaceID, userID, entityType, entityID); err != nil {
				return err
			}
			rows, err := queries.ListAccessibleDocumentsForRelationship(
				ctx,
				documentssql.ListAccessibleDocumentsForRelationshipParams{
					ActorID: userID, WorkspaceID: workspaceID,
					EntityType: string(entityType), EntityID: entityID,
				},
			)
			if err != nil {
				return fmt.Errorf("list accessible documents for relationship: %w", err)
			}
			result = make([]documentdomain.Summary, 0, len(rows))
			for _, row := range rows {
				result = append(result, summaryFromRelationship(row))
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func listRelationships(
	ctx context.Context,
	queries *documentssql.Queries,
	workspaceID, documentID, userID uuid.UUID,
) ([]documentdomain.RelatedWork, error) {
	rows, err := queries.ListVisibleDocumentRelationships(ctx, documentssql.ListVisibleDocumentRelationshipsParams{
		ActorID: userID, WorkspaceID: workspaceID, DocumentID: documentID,
	})
	if err != nil {
		return nil, fmt.Errorf("list visible document relationships: %w", err)
	}
	result := make([]documentdomain.RelatedWork, 0, len(rows))
	for _, row := range rows {
		result = append(result, documentdomain.RelatedWork{
			EntityID: row.EntityID, EntityType: documentdomain.RelationshipType(row.EntityType),
			Title: row.Title, Reference: stringPointer(row.Reference), TeamID: uuidPointer(row.TeamID),
		})
	}
	return result, nil
}

func getRelatedWork(
	ctx context.Context,
	queries *documentssql.Queries,
	workspaceID, userID uuid.UUID,
	entityType documentdomain.RelationshipType,
	entityID uuid.UUID,
) (documentdomain.RelatedWork, error) {
	switch entityType {
	case documentdomain.RelationshipStory:
		row, err := queries.GetVisibleStoryRelationshipTarget(ctx, documentssql.GetVisibleStoryRelationshipTargetParams{
			ActorID: userID, EntityID: entityID, WorkspaceID: workspaceID,
		})
		if err != nil {
			return documentdomain.RelatedWork{}, mapNotFound("get visible story relationship target", err)
		}
		return documentdomain.RelatedWork{
			EntityID: row.EntityID, EntityType: documentdomain.RelationshipStory,
			Title: row.Title, Reference: stringPointer(row.Reference), TeamID: uuidPointer(row.TeamID),
		}, nil
	case documentdomain.RelationshipObjective:
		row, err := queries.GetVisibleObjectiveRelationshipTarget(ctx, documentssql.GetVisibleObjectiveRelationshipTargetParams{
			ActorID: userID, EntityID: entityID, WorkspaceID: uuidPointer(workspaceID),
		})
		if err != nil {
			return documentdomain.RelatedWork{}, mapNotFound("get visible objective relationship target", err)
		}
		return documentdomain.RelatedWork{
			EntityID: row.EntityID, EntityType: documentdomain.RelationshipObjective,
			Title: row.Title, Reference: stringPointer(row.Reference), TeamID: row.TeamID,
		}, nil
	default:
		return documentdomain.RelatedWork{}, documentdomain.ErrInvalidInput
	}
}
