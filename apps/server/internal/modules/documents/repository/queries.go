package documentsrepository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const documentColumns = `
	d.document_id,
	d.workspace_id,
	d.title,
	d.content_html,
	d.content_text,
	d.visibility,
	d.created_by,
	d.updated_by,
	d.created_at,
	d.updated_at,
	d.archived_at`

const documentSummaryColumns = `
	d.document_id,
	d.workspace_id,
	d.title,
	d.visibility,
	d.created_by,
	d.updated_by,
	d.created_at,
	d.updated_at`

func accessPredicate(documentAlias string) string {
	return fmt.Sprintf(`(
		%s.visibility = 'workspace'
		OR %s.created_by = :user_id
		OR EXISTS (
			SELECT 1
			FROM document_members access_member
			WHERE access_member.document_id = %s.document_id
				AND access_member.user_id = :user_id
		)
	)`, documentAlias, documentAlias, documentAlias)
}

func editPredicate(documentAlias string) string {
	return fmt.Sprintf(`(
		%s.visibility = 'workspace'
		OR %s.created_by = :user_id
		OR EXISTS (
			SELECT 1
			FROM document_members edit_member
			WHERE edit_member.document_id = %s.document_id
				AND edit_member.user_id = :user_id
				AND edit_member.role = 'editor'
		)
	)`, documentAlias, documentAlias, documentAlias)
}

func visibleRelationshipCount(documentAlias string) string {
	return fmt.Sprintf(`(
		SELECT COUNT(*)
		FROM document_relationships relationship_count
		LEFT JOIN stories related_story
			ON relationship_count.entity_type = 'story'
			AND related_story.id = relationship_count.entity_id
			AND related_story.workspace_id = relationship_count.workspace_id
			AND related_story.deleted_at IS NULL
		LEFT JOIN objectives related_objective
			ON relationship_count.entity_type = 'objective'
			AND related_objective.objective_id = relationship_count.entity_id
			AND related_objective.workspace_id = relationship_count.workspace_id
		WHERE relationship_count.document_id = %s.document_id
			AND relationship_count.workspace_id = %s.workspace_id
			AND (related_story.id IS NOT NULL OR related_objective.objective_id IS NOT NULL)
			AND EXISTS (
				SELECT 1
				FROM team_members relationship_viewer
				WHERE relationship_viewer.user_id = :user_id
					AND relationship_viewer.team_id = COALESCE(related_story.team_id, related_objective.team_id)
			)
	)`, documentAlias, documentAlias)
}

func (r *repo) List(ctx context.Context, input documents.CoreListInput) ([]documents.CoreDocumentSummary, error) {
	query, params := buildDocumentListQuery(input)

	rows := []dbDocumentSummary{}
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	if err := stmt.SelectContext(ctx, &rows, params); err != nil {
		return nil, err
	}
	return toCoreDocumentSummaries(rows), nil
}

func buildDocumentListQuery(input documents.CoreListInput) (string, map[string]any) {
	query := `
		SELECT ` + documentSummaryColumns + `,
			` + editPredicate("d") + ` AS can_edit,
			` + visibleRelationshipCount("d") + ` AS related_work_count
		FROM documents d
		WHERE d.workspace_id = :workspace_id
			AND d.archived_at IS NULL
			AND ` + accessPredicate("d")

	params := map[string]any{
		"workspace_id": input.WorkspaceID,
		"user_id":      input.UserID,
	}
	if input.Search != "" {
		query += " AND (d.title ILIKE :search OR d.content_text ILIKE :search)"
		params["search"] = "%" + input.Search + "%"
	}
	switch input.Scope {
	case "mine":
		query += " AND d.created_by = :user_id"
	case "shared":
		query += ` AND d.created_by <> :user_id
			AND EXISTS (
				SELECT 1
				FROM document_members shared_scope_member
				WHERE shared_scope_member.document_id = d.document_id
					AND shared_scope_member.user_id = :user_id
			)`
	}
	query += " ORDER BY d.updated_at DESC, d.document_id"
	if input.Limit != nil {
		query += " LIMIT :limit"
		params["limit"] = *input.Limit
	}

	return query, params
}

