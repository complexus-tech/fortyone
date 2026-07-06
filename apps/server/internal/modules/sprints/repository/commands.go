package sprintsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (r *repo) Create(ctx context.Context, sprint sprints.CoreNewSprint, actorID *uuid.UUID) (sprints.CoreSprint, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.sprints.Create")
	defer span.End()

	var s dbSprint
	query := `
		INSERT INTO sprints (
			name,
			goal,
			objective_id,
			team_id,
			workspace_id,
			start_date,
			end_date
		) VALUES (
			:name,
			:goal,
			:objective_id,
			:team_id,
			:workspace_id,
			:start_date,
			:end_date
		)
		RETURNING
			sprint_id,
			name,
			goal,
			objective_id,
			team_id,
			workspace_id,
			start_date,
			end_date,
			created_at,
			updated_at,
			0 as total_stories,
			0 as cancelled_stories,
			0 as completed_stories,
			0 as started_stories,
			0 as unstarted_stories,
			0 as backlog_stories
	`

	params := map[string]any{
		"name":         sprint.Name,
		"goal":         sprint.Goal,
		"objective_id": sprint.Objective,
		"team_id":      sprint.Team,
		"workspace_id": sprint.Workspace,
		"start_date":   sprint.StartDate,
		"end_date":     sprint.EndDate,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return sprints.CoreSprint{}, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &s, params); err != nil {
		errMsg := fmt.Sprintf("Failed to create sprint: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create sprint"), trace.WithAttributes(attribute.String("error", errMsg)))
		return sprints.CoreSprint{}, err
	}

	r.log.Info(ctx, "sprint created successfully.")
	span.AddEvent("sprint created.", trace.WithAttributes(
		attribute.String("sprint.id", s.ID.String()),
		attribute.String("workspace.id", s.Workspace.String()),
	))

	r.recordSprintAuditEvent(ctx, sprint.Workspace, sprint.Team, actorID, "sprint", &s.ID, "sprint.created", map[string]any{
		"name":       sprint.Name,
		"start_date": sprint.StartDate.Format("2006-01-02"),
		"end_date":   sprint.EndDate.Format("2006-01-02"),
	})

	return toCoreSprint(s), nil
}

func (r *repo) Update(ctx context.Context, sprintID uuid.UUID, workspaceID uuid.UUID, updates sprints.CoreUpdateSprint, actorID *uuid.UUID) (sprints.CoreSprint, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.sprints.Update")
	defer span.End()

	query := "UPDATE sprints SET "
	var setClauses []string
	params := map[string]any{
		"sprint_id":    sprintID,
		"workspace_id": workspaceID,
	}

	if updates.Name != nil {
		setClauses = append(setClauses, "name = :name")
		params["name"] = *updates.Name
	}
	if updates.Goal != nil {
		setClauses = append(setClauses, "goal = :goal")
		params["goal"] = *updates.Goal
	}
	if updates.Objective != nil {
		setClauses = append(setClauses, "objective_id = :objective_id")
		params["objective_id"] = *updates.Objective
	}
	if updates.StartDate != nil {
		setClauses = append(setClauses, "start_date = :start_date")
		params["start_date"] = *updates.StartDate
	}
	if updates.EndDate != nil {
		setClauses = append(setClauses, "end_date = :end_date")
		params["end_date"] = *updates.EndDate
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	query += strings.Join(setClauses, ", ")
	query += " WHERE sprint_id = :sprint_id AND workspace_id = :workspace_id;"

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return sprints.CoreSprint{}, err
	}
	defer stmt.Close()

	if _, err := stmt.ExecContext(ctx, params); err != nil {
		errMsg := fmt.Sprintf("Failed to update sprint: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to update sprint"), trace.WithAttributes(attribute.String("error", errMsg)))
		return sprints.CoreSprint{}, err
	}

	// Get the updated sprint
	var s dbSprint
	getQuery := `
		WITH story_stats AS (
			SELECT 
				s.sprint_id,
				COUNT(*) as total,
				COUNT(CASE WHEN st.category = 'cancelled' THEN 1 END) as cancelled,
				COUNT(CASE WHEN st.category = 'completed' THEN 1 END) as completed,
				COUNT(CASE WHEN st.category = 'started' THEN 1 END) as started,
				COUNT(CASE WHEN st.category = 'unstarted' THEN 1 END) as unstarted,
				COUNT(CASE WHEN st.category = 'backlog' THEN 1 END) as backlog
			FROM sprints s
			LEFT JOIN stories str ON s.sprint_id = str.sprint_id
			LEFT JOIN statuses st ON str.status_id = st.status_id
			WHERE str.deleted_at IS NULL AND str.archived_at IS NULL
			GROUP BY s.sprint_id
		)
		SELECT
			s.sprint_id,
			s.name,
			s.goal,
			s.objective_id,
			s.team_id,
			s.workspace_id,
			s.start_date,
			s.end_date,
			s.created_at,
			s.updated_at,
			COALESCE(ss.total, 0) as total_stories,
			COALESCE(ss.cancelled, 0) as cancelled_stories,
			COALESCE(ss.completed, 0) as completed_stories,
			COALESCE(ss.started, 0) as started_stories,
			COALESCE(ss.unstarted, 0) as unstarted_stories,
			COALESCE(ss.backlog, 0) as backlog_stories
		FROM sprints s
		LEFT JOIN story_stats ss ON s.sprint_id = ss.sprint_id
		WHERE s.sprint_id = :sprint_id AND s.workspace_id = :workspace_id
	`

	stmt, err = r.db.PrepareNamedContext(ctx, getQuery)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare get statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare get statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return sprints.CoreSprint{}, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &s, params); err != nil {
		errMsg := fmt.Sprintf("Failed to get updated sprint: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get updated sprint"), trace.WithAttributes(attribute.String("error", errMsg)))
		return sprints.CoreSprint{}, err
	}

	r.log.Info(ctx, "sprint updated successfully.")
	span.AddEvent("sprint updated.", trace.WithAttributes(
		attribute.String("sprint.id", s.ID.String()),
		attribute.String("workspace.id", s.Workspace.String()),
	))

	r.recordSprintAuditEvent(ctx, workspaceID, s.Team, actorID, "sprint", &sprintID, "sprint.updated", map[string]any{
		"name_changed":       updates.Name != nil,
		"goal_changed":       updates.Goal != nil,
		"objective_changed":  updates.Objective != nil,
		"start_date_changed": updates.StartDate != nil,
		"end_date_changed":   updates.EndDate != nil,
	})

	return toCoreSprint(s), nil
}

