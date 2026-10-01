package scimrepository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/scim/domain"
	sql "github.com/complexus-tech/projects-api/internal/modules/scim/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ transactor database.Transactor }

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{transactor: database.NewTransactor(pool)}
}
func (r *Repository) within(ctx context.Context, write bool, fn func(*sql.Queries) error) error {
	options := pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly}
	if write {
		options = pgx.TxOptions{IsoLevel: pgx.Serializable}
	}
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		err = r.transactor.WithinTransaction(ctx, options, func(tx pgx.Tx) error { return fn(sql.New(tx)) })
		if !write || !database.IsRetryableTransactionError(err) {
			break
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if database.Classify(err) == database.ErrorClassUniqueViolation {
		return domain.ErrConflict
	}
	if database.IsRetryableTransactionError(err) {
		return domain.ErrChanged
	}
	return err
}
func admin(ctx context.Context, q *sql.Queries, scope domain.AdminScope, write bool) error {
	var err error
	if write {
		_, err = q.LockAdmin(ctx, sql.LockAdminParams{WorkspaceID: scope.WorkspaceID, ActorID: scope.ActorID})
	} else {
		_, err = q.AuthorizeAdmin(ctx, sql.AuthorizeAdminParams{WorkspaceID: scope.WorkspaceID, ActorID: scope.ActorID})
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrForbidden
	}
	return err
}
func authorized(ctx context.Context, q *sql.Queries, scope domain.Scope, write bool) error {
	if write {
		if err := admin(ctx, q, domain.AdminScope{WorkspaceID: scope.WorkspaceID, ActorID: scope.IssuerID}, true); err != nil {
			if errors.Is(err, domain.ErrForbidden) {
				return domain.ErrUnauthorized
			}
			return err
		}
		_, err := q.LockCredential(ctx, sql.LockCredentialParams{CredentialID: scope.CredentialID, WorkspaceID: scope.WorkspaceID, IssuerID: scope.IssuerID})
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrUnauthorized
		}
		return err
	}
	_, err := q.CheckScope(ctx, sql.CheckScopeParams{CredentialID: scope.CredentialID, WorkspaceID: scope.WorkspaceID, IssuerID: scope.IssuerID})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrUnauthorized
	}
	return err
}
func (r *Repository) Authenticate(ctx context.Context, slug string, digest []byte) (domain.Scope, error) {
	var scope domain.Scope
	err := r.within(ctx, false, func(q *sql.Queries) error {
		row, err := q.Authenticate(ctx, sql.AuthenticateParams{Slug: slug, Digest: digest})
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrUnauthorized
		}
		scope = domain.Scope{WorkspaceID: row.WorkspaceID, CredentialID: row.ID, IssuerID: row.IssuerID}
		return err
	})
	return scope, err
}
func (r *Repository) AuthorizeAdmin(ctx context.Context, scope domain.AdminScope) error {
	return r.within(ctx, false, func(q *sql.Queries) error { return admin(ctx, q, scope, false) })
}
func credential(row sql.ListCredentialsRow) domain.Credential {
	return domain.Credential{ID: row.ID, Name: row.Name, Prefix: row.TokenPrefix, CreatedAt: row.CreatedAt, ExpiresAt: row.ExpiresAt, RevokedAt: row.RevokedAt}
}
func user(row sql.GetUserRow) domain.User {
	return domain.User{ID: row.ID, WorkspaceID: row.WorkspaceID, UserID: row.UserID, UserName: row.UserName, ExternalID: row.ExternalID, Active: row.Active, Profile: row.Profile, PreviousRole: row.PreviousRole, Version: row.Version, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}
func audit(ctx context.Context, q *sql.Queries, workspace, actor uuid.UUID, credential *uuid.UUID, resource uuid.UUID, operation string, metadata any) error {
	body, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	return q.AppendAudit(ctx, sql.AppendAuditParams{ID: uuid.New(), WorkspaceID: workspace, ActorID: actor, CredentialID: credential, ResourceID: resource, Operation: operation, Metadata: body})
}
func (r *Repository) Status(ctx context.Context, scope domain.AdminScope) (domain.Status, error) {
	result := domain.Status{Credentials: []domain.Credential{}}
	err := r.within(ctx, false, func(q *sql.Queries) error {
		if err := admin(ctx, q, scope, false); err != nil {
			return err
		}
		rows, err := q.ListCredentials(ctx, sql.ListCredentialsParams{WorkspaceID: scope.WorkspaceID})
		if err != nil {
			return err
		}
		result.Credentials = make([]domain.Credential, 0, len(rows))
		for _, row := range rows {
			result.Credentials = append(result.Credentials, credential(row))
		}
		result.ManagedUsers, err = q.CountUsers(ctx, sql.CountUsersParams{WorkspaceID: scope.WorkspaceID, Field: "", Value: ""})
		if err != nil {
			return err
		}
		pending, err := q.SeatSync(ctx, sql.SeatSyncParams{WorkspaceID: scope.WorkspaceID})
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		result.PendingSeatSync = err == nil
		result.SeatSyncError = pending.LastError
		return err
	})
	return result, err
}
func (r *Repository) Mint(ctx context.Context, scope domain.AdminScope, name string, digest []byte, prefix string, expiry time.Time) (domain.Credential, error) {
	var result domain.Credential
	err := r.within(ctx, true, func(q *sql.Queries) error {
		if err := admin(ctx, q, scope, true); err != nil {
			return err
		}
		count, err := q.CountLiveCredentials(ctx, sql.CountLiveCredentialsParams{WorkspaceID: scope.WorkspaceID})
		if err != nil {
			return err
		}
		if count >= 10 {
			return domain.ErrConflict
		}
		row, err := q.CreateCredential(ctx, sql.CreateCredentialParams{ID: uuid.New(), WorkspaceID: scope.WorkspaceID, IssuerID: scope.ActorID, Name: name, Digest: digest, Prefix: prefix, ExpiresAt: expiry})
		if err != nil {
			return err
		}
		result = credential(sql.ListCredentialsRow(row))
		return audit(ctx, q, scope.WorkspaceID, scope.ActorID, nil, result.ID, "workspace.scim_credential_created", map[string]any{"expiresAt": expiry})
	})
	return result, err
}
func (r *Repository) Revoke(ctx context.Context, scope domain.AdminScope, id uuid.UUID) error {
	return r.within(ctx, true, func(q *sql.Queries) error {
		if err := admin(ctx, q, scope, true); err != nil {
			return err
		}
		count, err := q.RevokeCredential(ctx, sql.RevokeCredentialParams{WorkspaceID: scope.WorkspaceID, ID: id})
		if err != nil {
			return err
		}
		if count == 0 {
			return domain.ErrNotFound
		}
		return audit(ctx, q, scope.WorkspaceID, scope.ActorID, nil, id, "workspace.scim_credential_revoked", map[string]any{})
	})
}
func (r *Repository) Get(ctx context.Context, scope domain.Scope, id uuid.UUID) (domain.User, error) {
	var result domain.User
	err := r.within(ctx, false, func(q *sql.Queries) error {
		if err := authorized(ctx, q, scope, false); err != nil {
			return err
		}
		row, err := q.GetUser(ctx, sql.GetUserParams{WorkspaceID: scope.WorkspaceID, ID: id})
		result = user(row)
		return err
	})
	return result, err
}
func (r *Repository) List(ctx context.Context, scope domain.Scope, page domain.Page) ([]domain.User, int64, error) {
	result := []domain.User{}
	var total int64
	err := r.within(ctx, false, func(q *sql.Queries) error {
		if err := authorized(ctx, q, scope, false); err != nil {
			return err
		}
		var err error
		total, err = q.CountUsers(ctx, sql.CountUsersParams{WorkspaceID: scope.WorkspaceID, Field: page.Filter.Field, Value: page.Filter.Value, Active: page.Filter.Active})
		if err != nil {
			return err
		}
		rows, err := q.ListUsers(ctx, sql.ListUsersParams{WorkspaceID: scope.WorkspaceID, Field: page.Filter.Field, Value: page.Filter.Value, Active: page.Filter.Active, PageLimit: int32(page.Count), PageOffset: int32(page.StartIndex - 1)})
		if err != nil {
			return err
		}
		result = make([]domain.User, 0, len(rows))
		for _, row := range rows {
			result = append(result, user(sql.GetUserRow(row)))
		}
		return nil
	})
	return result, total, err
}