func (r *repo) Get(ctx context.Context, workspaceID, userID, documentID uuid.UUID) (documents.CoreDocument, error) {
	document, err := getDocument(ctx, r.db, workspaceID, userID, documentID)
	if err != nil {
		return documents.CoreDocument{}, err
	}
	if err := r.hydrateDocument(ctx, &document, userID); err != nil {
		return documents.CoreDocument{}, err
	}
	return document, nil
}

type namedPreparer interface {
	PrepareNamedContext(context.Context, string) (*sqlx.NamedStmt, error)
}

func getDocument(ctx context.Context, db namedPreparer, workspaceID, userID, documentID uuid.UUID) (documents.CoreDocument, error) {
	query := `
		SELECT ` + documentColumns + `,
			` + editPredicate("d") + ` AS can_edit
		FROM documents d
		WHERE d.document_id = :document_id
			AND d.workspace_id = :workspace_id
			AND d.archived_at IS NULL
			AND ` + accessPredicate("d")

	params := map[string]any{"document_id": documentID, "workspace_id": workspaceID, "user_id": userID}
	stmt, err := db.PrepareNamedContext(ctx, query)
	if err != nil {
		return documents.CoreDocument{}, err
	}
	defer stmt.Close()
	var row dbDocument
	if err := stmt.GetContext(ctx, &row, params); err != nil {
		return documents.CoreDocument{}, err
	}
	return toCoreDocument(row), nil
}

func (r *repo) hydrateDocument(ctx context.Context, document *documents.CoreDocument, userID uuid.UUID) error {
	members := []dbDocumentMember{}
	if err := r.db.SelectContext(ctx, &members, `
		SELECT user_id, role
		FROM document_members
		WHERE document_id = $1
		ORDER BY created_at, user_id`, document.ID); err != nil {
		return err
	}
	document.SharedWith = make([]documents.CoreDocumentMember, len(members))
	for i, member := range members {
		document.SharedWith[i] = documents.CoreDocumentMember{UserID: member.UserID, Role: member.Role}
	}

	related, err := r.listRelationships(ctx, document.WorkspaceID, document.ID, userID)
	if err != nil {
		return err
	}
	document.RelatedWork = related
	document.RelatedWorkCount = len(related)
	return nil
}

func (r *repo) Create(ctx context.Context, input documents.CoreCreateInput) (documents.CoreDocument, error) {
	query := `
		INSERT INTO documents (
			workspace_id, title, content_html, content_text, visibility, created_by, updated_by
		)
		VALUES (
			:workspace_id, :title, :content_html, :content_text, :visibility, :user_id, :user_id
		)
		RETURNING document_id, workspace_id, title, content_html, content_text,
			visibility, created_by, updated_by, created_at, updated_at, archived_at,
			TRUE AS can_edit`
	params := map[string]any{
		"workspace_id": input.WorkspaceID,
		"title":        input.Title,
		"content_html": input.ContentHTML,
		"content_text": input.ContentText,
		"visibility":   input.Visibility,
		"user_id":      input.UserID,
	}
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return documents.CoreDocument{}, err
	}
	defer stmt.Close()
	var row dbDocument
	if err := stmt.GetContext(ctx, &row, params); err != nil {
		return documents.CoreDocument{}, err
	}
	return toCoreDocument(row), nil
}

