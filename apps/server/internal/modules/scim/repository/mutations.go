package scimrepository

import (
	"context"
	"errors"

	domain "github.com/complexus-tech/projects-api/internal/modules/scim/domain"
	sql "github.com/complexus-tech/projects-api/internal/modules/scim/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func membership(ctx context.Context, q *sql.Queries, scope domain.Scope, userID uuid.UUID, active bool, previousRole string) (string, bool, error) {
	account, err := q.LockAccount(ctx, sql.LockAccountParams{UserID: userID})
	if err != nil {
		return "", false, err
	}
	if !account.IsActive || account.IsSystem || account.IsInternal {
		return "", false, domain.ErrForbidden
	}
	role, err := q.CurrentRole(ctx, sql.CurrentRoleParams{WorkspaceID: scope.WorkspaceID, UserID: userID})
	exists := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", false, err
	}
	if exists {
		previousRole = role
	}
	if active {
		if exists {
			return previousRole, false, nil
		}
		err = q.ActivateMembership(ctx, sql.ActivateMembershipParams{WorkspaceID: scope.WorkspaceID, UserID: userID, Role: sql.UserRole(previousRole)})
		return previousRole, true, err
	}
	if !exists {
		return previousRole, false, nil
	}
	if userID == scope.IssuerID {
		return "", false, domain.ErrForbidden
	}
	if role == "admin" {
		count, err := q.CountAdmins(ctx, sql.CountAdminsParams{WorkspaceID: scope.WorkspaceID})
		if err != nil {
			return "", false, err
		}
		if count <= 1 {
			return "", false, domain.ErrForbidden
		}
	}
	if err := q.RemoveTeamMemberships(ctx, sql.RemoveTeamMembershipsParams{WorkspaceID: scope.WorkspaceID, UserID: userID}); err != nil {
		return "", false, err
	}
	err = q.RemoveMembership(ctx, sql.RemoveMembershipParams{WorkspaceID: scope.WorkspaceID, UserID: userID})
	return previousRole, true, err
}
func (r *Repository) Create(ctx context.Context, scope domain.Scope, input domain.Mutation) (domain.User, error) {
	var result domain.User
	err := r.within(ctx, true, func(q *sql.Queries) error {
		if err := authorized(ctx, q, scope, true); err != nil {
			return err
		}
		account, err := q.FindAccount(ctx, sql.FindAccountParams{Email: input.Email})
		if errors.Is(err, pgx.ErrNoRows) {
			row, createErr := q.CreateAccount(ctx, sql.CreateAccountParams{ID: uuid.New(), UserName: input.UserName, Email: input.Email, FullName: &input.FullName})
			err = createErr
			account = sql.FindAccountRow(row)
		}
		if err != nil {
			return err
		}
		if !account.IsActive || account.IsSystem || account.IsInternal {
			return domain.ErrForbidden
		}
		role, changed, err := membership(ctx, q, scope, account.UserID, input.Active, "member")
		if err != nil {
			return err
		}
		row, err := q.CreateUser(ctx, sql.CreateUserParams{ID: uuid.New(), WorkspaceID: scope.WorkspaceID, UserID: account.UserID, UserName: input.UserName, ExternalID: input.ExternalID, Active: input.Active, Profile: input.Profile, PreviousRole: role})
		if err != nil {
			return err
		}
		result = user(sql.GetUserRow(row))
		if changed {
			if err = q.MarkSeatSync(ctx, sql.MarkSeatSyncParams{WorkspaceID: scope.WorkspaceID}); err != nil {
				return err
			}
		}
		return audit(ctx, q, scope.WorkspaceID, scope.IssuerID, &scope.CredentialID, result.ID, "workspace.scim_user_created", map[string]any{"active": result.Active, "userId": result.UserID})
	})
	return result, err
}
func (r *Repository) Update(ctx context.Context, scope domain.Scope, id uuid.UUID, expected int64, input domain.Mutation, deleted bool) (domain.User, error) {
	var result domain.User
	err := r.within(ctx, true, func(q *sql.Queries) error {
		if err := authorized(ctx, q, scope, true); err != nil {
			return err
		}
		old, err := q.LockUser(ctx, sql.LockUserParams{WorkspaceID: scope.WorkspaceID, ID: id})
		if err != nil {
			return err
		}
		if old.Version != expected {
			return domain.ErrChanged
		}
		role, changed, err := membership(ctx, q, scope, old.UserID, input.Active, old.PreviousRole)
		if err != nil {
			return err
		}
		row, err := q.UpdateUser(ctx, sql.UpdateUserParams{WorkspaceID: scope.WorkspaceID, ID: id, ExpectedVersion: expected, UserName: input.UserName, ExternalID: input.ExternalID, Active: input.Active, Profile: input.Profile, PreviousRole: role, Deleted: deleted})
		if err != nil {
			return err
		}
		result = user(sql.GetUserRow(row))
		if changed {
			if err = q.MarkSeatSync(ctx, sql.MarkSeatSyncParams{WorkspaceID: scope.WorkspaceID}); err != nil {
				return err
			}
		}
		operation := "workspace.scim_user_updated"
		if deleted {
			operation = "workspace.scim_user_deleted"
		}
		return audit(ctx, q, scope.WorkspaceID, scope.IssuerID, &scope.CredentialID, id, operation, map[string]any{"active": result.Active, "userId": result.UserID, "version": result.Version})
	})
	return result, err
}
func (r *Repository) PendingSeats(ctx context.Context, scope domain.AdminScope) (int64, error) {
	var generation int64
	err := r.within(ctx, false, func(q *sql.Queries) error {
		if err := admin(ctx, q, scope, false); err != nil {
			return err
		}
		row, err := q.SeatSync(ctx, sql.SeatSyncParams{WorkspaceID: scope.WorkspaceID})
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		generation = row.Generation
		return err
	})
	return generation, err
}
func (r *Repository) FinishSeats(ctx context.Context, scope domain.AdminScope, generation int64, succeeded bool) error {
	return r.within(ctx, true, func(q *sql.Queries) error {
		if err := admin(ctx, q, scope, true); err != nil {
			return err
		}
		if succeeded {
			return q.FinishSeatSync(ctx, sql.FinishSeatSyncParams{WorkspaceID: scope.WorkspaceID, Generation: generation})
		}
		return q.FailSeatSync(ctx, sql.FailSeatSyncParams{WorkspaceID: scope.WorkspaceID, Generation: generation})
	})
}
