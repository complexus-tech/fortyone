package workspacesecurity

import (
	"context"
	"errors"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/workspacesecurity/domain"
	"github.com/google/uuid"
)

type testRepository struct {
	Repository
	called bool
}

func (r *testRepository) UpdatePolicy(context.Context, domain.Scope, domain.PolicyUpdate) (domain.Policy, error) {
	r.called = true
	return domain.Policy{}, nil
}
func (r *testRepository) RevokeMember(context.Context, domain.Scope, uuid.UUID, string) error {
	r.called = true
	return nil
}
func TestServiceRejectsInvalidInputsBeforePersistence(t *testing.T) {
	r := &testRepository{}
	s := New(r)
	scope := domain.Scope{ActorID: uuid.New(), WorkspaceID: uuid.New()}
	if _, err := s.UpdatePolicy(t.Context(), domain.Scope{}, domain.PolicyUpdate{}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal(err)
	}
	if _, err := s.UpdatePolicy(t.Context(), scope, domain.PolicyUpdate{AllowedDomains: []string{"*.example.com"}}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatal(err)
	}
	if err := s.RevokeMember(t.Context(), scope, uuid.New(), " "); !errors.Is(err, domain.ErrInvalid) {
		t.Fatal(err)
	}
	if r.called {
		t.Fatal("invalid security input reached repository")
	}
}
