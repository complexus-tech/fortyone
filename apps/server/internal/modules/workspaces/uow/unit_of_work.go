package workspaceuow

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	teamsrepository "github.com/complexus-tech/projects-api/internal/modules/teams/repository"
	usersrepository "github.com/complexus-tech/projects-api/internal/modules/users/repository"
	workspacesrepository "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrNilOperation           = errors.New("workspace unit-of-work operation is required")
	ErrTransactionScopeClosed = errors.New("workspace transaction scope is closed")
)

type workspaceBinder interface {
	BindWorkspaceTransaction(pgx.Tx) (workspacesrepository.WorkspaceTransaction, error)
}

type teamBinder interface {
	BindWorkspaceTransaction(pgx.Tx) (teamsrepository.WorkspaceTransaction, error)
}

type userBinder interface {
	BindWorkspaceTransaction(pgx.Tx) (usersrepository.WorkspaceTransaction, error)
}

// Manager creates a fresh, transaction-bound capability scope for each
// callback. Pool-backed repositories are never exposed to the callback.
type Manager struct {
	transactions platformdatabase.Transactor
	workspaces   workspaceBinder
	teams        teamBinder
	users        userBinder
}

func New(
	beginner platformdatabase.Beginner,
	workspaces workspaceBinder,
	teams teamBinder,
	users userBinder,
) (*Manager, error) {
	if beginner == nil {
		return nil, errors.New("workspace unit-of-work transaction beginner is required")
	}
	if workspaces == nil {
		return nil, errors.New("workspace unit-of-work workspace binder is required")
	}
	if teams == nil {
		return nil, errors.New("workspace unit-of-work team binder is required")
	}
	if users == nil {
		return nil, errors.New("workspace unit-of-work user binder is required")
	}
	return &Manager{
		transactions: platformdatabase.NewTransactor(beginner),
		workspaces:   workspaces,
		teams:        teams,
		users:        users,
	}, nil
}

func (manager *Manager) WithinTransaction(
	ctx context.Context,
	operation func(workspaces.Transaction) error,
) error {
	if operation == nil {
		return ErrNilOperation
	}
	return manager.transactions.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		workspaceStore, err := manager.workspaces.BindWorkspaceTransaction(tx)
		if err != nil {
			return err
		}
		teamStore, err := manager.teams.BindWorkspaceTransaction(tx)
		if err != nil {
			return err
		}
		userStore, err := manager.users.BindWorkspaceTransaction(tx)
		if err != nil {
			return err
		}

		scope := &transactionScope{
			workspaces: workspaceStore,
			teams:      teamStore,
			users:      userStore,
		}
		scope.open.Store(true)
		defer scope.close()
		return operation(scope)
	})
}

type transactionScope struct {
	mu         sync.Mutex
	open       atomic.Bool
	workspaces workspacesrepository.WorkspaceTransaction
	teams      teamsrepository.WorkspaceTransaction
	users      usersrepository.WorkspaceTransaction
}

func (scope *transactionScope) locked(operation func() error) error {
	scope.mu.Lock()
	defer scope.mu.Unlock()
	if !scope.open.Load() {
		return ErrTransactionScopeClosed
	}
	return operation()
}

// close waits for any in-flight capability call before the transactor attempts
// commit. A callback that accidentally launches work it does not await cannot
// race a pgx query against commit; queued late calls observe the closed scope.
func (scope *transactionScope) close() {
	scope.mu.Lock()
	defer scope.mu.Unlock()
	scope.open.Store(false)
}

func (scope *transactionScope) CreateWorkspace(
	ctx context.Context,
	workspace workspaces.CoreWorkspace,
	createdBy uuid.UUID,
) (workspaces.CoreWorkspace, error) {
	var created workspaces.CoreWorkspace
	err := scope.locked(func() error {
		var err error
		created, err = scope.workspaces.CreateWorkspace(ctx, workspace, createdBy)
		return err
	})
	return created, err
}

func (scope *transactionScope) AddWorkspaceMember(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID uuid.UUID,
	role string,
) error {
	return scope.locked(func() error {
		return scope.workspaces.AddMember(ctx, workspaceID, userID, role)
	})
}

func (scope *transactionScope) CreateTeam(
	ctx context.Context,
	team workspaces.DefaultTeam,
) (workspaces.CreatedTeam, error) {
	var created workspaces.CreatedTeam
	err := scope.locked(func() error {
		stored, err := scope.teams.CreateTeam(ctx, teamsrepository.WorkspaceTeamInput{
			Name:      team.Name,
			Code:      team.Code,
			Color:     team.Color,
			Workspace: team.Workspace,
		})
		created.ID = stored.ID
		return err
	})
	return created, err
}

func (scope *transactionScope) AddTeamMember(
	ctx context.Context,
	teamID uuid.UUID,
	userID uuid.UUID,
	workspaceID uuid.UUID,
) error {
	return scope.locked(func() error {
		return scope.teams.AddMember(ctx, teamID, userID, workspaceID)
	})
}

func (scope *transactionScope) UpdateLastUsedWorkspace(
	ctx context.Context,
	userID uuid.UUID,
	workspaceID uuid.UUID,
) error {
	return scope.locked(func() error {
		return scope.users.UpdateLastUsedWorkspace(ctx, userID, workspaceID)
	})
}

func (scope *transactionScope) InitializeWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID) error {
	return scope.locked(func() error {
		return scope.workspaces.InitializeSettings(ctx, workspaceID)
	})
}

var _ workspaces.UnitOfWork = (*Manager)(nil)
