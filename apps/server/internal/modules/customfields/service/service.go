package customfields

import (
	"context"
	"fmt"
	"strings"

	domain "github.com/complexus-tech/projects-api/internal/modules/customfields/domain"
	"github.com/google/uuid"
)

type Repository interface {
	List(context.Context, domain.Scope) ([]domain.Field, error)
	Create(context.Context, domain.Scope, domain.Definition) (domain.Field, error)
	Update(context.Context, domain.Scope, uuid.UUID, domain.Definition) (domain.Field, error)
	Archive(context.Context, domain.Scope, uuid.UUID) error
	Snapshot(context.Context, domain.Scope, uuid.UUID) (domain.Snapshot, error)
	Batch(context.Context, domain.Scope, []uuid.UUID) (domain.BatchSnapshot, error)
	Patch(context.Context, domain.Scope, uuid.UUID, domain.ValuePatch) (domain.Snapshot, error)
	Report(context.Context, domain.Scope, domain.ReportInput) (domain.Report, error)
}

type Service struct{ repository Repository }

func New(repository Repository) *Service { return &Service{repository: repository} }

func validateScope(scope domain.Scope) error {
	if scope.ActorID == uuid.Nil || scope.WorkspaceID == uuid.Nil {
		return domain.ErrForbidden
	}
	return nil
}

func (s *Service) List(ctx context.Context, scope domain.Scope) ([]domain.Field, error) {
	if err := validateScope(scope); err != nil {
		return nil, err
	}
	if scope.TeamID == uuid.Nil {
		return nil, domain.ErrInvalid
	}
	return s.repository.List(ctx, scope)
}

func (s *Service) Create(ctx context.Context, scope domain.Scope, input domain.Definition) (domain.Field, error) {
	if err := validateScope(scope); err != nil {
		return domain.Field{}, err
	}
	if scope.TeamID == uuid.Nil {
		return domain.Field{}, domain.ErrInvalid
	}
	input.Name = strings.TrimSpace(input.Name)
	if err := domain.ValidateDefinition(input, nil); err != nil {
		return domain.Field{}, err
	}
	for _, option := range input.Options {
		if option.ID != uuid.Nil {
			return domain.Field{}, fmt.Errorf("%w: new options must not specify IDs", domain.ErrInvalid)
		}
	}
	return s.repository.Create(ctx, scope, input)
}

func (s *Service) Update(ctx context.Context, scope domain.Scope, id uuid.UUID, input domain.Definition) (domain.Field, error) {
	if err := validateScope(scope); err != nil {
		return domain.Field{}, err
	}
	if scope.TeamID == uuid.Nil || id == uuid.Nil {
		return domain.Field{}, domain.ErrInvalid
	}
	input.Name = strings.TrimSpace(input.Name)
	input.IconSet = input.IconSet || input.Icon != nil
	return s.repository.Update(ctx, scope, id, input)
}

func (s *Service) Archive(ctx context.Context, scope domain.Scope, id uuid.UUID) error {
	if err := validateScope(scope); err != nil {
		return err
	}
	if scope.TeamID == uuid.Nil || id == uuid.Nil {
		return domain.ErrInvalid
	}
	return s.repository.Archive(ctx, scope, id)
}

func (s *Service) Snapshot(ctx context.Context, scope domain.Scope, storyID uuid.UUID) (domain.Snapshot, error) {
	if err := validateScope(scope); err != nil {
		return domain.Snapshot{}, err
	}
	if storyID == uuid.Nil {
		return domain.Snapshot{}, domain.ErrInvalid
	}
	return s.repository.Snapshot(ctx, scope, storyID)
}

func (s *Service) Batch(ctx context.Context, scope domain.Scope, storyIDs []uuid.UUID) (domain.BatchSnapshot, error) {
	if err := validateScope(scope); err != nil {
		return domain.BatchSnapshot{}, err
	}
	if len(storyIDs) < 1 || len(storyIDs) > 100 {
		return domain.BatchSnapshot{}, domain.ErrInvalid
	}
	seen := make(map[uuid.UUID]bool, len(storyIDs))
	for _, id := range storyIDs {
		if id == uuid.Nil || seen[id] {
			return domain.BatchSnapshot{}, domain.ErrInvalid
		}
		seen[id] = true
	}
	return s.repository.Batch(ctx, scope, storyIDs)
}

func (s *Service) Patch(ctx context.Context, scope domain.Scope, storyID uuid.UUID, patch domain.ValuePatch) (domain.Snapshot, error) {
	if err := validateScope(scope); err != nil {
		return domain.Snapshot{}, err
	}
	if storyID == uuid.Nil {
		return domain.Snapshot{}, domain.ErrInvalid
	}
	if err := domain.ValidatePatch(patch); err != nil {
		return domain.Snapshot{}, err
	}
	return s.repository.Patch(ctx, scope, storyID, patch)
}

func (s *Service) Report(ctx context.Context, scope domain.Scope, input domain.ReportInput) (domain.Report, error) {
	if err := validateScope(scope); err != nil {
		return domain.Report{}, err
	}
	if err := domain.ValidateReport(input); err != nil {
		return domain.Report{}, err
	}
	return s.repository.Report(ctx, scope, input)
}