func (r *repo) Delete(ctx context.Context, sprintID uuid.UUID, workspaceID uuid.UUID, actorID *uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.sprints.Delete")
	defer span.End()

	query := `
		DELETE FROM sprints
		WHERE sprint_id = :sprint_id
		AND workspace_id = :workspace_id
		RETURNING team_id, name, CAST(start_date AS text), CAST(end_date AS text)
	`

	params := map[string]any{
		"sprint_id":    sprintID,
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	var deleted struct {
		TeamID    uuid.UUID `db:"team_id"`
		Name      string    `db:"name"`
		StartDate string    `db:"start_date"`
		EndDate   string    `db:"end_date"`
	}
	if err := stmt.GetContext(ctx, &deleted, params); err != nil {
		errMsg := fmt.Sprintf("Failed to delete sprint: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to delete sprint"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	r.log.Info(ctx, "sprint deleted successfully.")
	span.AddEvent("sprint deleted.", trace.WithAttributes(
		attribute.String("sprint.id", sprintID.String()),
		attribute.String("workspace.id", workspaceID.String()),
	))

	r.recordSprintAuditEvent(ctx, workspaceID, deleted.TeamID, actorID, "sprint", &sprintID, "sprint.deleted", map[string]any{
		"name":       deleted.Name,
		"start_date": deleted.StartDate,
		"end_date":   deleted.EndDate,
	})

	return nil
}

func (r *repo) recordSprintAuditEvent(ctx context.Context, workspaceID, teamID uuid.UUID, actorID *uuid.UUID, entityType string, entityID *uuid.UUID, eventType string, metadata map[string]any) {
	data, err := json.Marshal(metadata)
	if err != nil {
		r.log.Error(ctx, "failed to marshal sprint audit metadata", "error", err, "event_type", eventType)
		return
	}

	query := `
		INSERT INTO audit_events (
			workspace_id,
			team_id,
			actor_type,
			actor_id,
			entity_type,
			entity_id,
			event_type,
			metadata
		) VALUES (
			:workspace_id,
			:team_id,
			:user_actor_type,
			:actor_id,
			:entity_type,
			:entity_id,
			:event_type,
			CAST(:metadata AS jsonb)
		)
	`

	params := map[string]any{
		"workspace_id":    workspaceID,
		"team_id":         teamID,
		"user_actor_type": "user",
		"actor_id":        actorID,
		"entity_type":     entityType,
		"entity_id":       entityID,
		"event_type":      eventType,
		"metadata":        string(data),
	}

	if _, err := r.db.NamedExecContext(ctx, query, params); err != nil {
		r.log.Error(ctx, "failed to record sprint audit event", "error", err, "event_type", eventType, "entity_id", entityID)
	}
}
