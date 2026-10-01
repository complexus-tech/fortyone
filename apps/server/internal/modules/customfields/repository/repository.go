package customfieldsrepository

import (
	"context"
	"errors"
	"fmt"

	domain "github.com/complexus-tech/projects-api/internal/modules/customfields/domain"
	sql "github.com/complexus-tech/projects-api/internal/modules/customfields/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	transactor database.Transactor
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{transactor: database.NewTransactor(pool)}
}

func (r *Repository) within(ctx context.Context, write bool, fn func(*sql.Queries) error) error {
	options := pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly}
	if write {
		options = pgx.TxOptions{IsoLevel: pgx.ReadCommitted}
	}
	err := r.transactor.WithinTransaction(ctx, options, func(tx pgx.Tx) error { return fn(sql.New(tx)) })
	return mapError(err)
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var postgres *pgconn.PgError
	if errors.As(err, &postgres) {
		switch postgres.Code {
		case "23505", "40001", "40P01":
			return domain.ErrConflict
		case "23503", "23514", "22003", "22P02":
			return domain.ErrInvalid
		}
	}
	return err
}

func authorize(ctx context.Context, q *sql.Queries, scope domain.Scope, write, admin bool) error {
	var role sql.UserRole
	var err error
	if write {
		role, err = q.AuthorizeTeamMutation(ctx, sql.AuthorizeTeamMutationParams{ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, TeamID: scope.TeamID})
	} else {
		role, err = q.AuthorizeTeam(ctx, sql.AuthorizeTeamParams{ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, TeamID: scope.TeamID})
	}
	if err != nil {
		return mapError(err)
	}
	if admin && role != sql.UserRoleAdmin {
		return domain.ErrForbidden
	}
	if write && role != sql.UserRoleAdmin && role != sql.UserRoleMember {
		return domain.ErrForbidden
	}
	if write && role != sql.UserRoleAdmin {
		if _, err := q.LockActorTeamMembership(ctx, sql.LockActorTeamMembershipParams{TeamID: scope.TeamID, ActorID: scope.ActorID}); err != nil {
			return mapError(err)
		}
	}
	if !write && role != sql.UserRoleAdmin && role != sql.UserRoleMember && role != sql.UserRoleGuest {
		return domain.ErrForbidden
	}
	return nil
}

func fieldFromRow(row sql.ListFieldsRow) domain.Field {
	return domain.Field{ID: row.ID, TeamID: row.TeamID, Name: row.Name, Type: domain.FieldType(row.FieldType), Currency: row.Currency, Icon: row.Icon, ShowOnCreate: row.ShowOnCreate, ArchivedAt: row.ArchivedAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt, Options: []domain.Option{}}
}

func listFields(ctx context.Context, q *sql.Queries, scope domain.Scope) ([]domain.Field, error) {
	rows, err := q.ListFields(ctx, sql.ListFieldsParams{WorkspaceID: scope.WorkspaceID, TeamID: scope.TeamID})
	if err != nil {
		return nil, fmt.Errorf("list custom fields: %w", err)
	}
	options, err := q.ListFieldOptions(ctx, sql.ListFieldOptionsParams{WorkspaceID: scope.WorkspaceID, TeamID: scope.TeamID})
	if err != nil {
		return nil, fmt.Errorf("list custom field options: %w", err)
	}
	byField := make(map[[16]byte][]domain.Option)
	for _, option := range options {
		byField[option.FieldID] = append(byField[option.FieldID], domain.Option{ID: option.ID, Name: option.Name, ArchivedAt: option.ArchivedAt})
	}
	fields := make([]domain.Field, 0, len(rows))
	for _, row := range rows {
		field := fieldFromRow(row)
		if choices := byField[field.ID]; choices != nil {
			field.Options = choices
		}
		fields = append(fields, field)
	}
	return fields, nil
}

func (r *Repository) List(ctx context.Context, scope domain.Scope) ([]domain.Field, error) {
	var result []domain.Field
	err := r.within(ctx, false, func(q *sql.Queries) error {
		if err := authorize(ctx, q, scope, false, false); err != nil {
			return err
		}
		var err error
		result, err = listFields(ctx, q, scope)
		return err
	})
	return result, err
}
