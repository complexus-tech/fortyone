package attachmentsrepo

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/core/attachments"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository provides access to the attachments store
type Repository struct {
	log *logger.Logger
	db  *sqlx.DB
}

// New creates a new attachments repository
func New(log *logger.Logger, db *sqlx.DB) *Repository {
	return &Repository{
		log: log,
		db:  db,
	}
}

// CreateAttachment creates a new attachment in the database
func (r *Repository) CreateAttachment(ctx context.Context, attachment attachments.CoreAttachment) (attachments.CoreAttachment, error) {
	r.log.Info(ctx, "repo.attachments.create")

	const query = `
		INSERT INTO attachments 
		(filename, size, mime_type, uploaded_by, workspace_id)
		VALUES (:filename, :size, :mime_type, :uploaded_by, :workspace_id)
		RETURNING attachment_id, filename, size, mime_type, uploaded_by, workspace_id, created_at
	`

	params := map[string]any{
		"filename":     attachment.Filename,
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

// GetAttachmentByID gets an attachment by ID
func (r *Repository) GetAttachmentByID(ctx context.Context, id uuid.UUID) (attachments.CoreAttachment, error) {
	r.log.Info(ctx, "repo.attachments.getByID")

	const query = `
		SELECT attachment_id, filename, size, mime_type, uploaded_by, team_id, workspace_id, created_at
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
		SELECT a.attachment_id, a.filename, a.size, a.mime_type, a.uploaded_by, a.workspace_id, a.created_at
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

// UnlinkAttachmentFromStory unlinks an attachment from a story
func (r *Repository) UnlinkAttachmentFromStory(ctx context.Context, storyID, attachmentID uuid.UUID) error {
	r.log.Info(ctx, "repo.attachments.unlinkFromStory")

	const query = `DELETE FROM story_attachments WHERE story_id = :story_id AND attachment_id = :attachment_id`

	params := map[string]any{
		"story_id":      storyID,
		"attachment_id": attachmentID,
	}

	result, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to unlink attachment from story: %w", err)
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
