package integrationrequestsrepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repo struct {
	log *logger.Logger
	db  *sqlx.DB
}

func New(log *logger.Logger, db *sqlx.DB) *Repo {
	return &Repo{log: log, db: db}
}

type requestRow struct {
	ID               uuid.UUID    `db:"id"`
	WorkspaceID      uuid.UUID    `db:"workspace_id"`
	TeamID           uuid.UUID    `db:"team_id"`
	Provider         string       `db:"provider"`
	SourceType       string       `db:"source_type"`
	SourceExternalID string       `db:"source_external_id"`
	SourceNumber     *int         `db:"source_number"`
	SourceURL        *string      `db:"source_url"`
	Title            string       `db:"title"`
	Description      *string      `db:"description"`
	StatusID         *uuid.UUID   `db:"status_id"`
	Priority         string       `db:"priority"`
	AssigneeID       *uuid.UUID   `db:"assignee_id"`
	Status           string       `db:"status"`
	Metadata         mapJSON      `db:"metadata"`
	AcceptedStoryID  *uuid.UUID   `db:"accepted_story_id"`
	AcceptedByUserID *uuid.UUID   `db:"accepted_by_user_id"`
	AcceptedAt       sql.NullTime `db:"accepted_at"`
	DeclinedByUserID *uuid.UUID   `db:"declined_by_user_id"`
	DeclinedAt       sql.NullTime `db:"declined_at"`
	CreatedByUserID  *uuid.UUID   `db:"created_by_user_id"`
	CreatedAt        sql.NullTime `db:"created_at"`
	UpdatedAt        sql.NullTime `db:"updated_at"`
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
			source_url, title, description, status_id, priority, assignee_id, metadata, created_by_user_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, COALESCE(NULLIF($11, ''), 'No Priority'), $12, $13::jsonb, $14)
		ON CONFLICT (workspace_id, provider, source_type, source_external_id) DO UPDATE SET
			team_id = EXCLUDED.team_id,
			source_number = EXCLUDED.source_number,
			source_url = EXCLUDED.source_url,
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			status_id = COALESCE(integration_requests.status_id, EXCLUDED.status_id),
			priority = COALESCE(NULLIF(integration_requests.priority, ''), EXCLUDED.priority),
			assignee_id = COALESCE(integration_requests.assignee_id, EXCLUDED.assignee_id),
			metadata = EXCLUDED.metadata,
			updated_at = NOW()
		WHERE integration_requests.status = 'pending'
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

func (r *Repo) UpdatePending(ctx context.Context, workspaceID, requestID uuid.UUID, input integrationrequests.CoreUpdateRequestInput) (integrationrequests.CoreIntegrationRequest, error) {
	var row requestRow
	err := r.db.GetContext(ctx, &row, `
		UPDATE integration_requests
		SET title = COALESCE($3, title),
			description = COALESCE($4, description),
			status_id = COALESCE($5, status_id),
			priority = COALESCE(NULLIF($6, ''), priority),
			assignee_id = COALESCE($7, assignee_id),
			updated_at = NOW()
		WHERE workspace_id = $1 AND id = $2 AND status = 'pending'
		RETURNING *
	`, workspaceID, requestID, input.Title, input.Description, input.StatusID, input.Priority, input.AssigneeID)
	if err != nil {
		return integrationrequests.CoreIntegrationRequest{}, err
	}
	return toCore(row), nil
}

func (r *Repo) ListByTeam(ctx context.Context, workspaceID, teamID uuid.UUID, filter integrationrequests.CoreListRequestsFilter) ([]integrationrequests.CoreIntegrationRequest, error) {
	status := filter.Status
	if status == "" {
		status = integrationrequests.StatusPending
	}
	var rows []requestRow
	err := r.db.SelectContext(ctx, &rows, `
		SELECT *
		FROM integration_requests
		WHERE workspace_id = $1 AND team_id = $2 AND status = $3
		ORDER BY created_at DESC
	`, workspaceID, teamID, status)
	if err != nil {
		return nil, err
	}
	result := make([]integrationrequests.CoreIntegrationRequest, 0, len(rows))
	for _, row := range rows {
		result = append(result, toCore(row))
	}
	return result, nil
}

func (r *Repo) Get(ctx context.Context, workspaceID, requestID uuid.UUID) (integrationrequests.CoreIntegrationRequest, error) {
	var row requestRow
	err := r.db.GetContext(ctx, &row, `SELECT * FROM integration_requests WHERE workspace_id = $1 AND id = $2`, workspaceID, requestID)
	if err != nil {
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

func (r *Repo) MarkAccepted(ctx context.Context, workspaceID, requestID, storyID, acceptedByUserID uuid.UUID) (integrationrequests.CoreIntegrationRequest, error) {
	var row requestRow
	err := r.db.GetContext(ctx, &row, `
		UPDATE integration_requests
		SET status = 'accepted',
			accepted_story_id = $3,
			accepted_by_user_id = $4,
			accepted_at = NOW(),
			updated_at = NOW()
		WHERE workspace_id = $1 AND id = $2 AND status = 'pending'
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
		WHERE workspace_id = $1 AND id = $2 AND status = 'pending'
		RETURNING *
	`, workspaceID, requestID, declinedByUserID)
	if err != nil {
		return integrationrequests.CoreIntegrationRequest{}, err
	}
	return toCore(row), nil
}

func toCore(row requestRow) integrationrequests.CoreIntegrationRequest {
	result := integrationrequests.CoreIntegrationRequest{
		ID:               row.ID,
		WorkspaceID:      row.WorkspaceID,
		TeamID:           row.TeamID,
		Provider:         row.Provider,
		SourceType:       row.SourceType,
		SourceExternalID: row.SourceExternalID,
		SourceNumber:     row.SourceNumber,
		SourceURL:        row.SourceURL,
		Title:            row.Title,
		Description:      row.Description,
		StatusID:         row.StatusID,
		Priority:         row.Priority,
		AssigneeID:       row.AssigneeID,
		Status:           row.Status,
		Metadata:         map[string]any(row.Metadata),
		AcceptedStoryID:  row.AcceptedStoryID,
		AcceptedByUserID: row.AcceptedByUserID,
		DeclinedByUserID: row.DeclinedByUserID,
		CreatedByUserID:  row.CreatedByUserID,
	}
	result.AcceptedAt = nullTimePtr(row.AcceptedAt)
	result.DeclinedAt = nullTimePtr(row.DeclinedAt)
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

func nullTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}
