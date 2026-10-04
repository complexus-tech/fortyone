package usershttp

import (
	"context"
	"net/http"
	"time"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/deployment"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/google/uuid"
)

// WorkspaceSSOSessionWriter is a narrow user-owned session issuance capability.
// Enterprise SSO never reaches account storage or writes cookies itself.
type WorkspaceSSOSessionWriter struct{ handlers *Handlers }

func NewWorkspaceSSOSessionWriter(users *users.Service, sessions *cache.Service, cookieDomain string, mode deployment.Mode) *WorkspaceSSOSessionWriter {
	return &WorkspaceSSOSessionWriter{handlers: &Handlers{users: users, cache: sessions, cookieDomain: cookieDomain, deploymentMode: mode}}
}
func (w *WorkspaceSSOSessionWriter) Issue(ctx context.Context, response http.ResponseWriter, request *http.Request, userID uuid.UUID, proof auth.WorkspaceSSOAuthentication) error {
	version, active, err := w.handlers.users.ResolveActiveBrowserSessionVersion(ctx, userID)
	if err != nil {
		return err
	}
	if !active {
		return auth.ErrInvalidBrowserSession
	}
	session, err := auth.NewBrowserSession(userID, version)
	if err != nil {
		return err
	}
	if proof.WorkspaceID == uuid.Nil || proof.ConnectionID == uuid.Nil || proof.Generation < 1 || proof.AuthenticatedAt.IsZero() {
		return auth.ErrInvalidBrowserSession
	}
	token, err := w.handlers.createSessionToken()
	if err != nil {
		return err
	}
	expires := time.Now().UTC().Add(SessionDuration)
	session.AuthenticatedAt = proof.AuthenticatedAt
	session.ExpiresAt = expires
	session.WorkspaceSSO = &proof
	session.BrowserName = auth.BrowserName(request.UserAgent(), request.Header.Get("Sec-CH-UA"))
	if err := w.handlers.cache.Set(ctx, cache.AuthSessionCacheKey(token), session, time.Until(expires)); err != nil {
		return err
	}
	w.handlers.setSessionCookie(response, request, token, expires)
	return nil
}
