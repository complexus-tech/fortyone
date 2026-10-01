//go:build integration

package workspacesecurityrepository_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/workspacesecurity/domain"
	repository "github.com/complexus-tech/projects-api/internal/modules/workspacesecurity/repository"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type fixture struct {
	pool                                    *pgxpool.Pool
	repo                                    *repository.Repository
	scope                                   domain.Scope
	member, guest, inactive, otherWorkspace uuid.UUID
}

func seed(t *testing.T) fixture {
	t.Helper()
	database := testkit.NewPostgres(t)
	f := fixture{pool: database.Pool, repo: repository.New(database.Pool), scope: domain.Scope{ActorID: uuid.New(), WorkspaceID: uuid.New()}, member: uuid.New(), guest: uuid.New(), inactive: uuid.New(), otherWorkspace: uuid.New()}
	for _, workspace := range []uuid.UUID{f.scope.WorkspaceID, f.otherWorkspace} {
		exec(t, f.pool, `INSERT INTO workspaces(workspace_id,name,slug) VALUES ($1,'Security',$2)`, workspace, uuid.NewString())
	}
	for _, entry := range []struct {
		id     uuid.UUID
		role   string
		active bool
	}{{f.scope.ActorID, "admin", true}, {f.member, "member", true}, {f.guest, "guest", true}, {f.inactive, "admin", false}} {
		exec(t, f.pool, `INSERT INTO users(user_id,username,email,full_name,is_active) VALUES ($1,$2,$3,'Person',$4)`, entry.id, entry.id.String(), entry.id.String()+"@example.com", entry.active)
		for _, workspace := range []uuid.UUID{f.scope.WorkspaceID, f.otherWorkspace} {
			exec(t, f.pool, `INSERT INTO workspace_members(workspace_id,user_id,role) VALUES ($1,$2,CAST($3 AS user_role))`, workspace, entry.id, entry.role)
		}
	}
	return f
}
func exec(t *testing.T, pool *pgxpool.Pool, query string, args ...any) {
	t.Helper()
	if _, err := pool.Exec(t.Context(), query, args...); err != nil {
		t.Fatal(err)
	}
}
func identity() domain.SessionIdentity {
	now := time.Now().UTC()
	return domain.SessionIdentity{ID: uuid.New(), AuthenticatedAt: now.Add(-time.Hour), ExpiresAt: now.Add(24 * time.Hour)}
}
func TestTenantSessionRevocationAndAuthoritativePolicies(t *testing.T) {
	f := seed(t)
	ctx := t.Context()
	member := f.scope
	member.ActorID = f.member
	session := identity()
	if err := f.repo.CheckSession(ctx, member, session); err != nil {
		t.Fatal(err)
	}
	other := member
	other.WorkspaceID = f.otherWorkspace
	if err := f.repo.CheckSession(ctx, other, session); err != nil {
		t.Fatal(err)
	}
	if err := f.repo.CheckSession(ctx, member, session); err != nil {
		t.Fatal("tracking second request", err)
	}
	list, err := f.repo.Sessions(ctx, f.scope, nil, false)
	if err != nil || len(list.Items) != 1 {
		t.Fatalf("sessions %#v %v", list, err)
	}
	if err := f.repo.RevokeSession(ctx, f.scope, session.ID, "Lost device"); err != nil {
		t.Fatal(err)
	}
	if err := f.repo.CheckSession(ctx, member, session); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("revoked access", err)
	}
	if err := f.repo.CheckSession(ctx, other, session); err != nil {
		t.Fatal("unrelated tenant revoked", err)
	}
	if err := f.repo.RevokeMember(ctx, f.scope, f.member, "Offboarding"); err != nil {
		t.Fatal(err)
	}
	if err := f.repo.CheckSession(ctx, member, domain.SessionIdentity{}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("legacy bypass", err)
	}
	fresh := identity()
	fresh.AuthenticatedAt = time.Now().UTC().Add(time.Second)
	if err := f.repo.CheckSession(ctx, member, fresh); err != nil {
		t.Fatal("reauthentication", err)
	}
	if _, err := f.repo.Policy(ctx, member); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("member admin access", err)
	}
	inactive := f.scope
	inactive.ActorID = f.inactive
	if _, err := f.repo.Policy(ctx, inactive); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("inactive admin", err)
	}
	policy, err := f.repo.UpdatePolicy(ctx, f.scope, domain.PolicyUpdate{AllowedDomains: []string{"EXAMPLE.COM"}, AllowGuests: false, MaxSessionAgeHours: 24})
	if err != nil || policy.Version != 1 {
		t.Fatalf("policy %#v %v", policy, err)
	}
	if _, err := f.repo.UpdatePolicy(ctx, f.scope, domain.PolicyUpdate{AllowGuests: true}); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("stale policy", err)
	}
	if _, err := f.repo.UpdatePolicy(ctx, f.scope, domain.PolicyUpdate{AllowedDomains: []string{"foreign.test"}, ExpectedVersion: 1}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatal("admin lockout", err)
	}
	guest := f.scope
	guest.ActorID = f.guest
	if err := f.repo.CheckSession(ctx, guest, identity()); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("guest policy", err)
	}
	old := identity()
	old.AuthenticatedAt = time.Now().Add(-25 * time.Hour)
	old.ExpiresAt = time.Now().Add(30 * 24 * time.Hour)
	if err := f.repo.CheckSession(ctx, f.scope, old); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("renewed session age bypass", err)
	}
	exec(t, f.pool, `UPDATE users SET email='admin@foreign.test' WHERE user_id=$1`, f.scope.ActorID)
	if err := f.repo.CheckSession(ctx, f.scope, identity()); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("live email policy", err)
	}
}
func TestAuditTenantFiltersKeysetsAndImmutability(t *testing.T) {
	f := seed(t)
	ctx := t.Context()
	if err := f.repo.RecordExport(ctx, f.scope, 42); err != nil {
		t.Fatal(err)
	}
	if err := f.repo.RecordAuditExport(ctx, f.scope, 1); err != nil {
		t.Fatal(err)
	}
	foreign := f.scope
	foreign.WorkspaceID = f.otherWorkspace
	if err := f.repo.RecordExport(ctx, foreign, 99); err != nil {
		t.Fatal(err)
	}
	exec(t, f.pool, `INSERT INTO audit_events(workspace_id,actor_type,actor_id,entity_type,entity_id,event_type,metadata) VALUES ($1,'human_user',$2,'workspace',$1,'test.event',CAST($3 AS jsonb))`, f.scope.WorkspaceID, f.scope.ActorID, `{"reason":"Scheduled","secret":"DO_NOT_EXPORT"}`)
	filter := domain.AuditFilter{Limit: 2}
	page, err := f.repo.Audit(ctx, f.scope, filter)
	if err != nil || len(page) != 2 {
		t.Fatalf("first page %#v %v", page, err)
	}
	last := page[1]
	filter.Before = &domain.Position{CreatedAt: last.CreatedAt, ID: last.ID, Source: last.Source}
	next, err := f.repo.Audit(ctx, f.scope, filter)
	if err != nil || len(next) != 1 {
		t.Fatalf("next page %#v %v", next, err)
	}
	for _, item := range append(page, next...) {
		if strings.Contains(string(item.Metadata), "DO_NOT_EXPORT") || strings.Contains(string(item.Metadata), "99") {
			t.Fatal("metadata or tenant leaked")
		}
	}
	filter = domain.AuditFilter{Limit: 100, ActorID: &f.scope.ActorID, ResourceType: "workspace", ResourceID: &f.scope.WorkspaceID}
	events, err := f.repo.Audit(ctx, f.scope, filter)
	if err != nil || len(events) != 3 {
		t.Fatalf("filtered %#v %v", events, err)
	}
	filter.ResourceID = &f.otherWorkspace
	events, err = f.repo.Audit(ctx, f.scope, filter)
	if err != nil || len(events) != 0 {
		t.Fatal("foreign resource leaked", err)
	}
	member := f.scope
	member.ActorID = f.member
	if _, err := f.repo.Audit(ctx, member, filter); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("member audit access", err)
	}
	if _, err := f.pool.Exec(ctx, `UPDATE workspace_security_audit_events SET operation='tampered' WHERE workspace_id=$1`, f.scope.WorkspaceID); err == nil {
		t.Fatal("audit mutable")
	}
	exec(t, f.pool, `UPDATE workspaces SET deleted_at=CURRENT_TIMESTAMP WHERE workspace_id=$1`, f.scope.WorkspaceID)
	if err := f.repo.CheckSession(ctx, f.scope, identity()); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("deleted tenant accessible", err)
	}
}
