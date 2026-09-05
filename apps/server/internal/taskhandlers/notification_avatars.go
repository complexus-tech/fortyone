package taskhandlers

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
)

type EmailAvatarHandleStore interface {
	EnsureEmailAvatarHandle(context.Context, uuid.UUID) (uuid.UUID, error)
}

func emailAvatarURL(base string, handle uuid.UUID) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(base))
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || handle == uuid.Nil {
		return "", errors.New("email portraits require a public HTTPS API URL and an avatar handle")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/media/email-avatars/" + handle.String() + "/avatar"
	parsed.RawPath = ""
	return parsed.String(), nil
}

// Resolve only visible actors. The email contains a durable application URL,
// never an expiring storage URL or an unrestricted storage filename.
func (h *handlers) resolveDigestAvatars(ctx context.Context, digest *mailer.Digest) {
	resolved := make(map[uuid.UUID]string)
	for i := range digest.Rows {
		actor := &digest.Rows[i].Actor
		if actor.AvatarURL == "" {
			continue
		}
		actor.AvatarURL = ""
		if h.emailAvatars == nil || actor.ID == uuid.Nil {
			continue
		}
		if cached, exists := resolved[actor.ID]; exists {
			actor.AvatarURL = cached
			continue
		}
		handle, err := h.emailAvatars.EnsureEmailAvatarHandle(ctx, actor.ID)
		if err != nil {
			h.log.Warn(ctx, "Email portrait unavailable; using initials", "error", err)
			resolved[actor.ID] = ""
			continue
		}
		location, err := emailAvatarURL(h.apiPublicURL, handle)
		if err != nil {
			h.log.Warn(ctx, "Email portrait URL unavailable; using initials", "error", err)
			resolved[actor.ID] = ""
			continue
		}
		resolved[actor.ID], actor.AvatarURL = location, location
	}
}
