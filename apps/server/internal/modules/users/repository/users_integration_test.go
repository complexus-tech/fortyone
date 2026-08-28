//go:build integration

package usersrepository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestUsersEnforceWorkspacePrivateDataAndActorSemantics(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	workspaceA := insertUserTestWorkspace(t, ctx, postgres.Pool, "a")
	workspaceB := insertUserTestWorkspace(t, ctx, postgres.Pool, "b")
	activeA := insertUserTestAccount(t, ctx, postgres.Pool, "active-a", true, false)
	activeB := insertUserTestAccount(t, ctx, postgres.Pool, "active-b", true, false)
	inactiveA := insertUserTestAccount(t, ctx, postgres.Pool, "inactive-a", false, false)
	systemA := insertUserTestAccount(t, ctx, postgres.Pool, "system-a", true, true)
	insertUserTestWorkspaceMember(t, ctx, postgres.Pool, workspaceA, activeA, "member")
	insertUserTestWorkspaceMember(t, ctx, postgres.Pool, workspaceA, inactiveA, "member")
	insertUserTestWorkspaceMember(t, ctx, postgres.Pool, workspaceA, systemA, "system")
	insertUserTestWorkspaceMember(t, ctx, postgres.Pool, workspaceB, activeB, "member")
	teamA := insertUserTestTeam(t, ctx, postgres.Pool, workspaceA, "Team A")
	teamB := insertUserTestTeam(t, ctx, postgres.Pool, workspaceB, "Team B")
	insertUserTestTeamMember(t, ctx, postgres.Pool, teamA, activeA)
	insertUserTestTeamMember(t, ctx, postgres.Pool, teamB, activeB)

	repository := New(postgres.Pool)
	workspaceUsers, err := repository.List(ctx, workspaceA, users.CoreListUsersFilter{})
	if err != nil {
		t.Fatalf("list workspace users: %v", err)
	}
	assertUserIDs(t, workspaceUsers, activeA)

	crossWorkspaceTeamUsers, err := repository.List(ctx, workspaceA, users.CoreListUsersFilter{TeamID: &teamB})
	if err != nil {
		t.Fatalf("list with cross-workspace team: %v", err)
	}
	if len(crossWorkspaceTeamUsers) != 0 {
		t.Fatalf("cross-workspace team users = %#v, want none", crossWorkspaceTeamUsers)
	}

	if _, err := repository.GetUser(ctx, inactiveA); !errors.Is(err, users.ErrNotFound) {
		t.Fatalf("inactive user get error = %v, want ErrNotFound", err)
	}
	inactive, err := repository.GetUserByEmailAnyStatus(ctx, userTestEmail("inactive-a", inactiveA))
	if err != nil {
		t.Fatalf("get inactive user by email: %v", err)
	}
	if inactive.IsActive {
		t.Fatalf("inactive user = %#v, want inactive", inactive)
	}

	actors, err := repository.GetUsersByIDs(ctx, []uuid.UUID{inactiveA, systemA})
	if err != nil {
		t.Fatalf("get system/inactive actors: %v", err)
	}
	if len(actors) != 2 || !containsSystemUser(actors) || !containsInactiveUser(actors) {
		t.Fatalf("system/inactive actors = %#v", actors)
	}

	memory, err := repository.AddUserMemory(ctx, users.NewUserMemoryItem{
		UserID: activeA, WorkspaceID: workspaceA, Content: "private-a",
	})
	if err != nil {
		t.Fatalf("add scoped memory: %v", err)
	}
	if _, err := repository.AddUserMemory(ctx, users.NewUserMemoryItem{
		UserID: activeA, WorkspaceID: workspaceB, Content: "cross-tenant",
	}); !errors.Is(err, users.ErrMemoryNotFound) {
		t.Fatalf("cross-workspace memory create error = %v, want ErrMemoryNotFound", err)
	}

	updatedContent := "must-not-cross"
	if err := repository.UpdateUserMemory(ctx, memory.ID, users.UserMemoryScope{
		UserID: activeA, WorkspaceID: workspaceB,
	}, users.UpdateUserMemoryItem{Content: &updatedContent}); !errors.Is(err, users.ErrMemoryNotFound) {
		t.Fatalf("cross-workspace memory update error = %v, want ErrMemoryNotFound", err)
	}
	if err := repository.DeleteUserMemory(ctx, memory.ID, users.UserMemoryScope{
		UserID: activeB, WorkspaceID: workspaceA,
	}); !errors.Is(err, users.ErrMemoryNotFound) {
		t.Fatalf("cross-user memory delete error = %v, want ErrMemoryNotFound", err)
	}

	memoriesA, err := repository.ListUserMemories(ctx, activeA, workspaceA)
	if err != nil {
		t.Fatalf("list owner memories: %v", err)
	}
	if len(memoriesA) != 1 || memoriesA[0].Content != "private-a" {
		t.Fatalf("owner memories = %#v", memoriesA)
	}
	memoriesB, err := repository.ListUserMemories(ctx, activeA, workspaceB)
	if err != nil {
		t.Fatalf("list cross-workspace memories: %v", err)
	}
	if len(memoriesB) != 0 {
		t.Fatalf("cross-workspace memories = %#v, want none", memoriesB)
	}

	preferences, err := repository.GetAutomationPreferences(ctx, activeA, workspaceA)
	if err != nil {
		t.Fatalf("get default automation preferences: %v", err)
	}
	if preferences.AutoAssignSelf || !preferences.AutoScheduling || !preferences.OpenStoryInDialog {
		t.Fatalf("default preferences = %#v", preferences)
	}
	if _, err := repository.GetAutomationPreferences(ctx, activeA, workspaceB); !errors.Is(err, users.ErrNotFound) {
		t.Fatalf("cross-workspace preferences error = %v, want ErrNotFound", err)
	}
	if _, err := repository.GetAutomationPreferences(ctx, inactiveA, workspaceA); !errors.Is(err, users.ErrNotFound) {
		t.Fatalf("inactive preferences error = %v, want ErrNotFound", err)
	}
	enabled := true
	disabled := false
	if err := repository.UpdateAutomationPreferences(ctx, activeA, workspaceA, users.CoreUpdateAutomationPreferences{
		AutoAssignSelf: &enabled,
		AutoScheduling: &disabled,
	}); err != nil {
		t.Fatalf("update automation preferences: %v", err)
	}
	preferences, err = repository.GetAutomationPreferences(ctx, activeA, workspaceA)
	if err != nil {
		t.Fatalf("get updated automation preferences: %v", err)
	}
	if !preferences.AutoAssignSelf || preferences.AutoScheduling {
		t.Fatalf("updated preferences = %#v", preferences)
	}
}

