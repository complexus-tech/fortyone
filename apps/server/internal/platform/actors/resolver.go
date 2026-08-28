// Package actors resolves the small, closed set of first-party system actors.
package actors

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Key identifies a first-party actor without exposing its persisted UUID to
// callers or configuration.
type Key string

const (
	KeySystem Key = "system"
	KeyGitHub Key = "github"
)

var actorEmails = map[Key]string{
	KeySystem: "maya@fortyone.app",
	KeyGitHub: "github@fortyone.app",
}

// Lookup is the persistence capability required by the resolver. Bootstrap
// supplies the SQLC/pgx adapter; callers and tests do not depend on generated
// query types.
type Lookup interface {
	FindActiveSystemActorByEmail(ctx context.Context, email string) (uuid.UUID, error)
}

// Resolver maps stable application keys to live system-user records. System
// actors are resolved during process startup, so a second cache would add stale
// authorization state without removing a meaningful database hot path.
type Resolver struct {
	lookup Lookup
}

func NewResolver(lookup Lookup) *Resolver {
	return &Resolver{lookup: lookup}
}

// EmailForKey returns the configured system actor email for a key.
func EmailForKey(key Key) (string, bool) {
	email, ok := actorEmails[key]
	return email, ok
}

func (resolver *Resolver) Resolve(ctx context.Context, key Key) (uuid.UUID, error) {
	email, ok := actorEmails[key]
	if !ok {
		return uuid.Nil, fmt.Errorf("unknown actor key: %s", key)
	}
	if resolver == nil || resolver.lookup == nil {
		return uuid.Nil, fmt.Errorf("system actor lookup is not configured")
	}

	actorID, err := resolver.lookup.FindActiveSystemActorByEmail(ctx, email)
	if err != nil {
		return uuid.Nil, fmt.Errorf("resolve system actor %q: %w", email, err)
	}
	if actorID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("resolve system actor %q: lookup returned a zero id", email)
	}
	return actorID, nil
}
