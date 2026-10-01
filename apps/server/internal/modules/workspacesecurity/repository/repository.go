package workspacesecurityrepository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/workspacesecurity/domain"
	sql "github.com/complexus-tech/projects-api/internal/modules/workspacesecurity/repository/sqlc"
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
	var postgres *pgconn.PgError
	if errors.As(err, &postgres) && (postgres.Code == "40001" || postgres.Code == "40P01" || postgres.Code == "23505") {
		return domain.ErrConflict
	}
	return err
}
func admin(ctx context.Context, q *sql.Queries, scope domain.Scope, write bool) (string, error) {
	var email string
	var err error
	if write {
		email, err = q.LockWorkspaceAdmin(ctx, sql.LockWorkspaceAdminParams{WorkspaceID: scope.WorkspaceID, ActorID: scope.ActorID})
	} else {
		email, err = q.AuthorizeWorkspaceAdmin(ctx, sql.AuthorizeWorkspaceAdminParams{WorkspaceID: scope.WorkspaceID, ActorID: scope.ActorID})
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrForbidden
	}
	return email, err
}
func appendAudit(ctx context.Context, q *sql.Queries, scope domain.Scope, resourceType string, resourceID uuid.UUID, operation string, metadata any) error {
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	return q.AppendSecurityAudit(ctx, sql.AppendSecurityAuditParams{EventID: uuid.New(), WorkspaceID: scope.WorkspaceID, ActorID: scope.ActorID, ResourceType: resourceType, ResourceID: resourceID, Operation: operation, Metadata: encoded})
}