func TestUserMemoriesRejectRevokedOwnerMembership(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()

	workspaceID := insertUserTestWorkspace(t, ctx, postgres.Pool, "revoked-memory")
	userID := insertUserTestAccount(t, ctx, postgres.Pool, "revoked-memory-owner", true, false)
	insertUserTestWorkspaceMember(t, ctx, postgres.Pool, workspaceID, userID, "member")
	repository := New(postgres.Pool)
	memory, err := repository.AddUserMemory(ctx, users.NewUserMemoryItem{
		UserID: userID, WorkspaceID: workspaceID, Content: "private before revocation",
	})
	if err != nil {
		t.Fatalf("add memory before revocation: %v", err)
	}

	if _, err := postgres.Pool.Exec(ctx, `
		DELETE FROM public.workspace_members
		WHERE workspace_id = $1 AND user_id = $2
	`, workspaceID, userID); err != nil {
		t.Fatalf("revoke memory owner membership: %v", err)
	}

	memories, err := repository.ListUserMemories(ctx, userID, workspaceID)
	if err != nil {
		t.Fatalf("list memories after revocation: %v", err)
	}
	if len(memories) != 0 {
		t.Fatalf("memories after revocation = %#v, want none", memories)
	}
	updatedContent := "must not update after revocation"
	if err := repository.UpdateUserMemory(ctx, memory.ID, users.UserMemoryScope{
		UserID: userID, WorkspaceID: workspaceID,
	}, users.UpdateUserMemoryItem{Content: &updatedContent}); !errors.Is(err, users.ErrMemoryNotFound) {
		t.Fatalf("update after revocation error = %v, want ErrMemoryNotFound", err)
	}
	if err := repository.DeleteUserMemory(ctx, memory.ID, users.UserMemoryScope{
		UserID: userID, WorkspaceID: workspaceID,
	}); !errors.Is(err, users.ErrMemoryNotFound) {
		t.Fatalf("delete after revocation error = %v, want ErrMemoryNotFound", err)
	}
}

