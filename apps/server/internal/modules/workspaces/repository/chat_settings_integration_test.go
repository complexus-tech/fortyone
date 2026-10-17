//go:build integration

package workspacesrepository

import (
	"context"
	"fmt"
	workspacedomain "github.com/complexus-tech/projects-api/internal/modules/workspaces/domain"
	"sync"
	"testing"
	"time"

	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestChatSettingsDefaultOnReadsDoNotWriteAndToggleLive(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	workspace := insertWorkspaceTestWorkspace(t, ctx, postgres.Pool, "chat-settings")
	user := insertWorkspaceTestUser(t, ctx, postgres.Pool, "chat-settings", true)
	insertWorkspaceTestMembership(t, ctx, postgres.Pool, workspace, user, "admin")
	repository := New(postgres.Pool)
	var slug string
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT slug FROM workspaces WHERE workspace_id=$1`, workspace).Scan(&slug))
	for range 5 {
		enabled, err := repository.ChatEnabled(ctx, workspace)
		require.NoError(t, err)
		require.True(t, enabled, "legacy workspace without settings defaults on")
	}
	var count int
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT count(*) FROM workspace_settings WHERE workspace_id=$1`, workspace).Scan(&count))
	require.Zero(t, count, "authorization reads never initialize settings")
	settings, err := repository.GetOrCreateWorkspaceSettings(ctx, workspace)
	require.NoError(t, err)
	require.True(t, settings.ChatEnabled, "new settings use the database default")
	var before, after string
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT CAST(xmin AS text) FROM workspace_settings WHERE workspace_id=$1`, workspace).Scan(&before))
	for range 5 {
		enabled, err := repository.ChatEnabled(ctx, workspace)
		require.NoError(t, err)
		require.True(t, enabled)
		membership, err := repository.ResolveCurrentMembership(ctx, slug, user)
		require.NoError(t, err)
		require.True(t, membership.ChatEnabled)
	}
	require.NoError(t, postgres.Pool.QueryRow(ctx, `SELECT CAST(xmin AS text) FROM workspace_settings WHERE workspace_id=$1`, workspace).Scan(&after))
	require.Equal(t, before, after, "live checks must not rewrite shared settings")
	for _, enabled := range []bool{false, true} {
		settings.ChatEnabled = enabled
		saved, err := repository.UpdateWorkspaceSettings(ctx, workspace, settings)
		require.NoError(t, err)
		require.Equal(t, enabled, saved.ChatEnabled)
		settings = saved
		current, err := repository.ChatEnabled(ctx, workspace)
		require.NoError(t, err)
		require.Equal(t, enabled, current)
		membership, err := repository.ResolveCurrentMembership(ctx, slug, user)
		require.NoError(t, err)
		require.Equal(t, enabled, membership.ChatEnabled, "no process restart or auth cache")
	}
	enabled, err := repository.ChatEnabled(ctx, uuid.New())
	require.Error(t, err)
	require.False(t, enabled, "unknown workspace cannot inherit default-on access")
	_, err = postgres.Pool.Exec(ctx, `UPDATE workspaces SET deleted_at=NOW() WHERE workspace_id=$1`, workspace)
	require.NoError(t, err)
	enabled, err = repository.ChatEnabled(ctx, workspace)
	require.Error(t, err)
	require.False(t, enabled)
}

func TestChatSettingChangeSerializesWithAuthorizedUnit(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	workspace := insertWorkspaceTestWorkspace(t, ctx, postgres.Pool, "chat-setting-lock")
	repository := New(postgres.Pool)
	settings, err := repository.GetOrCreateWorkspaceSettings(ctx, workspace)
	require.NoError(t, err)
	ready, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unlock := func() { once.Do(func() { close(release) }) }
	defer unlock()
	done := make(chan error, 1)
	go func() {
		done <- platformdatabase.WithinUnitOfWork(ctx, postgres.Pool, func(unit context.Context, _ pgx.Tx) error {
			enabled, err := repository.ChatEnabled(unit, workspace)
			close(ready)
			if err != nil {
				return err
			}
			if !enabled {
				return fmt.Errorf("expected initial chat access")
			}
			<-release
			return nil
		})
	}()
	select {
	case <-ready:
	case err := <-done:
		t.Fatalf("authorization unit failed before readiness: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("authorization unit did not become ready")
	}
	settings.ChatEnabled = false
	bounded, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	_, err = repository.UpdateWorkspaceSettings(bounded, workspace, settings)
	cancel()
	require.Error(t, err, "feature toggle must wait for the authorized unit to finish")
	unlock()
	require.NoError(t, <-done)
	_, err = repository.UpdateWorkspaceSettings(ctx, workspace, settings)
	require.NoError(t, err)
	enabled, err := repository.ChatEnabled(ctx, workspace)
	require.NoError(t, err)
	require.False(t, enabled)
}

func TestChatSettingsRejectStaleUnrelatedSaveAfterDisable(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	workspace := insertWorkspaceTestWorkspace(t, ctx, postgres.Pool, "stale-chat-settings")
	repository := New(postgres.Pool)
	initial, err := repository.GetOrCreateWorkspaceSettings(ctx, workspace)
	require.NoError(t, err)
	unrelated := initial
	unrelated.ObjectiveEnabled = !initial.ObjectiveEnabled
	disabled := initial
	disabled.ChatEnabled = false
	saved, err := repository.UpdateWorkspaceSettings(ctx, workspace, disabled)
	require.NoError(t, err)
	require.True(t, saved.UpdatedAt.After(initial.UpdatedAt))
	_, err = repository.UpdateWorkspaceSettings(ctx, workspace, unrelated)
	require.ErrorIs(t, err, workspacedomain.ErrSettingsConflict)
	current, err := repository.GetWorkspaceSettings(ctx, workspace)
	require.NoError(t, err)
	require.False(t, current.ChatEnabled, "a stale unrelated save must not re-enable Messages")
	require.Equal(t, initial.ObjectiveEnabled, current.ObjectiveEnabled, "conflicting writes are fully rolled back")
	require.Equal(t, saved.UpdatedAt, current.UpdatedAt)
	current.ObjectiveEnabled = unrelated.ObjectiveEnabled
	refreshed, err := repository.UpdateWorkspaceSettings(ctx, workspace, current)
	require.NoError(t, err)
	require.False(t, refreshed.ChatEnabled)
	require.Equal(t, unrelated.ObjectiveEnabled, refreshed.ObjectiveEnabled)
}