func (r *Repository) CheckSession(ctx context.Context, scope domain.Scope, session domain.SessionIdentity) error {
	return r.within(ctx, true, func(q *sql.Queries) error {
		row, err := q.SessionAccessState(ctx, sql.SessionAccessStateParams{WorkspaceID: scope.WorkspaceID, ActorID: scope.ActorID, SessionID: session.ID})
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrForbidden
		}
		if err != nil {
			return err
		}
		state := domain.AccessState{Email: row.Email, Role: row.Role, Policy: domain.Policy{AllowedDomains: row.AllowedDomains, AllowGuests: row.AllowGuests, MaxSessionAgeHours: row.MaxSessionAgeHours}, RevokedBefore: row.RevokedBefore, RevokedAt: row.RevokedAt}
		if err := domain.CheckAccess(state, session, time.Now().UTC()); err != nil {
			return err
		}
		if session.ID == uuid.Nil || session.AuthenticatedAt.IsZero() || session.ExpiresAt.IsZero() {
			return nil
		}
		count, err := q.TrackBrowserSession(ctx, sql.TrackBrowserSessionParams{WorkspaceID: scope.WorkspaceID, ActorID: scope.ActorID, SessionID: session.ID, AuthenticatedAt: session.AuthenticatedAt, ExpiresAt: session.ExpiresAt})
		if err == nil && count != 1 {
			return domain.ErrForbidden
		}
		return err
	})
}
func readPolicy(ctx context.Context, q *sql.Queries, workspaceID uuid.UUID) (domain.Policy, error) {
	row, err := q.GetPolicy(ctx, sql.GetPolicyParams{WorkspaceID: workspaceID})
	return domain.Policy{AllowedDomains: row.AllowedDomains, AllowGuests: row.AllowGuests, MaxSessionAgeHours: row.MaxSessionAgeHours, Version: row.Version, UpdatedAt: row.UpdatedAt}, err
}
func (r *Repository) Policy(ctx context.Context, scope domain.Scope) (domain.Policy, error) {
	var result domain.Policy
	err := r.within(ctx, false, func(q *sql.Queries) error {
		if _, err := admin(ctx, q, scope, false); err != nil {
			return err
		}
		var err error
		result, err = readPolicy(ctx, q, scope.WorkspaceID)
		return err
	})
	return result, err
}
func (r *Repository) UpdatePolicy(ctx context.Context, scope domain.Scope, input domain.PolicyUpdate) (domain.Policy, error) {
	input, err := domain.NormalizePolicy(input)
	if err != nil {
		return domain.Policy{}, err
	}
	var result domain.Policy
	err = r.within(ctx, true, func(q *sql.Queries) error {
		email, err := admin(ctx, q, scope, true)
		if err != nil {
			return err
		}
		if !domain.EmailAllowed(email, input.AllowedDomains) {
			return errors.Join(domain.ErrInvalid, errors.New("allowed domains must include your current email domain"))
		}
		previous, err := readPolicy(ctx, q, scope.WorkspaceID)
		if err != nil {
			return err
		}
		if previous.Version != input.ExpectedVersion {
			return domain.ErrConflict
		}
		row, err := q.PutPolicy(ctx, sql.PutPolicyParams{WorkspaceID: scope.WorkspaceID, AllowedDomains: input.AllowedDomains, AllowGuests: input.AllowGuests, MaxSessionAgeHours: input.MaxSessionAgeHours})
		if err != nil {
			return err
		}
		result = domain.Policy{AllowedDomains: row.AllowedDomains, AllowGuests: row.AllowGuests, MaxSessionAgeHours: row.MaxSessionAgeHours, Version: row.Version, UpdatedAt: &row.UpdatedAt}
		return appendAudit(ctx, q, scope, "workspace_policy", scope.WorkspaceID, "workspace.security_policy_updated", map[string]any{"previous": previous, "current": result})
	})
	return result, err
}
func (r *Repository) Sessions(ctx context.Context, scope domain.Scope, userID *uuid.UUID, includeRevoked bool) (domain.SessionList, error) {
	result := domain.SessionList{Items: []domain.Session{}}
	err := r.within(ctx, false, func(q *sql.Queries) error {
		if _, err := admin(ctx, q, scope, false); err != nil {
			return err
		}
		rows, err := q.ListBrowserSessions(ctx, sql.ListBrowserSessionsParams{WorkspaceID: scope.WorkspaceID, UserID: userID, IncludeRevoked: includeRevoked})
		if err != nil {
			return err
		}
		if len(rows) > 500 {
			result.HasMore = true
			rows = rows[:500]
		}
		for _, row := range rows {
			name := row.Email
			if row.FullName != nil && *row.FullName != "" {
				name = *row.FullName
			}
			result.Items = append(result.Items, domain.Session{ID: row.SessionID, UserID: row.UserID, Name: name, Email: row.Email, Role: row.Role, AuthenticatedAt: row.AuthenticatedAt, LastSeenAt: row.LastSeenAt, ExpiresAt: row.ExpiresAt, RevokedAt: row.RevokedAt, RevokedBefore: row.RevokedBefore})
		}
		return nil
	})
	return result, err
}
func (r *Repository) RevokeSession(ctx context.Context, scope domain.Scope, id uuid.UUID, reason string) error {
	if err := domain.ValidateReason(reason); err != nil {
		return err
	}
	return r.within(ctx, true, func(q *sql.Queries) error {
		if _, err := admin(ctx, q, scope, true); err != nil {
			return err
		}
		row, err := q.LockBrowserSession(ctx, sql.LockBrowserSessionParams{WorkspaceID: scope.WorkspaceID, SessionID: id})
		if err != nil {
			return err
		}
		if row.RevokedAt != nil {
			return nil
		}
		count, err := q.RevokeBrowserSession(ctx, sql.RevokeBrowserSessionParams{WorkspaceID: scope.WorkspaceID, SessionID: id})
		if err != nil {
			return err
		}
		if count != 1 {
			return domain.ErrConflict
		}
		return appendAudit(ctx, q, scope, "browser_session", id, "workspace.session_revoked", map[string]any{"userId": row.UserID, "reason": reason})
	})
}
func (r *Repository) RevokeMember(ctx context.Context, scope domain.Scope, id uuid.UUID, reason string) error {
	if err := domain.ValidateReason(reason); err != nil {
		return err
	}
	return r.within(ctx, true, func(q *sql.Queries) error {
		if _, err := admin(ctx, q, scope, true); err != nil {
			return err
		}
		if _, err := q.LockTargetMember(ctx, sql.LockTargetMemberParams{WorkspaceID: scope.WorkspaceID, UserID: id}); err != nil {
			return err
		}
		at, err := q.RevokeMemberSessions(ctx, sql.RevokeMemberSessionsParams{WorkspaceID: scope.WorkspaceID, UserID: id})
		if err != nil {
			return err
		}
		return appendAudit(ctx, q, scope, "workspace_member", id, "workspace.member_sessions_revoked", map[string]any{"revokedBefore": at, "reason": reason})
	})
}