func (r *repo) Update(ctx context.Context, input documents.CoreUpdateInput) (documents.CoreDocument, error) {
	query := `
		UPDATE documents d
		SET title = COALESCE(:title, d.title),
			content_html = COALESCE(:content_html, d.content_html),
			content_text = COALESCE(:content_text, d.content_text),
			updated_by = :user_id,
			updated_at = CURRENT_TIMESTAMP
		WHERE d.document_id = :document_id
			AND d.workspace_id = :workspace_id
			AND d.archived_at IS NULL
			AND ` + editPredicate("d") + `
		RETURNING document_id, workspace_id, title, content_html, content_text,
			visibility, created_by, updated_by, created_at, updated_at, archived_at,
			TRUE AS can_edit`
	params := map[string]any{
		"document_id":  input.DocumentID,
		"workspace_id": input.WorkspaceID,
		"user_id":      input.UserID,
		"title":        input.Title,
		"content_html": input.ContentHTML,
		"content_text": input.ContentText,
	}
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return documents.CoreDocument{}, err
	}
	defer stmt.Close()
	var row dbDocument
	if err := stmt.GetContext(ctx, &row, params); err != nil {
		return documents.CoreDocument{}, err
	}
	document := toCoreDocument(row)
	if err := r.hydrateDocument(ctx, &document, input.UserID); err != nil {
		return documents.CoreDocument{}, err
	}
	return document, nil
}

func (r *repo) Duplicate(ctx context.Context, workspaceID, userID, documentID uuid.UUID) (documents.CoreDocument, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return documents.CoreDocument{}, err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO documents (
			workspace_id, title, content_html, content_text, visibility, created_by, updated_by
		)
		SELECT d.workspace_id, LEFT('Copy of ' || d.title, 255), d.content_html,
			d.content_text, 'private', :user_id, :user_id
		FROM documents d
		WHERE d.document_id = :document_id
			AND d.workspace_id = :workspace_id
			AND d.archived_at IS NULL
			AND ` + accessPredicate("d") + `
		RETURNING document_id, workspace_id, title, content_html, content_text,
			visibility, created_by, updated_by, created_at, updated_at, archived_at,
			TRUE AS can_edit`
	params := map[string]any{
		"workspace_id": workspaceID,
		"user_id":      userID,
		"document_id":  documentID,
	}
	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		return documents.CoreDocument{}, err
	}
	defer stmt.Close()

	var row dbDocument
	if err := stmt.GetContext(ctx, &row, params); err != nil {
		return documents.CoreDocument{}, err
	}

	oldMediaPath := "/documents/" + documentID.String() + "/media/"
	newMediaPath := "/documents/" + row.ID.String() + "/media/"
	if _, err := tx.ExecContext(ctx, `
		UPDATE documents
		SET content_html = REPLACE(content_html, $1, $2)
		WHERE document_id = $3`, oldMediaPath, newMediaPath, row.ID); err != nil {
		return documents.CoreDocument{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO document_attachments (document_id, attachment_id, created_by)
		SELECT $1, attachment_id, $2
		FROM document_attachments
		WHERE document_id = $3`, row.ID, userID, documentID); err != nil {
		return documents.CoreDocument{}, err
	}
	if err := tx.Commit(); err != nil {
		return documents.CoreDocument{}, err
	}

	document := toCoreDocument(row)
	document.ContentHTML = strings.ReplaceAll(document.ContentHTML, oldMediaPath, newMediaPath)
	if err := r.hydrateDocument(ctx, &document, userID); err != nil {
		return documents.CoreDocument{}, err
	}
	return document, nil
}

func (r *repo) Archive(ctx context.Context, workspaceID, userID, documentID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE documents
		SET archived_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP, updated_by = $1
		WHERE document_id = $2 AND workspace_id = $3 AND created_by = $1 AND archived_at IS NULL`, userID, documentID, workspaceID)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func (r *repo) Delete(ctx context.Context, workspaceID, userID, documentID uuid.UUID) ([]uuid.UUID, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	orphanedAttachmentIDs := []uuid.UUID{}
	if err := tx.SelectContext(ctx, &orphanedAttachmentIDs, `
		SELECT media.attachment_id
		FROM document_attachments media
		JOIN documents d ON d.document_id = media.document_id
		WHERE d.document_id = $1
			AND d.workspace_id = $2
			AND d.created_by = $3
			AND NOT EXISTS (
				SELECT 1
				FROM document_attachments other_media
				WHERE other_media.attachment_id = media.attachment_id
					AND other_media.document_id <> d.document_id
			)
			AND NOT EXISTS (
				SELECT 1
				FROM story_attachments story_media
				WHERE story_media.attachment_id = media.attachment_id
			)`, documentID, workspaceID, userID); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		DELETE FROM documents
		WHERE document_id = $1 AND workspace_id = $2 AND created_by = $3`, documentID, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	if err := requireAffected(result); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return orphanedAttachmentIDs, nil
}