func TestUserAccountLifecycleAndCallerTransactionRollback(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	repository := New(postgres.Pool)
	workspaceID := insertUserTestWorkspace(t, ctx, postgres.Pool, "lifecycle")

	created, err := repository.Create(ctx, users.CoreUser{
		Username: "grace", Email: "grace@example.com", FullName: "Grace Hopper",
		Timezone: "Africa/Harare", LastLoginAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if _, err := repository.Create(ctx, users.CoreUser{
		Username: "duplicate", Email: created.Email, Timezone: "Africa/Harare",
	}); !errors.Is(err, users.ErrEmailTaken) {
		t.Fatalf("duplicate user error = %v, want ErrEmailTaken", err)
	}
	insertUserTestWorkspaceMember(t, ctx, postgres.Pool, workspaceID, created.ID, "member")

	fullName := "Rear Admiral Grace Hopper"
	start := 480
	end := 960
	updated, err := repository.UpdateUser(ctx, created.ID, users.CoreUpdateUser{
		FullName: &fullName,
		WorkSchedule: &users.CoreWorkScheduleOverride{
			WorkingDays:        []int{1, 2, 3, 4, 5},
			WorkingStartMinute: &start,
			WorkingEndMinute:   &end,
		},
	})
	if err != nil {
		t.Fatalf("update user: %v", err)
	}
	if updated.FullName != fullName || len(updated.WorkingDays) != 5 ||
		updated.WorkingStartMinute == nil || *updated.WorkingStartMinute != start {
		t.Fatalf("updated user = %#v", updated)
	}

	tx, err := postgres.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin caller transaction: %v", err)
	}
	workspaceTransaction, err := repository.BindWorkspaceTransaction(tx)
	if err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("bind workspace transaction: %v", err)
	}
	if err := workspaceTransaction.UpdateLastUsedWorkspace(ctx, created.ID, workspaceID); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("update workspace in caller transaction: %v", err)
	}
	var insideWorkspace *uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT last_used_workspace_id FROM users WHERE user_id = $1`, created.ID).Scan(&insideWorkspace); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("read transaction-local workspace: %v", err)
	}
	if insideWorkspace == nil || *insideWorkspace != workspaceID {
		_ = tx.Rollback(ctx)
		t.Fatalf("transaction-local workspace = %v, want %s", insideWorkspace, workspaceID)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("roll back caller transaction: %v", err)
	}
	assertLastUsedWorkspace(t, ctx, postgres.Pool, created.ID, nil)

	if err := repository.UpdateUserWorkspace(ctx, created.ID, workspaceID); err != nil {
		t.Fatalf("update committed workspace: %v", err)
	}
	assertLastUsedWorkspace(t, ctx, postgres.Pool, created.ID, &workspaceID)

	deactivatedAt := time.Date(2026, time.August, 28, 10, 0, 0, 0, time.UTC)
	if err := repository.DeleteUser(ctx, created.ID, deactivatedAt); err != nil {
		t.Fatalf("deactivate user: %v", err)
	}
	if _, err := repository.GetUser(ctx, created.ID); !errors.Is(err, users.ErrNotFound) {
		t.Fatalf("get deactivated user error = %v, want ErrNotFound", err)
	}
	signedInAt := deactivatedAt.Add(time.Hour)
	activated, err := repository.ReactivateUserForVerifiedSignIn(ctx, users.VerifiedSignInReactivation{
		UserID: created.ID, SignedInAt: signedInAt,
	})
	if err != nil {
		t.Fatalf("activate user: %v", err)
	}
	if !activated.IsActive {
		t.Fatalf("activated user = %#v", activated)
	}
	var policy string
	var sessionVersion int64
	var persistedLastLogin, persistedUpdatedAt time.Time
	if err := postgres.Pool.QueryRow(ctx, `
		SELECT login_reactivation_policy, auth_session_version, last_login_at, updated_at
		FROM users
		WHERE user_id = $1
	`, created.ID).Scan(&policy, &sessionVersion, &persistedLastLogin, &persistedUpdatedAt); err != nil {
		t.Fatalf("inspect verified reactivation: %v", err)
	}
	if policy != "verified_sign_in" || sessionVersion != 2 {
		t.Fatalf("reactivation policy/version = %q/%d, want verified_sign_in/2", policy, sessionVersion)
	}
	if !persistedLastLogin.Equal(signedInAt) || !persistedUpdatedAt.Equal(signedInAt) {
		t.Fatalf("reactivation times = %v/%v, want %v", persistedLastLogin, persistedUpdatedAt, signedInAt)
	}
}

func TestVerifiedSignInReactivationFailsClosedForAdministratorAndLegacyPolicies(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	repository := New(postgres.Pool)
	signedInAt := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)

	for _, policy := range []string{"admin_only", "legacy_admin_review"} {
		t.Run(policy, func(t *testing.T) {
			userID := insertUserTestAccount(t, ctx, postgres.Pool, "blocked-"+policy, false, false)
			if _, err := postgres.Pool.Exec(ctx, `
				UPDATE users
				SET login_reactivation_policy = $2,
				    inactivity_warning_sent_at = $3
				WHERE user_id = $1
			`, userID, policy, signedInAt.Add(-24*time.Hour)); err != nil {
				t.Fatalf("arrange blocked account: %v", err)
			}

			_, err := repository.ReactivateUserForVerifiedSignIn(ctx, users.VerifiedSignInReactivation{
				UserID: userID, SignedInAt: signedInAt,
			})
			if !errors.Is(err, users.ErrInvalidCredentials) {
				t.Fatalf("reactivation error = %v, want generic invalid credentials", err)
			}

			var active bool
			var persistedPolicy string
			var warningAt *time.Time
			if err := postgres.Pool.QueryRow(ctx, `
				SELECT is_active, login_reactivation_policy, inactivity_warning_sent_at
				FROM users
				WHERE user_id = $1
			`, userID).Scan(&active, &persistedPolicy, &warningAt); err != nil {
				t.Fatalf("inspect blocked account: %v", err)
			}
			if active || persistedPolicy != policy || warningAt == nil {
				t.Fatalf("blocked account state = active:%t policy:%q warning:%v", active, persistedPolicy, warningAt)
			}
		})
	}
}

func TestAutomatedInactivityDeactivationRemainsEligibleForVerifiedSignIn(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	repository := New(postgres.Pool)
	userID := insertUserTestAccount(t, ctx, postgres.Pool, "automated-inactivity", true, false)
	lastLoginAt := time.Date(2025, time.August, 1, 8, 0, 0, 0, time.UTC)
	warningAt := time.Date(2026, time.July, 1, 8, 0, 0, 0, time.UTC)
	deactivatedAt := time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC)
	if _, err := postgres.Pool.Exec(ctx, `
		UPDATE users
		SET last_login_at = $2,
		    inactivity_warning_sent_at = $3
		WHERE user_id = $1
	`, userID, lastLoginAt, warningAt); err != nil {
		t.Fatalf("arrange inactive account: %v", err)
	}

	rows, err := repository.DeactivateInactiveUsers(
		ctx,
		lastLoginAt.Add(24*time.Hour),
		warningAt.Add(24*time.Hour),
		deactivatedAt,
		1,
	)
	if err != nil {
		t.Fatalf("deactivate inactive account: %v", err)
	}
	if rows != 1 {
		t.Fatalf("deactivated rows = %d, want 1", rows)
	}

	var policy string
	var version int64
	if err := postgres.Pool.QueryRow(ctx, `
		SELECT login_reactivation_policy, auth_session_version
		FROM users
		WHERE user_id = $1
	`, userID).Scan(&policy, &version); err != nil {
		t.Fatalf("inspect automated deactivation: %v", err)
	}
	if policy != "verified_sign_in" || version != 2 {
		t.Fatalf("automated deactivation policy/version = %q/%d", policy, version)
	}

	signedInAt := deactivatedAt.Add(time.Hour)
	user, err := repository.ReactivateUserForVerifiedSignIn(ctx, users.VerifiedSignInReactivation{
		UserID: userID, SignedInAt: signedInAt,
	})
	if err != nil {
		t.Fatalf("reactivate automated inactivity account: %v", err)
	}
	if !user.IsActive || !user.LastLoginAt.Equal(signedInAt) {
		t.Fatalf("reactivated user = %#v", user)
	}
}

func TestExternalIdentityResolutionIsAtomicAndPreservesInactiveIdentity(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	repository := New(postgres.Pool)

	inactiveID := insertUserTestAccount(t, ctx, postgres.Pool, "identity-inactive", false, false)
	inactiveEmail := userTestEmail("identity-inactive", inactiveID)
	inactiveResult, err := repository.ResolveExternalIdentity(ctx, users.CoreExternalIdentityInput{
		Provider: "microsoft", Issuer: "https://login.microsoftonline.com/test/v2.0",
		Subject: "inactive-subject", Email: inactiveEmail, FullName: "Inactive Identity",
		Timezone: "Africa/Harare",
	})
	if err != nil {
		t.Fatalf("resolve inactive external identity: %v", err)
	}
	if inactiveResult.Created || inactiveResult.User.ID != inactiveID || inactiveResult.User.IsActive {
		t.Fatalf("inactive identity result = %#v", inactiveResult)
	}

	input := users.CoreExternalIdentityInput{
		Provider: "microsoft", Issuer: "https://login.microsoftonline.com/test/v2.0",
		Subject: "concurrent-subject", Email: "concurrent-identity@example.com",
		FullName: "Concurrent Identity", Timezone: "Africa/Harare",
	}
	const requests = 8
	start := make(chan struct{})
	results := make(chan users.CoreExternalIdentityResult, requests)
	errorsChannel := make(chan error, requests)
	var waitGroup sync.WaitGroup
	for range requests {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			result, resolveErr := repository.ResolveExternalIdentity(context.Background(), input)
			results <- result
			errorsChannel <- resolveErr
		}()
	}
	close(start)
	waitGroup.Wait()
	close(results)
	close(errorsChannel)

	for resolveErr := range errorsChannel {
		if resolveErr != nil {
			t.Fatalf("concurrent identity resolution: %v", resolveErr)
		}
	}
	var resolvedID uuid.UUID
	createdCount := 0
	for result := range results {
		if resolvedID == uuid.Nil {
			resolvedID = result.User.ID
		}
		if result.User.ID != resolvedID {
			t.Fatalf("resolved user IDs differ: %s and %s", resolvedID, result.User.ID)
		}
		if result.Created {
			createdCount++
		}
	}
	if createdCount != 1 {
		t.Fatalf("created identity users = %d, want 1", createdCount)
	}

	var userCount, identityCount int
	if err := postgres.Pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM users WHERE email = $1),
			(SELECT COUNT(*) FROM user_external_identities WHERE provider = $2 AND issuer = $3 AND subject = $4)
	`, input.Email, input.Provider, input.Issuer, input.Subject).Scan(&userCount, &identityCount); err != nil {
		t.Fatalf("inspect resolved identity: %v", err)
	}
	if userCount != 1 || identityCount != 1 {
		t.Fatalf("users/identities = %d/%d, want 1/1", userCount, identityCount)
	}
}

