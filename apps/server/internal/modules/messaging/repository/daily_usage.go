package messagingrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const DefaultDailyWorkspaceTokenLimit int64 = 1_000_000

var ErrDailyWorkspaceTokenLimit = errors.New("daily workspace assistant token limit reached")

type DailyTokenLimitError struct {
	WorkspaceID uuid.UUID
	Used        int64
	Limit       int64
}

func (e *DailyTokenLimitError) Error() string {
	return fmt.Sprintf("workspace %s used %d of %d daily assistant tokens: %v", e.WorkspaceID, e.Used, e.Limit, ErrDailyWorkspaceTokenLimit)
}

func (e *DailyTokenLimitError) Unwrap() error {
	return ErrDailyWorkspaceTokenLimit
}

type DailyUsageSnapshot struct {
	InputTokens  int64
	OutputTokens int64
	TotalTokens  int64
	RequestCount int64
	Limit        int64
	Remaining    int64
	Allowed      bool
}

// DailyUsageRecordInput identifies one assistant execution within a provider
// event. AttemptCount must be the durable inbound receipt attempt returned by
// StartInboundEvent. Replaying the same attempt is idempotent; a later attempt
// can record additional Responses Usage legitimately incurred by a retry.
type DailyUsageRecordInput struct {
	InboundEventID      uuid.UUID
	WorkspaceID         uuid.UUID
	Provider            string
	ExternalWorkspaceID string
	ExternalEventID     string
	AttemptCount        int
	Usage               messaging.Usage
}

type dailyUsageQuerier interface {
	GetContext(ctx context.Context, destination any, query string, args ...any) error
}

type dailyUsageTransaction interface {
	dailyUsageQuerier
	Commit() error
	Rollback() error
}

type dailyUsageDatabase interface {
	dailyUsageQuerier
	Begin(ctx context.Context) (dailyUsageTransaction, error)
}

type sqlxDailyUsageDatabase struct {
	db *sqlx.DB
}

func (d sqlxDailyUsageDatabase) GetContext(ctx context.Context, destination any, query string, args ...any) error {
	return d.db.GetContext(ctx, destination, query, args...)
}

func (d sqlxDailyUsageDatabase) Begin(ctx context.Context) (dailyUsageTransaction, error) {
	return d.db.BeginTxx(ctx, nil)
}

// DailyUsageRepository persists the workspace-wide messaging assistant budget.
// Slack, Teams, and future provider adapters intentionally share the same UTC
// daily ceiling so adding a provider cannot bypass the cost control.
type DailyUsageRepository struct {
	db dailyUsageDatabase
}

func NewDailyUsageRepository(db *sqlx.DB) *DailyUsageRepository {
	if db == nil {
		return &DailyUsageRepository{}
	}
	return &DailyUsageRepository{db: sqlxDailyUsageDatabase{db: db}}
}

// Check returns ErrDailyWorkspaceTokenLimit when the current UTC-day total has
// reached the configured ceiling. A zero limit selects the production default.
func (r *DailyUsageRepository) Check(ctx context.Context, workspaceID uuid.UUID, limit int64) (DailyUsageSnapshot, error) {
	if err := validateDailyUsageRepository(r, workspaceID); err != nil {
		return DailyUsageSnapshot{}, err
	}
	limit, err := normalizedDailyTokenLimit(limit)
	if err != nil {
		return DailyUsageSnapshot{}, err
	}
	row, err := r.load(ctx, workspaceID)
	if err != nil {
		return DailyUsageSnapshot{}, err
	}
	snapshot := dailyUsageSnapshot(row, limit)
	if !snapshot.Allowed {
		return snapshot, &DailyTokenLimitError{
			WorkspaceID: workspaceID,
			Used:        snapshot.TotalTokens,
			Limit:       limit,
		}
	}
	return snapshot, nil
}

