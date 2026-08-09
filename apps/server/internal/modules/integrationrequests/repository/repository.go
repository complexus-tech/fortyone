package integrationrequestsrepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Repo struct {
	log *logger.Logger
	db  *sqlx.DB
}

func New(log *logger.Logger, db *sqlx.DB) *Repo {
	return &Repo{log: log, db: db}
}

type requestRow struct {
	ID                        uuid.UUID      `db:"id"`
	WorkspaceID               uuid.UUID      `db:"workspace_id"`
	TeamID                    uuid.UUID      `db:"team_id"`
	Provider                  string         `db:"provider"`
	SourceType                string         `db:"source_type"`
	SourceExternalID          string         `db:"source_external_id"`
	SourceNumber              *int           `db:"source_number"`
	SourceURL                 *string        `db:"source_url"`
	Title                     string         `db:"title"`
	Description               *string        `db:"description"`
	StatusID                  *uuid.UUID     `db:"status_id"`
	Priority                  string         `db:"priority"`
	AssigneeID                *uuid.UUID     `db:"assignee_id"`
	EstimateValue             *int16         `db:"estimate_unit"`
	ObjectiveID               *uuid.UUID     `db:"objective_id"`
	KeyResultID               *uuid.UUID     `db:"key_result_id"`
	SprintID                  *uuid.UUID     `db:"sprint_id"`
	StartDate                 *time.Time     `db:"start_date"`
	EndDate                   *time.Time     `db:"end_date"`
	LabelIDs                  pq.StringArray `db:"label_ids"`
	Status                    string         `db:"status"`
	Metadata                  mapJSON        `db:"metadata"`
	AcceptedStoryID           *uuid.UUID     `db:"accepted_story_id"`
	AcceptedByUserID          *uuid.UUID     `db:"accepted_by_user_id"`
	AcceptedAt                sql.NullTime   `db:"accepted_at"`
	DeclinedByUserID          *uuid.UUID     `db:"declined_by_user_id"`
	DeclinedAt                sql.NullTime   `db:"declined_at"`
	AcceptanceState           string         `db:"acceptance_state"`
	AcceptanceStartedByUserID *uuid.UUID     `db:"acceptance_started_by_user_id"`
	AcceptanceStartedAt       sql.NullTime   `db:"acceptance_started_at"`
	CreatedByUserID           *uuid.UUID     `db:"created_by_user_id"`
	CreatedAt                 sql.NullTime   `db:"created_at"`
	UpdatedAt                 sql.NullTime   `db:"updated_at"`
}

type mapJSON map[string]any

func (m *mapJSON) Scan(src any) error {
	if src == nil {
		*m = map[string]any{}
		return nil
	}
	var raw []byte
	switch value := src.(type) {
	case []byte:
		raw = value
	case string:
		raw = []byte(value)
	default:
		raw = []byte("{}")
	}
	if len(raw) == 0 {
		raw = []byte("{}")
	}
	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return err
	}
	*m = parsed
	return nil
}

