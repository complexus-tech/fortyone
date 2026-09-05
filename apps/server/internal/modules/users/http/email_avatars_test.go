package usershttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type emailAvatarStoreStub struct {
	handle uuid.UUID
	file   string
	err    error
	calls  int
}

func (s *emailAvatarStoreStub) GetEmailAvatar(_ context.Context, handle uuid.UUID) (string, error) {
	s.calls++
	if handle != s.handle {
		return "", users.ErrNotFound
	}
	return s.file, s.err
}

type emailAvatarSignerStub struct {
	now    time.Time
	file   string
	expiry time.Duration
	calls  int
	err    error
}

func (s *emailAvatarSignerStub) ResolveProfileImageURL(_ context.Context, file string, expiry time.Duration) (string, error) {
	s.calls++
	s.file, s.expiry = file, expiry
	return fmt.Sprintf("https://storage.example.com/%s?signature=%d", file, s.now.Unix()), s.err
}

func TestEmailAvatarURLResolvesAgainAfterAWeekAndProfileReplacement(t *testing.T) {
	handle := uuid.New()
	store := &emailAvatarStoreStub{handle: handle, file: "profiles/original.png"}
	signer := &emailAvatarSignerStub{now: time.Date(2026, 9, 5, 9, 0, 0, 0, time.UTC)}
	h := emailAvatarHandler{users: store, images: signer}
	path := "/media/email-avatars/" + handle.String() + "/avatar"
	request := func() *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		r.SetPathValue("handle", handle.String())
		w := httptest.NewRecorder()
		require.NoError(t, h.redirect(t.Context(), w, r))
		require.Equal(t, http.StatusFound, w.Code)
		require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
		require.Equal(t, "no-referrer", w.Header().Get("Referrer-Policy"))
		return w
	}
	first := request().Header().Get("Location")
	signer.now = signer.now.Add(8 * 24 * time.Hour)
	second := request().Header().Get("Location")
	require.NotEqual(t, first, second, "same permanent endpoint must mint a fresh storage signature")
	require.Equal(t, 5*time.Minute, signer.expiry)
	store.file = "profiles/replacement.jpg"
	third := request().Header().Get("Location")
	require.Contains(t, third, "replacement.jpg")
	require.Equal(t, 3, store.calls)
	require.Equal(t, 3, signer.calls)
}

func TestEmailAvatarEndpointRejectsUnknownHandlesAndRemovedPhotos(t *testing.T) {
	for _, mode := range []string{"invalid", "unknown", "removed", "storage failure"} {
		t.Run(mode, func(t *testing.T) {
			handle := uuid.New()
			store := &emailAvatarStoreStub{handle: handle, file: "profiles/a.png"}
			signer := &emailAvatarSignerStub{now: time.Now()}
			r := httptest.NewRequest(http.MethodGet, "/media/email-avatars/test/avatar?key=private-document.pdf", nil)
			r.SetPathValue("handle", handle.String())
			switch mode {
			case "invalid":
				r.SetPathValue("handle", "../../private.pdf")
			case "unknown":
				r.SetPathValue("handle", uuid.New().String())
			case "removed":
				store.err = users.ErrNotFound
			case "storage failure":
				signer.err = errors.New("storage unavailable")
			}
			w := httptest.NewRecorder()
			err := (emailAvatarHandler{users: store, images: signer}).redirect(t.Context(), w, r)
			if mode == "storage failure" {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, http.StatusNotFound, w.Code)
				require.Zero(t, signer.calls)
			}
			require.Empty(t, w.Header().Get("Location"))
			require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
		})
	}
}
