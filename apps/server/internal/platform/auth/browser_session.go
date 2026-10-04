package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidBrowserSession = errors.New("invalid browser session")

// BrowserSession is the versioned, server-side identity bound to an opaque
// first-party session token. Version is compared with authoritative account
// state on every authenticated request so deactivation and revocation take
// effect without enumerating Redis keys.
type BrowserSession struct {
	UserID          uuid.UUID                   `json:"user_id"`
	Version         int64                       `json:"version"`
	SessionID       uuid.UUID                   `json:"session_id,omitempty"`
	AuthenticatedAt time.Time                   `json:"authenticated_at,omitempty"`
	ExpiresAt       time.Time                   `json:"expires_at,omitempty"`
	BrowserName     *string                     `json:"browser_name,omitempty"`
	WorkspaceSSO    *WorkspaceSSOAuthentication `json:"workspace_sso,omitempty"`
}

// This assertion is created only after verified tenant OIDC authentication.
// It is metadata, never an identity-provider token or bearer credential.
type WorkspaceSSOAuthentication struct {
	WorkspaceID     uuid.UUID `json:"workspace_id"`
	ConnectionID    uuid.UUID `json:"connection_id"`
	Generation      int64     `json:"generation"`
	AuthenticatedAt time.Time `json:"authenticated_at"`
}

func NewBrowserSession(userID uuid.UUID, version int64) (BrowserSession, error) {
	session := BrowserSession{UserID: userID, Version: version, SessionID: uuid.New(), AuthenticatedAt: time.Now().UTC()}
	if err := session.Validate(); err != nil {
		return BrowserSession{}, err
	}
	return session, nil
}

const browserSessionKey contextKey = "browser-session"

// Session metadata is server-sourced and contains no bearer credential. Its
// authentication time must survive cookie renewal and native handoffs.
func SetBrowserSession(ctx context.Context, session BrowserSession) context.Context {
	return context.WithValue(ctx, browserSessionKey, session)
}

func GetBrowserSession(ctx context.Context) (BrowserSession, bool) {
	session, ok := ctx.Value(browserSessionKey).(BrowserSession)
	return session, ok
}

func (session BrowserSession) Validate() error {
	if session.UserID == uuid.Nil || session.Version <= 0 {
		return ErrInvalidBrowserSession
	}
	return nil
}
