package teamsrepository

import (
	"context"
	"errors"
	"testing"
	"time"

	teamsql "github.com/complexus-tech/projects-api/internal/modules/teams/repository/sqlc"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type fakeTeamQueries struct {
	teamsql.Querier
	list             func(context.Context, teamsql.ListTeamsForActorParams) ([]teamsql.ListTeamsForActorRow, error)
	get              func(context.Context, teamsql.GetTeamForActorParams) (teamsql.GetTeamForActorRow, error)
	create           func(context.Context, teamsql.CreateTeamParams) (teamsql.CreateTeamRow, error)
	createAutomation func(context.Context, teamsql.CreateDefaultStoryAutomationSettingsParams) (int64, error)
	createStatus     func(context.Context, teamsql.CreateDefaultStoryStatusParams) (int64, error)
	addMember        func(context.Context, teamsql.AddTeamMemberForWorkspaceParams) (teamsql.AddTeamMemberForWorkspaceRow, error)
	joinPublic       func(context.Context, teamsql.JoinPublicTeamForActorParams) (teamsql.JoinPublicTeamForActorRow, error)
	canOrder         func(context.Context, teamsql.ActorCanOrderWorkspaceTeamsParams) (bool, error)
	deleteOrdering   func(context.Context, teamsql.DeleteActorTeamOrderingParams) error
	insertOrdering   func(context.Context, teamsql.InsertActorTeamOrderParams) (int64, error)
}

func (fake fakeTeamQueries) ListTeamsForActor(ctx context.Context, params teamsql.ListTeamsForActorParams) ([]teamsql.ListTeamsForActorRow, error) {
	return fake.list(ctx, params)
}

func (fake fakeTeamQueries) GetTeamForActor(ctx context.Context, params teamsql.GetTeamForActorParams) (teamsql.GetTeamForActorRow, error) {
	return fake.get(ctx, params)
}

func (fake fakeTeamQueries) CreateTeam(ctx context.Context, params teamsql.CreateTeamParams) (teamsql.CreateTeamRow, error) {
	return fake.create(ctx, params)
}

func (fake fakeTeamQueries) CreateDefaultStoryAutomationSettings(ctx context.Context, params teamsql.CreateDefaultStoryAutomationSettingsParams) (int64, error) {
	return fake.createAutomation(ctx, params)
}

func (fake fakeTeamQueries) CreateDefaultStoryStatus(ctx context.Context, params teamsql.CreateDefaultStoryStatusParams) (int64, error) {
	return fake.createStatus(ctx, params)
}

func (fake fakeTeamQueries) AddTeamMemberForWorkspace(ctx context.Context, params teamsql.AddTeamMemberForWorkspaceParams) (teamsql.AddTeamMemberForWorkspaceRow, error) {
	return fake.addMember(ctx, params)
}

func (fake fakeTeamQueries) JoinPublicTeamForActor(ctx context.Context, params teamsql.JoinPublicTeamForActorParams) (teamsql.JoinPublicTeamForActorRow, error) {
	return fake.joinPublic(ctx, params)
}

func (fake fakeTeamQueries) ActorCanOrderWorkspaceTeams(ctx context.Context, params teamsql.ActorCanOrderWorkspaceTeamsParams) (bool, error) {
	return fake.canOrder(ctx, params)
}

func (fake fakeTeamQueries) DeleteActorTeamOrdering(ctx context.Context, params teamsql.DeleteActorTeamOrderingParams) error {
	return fake.deleteOrdering(ctx, params)
}

func (fake fakeTeamQueries) InsertActorTeamOrder(ctx context.Context, params teamsql.InsertActorTeamOrderParams) (int64, error) {
	return fake.insertOrdering(ctx, params)
}

func TestListMapsTypedScopePaginationAndRows(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	actorID := uuid.New()
	teamID := uuid.New()
	createdAt := time.Date(2026, time.August, 28, 9, 0, 0, 0, time.UTC)
	repository := newTeamTestRepository(fakeTeamQueries{
		list: func(_ context.Context, params teamsql.ListTeamsForActorParams) ([]teamsql.ListTeamsForActorRow, error) {
			if params.WorkspaceID != workspaceID || params.ActorID != actorID || !params.JoinedOnly {
				t.Fatalf("scope params = %#v", params)
			}
			if params.Search != "Platform" || params.PageLimit != 21 || params.PageOffset != 40 {
				t.Fatalf("filter params = %#v", params)
			}
			return []teamsql.ListTeamsForActorRow{{
				TeamID:         teamID,
				Name:           "Platform",
				Code:           "PLT",
				Color:          "#000000",
				WorkspaceID:    workspaceID,
				CreatedAt:      createdAt,
				UpdatedAt:      createdAt,
				MemberCount:    7,
				SprintsEnabled: true,
			}}, nil
		},
	})

	got, err := repository.List(context.Background(), workspaceID, actorID, teams.CoreListTeamsFilter{
		Search: " Platform ", Limit: 21, Offset: 40, JoinedOnly: true,
	})
	if err != nil {
		t.Fatalf("list teams: %v", err)
	}
	if len(got) != 1 || got[0].ID != teamID || got[0].MemberCount != 7 || !got[0].SprintsEnabled {
		t.Fatalf("mapped teams = %#v", got)
	}
}

func TestGetMapsHiddenTeamToDomainNotFound(t *testing.T) {
	t.Parallel()

	repository := newTeamTestRepository(fakeTeamQueries{
		get: func(context.Context, teamsql.GetTeamForActorParams) (teamsql.GetTeamForActorRow, error) {
			return teamsql.GetTeamForActorRow{}, pgx.ErrNoRows
		},
	})
	_, err := repository.GetByID(context.Background(), uuid.New(), uuid.New(), uuid.New())
	if !errors.Is(err, teams.ErrTeamNotFound) {
		t.Fatalf("get error = %v, want ErrTeamNotFound", err)
	}
}

func TestCreateMapsUniqueCodeAndInitializesEveryDefault(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	teamID := uuid.New()
	statusCount := 0
	queries := fakeTeamQueries{
		create: func(_ context.Context, params teamsql.CreateTeamParams) (teamsql.CreateTeamRow, error) {
			if params.WorkspaceID != workspaceID || params.Code != "PLT" {
				t.Fatalf("create params = %#v", params)
			}
			return teamsql.CreateTeamRow{
				TeamID: teamID, Name: params.Name, Code: params.Code, Color: params.Color,
				WorkspaceID: workspaceID, IsPrivate: params.IsPrivate,
			}, nil
		},
		createAutomation: func(_ context.Context, params teamsql.CreateDefaultStoryAutomationSettingsParams) (int64, error) {
			if params.TeamID != teamID || params.WorkspaceID != workspaceID {
				t.Fatalf("automation params = %#v", params)
			}
			return 1, nil
		},
		createStatus: func(_ context.Context, params teamsql.CreateDefaultStoryStatusParams) (int64, error) {
			if params.TeamID != teamID || params.WorkspaceID != workspaceID {
				t.Fatalf("status params = %#v", params)
			}
			statusCount++
			return 1, nil
		},
	}
	repository := newTeamTestRepository(queries)
	repository.runTransaction = func(ctx context.Context, operation func(teamsql.Querier) error) error {
		return operation(queries)
	}

	created, err := repository.Create(context.Background(), teams.CoreTeam{
		Name: "Platform", Code: "PLT", Color: "#000000", Workspace: workspaceID, IsPrivate: true,
	})
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	if created.ID != teamID || created.MemberCount != 1 || !created.IsPrivate {
		t.Fatalf("created team = %#v", created)
	}
	if statusCount != len(teams.DefaultStoryStatuses) {
		t.Fatalf("created statuses = %d, want %d", statusCount, len(teams.DefaultStoryStatuses))
	}

	uniqueQueries := fakeTeamQueries{
		create: func(context.Context, teamsql.CreateTeamParams) (teamsql.CreateTeamRow, error) {
			return teamsql.CreateTeamRow{}, &pgconn.PgError{Code: "23505"}
		},
	}
	uniqueRepository := newTeamTestRepository(uniqueQueries)
	uniqueRepository.runTransaction = func(ctx context.Context, operation func(teamsql.Querier) error) error {
		return operation(uniqueQueries)
	}
	_, err = uniqueRepository.Create(context.Background(), teams.CoreTeam{Workspace: workspaceID})
	if !errors.Is(err, teams.ErrTeamCodeExists) {
		t.Fatalf("unique create error = %v, want ErrTeamCodeExists", err)
	}
}

func TestMembershipCommandsMapTypedScopeAndOutcomes(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	repository := newTeamTestRepository(fakeTeamQueries{
		addMember: func(_ context.Context, params teamsql.AddTeamMemberForWorkspaceParams) (teamsql.AddTeamMemberForWorkspaceRow, error) {
			if params.WorkspaceID != workspaceID || params.TeamID != teamID || params.UserID != actorID {
				t.Fatalf("add params = %#v", params)
			}
			return teamsql.AddTeamMemberForWorkspaceRow{Eligible: true}, nil
		},
		joinPublic: func(_ context.Context, params teamsql.JoinPublicTeamForActorParams) (teamsql.JoinPublicTeamForActorRow, error) {
			if params.WorkspaceID != workspaceID || params.TeamID != teamID || params.ActorID != actorID {
				t.Fatalf("join params = %#v", params)
			}
			return teamsql.JoinPublicTeamForActorRow{}, nil
		},
	})

	if err := repository.AddMember(context.Background(), teamID, actorID, workspaceID); !errors.Is(err, teams.ErrTeamMemberExists) {
		t.Fatalf("duplicate add error = %v, want ErrTeamMemberExists", err)
	}
	if err := repository.JoinPublicTeam(context.Background(), teams.CorePublicTeamJoin{
		TeamID: teamID, ActorID: actorID, WorkspaceID: workspaceID,
	}); !errors.Is(err, teams.ErrTeamNotFound) {
		t.Fatalf("ineligible join error = %v, want ErrTeamNotFound", err)
	}
}

func TestOrderingRejectsCrossScopeTeamInsideTransaction(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	actorID := uuid.New()
	teamIDs := []uuid.UUID{uuid.New(), uuid.New()}
	inserted := 0
	queries := fakeTeamQueries{
		canOrder: func(_ context.Context, params teamsql.ActorCanOrderWorkspaceTeamsParams) (bool, error) {
			return params.WorkspaceID == workspaceID && params.ActorID == actorID, nil
		},
		deleteOrdering: func(_ context.Context, params teamsql.DeleteActorTeamOrderingParams) error {
			if params.WorkspaceID != workspaceID || params.ActorID != actorID {
				t.Fatalf("delete params = %#v", params)
			}
			return nil
		},
		insertOrdering: func(_ context.Context, params teamsql.InsertActorTeamOrderParams) (int64, error) {
			if params.OrderIndex != int32(inserted) || params.TeamID != teamIDs[inserted] {
				t.Fatalf("insert params = %#v", params)
			}
			inserted++
			if inserted == 2 {
				return 0, nil
			}
			return 1, nil
		},
	}
	repository := newTeamTestRepository(queries)
	repository.runTransaction = func(ctx context.Context, operation func(teamsql.Querier) error) error {
		return operation(queries)
	}

	err := repository.UpdateUserTeamOrdering(context.Background(), actorID, workspaceID, teamIDs)
	if !errors.Is(err, teams.ErrTeamNotFound) {
		t.Fatalf("ordering error = %v, want ErrTeamNotFound", err)
	}
	if inserted != 2 {
		t.Fatalf("insert attempts = %d, want 2", inserted)
	}
}

func TestPaginationParamsRejectsUnsafeValues(t *testing.T) {
	t.Parallel()

	if _, _, err := paginationParams(teams.CoreListTeamsFilter{Limit: 1, Offset: -1}); err == nil {
		t.Fatal("negative offset error = nil")
	}
	limit, offset, err := paginationParams(teams.CoreListTeamsFilter{Limit: -1, Offset: 99})
	if err != nil || limit != 0 || offset != 0 {
		t.Fatalf("unbounded pagination = (%d, %d, %v), want zeros", limit, offset, err)
	}
}

func newTeamTestRepository(queries teamsql.Querier) *repo {
	return newWithQueries(queries)
}