func (r *Repo) UpsertPending(ctx context.Context, input integrationrequests.CoreUpsertRequestInput) (integrationrequests.CoreIntegrationRequest, error) {
	metadata, err := json.Marshal(input.Metadata)
	if err != nil {
		return integrationrequests.CoreIntegrationRequest{}, err
	}
	var row requestRow
	query := `
		INSERT INTO integration_requests (
			workspace_id, team_id, provider, source_type, source_external_id, source_number,
			source_url, title, description, status_id, priority, assignee_id, estimate_unit,
			objective_id, key_result_id, sprint_id, start_date, end_date, label_ids, metadata, created_by_user_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, COALESCE(NULLIF($11, ''), 'No Priority'), $12, $13, $14, $15, $16, $17, $18, $19, CAST($20 AS jsonb), $21)
		ON CONFLICT (workspace_id, provider, source_type, source_external_id) DO UPDATE SET
			team_id = EXCLUDED.team_id,
			source_number = EXCLUDED.source_number,
			source_url = EXCLUDED.source_url,
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			status_id = COALESCE(integration_requests.status_id, EXCLUDED.status_id),
			priority = EXCLUDED.priority,
			assignee_id = COALESCE(integration_requests.assignee_id, EXCLUDED.assignee_id),
			estimate_unit = COALESCE(integration_requests.estimate_unit, EXCLUDED.estimate_unit),
			objective_id = COALESCE(integration_requests.objective_id, EXCLUDED.objective_id),
			key_result_id = COALESCE(integration_requests.key_result_id, EXCLUDED.key_result_id),
			sprint_id = COALESCE(integration_requests.sprint_id, EXCLUDED.sprint_id),
			start_date = COALESCE(integration_requests.start_date, EXCLUDED.start_date),
			end_date = COALESCE(integration_requests.end_date, EXCLUDED.end_date),
			label_ids = CASE WHEN cardinality(EXCLUDED.label_ids) > 0 THEN EXCLUDED.label_ids ELSE integration_requests.label_ids END,
			metadata = EXCLUDED.metadata,
			updated_at = NOW()
		WHERE integration_requests.status = 'pending'
		  AND integration_requests.acceptance_state = 'idle'
		RETURNING *
	`
	err = r.db.GetContext(
		ctx,
		&row,
		query,
		input.WorkspaceID,
		input.TeamID,
		input.Provider,
		input.SourceType,
		input.SourceExternalID,
		input.SourceNumber,
		input.SourceURL,
		input.Title,
		input.Description,
		input.StatusID,
		input.Priority,
		input.AssigneeID,
		input.EstimateValue,
		input.ObjectiveID,
		input.KeyResultID,
		input.SprintID,
		input.StartDate,
		input.EndDate,
		pq.Array(uuidStrings(input.LabelIDs)),
		string(metadata),
		input.CreatedByUserID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return r.GetByExternal(ctx, input.WorkspaceID, input.Provider, input.SourceType, input.SourceExternalID)
		}
		return integrationrequests.CoreIntegrationRequest{}, err
	}
	return toCore(row), nil
}

func (r *Repo) AuthorizeTeam(ctx context.Context, workspaceID, teamID, userID uuid.UUID) error {
	var allowed int
	err := r.db.GetContext(ctx, &allowed, `
		SELECT 1
		FROM teams t
		WHERE t.workspace_id = $1
		  AND t.team_id = $2
		  AND `+teamAccessPredicate("t.team_id", "t.workspace_id", "$3")+`
	`, workspaceID, teamID, userID)
	return err
}

