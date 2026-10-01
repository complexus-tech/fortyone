//go:build integration

package enterprisessorepository_test

import (
	"errors"
	"testing"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/enterprisesso/domain"
	repository "github.com/complexus-tech/projects-api/internal/modules/enterprisesso/repository"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type fixture struct {
	repo                     *repository.Repository
	pool                     *pgxpool.Pool
	scope                    domain.Scope
	member, foreignWorkspace uuid.UUID
	connection               domain.Connection
}

func seed(t *testing.T) fixture {
	t.Helper()
	database := testkit.NewPostgres(t)
	f := fixture{repo: repository.New(database.Pool), pool: database.Pool, scope: domain.Scope{ActorID: uuid.New(), WorkspaceID: uuid.New()}, member: uuid.New(), foreignWorkspace: uuid.New()}
	for _, workspace := range []uuid.UUID{f.scope.WorkspaceID, f.foreignWorkspace} {
		exec(t, f.pool, `INSERT INTO workspaces(workspace_id,name,slug) VALUES ($1,'SSO',$2)`, workspace, uuid.NewString())
	}
	for _, account := range []struct {
		id   uuid.UUID
		role string
	}{{f.scope.ActorID, "admin"}, {f.member, "member"}} {
		exec(t, f.pool, `INSERT INTO users(user_id,username,email,is_active) VALUES ($1,$2,$3,TRUE)`, account.id, account.id.String(), account.id.String()+"@example.com")
		exec(t, f.pool, `INSERT INTO workspace_members(workspace_id,user_id,role) VALUES ($1,$2,CAST($3 AS user_role))`, f.scope.WorkspaceID, account.id, account.role)
	}
	var err error
	f.connection, err = f.repo.Create(t.Context(), f.scope, domain.Connection{ID: uuid.New(), Issuer: "https://issuer.example.com", ClientID: "client", SecretEnvelope: "encrypted-fixture"})
	if err != nil {
		t.Fatal(err)
	}
	return f
}
func exec(t *testing.T, pool *pgxpool.Pool, query string, args ...any) {
	t.Helper()
	if _, err := pool.Exec(t.Context(), query, args...); err != nil {
		t.Fatal(err)
	}
}
func TestSSORequiresAccountProofThenUsesStableSubject(t *testing.T) {
	f := seed(t)
	ctx := t.Context()
	identity := domain.Identity{Subject: "idp-subject", Email: f.member.String() + "@example.com", AuthenticatedAt: time.Now().UTC()}
	attempt := domain.Attempt{WorkspaceID: f.scope.WorkspaceID, ConnectionID: f.connection.ID, Generation: 1}
	if _, err := f.repo.Authenticate(ctx, attempt, identity); !errors.Is(err, domain.ErrLinkRequired) {
		t.Fatal("email claim took over existing account", err)
	}
	attempt.LinkUserID = f.scope.ActorID
	if _, err := f.repo.Authenticate(ctx, attempt, identity); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("linked another account's email", err)
	}
	attempt.LinkUserID = f.member
	userID, err := f.repo.Authenticate(ctx, attempt, identity)
	if err != nil || userID != f.member {
		t.Fatal("verified initial link", err)
	}
	attempt.LinkUserID = uuid.Nil
	identity.Email = "changed-email@example.com"
	userID, err = f.repo.Authenticate(ctx, attempt, identity)
	if err != nil || userID != f.member {
		t.Fatal("sign-in did not resolve stable subject", err)
	}
	attempt.LinkUserID = f.scope.ActorID
	if _, err := f.repo.Authenticate(ctx, attempt, identity); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("subject rebound to different authenticated user", err)
	}
	attempt.LinkUserID = uuid.Nil
	attempt.WorkspaceID = f.foreignWorkspace
	if _, err := f.repo.Authenticate(ctx, attempt, identity); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("cross-tenant subject authentication", err)
	}
	exec(t, f.pool, `DELETE FROM workspace_members WHERE workspace_id=$1 AND user_id=$2`, f.scope.WorkspaceID, f.member)
	attempt.WorkspaceID = f.scope.WorkspaceID
	if _, err := f.repo.Authenticate(ctx, attempt, identity); !errors.Is(err, domain.ErrLinkRequired) {
		t.Fatal("removed member reused SSO", err)
	}
}
func TestSSORequiredGenerationLifecycleAndImmutableAudit(t *testing.T) {
	f := seed(t)
	ctx := t.Context()
	c := f.connection
	c.RequireSSO = true
	c.Enabled = true
	updated, err := f.repo.Update(ctx, f.scope, c, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.repo.CheckSession(ctx, f.scope, domain.SessionProof{}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("social session bypassed required SSO", err)
	}
	proof := domain.SessionProof{ConnectionID: c.ID, Generation: 1, AuthenticatedAt: time.Now()}
	if err := f.repo.CheckSession(ctx, f.scope, proof); err != nil {
		t.Fatal(err)
	}
	foreign := f.scope
	foreign.WorkspaceID = f.foreignWorkspace
	if err := f.repo.CheckSession(ctx, foreign, domain.SessionProof{}); err != nil {
		t.Fatal("SSO enforcement leaked to other tenant", err)
	}
	updated.Generation = 2
	updated.SecretEnvelope = "new-encrypted-fixture"
	updated, err = f.repo.Update(ctx, f.scope, updated, updated.Version)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.repo.CheckSession(ctx, f.scope, proof); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("old credential generation bypassed enforcement", err)
	}
	if _, err := f.repo.Update(ctx, f.scope, c, 1); !errors.Is(err, domain.ErrConflict) {
		t.Fatal("stale connection update", err)
	}
	member := f.scope
	member.ActorID = f.member
	if _, err := f.repo.Get(ctx, member); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("member read provider credential config", err)
	}
	if _, err := f.pool.Exec(ctx, `DELETE FROM workspace_sso_audit_events WHERE workspace_id=$1`, f.scope.WorkspaceID); err == nil {
		t.Fatal("SSO audit mutable")
	}
	if err := f.repo.Archive(ctx, f.scope, "Replace identity provider"); err != nil {
		t.Fatal(err)
	}
	if err := f.repo.CheckSession(ctx, f.scope, domain.SessionProof{}); err != nil {
		t.Fatal("archived provider remained enforced", err)
	}
	if result, err := f.repo.Get(ctx, f.scope); err != nil || result != nil {
		t.Fatal("archived config returned", err)
	}
}