func (r *repo) SetAccess(ctx context.Context, input documents.CoreAccessInput) (documents.CoreDocument, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return documents.CoreDocument{}, err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		UPDATE documents
		SET visibility = $1, updated_at = CURRENT_TIMESTAMP, updated_by = $2
		WHERE document_id = $3 AND workspace_id = $4 AND created_by = $2 AND archived_at IS NULL`, input.Visibility, input.UserID, input.DocumentID, input.WorkspaceID)
	if err != nil {
		return documents.CoreDocument{}, err
	}
	if err := requireAffected(result); err != nil {
		return documents.CoreDocument{}, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM document_members WHERE document_id = $1`, input.DocumentID); err != nil {
		return documents.CoreDocument{}, err
	}
	if input.Visibility == documents.VisibilityRestricted {
		for _, member := range input.Members {
			result, err := tx.ExecContext(ctx, `
				INSERT INTO document_members (document_id, user_id, role)
				SELECT $1, wm.user_id,
					CASE WHEN wm.role = 'guest' THEN 'viewer' ELSE $3 END
				FROM workspace_members wm
				WHERE wm.workspace_id = $2 AND wm.user_id = $4
				ON CONFLICT (document_id, user_id) DO UPDATE SET role = EXCLUDED.role`, input.DocumentID, input.WorkspaceID, member.Role, member.UserID)
			if err != nil {
				return documents.CoreDocument{}, err
			}
			if err := requireAffected(result); err != nil {
				return documents.CoreDocument{}, documents.ErrInvalidInput
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return documents.CoreDocument{}, err
	}
	return r.Get(ctx, input.WorkspaceID, input.UserID, input.DocumentID)
}

func (r *repo) AddRelationship(ctx context.Context, input documents.CoreRelationshipInput) (documents.CoreRelatedWork, error) {
	document, err := getDocument(ctx, r.db, input.WorkspaceID, input.UserID, input.DocumentID)
	if err != nil {
		return documents.CoreRelatedWork{}, err
	}
	if !document.CanEdit {
		return documents.CoreRelatedWork{}, documents.ErrForbidden
	}
	related, err := r.getRelatedWork(ctx, input.WorkspaceID, input.UserID, input.EntityType, input.EntityID)
	if err != nil {
		return documents.CoreRelatedWork{}, err
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO document_relationships (document_id, workspace_id, entity_type, entity_id, created_by)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (document_id, entity_type, entity_id) DO NOTHING`, input.DocumentID, input.WorkspaceID, input.EntityType, input.EntityID, input.UserID)
	if err != nil {
		return documents.CoreRelatedWork{}, err
	}
	return related, nil
}

func (r *repo) RemoveRelationship(ctx context.Context, input documents.CoreRelationshipInput) error {
	if _, err := r.getRelatedWork(ctx, input.WorkspaceID, input.UserID, input.EntityType, input.EntityID); err != nil {
		return err
	}
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM document_relationships relationship
		USING documents d
		WHERE relationship.document_id = d.document_id
			AND relationship.document_id = $1
			AND relationship.workspace_id = $2
			AND relationship.entity_type = $3
			AND relationship.entity_id = $4
			AND d.archived_at IS NULL
			AND (
				d.visibility = 'workspace'
				OR d.created_by = $5
				OR EXISTS (
					SELECT 1 FROM document_members member
					WHERE member.document_id = d.document_id
						AND member.user_id = $5
						AND member.role = 'editor'
				)
			)`, input.DocumentID, input.WorkspaceID, input.EntityType, input.EntityID, input.UserID)
	if err != nil {
		return err
	}
	return requireAffected(result)
}

func (r *repo) ListRelatedDocuments(ctx context.Context, workspaceID, userID uuid.UUID, entityType documents.RelationshipType, entityID uuid.UUID) ([]documents.CoreDocumentSummary, error) {
	if _, err := r.getRelatedWork(ctx, workspaceID, userID, entityType, entityID); err != nil {
		return nil, err
	}
	query := `
		SELECT ` + documentSummaryColumns + `,
			` + editPredicate("d") + ` AS can_edit,
			` + visibleRelationshipCount("d") + ` AS related_work_count
		FROM document_relationships relationship
		JOIN documents d ON d.document_id = relationship.document_id
		WHERE relationship.workspace_id = :workspace_id
			AND relationship.entity_type = :entity_type
			AND relationship.entity_id = :entity_id
			AND d.archived_at IS NULL
			AND ` + accessPredicate("d") + `
		ORDER BY d.updated_at DESC`
	params := map[string]any{"workspace_id": workspaceID, "user_id": userID, "entity_type": entityType, "entity_id": entityID}
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows := []dbDocumentSummary{}
	if err := stmt.SelectContext(ctx, &rows, params); err != nil {
		return nil, err
	}
	return toCoreDocumentSummaries(rows), nil
}

func (r *repo) LinkMedia(ctx context.Context, input documents.CoreMediaInput) error {
	query := `
		INSERT INTO document_attachments (document_id, attachment_id, created_by)
		SELECT d.document_id, attachment.attachment_id, :user_id
		FROM documents d
		JOIN attachments attachment
			ON attachment.attachment_id = :attachment_id
			AND attachment.workspace_id = d.workspace_id
		WHERE d.document_id = :document_id
			AND d.workspace_id = :workspace_id
			AND d.archived_at IS NULL
			AND ` + editPredicate("d") + `
		ON CONFLICT (document_id, attachment_id) DO NOTHING`
	params := mediaParams(input)
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	// A duplicate link is idempotent, but only for a user who still has edit
	// access. This also prevents a zero-row insert from masking denied access.
	return r.authorizeMedia(ctx, input, editPredicate("d"))
}

func (r *repo) UnlinkMedia(ctx context.Context, input documents.CoreMediaInput) (bool, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	query := `
		DELETE FROM document_attachments media
		USING documents d
		WHERE media.document_id = d.document_id
			AND media.attachment_id = :attachment_id
			AND d.document_id = :document_id
			AND d.workspace_id = :workspace_id
			AND d.archived_at IS NULL
			AND ` + editPredicate("d") + `
		RETURNING media.attachment_id`
	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		return false, err
	}
	defer stmt.Close()
	var attachmentID uuid.UUID
	if err := stmt.GetContext(ctx, &attachmentID, mediaParams(input)); err != nil {
		return false, err
	}

	var isOrphaned bool
	if err := tx.GetContext(ctx, &isOrphaned, `
		SELECT NOT EXISTS (
			SELECT 1 FROM document_attachments WHERE attachment_id = $1
		) AND NOT EXISTS (
			SELECT 1 FROM story_attachments WHERE attachment_id = $1
		)`, attachmentID); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return isOrphaned, nil
}

