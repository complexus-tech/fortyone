//go:build integration

package usersrepository

import (
	"sync"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestEmailAvatarHandleIsStableConcurrentAndFollowsCurrentProfile(t *testing.T) {
	db := testkit.NewPostgres(t)
	ctx := t.Context()
	userID := insertUserTestAccount(t, ctx, db.Pool, "email-avatar", true, false)
	_, err := db.Pool.Exec(ctx, "UPDATE users SET avatar_url=$1 WHERE user_id=$2", "profiles/original.png", userID)
	require.NoError(t, err)
	repository := New(db.Pool)
	var group sync.WaitGroup
	handles := make(chan uuid.UUID, 8)
	errors := make(chan error, 8)
	for range 8 {
		group.Add(1)
		go func() {
			defer group.Done()
			handle, err := repository.EnsureEmailAvatarHandle(ctx, userID)
			handles <- handle
			errors <- err
		}()
	}
	group.Wait()
	close(handles)
	close(errors)
	for err := range errors {
		require.NoError(t, err)
	}
	var stable uuid.UUID
	for handle := range handles {
		require.NotEqual(t, uuid.Nil, handle)
		if stable == uuid.Nil {
			stable = handle
		}
		require.Equal(t, stable, handle)
	}
	_, err = db.Pool.Exec(ctx, "UPDATE email_avatar_handles SET created_at=$1 WHERE handle=$2", time.Now().Add(-365*24*time.Hour), stable)
	require.NoError(t, err)
	file, err := repository.GetEmailAvatar(ctx, stable)
	require.NoError(t, err)
	require.Equal(t, "profiles/original.png", file)
	_, err = db.Pool.Exec(ctx, "UPDATE users SET avatar_url=$1 WHERE user_id=$2", "profiles/replaced.jpg", userID)
	require.NoError(t, err)
	file, err = repository.GetEmailAvatar(ctx, stable)
	require.NoError(t, err)
	require.Equal(t, "profiles/replaced.jpg", file)
	_, err = db.Pool.Exec(ctx, "UPDATE users SET avatar_url=NULL WHERE user_id=$1", userID)
	require.NoError(t, err)
	_, err = repository.GetEmailAvatar(ctx, stable)
	require.ErrorIs(t, err, ErrNotFound)
	_, err = repository.GetEmailAvatar(ctx, uuid.New())
	require.ErrorIs(t, err, ErrNotFound)
}
