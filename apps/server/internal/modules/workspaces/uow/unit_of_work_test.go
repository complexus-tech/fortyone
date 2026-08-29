package workspaceuow

import (
	"context"
	"errors"
	"testing"
	"time"

	teamsrepository "github.com/complexus-tech/projects-api/internal/modules/teams/repository"
	usersrepository "github.com/complexus-tech/projects-api/internal/modules/users/repository"
	workspacesrepository "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5"
)

type nilBeginner struct{}

func (nilBeginner) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	return nil, errors.New("not used")
}

type nilWorkspaceBinder struct{}
type nilTeamBinder struct{}
type nilUserBinder struct{}

func TestNewRequiresEveryTransactionBinder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		beginner   platformdatabase.Beginner
		workspaces workspaceBinder
		teams      teamBinder
		users      userBinder
	}{
		{name: "beginner", workspaces: nilWorkspaceBinder{}, teams: nilTeamBinder{}, users: nilUserBinder{}},
		{name: "workspaces", beginner: nilBeginner{}, teams: nilTeamBinder{}, users: nilUserBinder{}},
		{name: "teams", beginner: nilBeginner{}, workspaces: nilWorkspaceBinder{}, users: nilUserBinder{}},
		{name: "users", beginner: nilBeginner{}, workspaces: nilWorkspaceBinder{}, teams: nilTeamBinder{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := New(test.beginner, test.workspaces, test.teams, test.users); err == nil {
				t.Fatal("New error = nil, want missing dependency error")
			}
		})
	}
}

func (nilWorkspaceBinder) BindWorkspaceTransaction(pgx.Tx) (workspacesrepository.WorkspaceTransaction, error) {
	return nil, nil
}

func (nilTeamBinder) BindWorkspaceTransaction(pgx.Tx) (teamsrepository.WorkspaceTransaction, error) {
	return nil, nil
}

func (nilUserBinder) BindWorkspaceTransaction(pgx.Tx) (usersrepository.WorkspaceTransaction, error) {
	return nil, nil
}

func TestTransactionScopeCloseWaitsForInFlightCallAndRejectsLateWork(t *testing.T) {
	t.Parallel()

	scope := &transactionScope{}
	scope.open.Store(true)
	entered := make(chan struct{})
	release := make(chan struct{})
	operationDone := make(chan error, 1)
	go func() {
		operationDone <- scope.locked(func() error {
			close(entered)
			<-release
			return nil
		})
	}()
	<-entered

	closed := make(chan struct{})
	go func() {
		scope.close()
		close(closed)
	}()
	select {
	case <-closed:
		t.Fatal("scope closed while a capability call was still in flight")
	case <-time.After(25 * time.Millisecond):
	}
	close(release)
	if err := <-operationDone; err != nil {
		t.Fatalf("in-flight operation error = %v", err)
	}
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("scope close did not finish after the operation completed")
	}
	if err := scope.locked(func() error { return nil }); !errors.Is(err, ErrTransactionScopeClosed) {
		t.Fatalf("late operation error = %v, want ErrTransactionScopeClosed", err)
	}
}
