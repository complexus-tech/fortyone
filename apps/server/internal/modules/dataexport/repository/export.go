package dataexportrepository

import (
	"context"
	"fmt"

	exportdomain "github.com/complexus-tech/projects-api/internal/modules/dataexport/domain"
	exportsql "github.com/complexus-tech/projects-api/internal/modules/dataexport/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (repository *Repository) Snapshot(ctx context.Context, scope exportdomain.Scope) (exportdomain.Snapshot, error) {
	var snapshot exportdomain.Snapshot
	if err := scope.Validate(); err != nil {
		return snapshot, err
	}
	err := platformdatabase.NewTransactor(repository.pool).WithinTransaction(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly,
	}, func(transaction pgx.Tx) error {
		queries := exportsql.New(transaction)
		budget, err := queries.GetExportBudget(ctx, exportsql.GetExportBudgetParams{ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, TeamID: scope.TeamID})
		if err != nil {
			return fmt.Errorf("inspect export budget: %w", err)
		}
		if !budget.Authorized || !budget.TeamVisible {
			return exportdomain.ErrForbidden
		}
		if budget.TaskCount > exportdomain.MaximumTasks || budget.EstimatedBytes > exportdomain.MaximumBytes || budget.TeamCount > 100 || budget.PersonCount > 500 || budget.LabelCount > 500 || budget.ObjectiveCount > 250 || budget.SprintCount > 250 || budget.KeyResultCount > 500 || budget.FieldCount > 2500 {
			return exportdomain.ErrTooLarge
		}
		analysis, err := queries.ExportWorkGraph(ctx, exportsql.ExportWorkGraphParams{ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, TeamID: scope.TeamID})
		if err != nil {
			return fmt.Errorf("export work graph: %w", err)
		}
		fields, err := queries.ExportFieldDefinitions(ctx, exportsql.ExportFieldDefinitionsParams{ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, TeamID: scope.TeamID})
		if err != nil {
			return fmt.Errorf("export field definitions: %w", err)
		}
		taskData, err := queries.ExportTaskData(ctx, exportsql.ExportTaskDataParams{ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, TeamID: scope.TeamID})
		if err != nil {
			return fmt.Errorf("export canonical task data: %w", err)
		}
		snapshot = exportdomain.Snapshot{TaskCount: int(budget.TaskCount), Envelope: exportdomain.Envelope{
			Format: "fortyone-work-export", Version: 1, GeneratedAt: budget.GeneratedAt,
			Analysis: analysis, CustomFields: fields, TaskData: taskData,
		}}
		return nil
	})
	return snapshot, err
}
