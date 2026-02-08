package attachmentsrepository

import (
	"context"
	"fmt"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	"github.com/google/uuid"
)

// GetAttachmentByID gets an attachment by ID
func (r *Repository) GetAttachmentByID(ctx context.Context, id uuid.UUID) (attachments.CoreAttachment, error) {
	r.log.Info(ctx, "repo.attachments.getByID")

	const query = `
		SELECT attachment_id, filename, blob_name, size, mime_type, uploaded_by, workspace_id, created_at
		FROM attachments 
		WHERE attachment_id = :id
	`

	params := map[string]any{
		"id": id,
	}

	rows, err := r.db.NamedQueryContext(ctx, query, params)
	if err != nil {
		return attachments.CoreAttachment{}, fmt.Errorf("failed to get attachment: %w", err)
	}
	defer rows.Close()

	var dbAttachment dbAttachment
	if !rows.Next() {
		return attachments.CoreAttachment{}, attachments.ErrNotFound
	}

	if err := rows.StructScan(&dbAttachment); err != nil {
		return attachments.CoreAttachment{}, fmt.Errorf("failed to scan attachment: %w", err)
	}

	return toCoreAttachment(dbAttachment), nil
}

// GetAttachmentsByStoryID gets all attachments for a story
func (r *Repository) GetAttachmentsByStoryID(ctx context.Context, storyID uuid.UUID) ([]attachments.CoreAttachment, error) {
	r.log.Info(ctx, "repo.attachments.getByStoryID")

	const query = `
		SELECT a.attachment_id, a.filename, a.blob_name, a.size, a.mime_type, a.uploaded_by, a.workspace_id, a.created_at
		FROM attachments a
		JOIN story_attachments sa ON a.attachment_id = sa.attachment_id
		WHERE sa.story_id = :story_id
		ORDER BY a.created_at DESC
	`

	params := map[string]any{
		"story_id": storyID,
	}

	rows, err := r.db.NamedQueryContext(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get attachments: %w", err)
	}
	defer rows.Close()

	var attachmentsList []attachments.CoreAttachment
	for rows.Next() {
		var dbAttachment dbAttachment
		if err := rows.StructScan(&dbAttachment); err != nil {
			return nil, fmt.Errorf("failed to scan attachment: %w", err)
		}
		attachmentsList = append(attachmentsList, toCoreAttachment(dbAttachment))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating attachments: %w", err)
	}

	return attachmentsList, nil
}
