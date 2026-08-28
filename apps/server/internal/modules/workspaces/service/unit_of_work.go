package workspaces

import (
	"context"

	"github.com/google/uuid"
)

type DefaultTeam struct {
	Name      string
	Color     string
	Code      string
	Workspace uuid.UUID
}

type CreatedTeam struct {
	ID uuid.UUID
}

// Transaction exposes only the persistence capabilities that participate in
// workspace creation. Implementations bind every method to the same database
// transaction and reject calls after the callback returns.
type Transaction interface {
	CreateWorkspace(context.Context, CoreWorkspace, uuid.UUID) (CoreWorkspace, error)
	AddWorkspaceMember(context.Context, uuid.UUID, uuid.UUID, string) error
	CreateTeam(context.Context, DefaultTeam) (CreatedTeam, error)
	AddTeamMember(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error
	UpdateLastUsedWorkspace(context.Context, uuid.UUID, uuid.UUID) error
	InitializeWorkspaceSettings(context.Context, uuid.UUID) error
}

// UnitOfWork owns the workspace-creation transaction boundary. The callback
// must not retain Transaction; concrete implementations enforce that rule.
type UnitOfWork interface {
	WithinTransaction(context.Context, func(Transaction) error) error
}

type unavailableUnitOfWork struct{}

func (unavailableUnitOfWork) WithinTransaction(context.Context, func(Transaction) error) error {
	return ErrTx
}
