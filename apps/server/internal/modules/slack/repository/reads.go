package slackrepository

import (
	"context"
	"errors"
	"strings"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	slacksql "github.com/complexus-tech/projects-api/internal/modules/slack/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repo) FindWorkspaceBySlug(ctx context.Context, slug string) (WorkspaceRecord, error) {
	row, err := repository.queries.FindWorkspaceBySlug(ctx, slacksql.FindWorkspaceBySlugParams{Slug: strings.TrimSpace(slug)})
	if err != nil {
		return WorkspaceRecord{}, mapDatabaseError(err)
	}
	return slackdomain.Workspace{ID: row.WorkspaceID, Slug: row.Slug, Name: row.Name}, nil
}

func (repository *Repo) FindWorkspaceByID(ctx context.Context, workspaceID uuid.UUID) (WorkspaceRecord, error) {
	row, err := repository.queries.FindWorkspaceByID(ctx, slacksql.FindWorkspaceByIDParams{WorkspaceID: workspaceID})
	if err != nil {
		return WorkspaceRecord{}, mapDatabaseError(err)
	}
	return slackdomain.Workspace{ID: row.WorkspaceID, Slug: row.Slug, Name: row.Name}, nil
}

func (repository *Repo) FindTeamByCode(ctx context.Context, workspaceID uuid.UUID, code string) (TeamRecord, error) {
	row, err := repository.queries.FindTeamByCode(ctx, slacksql.FindTeamByCodeParams{WorkspaceID: workspaceID, Code: strings.TrimSpace(code)})
	if err != nil {
		return TeamRecord{}, mapDatabaseError(err)
	}
	return slackdomain.Team{ID: row.TeamID, Code: row.Code, Name: row.Name, Color: row.Color}, nil
}

func (repository *Repo) FindTeamByID(ctx context.Context, workspaceID, teamID uuid.UUID) (TeamRecord, error) {
	row, err := repository.queries.FindTeamByID(ctx, slacksql.FindTeamByIDParams{WorkspaceID: workspaceID, TeamID: teamID})
	if err != nil {
		return TeamRecord{}, mapDatabaseError(err)
	}
	return slackdomain.Team{ID: row.TeamID, Code: row.Code, Name: row.Name, Color: row.Color}, nil
}

func (repository *Repo) GetWorkspaceBySlackTeamID(ctx context.Context, slackTeamID string) (WorkspaceRecord, error) {
	row, err := repository.queries.GetWorkspaceBySlackTeamID(ctx, slacksql.GetWorkspaceBySlackTeamIDParams{SlackTeamID: strings.TrimSpace(slackTeamID)})
	if err != nil {
		return WorkspaceRecord{}, mapDatabaseError(err)
	}
	return slackdomain.Workspace{ID: row.WorkspaceID, Slug: row.Slug, Name: row.Name}, nil
}

func (repository *Repo) ListWorkspaceTeams(ctx context.Context, workspaceID uuid.UUID) ([]TeamRecord, error) {
	rows, err := repository.queries.ListWorkspaceTeams(ctx, slacksql.ListWorkspaceTeamsParams{WorkspaceID: workspaceID})
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	result := make([]TeamRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, slackdomain.Team{ID: row.TeamID, Code: row.Code, Name: row.Name, Color: row.Color})
	}
	return result, nil
}

func (repository *Repo) ListWorkspaceTeamsForUser(ctx context.Context, workspaceID, userID uuid.UUID) ([]TeamRecord, error) {
	rows, err := repository.queries.ListWorkspaceTeamsForUser(ctx, slacksql.ListWorkspaceTeamsForUserParams{WorkspaceID: workspaceID, UserID: userID})
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	result := make([]TeamRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, slackdomain.Team{ID: row.TeamID, Code: row.Code, Name: row.Name, Color: row.Color})
	}
	return result, nil
}

func (repository *Repo) ListTeamStatuses(ctx context.Context, teamID uuid.UUID) ([]StatusRecord, error) {
	rows, err := repository.queries.ListTeamStatuses(ctx, slacksql.ListTeamStatusesParams{TeamID: teamID})
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	result := make([]StatusRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, slackdomain.Status{ID: row.StatusID, Name: row.Name, Category: row.Category})
	}
	return result, nil
}

func (repository *Repo) ListTeamMembers(ctx context.Context, teamID uuid.UUID) ([]TeamMemberRecord, error) {
	rows, err := repository.queries.ListTeamMembers(ctx, slacksql.ListTeamMembersParams{TeamID: teamID})
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	return mapTeamMembers(rows), nil
}

