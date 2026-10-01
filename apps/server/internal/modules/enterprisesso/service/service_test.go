package enterprisesso

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/enterprisesso/domain"
	"github.com/complexus-tech/projects-api/internal/platform/oidcclient"
	"github.com/google/uuid"
)

type repository struct {
	Repository
	connection domain.Connection
	updated    bool
	denied     bool
}

func (r *repository) AuthorizeAdmin(context.Context, domain.Scope) error {
	if r.denied {
		return domain.ErrForbidden
	}
	return nil
}
func (r *repository) Get(context.Context, domain.Scope) (*domain.Connection, error) {
	return &r.connection, nil
}
func (r *repository) Update(_ context.Context, _ domain.Scope, c domain.Connection, _ int64) (domain.Connection, error) {
	r.updated = true
	return c, nil
}

type provider struct {
	Provider
	called bool
}

func (p *provider) Validate(context.Context, oidcclient.Credential, string) error {
	p.called = true
	return nil
}
func TestRequiredSSOOnlyAfterTestingCurrentConnection(t *testing.T) {
	r := &repository{connection: domain.Connection{ID: uuid.New(), WorkspaceID: uuid.New(), Generation: 2, Version: 3}}
	s := New(r, nil, nil, "https://api.example/callback")
	scope := domain.Scope{ActorID: uuid.New(), WorkspaceID: r.connection.WorkspaceID}
	input := domain.Update{Enabled: true, RequireSSO: true, ExpectedVersion: 3}
	for _, proof := range []domain.SessionProof{{}, {ConnectionID: uuid.New(), Generation: 2, AuthenticatedAt: time.Now()}, {ConnectionID: r.connection.ID, Generation: 1, AuthenticatedAt: time.Now()}} {
		if _, err := s.Update(t.Context(), scope, input, proof); !errors.Is(err, domain.ErrInvalid) {
			t.Fatal("untested SSO enforced", err)
		}
	}
	if r.updated {
		t.Fatal("untested setting reached storage")
	}
	proof := domain.SessionProof{ConnectionID: r.connection.ID, Generation: 2, AuthenticatedAt: time.Now()}
	if _, err := s.Update(t.Context(), scope, input, proof); err != nil || !r.updated {
		t.Fatal("verified connection could not be enforced", err)
	}
}
func TestSSOAuthorizationPrecedesProviderDiscovery(t *testing.T) {
	r := &repository{denied: true}
	p := &provider{}
	s := New(r, nil, p, "https://api.example/callback")
	_, err := s.Create(t.Context(), domain.Scope{ActorID: uuid.New(), WorkspaceID: uuid.New()}, domain.Create{Issuer: "https://untrusted.example", ClientID: "client", ClientSecret: "secret"})
	if !errors.Is(err, domain.ErrForbidden) || p.called {
		t.Fatal("unauthorized user initiated provider request", err)
	}
}
