package documentsrepository

import (
	"context"
	"fmt"

	documentdomain "github.com/complexus-tech/projects-api/internal/modules/documents/domain"
	documentssql "github.com/complexus-tech/projects-api/internal/modules/documents/repository/sqlc"
	"github.com/google/uuid"
)

func (repository *Repository) SetAccess(
	ctx context.Context,
	input documentdomain.AccessInput,
) (documentdomain.Document, error) {
	if err := repository.configured(); err != nil {
		return documentdomain.Document{}, err
	}
	var document documentdomain.Document
	err := repository.withinSerializable(ctx, func(queries *documentssql.Queries) error {
		if _, err := queries.SetOwnedDocumentVisibility(ctx, documentssql.SetOwnedDocumentVisibilityParams{
			Visibility: string(input.Visibility), ActorID: input.UserID,
			DocumentID: input.DocumentID, WorkspaceID: input.WorkspaceID,
		}); err != nil {
			return mapNotFound("set owned document visibility", err)
		}
		deleteParams := documentssql.DeleteOwnedDocumentMembersParams{
			DocumentID: input.DocumentID, WorkspaceID: input.WorkspaceID, ActorID: input.UserID,
		}
		if err := queries.DeleteOwnedDocumentMembers(ctx, deleteParams); err != nil {
			return fmt.Errorf("delete owned document members: %w", err)
		}
		if input.Visibility == documentdomain.VisibilityRestricted && len(input.Members) > 0 {
			memberIDs := make([]uuid.UUID, 0, len(input.Members))
			memberRoles := make([]string, 0, len(input.Members))
			for _, member := range input.Members {
				memberIDs = append(memberIDs, member.UserID)
				memberRoles = append(memberRoles, member.Role)
			}
			inserted, err := queries.InsertActiveWorkspaceDocumentMembers(
				ctx,
				documentssql.InsertActiveWorkspaceDocumentMembersParams{
					DocumentID: input.DocumentID, WorkspaceID: input.WorkspaceID,
					ActorID: input.UserID, MemberIds: memberIDs, MemberRoles: memberRoles,
				},
			)
			if err != nil {
				return fmt.Errorf("insert active workspace document members: %w", err)
			}
			if inserted != int64(len(input.Members)) {
				return documentdomain.ErrInvalidInput
			}
		}
		var err error
		document, err = getDocument(ctx, queries, input.WorkspaceID, input.UserID, input.DocumentID)
		return err
	})
	if err != nil {
		return documentdomain.Document{}, err
	}
	return document, nil
}
