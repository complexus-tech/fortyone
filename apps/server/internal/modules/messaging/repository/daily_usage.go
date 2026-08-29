package messagingrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/domain"
	messagingsql "github.com/complexus-tech/projects-api/internal/modules/messaging/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

type DailyUsageRecordInput struct {
	InboundEventID      uuid.UUID
	WorkspaceID         uuid.UUID
	Provider            string
	ExternalWorkspaceID string
	ExternalEventID     string
	AttemptCount        int
	Usage               messaging.Usage
}

type dailyUsageQueries interface {
	AddAssistantDailyUsage(context.Context, messagingsql.AddAssistantDailyUsageParams) (messagingsql.AddAssistantDailyUsageRow, error)
	ClaimAssistantUsageEvent(context.Context, messagingsql.ClaimAssistantUsageEventParams) (int32, error)
	GetAssistantDailyUsage(context.Context, messagingsql.GetAssistantDailyUsageParams) (messagingsql.GetAssistantDailyUsageRow, error)
}

type DailyUsageRepository struct {
	queries        dailyUsageQueries
	runTransaction func(context.Context, func(dailyUsageQueries) error) error
}

func NewDailyUsageRepository(pool *pgxpool.Pool) *DailyUsageRepository {
	if pool == nil {
		return &DailyUsageRepository{}
	}
	queries := messagingsql.New(pool)
	transactor := platformdatabase.NewTransactor(pool)
	return &DailyUsageRepository{
		queries: queries,
		runTransaction: func(ctx context.Context, operation func(dailyUsageQueries) error) error {
			return transactor.WithinTransaction(ctx, pgx.TxOptions{
				IsoLevel:   pgx.ReadCommitted,
				AccessMode: pgx.ReadWrite,
			}, func(tx pgx.Tx) error {
				return operation(queries.WithTx(tx))
			})
		},
	}
}

func newDailyUsageRepositoryWithQueries(queries dailyUsageQueries) *DailyUsageRepository {
	return &DailyUsageRepository{
		queries: queries,
		runTransaction: func(ctx context.Context, operation func(dailyUsageQueries) error) error {
			return operation(queries)
		},
	}
}

func (repository *DailyUsageRepository) Check(
	ctx context.Context,
	workspaceID uuid.UUID,
	limit int64,
) (DailyUsageSnapshot, error) {
	if err := validateDailyUsageRepository(repository, workspaceID); err != nil {
		return DailyUsageSnapshot{}, err
	}
	limit, err := normalizedDailyTokenLimit(limit)
	if err != nil {
		return DailyUsageSnapshot{}, err
	}
	row, err := repository.load(ctx, workspaceID)
	if err != nil {
		return DailyUsageSnapshot{}, err
	}
	snapshot := dailyUsageSnapshot(row, limit)
	if !snapshot.Allowed {
		return snapshot, &DailyTokenLimitError{WorkspaceID: workspaceID, Used: snapshot.TotalTokens, Limit: limit}
	}
	return snapshot, nil
}

func (repository *DailyUsageRepository) Record(
	ctx context.Context,
	input DailyUsageRecordInput,
	limit int64,
) (DailyUsageSnapshot, error) {
	input.Provider = strings.ToLower(strings.TrimSpace(input.Provider))
	input.ExternalWorkspaceID = strings.TrimSpace(input.ExternalWorkspaceID)
	input.ExternalEventID = strings.TrimSpace(input.ExternalEventID)
	if err := validateDailyUsageRepository(repository, input.WorkspaceID); err != nil {
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
	databaseAttemptCount, err := safecast.Int32(input.AttemptCount)
	if err != nil {
		return DailyUsageSnapshot{}, fmt.Errorf("validate messaging assistant usage attempt count: %w", err)
	}
	if input.Usage.TotalTokens == 0 {
		row, loadErr := repository.load(ctx, input.WorkspaceID)
		return dailyUsageSnapshot(row, limit), loadErr
	}

	var result dailyUsageRow
	err = repository.runTransaction(ctx, func(queries dailyUsageQueries) error {
		_, claimErr := queries.ClaimAssistantUsageEvent(ctx, messagingsql.ClaimAssistantUsageEventParams{
			InboundEventID: input.InboundEventID, WorkspaceID: input.WorkspaceID,
			Provider: input.Provider, ExternalWorkspaceID: input.ExternalWorkspaceID,
			ExternalEventID: input.ExternalEventID, AttemptCount: databaseAttemptCount,
			InputTokens: int64(input.Usage.InputTokens), OutputTokens: int64(input.Usage.OutputTokens),
			TotalTokens: int64(input.Usage.TotalTokens),
		})
		if claimErr != nil && !errors.Is(claimErr, pgx.ErrNoRows) {
			return fmt.Errorf("claim messaging assistant usage event: %w", claimErr)
		}
		if errors.Is(claimErr, pgx.ErrNoRows) {
			loaded, loadErr := loadDailyUsage(ctx, queries, input.WorkspaceID)
			result = loaded
			return loadErr
		}
		row, addErr := queries.AddAssistantDailyUsage(ctx, messagingsql.AddAssistantDailyUsageParams{
			WorkspaceID: input.WorkspaceID, InputTokens: int64(input.Usage.InputTokens),
			OutputTokens: int64(input.Usage.OutputTokens), TotalTokens: int64(input.Usage.TotalTokens),
		})
		if addErr != nil {
			return fmt.Errorf("record messaging assistant daily usage: %w", addErr)
		}
		result = dailyUsageRow(row)
		return nil
	})
	if err != nil {
		return DailyUsageSnapshot{}, fmt.Errorf("record messaging assistant usage: %w", err)
	}
	return dailyUsageSnapshot(result, limit), nil
}

func (repository *DailyUsageRepository) load(ctx context.Context, workspaceID uuid.UUID) (dailyUsageRow, error) {
	return loadDailyUsage(ctx, repository.queries, workspaceID)
}

func loadDailyUsage(ctx context.Context, queries dailyUsageQueries, workspaceID uuid.UUID) (dailyUsageRow, error) {
	row, err := queries.GetAssistantDailyUsage(ctx, messagingsql.GetAssistantDailyUsageParams{WorkspaceID: workspaceID})
	if err != nil {
		return dailyUsageRow{}, fmt.Errorf("check messaging assistant daily usage: %w", err)
	}
	return dailyUsageRow(row), nil
}

type dailyUsageRow struct {
	InputTokens  int64
	OutputTokens int64
	TotalTokens  int64
	RequestCount int64
}

func validateDailyUsageRepository(repository *DailyUsageRepository, workspaceID uuid.UUID) error {
	if repository == nil || repository.queries == nil || repository.runTransaction == nil {
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
		InputTokens: row.InputTokens, OutputTokens: row.OutputTokens,
		TotalTokens: row.TotalTokens, RequestCount: row.RequestCount,
		Limit: limit, Remaining: remaining, Allowed: row.TotalTokens < limit,
	}
}
