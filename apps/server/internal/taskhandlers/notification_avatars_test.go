package taskhandlers

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type avatarHandleStoreStub struct {
	id    uuid.UUID
	calls int
}

func (s *avatarHandleStoreStub) EnsureEmailAvatarHandle(context.Context, uuid.UUID) (uuid.UUID, error) {
	s.calls++
	return s.id, nil
}

func TestNotificationPortraitsUseDurableURLsAndResolveEachActorOnce(t *testing.T) {
	store := &avatarHandleStoreStub{id: uuid.New()}
	h := &handlers{emailAvatars: store, apiPublicURL: "https://api.fortyone.app", log: logger.NewWithText(io.Discard, slog.LevelError, "test")}
	actor := mailer.EmailActor{ID: uuid.New(), Name: "Sam Taylor", AvatarURL: "profiles/private.png"}
	digest := mailer.Digest{Rows: []mailer.DigestRow{{Actor: actor}, {Actor: actor}, {Actor: mailer.EmailActor{Name: "No photo"}}}}
	h.resolveDigestAvatars(t.Context(), &digest)
	want := "https://api.fortyone.app/media/email-avatars/" + store.id.String() + "/avatar"
	require.Equal(t, want, digest.Rows[0].Actor.AvatarURL)
	require.Equal(t, want, digest.Rows[1].Actor.AvatarURL)
	require.Empty(t, digest.Rows[2].Actor.AvatarURL)
	require.Equal(t, 1, store.calls)
	require.NotContains(t, want, "X-Amz")
	for _, invalid := range []string{"", "http://api.fortyone.app", "https://name:secret@api.fortyone.app", "https://api.fortyone.app?token=secret"} {
		_, err := emailAvatarURL(invalid, store.id)
		require.Error(t, err)
	}
}
