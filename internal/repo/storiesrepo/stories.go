package storiesrepo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/comments"
	"github.com/complexus-tech/projects-api/internal/core/links"
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/internal/repo/commentsrepo"
	"github.com/complexus-tech/projects-api/internal/repo/linksrepo"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type repo struct {
	db  *sqlx.DB
	log *logger.Logger
}

func New(log *logger.Logger, db *sqlx.DB) *repo {
	return &repo{
		db:  db,
		log: log,
	}
}

// GetNextSequenceID returns the next sequence ID for a team.
func (r *repo) GetNextSequenceID(ctx context.Context, teamID uuid.UUID, workspaceId uuid.UUID) (int, func() error, func() error, error) {
	var currentSequence int

	// Start a transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	query := `
		UPDATE team_story_sequences
		SET current_sequence = current_sequence + 1
		WHERE workspace_id = :workspace_id AND team_id = :team_id
		RETURNING current_sequence
	`
	params := map[string]any{
		"team_id":      teamID,
		"workspace_id": workspaceId,
	}

	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		tx.Rollback()
		return 0, nil, nil, fmt.Errorf("failed to prepare named statement: %w", err)
	}
	defer stmt.Close()

	err = stmt.GetContext(ctx, &currentSequence, params)
	if errors.Is(err, sql.ErrNoRows) {
		// If no record exists, insert a new one starting from 1
		q := `
			INSERT INTO team_story_sequences (workspace_id, team_id, current_sequence)
			VALUES (:workspace_id, :team_id, 0)
			RETURNING current_sequence
		`
		stmt, err = tx.PrepareNamedContext(ctx, q)
		if err != nil {
			tx.Rollback()
			return 0, nil, nil, fmt.Errorf("failed to prepare named statement for insert: %w", err)
		}
		defer stmt.Close()

		err = stmt.GetContext(ctx, &currentSequence, params)
	}

	if err != nil {
		tx.Rollback()
		return 0, nil, nil, fmt.Errorf("failed to get/update sequence: %w", err)
	}

	commit := func() error {
		return tx.Commit()
	}

	rollback := func() error {
		return tx.Rollback()
	}

	return currentSequence, commit, rollback, nil
}

// Create creates a new story.
func (r *repo) Create(ctx context.Context, story *stories.CoreSingleStory) (stories.CoreSingleStory, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.Create")
	defer span.End()

	// Validate status belongs to the same team
	if story.Status != nil {
		if err := r.validateStatusTeam(ctx, *story.Status, story.Team); err != nil {
			return stories.CoreSingleStory{}, err
		}
	}

	lastSequence, commit, rollback, err := r.GetNextSequenceID(ctx, story.Team, story.Workspace)
	if err != nil {
		return stories.CoreSingleStory{}, fmt.Errorf("failed to get next sequence ID: %w", err)
	}
	story.SequenceID = lastSequence + 1

	cs, err := r.insertStory(ctx, story)
	if err != nil {
		rollback()
		return stories.CoreSingleStory{}, fmt.Errorf("failed to insert story: %w", err)
	}

	if err := commit(); err != nil {
		return stories.CoreSingleStory{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return toCoreStory(cs), nil
}

func (r *repo) insertStory(ctx context.Context, story *stories.CoreSingleStory) (dbStory, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.insertStory")
	defer span.End()

	q := `
			INSERT INTO stories (
					sequence_id, title, description, description_html,
					parent_id, objective_id, status_id, assignee_id, 
					blocked_by_id, blocking_id, related_id, reporter_id,
					priority, sprint_id, key_result_id, team_id, workspace_id, start_date, 
					end_date, created_at, updated_at
			) VALUES (
					:sequence_id, :title, :description, :description_html,
					:parent_id, :objective_id, :status_id, :assignee_id, :blocked_by_id,
					:blocking_id, :related_id, :reporter_id, :priority, :sprint_id,
					:key_result_id, :team_id, :workspace_id, :start_date, :end_date, :created_at, :updated_at
			) RETURNING stories.id, stories.sequence_id, stories.title, stories.description, stories.description_html, stories.parent_id, stories.objective_id, stories.status_id, stories.assignee_id, stories.blocked_by_id, stories.blocking_id, stories.related_id, stories.reporter_id, stories.priority, stories.sprint_id, stories.key_result_id, stories.team_id, stories.workspace_id, stories.start_date, stories.end_date, stories.created_at, stories.updated_at;
		`

	var cs dbStory
	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return dbStory{}, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "creating story.")
	if err := stmt.GetContext(ctx, &cs, toDBStory(*story)); err != nil {
		errMsg := fmt.Sprintf("failed to create story: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create story"), trace.WithAttributes(attribute.String("error", errMsg)))
		return dbStory{}, err
	}

	r.log.Info(ctx, "Story created successfully.")
	span.AddEvent("Story created.", trace.WithAttributes(
		attribute.String("story.title", story.Title),
	))

	return cs, err
}

func (r *repo) GetStoryLinks(ctx context.Context, storyID uuid.UUID) ([]links.CoreLink, error) {
	r.log.Info(ctx, "business.repository.stories.GetStoryLinks")
	ctx, span := web.AddSpan(ctx, "business.repository.stories.GetStoryLinks")
	defer span.End()

	span.SetAttributes(attribute.String("storyId", storyID.String()))

	var dbLinks []linksrepo.DbLink
	query := `
		SELECT 
			link_id,
			title,
			url,
			story_id,
			created_at,
			updated_at
		FROM story_links
		WHERE story_id = :story_id
	`

	params := map[string]any{
		"story_id": storyID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, "error preparing statement", err)
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &dbLinks, params); err != nil {
		r.log.Error(ctx, "error selecting links", err)
		return nil, err
	}

	return linksrepo.ToCoreLinks(dbLinks), nil
}

func (r *repo) UpdateLabels(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, labels []uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.UpdateLabels")
	defer span.End()

	// Start a transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// First, delete all existing labels for the story
	deleteQuery := `
		DELETE FROM story_labels 
		WHERE story_id = :story_id 
		AND story_id IN (
			SELECT id FROM stories 
			WHERE id = :story_id 
			AND workspace_id = :workspace_id
		)
	`
	params := map[string]any{
		"story_id":     id,
		"workspace_id": workspaceId,
	}

	if _, err = tx.NamedExecContext(ctx, deleteQuery, params); err != nil {
		return fmt.Errorf("failed to delete existing labels: %w", err)
	}

	// If we have new labels to insert
	if len(labels) > 0 {
		// Prepare values for bulk insert
		values := make([]string, len(labels))
		args := make([]any, 0, len(labels)*2)
		for i, labelID := range labels {
			values[i] = fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2)
			args = append(args, id, labelID)
		}

		// Insert new labels
		insertQuery := fmt.Sprintf(`
			INSERT INTO story_labels (story_id, label_id)
			VALUES %s
		`, strings.Join(values, ","))

		if _, err = tx.ExecContext(ctx, insertQuery, args...); err != nil {
			return fmt.Errorf("failed to insert new labels: %w", err)
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// List returns a list of stories for a workspace with additional filters.

// MyStories returns a list of stories.
func (r *repo) MyStories(ctx context.Context, workspaceId uuid.UUID) ([]stories.CoreStoryList, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.MyStories")
	defer span.End()

	currentUser, _ := mid.GetUserID(ctx)

	params := map[string]any{
		"workspace_id": workspaceId,
		"current_user": currentUser,
	}

	var stories []dbStory

	q := `
		SELECT
			s.id,
			s.sequence_id,
			s.title,
			s.priority,
			s.status_id,
			s.start_date,
			s.end_date,
			s.sprint_id,
			s.team_id,
			s.objective_id,
			s.workspace_id,
			s.assignee_id,
			s.reporter_id,
			s.created_at,
			s.updated_at,
			s.deleted_at,
			s.archived_at,
			COALESCE(
				(
					SELECT
						json_agg(
							json_build_object(
								'id', sub.id,
								'sequence_id', sub.sequence_id,
								'title', sub.title,
								'priority', sub.priority,
								'status_id', sub.status_id,
								'start_date', sub.start_date,
								'end_date', sub.end_date,
								'sprint_id', sub.sprint_id,
								'team_id', sub.team_id,
								'objective_id', sub.objective_id,
								'workspace_id', sub.workspace_id,
								'assignee_id', sub.assignee_id,
								'reporter_id', sub.reporter_id,
								'created_at', sub.created_at,
								'updated_at', sub.updated_at,
								'completed_at', sub.completed_at,
								'deleted_at', sub.deleted_at,
								'archived_at', sub.archived_at,
								'labels', '[]'
							)
						)
					FROM
						stories sub
					WHERE
						sub.parent_id = s.id
						AND sub.deleted_at IS NULL
				), '[]'
			) AS sub_stories,
			COALESCE(
				(
					SELECT
						json_agg(l.label_id)
					FROM
						labels l
						INNER JOIN story_labels sl ON sl.label_id = l.label_id
					WHERE
						sl.story_id = s.id
				), '[]'
			) AS labels
		FROM
			stories s
			INNER JOIN team_members tm ON 
				tm.team_id = s.team_id 
				AND tm.user_id = :current_user
		WHERE s.workspace_id = :workspace_id 
		AND s.deleted_at IS NULL
		AND s.parent_id IS NULL
		AND (s.assignee_id = :current_user OR s.reporter_id = :current_user)
		ORDER BY s.created_at DESC;
	`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "fetching stories.")
	if err := stmt.SelectContext(ctx, &stories, params); err != nil {
		errMsg := fmt.Sprintf("failed to retrieve stories from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("stories not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	r.log.Info(ctx, "stories retrieved successfully.")
	span.AddEvent("stories retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(stories)),
		attribute.String("query", q),
	))

	return toCoreStories(stories), nil
}

// Get returns the story with the specified ID.
func (r *repo) Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (stories.CoreSingleStory, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.Get")
	defer span.End()

	story, err := r.getStoryById(ctx, id, workspaceId)
	if err != nil {
		return stories.CoreSingleStory{}, err
	}

	return toCoreStory(story), nil
}

func (r *repo) getStoryById(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (dbStory, error) {
	query := `
        SELECT
					s.id,
					s.title,
					s.priority,
					s.sequence_id,
					s.status_id,
					s.description,
					s.description_html,
					s.team_id,
					s.objective_id,
					s.sprint_id,
					s.key_result_id,
					s.workspace_id,
					s.assignee_id,
					s.reporter_id,
					s.start_date,
					s.end_date,
					s.created_at,
					s.updated_at,
					s.deleted_at,
					s.archived_at,
					s.completed_at,
					COALESCE(
							(
									SELECT
											json_agg(sub.*)
									FROM
											stories sub
									WHERE
											sub.parent_id = s.id AND sub.deleted_at IS NULL
							), '[]'
					) AS sub_stories,
					COALESCE(
						(
								SELECT
										json_agg(l.label_id)
								FROM
										labels l
										INNER JOIN story_labels sl ON sl.label_id = l.label_id
								WHERE
										sl.story_id = s.id
						), '[]') AS labels
				FROM
					stories s
				WHERE
					s.id =:id
					AND s.workspace_id =:workspace_id;
    `

	params := map[string]any{
		"id":           id,
		"workspace_id": workspaceId,
	}

	var story dbStory

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to prepare named statement: %s", err), "id", id)
		return dbStory{}, err
	}
	defer stmt.Close()

	err = stmt.GetContext(ctx, &story, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return dbStory{}, errors.New("story not found")
		}
		r.log.Error(ctx, fmt.Sprintf("failed to execute query: %s", err), "id", id)
		return dbStory{}, err
	}

	return story, nil
}

// Delete deletes the story with the specified ID.
func (r *repo) Delete(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.Delete")
	defer span.End()
	params := map[string]any{"id": id, "workspace_id": workspaceId}

	stmt, err := r.db.PrepareNamedContext(ctx, `
		UPDATE stories 
		SET deleted_at = NOW(),
				updated_at = NOW() 
		WHERE id = :id
		AND workspace_id = :workspace_id;
	`)

	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to prepare named statement: %s", err), "id", id)
		return err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("Deleting story #%s", id), "id", id)
	if _, err := stmt.ExecContext(ctx, params); err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to delete story: %s", err), "id", id)
		return err
	}

	r.log.Info(ctx, fmt.Sprintf("Story #%s deleted successfully", id), "id", id)
	span.AddEvent("Story deleted.", trace.WithAttributes(attribute.String("story.id", id.String())))

	return nil
}

// List returns a list of stories for a workspace with additional filters.
func (r *repo) List(ctx context.Context, workspaceId uuid.UUID, filters map[string]any) ([]stories.CoreStoryList, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.List")
	defer span.End()

	// Legacy method - for backwards compatibility, convert map filters to CoreStoryFilters if needed
	// If it's a simple filter map (old style), use the simple approach
	// If it contains complex filters (new style from coreFiltersToMap), use the advanced approach

	// Check if this is a complex filter (contains date filters or advanced filters)
	hasComplexFilters := false
	for key := range filters {
		if key == "created_after" || key == "created_before" || key == "updated_after" ||
			key == "updated_before" || key == "deadline_after" || key == "deadline_before" ||
			key == "assigned_to_me" || key == "created_by_me" || key == "has_no_assignee" ||
			key == "current_user_id" {
			hasComplexFilters = true
			break
		}
	}

	if hasComplexFilters {
		// Use the advanced filtering approach - convert back to CoreStoryFilters
		coreFilters := mapToCoreFilters(filters, workspaceId)
		query := r.buildSimpleStoriesQuery(coreFilters)
		query += " ORDER BY s.created_at DESC"

		params := r.buildQueryParams(coreFilters)

		var stories []dbStory
		stmt, err := r.db.PrepareNamedContext(ctx, query)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
			r.log.Error(ctx, errMsg)
			span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
			return nil, err
		}
		defer stmt.Close()

		if err := stmt.SelectContext(ctx, &stories, params); err != nil {
			errMsg := fmt.Sprintf("Failed to retrieve stories from the database: %s", err)
			r.log.Error(ctx, errMsg)
			span.RecordError(errors.New("stories not found"), trace.WithAttributes(attribute.String("error", errMsg)))
			return nil, err
		}

		span.AddEvent("stories retrieved.", trace.WithAttributes(
			attribute.Int("story.count", len(stories)),
		))

		return toCoreStories(stories), nil
	}

	// Legacy simple filtering approach for backwards compatibility
	query := `
		SELECT
			s.id,
			s.sequence_id,
			s.title,
			s.priority,
			s.status_id,
			s.start_date,
			s.end_date,
			s.sprint_id,
			s.team_id,
			s.objective_id,
			s.workspace_id,
			s.assignee_id,
			s.reporter_id,
			s.created_at,
			s.updated_at,
			s.completed_at,
			s.deleted_at,
			s.archived_at,
			COALESCE(
				(
					SELECT
						json_agg(
							json_build_object(
								'id', sub.id,
								'sequence_id', sub.sequence_id,
								'title', sub.title,
								'priority', sub.priority,
								'status_id', sub.status_id,
								'start_date', sub.start_date,
								'end_date', sub.end_date,
								'sprint_id', sub.sprint_id,
								'team_id', sub.team_id,
								'objective_id', sub.objective_id,
								'workspace_id', sub.workspace_id,
								'assignee_id', sub.assignee_id,
								'reporter_id', sub.reporter_id,
								'created_at', sub.created_at,
								'updated_at', sub.updated_at,
								'completed_at', sub.completed_at,
								'deleted_at', sub.deleted_at,
								'archived_at', sub.archived_at,
								'labels', '[]'
							)
						)
					FROM
						stories sub
					WHERE
						sub.parent_id = s.id 
						AND sub.deleted_at IS NULL
				), '[]'
			) AS sub_stories,
			COALESCE(
				(
					SELECT
						json_agg(l.label_id)
					FROM
						labels l
						INNER JOIN story_labels sl ON sl.label_id = l.label_id
					WHERE
						sl.story_id = s.id
				), '[]'
			) AS labels
		FROM
			stories s
	`
	var setClauses []string

	for field := range filters {
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", field, field))
	}

	filters["workspace_id"] = workspaceId

	query += " WHERE " + strings.Join(setClauses, " AND ") + " AND deleted_at IS NULL AND workspace_id = :workspace_id AND parent_id IS NULL;"

	var stories []dbStory

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	r.log.Info(ctx, "fetching stories.")
	if err := stmt.SelectContext(ctx, &stories, filters); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve stories from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("stories not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	r.log.Info(ctx, "stories retrieved successfully.")
	span.AddEvent("stories retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(stories)),
		attribute.String("query", query),
	))

	return toCoreStories(stories), nil
}

// mapToCoreFilters converts map[string]any filters back to CoreStoryFilters
func mapToCoreFilters(filters map[string]any, workspaceId uuid.UUID) stories.CoreStoryFilters {
	coreFilters := stories.CoreStoryFilters{
		WorkspaceID: workspaceId,
	}

	// Handle CurrentUserID
	if currentUserId, ok := filters["current_user_id"].(uuid.UUID); ok {
		coreFilters.CurrentUserID = currentUserId
	}

	if statusIds, ok := filters["status_ids"].([]uuid.UUID); ok {
		coreFilters.StatusIDs = statusIds
	}
	if assigneeIds, ok := filters["assignee_ids"].([]uuid.UUID); ok {
		coreFilters.AssigneeIDs = assigneeIds
	}
	if reporterIds, ok := filters["reporter_ids"].([]uuid.UUID); ok {
		coreFilters.ReporterIDs = reporterIds
	}
	if priorities, ok := filters["priorities"].([]string); ok {
		coreFilters.Priorities = priorities
	}
	if categories, ok := filters["categories"].([]string); ok {
		coreFilters.Categories = categories
	}
	if teamIds, ok := filters["team_ids"].([]uuid.UUID); ok {
		coreFilters.TeamIDs = teamIds
	}
	if sprintIds, ok := filters["sprint_ids"].([]uuid.UUID); ok {
		coreFilters.SprintIDs = sprintIds
	}
	if labelIds, ok := filters["label_ids"].([]uuid.UUID); ok {
		coreFilters.LabelIDs = labelIds
	}
	if parentId, ok := filters["parent_id"].(uuid.UUID); ok {
		coreFilters.Parent = &parentId
	}
	if objectiveId, ok := filters["objective_id"].(uuid.UUID); ok {
		coreFilters.Objective = &objectiveId
	}
	if epicId, ok := filters["epic_id"].(uuid.UUID); ok {
		coreFilters.Epic = &epicId
	}
	if keyResultId, ok := filters["key_result_id"].(uuid.UUID); ok {
		coreFilters.KeyResult = &keyResultId
	}
	if hasNoAssignee, ok := filters["has_no_assignee"].(bool); ok {
		coreFilters.HasNoAssignee = &hasNoAssignee
	}
	if assignedToMe, ok := filters["assigned_to_me"].(bool); ok {
		coreFilters.AssignedToMe = &assignedToMe
	}
	if createdByMe, ok := filters["created_by_me"].(bool); ok {
		coreFilters.CreatedByMe = &createdByMe
	}
	if createdAfter, ok := filters["created_after"].(time.Time); ok {
		coreFilters.CreatedAfter = &createdAfter
	}
	if createdBefore, ok := filters["created_before"].(time.Time); ok {
		coreFilters.CreatedBefore = &createdBefore
	}
	if updatedAfter, ok := filters["updated_after"].(time.Time); ok {
		coreFilters.UpdatedAfter = &updatedAfter
	}
	if updatedBefore, ok := filters["updated_before"].(time.Time); ok {
		coreFilters.UpdatedBefore = &updatedBefore
	}
	if deadlineAfter, ok := filters["deadline_after"].(time.Time); ok {
		coreFilters.DeadlineAfter = &deadlineAfter
	}
	if deadlineBefore, ok := filters["deadline_before"].(time.Time); ok {
		coreFilters.DeadlineBefore = &deadlineBefore
	}
	if completedAfter, ok := filters["completed_after"].(time.Time); ok {
		coreFilters.CompletedAfter = &completedAfter
	}
	if completedBefore, ok := filters["completed_before"].(time.Time); ok {
		coreFilters.CompletedBefore = &completedBefore
	}
	if includeArchived, ok := filters["include_archived"].(bool); ok {
		coreFilters.IncludeArchived = &includeArchived
	}

	return coreFilters
}

// BulkDelete removes the stories with the specified IDs.
func (r *repo) BulkDelete(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.BulkDelete")
	defer span.End()

	params := map[string]any{"ids": ids, "workspace_id": workspaceId}

	query := `
        UPDATE stories
        SET deleted_at = NOW(), updated_at = NOW()
        WHERE id = ANY(:ids) AND workspace_id = :workspace_id;
    `

	r.log.Info(ctx, fmt.Sprintf("Deleting stories: %v", ids), "ids", ids)
	_, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to delete stories: %s", err), "ids", ids)
		return err
	}

	r.log.Info(ctx, fmt.Sprintf("Stories: %v deleted successfully", ids), "ids", ids)
	span.AddEvent("Stories deleted.", trace.WithAttributes(attribute.Int("stories.length", len(ids))))

	return nil
}

// HardBulkDelete performs permanent removal of the stories with the specified IDs.
func (r *repo) HardBulkDelete(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.HardBulkDelete")
	defer span.End()

	params := map[string]any{"ids": ids, "workspace_id": workspaceId}
	query := `
		DELETE FROM stories
		WHERE id = ANY(:ids)
			AND workspace_id = :workspace_id;
	`

	r.log.Info(ctx, fmt.Sprintf("Hard deleting stories: %v", ids), "ids", ids)

	result, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to hard delete stories: %s", err)
		r.log.Error(ctx, errMsg, "ids", ids)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to get rows affected: %s", err), "ids", ids)
		return err
	}
	if rowsAffected == 0 {
		r.log.Warn(ctx, "No stories found to hard delete", "ids", ids)
		return fmt.Errorf("no stories found to delete")
	}

	r.log.Info(ctx, fmt.Sprintf("Stories hard deleted successfully: %v (%d rows)", ids, rowsAffected),
		"ids", ids, "rows_affected", rowsAffected)
	span.AddEvent("Stories hard deleted.", trace.WithAttributes(
		attribute.Int("stories.length", len(ids)),
		attribute.Int64("rows.affected", rowsAffected)))

	return nil
}

// Restore restores a story with the specified ID.
func (r *repo) Restore(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.Restore")
	defer span.End()

	query := `
			UPDATE stories 
			SET deleted_at = NULL, 
					updated_at = NOW() 
			WHERE id = :id
			AND workspace_id = :workspace_id;
	`
	params := map[string]any{"id": id, "workspace_id": workspaceId}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to prepare restore statement: %s", err), "id", id)
		return err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("restoring story #%s", id), "id", id)
	_, err = stmt.ExecContext(ctx, params)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to restore story: %s", err), "id", id)
		return err
	}

	r.log.Info(ctx, fmt.Sprintf("story #%s restored successfully", id), "id", id)
	span.AddEvent("story restored.", trace.WithAttributes(attribute.String("story.id", id.String())))

	return nil
}

// BulkRestore restores the stories with the specified IDs.
func (r *repo) BulkRestore(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.BulkRestore")
	defer span.End()

	params := map[string]any{"ids": ids, "workspace_id": workspaceId}

	query := `
				UPDATE stories
				SET deleted_at = NULL, updated_at = NOW()
				WHERE id = ANY(:ids)
				AND workspace_id = :workspace_id;
			`

	r.log.Info(ctx, fmt.Sprintf("restoring stories: %v", ids), "ids", ids)
	_, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to restore stories: %s", err), "ids", ids)
		return err
	}

	r.log.Info(ctx, fmt.Sprintf("stories: %v restored successfully", ids), "ids", ids)
	span.AddEvent("stories restored.", trace.WithAttributes(attribute.Int("stories.length", len(ids))))

	return nil
}

// BulkUnarchive unarchives the stories with the specified IDs.
func (r *repo) BulkUnarchive(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.BulkUnarchive")
	defer span.End()

	params := map[string]any{"ids": ids, "workspace_id": workspaceId}

	query := `
		UPDATE stories
		SET archived_at = NULL, updated_at = NOW()
		WHERE id = ANY(:ids)
			AND workspace_id = :workspace_id
			AND archived_at IS NOT NULL;
	`

	r.log.Info(ctx, fmt.Sprintf("Bulk unarchiving stories: %v", ids), "ids", ids)

	result, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to bulk unarchive stories: %s", err)
		r.log.Error(ctx, errMsg, "ids", ids)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to get rows affected: %s", err), "ids", ids)
		return err
	}
	if rowsAffected == 0 {
		r.log.Warn(ctx, "No stories found to unarchive", "ids", ids)
		return fmt.Errorf("no stories found to unarchive")
	}

	r.log.Info(ctx, fmt.Sprintf("Stories unarchived successfully: %v (%d rows)", ids, rowsAffected),
		"ids", ids, "rows_affected", rowsAffected)
	span.AddEvent("Stories unarchived.", trace.WithAttributes(
		attribute.Int("stories.length", len(ids)),
		attribute.Int64("rows.affected", rowsAffected)))

	return nil
}

// BulkArchive archives the stories with the specified IDs.
func (r *repo) BulkArchive(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.BulkArchive")
	defer span.End()

	params := map[string]any{"ids": ids, "workspace_id": workspaceId}

	query := `
		UPDATE stories
		SET archived_at = NOW()
		WHERE id = ANY(:ids)
			AND workspace_id = :workspace_id
			AND archived_at IS NULL;
	`

	r.log.Info(ctx, fmt.Sprintf("Bulk archiving stories: %v", ids), "ids", ids)

	result, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to bulk archive stories: %s", err)
		r.log.Error(ctx, errMsg, "ids", ids)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to get rows affected: %s", err), "ids", ids)
		return err
	}
	if rowsAffected == 0 {
		r.log.Warn(ctx, "No stories found to archive", "ids", ids)
		return fmt.Errorf("no stories found to archive")
	}

	r.log.Info(ctx, fmt.Sprintf("Stories archived successfully: %v (%d rows)", ids, rowsAffected),
		"ids", ids, "rows_affected", rowsAffected)
	span.AddEvent("Stories archived.", trace.WithAttributes(
		attribute.Int("stories.length", len(ids)),
		attribute.Int64("rows.affected", rowsAffected)))

	return nil
}

// Update updates the story with the specified ID.
func (r *repo) Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, updates map[string]any) error {
	r.log.Info(ctx, "business.repository.stories.Update")
	ctx, span := web.AddSpan(ctx, "business.repository.stories.Update")
	defer span.End()

	// If status is being updated, validate it belongs to the same team
	if statusId, ok := updates["status_id"].(uuid.UUID); ok {
		// We need to get the story's team ID first
		var teamId uuid.UUID
		q := `SELECT team_id FROM stories WHERE id = :story_id AND workspace_id = :workspace_id`
		params := map[string]any{
			"story_id":     id,
			"workspace_id": workspaceId,
		}
		stmt, err := r.db.PrepareNamedContext(ctx, q)
		if err != nil {
			errMsg := fmt.Sprintf("failed to prepare team query statement: %s", err)
			r.log.Error(ctx, errMsg)
			span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
			return err
		}
		defer stmt.Close()

		if err := stmt.GetContext(ctx, &teamId, params); err != nil {
			errMsg := fmt.Sprintf("failed to get story team: %s", err)
			r.log.Error(ctx, errMsg)
			span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
			return err
		}

		if err := r.validateStatusTeam(ctx, statusId, teamId); err != nil {
			return err
		}
	}

	query := "UPDATE stories SET "
	var setClauses []string
	params := map[string]any{"id": id, "workspace_id": workspaceId}

	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", field, field))
		params[field] = value
	}

	setClauses = append(setClauses, "updated_at = NOW()")

	query += strings.Join(setClauses, ", ")
	query += " WHERE id = :id AND workspace_id = :workspace_id;"

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to prepare named update statement: %s", err), "id", id)
		return err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("Updating story #%s", id), "id", id)
	_, err = stmt.ExecContext(ctx, params)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to update story: %s", err), "id", id)
		return err
	}

	r.log.Info(ctx, fmt.Sprintf("Story #%s updated successfully", id), "id", id)
	span.AddEvent("Story updated.", trace.WithAttributes(attribute.String("story.id", id.String())))

	return nil
}

// BulkUpdate updates the stories with the specified IDs.
func (r *repo) BulkUpdate(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID, updates map[string]any) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.BulkUpdate")
	defer span.End()

	query := "UPDATE stories SET "
	var setClauses []string
	params := map[string]any{"ids": ids, "workspace_id": workspaceId}

	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = :%s", field, field))
		params[field] = value
	}

	setClauses = append(setClauses, "updated_at = NOW()")

	query += strings.Join(setClauses, ", ")
	query += " WHERE id IN (:ids) AND workspace_id = :workspace_id;"

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to prepare named bulk update statement: %s", err), "ids", ids)
		return err
	}
	defer stmt.Close()

	r.log.Info(ctx, fmt.Sprintf("Updating stories: %v", ids), "ids", ids)
	if _, err := stmt.ExecContext(ctx, params); err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to update stories: %s", err), "ids", ids)
		return err
	}

	r.log.Info(ctx, fmt.Sprintf("Stories: %v updated successfully", ids), "ids", ids)
	span.AddEvent("Stories updated.", trace.WithAttributes(attribute.Int("stories.length", len(ids))))

	return nil
}

func (r *repo) GetSubStories(ctx context.Context, parentId uuid.UUID, workspaceId uuid.UUID) ([]stories.CoreStoryList, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.GetSubStories")
	defer span.End()

	subStories, err := r.getSubStories(ctx, parentId, workspaceId)
	if err != nil {
		return nil, err
	}

	return toCoreStories(subStories), nil
}

func (r *repo) getSubStories(ctx context.Context, parentId uuid.UUID, workspaceId uuid.UUID) ([]dbStory, error) {
	query := `
        SELECT
          id,
					sequence_id,
					title,
					priority,
					description,
					status_id,
					start_date,
					end_date,
					sprint_id,
					team_id,
					objective_id,
					workspace_id,
					assignee_id,
					reporter_id,
					created_at,
					updated_at,
					completed_at,
					deleted_at,
					archived_at
        FROM
            stories
        WHERE
            parent_id = :parent_id
					AND workspace_id = :workspace_id
					AND deleted_at IS NULL
        ORDER BY sequence_id ASC;
    `
	params := map[string]any{
		"parent_id":    parentId,
		"workspace_id": workspaceId,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("Failed to prepare named statement: %s", err), "id", parentId)
		return nil, err
	}
	defer stmt.Close()

	var subStories []dbStory
	if err := stmt.SelectContext(ctx, &subStories, params); err != nil {
		return nil, err
	}

	return subStories, nil
}

func (r *repo) RecordActivities(ctx context.Context, activities []stories.CoreActivity) ([]stories.CoreActivity, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.RecordActivities")
	defer span.End()

	dbActivities, err := r.recordActivities(ctx, activities)
	if err != nil {
		return nil, fmt.Errorf("failed to insert activities: %w", err)
	}

	return toCoreActivities(dbActivities), nil
}

func (r *repo) recordActivities(ctx context.Context, activities []stories.CoreActivity) ([]dbActivity, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.recordActivity")
	defer span.End()

	// Prepare the base query for bulk insert
	q := `
		INSERT INTO story_activities (
			story_id, 
			activity_type, 
			field_changed, 
			current_value,
			user_id,
			workspace_id
		)
		VALUES (
			:story_id, 
			:activity_type, 
			:field_changed, 
			:current_value,
			:user_id,
			:workspace_id
		)
		RETURNING story_activities.*;
	`

	// Start a transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Prepare the statement
	stmt, err := tx.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	// Insert each activity and collect results
	var result []dbActivity
	for _, activity := range activities {
		var da dbActivity
		if err := stmt.GetContext(ctx, &da, toDBActivity(activity)); err != nil {
			errMsg := fmt.Sprintf("Failed to insert activity: %s", err)
			r.log.Error(ctx, errMsg)
			span.RecordError(errors.New("failed to insert activity"), trace.WithAttributes(attribute.String("error", errMsg)))
			return nil, err
		}
		result = append(result, da)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.log.Info(ctx, fmt.Sprintf("Successfully created %d activities", len(activities)))
	span.AddEvent("Activities created.", trace.WithAttributes(
		attribute.Int("activity.count", len(activities)),
	))

	return result, nil
}

// GetActivities returns activities for a given story ID with pagination.
func (r *repo) GetActivities(ctx context.Context, storyID uuid.UUID, page, pageSize int) ([]stories.CoreActivity, bool, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.GetActivities")
	defer span.End()

	// Calculate offset and fetch one extra to check if there are more
	offset := (page - 1) * pageSize
	limit := pageSize + 1

	params := map[string]any{
		"story_id": storyID,
		"limit":    limit,
		"offset":   offset,
	}

	q := `
		SELECT 
			activity_id,
			story_id,
			user_id,
			activity_type,
			field_changed,
			current_value,
			created_at,
			workspace_id
		FROM story_activities
		WHERE story_id = :story_id
		ORDER BY created_at DESC
		LIMIT :limit OFFSET :offset
	`

	var activities []dbActivity

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, false, fmt.Errorf("failed to get activities: %w", err)
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &activities, params); err != nil {
		errMsg := fmt.Sprintf("Failed to get activities: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get activities"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, false, fmt.Errorf("failed to get activities: %w", err)
	}

	// Check if there are more activities
	hasMore := len(activities) > pageSize
	if hasMore {
		// Remove the extra activity we fetched
		activities = activities[:pageSize]
	}

	span.AddEvent("activities retrieved.", trace.WithAttributes(
		attribute.Int("activity.count", len(activities)),
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
		attribute.Bool("has.more", hasMore),
	))

	return toCoreActivities(activities), hasMore, nil
}

func (r *repo) CreateComment(ctx context.Context, cnc stories.CoreNewComment) (comments.CoreComment, error) {
	r.log.Info(ctx, "business.repository.stories.CreateComment")
	ctx, span := web.AddSpan(ctx, "business.repository.stories.CreateComment")
	defer span.End()

	q := `
	INSERT INTO story_comments (
		content, story_id, commenter_id, parent_id
	) VALUES (
		:content, :story_id, :commenter_id, :parent_id
	) RETURNING story_comments.*;
`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to prepare named statement: %s", err))
		return comments.CoreComment{}, err
	}
	defer stmt.Close()

	var comment commentsrepo.DbComment
	if err := stmt.GetContext(ctx, &comment, toDBNewComment(cnc)); err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to insert comment: %s", err))
		return comments.CoreComment{}, err
	}

	return toCoreComment(comment), nil
}

func (r *repo) GetComments(ctx context.Context, storyID uuid.UUID, page, pageSize int) ([]comments.CoreComment, bool, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.GetComments")
	defer span.End()

	// Calculate offset and fetch one extra to check if there are more
	offset := (page - 1) * pageSize
	limit := pageSize + 1

	params := map[string]any{
		"story_id": storyID,
		"limit":    limit,
		"offset":   offset,
	}

	q := `
		SELECT 
			sc.comment_id,
			sc.story_id,
			sc.commenter_id,
			sc.content,
			sc.parent_id,
			sc.created_at,
			sc.updated_at,
			COALESCE(
				(
					SELECT
						json_agg(sub.*)
					FROM
						story_comments sub
					WHERE
						sub.parent_id = sc.comment_id			
				), '[]'
			) AS sub_comments
		FROM story_comments sc 
		WHERE sc.story_id = :story_id AND sc.parent_id IS NULL 
		ORDER BY sc.created_at DESC
		LIMIT :limit OFFSET :offset
	`

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to prepare named statement: %s", err))
		return nil, false, fmt.Errorf("failed to prepare named statement: %w", err)
	}
	defer stmt.Close()

	var comments []commentsrepo.DbComment
	if err := stmt.SelectContext(ctx, &comments, params); err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to get comments: %s", err))
		return nil, false, fmt.Errorf("failed to get comments: %w", err)
	}

	// Check if there are more comments
	hasMore := len(comments) > pageSize
	if hasMore {
		// Remove the extra comment we fetched
		comments = comments[:pageSize]
	}

	span.AddEvent("comments retrieved.", trace.WithAttributes(
		attribute.Int("comment.count", len(comments)),
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
		attribute.Bool("has.more", hasMore),
	))

	return toCoreComments(comments), hasMore, nil
}

func (r *repo) GetComment(ctx context.Context, commentID uuid.UUID) (comments.CoreComment, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.GetComment")
	defer span.End()

	q := `
		SELECT comment_id, story_id, commenter_id, content, parent_id, created_at, updated_at
		FROM story_comments 
		WHERE comment_id = :comment_id
	`

	params := map[string]any{"comment_id": commentID}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to prepare named statement: %s", err))
		return comments.CoreComment{}, fmt.Errorf("failed to prepare named statement: %w", err)
	}
	defer stmt.Close()

	var comment commentsrepo.DbComment
	if err := stmt.GetContext(ctx, &comment, params); err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to get comment: %s", err))
		return comments.CoreComment{}, fmt.Errorf("failed to get comment: %w", err)
	}

	return toCoreComment(comment), nil
}

func (r *repo) validateStatusTeam(ctx context.Context, statusId, teamId uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.validateStatusTeam")
	defer span.End()

	q := `
		SELECT EXISTS (
			SELECT 1 FROM statuses 
			WHERE status_id = :status_id 
			AND team_id = :team_id
		)
	`
	params := map[string]any{
		"status_id": statusId,
		"team_id":   teamId,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare validation statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	var exists bool
	if err := stmt.GetContext(ctx, &exists, params); err != nil {
		errMsg := fmt.Sprintf("failed to validate status team: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("database error"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	if !exists {
		return errors.New("status does not belong to the story's team")
	}

	return nil
}

// DuplicateStory creates a copy of an existing story with a new sequence ID
func (r *repo) DuplicateStory(ctx context.Context, originalStoryID uuid.UUID, workspaceId uuid.UUID, userID uuid.UUID) (stories.CoreSingleStory, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.DuplicateStory")
	defer span.End()

	// Start a transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return stories.CoreSingleStory{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Get the original story
	originalStory, err := r.getStoryById(ctx, originalStoryID, workspaceId)
	if err != nil {
		return stories.CoreSingleStory{}, fmt.Errorf("failed to get original story: %w", err)
	}

	// Get new sequence ID
	lastSequence, commit, rollback, err := r.GetNextSequenceID(ctx, originalStory.Team, workspaceId)
	if err != nil {
		return stories.CoreSingleStory{}, fmt.Errorf("failed to get next sequence ID: %w", err)
	}
	defer rollback()

	// Prepare the insert query for the new story
	q := `
		INSERT INTO stories (
			sequence_id,
			title,
			description,
			description_html,
			team_id,
			objective_id,
			status_id,
			assignee_id,
			priority,
			sprint_id,
			workspace_id,
			reporter_id,
			created_at,
			updated_at
		) VALUES (
			:sequence_id,
			:title,
			:description,
			:description_html,
			:team_id,
			:objective_id,
			:status_id,
			:assignee_id,
			:priority,
			:sprint_id,
			:workspace_id,
			:reporter_id,
			NOW(),
			NOW()
		) RETURNING stories.id, stories.sequence_id, stories.title, stories.description, stories.description_html, stories.parent_id, stories.objective_id, stories.status_id, stories.assignee_id, stories.blocked_by_id, stories.blocking_id, stories.related_id, stories.reporter_id, stories.priority, stories.sprint_id, stories.team_id, stories.workspace_id, stories.start_date, stories.end_date, stories.created_at, stories.updated_at;
	`

	// Prepare parameters for the new story
	params := map[string]any{
		"sequence_id":      lastSequence + 1,
		"title":            "Copy of " + originalStory.Title,
		"description":      originalStory.Description,
		"description_html": originalStory.DescriptionHTML,
		"team_id":          originalStory.Team,
		"objective_id":     originalStory.Objective,
		"status_id":        originalStory.Status,
		"assignee_id":      originalStory.Assignee,
		"priority":         originalStory.Priority,
		"sprint_id":        originalStory.Sprint,
		"workspace_id":     workspaceId,
		"reporter_id":      userID,
	}

	// Execute the insert
	stmt, err := tx.PrepareNamedContext(ctx, q)
	if err != nil {
		return stories.CoreSingleStory{}, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	var newStory dbStory
	if err := stmt.GetContext(ctx, &newStory, params); err != nil {
		return stories.CoreSingleStory{}, fmt.Errorf("failed to create duplicate story: %w", err)
	}

	// Commit the sequence ID transaction
	if err := commit(); err != nil {
		return stories.CoreSingleStory{}, fmt.Errorf("failed to commit sequence ID transaction: %w", err)
	}

	// Commit the main transaction
	if err := tx.Commit(); err != nil {
		return stories.CoreSingleStory{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.log.Info(ctx, fmt.Sprintf("Successfully duplicated story #%s", originalStoryID))
	span.AddEvent("Story duplicated.", trace.WithAttributes(
		attribute.String("original_story.id", originalStoryID.String()),
		attribute.String("new_story.id", newStory.ID.String()),
	))

	return toCoreStory(newStory), nil
}

// CountStoriesInWorkspace returns the count of stories in a workspace.
func (r *repo) CountStoriesInWorkspace(ctx context.Context, workspaceId uuid.UUID) (int, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.CountStoriesInWorkspace")
	defer span.End()

	query := `
		SELECT COUNT(*)
		FROM stories
		WHERE workspace_id = :workspace_id
		AND deleted_at IS NULL AND archived_at IS NULL
	`

	params := map[string]any{
		"workspace_id": workspaceId,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare count stories statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return 0, err
	}
	defer stmt.Close()

	var count int
	if err := stmt.GetContext(ctx, &count, params); err != nil {
		errMsg := fmt.Sprintf("failed to count stories in workspace: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to count stories"), trace.WithAttributes(attribute.String("error", errMsg)))
		return 0, err
	}

	span.AddEvent("stories counted.", trace.WithAttributes(
		attribute.Int("stories.count", count),
	))

	return count, nil
}

// ListGroupedStories returns stories grouped by the specified field with limited stories per group
func (r *repo) ListGroupedStories(ctx context.Context, query stories.CoreStoryQuery) ([]stories.CoreStoryGroup, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.ListGroupedStories")
	defer span.End()

	// Handle "none" groupBy case - create a single group with all stories
	if query.GroupBy == "none" {
		return r.listGroupedStoriesNone(ctx, query)
	}

	// Use SQL-based grouping for better performance with large datasets
	if r.shouldUseSQLGrouping(query) {
		return r.listGroupedStoriesSQL(ctx, query)
	}

	// Fall back to Go-based grouping for complex cases
	return r.listGroupedStoriesGo(ctx, query)
}

// shouldUseSQLGrouping determines whether to use SQL-based or Go-based grouping
func (r *repo) shouldUseSQLGrouping(query stories.CoreStoryQuery) bool {
	// Always use optimized SQL grouping for better performance
	switch query.GroupBy {
	case "status", "assignee", "priority", "team", "sprint":
		return true
	default:
		return false
	}
}

// listGroupedStoriesSQL performs grouping directly in SQL for optimal performance
func (r *repo) listGroupedStoriesSQL(ctx context.Context, query stories.CoreStoryQuery) ([]stories.CoreStoryGroup, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.listGroupedStoriesSQL")
	defer span.End()

	// Build optimized query that gets all possible groups and their stories
	limit := query.StoriesPerGroup + 1 // Fetch one extra to check if more exist

	// Build the all groups CTE based on group type
	allGroupsCTE := r.buildAllGroupsCTE(query.GroupBy, query.Filters)
	groupColumn := r.getGroupColumn(query.GroupBy)
	orderByClause := r.buildOrderByClause(query.OrderBy, query.OrderDirection)

	// For simple ordering (created, updated), we can use a more efficient approach
	var rowNumberOrder string
	var jsonAggOrder string

	if query.OrderBy == "created" || query.OrderBy == "updated" {
		// Use simple column ordering for better performance
		if query.OrderBy == "created" {
			if query.OrderDirection == "asc" {
				rowNumberOrder = "s.created_at ASC"
				jsonAggOrder = "ls.created_at ASC"
			} else {
				rowNumberOrder = "s.created_at DESC"
				jsonAggOrder = "ls.created_at DESC"
			}
		} else {
			if query.OrderDirection == "asc" {
				rowNumberOrder = "s.updated_at ASC"
				jsonAggOrder = "ls.updated_at ASC"
			} else {
				rowNumberOrder = "s.updated_at DESC"
				jsonAggOrder = "ls.updated_at DESC"
			}
		}
	} else {
		// Use complex ordering for priority and deadline
		rowNumberOrder = orderByClause
		jsonAggOrder = r.buildOrderByClauseWithAlias(query.OrderBy, query.OrderDirection, "ls")
	}

	// Build the join clauses for categories filter
	var joinClauses string
	if len(query.Filters.Categories) > 0 {
		joinClauses = "INNER JOIN statuses stat ON s.status_id = stat.status_id"
	}

	// Simplified query that includes all possible groups
	sqlQuery := fmt.Sprintf(`
		WITH all_groups AS (
			%s
		),
		grouped_stories AS (
			SELECT 
				s.id,
				s.sequence_id,
				s.title,
				s.priority,
				s.status_id,
				s.start_date,
				s.end_date,
				s.sprint_id,
				s.team_id,
				s.objective_id,
				s.workspace_id,
				s.assignee_id,
				s.reporter_id,
				s.created_at,
				s.updated_at,
				s.completed_at,
				s.deleted_at,
				s.archived_at,
				COALESCE(CAST(%s AS text), 'null') as group_key,
				COUNT(*) OVER (PARTITION BY COALESCE(CAST(%s AS text), 'null')) as total_count,
				ROW_NUMBER() OVER (PARTITION BY COALESCE(CAST(%s AS text), 'null') ORDER BY %s) as row_num
			FROM stories s
			%s
			%s
		),
		limited_stories AS (
			SELECT *
			FROM grouped_stories
			WHERE row_num <= %d
		)
		SELECT 
			ag.group_key,
			MAX(ls.total_count) as total_count,
			COALESCE(
				json_agg(
					json_build_object(
						'id', ls.id,
						'sequence_id', ls.sequence_id,
						'title', ls.title,
						'priority', ls.priority,
						'status_id', ls.status_id,
						'start_date', ls.start_date,
						'end_date', ls.end_date,
						'sprint_id', ls.sprint_id,
						'team_id', ls.team_id,
						'objective_id', ls.objective_id,
						'workspace_id', ls.workspace_id,
						'assignee_id', ls.assignee_id,
						'reporter_id', ls.reporter_id,
						'created_at', ls.created_at,
						'updated_at', ls.updated_at,
						'completed_at', ls.completed_at,
						'deleted_at', ls.deleted_at,
						'archived_at', ls.archived_at,
						'sub_stories', COALESCE(
							(
								SELECT
									json_agg(
										json_build_object(
											'id', sub.id,
											'sequence_id', sub.sequence_id,
											'title', sub.title,
											'priority', sub.priority,
											'status_id', sub.status_id,
											'start_date', sub.start_date,
											'end_date', sub.end_date,
											'sprint_id', sub.sprint_id,
											'team_id', sub.team_id,
											'objective_id', sub.objective_id,
											'workspace_id', sub.workspace_id,
											'assignee_id', sub.assignee_id,
											'reporter_id', sub.reporter_id,
											'created_at', sub.created_at,
											'updated_at', sub.updated_at,
											'completed_at', sub.completed_at,
											'deleted_at', sub.deleted_at,
											'archived_at', sub.archived_at,
											'labels', '[]'
										)
									)
								FROM
									stories sub
								WHERE
									sub.parent_id = ls.id 
									AND sub.deleted_at IS NULL
							), '[]'
						),
						'labels', COALESCE(
							(
								SELECT
									json_agg(l.label_id)
								FROM
									labels l
									INNER JOIN story_labels sl ON sl.label_id = l.label_id
								WHERE
									sl.story_id = ls.id
							), '[]'
						)
					) ORDER BY %s
				) FILTER (WHERE ls.id IS NOT NULL), 
				CAST('[]' AS json)
			) as stories_json
		FROM all_groups ag
		LEFT JOIN limited_stories ls ON ag.group_key = ls.group_key
		GROUP BY ag.group_key, ag.sort_order
		ORDER BY ag.sort_order, ag.group_key`,
		allGroupsCTE,
		groupColumn, groupColumn, groupColumn, rowNumberOrder,
		joinClauses,
		r.buildSimpleWhereClause(query.Filters),
		limit, jsonAggOrder)

	params := r.buildQueryParams(query.Filters)

	type groupResult struct {
		GroupKey    string          `db:"group_key"`
		TotalCount  *int            `db:"total_count"`
		StoriesJSON json.RawMessage `db:"stories_json"`
	}

	var results []groupResult
	stmt, err := r.db.PrepareNamedContext(ctx, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare group query: %w", err)
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &results, params); err != nil {
		return nil, fmt.Errorf("failed to execute group query: %w", err)
	}

	// Convert results to CoreStoryGroup
	var groups []stories.CoreStoryGroup
	for _, result := range results {
		var storyMaps []map[string]any
		if err := json.Unmarshal(result.StoriesJSON, &storyMaps); err != nil {
			r.log.Error(ctx, "failed to unmarshal stories JSON", "error", err)
			continue
		}

		// Convert story maps to CoreStoryList
		var coreStories []stories.CoreStoryList
		for _, storyMap := range storyMaps {
			coreStory := r.mapToStoryList(storyMap)
			coreStories = append(coreStories, coreStory)
		}

		// Check if we have more stories (we fetched storiesPerGroup+1)
		hasMore := len(coreStories) > query.StoriesPerGroup
		if hasMore {
			// Remove the extra story we fetched
			coreStories = coreStories[:query.StoriesPerGroup]
		}

		loadedCount := len(coreStories)
		totalCount := 0
		if result.TotalCount != nil {
			totalCount = *result.TotalCount
		}

		group := stories.CoreStoryGroup{
			Key:         result.GroupKey,
			LoadedCount: loadedCount,
			TotalCount:  totalCount,
			HasMore:     hasMore,
			Stories:     coreStories,
			NextPage:    2, // Next page for load more
		}
		groups = append(groups, group)
	}

	span.AddEvent("SQL grouped stories retrieved.", trace.WithAttributes(
		attribute.Int("groups.count", len(groups)),
		attribute.String("group.by", query.GroupBy),
	))

	return groups, nil
}

// listGroupedStoriesNone handles the special case of groupBy=none - creates a single group with all stories
func (r *repo) listGroupedStoriesNone(ctx context.Context, query stories.CoreStoryQuery) ([]stories.CoreStoryGroup, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.listGroupedStoriesNone")
	defer span.End()

	// Use StoriesPerGroup as page size for pagination
	pageSize := query.StoriesPerGroup
	if pageSize == 0 {
		pageSize = 15 // Default page size
	}

	// Calculate offset based on page (page 1 = offset 0)
	page := query.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	// Fetch one extra story to check if there are more
	limit := pageSize + 1

	// Build query with pagination and total count using window function
	baseQuery := r.buildSimpleStoriesQuery(query.Filters)

	// Modify the SELECT clause to include the window function
	modifiedQuery := strings.Replace(baseQuery,
		"CAST('[]' AS json) AS labels",
		"CAST('[]' AS json) AS labels, COUNT(*) OVER() as total_count", 1)

	orderByClause := r.buildOrderByClause(query.OrderBy, query.OrderDirection)
	sqlQuery := fmt.Sprintf("%s ORDER BY %s LIMIT %d OFFSET %d", modifiedQuery, orderByClause, limit, offset)

	params := r.buildQueryParams(query.Filters)

	// Struct to capture both story data and total count
	type storyWithCount struct {
		dbStory
		TotalCount int `db:"total_count"`
	}

	var storiesWithCount []storyWithCount
	stmt, err := r.db.PrepareNamedContext(ctx, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare none group query: %w", err)
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &storiesWithCount, params); err != nil {
		return nil, fmt.Errorf("failed to execute none group query: %w", err)
	}

	// Extract dbStories and total count
	dbStories := make([]dbStory, len(storiesWithCount))
	var totalCount int
	for i, swc := range storiesWithCount {
		dbStories[i] = swc.dbStory
		if i == 0 {
			totalCount = swc.TotalCount // All rows will have the same total count
		}
	}

	// Handle empty results case
	if len(storiesWithCount) == 0 {
		totalCount = 0
	}

	// Check if we have more stories
	hasMore := len(dbStories) > pageSize
	if hasMore {
		// Remove the extra story we fetched
		dbStories = dbStories[:pageSize]
	}

	// Convert to CoreStoryList
	coreStories := toCoreStories(dbStories)

	// Calculate next page
	nextPage := page + 1
	if !hasMore {
		nextPage = 0 // No more pages
	}

	// Create single group with key "none"
	group := stories.CoreStoryGroup{
		Key:         "none",
		LoadedCount: len(coreStories),
		TotalCount:  totalCount,
		HasMore:     hasMore,
		Stories:     coreStories,
		NextPage:    nextPage,
	}

	span.AddEvent("none grouped stories retrieved.", trace.WithAttributes(
		attribute.Int("stories.count", len(coreStories)),
		attribute.Int("page", page),
		attribute.Bool("has.more", hasMore),
	))

	return []stories.CoreStoryGroup{group}, nil
}

// buildSimpleWhereClause builds a simplified WHERE clause without subqueries
func (r *repo) buildSimpleWhereClause(filters stories.CoreStoryFilters) string {
	whereClauses := []string{
		"s.workspace_id = :workspace_id",
		"s.parent_id IS NULL",
	}

	// Archived filtering: If includeArchived=true return ONLY archived stories,
	// otherwise return only active (non-archived) stories.
	if filters.IncludeArchived != nil && *filters.IncludeArchived {
		whereClauses = append(whereClauses, "s.archived_at IS NOT NULL")
	} else {
		whereClauses = append(whereClauses, "s.archived_at IS NULL")
	}

	// Exclude deleted unless explicitly included
	if filters.IncludeDeleted != nil && *filters.IncludeDeleted {
		whereClauses = append(whereClauses, "s.deleted_at IS NOT NULL")
	} else {
		whereClauses = append(whereClauses, "s.deleted_at IS NULL")
	}

	// Handle parent filtering
	if filters.Parent != nil {
		whereClauses = append(whereClauses, "s.parent_id = :parent_id")
	} else {
		// Default: only show top-level stories (no parent)
		whereClauses = append(whereClauses, "s.parent_id IS NULL")
	}

	// Add filter conditions
	if len(filters.StatusIDs) > 0 {
		whereClauses = append(whereClauses, "s.status_id = ANY(:status_ids)")
	}

	if len(filters.AssigneeIDs) > 0 {
		whereClauses = append(whereClauses, "s.assignee_id = ANY(:assignee_ids)")
	}

	if len(filters.ReporterIDs) > 0 {
		whereClauses = append(whereClauses, "s.reporter_id = ANY(:reporter_ids)")
	}

	if len(filters.Priorities) > 0 {
		whereClauses = append(whereClauses, "s.priority = ANY(:priorities)")
	}

	if len(filters.Categories) > 0 {
		whereClauses = append(whereClauses, "stat.category = ANY(:categories)")
	}

	if len(filters.TeamIDs) > 0 {
		whereClauses = append(whereClauses, "s.team_id = ANY(:team_ids)")
	} else {
		// If no specific teams are provided, only show stories from teams user is a member of
		whereClauses = append(whereClauses, "EXISTS (SELECT 1 FROM team_members tm WHERE tm.team_id = s.team_id AND tm.user_id = :current_user_id)")
	}

	if len(filters.SprintIDs) > 0 {
		whereClauses = append(whereClauses, "s.sprint_id = ANY(:sprint_ids)")
	}

	if filters.Objective != nil {
		whereClauses = append(whereClauses, "s.objective_id = :objective_id")
	}

	if filters.HasNoAssignee != nil && *filters.HasNoAssignee {
		whereClauses = append(whereClauses, "s.assignee_id IS NULL")
	}

	// Handle createdByMe and assignedToMe with OR logic when both are true
	if filters.AssignedToMe != nil && *filters.AssignedToMe && filters.CreatedByMe != nil && *filters.CreatedByMe {
		whereClauses = append(whereClauses, "(s.assignee_id = :current_user_id OR s.reporter_id = :current_user_id)")
	} else {
		if filters.AssignedToMe != nil && *filters.AssignedToMe {
			whereClauses = append(whereClauses, "s.assignee_id = :current_user_id")
		}

		if filters.CreatedByMe != nil && *filters.CreatedByMe {
			whereClauses = append(whereClauses, "s.reporter_id = :current_user_id")
		}
	}

	// Date range filters
	if filters.CreatedAfter != nil {
		whereClauses = append(whereClauses, "s.created_at >= :created_after")
	}

	if filters.CreatedBefore != nil {
		whereClauses = append(whereClauses, "s.created_at <= :created_before")
	}

	if filters.UpdatedAfter != nil {
		whereClauses = append(whereClauses, "s.updated_at >= :updated_after")
	}

	if filters.UpdatedBefore != nil {
		whereClauses = append(whereClauses, "s.updated_at <= :updated_before")
	}

	if filters.DeadlineAfter != nil {
		whereClauses = append(whereClauses, "(s.end_date >= :deadline_after)")
	}

	if filters.DeadlineBefore != nil {
		whereClauses = append(whereClauses, "(s.end_date <= :deadline_before)")
	}

	// Completion filters
	if filters.CompletedAfter != nil {
		whereClauses = append(whereClauses, "s.completed_at >= :completed_after")
	}

	if filters.CompletedBefore != nil {
		whereClauses = append(whereClauses, "s.completed_at <= :completed_before")
	}

	return "WHERE " + strings.Join(whereClauses, " AND ")
}

// mapToStoryList converts a map to CoreStoryList
func (r *repo) mapToStoryList(storyMap map[string]any) stories.CoreStoryList {
	story := stories.CoreStoryList{
		Labels:     []uuid.UUID{},
		SubStories: []stories.CoreStoryList{},
	}

	// Helper function to safely convert values
	if id, ok := storyMap["id"].(string); ok {
		if parsed, err := uuid.Parse(id); err == nil {
			story.ID = parsed
		}
	}

	if seqID, ok := storyMap["sequence_id"].(float64); ok {
		story.SequenceID = int(seqID)
	}

	if title, ok := storyMap["title"].(string); ok {
		story.Title = title
	}

	if priority, ok := storyMap["priority"].(string); ok {
		story.Priority = priority
	}

	if statusID, ok := storyMap["status_id"].(string); ok {
		if parsed, err := uuid.Parse(statusID); err == nil {
			story.Status = &parsed
		}
	}

	if assigneeID, ok := storyMap["assignee_id"].(string); ok && assigneeID != "" {
		if parsed, err := uuid.Parse(assigneeID); err == nil {
			story.Assignee = &parsed
		}
	}

	if reporterID, ok := storyMap["reporter_id"].(string); ok && reporterID != "" {
		if parsed, err := uuid.Parse(reporterID); err == nil {
			story.Reporter = &parsed
		}
	}

	if objectiveID, ok := storyMap["objective_id"].(string); ok && objectiveID != "" {
		if parsed, err := uuid.Parse(objectiveID); err == nil {
			story.Objective = &parsed
		}
	}

	if sprintID, ok := storyMap["sprint_id"].(string); ok && sprintID != "" {
		if parsed, err := uuid.Parse(sprintID); err == nil {
			story.Sprint = &parsed
		}
	}

	if teamID, ok := storyMap["team_id"].(string); ok {
		if parsed, err := uuid.Parse(teamID); err == nil {
			story.Team = parsed
		}
	}

	if workspaceID, ok := storyMap["workspace_id"].(string); ok {
		if parsed, err := uuid.Parse(workspaceID); err == nil {
			story.Workspace = parsed
		}
	}

	// Handle time fields
	if createdAt, ok := storyMap["created_at"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, createdAt); err == nil {
			story.CreatedAt = parsed
		}
	}

	if updatedAt, ok := storyMap["updated_at"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			story.UpdatedAt = parsed
		}
	}

	// Handle nullable time fields
	if startDate, ok := storyMap["start_date"].(string); ok && startDate != "" {
		if parsed, err := time.Parse(time.RFC3339, startDate); err == nil {
			story.StartDate = &parsed
		}
	}

	if endDate, ok := storyMap["end_date"].(string); ok && endDate != "" {
		if parsed, err := time.Parse(time.RFC3339, endDate); err == nil {
			story.EndDate = &parsed
		}
	}

	if completedAt, ok := storyMap["completed_at"].(string); ok && completedAt != "" {
		if parsed, err := time.Parse(time.RFC3339, completedAt); err == nil {
			story.CompletedAt = &parsed
		}
	}

	if archivedAt, ok := storyMap["archived_at"].(string); ok && archivedAt != "" {
		if parsed, err := time.Parse(time.RFC3339, archivedAt); err == nil {
			story.ArchivedAt = &parsed
		}
	}

	if deletedAt, ok := storyMap["deleted_at"].(string); ok && deletedAt != "" {
		if parsed, err := time.Parse(time.RFC3339, deletedAt); err == nil {
			story.DeletedAt = &parsed
		}
	}

	// Handle sub-stories
	if subStoriesData, ok := storyMap["sub_stories"].([]any); ok {
		for _, subStoryData := range subStoriesData {
			if subStoryMap, ok := subStoryData.(map[string]any); ok {
				subStory := r.mapToStoryList(subStoryMap)
				story.SubStories = append(story.SubStories, subStory)
			}
		}
	}

	// Handle labels
	if labelsData, ok := storyMap["labels"].([]any); ok {
		for _, labelData := range labelsData {
			if labelID, ok := labelData.(string); ok {
				if parsed, err := uuid.Parse(labelID); err == nil {
					story.Labels = append(story.Labels, parsed)
				}
			}
		}
	}

	return story
}

// buildAllGroupsCTE builds a CTE that returns all possible group values
func (r *repo) buildAllGroupsCTE(groupBy string, filters stories.CoreStoryFilters) string {
	switch groupBy {
	case "status":
		if len(filters.TeamIDs) > 0 {
			// If specific teams are filtered, get all statuses from those teams
			return `
				SELECT CAST(s.status_id AS text) as group_key, s.order_index as sort_order
				FROM statuses s
				WHERE s.workspace_id = :workspace_id
				AND s.team_id = ANY(:team_ids)
			`
		} else {
			// If no team filter, get statuses used by teams user belongs to
			return `
				SELECT CAST(s.status_id AS text) as group_key, s.order_index as sort_order
				FROM statuses s
				WHERE s.workspace_id = :workspace_id
				AND EXISTS (
					SELECT 1
					FROM team_members tm
					WHERE tm.team_id = s.team_id
					AND tm.user_id = :current_user_id
				)
			`
		}
	case "team":
		if len(filters.TeamIDs) > 0 {
			// If specific teams are filtered, only include those
			return `
				SELECT CAST(team_id AS text) as group_key, name as sort_order
				FROM teams 
				WHERE workspace_id = :workspace_id 
				AND team_id = ANY(:team_ids)
			`
		} else {
			// If no team filter, only include teams user belongs to
			return `
				SELECT CAST(t.team_id AS text) as group_key, t.name as sort_order
				FROM teams t
				INNER JOIN team_members tm ON tm.team_id = t.team_id
				WHERE t.workspace_id = :workspace_id 
				AND tm.user_id = :current_user_id
			`
		}
	case "priority":
		return `
			SELECT priority as group_key, 
				CASE priority
					WHEN 'Urgent' THEN 1
					WHEN 'High' THEN 2
					WHEN 'Medium' THEN 3
					WHEN 'Low' THEN 4
					WHEN 'No Priority' THEN 5
					ELSE 6
				END as sort_order
			FROM (VALUES ('Urgent'), ('High'), ('Medium'), ('Low'), ('No Priority')) AS priorities(priority)
		`
	case "assignee":
		if len(filters.TeamIDs) > 0 {
			// If specific teams are filtered, get all users from those teams
			return `
				SELECT CAST(u.user_id AS text) as group_key, u.username as sort_order
				FROM users u
				INNER JOIN team_members tm ON tm.user_id = u.user_id
				WHERE tm.team_id = ANY(:team_ids)
				AND u.is_active = true
				UNION ALL
				SELECT 'null' as group_key, 'Unassigned' as sort_order
			`
		} else {
			return `
				SELECT CAST(u.user_id AS text) as group_key, u.username as sort_order
				FROM users u
				INNER JOIN team_members tm ON tm.user_id = u.user_id
				INNER JOIN teams t ON t.team_id = tm.team_id
				INNER JOIN team_members tm2 ON tm2.team_id = t.team_id
				WHERE t.workspace_id = :workspace_id
				AND tm2.user_id = :current_user_id
				AND u.is_active = true
				UNION ALL
				SELECT 'null' as group_key, 'Unassigned' as sort_order
			`
		}
	case "sprint":
		if len(filters.TeamIDs) > 0 {
			// If specific teams are filtered, get all sprints from those teams
			return `
				SELECT CAST(sprint_id AS text) as group_key, name as sort_order
				FROM sprints 
				WHERE workspace_id = :workspace_id
				AND team_id = ANY(:team_ids)
				UNION ALL
				SELECT 'null' as group_key, 'No Sprint' as sort_order
			`
		} else {
			return `
				SELECT CAST(s.sprint_id AS text) as group_key, s.name as sort_order
				FROM sprints s
				INNER JOIN team_members tm ON tm.team_id = s.team_id
				WHERE s.workspace_id = :workspace_id
				AND tm.user_id = :current_user_id
				UNION ALL
				SELECT 'null' as group_key, 'No Sprint' as sort_order
			`
		}
	default:
		return `SELECT 'all' as group_key, 0 as sort_order`
	}
}

// getGroupColumn returns the column name for grouping
func (r *repo) getGroupColumn(groupBy string) string {
	switch groupBy {
	case "status":
		return "s.status_id"
	case "assignee":
		return "s.assignee_id"
	case "priority":
		return "s.priority"
	case "team":
		return "s.team_id"
	case "sprint":
		return "s.sprint_id"
	default:
		return "s.id"
	}
}

// buildOrderByClause builds the ORDER BY clause based on orderBy and orderDirection
func (r *repo) buildOrderByClause(orderBy, orderDirection string) string {
	return r.buildOrderByClauseWithAlias(orderBy, orderDirection, "s")
}

// buildOrderByClauseWithAlias builds the ORDER BY clause with a specific table alias
func (r *repo) buildOrderByClauseWithAlias(orderBy, orderDirection, tableAlias string) string {
	var column string
	switch orderBy {
	case "created":
		column = fmt.Sprintf("%s.created_at", tableAlias)
	case "updated":
		column = fmt.Sprintf("%s.updated_at", tableAlias)
	case "priority":
		// Use a more efficient approach for priority ordering
		// First sort by priority using a simple mapping, then by created_at for consistency
		direction := "ASC"
		if orderDirection == "desc" {
			direction = "DESC"
		}
		return fmt.Sprintf(`CASE %s.priority
			WHEN 'Urgent' THEN 1
			WHEN 'High' THEN 2
			WHEN 'Medium' THEN 3
			WHEN 'Low' THEN 4
			WHEN 'No Priority' THEN 5
			ELSE 6
		END %s, %s.created_at DESC`, tableAlias, direction, tableAlias)
	case "deadline":
		// Handle NULL values efficiently - put them last for both ASC and DESC
		direction := "ASC"
		if orderDirection == "desc" {
			direction = "DESC"
		}
		return fmt.Sprintf("%s.end_date %s NULLS LAST, %s.created_at DESC", tableAlias, direction, tableAlias)
	default:
		column = fmt.Sprintf("%s.created_at", tableAlias)
	}

	direction := "DESC"
	if orderDirection == "asc" {
		direction = "ASC"
	}

	return fmt.Sprintf("%s %s", column, direction)
}

// buildQueryParams builds the parameter map for the SQL query
func (r *repo) buildQueryParams(filters stories.CoreStoryFilters) map[string]any {
	params := map[string]any{
		"workspace_id":    filters.WorkspaceID,
		"current_user_id": filters.CurrentUserID,
	}

	// Direct assignment - let database driver handle array conversion
	if len(filters.StatusIDs) > 0 {
		params["status_ids"] = filters.StatusIDs
	}
	if len(filters.AssigneeIDs) > 0 {
		params["assignee_ids"] = filters.AssigneeIDs
	}
	if len(filters.ReporterIDs) > 0 {
		params["reporter_ids"] = filters.ReporterIDs
	}
	if len(filters.Priorities) > 0 {
		params["priorities"] = filters.Priorities
	}
	if len(filters.Categories) > 0 {
		params["categories"] = filters.Categories
	}
	if len(filters.TeamIDs) > 0 {
		params["team_ids"] = filters.TeamIDs
	}
	if len(filters.SprintIDs) > 0 {
		params["sprint_ids"] = filters.SprintIDs
	}
	if len(filters.LabelIDs) > 0 {
		params["label_ids"] = filters.LabelIDs
	}

	// Single value parameters
	if filters.Parent != nil {
		params["parent_id"] = *filters.Parent
	}
	if filters.Objective != nil {
		params["objective_id"] = *filters.Objective
	}
	if filters.Epic != nil {
		params["epic_id"] = *filters.Epic
	}

	// Date range parameters
	if filters.CreatedAfter != nil {
		params["created_after"] = *filters.CreatedAfter
	}
	if filters.CreatedBefore != nil {
		params["created_before"] = *filters.CreatedBefore
	}
	if filters.UpdatedAfter != nil {
		params["updated_after"] = *filters.UpdatedAfter
	}
	if filters.UpdatedBefore != nil {
		params["updated_before"] = *filters.UpdatedBefore
	}
	if filters.DeadlineAfter != nil {
		params["deadline_after"] = *filters.DeadlineAfter
	}
	if filters.DeadlineBefore != nil {
		params["deadline_before"] = *filters.DeadlineBefore
	}
	if filters.CompletedAfter != nil {
		params["completed_after"] = *filters.CompletedAfter
	}
	if filters.CompletedBefore != nil {
		params["completed_before"] = *filters.CompletedBefore
	}

	return params
}

// buildStoriesQuery builds the base SQL query for stories with filters
func (r *repo) buildStoriesQuery(filters stories.CoreStoryFilters) string {
	query := `
		SELECT
			s.id,
			s.sequence_id,
			s.title,
			s.priority,
			s.status_id,
			s.start_date,
			s.end_date,
			s.sprint_id,
			s.team_id,
			s.objective_id,
			s.workspace_id,
			s.assignee_id,
			s.reporter_id,
			s.created_at,
			s.updated_at,
			s.completed_at,
			s.deleted_at,
			s.archived_at,
			COALESCE(
				(
					SELECT
						json_agg(
							json_build_object(
								'id', sub.id,
								'sequence_id', sub.sequence_id,
								'title', sub.title,
								'priority', sub.priority,
								'status_id', sub.status_id,
								'start_date', sub.start_date,
								'end_date', sub.end_date,
								'sprint_id', sub.sprint_id,
								'team_id', sub.team_id,
								'objective_id', sub.objective_id,
								'workspace_id', sub.workspace_id,
								'assignee_id', sub.assignee_id,
								'reporter_id', sub.reporter_id,
								'created_at', sub.created_at,
								'updated_at', sub.updated_at,
								'completed_at', sub.completed_at,
								'deleted_at', sub.deleted_at,
								'archived_at', sub.archived_at
							)
						)
					FROM
						stories sub
					WHERE
						sub.parent_id = s.id
						AND sub.deleted_at IS NULL
				), CAST('[]' AS json)
			) AS sub_stories,
			COALESCE(
				(
					SELECT
						json_agg(l.label_id)
					FROM
						labels l
						INNER JOIN story_labels sl ON sl.label_id = l.label_id
					WHERE
						sl.story_id = s.id
				), CAST('[]' AS json)
			) AS labels
		FROM
			stories s
	`

	// Add team member join if needed for assigned_to_me or created_by_me filters
	needsTeamJoin := filters.AssignedToMe != nil || filters.CreatedByMe != nil
	if needsTeamJoin {
		query += `
			INNER JOIN team_members tm ON 
				tm.team_id = s.team_id 
				AND tm.user_id = :current_user_id
		`
	}

	// Add status join if categories filter is present
	if len(filters.Categories) > 0 {
		query += `
			INNER JOIN statuses stat ON s.status_id = stat.status_id
		`
	}

	// Build WHERE clauses
	whereClauses := []string{
		"s.workspace_id = :workspace_id",
		"s.parent_id IS NULL",
	}

	// Exclude archived unless explicitly included
	if filters.IncludeArchived != nil && *filters.IncludeArchived {
		whereClauses = append(whereClauses, "s.archived_at IS NOT NULL")
	} else {
		whereClauses = append(whereClauses, "s.archived_at IS NULL")
	}

	// Exclude deleted unless explicitly included
	if filters.IncludeDeleted != nil && *filters.IncludeDeleted {
		whereClauses = append(whereClauses, "s.deleted_at IS NOT NULL")
	} else {
		whereClauses = append(whereClauses, "s.deleted_at IS NULL")
	}

	// Add filter conditions
	if len(filters.StatusIDs) > 0 {
		whereClauses = append(whereClauses, "s.status_id = ANY(:status_ids)")
	}

	if len(filters.AssigneeIDs) > 0 {
		whereClauses = append(whereClauses, "s.assignee_id = ANY(:assignee_ids)")
	}

	if len(filters.ReporterIDs) > 0 {
		whereClauses = append(whereClauses, "s.reporter_id = ANY(:reporter_ids)")
	}

	if len(filters.Priorities) > 0 {
		whereClauses = append(whereClauses, "s.priority = ANY(:priorities)")
	}

	if len(filters.Categories) > 0 {
		whereClauses = append(whereClauses, "stat.category = ANY(:categories)")
	}

	if len(filters.TeamIDs) > 0 {
		whereClauses = append(whereClauses, "s.team_id = ANY(:team_ids)")
	} else {
		// If no specific teams are provided, only show stories from teams user is a member of
		whereClauses = append(whereClauses, "EXISTS (SELECT 1 FROM team_members tm WHERE tm.team_id = s.team_id AND tm.user_id = :current_user_id)")
	}

	if len(filters.SprintIDs) > 0 {
		whereClauses = append(whereClauses, "s.sprint_id = ANY(:sprint_ids)")
	}

	if len(filters.LabelIDs) > 0 {
		query += `
			INNER JOIN story_labels sl_filter ON sl_filter.story_id = s.id
		`
		whereClauses = append(whereClauses, "sl_filter.label_id = ANY(:label_ids)")
	}

	if filters.Parent != nil {
		whereClauses = append(whereClauses, "s.parent_id = :parent_id")
	}

	if filters.Objective != nil {
		whereClauses = append(whereClauses, "s.objective_id = :objective_id")
	}

	if filters.Epic != nil {
		whereClauses = append(whereClauses, "s.epic_id = :epic_id")
	}

	if filters.HasNoAssignee != nil && *filters.HasNoAssignee {
		whereClauses = append(whereClauses, "s.assignee_id IS NULL")
	}

	// Handle createdByMe and assignedToMe with OR logic when both are true
	if filters.AssignedToMe != nil && *filters.AssignedToMe && filters.CreatedByMe != nil && *filters.CreatedByMe {
		whereClauses = append(whereClauses, "(s.assignee_id = :current_user_id OR s.reporter_id = :current_user_id)")
	} else {
		if filters.AssignedToMe != nil && *filters.AssignedToMe {
			whereClauses = append(whereClauses, "s.assignee_id = :current_user_id")
		}

		if filters.CreatedByMe != nil && *filters.CreatedByMe {
			whereClauses = append(whereClauses, "s.reporter_id = :current_user_id")
		}
	}

	// Date range filters
	if filters.CreatedAfter != nil {
		whereClauses = append(whereClauses, "s.created_at >= :created_after")
	}

	if filters.CreatedBefore != nil {
		whereClauses = append(whereClauses, "s.created_at <= :created_before")
	}

	if filters.UpdatedAfter != nil {
		whereClauses = append(whereClauses, "s.updated_at >= :updated_after")
	}

	if filters.UpdatedBefore != nil {
		whereClauses = append(whereClauses, "s.updated_at <= :updated_before")
	}

	if filters.DeadlineAfter != nil {
		whereClauses = append(whereClauses, "(s.end_date >= :deadline_after)")
	}

	if filters.DeadlineBefore != nil {
		whereClauses = append(whereClauses, "(s.end_date <= :deadline_before)")
	}

	// Completion filters
	if filters.CompletedAfter != nil {
		whereClauses = append(whereClauses, "s.completed_at >= :completed_after")
	}

	if filters.CompletedBefore != nil {
		whereClauses = append(whereClauses, "s.completed_at <= :completed_before")
	}

	query += " WHERE " + strings.Join(whereClauses, " AND ")

	return query
}

// listGroupedStoriesGo performs grouping in Go (original implementation)
func (r *repo) listGroupedStoriesGo(ctx context.Context, query stories.CoreStoryQuery) ([]stories.CoreStoryGroup, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.listGroupedStoriesGo")
	defer span.End()

	// For now, implement a simple approach - get all stories and group in Go
	// TODO: Optimize with SQL-based grouping for better performance

	baseQuery := r.buildStoriesQuery(query.Filters)
	orderByClause := r.buildOrderByClause(query.OrderBy, query.OrderDirection)
	baseQuery += fmt.Sprintf(" ORDER BY %s", orderByClause)

	params := r.buildQueryParams(query.Filters)

	var allStories []dbStory
	stmt, err := r.db.PrepareNamedContext(ctx, baseQuery)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &allStories, params); err != nil {
		errMsg := fmt.Sprintf("Failed to retrieve stories from the database: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("stories not found"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	// Group stories in Go
	groups := r.groupStories(allStories, query.GroupBy, query.StoriesPerGroup, query.OrderBy, query.OrderDirection)

	span.AddEvent("Go grouped stories retrieved.", trace.WithAttributes(
		attribute.Int("groups.count", len(groups)),
		attribute.String("group.by", query.GroupBy),
	))

	return groups, nil
}

// groupStories groups stories by the specified field and limits stories per group
func (r *repo) groupStories(allStories []dbStory, groupBy string, storiesPerGroup int, orderBy, orderDirection string) []stories.CoreStoryGroup {
	groupMap := make(map[string]*stories.CoreStoryGroup)

	for _, story := range allStories {
		key := r.getGroupKey(story, groupBy)

		// Initialize group if not exists
		if groupMap[key] == nil {
			groupMap[key] = &stories.CoreStoryGroup{
				Key:         key,
				LoadedCount: 0,
				TotalCount:  0, // Will be calculated after all stories are processed
				HasMore:     false,
				Stories:     []stories.CoreStoryList{},
				NextPage:    2, // Next page for load more
			}
		}

		group := groupMap[key]

		// Only add to Stories slice if under the limit
		if group.LoadedCount < storiesPerGroup {
			coreStories := toCoreStories([]dbStory{story})
			if len(coreStories) > 0 {
				group.Stories = append(group.Stories, coreStories[0])
				group.LoadedCount++
			}
		} else if group.LoadedCount == storiesPerGroup {
			// We found one more story beyond the limit, so there are more
			group.HasMore = true
		}
	}

	// Calculate total count for each group by counting all stories in that group
	groupCounts := make(map[string]int)
	for _, story := range allStories {
		key := r.getGroupKey(story, groupBy)
		groupCounts[key]++
	}

	// Convert map to slice and set total counts
	var groups []stories.CoreStoryGroup
	for _, group := range groupMap {
		group.TotalCount = groupCounts[group.Key]
		groups = append(groups, *group)
	}

	// Sort groups according to the same logic as SQL-based grouping
	r.sortGroups(groups, groupBy)

	// Sort stories within each group (only if not using default ordering)
	if orderBy != "created" || orderDirection != "desc" {
		for i := range groups {
			r.sortStoriesInGroup(&groups[i], orderBy, orderDirection)
		}
	}

	return groups
}

// sortGroups sorts groups according to the same logic as SQL-based grouping
func (r *repo) sortGroups(groups []stories.CoreStoryGroup, groupBy string) {
	sort.Slice(groups, func(i, j int) bool {
		sortOrderI := r.getGroupSortOrder(groups[i].Key, groupBy)
		sortOrderJ := r.getGroupSortOrder(groups[j].Key, groupBy)

		// First sort by sort order, then by group key for consistent ordering
		if sortOrderI != sortOrderJ {
			return sortOrderI < sortOrderJ
		}
		return groups[i].Key < groups[j].Key
	})
}

// getGroupSortOrder returns the sort order for a group key based on group type
func (r *repo) getGroupSortOrder(groupKey, groupBy string) int {
	switch groupBy {
	case "priority":
		switch groupKey {
		case "Urgent":
			return 1
		case "High":
			return 2
		case "Medium":
			return 3
		case "Low":
			return 4
		case "No Priority":
			return 5
		default:
			return 6
		}
	case "status":
		// For status, we can't easily get order_index in Go grouping without additional DB call
		// So we'll just use alphabetical for now (SQL grouping will still use proper order_index)
		// This fallback is only used for complex grouping scenarios anyway
		return 0
	default:
		// For assignee, team, sprint - use alphabetical ordering
		return 0
	}
}

// sortStoriesInGroup sorts stories within a group based on orderBy and orderDirection
func (r *repo) sortStoriesInGroup(group *stories.CoreStoryGroup, orderBy, orderDirection string) {
	// Skip sorting if group is empty or has only one story
	if len(group.Stories) <= 1 {
		return
	}

	isAsc := orderDirection == "asc"

	sort.Slice(group.Stories, func(i, j int) bool {
		var compareResult int

		switch orderBy {
		case "created":
			compareResult = group.Stories[i].CreatedAt.Compare(group.Stories[j].CreatedAt)
		case "updated":
			compareResult = group.Stories[i].UpdatedAt.Compare(group.Stories[j].UpdatedAt)
		case "priority":
			priorityOrderI := r.getPrioritySortOrder(group.Stories[i].Priority)
			priorityOrderJ := r.getPrioritySortOrder(group.Stories[j].Priority)
			compareResult = priorityOrderI - priorityOrderJ
			// If priorities are equal, sort by created_at DESC for consistency
			if compareResult == 0 {
				compareResult = -group.Stories[i].CreatedAt.Compare(group.Stories[j].CreatedAt)
			}
		case "deadline":
			// Handle nil end dates efficiently - put them last for both ASC and DESC
			if group.Stories[i].EndDate == nil && group.Stories[j].EndDate == nil {
				// Both nil, sort by created_at DESC for consistency
				compareResult = -group.Stories[i].CreatedAt.Compare(group.Stories[j].CreatedAt)
			} else if group.Stories[i].EndDate == nil {
				compareResult = 1 // nil goes to end
			} else if group.Stories[j].EndDate == nil {
				compareResult = -1 // nil goes to end
			} else {
				compareResult = group.Stories[i].EndDate.Compare(*group.Stories[j].EndDate)
				// If deadlines are equal, sort by created_at DESC for consistency
				if compareResult == 0 {
					compareResult = -group.Stories[i].CreatedAt.Compare(group.Stories[j].CreatedAt)
				}
			}
		default:
			// Default to created date
			compareResult = group.Stories[i].CreatedAt.Compare(group.Stories[j].CreatedAt)
		}

		// Apply direction
		if isAsc {
			return compareResult < 0
		} else {
			return compareResult > 0
		}
	})
}

// getPrioritySortOrder returns the sort order for a priority value
func (r *repo) getPrioritySortOrder(priority string) int {
	switch priority {
	case "Urgent":
		return 1
	case "High":
		return 2
	case "Medium":
		return 3
	case "Low":
		return 4
	case "No Priority":
		return 5
	default:
		return 6
	}
}

// getGroupKey extracts the group key from a story
func (r *repo) getGroupKey(story dbStory, groupBy string) string {
	switch groupBy {
	case "status":
		if story.Status != nil {
			return story.Status.String()
		}
		return "null"
	case "assignee":
		if story.Assignee != nil {
			return story.Assignee.String()
		}
		return "null"
	case "priority":
		return story.Priority
	case "team":
		return story.Team.String()
	case "sprint":
		if story.Sprint != nil {
			return story.Sprint.String()
		}
		return "null"
	default:
		return "all"
	}
}

// ListGroupStories returns more stories for a specific group (for load more functionality)
func (r *repo) ListGroupStories(ctx context.Context, groupKey string, query stories.CoreStoryQuery) ([]stories.CoreStoryList, bool, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.ListGroupStories")
	defer span.End()

	// Build simplified query for list view (no N+1 subqueries)
	baseQuery := r.buildSimpleStoriesQuery(query.Filters)

	// Add group-specific filter based on groupBy type, unless groupKey is "none"
	if groupKey != "none" {
		if groupKey == "null" || groupKey == "" {
			// Handle NULL values specially
			baseQuery += fmt.Sprintf(" AND %s IS NULL", r.getGroupColumn(query.GroupBy))
		} else {
			baseQuery += fmt.Sprintf(" AND %s = :group_key", r.getGroupColumn(query.GroupBy))
		}
	}
	// If groupKey is "none", don't add any group-specific filters - return all matching stories

	// Get paginated stories - fetch one extra to check if there are more
	pageSize := query.PageSize
	offset := (query.Page - 1) * pageSize
	orderByClause := r.buildOrderByClause(query.OrderBy, query.OrderDirection)
	baseQuery += fmt.Sprintf(" ORDER BY %s LIMIT %d OFFSET %d", orderByClause, pageSize+1, offset)

	params := r.buildQueryParams(query.Filters)
	if groupKey != "null" && groupKey != "" && groupKey != "none" {
		params["group_key"] = groupKey
	}

	var dbStories []dbStory
	stmt, err := r.db.PrepareNamedContext(ctx, baseQuery)
	if err != nil {
		return nil, false, fmt.Errorf("failed to prepare stories query: %w", err)
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &dbStories, params); err != nil {
		return nil, false, fmt.Errorf("failed to get stories: %w", err)
	}

	// Check if there are more stories
	hasMore := len(dbStories) > pageSize
	if hasMore {
		// Remove the extra story we fetched
		dbStories = dbStories[:pageSize]
	}

	coreStories := toCoreStories(dbStories)

	span.AddEvent("group stories retrieved.", trace.WithAttributes(
		attribute.Int("stories.count", len(coreStories)),
		attribute.String("group.key", groupKey),
		attribute.Bool("has.more", hasMore),
	))

	return coreStories, hasMore, nil
}

// buildSimpleStoriesQuery builds a simplified query for list views without N+1 subqueries
func (r *repo) buildSimpleStoriesQuery(filters stories.CoreStoryFilters) string {
	query := `
		SELECT
			s.id,
			s.sequence_id,
			s.title,
			s.priority,
			s.status_id,
			s.start_date,
			s.end_date,
			s.sprint_id,
			s.team_id,
			s.objective_id,
			s.workspace_id,
			s.assignee_id,
			s.reporter_id,
			s.created_at,
			s.updated_at,
			s.completed_at,
			s.deleted_at,
			s.archived_at,
			COALESCE(
				(
					SELECT
						json_agg(
							json_build_object(
								'id', sub.id,
								'sequence_id', sub.sequence_id,
								'title', sub.title,
								'priority', sub.priority,
								'status_id', sub.status_id,
								'start_date', sub.start_date,
								'end_date', sub.end_date,
								'sprint_id', sub.sprint_id,
								'team_id', sub.team_id,
								'objective_id', sub.objective_id,
								'workspace_id', sub.workspace_id,
								'assignee_id', sub.assignee_id,
								'reporter_id', sub.reporter_id,
								'created_at', sub.created_at,
								'updated_at', sub.updated_at,
								'completed_at', sub.completed_at,
								'deleted_at', sub.deleted_at,
								'archived_at', sub.archived_at,
								'labels', '[]'
							)
						)
					FROM
						stories sub
					WHERE
						sub.parent_id = s.id 
						AND sub.deleted_at IS NULL
				), '[]'
			) AS sub_stories,
			COALESCE(
				(
					SELECT
						json_agg(l.label_id)
					FROM
						labels l
						INNER JOIN story_labels sl ON sl.label_id = l.label_id
					WHERE
						sl.story_id = s.id
				), '[]'
			) AS labels
		FROM
			stories s
	`

	// Add team member join if needed for assigned_to_me or created_by_me filters
	needsTeamJoin := filters.AssignedToMe != nil || filters.CreatedByMe != nil
	if needsTeamJoin {
		query += `
			INNER JOIN team_members tm ON 
				tm.team_id = s.team_id 
				AND tm.user_id = :current_user_id
		`
	}

	// Add status join if categories filter is present
	if len(filters.Categories) > 0 {
		query += `
			INNER JOIN statuses stat ON s.status_id = stat.status_id
		`
	}

	// Add label join if needed for filtering (but don't fetch label data)
	if len(filters.LabelIDs) > 0 {
		query += `
			INNER JOIN story_labels sl_filter ON sl_filter.story_id = s.id
		`
	}

	// Build WHERE clauses (same as buildStoriesQuery)
	whereClauses := []string{
		"s.workspace_id = :workspace_id",
		"s.parent_id IS NULL",
	}

	// Exclude archived unless explicitly included
	if filters.IncludeArchived != nil && *filters.IncludeArchived {
		whereClauses = append(whereClauses, "s.archived_at IS NOT NULL")
	} else {
		whereClauses = append(whereClauses, "s.archived_at IS NULL")
	}

	// Exclude deleted unless explicitly included
	if filters.IncludeDeleted != nil && *filters.IncludeDeleted {
		whereClauses = append(whereClauses, "s.deleted_at IS NOT NULL")
	} else {
		whereClauses = append(whereClauses, "s.deleted_at IS NULL")
	}

	// Add filter conditions
	if len(filters.StatusIDs) > 0 {
		whereClauses = append(whereClauses, "s.status_id = ANY(:status_ids)")
	}

	if len(filters.AssigneeIDs) > 0 {
		whereClauses = append(whereClauses, "s.assignee_id = ANY(:assignee_ids)")
	}

	if len(filters.ReporterIDs) > 0 {
		whereClauses = append(whereClauses, "s.reporter_id = ANY(:reporter_ids)")
	}

	if len(filters.Priorities) > 0 {
		whereClauses = append(whereClauses, "s.priority = ANY(:priorities)")
	}

	if len(filters.Categories) > 0 {
		whereClauses = append(whereClauses, "stat.category = ANY(:categories)")
	}

	if len(filters.TeamIDs) > 0 {
		whereClauses = append(whereClauses, "s.team_id = ANY(:team_ids)")
	} else {
		// If no specific teams are provided, only show stories from teams user is a member of
		whereClauses = append(whereClauses, "EXISTS (SELECT 1 FROM team_members tm WHERE tm.team_id = s.team_id AND tm.user_id = :current_user_id)")
	}

	if len(filters.SprintIDs) > 0 {
		whereClauses = append(whereClauses, "s.sprint_id = ANY(:sprint_ids)")
	}

	if len(filters.LabelIDs) > 0 {
		whereClauses = append(whereClauses, "sl_filter.label_id = ANY(:label_ids)")
	}

	if filters.Parent != nil {
		whereClauses = append(whereClauses, "s.parent_id = :parent_id")
	}

	if filters.Objective != nil {
		whereClauses = append(whereClauses, "s.objective_id = :objective_id")
	}

	if filters.Epic != nil {
		whereClauses = append(whereClauses, "s.epic_id = :epic_id")
	}

	if filters.HasNoAssignee != nil && *filters.HasNoAssignee {
		whereClauses = append(whereClauses, "s.assignee_id IS NULL")
	}

	// Handle createdByMe and assignedToMe with OR logic when both are true
	if filters.AssignedToMe != nil && *filters.AssignedToMe && filters.CreatedByMe != nil && *filters.CreatedByMe {
		whereClauses = append(whereClauses, "(s.assignee_id = :current_user_id OR s.reporter_id = :current_user_id)")
	} else {
		if filters.AssignedToMe != nil && *filters.AssignedToMe {
			whereClauses = append(whereClauses, "s.assignee_id = :current_user_id")
		}

		if filters.CreatedByMe != nil && *filters.CreatedByMe {
			whereClauses = append(whereClauses, "s.reporter_id = :current_user_id")
		}
	}

	// Date range filters
	if filters.CreatedAfter != nil {
		whereClauses = append(whereClauses, "s.created_at >= :created_after")
	}

	if filters.CreatedBefore != nil {
		whereClauses = append(whereClauses, "s.created_at <= :created_before")
	}

	if filters.UpdatedAfter != nil {
		whereClauses = append(whereClauses, "s.updated_at >= :updated_after")
	}

	if filters.UpdatedBefore != nil {
		whereClauses = append(whereClauses, "s.updated_at <= :updated_before")
	}

	if filters.DeadlineAfter != nil {
		whereClauses = append(whereClauses, "(s.end_date >= :deadline_after)")
	}

	if filters.DeadlineBefore != nil {
		whereClauses = append(whereClauses, "(s.end_date <= :deadline_before)")
	}

	// Completion filters
	if filters.CompletedAfter != nil {
		whereClauses = append(whereClauses, "s.completed_at >= :completed_after")
	}

	if filters.CompletedBefore != nil {
		whereClauses = append(whereClauses, "s.completed_at <= :completed_before")
	}

	query += " WHERE " + strings.Join(whereClauses, " AND ")

	return query
}

// ListByCategory returns stories filtered by category with pagination
func (r *repo) ListByCategory(ctx context.Context, workspaceId, userID, teamId uuid.UUID, category string, page, pageSize int) ([]stories.CoreStoryList, bool, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.ListByCategory")
	defer span.End()

	// Build simplified query with status join for category filtering
	query := `
		SELECT
			s.id,
			s.sequence_id,
			s.title,
			s.priority,
			s.status_id,
			s.start_date,
			s.end_date,
			s.sprint_id,
			s.team_id,
			s.objective_id,
			s.workspace_id,
			s.assignee_id,
			s.reporter_id,
			s.created_at,
			s.updated_at,
			s.completed_at,
			s.deleted_at,
			s.archived_at,
			COALESCE(
				(
					SELECT
						json_agg(
							json_build_object(
								'id', sub.id,
								'sequence_id', sub.sequence_id,
								'title', sub.title,
								'priority', sub.priority,
								'status_id', sub.status_id,
								'start_date', sub.start_date,
								'end_date', sub.end_date,
								'sprint_id', sub.sprint_id,
								'team_id', sub.team_id,
								'objective_id', sub.objective_id,
								'workspace_id', sub.workspace_id,
								'assignee_id', sub.assignee_id,
								'reporter_id', sub.reporter_id,
								'created_at', sub.created_at,
								'updated_at', sub.updated_at,
								'completed_at', sub.completed_at,
								'deleted_at', sub.deleted_at,
								'archived_at', sub.archived_at,
								'labels', '[]'
							)
						)
					FROM
						stories sub
					WHERE
						sub.parent_id = s.id 
						AND sub.deleted_at IS NULL
				), '[]'
			) AS sub_stories,
			COALESCE(
				(
					SELECT
						json_agg(l.label_id)
					FROM
						labels l
						INNER JOIN story_labels sl ON sl.label_id = l.label_id
					WHERE
						sl.story_id = s.id
				), '[]'
			) AS labels
		FROM
			stories s
		INNER JOIN statuses stat ON s.status_id = stat.status_id
		WHERE
			s.workspace_id = :workspace_id
			AND s.team_id = :team_id
			AND s.deleted_at IS NULL
			AND s.archived_at IS NULL
			AND s.parent_id IS NULL
			AND stat.category = :category
		ORDER BY s.created_at DESC
		LIMIT :limit OFFSET :offset
	`

	// Calculate offset and fetch one extra to check if there are more
	offset := (page - 1) * pageSize
	limit := pageSize + 1

	params := map[string]any{
		"workspace_id": workspaceId,
		"team_id":      teamId,
		"category":     category,
		"limit":        limit,
		"offset":       offset,
	}

	var dbStories []dbStory
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, false, fmt.Errorf("failed to prepare category stories query: %w", err)
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &dbStories, params); err != nil {
		return nil, false, fmt.Errorf("failed to get stories by category: %w", err)
	}

	// Check if there are more stories
	hasMore := len(dbStories) > pageSize
	if hasMore {
		// Remove the extra story we fetched
		dbStories = dbStories[:pageSize]
	}

	coreStories := toCoreStories(dbStories)

	span.AddEvent("category stories retrieved.", trace.WithAttributes(
		attribute.Int("stories.count", len(coreStories)),
		attribute.String("category", category),
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
		attribute.Bool("has.more", hasMore),
	))

	return coreStories, hasMore, nil
}

// QueryByRef returns a story by team code and sequence ID.
func (r *repo) QueryByRef(ctx context.Context, workspaceId uuid.UUID, teamCode string, sequenceID int) (stories.CoreSingleStory, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.QueryByRef")
	defer span.End()

	story, err := r.getStoryByRef(ctx, workspaceId, teamCode, sequenceID)
	if err != nil {
		span.RecordError(err)
		return stories.CoreSingleStory{}, err
	}

	return toCoreStory(story), nil
}

func (r *repo) getStoryByRef(ctx context.Context, workspaceId uuid.UUID, teamCode string, sequenceID int) (dbStory, error) {
	query := `
        SELECT
					s.id,
					s.title,
					s.priority,
					s.sequence_id,
					s.status_id,
					s.description,
					s.description_html,
					s.team_id,
					s.objective_id,
					s.sprint_id,
					s.key_result_id,
					s.workspace_id,
					s.assignee_id,
					s.reporter_id,
					s.start_date,
					s.end_date,
					s.created_at,
					s.updated_at,
					s.deleted_at,
					s.archived_at,
					s.completed_at,
					COALESCE(
							(
									SELECT
											json_agg(sub.*)
									FROM
											stories sub
									WHERE
											sub.parent_id = s.id AND sub.deleted_at IS NULL
							), '[]'
					) AS sub_stories,
					COALESCE(
						(
								SELECT
										json_agg(l.label_id)
								FROM
										labels l
										INNER JOIN story_labels sl ON sl.label_id = l.label_id
								WHERE
										sl.story_id = s.id
						), '[]') AS labels
				FROM
					stories s
					INNER JOIN teams t ON s.team_id = t.team_id
				WHERE
					s.sequence_id = :sequence_id
					AND t.code = :team_code
					AND s.workspace_id = :workspace_id
					AND s.deleted_at IS NULL;
    `

	params := map[string]any{
		"sequence_id":  sequenceID,
		"team_code":    teamCode,
		"workspace_id": workspaceId,
	}

	var story dbStory

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, fmt.Sprintf("failed to prepare named statement: %s", err), "teamCode", teamCode, "sequenceID", sequenceID)
		return dbStory{}, err
	}
	defer stmt.Close()

	err = stmt.GetContext(ctx, &story, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return dbStory{}, errors.New("story not found")
		}
		r.log.Error(ctx, fmt.Sprintf("failed to execute query: %s", err), "teamCode", teamCode, "sequenceID", sequenceID)
		return dbStory{}, err
	}

	return story, nil
}

// GetStoryIDByRef returns story UUID by workspace, team code and sequence.
func (r *repo) GetStoryIDByRef(ctx context.Context, workspaceId uuid.UUID, teamCode string, sequenceID int) (uuid.UUID, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.GetStoryIDByRef")
	defer span.End()

	q := `SELECT s.id
		  FROM stories s
		  INNER JOIN teams t ON s.team_id = t.team_id
		  WHERE t.code = :team_code
			AND s.sequence_id = :sequence_id
			AND s.workspace_id = :workspace_id
			AND s.deleted_at IS NULL
		  LIMIT 1`

	params := map[string]any{
		"team_code":    teamCode,
		"sequence_id":  sequenceID,
		"workspace_id": workspaceId,
	}

	var id uuid.UUID
	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		r.log.Error(ctx, "failed to prepare GetStoryIDByRef", "error", err)
		return uuid.Nil, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &id, params); err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, errors.New("story not found")
		}
		return uuid.Nil, err
	}
	return id, nil
}

// CreateStoryFromIssue inserts a minimal story when an issue is opened in GitHub.
// It returns the new story id.
func (r *repo) CreateStoryFromIssue(ctx context.Context, workspaceID, teamID uuid.UUID, title, description string, reporterID uuid.UUID) (uuid.UUID, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.CreateStoryFromIssue")
	defer span.End()

	// 1) Get default status for team (category 'unstarted')
	var statusID uuid.UUID
	statusQuery := `SELECT status_id FROM statuses WHERE team_id = :team_id AND category = 'unstarted' LIMIT 1`
	stmtStatus, err := r.db.PrepareNamedContext(ctx, statusQuery)
	if err != nil {
		return uuid.Nil, fmt.Errorf("prepare status query: %w", err)
	}
	if err := stmtStatus.GetContext(ctx, &statusID, map[string]any{"team_id": teamID}); err != nil {
		stmtStatus.Close()
		return uuid.Nil, fmt.Errorf("fetch default status: %w", err)
	}
	stmtStatus.Close()

	// 2) Get next sequence id
	sequenceID, commit, rollback, err := r.GetNextSequenceID(ctx, teamID, workspaceID)
	if err != nil {
		return uuid.Nil, err
	}

	// 3) Insert story
	insertQuery := `
		INSERT INTO stories (
			sequence_id, title, description, description_html, status_id, priority, team_id, workspace_id, reporter_id, created_at, updated_at
		) VALUES (
			:sequence_id, :title, :description, :description_html, :status_id, :priority, :team_id, :workspace_id, :reporter_id, NOW(), NOW()
		) RETURNING id`

	params := map[string]any{
		"sequence_id":      sequenceID + 1,
		"title":            title,
		"description":      description,
		"description_html": description,
		"status_id":        statusID,
		"team_id":          teamID,
		"workspace_id":     workspaceID,
		"reporter_id":      reporterID,
		"priority":         "No Priority",
	}

	stmt, err := r.db.PrepareNamedContext(ctx, insertQuery)
	if err != nil {
		rollback()
		return uuid.Nil, fmt.Errorf("prepare insert story: %w", err)
	}
	var storyID uuid.UUID
	if err := stmt.GetContext(ctx, &storyID, params); err != nil {
		stmt.Close()
		rollback()
		return uuid.Nil, fmt.Errorf("insert story: %w", err)
	}
	stmt.Close()

	if err := commit(); err != nil {
		return uuid.Nil, err
	}
	return storyID, nil
}

// UpdateStoryStatus updates only the status of a story - used for automated transitions
func (r *repo) UpdateStoryStatus(ctx context.Context, storyID uuid.UUID, workspaceID uuid.UUID, statusID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.stories.UpdateStoryStatus")
	defer span.End()

	updates := map[string]any{
		"status_id": statusID,
	}

	return r.Update(ctx, storyID, workspaceID, updates)
}

// GetStatusCategory returns the category for a given status ID
func (r *repo) GetStatusCategory(ctx context.Context, statusID string) (string, error) {
	ctx, span := web.AddSpan(ctx, "storiesrepo.GetStatusCategory")
	defer span.End()

	query := `
		SELECT category 
		FROM statuses 
		WHERE status_id = :status_id
	`

	params := map[string]any{
		"status_id": statusID,
	}

	var category string
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, "failed to prepare status category query", "error", err)
		return "", err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &category, params); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("status not found: %s", statusID)
		}
		r.log.Error(ctx, "failed to get status category", "statusID", statusID, "error", err)
		return "", err
	}

	return category, nil
}
