package labelsrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	labelsdomain "github.com/complexus-tech/projects-api/internal/modules/labels/domain"
	labelssql "github.com/complexus-tech/projects-api/internal/modules/labels/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/pagination"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errRepositoryNotConfigured = errors.New("labels repository is not configured")

type Repository struct {
	queries labelssql.Querier
}

func New(pool *pgxpool.Pool) *Repository {
	if pool == nil {
		return &Repository{}
	}
	return &Repository{queries: labelssql.New(pool)}
}

func (repository *Repository) configured() error {
	if repository == nil || repository.queries == nil {
		return errRepositoryNotConfigured
	}
	return nil
}

func (repository *Repository) GetLabels(ctx context.Context, actorID, workspaceID uuid.UUID, filters labelsdomain.Filters) ([]labelsdomain.Label, error) {
	if err := repository.configured(); err != nil {
		return nil, err
	}

	search := strings.TrimSpace(filters.Search)
	params := labelssql.ListLabelsForMemberParams{
		ActorID:      actorID,
		WorkspaceID:  workspaceID,
		FilterTeam:   filters.TeamID != nil,
		TeamID:       filters.TeamID,
		FilterSearch: search != "",
		Search:       search,
	}

	if filters.Limit == nil {
		if filters.Offset != 0 {
			return nil, labelsdomain.ErrInvalidPagination
		}
		rows, err := repository.queries.ListLabelsForMember(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("list labels: %w", err)
		}
		return mapLabels(rows), nil
	}
	resultLimit, resultOffset, err := labelPageBounds(*filters.Limit, filters.Offset)
	if err != nil {
		return nil, err
	}

	rows, err := repository.queries.ListLabelsPageForMember(ctx, labelssql.ListLabelsPageForMemberParams{
		ActorID:      params.ActorID,
		WorkspaceID:  params.WorkspaceID,
		FilterTeam:   params.FilterTeam,
		TeamID:       params.TeamID,
		FilterSearch: params.FilterSearch,
		Search:       params.Search,
		ResultLimit:  resultLimit,
		ResultOffset: resultOffset,
	})
	if err != nil {
		return nil, fmt.Errorf("list label page: %w", err)
	}
	return mapLabelPage(rows), nil
}

func labelPageBounds(limit, offset int) (int32, int32, error) {
	if limit < 1 || limit > pagination.MaximumPageSize+1 || offset < 0 {
		return 0, 0, labelsdomain.ErrInvalidPagination
	}
	resultLimit, err := safecast.Int32(limit)
	if err != nil {
		return 0, 0, labelsdomain.ErrInvalidPagination
	}
	resultOffset, err := safecast.Int32(offset)
	if err != nil {
		return 0, 0, labelsdomain.ErrInvalidPagination
	}
	return resultLimit, resultOffset, nil
}

func (repository *Repository) CreateLabel(ctx context.Context, actorID uuid.UUID, input labelsdomain.NewLabel) (labelsdomain.Label, error) {
	if err := repository.configured(); err != nil {
		return labelsdomain.Label{}, err
	}
	row, err := repository.queries.CreateLabelForMember(ctx, labelssql.CreateLabelForMemberParams{
		Name:        input.Name,
		TeamID:      input.TeamID,
		WorkspaceID: input.WorkspaceID,
		ActorID:     actorID,
		Color:       input.Color,
	})
	if err != nil {
		return labelsdomain.Label{}, mapError("create label", err)
	}
	return mapLabel(row), nil
}

func (repository *Repository) GetLabel(ctx context.Context, actorID, labelID, workspaceID uuid.UUID) (labelsdomain.Label, error) {
	if err := repository.configured(); err != nil {
		return labelsdomain.Label{}, err
	}
	row, err := repository.queries.GetLabelForMember(ctx, labelssql.GetLabelForMemberParams{
		ActorID:     actorID,
		LabelID:     labelID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return labelsdomain.Label{}, mapError("get label", err)
	}
	return mapLabel(row), nil
}

func (repository *Repository) UpdateLabel(ctx context.Context, actorID, labelID, workspaceID uuid.UUID, name, color string) (labelsdomain.Label, error) {
	if err := repository.configured(); err != nil {
		return labelsdomain.Label{}, err
	}
	row, err := repository.queries.UpdateLabelForMember(ctx, labelssql.UpdateLabelForMemberParams{
		Name:        name,
		Color:       color,
		LabelID:     labelID,
		WorkspaceID: workspaceID,
		ActorID:     actorID,
	})
	if err != nil {
		return labelsdomain.Label{}, mapError("update label", err)
	}
	return mapLabel(row), nil
}

func (repository *Repository) DeleteLabel(ctx context.Context, actorID, labelID, workspaceID uuid.UUID) error {
	if err := repository.configured(); err != nil {
		return err
	}
	rows, err := repository.queries.DeleteLabelForMember(ctx, labelssql.DeleteLabelForMemberParams{
		LabelID:     labelID,
		WorkspaceID: workspaceID,
		ActorID:     actorID,
	})
	if err != nil {
		return fmt.Errorf("delete label: %w", err)
	}
	if rows != 1 {
		return labelsdomain.ErrNotFound
	}
	return nil
}

func mapError(operation string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return labelsdomain.ErrNotFound
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func mapLabel(row labelssql.Label) labelsdomain.Label {
	workspaceID := uuid.Nil
	if row.WorkspaceID != nil {
		workspaceID = *row.WorkspaceID
	}
	color := ""
	if row.Color != nil {
		color = *row.Color
	}
	return labelsdomain.Label{
		ID:          row.LabelID,
		Name:        row.Name,
		TeamID:      row.TeamID,
		WorkspaceID: workspaceID,
		Color:       color,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func mapLabels(rows []labelssql.Label) []labelsdomain.Label {
	labels := make([]labelsdomain.Label, len(rows))
	for index, row := range rows {
		labels[index] = mapLabel(row)
	}
	return labels
}

func mapLabelPage(rows []labelssql.Label) []labelsdomain.Label {
	return mapLabels(rows)
}
