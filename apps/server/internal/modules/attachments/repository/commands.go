package attachmentsrepository

import (
	"context"
	"fmt"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	"github.com/google/uuid"
)

// CreateAttachment creates a new attachment in the database
func (r *Repository) CreateAttachment(ctx context.Context, attachment attachments.CoreAttachment) (attachments.CoreAttachment, error) {
	r.log.Info(ctx, "repo.attachments.create")

	const query = `
		INSERT INTO attachments 
		(filename, blob_name, size, mime_type, uploaded_by, workspace_id)
		VALUES (:filename, :blob_name, :size, :mime_type, :uploaded_by, :workspace_id)
		RETURNING attachment_id, filename, blob_name, size, mime_type, uploaded_by, workspace_id, created_at
	`

	params := map[string]any{
		"filename":     attachment.Filename,
		"blob_name":    attachment.BlobName,
		"size":         attachment.Size,
		"mime_type":    attachment.MimeType,
		"uploaded_by":  attachment.UploadedBy,
		"workspace_id": attachment.WorkspaceID,
	}

	rows, err := r.db.NamedQueryContext(ctx, query, params)
	if err != nil {
		return attachments.CoreAttachment{}, fmt.Errorf("failed to create attachment: %w", err)
	}
	defer rows.Close()

	var dbAttachment dbAttachment
	if rows.Next() {
		if err := rows.StructScan(&dbAttachment); err != nil {
			return attachments.CoreAttachment{}, fmt.Errorf("failed to scan attachment: %w", err)
		}
	}

	return toCoreAttachment(dbAttachment), nil
}

// DeleteAttachment deletes an attachment
func (r *Repository) DeleteAttachment(ctx context.Context, id uuid.UUID) error {
	r.log.Info(ctx, "repo.attachments.delete")

	const query = `DELETE FROM attachments WHERE attachment_id = :id`

	params := map[string]any{
		"id": id,
	}

	result, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to delete attachment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return attachments.ErrNotFound
	}

	return nil
}

// LinkAttachmentToStory links an attachment to a story
func (r *Repository) LinkAttachmentToStory(ctx context.Context, storyID, attachmentID uuid.UUID) error {
	r.log.Info(ctx, "repo.attachments.linkToStory")

	const query = `
		INSERT INTO story_attachments (story_id, attachment_id)
		VALUES (:story_id, :attachment_id)
		ON CONFLICT (story_id, attachment_id) DO NOTHING
	`

	params := map[string]any{
		"story_id":      storyID,
		"attachment_id": attachmentID,
	}

	_, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to link attachment to story: %w", err)
	}

	return nil
}
