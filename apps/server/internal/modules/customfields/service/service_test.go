package customfields

import (
	"context"
	"errors"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/customfields/domain"
	"github.com/google/uuid"
)

type policyRepository struct {
	Repository
	called bool
}

func (r *policyRepository) Patch(context.Context, domain.Scope, uuid.UUID, domain.ValuePatch) (domain.Snapshot, error) {
	r.called = true
	return domain.Snapshot{}, nil
}
func (r *policyRepository) Report(context.Context, domain.Scope, domain.ReportInput) (domain.Report, error) {
	r.called = true
	return domain.Report{}, nil
}

func TestServiceRejectsUnauthenticatedScopeAndDuplicatePatches(t *testing.T) {
	repository := &policyRepository{}
	service := New(repository)
	if _, err := service.Patch(t.Context(), domain.Scope{}, uuid.New(), domain.ValuePatch{}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("empty scope = %v", err)
	}
	id := uuid.New()
	scope := domain.Scope{ActorID: uuid.New(), WorkspaceID: uuid.New()}
	if _, err := service.Patch(t.Context(), scope, uuid.New(), domain.ValuePatch{Values: []domain.Value{{FieldID: id}, {FieldID: id}}}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("duplicate fields = %v", err)
	}
	if repository.called {
		t.Fatal("invalid patch reached persistence")
	}
	if _, err := service.Report(t.Context(), scope, domain.ReportInput{FieldID: id, Aggregation: "sum", GroupBy: "status", DateBasis: "bad"}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("report basis = %v", err)
	}
	if repository.called {
		t.Fatal("invalid report reached persistence")
	}
}
