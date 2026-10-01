package workpresetsrepository

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	domain "github.com/complexus-tech/projects-api/internal/modules/workpresets/domain"
	workpresetssql "github.com/complexus-tech/projects-api/internal/modules/workpresets/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ queries workpresetssql.Querier }

func New(pool *pgxpool.Pool) *Repository { return &Repository{queries: workpresetssql.New(pool)} }

func (r *Repository) List(ctx context.Context, actorID, workspaceID uuid.UUID, filters domain.List) ([]domain.Preset, error) {
	limit, err := safecast.Int32(filters.Limit)
	if err != nil || limit < 1 || limit > 101 {
		return nil, domain.ErrInvalidInput
	}
	params := workpresetssql.ListPresetsForActorParams{ActorID: actorID, WorkspaceID: workspaceID, TeamID: filters.TeamID, Kind: string(filters.Kind), ResultLimit: limit}
	if filters.Before != nil {
		params.BeforeID = &filters.Before.ID
		params.BeforeCreatedAt = &filters.Before.CreatedAt
	}
	rows, err := r.queries.ListPresetsForActor(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list work presets: %w", err)
	}
	items := make([]domain.Preset, len(rows))
	for index, row := range rows {
		items[index] = mapRow(row)
	}
	return items, nil
}

func (r *Repository) Create(ctx context.Context, actorID, workspaceID uuid.UUID, input domain.Input) (domain.Preset, error) {
	row, err := r.queries.CreatePresetForActor(ctx, workpresetssql.CreatePresetForActorParams{ActorID: actorID, WorkspaceID: workspaceID, TeamID: input.TeamID, Kind: string(input.Kind), Visibility: string(input.Visibility), Name: input.Name, Configuration: input.Configuration})
	if err != nil {
		return domain.Preset{}, mapError("create work preset", err)
	}
	return mapRow(workpresetssql.ListPresetsForActorRow(row)), nil
}

func (r *Repository) Update(ctx context.Context, actorID, workspaceID, id uuid.UUID, input domain.Update) (domain.Preset, error) {
	row, err := r.queries.RenamePresetForActor(ctx, workpresetssql.RenamePresetForActorParams{ActorID: actorID, WorkspaceID: workspaceID, ID: id, Name: input.Name})
	if err != nil {
		return domain.Preset{}, mapError("rename work preset", err)
	}
	return mapRow(workpresetssql.ListPresetsForActorRow(row)), nil
}

func (r *Repository) Archive(ctx context.Context, actorID, workspaceID, id uuid.UUID) error {
	rows, err := r.queries.ArchivePresetForActor(ctx, workpresetssql.ArchivePresetForActorParams{ActorID: actorID, WorkspaceID: workspaceID, ID: id})
	if err != nil {
		return mapError("archive work preset", err)
	}
	if rows != 1 {
		return domain.ErrNotFound
	}
	return nil
}

func mapRow(row workpresetssql.ListPresetsForActorRow) domain.Preset {
	return domain.Preset{ID: row.ID, TeamID: row.TeamID, OwnerID: row.OwnerID, Kind: domain.Kind(row.Kind), Visibility: domain.Visibility(row.Visibility), Name: row.Name, Configuration: bytes.Clone(row.Configuration), CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt, CanEdit: row.CanEdit}
}

func mapError(operation string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	return fmt.Errorf("%s: %w", operation, err)
}
