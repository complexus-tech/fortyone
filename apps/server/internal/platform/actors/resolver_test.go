package actors_test

import (
	"context"
	"errors"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/actors"
	"github.com/google/uuid"
)

type actorLookupStub struct {
	email string
	id    uuid.UUID
	err   error
}

func (stub *actorLookupStub) FindActiveSystemActorByEmail(_ context.Context, email string) (uuid.UUID, error) {
	stub.email = email
	return stub.id, stub.err
}

func TestResolverUsesClosedKeyCatalogAndLiveLookup(t *testing.T) {
	t.Parallel()

	actorID := uuid.New()
	lookup := &actorLookupStub{id: actorID}
	resolver := actors.NewResolver(lookup)
	resolved, err := resolver.Resolve(context.Background(), actors.KeySystem)
	if err != nil {
		t.Fatalf("resolve system actor: %v", err)
	}
	if resolved != actorID || lookup.email != "maya@fortyone.app" {
		t.Fatalf("resolved actor/email = %s/%q", resolved, lookup.email)
	}
	if _, err := resolver.Resolve(context.Background(), actors.Key("unregistered")); err == nil {
		t.Fatal("unknown actor key error = nil")
	}
}

func TestResolverRejectsLookupFailuresAndZeroIDs(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("lookup failed")
	if _, err := actors.NewResolver(&actorLookupStub{err: sentinel}).Resolve(
		context.Background(),
		actors.KeyGitHub,
	); !errors.Is(err, sentinel) {
		t.Fatalf("lookup error = %v, want wrapped sentinel", err)
	}
	if _, err := actors.NewResolver(&actorLookupStub{}).Resolve(
		context.Background(),
		actors.KeyGitHub,
	); err == nil {
		t.Fatal("zero actor id error = nil")
	}
}