func insertUserTestWorkspace(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	label string,
) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug)
		VALUES ($1, $2, $3)
	`, id, "Users "+label, "users-"+label+"-"+uuid.NewString()); err != nil {
		t.Fatalf("insert workspace %s: %v", label, err)
	}
	return id
}

func insertUserTestAccount(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	label string,
	active bool,
	system bool,
) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email, full_name, is_active, is_system)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, label+"-"+id.String(), userTestEmail(label, id), "Users "+label, active, system); err != nil {
		t.Fatalf("insert user %s: %v", label, err)
	}
	return id
}

func userTestEmail(label string, userID uuid.UUID) string {
	return fmt.Sprintf("%s-%s@example.com", label, userID)
}

func insertUserTestWorkspaceMember(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	workspaceID uuid.UUID,
	userID uuid.UUID,
	role string,
) {
	t.Helper()
	if _, err := pool.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, CAST($3 AS user_role))
	`, workspaceID, userID, role); err != nil {
		t.Fatalf("insert workspace member: %v", err)
	}
}

func insertUserTestTeam(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	workspaceID uuid.UUID,
	name string,
) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO teams (team_id, name, workspace_id, code, color)
		VALUES ($1, $2, $3, $4, '#000000')
	`, id, name, workspaceID, "U"+uuid.NewString()[:7]); err != nil {
		t.Fatalf("insert team: %v", err)
	}
	return id
}

func insertUserTestTeamMember(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	teamID uuid.UUID,
	userID uuid.UUID,
) {
	t.Helper()
	if _, err := pool.Exec(ctx, `
		INSERT INTO team_members (team_id, user_id)
		VALUES ($1, $2)
	`, teamID, userID); err != nil {
		t.Fatalf("insert team member: %v", err)
	}
}

func assertUserIDs(t *testing.T, actual []users.CoreUser, expected ...uuid.UUID) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Fatalf("user count = %d, want %d: %#v", len(actual), len(expected), actual)
	}
	for _, expectedID := range expected {
		found := false
		for _, user := range actual {
			if user.ID == expectedID {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("users %#v do not contain %s", actual, expectedID)
		}
	}
}

func containsSystemUser(usersList []users.CoreUser) bool {
	for _, user := range usersList {
		if user.IsSystem {
			return true
		}
	}
	return false
}

func containsInactiveUser(usersList []users.CoreUser) bool {
	for _, user := range usersList {
		if !user.IsActive {
			return true
		}
	}
	return false
}

func assertLastUsedWorkspace(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	userID uuid.UUID,
	want *uuid.UUID,
) {
	t.Helper()
	var got *uuid.UUID
	if err := pool.QueryRow(ctx, `SELECT last_used_workspace_id FROM users WHERE user_id = $1`, userID).Scan(&got); err != nil {
		t.Fatalf("read last used workspace: %v", err)
	}
	if want == nil && got == nil {
		return
	}
	if want == nil || got == nil || *want != *got {
		t.Fatalf("last used workspace = %v, want %v", got, want)
	}
}
