package auth

import (
	"errors"

	"github.com/google/uuid"
)

var ErrInvalidBrowserSession = errors.New("invalid browser session")

// BrowserSession is the versioned, server-side identity bound to an opaque
// first-party session token. Version is compared with authoritative account
// state on every authenticated request so deactivation and revocation take
// effect without enumerating Redis keys.
type BrowserSession struct {
	UserID  uuid.UUID `json:"user_id"`
	Version int64     `json:"version"`
}

func NewBrowserSession(userID uuid.UUID, version int64) (BrowserSession, error) {
	session := BrowserSession{UserID: userID, Version: version}
	if err := session.Validate(); err != nil {
		return BrowserSession{}, err
	}
	return session, nil
}

func (session BrowserSession) Validate() error {
	if session.UserID == uuid.Nil || session.Version <= 0 {
		return ErrInvalidBrowserSession
	}
	return nil
}
