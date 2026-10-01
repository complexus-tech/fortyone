package enterprisessorepository

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	domain "github.com/complexus-tech/projects-api/internal/modules/enterprisesso/domain"
	sql "github.com/complexus-tech/projects-api/internal/modules/enterprisesso/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ transactor database.Transactor }

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{transactor: database.NewTransactor(pool)}
}
func (r *Repository) within(ctx context.Context, write bool, fn func(*sql.Queries) error) error {
	options := pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly}
	if write {
		options = pgx.TxOptions{IsoLevel: pgx.ReadCommitted}
	}
	err := r.transactor.WithinTransaction(ctx, options, func(tx pgx.Tx) error { return fn(sql.New(tx)) })
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pgerr *pgconn.PgError
	if errors.As(err, &pgerr) && (pgerr.Code == "23505" || pgerr.Code == "40001" || pgerr.Code == "40P01") {
		return domain.ErrConflict
	}
	return err
}
func admin(ctx context.Context, q *sql.Queries, scope domain.Scope, write bool) error {
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
func connection(row sql.GetConnectionRow) domain.Connection {
	return domain.Connection{ID: row.ID, WorkspaceID: row.WorkspaceID, Issuer: row.Issuer, ClientID: row.ClientID, SecretEnvelope: row.ClientSecretEnvelope, Enabled: row.Enabled, RequireSSO: row.RequireSso, Generation: row.Generation, Version: row.Version, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}
func audit(ctx context.Context, q *sql.Queries, scope domain.Scope, id uuid.UUID, operation string, metadata any) error {
	body, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	return q.AppendAudit(ctx, sql.AppendAuditParams{ID: uuid.New(), WorkspaceID: scope.WorkspaceID, ActorID: scope.ActorID, ConnectionID: id, Operation: operation, Metadata: body})
}
func (r *Repository) AuthorizeAdmin(ctx context.Context, scope domain.Scope) error {
	return r.within(ctx, false, func(q *sql.Queries) error { return admin(ctx, q, scope, false) })
}
func (r *Repository) Get(ctx context.Context, scope domain.Scope) (*domain.Connection, error) {
	var result *domain.Connection
	err := r.within(ctx, false, func(q *sql.Queries) error {
		if err := admin(ctx, q, scope, false); err != nil {
			return err
		}
		row, err := q.GetConnection(ctx, sql.GetConnectionParams{WorkspaceID: scope.WorkspaceID})
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		value := connection(row)
		result = &value
		return nil
	})
	return result, err
}
func (r *Repository) Public(ctx context.Context, slug string) (domain.Connection, error) {
	var result domain.Connection
	err := r.within(ctx, false, func(q *sql.Queries) error {
		row, err := q.PublicConnection(ctx, sql.PublicConnectionParams{Slug: slug})
		result = connection(sql.GetConnectionRow(row))
		return err
	})
	return result, err
}
func (r *Repository) Create(ctx context.Context, scope domain.Scope, input domain.Connection) (domain.Connection, error) {
	var result domain.Connection
	err := r.within(ctx, true, func(q *sql.Queries) error {
		if err := admin(ctx, q, scope, true); err != nil {
			return err
		}
		row, err := q.CreateConnection(ctx, sql.CreateConnectionParams{ID: input.ID, WorkspaceID: scope.WorkspaceID, Issuer: input.Issuer, ClientID: input.ClientID, SecretEnvelope: input.SecretEnvelope})
		if err != nil {
			return err
		}
		result = connection(sql.GetConnectionRow(row))
		return audit(ctx, q, scope, result.ID, "workspace.sso_configured", map[string]any{"issuer": result.Issuer, "clientId": result.ClientID})
	})
	return result, err
}
func (r *Repository) Update(ctx context.Context, scope domain.Scope, input domain.Connection, expected int64) (domain.Connection, error) {
	var result domain.Connection
	err := r.within(ctx, true, func(q *sql.Queries) error {
		if err := admin(ctx, q, scope, true); err != nil {
			return err
		}
		previous, err := q.LockConnection(ctx, sql.LockConnectionParams{WorkspaceID: scope.WorkspaceID})
		if err != nil {
			return err
		}
		if previous.ID != input.ID || previous.Version != expected || previous.Issuer != input.Issuer || previous.ClientID != input.ClientID {
			return domain.ErrConflict
		}
		row, err := q.UpdateConnection(ctx, sql.UpdateConnectionParams{ID: input.ID, WorkspaceID: scope.WorkspaceID, Enabled: input.Enabled, RequireSso: input.RequireSSO, Generation: input.Generation, SecretEnvelope: input.SecretEnvelope, ExpectedVersion: expected})
		if err != nil {
			return err
		}
		result = connection(sql.GetConnectionRow(row))
		return audit(ctx, q, scope, result.ID, "workspace.sso_updated", map[string]any{"enabled": result.Enabled, "requireSSO": result.RequireSSO, "generation": result.Generation})
	})
	return result, err
}
func (r *Repository) Archive(ctx context.Context, scope domain.Scope, reason string) error {
	return r.within(ctx, true, func(q *sql.Queries) error {
		if err := admin(ctx, q, scope, true); err != nil {
			return err
		}
		row, err := q.LockConnection(ctx, sql.LockConnectionParams{WorkspaceID: scope.WorkspaceID})
		if err != nil {
			return err
		}
		if _, err := q.ArchiveConnection(ctx, sql.ArchiveConnectionParams{WorkspaceID: scope.WorkspaceID}); err != nil {
			return err
		}
		return audit(ctx, q, scope, row.ID, "workspace.sso_removed", map[string]any{"reason": reason})
	})
}

func (r *Repository) Authenticate(ctx context.Context, attempt domain.Attempt, identity domain.Identity) (uuid.UUID, error) {
	var userID uuid.UUID
	err := r.within(ctx, true, func(q *sql.Queries) error {
		if _, err := q.LockLiveWorkspace(ctx, sql.LockLiveWorkspaceParams{WorkspaceID: attempt.WorkspaceID}); err != nil {
			return err
		}
		current, err := q.LockConnection(ctx, sql.LockConnectionParams{WorkspaceID: attempt.WorkspaceID})
		if err != nil {
			return err
		}
		if !current.Enabled || current.ID != attempt.ConnectionID || current.Generation != attempt.Generation {
			return domain.ErrConflict
		}
		linked, err := q.ResolveLinkedIdentity(ctx, sql.ResolveLinkedIdentityParams{ConnectionID: current.ID, WorkspaceID: attempt.WorkspaceID, Subject: identity.Subject})
		if err == nil {
			if attempt.LinkUserID != uuid.Nil && attempt.LinkUserID != linked {
				return domain.ErrForbidden
			}
			userID = linked
			if err := q.TouchIdentity(ctx, sql.TouchIdentityParams{ConnectionID: current.ID, WorkspaceID: attempt.WorkspaceID, Subject: identity.Subject, AuthenticatedAt: identity.AuthenticatedAt}); err != nil {
				return err
			}
		} else if errors.Is(err, pgx.ErrNoRows) {
			if attempt.LinkUserID == uuid.Nil {
				return domain.ErrLinkRequired
			}
			email, err := q.LockLinkAccount(ctx, sql.LockLinkAccountParams{WorkspaceID: attempt.WorkspaceID, UserID: attempt.LinkUserID})
			if err != nil {
				return err
			}
			if !strings.EqualFold(email, identity.Email) {
				return domain.ErrForbidden
			}
			userID = attempt.LinkUserID
			if err := q.LinkIdentity(ctx, sql.LinkIdentityParams{ConnectionID: current.ID, WorkspaceID: attempt.WorkspaceID, Subject: identity.Subject, UserID: userID, AuthenticatedAt: identity.AuthenticatedAt}); err != nil {
				return err
			}
			if err := audit(ctx, q, domain.Scope{ActorID: userID, WorkspaceID: attempt.WorkspaceID}, current.ID, "workspace.sso_identity_linked", map[string]any{}); err != nil {
				return err
			}
		} else {
			return err
		}
		return audit(ctx, q, domain.Scope{ActorID: userID, WorkspaceID: attempt.WorkspaceID}, current.ID, "workspace.sso_signed_in", map[string]any{"generation": current.Generation})
	})
	return userID, err
}
func (r *Repository) CheckSession(ctx context.Context, scope domain.Scope, proof domain.SessionProof) error {
	return r.within(ctx, false, func(q *sql.Queries) error {
		required, err := q.RequiredConnection(ctx, sql.RequiredConnectionParams{WorkspaceID: scope.WorkspaceID, ActorID: scope.ActorID})
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		if proof.ConnectionID != required.ID || proof.Generation != required.Generation || proof.AuthenticatedAt.IsZero() {
			return domain.ErrForbidden
		}
		return nil
	})
}
