package teamsrepository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	teamsdomain "github.com/complexus-tech/projects-api/internal/modules/teams/domain"
	teamsql "github.com/complexus-tech/projects-api/internal/modules/teams/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *repo) List(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID uuid.UUID,
	filter teamsdomain.ListFilter,
) ([]teamsdomain.Team, error) {
	pageLimit, pageOffset, err := paginationParams(filter)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListTeamsForActor(ctx, teamsql.ListTeamsForActorParams{
		ActorID:     userID,
		WorkspaceID: workspaceID,
		JoinedOnly:  filter.JoinedOnly,
		Search:      strings.TrimSpace(filter.Search),
		PageLimit:   pageLimit,
		PageOffset:  pageOffset,
	})
	if err != nil {
		return nil, fmt.Errorf("list scoped teams: %w", err)
	}
	return toCoreListTeams(rows), nil
}

func (r *repo) GetByID(
	ctx context.Context,
	teamID uuid.UUID,
	workspaceID uuid.UUID,
	userID uuid.UUID,
) (teamsdomain.Team, error) {
	row, err := r.queries.GetTeamForActor(ctx, teamsql.GetTeamForActorParams{
		TeamID:      teamID,
		WorkspaceID: workspaceID,
		ActorID:     userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return teamsdomain.Team{}, teamsdomain.ErrNotFound
		}
		return teamsdomain.Team{}, fmt.Errorf("get scoped team: %w", err)
	}
	return toCoreGetTeam(row), nil
}

func (r *repo) ListPublicTeams(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID uuid.UUID,
	filter teamsdomain.ListFilter,
) ([]teamsdomain.Team, error) {
	pageLimit, pageOffset, err := paginationParams(filter)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListPublicTeamsForActor(ctx, teamsql.ListPublicTeamsForActorParams{
		WorkspaceID: workspaceID,
		ActorID:     userID,
		Search:      strings.TrimSpace(filter.Search),
		PageLimit:   pageLimit,
		PageOffset:  pageOffset,
	})
	if err != nil {
		return nil, fmt.Errorf("list scoped public teams: %w", err)
	}
	return toCorePublicTeams(rows), nil
}

func paginationParams(filter teamsdomain.ListFilter) (int32, int32, error) {
	if filter.Limit <= 0 {
		return 0, 0, nil
	}
	if filter.Limit > math.MaxInt32 {
		return 0, 0, fmt.Errorf("team page limit exceeds database range: %d", filter.Limit)
	}
	if filter.Offset < 0 {
		return 0, 0, fmt.Errorf("team page offset cannot be negative: %d", filter.Offset)
	}
	if filter.Offset > math.MaxInt32 {
		return 0, 0, fmt.Errorf("team page offset exceeds database range: %d", filter.Offset)
	}
	return int32(filter.Limit), int32(filter.Offset), nil
}
