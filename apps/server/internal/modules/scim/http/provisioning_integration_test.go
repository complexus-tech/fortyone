//go:build integration

package scimhttp_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/scim/domain"
	scimhttp "github.com/complexus-tech/projects-api/internal/modules/scim/http"
	repository "github.com/complexus-tech/projects-api/internal/modules/scim/repository"
	service "github.com/complexus-tech/projects-api/internal/modules/scim/service"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type seats struct {
	calls int
	err   error
}

func (s *seats) UpdateSubscriptionSeats(context.Context, uuid.UUID) error { s.calls++; return s.err }

type fixture struct {
	pool                               *pgxpool.Pool
	repo                               *repository.Repository
	service                            *service.Service
	admin                              domain.AdminScope
	scope                              domain.Scope
	token, slug                        string
	seats                              *seats
	foreign, member, team, foreignTeam uuid.UUID
}

func seed(t *testing.T) fixture {
	t.Helper()
	database := testkit.NewPostgres(t)
	f := fixture{pool: database.Pool, repo: repository.New(database.Pool), admin: domain.AdminScope{ActorID: uuid.New(), WorkspaceID: uuid.New()}, slug: uuid.NewString(), seats: &seats{}, foreign: uuid.New(), member: uuid.New(), team: uuid.New(), foreignTeam: uuid.New()}
	exec(t, f.pool, `INSERT INTO workspaces(workspace_id,name,slug) VALUES($1,'SCIM',$2),($3,'Other',$4)`, f.admin.WorkspaceID, f.slug, f.foreign, uuid.NewString())
	for _, account := range []struct {
		id   uuid.UUID
		role string
	}{{f.admin.ActorID, "admin"}, {f.member, "member"}} {
		exec(t, f.pool, `INSERT INTO users(user_id,username,email,is_active) VALUES($1,$2,$3,TRUE)`, account.id, account.id.String(), account.id.String()+"@example.com")
		exec(t, f.pool, `INSERT INTO workspace_members(workspace_id,user_id,role) VALUES($1,$2,CAST($3 AS user_role))`, f.admin.WorkspaceID, account.id, account.role)
	}
	exec(t, f.pool, `INSERT INTO workspace_members(workspace_id,user_id,role) VALUES($1,$2,'member')`, f.foreign, f.member)
	exec(t, f.pool, `INSERT INTO teams(team_id,workspace_id,name,code,color) VALUES($1,$2,'One','ONE','#123456'),($3,$4,'Other','TWO','#654321')`, f.team, f.admin.WorkspaceID, f.foreignTeam, f.foreign)
	exec(t, f.pool, `INSERT INTO team_members(team_id,user_id) VALUES($1,$2),($3,$2)`, f.team, f.member, f.foreignTeam)
	var err error
	f.service, err = service.New(f.repo, f.seats, "integration-scim-credential-secret")
	if err != nil {
		t.Fatal(err)
	}
	minted, err := f.service.Mint(t.Context(), f.admin, domain.MintInput{Name: "Test provider", LifetimeDays: 90})
	if err != nil {
		t.Fatal(err)
	}
	f.token = minted.Token
	f.scope, err = f.service.Authenticate(t.Context(), f.slug, f.token)
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
func scalar[T any](t *testing.T, pool *pgxpool.Pool, query string, args ...any) T {
	t.Helper()
	var value T
	if err := pool.QueryRow(t.Context(), query, args...).Scan(&value); err != nil {
		t.Fatal(err)
	}
	return value
}
func input(f fixture) domain.Input {
	return domain.Input{Schemas: []string{domain.UserSchema}, UserName: f.member.String() + "@example.com", ExternalID: ptr("directory-member"), DisplayName: "Managed name", Emails: []domain.Email{{Value: f.member.String() + "@example.com", Primary: true}}}
}
func ptr[T any](v T) *T { return &v }
func TestTenantDeactivationPreservesGlobalAndForeignAccess(t *testing.T) {
	f := seed(t)
	ctx := t.Context()
	originalVersion := scalar[int64](t, f.pool, `SELECT auth_session_version FROM users WHERE user_id=$1`, f.member)
	user, err := f.service.Create(ctx, f.scope, input(f))
	if err != nil {
		t.Fatal(err)
	}
	if user.UserID != f.member {
		t.Fatal("created a duplicate global account")
	}
	f.seats.err = errors.New("provider-private-error")
	changed, err := f.service.Patch(ctx, f.scope, user.ID, domain.Patch{Schemas: []string{domain.PatchSchema}, Operations: []domain.Operation{{Op: "replace", Path: "active", Value: json.RawMessage(`false`)}}})
	if err != nil || changed.Active {
		t.Fatal("deactivation", err)
	}
	if scalar[int](t, f.pool, `SELECT count(*) FROM workspace_members WHERE workspace_id=$1 AND user_id=$2`, f.admin.WorkspaceID, f.member) != 0 || scalar[int](t, f.pool, `SELECT count(*) FROM team_members WHERE team_id=$1 AND user_id=$2`, f.team, f.member) != 0 {
		t.Fatal("tenant access survived deactivation")
	}
	if scalar[int](t, f.pool, `SELECT count(*) FROM workspace_members WHERE workspace_id=$1 AND user_id=$2`, f.foreign, f.member) != 1 || scalar[int](t, f.pool, `SELECT count(*) FROM team_members WHERE team_id=$1 AND user_id=$2`, f.foreignTeam, f.member) != 1 {
		t.Fatal("foreign access changed")
	}
	if !scalar[bool](t, f.pool, `SELECT is_active FROM users WHERE user_id=$1`, f.member) || scalar[int64](t, f.pool, `SELECT auth_session_version FROM users WHERE user_id=$1`, f.member) != originalVersion {
		t.Fatal("global auth state changed")
	}
	if scalar[string](t, f.pool, `SELECT username FROM users WHERE user_id=$1`, f.member) != f.member.String() {
		t.Fatal("SCIM overwrote global profile")
	}
	status, err := f.service.Status(ctx, f.admin)
	if err != nil || !status.PendingSeatSync || strings.Contains(status.SeatSyncError, "private") {
		t.Fatal("seat failure was lost or exposed", err)
	}
	f.seats.err = nil
	if err = f.service.RetrySeats(ctx, f.admin); err != nil {
		t.Fatal(err)
	}
	status, err = f.service.Status(ctx, f.admin)
	if err != nil || status.PendingSeatSync {
		t.Fatal("retry didn't clear completed generation", err)
	}
	changed, err = f.service.Patch(ctx, f.scope, user.ID, domain.Patch{Schemas: []string{domain.PatchSchema}, Operations: []domain.Operation{{Op: "replace", Path: "active", Value: json.RawMessage(`true`)}}})
	if err != nil || !changed.Active {
		t.Fatal("reactivation", err)
	}
	if scalar[string](t, f.pool, `SELECT CAST(role AS text) FROM workspace_members WHERE workspace_id=$1 AND user_id=$2`, f.admin.WorkspaceID, f.member) != "member" {
		t.Fatal("reactivation changed member role")
	}
	if scalar[int](t, f.pool, `SELECT count(*) FROM team_members WHERE team_id=$1 AND user_id=$2`, f.team, f.member) != 0 {
		t.Fatal("reactivation resurrected private team membership")
	}
	if _, err = f.pool.Exec(ctx, `UPDATE workspace_scim_audit_events SET operation='tamper'`); err == nil {
		t.Fatal("audit was mutable")
	}
}
func TestTokenAuthorityAndAtomicPatch(t *testing.T) {
	f := seed(t)
	ctx := t.Context()
	user, err := f.service.Create(ctx, f.scope, input(f))
	if err != nil {
		t.Fatal(err)
	}
	patch := domain.Patch{Schemas: []string{domain.PatchSchema}, Operations: []domain.Operation{{Op: "replace", Path: "active", Value: json.RawMessage(`false`)}, {Op: "replace", Path: "password", Value: json.RawMessage(`"secret"`)}}}
	if _, err = f.service.Patch(ctx, f.scope, user.ID, patch); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatal("unsupported patch accepted", err)
	}
	if scalar[int](t, f.pool, `SELECT count(*) FROM workspace_members WHERE workspace_id=$1 AND user_id=$2`, f.admin.WorkspaceID, f.member) != 1 {
		t.Fatal("partial patch revoked membership")
	}
	if _, err = f.repo.Update(ctx, f.scope, user.ID, user.Version-1, domain.Mutation{UserName: user.UserName, Profile: user.Profile, Active: false}, false); !errors.Is(err, domain.ErrChanged) {
		t.Fatal("stale version mutated resource", err)
	}
	adminInput := domain.Input{Schemas: []string{domain.UserSchema}, UserName: f.admin.ActorID.String() + "@example.com"}
	adminUser, err := f.service.Create(ctx, f.scope, adminInput)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.service.Delete(ctx, f.scope, adminUser.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("issuer or last admin removed", err)
	}
	other := f.scope
	other.WorkspaceID = f.foreign
	if _, err = f.service.Get(ctx, other, user.ID); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatal("cross tenant token accepted", err)
	}
	exec(t, f.pool, `UPDATE workspace_members SET role='member' WHERE workspace_id=$1 AND user_id=$2`, f.admin.WorkspaceID, f.admin.ActorID)
	if _, err = f.service.Authenticate(ctx, f.slug, f.token); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatal("demoted token issuer authenticated", err)
	}
	if _, err = f.service.Get(ctx, f.scope, user.ID); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatal("previously resolved scope bypassed current authority", err)
	}
	exec(t, f.pool, `UPDATE workspace_members SET role='admin' WHERE workspace_id=$1 AND user_id=$2`, f.admin.WorkspaceID, f.admin.ActorID)
	if err = f.service.Revoke(ctx, f.admin, f.scope.CredentialID); err != nil {
		t.Fatal(err)
	}
	if _, err = f.service.Authenticate(ctx, f.slug, f.token); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatal("revoked token authenticated", err)
	}
}
func TestSCIMProtocolLifecycleDiscoveryAndFiltering(t *testing.T) {
	f := seed(t)
	h := scimhttp.New(f.service, "https://api.example.com")
	send := func(method, path, body string, id uuid.UUID) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, "https://api.example.com/scim/v2/"+f.slug+path, strings.NewReader(body))
		r.SetPathValue("workspaceSlug", f.slug)
		if id != uuid.Nil {
			r.SetPathValue("resourceId", id.String())
		}
		r.Header.Set("Authorization", "Bearer "+f.token)
		r.Header.Set("Content-Type", "application/scim+json")
		w := httptest.NewRecorder()
		var err error
		if strings.HasPrefix(path, "/Users") {
			if id == uuid.Nil {
				err = h.Users(t.Context(), w, r)
			} else {
				err = h.User(t.Context(), w, r)
			}
		} else {
			err = h.Discovery(t.Context(), w, r)
		}
		if err != nil {
			t.Fatal(err)
		}
		if w.Header().Get("Content-Type") != "application/scim+json" {
			t.Fatal("wrong media type")
		}
		return w
	}
	created := send("POST", "/Users", `{"schemas":["`+domain.UserSchema+`"],"userName":"new@example.com","externalId":"directory-42","name":{"givenName":"New"},"emails":[{"value":"new@example.com","primary":true}]}`, uuid.Nil)
	if created.Code != 201 {
		t.Fatal(created.Code, created.Body.String())
	}
	var resource struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &resource); err != nil {
		t.Fatal(err)
	}
	id := uuid.MustParse(resource.ID)
	if !strings.HasSuffix(created.Header().Get("Location"), id.String()) || strings.Contains(created.Body.String(), `"data"`) {
		t.Fatal("non-SCIM create response")
	}
	list := send("GET", `/Users?filter=userName%20eq%20%22new%40example.com%22&count=0`, "", uuid.Nil)
	if list.Code != 200 || !strings.Contains(list.Body.String(), `"totalResults":1`) || !strings.Contains(list.Body.String(), `"Resources":[]`) {
		t.Fatal("count zero or exact filter", list.Body.String())
	}
	bad := send("GET", `/Users?filter=userName%20co%20%22new%22`, "", uuid.Nil)
	if bad.Code != 400 || !strings.Contains(bad.Body.String(), `"scimType":"invalidFilter"`) {
		t.Fatal("invalid filter semantics", bad.Body.String())
	}
	replaced := send("PUT", "/Users/"+id.String(), `{"schemas":["`+domain.UserSchema+`"],"id":"readonly-ignored","meta":{"resourceType":"User"},"userName":"renamed@example.com","active":true}`, id)
	if replaced.Code != 200 {
		t.Fatal("PUT", replaced.Body.String())
	}
	patched := send("PATCH", "/Users/"+id.String(), `{"schemas":["`+domain.PatchSchema+`"],"Operations":[{"op":"add","path":"displayName","value":"Visible"},{"op":"remove","path":"displayName"}]}`, id)
	if patched.Code != 200 || strings.Contains(patched.Body.String(), `"displayName"`) {
		t.Fatal("atomic attribute removal", patched.Body.String())
	}
	for _, path := range []string{"/ServiceProviderConfig", "/ResourceTypes", "/Schemas"} {
		response := send(http.MethodGet, path, "", uuid.Nil)
		if response.Code != 200 {
			t.Fatal(path, response.Body.String())
		}
	}
	deleted := send("DELETE", "/Users/"+id.String(), "", id)
	if deleted.Code != 204 || deleted.Body.Len() != 0 {
		t.Fatal("DELETE response", deleted.Body.String())
	}
	gone := send("GET", "/Users/"+id.String(), "", id)
	if gone.Code != 404 {
		t.Fatal("deleted resource remained visible")
	}
	if scalar[int](t, f.pool, `SELECT count(*) FROM internal_slack_alerts WHERE user_id=(SELECT user_id FROM workspace_scim_users WHERE id=$1) AND kind='account_created'`, id) != 1 {
		t.Fatal("new account omitted canonical alert")
	}
	if scalar[int](t, f.pool, `SELECT count(*) FROM workspace_scim_audit_events WHERE operation='workspace.scim_user_deleted'`) < 1 {
		t.Fatal("delete audit missing")
	}
}
func TestCredentialsExpireAndInactiveAccountsCannotProvision(t *testing.T) {
	f := seed(t)
	ctx := t.Context()
	exec(t, f.pool, `UPDATE users SET is_active=FALSE WHERE user_id=$1`, f.member)
	if _, err := f.service.Create(ctx, f.scope, input(f)); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("inactive account was provisioned", err)
	}
	exec(t, f.pool, `UPDATE workspace_scim_credentials SET created_at=CAST($2 AS timestamptz) - interval '1 day',expires_at=CAST($2 AS timestamptz) WHERE id=$1`, f.scope.CredentialID, time.Now().Add(-time.Hour))
	if _, err := f.service.Authenticate(ctx, f.slug, f.token); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatal("expired token authenticated", err)
	}
}
