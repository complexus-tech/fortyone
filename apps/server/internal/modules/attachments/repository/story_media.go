package attachmentsrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	"github.com/google/uuid"
)

// StoryExistsInWorkspace reports whether a story belongs to the workspace.
// It is used before uploading inline media so invalid story identities do not
// leave unlinked storage objects behind.
func (r *Repository) StoryExistsInWorkspace(ctx context.Context, storyID, workspaceID uuid.UUID) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM stories
			WHERE id = :story_id AND workspace_id = :workspace_id
		)`

	params := map[string]any{
		"story_id":     storyID,
		"workspace_id": workspaceID,
	}
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return false, fmt.Errorf("prepare story media workspace check: %w", err)
	}
	defer stmt.Close()

	var exists bool
	if err := stmt.GetContext(ctx, &exists, params); err != nil {
		return false, fmt.Errorf("check story media workspace: %w", err)
	}
	return exists, nil
}

// LinkStoryMedia links a dedicated inline attachment to a story while
// requiring both records to belong to the requested workspace.
func (r *Repository) LinkStoryMedia(ctx context.Context, storyID, attachmentID, createdBy, workspaceID uuid.UUID) error {
	const query = `
		INSERT INTO story_inline_attachments (story_id, attachment_id, created_by)
		SELECT story.id, attachment.attachment_id, :created_by
		FROM stories story
		JOIN attachments attachment
			ON attachment.attachment_id = :attachment_id
			AND attachment.workspace_id = story.workspace_id
		WHERE story.id = :story_id
			AND story.workspace_id = :workspace_id
		ON CONFLICT (story_id, attachment_id) DO NOTHING`

	params := storyMediaParams(storyID, attachmentID, workspaceID)
	params["created_by"] = createdBy
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare story media link: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		return fmt.Errorf("link story media: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read story media link result: %w", err)
	}
	if count > 0 {
		return nil
	}

	// A duplicate link is idempotent, but a workspace/story mismatch remains
	// indistinguishable from a missing attachment unless we re-authorize it.
	_, err = r.AuthorizeStoryMedia(ctx, storyID, attachmentID, workspaceID)
	return err
}

// AuthorizeStoryMedia returns one attachment only when the exact story,
// workspace, and inline-media link all match.
func (r *Repository) AuthorizeStoryMedia(ctx context.Context, storyID, attachmentID, workspaceID uuid.UUID) (attachments.CoreAttachment, error) {
	const query = `
		SELECT attachment.attachment_id, attachment.filename, attachment.blob_name,
			attachment.size, attachment.mime_type, attachment.uploaded_by,
			attachment.workspace_id, attachment.created_at
		FROM story_inline_attachments media
		JOIN stories story ON story.id = media.story_id
		JOIN attachments attachment ON attachment.attachment_id = media.attachment_id
		WHERE media.story_id = :story_id
			AND media.attachment_id = :attachment_id
			AND story.workspace_id = :workspace_id
			AND attachment.workspace_id = story.workspace_id`

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return attachments.CoreAttachment{}, fmt.Errorf("prepare story media authorization: %w", err)
	}
	defer stmt.Close()

	var row dbAttachment
	if err := stmt.GetContext(ctx, &row, storyMediaParams(storyID, attachmentID, workspaceID)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return attachments.CoreAttachment{}, attachments.ErrNotFound
		}
		return attachments.CoreAttachment{}, fmt.Errorf("authorize story media: %w", err)
	}
	return toCoreAttachment(row), nil
}

// UnlinkStoryMedia removes only the exact authorized story-media relation and
// reports whether no feature still references the attachment.
func (r *Repository) UnlinkStoryMedia(ctx context.Context, storyID, attachmentID, workspaceID uuid.UUID) (bool, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin story media unlink: %w", err)
	}
	defer tx.Rollback()

	const query = `
		DELETE FROM story_inline_attachments media
		USING stories story, attachments attachment
		WHERE media.story_id = story.id
			AND media.attachment_id = attachment.attachment_id
			AND media.story_id = :story_id
			AND media.attachment_id = :attachment_id
			AND story.workspace_id = :workspace_id
			AND attachment.workspace_id = story.workspace_id
		RETURNING media.attachment_id`
	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		return false, fmt.Errorf("prepare story media unlink: %w", err)
	}
	defer stmt.Close()

	var unlinkedAttachmentID uuid.UUID
	if err := stmt.GetContext(ctx, &unlinkedAttachmentID, storyMediaParams(storyID, attachmentID, workspaceID)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, attachments.ErrNotFound
		}
		return false, fmt.Errorf("unlink story media: %w", err)
	}

	var isOrphaned bool
	if err := tx.GetContext(ctx, &isOrphaned, `
		SELECT NOT EXISTS (
			SELECT 1 FROM story_inline_attachments WHERE attachment_id = $1
		) AND NOT EXISTS (
			SELECT 1 FROM story_attachments WHERE attachment_id = $1
		) AND NOT EXISTS (
			SELECT 1 FROM document_attachments WHERE attachment_id = $1
		)`, unlinkedAttachmentID); err != nil {
		return false, fmt.Errorf("check story media references: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit story media unlink: %w", err)
	}
	return isOrphaned, nil
}

func storyMediaParams(storyID, attachmentID, workspaceID uuid.UUID) map[string]any {
	return map[string]any{
		"story_id":      storyID,
		"attachment_id": attachmentID,
		"workspace_id":  workspaceID,
	}
}