func (r *repo) AuthorizeMedia(ctx context.Context, input documents.CoreMediaInput) error {
	return r.authorizeMedia(ctx, input, accessPredicate("d"))
}

func (r *repo) authorizeMedia(ctx context.Context, input documents.CoreMediaInput, permissionPredicate string) error {
	query := `
		SELECT 1
		FROM document_attachments media
		JOIN documents d ON d.document_id = media.document_id
		JOIN attachments attachment ON attachment.attachment_id = media.attachment_id
		WHERE d.document_id = :document_id
			AND media.attachment_id = :attachment_id
			AND d.workspace_id = :workspace_id
			AND attachment.workspace_id = d.workspace_id
			AND d.archived_at IS NULL
			AND ` + permissionPredicate
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	var authorized int
	if err := stmt.GetContext(ctx, &authorized, mediaParams(input)); err != nil {
		return err
	}
	return nil
}

func mediaParams(input documents.CoreMediaInput) map[string]any {
	return map[string]any{
		"workspace_id":  input.WorkspaceID,
		"user_id":       input.UserID,
		"document_id":   input.DocumentID,
		"attachment_id": input.AttachmentID,
	}
}

func (r *repo) listRelationships(ctx context.Context, workspaceID, documentID, userID uuid.UUID) ([]documents.CoreRelatedWork, error) {
	rows := []dbRelatedWork{}
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT relationship.entity_id, relationship.entity_type,
			COALESCE(story.title, objective.name) AS title,
			COALESCE(story_team.code || '-' || story.sequence_id, objective_team.code || '-' || objective.sequence_id) AS reference,
			COALESCE(story.team_id, objective.team_id) AS team_id
		FROM document_relationships relationship
		LEFT JOIN stories story ON relationship.entity_type = 'story' AND story.id = relationship.entity_id AND story.workspace_id = relationship.workspace_id AND story.deleted_at IS NULL
		LEFT JOIN teams story_team ON story_team.team_id = story.team_id
		LEFT JOIN objectives objective ON relationship.entity_type = 'objective' AND objective.objective_id = relationship.entity_id AND objective.workspace_id = relationship.workspace_id
		LEFT JOIN teams objective_team ON objective_team.team_id = objective.team_id
		WHERE relationship.workspace_id = $1 AND relationship.document_id = $2
			AND (story.id IS NOT NULL OR objective.objective_id IS NOT NULL)
			AND EXISTS (
				SELECT 1
				FROM team_members viewer_membership
				WHERE viewer_membership.user_id = $3
					AND viewer_membership.team_id = COALESCE(story.team_id, objective.team_id)
			)
		ORDER BY relationship.created_at`, workspaceID, documentID, userID); err != nil {
		return nil, err
	}
	result := make([]documents.CoreRelatedWork, len(rows))
	for i, row := range rows {
		result[i] = toCoreRelatedWork(row)
	}
	return result, nil
}

func (r *repo) getRelatedWork(ctx context.Context, workspaceID, userID uuid.UUID, entityType documents.RelationshipType, entityID uuid.UUID) (documents.CoreRelatedWork, error) {
	var row dbRelatedWork
	var err error
	switch entityType {
	case documents.RelationshipStory:
		err = r.db.GetContext(ctx, &row, `
			SELECT story.id AS entity_id, 'story' AS entity_type, story.title,
				team.code || '-' || story.sequence_id AS reference, story.team_id
			FROM stories story
			JOIN teams team ON team.team_id = story.team_id
			JOIN team_members membership ON membership.team_id = story.team_id AND membership.user_id = $3
			WHERE story.id = $1 AND story.workspace_id = $2 AND story.deleted_at IS NULL`, entityID, workspaceID, userID)
	case documents.RelationshipObjective:
		err = r.db.GetContext(ctx, &row, `
			SELECT objective.objective_id AS entity_id, 'objective' AS entity_type, objective.name AS title,
				team.code || '-' || objective.sequence_id AS reference, objective.team_id
			FROM objectives objective
			JOIN teams team ON team.team_id = objective.team_id
			JOIN team_members membership ON membership.team_id = objective.team_id AND membership.user_id = $3
			WHERE objective.objective_id = $1 AND objective.workspace_id = $2`, entityID, workspaceID, userID)
	default:
		return documents.CoreRelatedWork{}, documents.ErrInvalidInput
	}
	if err != nil {
		return documents.CoreRelatedWork{}, err
	}
	return toCoreRelatedWork(row), nil
}

func requireAffected(result sql.Result) error {
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}
