package usershttp

import (
	"context"
	"errors"
	"github.com/complexus-tech/projects-api/pkg/web"
	"net/http"
	"net/url"
	"time"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/google/uuid"
)

type emailAvatarReader interface {
	GetEmailAvatar(context.Context, uuid.UUID) (string, error)
}
type emailAvatarResolver interface {
	ResolveProfileImageURL(context.Context, string, time.Duration) (string, error)
}
type emailAvatarHandler struct {
	users  emailAvatarReader
	images emailAvatarResolver
}

// The URL in the email never expires. Only this uncached redirect points to a
// short-lived storage URL, minted anew for every image request. The opaque
// handle is limited to one person's current avatar and accepts no storage key.
func (h emailAvatarHandler) redirect(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	handle, err := web.UUIDPathParameter(r, "handle")
	if err != nil || handle == uuid.Nil {
		http.NotFound(w, r)
		return nil
	}
	avatar, err := h.users.GetEmailAvatar(ctx, handle)
	if errors.Is(err, users.ErrNotFound) {
		http.NotFound(w, r)
		return nil
	}
	if err != nil {
		return err
	}
	location, err := h.images.ResolveProfileImageURL(ctx, avatar, 5*time.Minute)
	if err != nil {
		return err
	}
	parsed, err := url.Parse(location)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
		http.NotFound(w, r)
		return nil
	}
	http.Redirect(w, r, parsed.String(), http.StatusFound)
	return nil
}
