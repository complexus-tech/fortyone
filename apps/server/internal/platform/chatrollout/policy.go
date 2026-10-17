// Package chatrollout reads the live workspace Messages setting. It never grants membership.
package chatrollout

import (
	"context"
	"github.com/google/uuid"
)

// Reader must perform an authoritative, read-only lookup. Missing workspaces
// and database failures must not enable Messages.
type Reader interface {
	ChatEnabled(context.Context, uuid.UUID) (bool, error)
}

type ReaderFunc func(context.Context, uuid.UUID) (bool, error)

func (f ReaderFunc) ChatEnabled(ctx context.Context, id uuid.UUID) (bool, error) { return f(ctx, id) }

// Policy has no permission cache; toggles apply without restarting processes.
// HTTP and SSE use the flag from their existing workspace authorization read.
type Policy struct{ reader Reader }

func New(reader Reader) Policy { return Policy{reader: reader} }

// Enabled preserves database errors so callers inside a unit of work can retry
// serialization conflicts instead of treating them as permission revocations.
func (p Policy) Enabled(ctx context.Context, id uuid.UUID) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if p.reader == nil || id == uuid.Nil {
		return false, nil
	}
	enabled, err := p.reader.ChatEnabled(ctx, id)
	if err != nil {
		return false, err
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return enabled, nil
}
func (p Policy) Allows(ctx context.Context, id uuid.UUID) bool {
	enabled, err := p.Enabled(ctx, id)
	return err == nil && enabled
}