func (r *Repo) UpdatePending(ctx context.Context, workspaceID, requestID, userID uuid.UUID, input integrationrequests.CoreUpdateRequestInput) (result integrationrequests.CoreIntegrationRequest, err error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return integrationrequests.CoreIntegrationRequest{}, fmt.Errorf("begin integration request update: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var current requestRow
	if err = tx.GetContext(ctx, &current, `
		SELECT *
		FROM integration_requests
		WHERE workspace_id = $1
		  AND id = $2
		  AND status = 'pending'
		  AND acceptance_state = 'idle'
		  AND `+teamAccessPredicate("integration_requests.team_id", "integration_requests.workspace_id", "$3")+`
		FOR UPDATE
	`, workspaceID, requestID, userID); err != nil {
		return integrationrequests.CoreIntegrationRequest{}, err
	}

	title := current.Title
	if input.Title != nil {
		title = strings.TrimSpace(*input.Title)
	}
	description := optionalValue(current.Description, input.Description)
	statusID := optionalValue(current.StatusID, input.StatusID)
	priority := current.Priority
	if input.Priority != nil {
		priority = strings.TrimSpace(*input.Priority)
	}
	assigneeID := optionalValue(current.AssigneeID, input.AssigneeID)
	estimateValue := optionalValue(current.EstimateValue, input.EstimateValue)
	objectiveID := optionalValue(current.ObjectiveID, input.ObjectiveID)
	keyResultID := optionalValue(current.KeyResultID, input.KeyResultID)
	sprintID := optionalValue(current.SprintID, input.SprintID)
	startDate := optionalValue(current.StartDate, input.StartDate)
	endDate := optionalValue(current.EndDate, input.EndDate)
	labelIDs, parseErr := requestLabelIDs(current.LabelIDs, input.LabelIDs)
	if parseErr != nil {
		return integrationrequests.CoreIntegrationRequest{}, parseErr
	}

	if err = validatePendingRequestProperties(ctx, tx, current.WorkspaceID, current.TeamID, statusID, assigneeID, objectiveID, keyResultID, sprintID, labelIDs, startDate, endDate); err != nil {
		return integrationrequests.CoreIntegrationRequest{}, err
	}

	var updated requestRow
	if err = tx.GetContext(ctx, &updated, `
		UPDATE integration_requests
		SET title = $4,
			description = $5,
			status_id = $6,
			priority = $7,
			assignee_id = $8,
			estimate_unit = $9,
			objective_id = $10,
			key_result_id = $11,
			sprint_id = $12,
			start_date = $13,
			end_date = $14,
			label_ids = $15,
			updated_at = NOW()
		WHERE workspace_id = $1
		  AND id = $2
		  AND status = 'pending'
		  AND acceptance_state = 'idle'
		  AND `+teamAccessPredicate("integration_requests.team_id", "integration_requests.workspace_id", "$3")+`
		RETURNING *
	`, workspaceID, requestID, userID, title, description, statusID, priority, assigneeID, estimateValue, objectiveID, keyResultID, sprintID, startDate, endDate, pq.Array(uuidStrings(labelIDs))); err != nil {
		return integrationrequests.CoreIntegrationRequest{}, err
	}
	if err = tx.Commit(); err != nil {
		return integrationrequests.CoreIntegrationRequest{}, fmt.Errorf("commit integration request update: %w", err)
	}
	return toCore(updated), nil
}

func (r *Repo) ListByTeam(ctx context.Context, workspaceID, teamID, userID uuid.UUID, filter integrationrequests.CoreListRequestsFilter) ([]integrationrequests.CoreIntegrationRequest, error) {
	query, args := teamRequestFilterQuery(workspaceID, teamID, userID, filter)
	query = `SELECT * ` + query + ` ORDER BY created_at DESC`
	if filter.PageSize > 0 {
		page := filter.Page
		if page <= 0 {
			page = 1
		}
		args = append(args, filter.PageSize, (page-1)*filter.PageSize)
		query += ` LIMIT $` + strconv.Itoa(len(args)-1) + ` OFFSET $` + strconv.Itoa(len(args))
	}

	var rows []requestRow
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	result := make([]integrationrequests.CoreIntegrationRequest, 0, len(rows))
	for _, row := range rows {
		result = append(result, toCore(row))
	}
	return result, nil
}

func (r *Repo) CountByTeam(ctx context.Context, workspaceID, teamID, userID uuid.UUID, filter integrationrequests.CoreListRequestsFilter) (int, error) {
	query, args := teamRequestFilterQuery(workspaceID, teamID, userID, filter)
	var count int
	if err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) `+query, args...); err != nil {
		return 0, err
	}
	return count, nil
}

func teamRequestFilterQuery(workspaceID, teamID, userID uuid.UUID, filter integrationrequests.CoreListRequestsFilter) (string, []any) {
	status := filter.Status
	if status == "" {
		status = integrationrequests.StatusPending
	}
	query := `FROM integration_requests
		WHERE workspace_id = $1
		  AND team_id = $2
		  AND status = $3
		  AND ` + teamAccessPredicate("integration_requests.team_id", "integration_requests.workspace_id", "$4") + `
	`
	args := []any{workspaceID, teamID, status, userID}
	if search := strings.TrimSpace(filter.Search); search != "" {
		args = append(args, "%"+search+"%")
		placeholder := `$` + strconv.Itoa(len(args))
		query += ` AND (
			title ILIKE ` + placeholder + `
			OR COALESCE(description, '') ILIKE ` + placeholder + `
			OR source_external_id ILIKE ` + placeholder + `
			OR COALESCE(source_number::text, '') ILIKE ` + placeholder + `
		)`
	}
	if filter.Provider != "" {
		args = append(args, filter.Provider)
		query += ` AND provider = $` + strconv.Itoa(len(args))
	}
	if filter.Priority != "" {
		args = append(args, filter.Priority)
		query += ` AND priority = $` + strconv.Itoa(len(args))
	}
	if filter.AssigneeID != nil {
		args = append(args, *filter.AssigneeID)
		query += ` AND assignee_id = $` + strconv.Itoa(len(args))
	}
	if filter.CreatedAfter != nil {
		args = append(args, *filter.CreatedAfter)
		query += ` AND created_at >= $` + strconv.Itoa(len(args))
	}
	if filter.CreatedBefore != nil {
		args = append(args, *filter.CreatedBefore)
		query += ` AND created_at <= $` + strconv.Itoa(len(args))
	}
	return query, args
}

func (r *Repo) GetForUser(ctx context.Context, workspaceID, requestID, userID uuid.UUID) (integrationrequests.CoreIntegrationRequest, error) {
	var row requestRow
	err := r.db.GetContext(ctx, &row, `
		SELECT *
		FROM integration_requests
		WHERE workspace_id = $1
		  AND id = $2
		  AND `+teamAccessPredicate("integration_requests.team_id", "integration_requests.workspace_id", "$3")+`
	`, workspaceID, requestID, userID)
	if err != nil {
		return integrationrequests.CoreIntegrationRequest{}, err
	}
	return toCore(row), nil
}

// Get is reserved for trusted provider callbacks that do not have a FortyOne
// actor in their execution context. User-facing reads must use GetForUser.
func (r *Repo) Get(ctx context.Context, workspaceID, requestID uuid.UUID) (integrationrequests.CoreIntegrationRequest, error) {
	var row requestRow
	if err := r.db.GetContext(ctx, &row, `
		SELECT *
		FROM integration_requests
		WHERE workspace_id = $1 AND id = $2
	`, workspaceID, requestID); err != nil {
		return integrationrequests.CoreIntegrationRequest{}, err
	}
	return toCore(row), nil
}

func (r *Repo) GetByExternal(ctx context.Context, workspaceID uuid.UUID, provider, sourceType, sourceExternalID string) (integrationrequests.CoreIntegrationRequest, error) {
	var row requestRow
	err := r.db.GetContext(ctx, &row, `
		SELECT *
		FROM integration_requests
		WHERE workspace_id = $1 AND provider = $2 AND source_type = $3 AND source_external_id = $4
	`, workspaceID, provider, sourceType, sourceExternalID)
	if err != nil {
		return integrationrequests.CoreIntegrationRequest{}, err
	}
	return toCore(row), nil
}

func (r *Repo) FindFirstStatusByCategory(ctx context.Context, teamID uuid.UUID, category string) (*uuid.UUID, error) {
	var statusID uuid.UUID
	err := r.db.GetContext(ctx, &statusID, `SELECT status_id FROM statuses WHERE team_id = $1 AND category = $2 ORDER BY order_index ASC LIMIT 1`, teamID, category)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &statusID, nil
}

// ReserveAcceptance durably fences a pending request before story creation.
// Repeated calls return the existing reservation so a crashed conversion can
// resume with the original actor and the story creation idempotency key.
func (r *Repo) ReserveAcceptance(ctx context.Context, workspaceID, requestID, userID uuid.UUID) (result integrationrequests.CoreIntegrationRequest, err error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return integrationrequests.CoreIntegrationRequest{}, fmt.Errorf("begin integration request acceptance reservation: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var current requestRow
	if err = tx.GetContext(ctx, &current, `
		SELECT *
		FROM integration_requests
		WHERE workspace_id = $1
		  AND id = $2
		  AND status = 'pending'
		  AND `+teamAccessPredicate("integration_requests.team_id", "integration_requests.workspace_id", "$3")+`
		FOR UPDATE
	`, workspaceID, requestID, userID); err != nil {
		return integrationrequests.CoreIntegrationRequest{}, err
	}

	if current.AcceptanceState == integrationrequests.AcceptanceStateReserved {
		if err = tx.Commit(); err != nil {
			return integrationrequests.CoreIntegrationRequest{}, fmt.Errorf("commit existing integration request acceptance reservation: %w", err)
		}
		return toCore(current), nil
	}
	if current.AcceptanceState != integrationrequests.AcceptanceStateIdle {
		return integrationrequests.CoreIntegrationRequest{}, integrationrequests.ErrRequestNotPending
	}

	labelIDs, parseErr := requestLabelIDs(current.LabelIDs, nil)
	if parseErr != nil {
		return integrationrequests.CoreIntegrationRequest{}, parseErr
	}
	if err = validatePendingRequestProperties(
		ctx,
		tx,
		current.WorkspaceID,
		current.TeamID,
		current.StatusID,
		current.AssigneeID,
		current.ObjectiveID,
		current.KeyResultID,
		current.SprintID,
		labelIDs,
		current.StartDate,
		current.EndDate,
	); err != nil {
		return integrationrequests.CoreIntegrationRequest{}, err
	}

	var reserved requestRow
	if err = tx.GetContext(ctx, &reserved, `
		UPDATE integration_requests
		SET acceptance_state = 'reserved',
			acceptance_started_by_user_id = $3,
			acceptance_started_at = NOW(),
			updated_at = NOW()
		WHERE workspace_id = $1
		  AND id = $2
		  AND status = 'pending'
		  AND acceptance_state = 'idle'
		  AND `+teamAccessPredicate("integration_requests.team_id", "integration_requests.workspace_id", "$3")+`
		RETURNING *
	`, workspaceID, requestID, userID); err != nil {
		return integrationrequests.CoreIntegrationRequest{}, err
	}
	if err = tx.Commit(); err != nil {
		return integrationrequests.CoreIntegrationRequest{}, fmt.Errorf("commit integration request acceptance reservation: %w", err)
	}
	return toCore(reserved), nil
}

func (r *Repo) MarkAccepted(ctx context.Context, workspaceID, requestID, storyID, acceptedByUserID uuid.UUID) (integrationrequests.CoreIntegrationRequest, error) {
	var row requestRow
	err := r.db.GetContext(ctx, &row, `
		UPDATE integration_requests
		SET status = 'accepted',
			accepted_story_id = $3,
			accepted_by_user_id = $4,
			accepted_at = NOW(),
			acceptance_state = 'idle',
			acceptance_started_by_user_id = NULL,
			acceptance_started_at = NULL,
			updated_at = NOW()
		WHERE workspace_id = $1
		  AND id = $2
		  AND status = 'pending'
		  AND acceptance_state = 'reserved'
		  AND acceptance_started_by_user_id = $4
		RETURNING *
	`, workspaceID, requestID, storyID, acceptedByUserID)
	if err != nil {
		return integrationrequests.CoreIntegrationRequest{}, err
	}
	return toCore(row), nil
}

func (r *Repo) MarkDeclined(ctx context.Context, workspaceID, requestID, declinedByUserID uuid.UUID) (integrationrequests.CoreIntegrationRequest, error) {
	var row requestRow
	err := r.db.GetContext(ctx, &row, `
		UPDATE integration_requests
		SET status = 'declined',
			declined_by_user_id = $3,
			declined_at = NOW(),
			updated_at = NOW()
		WHERE workspace_id = $1
		  AND id = $2
		  AND status = 'pending'
		  AND acceptance_state = 'idle'
		  AND `+teamAccessPredicate("integration_requests.team_id", "integration_requests.workspace_id", "$3")+`
		RETURNING *
	`, workspaceID, requestID, declinedByUserID)
	if err != nil {
		return integrationrequests.CoreIntegrationRequest{}, err
	}
	return toCore(row), nil
}

func teamAccessPredicate(teamExpression, workspaceExpression, userPlaceholder string) string {
	return fmt.Sprintf(`(
		EXISTS (
			SELECT 1
			FROM team_members request_team_member
			WHERE request_team_member.team_id = %s
			  AND request_team_member.user_id = %s
		)
		OR EXISTS (
			SELECT 1
			FROM workspace_members request_workspace_member
			WHERE request_workspace_member.workspace_id = %s
			  AND request_workspace_member.user_id = %s
			  AND request_workspace_member.role = 'admin'
		)
	)`, teamExpression, userPlaceholder, workspaceExpression, userPlaceholder)
}

func toCore(row requestRow) integrationrequests.CoreIntegrationRequest {
	result := integrationrequests.CoreIntegrationRequest{
		ID:                        row.ID,
		WorkspaceID:               row.WorkspaceID,
		TeamID:                    row.TeamID,
		Provider:                  row.Provider,
		SourceType:                row.SourceType,
		SourceExternalID:          row.SourceExternalID,
		SourceNumber:              row.SourceNumber,
		SourceURL:                 row.SourceURL,
		Title:                     row.Title,
		Description:               row.Description,
		StatusID:                  row.StatusID,
		Priority:                  row.Priority,
		AssigneeID:                row.AssigneeID,
		EstimateValue:             row.EstimateValue,
		ObjectiveID:               row.ObjectiveID,
		KeyResultID:               row.KeyResultID,
		SprintID:                  row.SprintID,
		StartDate:                 row.StartDate,
		EndDate:                   row.EndDate,
		LabelIDs:                  parseUUIDStrings(row.LabelIDs),
		Status:                    row.Status,
		Metadata:                  map[string]any(row.Metadata),
		AcceptedStoryID:           row.AcceptedStoryID,
		AcceptedByUserID:          row.AcceptedByUserID,
		DeclinedByUserID:          row.DeclinedByUserID,
		AcceptanceState:           row.AcceptanceState,
		AcceptanceStartedByUserID: row.AcceptanceStartedByUserID,
		CreatedByUserID:           row.CreatedByUserID,
	}
	result.AcceptedAt = nullTimePtr(row.AcceptedAt)
	result.DeclinedAt = nullTimePtr(row.DeclinedAt)
	result.AcceptanceStartedAt = nullTimePtr(row.AcceptanceStartedAt)
	if row.CreatedAt.Valid {
		result.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		result.UpdatedAt = row.UpdatedAt.Time
	}
	if result.Metadata == nil {
		result.Metadata = map[string]any{}
	}
	return result
}

func uuidStrings(values []uuid.UUID) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value != uuid.Nil {
			result = append(result, value.String())
		}
	}
	return result
}

func parseUUIDStrings(values []string) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		parsed, err := uuid.Parse(value)
		if err == nil && parsed != uuid.Nil {
			result = append(result, parsed)
		}
	}
	return result
}

func nullableUUIDArray(values *[]uuid.UUID) any {
	if values == nil {
		return nil
	}
	return pq.Array(uuidStrings(*values))
}

func optionalValue[T any](current *T, patch integrationrequests.OptionalValue[T]) *T {
	if patch.Set {
		return patch.Value
	}
	return current
}

func requestLabelIDs(current pq.StringArray, patch *[]uuid.UUID) ([]uuid.UUID, error) {
	if patch == nil {
		result := make([]uuid.UUID, 0, len(current))
		seen := make(map[uuid.UUID]struct{}, len(current))
		for _, raw := range current {
			labelID, err := uuid.Parse(raw)
			if err != nil || labelID == uuid.Nil {
				return nil, fmt.Errorf("%w: stored label id is invalid", integrationrequests.ErrInvalidRequestProperty)
			}
			if _, exists := seen[labelID]; exists {
				continue
			}
			seen[labelID] = struct{}{}
			result = append(result, labelID)
		}
		return result, nil
	}

	result := make([]uuid.UUID, 0, len(*patch))
	seen := make(map[uuid.UUID]struct{}, len(*patch))
	for _, labelID := range *patch {
		if labelID == uuid.Nil {
			return nil, fmt.Errorf("%w: label id is required", integrationrequests.ErrInvalidRequestProperty)
		}
		if _, exists := seen[labelID]; exists {
			continue
		}
		seen[labelID] = struct{}{}
		result = append(result, labelID)
	}
	return result, nil
}

func validatePendingRequestProperties(
	ctx context.Context,
	tx *sqlx.Tx,
	workspaceID uuid.UUID,
	teamID uuid.UUID,
	statusID *uuid.UUID,
	assigneeID *uuid.UUID,
	objectiveID *uuid.UUID,
	keyResultID *uuid.UUID,
	sprintID *uuid.UUID,
	labelIDs []uuid.UUID,
	startDate *time.Time,
	endDate *time.Time,
) error {
	if startDate != nil && endDate != nil && startDate.After(*endDate) {
		return fmt.Errorf("%w: start date must not be after deadline", integrationrequests.ErrInvalidRequestProperty)
	}

	if statusID != nil {
		if err := requirePendingReference(ctx, tx, "status", `
			SELECT 1
			FROM statuses
			WHERE status_id = $1 AND workspace_id = $2 AND team_id = $3
		`, *statusID, workspaceID, teamID); err != nil {
			return err
		}
	}
	if assigneeID != nil {
		if err := requirePendingReference(ctx, tx, "assignee", `
			SELECT 1
			FROM team_members tm
			JOIN users u ON u.user_id = tm.user_id AND u.is_active = true
			JOIN workspace_members wm ON wm.user_id = tm.user_id AND wm.workspace_id = $3
			WHERE tm.team_id = $2 AND tm.user_id = $1
		`, *assigneeID, teamID, workspaceID); err != nil {
			return err
		}
	}
	if objectiveID != nil {
		if err := requirePendingReference(ctx, tx, "objective", `
			SELECT 1
			FROM objectives
			WHERE objective_id = $1 AND workspace_id = $2 AND team_id = $3
		`, *objectiveID, workspaceID, teamID); err != nil {
			return err
		}
	}
	if keyResultID != nil {
		if objectiveID == nil {
			return fmt.Errorf("%w: key result requires an objective", integrationrequests.ErrInvalidRequestProperty)
		}
		if err := requirePendingReference(ctx, tx, "key result", `
			SELECT 1
			FROM key_results kr
			JOIN objectives o ON o.objective_id = kr.objective_id
			WHERE kr.id = $1
			  AND kr.objective_id = $2
			  AND o.workspace_id = $3
			  AND o.team_id = $4
		`, *keyResultID, *objectiveID, workspaceID, teamID); err != nil {
			return err
		}
	}
	if sprintID != nil {
		if err := requirePendingReference(ctx, tx, "sprint", `
			SELECT 1
			FROM sprints
			WHERE sprint_id = $1 AND workspace_id = $2 AND team_id = $3
		`, *sprintID, workspaceID, teamID); err != nil {
			return err
		}
	}
	if len(labelIDs) > 0 {
		var count int
		if err := tx.GetContext(ctx, &count, `
			SELECT COUNT(*)
			FROM labels
			WHERE workspace_id = $1
			  AND (team_id = $2 OR team_id IS NULL)
			  AND label_id = ANY($3)
		`, workspaceID, teamID, pq.Array(uuidStrings(labelIDs))); err != nil {
			return fmt.Errorf("validate labels: %w", err)
		}
		if count != len(labelIDs) {
			return fmt.Errorf("%w: one or more labels are not available to the team", integrationrequests.ErrInvalidRequestProperty)
		}
	}
	return nil
}

func requirePendingReference(ctx context.Context, tx *sqlx.Tx, name, query string, args ...any) error {
	var exists int
	if err := tx.GetContext(ctx, &exists, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: %s is not available to the team", integrationrequests.ErrInvalidRequestProperty, name)
		}
		return fmt.Errorf("validate %s: %w", name, err)
	}
	return nil
}

func nullTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}