func (repository *Repo) ListTeamLabels(ctx context.Context, workspaceID, teamID uuid.UUID) ([]LabelRecord, error) {
	rows, err := repository.queries.ListTeamLabels(ctx, slacksql.ListTeamLabelsParams{WorkspaceID: workspaceID, TeamID: teamID})
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	result := make([]LabelRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, slackdomain.Label{ID: row.LabelID, Name: row.Name})
	}
	return result, nil
}

func (repository *Repo) FindTeamMemberByID(ctx context.Context, teamID, userID uuid.UUID) (TeamMemberRecord, error) {
	row, err := repository.queries.FindTeamMemberByID(ctx, slacksql.FindTeamMemberByIDParams{TeamID: teamID, UserID: userID})
	if err != nil {
		return TeamMemberRecord{}, mapDatabaseError(err)
	}
	return slackdomain.TeamMember{UserID: row.UserID, Username: row.Username, FullName: row.FullName, Email: row.Email}, nil
}

func (repository *Repo) FindTeamLabelByID(ctx context.Context, workspaceID, teamID, labelID uuid.UUID) (LabelRecord, error) {
	row, err := repository.queries.FindTeamLabelByID(ctx, slacksql.FindTeamLabelByIDParams{WorkspaceID: workspaceID, TeamID: teamID, LabelID: labelID})
	if err != nil {
		return LabelRecord{}, mapDatabaseError(err)
	}
	return slackdomain.Label{ID: row.LabelID, Name: row.Name}, nil
}

func (repository *Repo) FindTeamObjectiveByID(ctx context.Context, workspaceID, teamID, objectiveID uuid.UUID) (ObjectiveRecord, error) {
	row, err := repository.queries.FindTeamObjectiveByID(ctx, slacksql.FindTeamObjectiveByIDParams{WorkspaceID: workspaceID, TeamID: teamID, ObjectiveID: objectiveID})
	if err != nil {
		return ObjectiveRecord{}, mapDatabaseError(err)
	}
	return slackdomain.Objective{ID: row.ObjectiveID, Name: row.Name}, nil
}

func (repository *Repo) SearchTeamMembers(ctx context.Context, teamID uuid.UUID, query string, limit int) ([]TeamMemberRecord, error) {
	queryLimit, err := normalizedSearchLimit(limit)
	if err != nil {
		return nil, err
	}
	rows, err := repository.queries.SearchTeamMembers(ctx, slacksql.SearchTeamMembersParams{TeamID: teamID, SearchQuery: strings.TrimSpace(query), ResultLimit: queryLimit})
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	result := make([]TeamMemberRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, slackdomain.TeamMember{UserID: row.UserID, Username: row.Username, FullName: row.FullName, Email: row.Email})
	}
	return result, nil
}

func (repository *Repo) SearchTeamLabels(ctx context.Context, workspaceID, teamID uuid.UUID, query string, limit int) ([]LabelRecord, error) {
	queryLimit, err := normalizedSearchLimit(limit)
	if err != nil {
		return nil, err
	}
	rows, err := repository.queries.SearchTeamLabels(ctx, slacksql.SearchTeamLabelsParams{WorkspaceID: workspaceID, TeamID: teamID, SearchQuery: strings.TrimSpace(query), ResultLimit: queryLimit})
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	result := make([]LabelRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, slackdomain.Label{ID: row.LabelID, Name: row.Name})
	}
	return result, nil
}

func (repository *Repo) SearchTeamObjectives(ctx context.Context, workspaceID, teamID uuid.UUID, query string, limit int) ([]ObjectiveRecord, error) {
	queryLimit, err := normalizedSearchLimit(limit)
	if err != nil {
		return nil, err
	}
	rows, err := repository.queries.SearchTeamObjectives(ctx, slacksql.SearchTeamObjectivesParams{WorkspaceID: workspaceID, TeamID: teamID, SearchQuery: strings.TrimSpace(query), ResultLimit: queryLimit})
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	result := make([]ObjectiveRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, slackdomain.Objective{ID: row.ObjectiveID, Name: row.Name})
	}
	return result, nil
}

func (repository *Repo) FindFirstStatusByCategory(ctx context.Context, teamID uuid.UUID, category string) (*uuid.UUID, error) {
	id, err := repository.queries.FindFirstStatusByCategory(ctx, slacksql.FindFirstStatusByCategoryParams{TeamID: teamID, Category: strings.TrimSpace(category)})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	return &id, nil
}

func normalizedSearchLimit(limit int) (int32, error) {
	if limit <= 0 || limit > 50 {
		limit = 25
	}
	return safecast.Int32(limit)
}

func mapTeamMembers(rows []slacksql.ListTeamMembersRow) []TeamMemberRecord {
	result := make([]TeamMemberRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, slackdomain.TeamMember{UserID: row.UserID, Username: row.Username, FullName: row.FullName, Email: row.Email})
	}
	return result
}
