package storiesrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

type linkedStoryMedia struct {
	AttachmentID uuid.UUID `db:"attachment_id"`
	MimeType     string    `db:"mime_type"`
}

var reconcilableStoryFields = map[string]struct{}{
	"assignee_id":                 {},
	"completed_at":                {},
	"description":                 {},
	"description_html":            {},
	"end_date":                    {},
	"estimate_unit":               {},
	"estimated_duration_minutes":  {},
	"minimum_focus_block_minutes": {},
	"auto_scheduling_enabled":     {},
	"auto_scheduling_locked":      {},
	"auto_scheduling_status":      {},
	"auto_scheduling_reason":      {},
	"auto_scheduling_updated_at":  {},
	"key_result_id":               {},
	"objective_id":                {},
	"parent_id":                   {},
	"priority":                    {},
	"sprint_id":                   {},
	"start_date":                  {},
	"status_id":                   {},
	"title":                       {},
}

// UpdateWithMediaReconciliation updates a story and reconciles its dedicated
// inline-media links in one transaction. referencedAttachmentIDs must come
// from a validated, authoritative editor snapshot.
func (r *repo) UpdateWithMediaReconciliation(
	ctx context.Context,
	storyID, workspaceID uuid.UUID,
	updates map[string]any,
	referencedAttachmentIDs []uuid.UUID,
) ([]uuid.UUID, error) {
	if storyID == uuid.Nil || workspaceID == uuid.Nil {
		return nil, stories.ErrInvalidStoryMediaReference
	}

	if statusID, ok := updates["status_id"].(uuid.UUID); ok {
		if err := r.validateStatusTeamForStory(ctx, storyID, workspaceID, statusID); err != nil {
			return nil, err
		}
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin story media reconciliation: %w", err)
	}
	defer tx.Rollback()

	var lockedStoryID uuid.UUID
	if err := tx.GetContext(ctx, &lockedStoryID, `
		SELECT id
		FROM stories
		WHERE id = $1 AND workspace_id = $2
		FOR UPDATE`, storyID, workspaceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, stories.ErrNotFound
		}
		return nil, fmt.Errorf("lock story for media reconciliation: %w", err)
	}

	linkedMedia := []linkedStoryMedia{}
	if err := tx.SelectContext(ctx, &linkedMedia, `
		SELECT media.attachment_id, attachment.mime_type
		FROM story_inline_attachments media
		JOIN attachments attachment ON attachment.attachment_id = media.attachment_id
		WHERE media.story_id = $1
			AND attachment.workspace_id = $2
		ORDER BY media.attachment_id
		FOR UPDATE OF media, attachment`, storyID, workspaceID); err != nil {
		return nil, fmt.Errorf("lock story media links: %w", err)
	}

	linkedByID := make(map[uuid.UUID]linkedStoryMedia, len(linkedMedia))
	for _, media := range linkedMedia {
		linkedByID[media.AttachmentID] = media
	}
	referencedIDs := make(map[uuid.UUID]struct{}, len(referencedAttachmentIDs))
	for _, attachmentID := range referencedAttachmentIDs {
		media, linked := linkedByID[attachmentID]
		if attachmentID == uuid.Nil || !linked || !isStoryInlineMediaType(media.MimeType) {
			return nil, stories.ErrInvalidStoryMediaReference
		}
		referencedIDs[attachmentID] = struct{}{}
	}

	if err := updateStoryInMediaTransaction(ctx, tx, storyID, workspaceID, updates); err != nil {
		return nil, err
	}

	orphanedAttachmentIDs := make([]uuid.UUID, 0)
	for _, media := range linkedMedia {
		if _, referenced := referencedIDs[media.AttachmentID]; referenced {
			continue
		}

		result, err := tx.ExecContext(ctx, `
			DELETE FROM story_inline_attachments
			WHERE story_id = $1 AND attachment_id = $2`, storyID, media.AttachmentID)
		if err != nil {
			return nil, fmt.Errorf("unlink removed story media: %w", err)
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return nil, fmt.Errorf("read removed story media result: %w", err)
		}
		if rowsAffected == 0 {
			continue
		}

		var orphaned bool
		if err := tx.GetContext(ctx, &orphaned, `
			SELECT NOT EXISTS (
				SELECT 1 FROM story_inline_attachments WHERE attachment_id = $1
			) AND NOT EXISTS (
				SELECT 1 FROM story_attachments WHERE attachment_id = $1
			) AND NOT EXISTS (
				SELECT 1 FROM document_attachments WHERE attachment_id = $1
			)`, media.AttachmentID); err != nil {
			return nil, fmt.Errorf("check removed story media references: %w", err)
		}
		if orphaned {
			orphanedAttachmentIDs = append(orphanedAttachmentIDs, media.AttachmentID)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit story media reconciliation: %w", err)
	}
	return orphanedAttachmentIDs, nil
}

func (r *repo) validateStatusTeamForStory(
	ctx context.Context,
	storyID, workspaceID, statusID uuid.UUID,
) error {
	var teamID uuid.UUID
	if err := r.db.GetContext(ctx, &teamID, `
		SELECT team_id
		FROM stories
		WHERE id = $1 AND workspace_id = $2`, storyID, workspaceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return stories.ErrNotFound
		}
		return fmt.Errorf("get story team for media reconciliation: %w", err)
	}
	return r.validateStatusTeam(ctx, statusID, teamID)
}

func updateStoryInMediaTransaction(
	ctx context.Context,
	tx interface {
		NamedExecContext(context.Context, string, any) (sql.Result, error)
	},
	storyID, workspaceID uuid.UUID,
	updates map[string]any,
) error {
	if len(updates) == 0 {
		return nil
	}

	fields := make([]string, 0, len(updates))
	params := make(map[string]any, len(updates)+2)
	params["id"] = storyID
	params["workspace_id"] = workspaceID
	for field, value := range updates {
		if _, allowed := reconcilableStoryFields[field]; !allowed {
			return fmt.Errorf("%w: unsupported story field %q", stories.ErrInvalidStoryMediaReference, field)
		}
		fields = append(fields, field)
		params[field] = value
	}
	sort.Strings(fields)

	setClauses := make([]string, 0, len(fields)+1)
	for _, field := range fields {
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", field, field))
	}
	setClauses = append(setClauses, "updated_at = NOW()")
	query := `
		WITH updated_story AS (
			UPDATE stories
			SET ` + strings.Join(setClauses, ", ") + `
			WHERE id = :id AND workspace_id = :workspace_id
			RETURNING id, assignee_id
		)
		DELETE FROM story_collaborators collaborator
		USING updated_story story
		WHERE collaborator.story_id = story.id
			AND collaborator.user_id = story.assignee_id`
	if _, err := tx.NamedExecContext(ctx, query, params); err != nil {
		return fmt.Errorf("update story with media reconciliation: %w", err)
	}
	return nil
}

func isStoryInlineMediaType(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/") || mimeType == "video/mp4"
}