// Record atomically claims an execution ledger key and adds aggregate Responses
// Usage only when that key is new. Callers must invoke it for both successful
// responses and errors containing partial tool-loop usage. It returns the new
// total but does not reject usage already incurred by the provider.
func (r *DailyUsageRepository) Record(ctx context.Context, input DailyUsageRecordInput, limit int64) (DailyUsageSnapshot, error) {
	input.Provider = strings.ToLower(strings.TrimSpace(input.Provider))
	input.ExternalWorkspaceID = strings.TrimSpace(input.ExternalWorkspaceID)
	input.ExternalEventID = strings.TrimSpace(input.ExternalEventID)
	if err := validateDailyUsageRepository(r, input.WorkspaceID); err != nil {
		return DailyUsageSnapshot{}, err
	}
	if input.Provider == "" || input.ExternalWorkspaceID == "" || input.ExternalEventID == "" {
		return DailyUsageSnapshot{}, errors.New("messaging assistant usage provider, external workspace, and external event are required")
	}
	if input.InboundEventID == uuid.Nil {
		return DailyUsageSnapshot{}, errors.New("messaging assistant usage inbound event is required")
	}
	if input.AttemptCount < 1 {
		return DailyUsageSnapshot{}, errors.New("messaging assistant usage attempt count must be positive")
	}
	limit, err := normalizedDailyTokenLimit(limit)
	if err != nil {
		return DailyUsageSnapshot{}, err
	}
	if err := validateUsage(input.Usage); err != nil {
		return DailyUsageSnapshot{}, err
	}
	if input.Usage.TotalTokens == 0 {
		row, loadErr := r.load(ctx, input.WorkspaceID)
		if loadErr != nil {
			return DailyUsageSnapshot{}, loadErr
		}
		return dailyUsageSnapshot(row, limit), nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return DailyUsageSnapshot{}, fmt.Errorf("begin messaging assistant usage record: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var inserted int
	err = tx.GetContext(ctx, &inserted, `
		INSERT INTO messaging_assistant_usage_events (
			inbound_event_id,
			workspace_id,
			provider,
			external_workspace_id,
			external_event_id,
			attempt_count,
			usage_date,
			input_tokens,
			output_tokens,
			total_tokens
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			CAST(NOW() AT TIME ZONE 'UTC' AS date),
			$7,
			$8,
			$9
		)
		ON CONFLICT (
			inbound_event_id,
			attempt_count
		) DO NOTHING
		RETURNING 1
	`,
		input.InboundEventID,
		input.WorkspaceID,
		input.Provider,
		input.ExternalWorkspaceID,
		input.ExternalEventID,
		input.AttemptCount,
		int64(input.Usage.InputTokens),
		int64(input.Usage.OutputTokens),
		int64(input.Usage.TotalTokens),
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return DailyUsageSnapshot{}, fmt.Errorf("claim messaging assistant usage event: %w", err)
	}

	var row dailyUsageRow
	if errors.Is(err, sql.ErrNoRows) {
		row, err = loadDailyUsage(ctx, tx, input.WorkspaceID)
		if err != nil {
			return DailyUsageSnapshot{}, err
		}
	} else {
		err = tx.GetContext(ctx, &row, `
			INSERT INTO messaging_assistant_daily_usage (
				workspace_id, usage_date, input_tokens, output_tokens, total_tokens, request_count
			) VALUES (
				$1,
				CAST(NOW() AT TIME ZONE 'UTC' AS date),
				$2,
				$3,
				$4,
				1
			)
			ON CONFLICT (workspace_id, usage_date) DO UPDATE SET
				input_tokens = messaging_assistant_daily_usage.input_tokens + EXCLUDED.input_tokens,
				output_tokens = messaging_assistant_daily_usage.output_tokens + EXCLUDED.output_tokens,
				total_tokens = messaging_assistant_daily_usage.total_tokens + EXCLUDED.total_tokens,
				request_count = messaging_assistant_daily_usage.request_count + 1,
				updated_at = NOW()
			RETURNING input_tokens, output_tokens, total_tokens, request_count
		`,
			input.WorkspaceID,
			int64(input.Usage.InputTokens),
			int64(input.Usage.OutputTokens),
			int64(input.Usage.TotalTokens),
		)
		if err != nil {
			return DailyUsageSnapshot{}, fmt.Errorf("record messaging assistant daily usage: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return DailyUsageSnapshot{}, fmt.Errorf("commit messaging assistant usage record: %w", err)
	}
	return dailyUsageSnapshot(row, limit), nil
}

func (r *DailyUsageRepository) load(ctx context.Context, workspaceID uuid.UUID) (dailyUsageRow, error) {
	return loadDailyUsage(ctx, r.db, workspaceID)
}

func loadDailyUsage(ctx context.Context, querier dailyUsageQuerier, workspaceID uuid.UUID) (dailyUsageRow, error) {
	var row dailyUsageRow
	err := querier.GetContext(ctx, &row, `
		SELECT
			CAST(COALESCE(SUM(input_tokens), 0) AS bigint) AS input_tokens,
			CAST(COALESCE(SUM(output_tokens), 0) AS bigint) AS output_tokens,
			CAST(COALESCE(SUM(total_tokens), 0) AS bigint) AS total_tokens,
			CAST(COALESCE(SUM(request_count), 0) AS bigint) AS request_count
		FROM messaging_assistant_daily_usage
		WHERE workspace_id = $1
		  AND usage_date = CAST(NOW() AT TIME ZONE 'UTC' AS date)
	`, workspaceID)
	if err != nil {
		return dailyUsageRow{}, fmt.Errorf("check messaging assistant daily usage: %w", err)
	}
	return row, nil
}

type dailyUsageRow struct {
	InputTokens  int64 `db:"input_tokens"`
	OutputTokens int64 `db:"output_tokens"`
	TotalTokens  int64 `db:"total_tokens"`
	RequestCount int64 `db:"request_count"`
}

func validateDailyUsageRepository(repository *DailyUsageRepository, workspaceID uuid.UUID) error {
	if repository == nil || repository.db == nil {
		return errors.New("messaging assistant daily usage repository is not configured")
	}
	if workspaceID == uuid.Nil {
		return errors.New("messaging assistant daily usage workspace is required")
	}
	return nil
}

func normalizedDailyTokenLimit(limit int64) (int64, error) {
	if limit == 0 {
		return DefaultDailyWorkspaceTokenLimit, nil
	}
	if limit < 0 {
		return 0, errors.New("messaging assistant daily token limit must be positive")
	}
	return limit, nil
}

func validateUsage(usage messaging.Usage) error {
	if usage.InputTokens < 0 || usage.OutputTokens < 0 || usage.TotalTokens < 0 {
		return errors.New("messaging assistant usage cannot be negative")
	}
	if usage.TotalTokens != usage.InputTokens+usage.OutputTokens {
		return errors.New("messaging assistant total usage must equal input plus output tokens")
	}
	return nil
}

func dailyUsageSnapshot(row dailyUsageRow, limit int64) DailyUsageSnapshot {
	remaining := limit - row.TotalTokens
	if remaining < 0 {
		remaining = 0
	}
	return DailyUsageSnapshot{
		InputTokens:  row.InputTokens,
		OutputTokens: row.OutputTokens,
		TotalTokens:  row.TotalTokens,
		RequestCount: row.RequestCount,
		Limit:        limit,
		Remaining:    remaining,
		Allowed:      row.TotalTokens < limit,
	}
}
