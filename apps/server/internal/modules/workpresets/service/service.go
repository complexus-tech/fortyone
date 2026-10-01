package workpresets

import (
	"context"
	"strings"

	domain "github.com/complexus-tech/projects-api/internal/modules/workpresets/domain"
	"github.com/google/uuid"
)

type Repository interface {
	List(context.Context, uuid.UUID, uuid.UUID, domain.List) ([]domain.Preset, error)
	Create(context.Context, uuid.UUID, uuid.UUID, domain.Input) (domain.Preset, error)
	Update(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, domain.Update) (domain.Preset, error)
	Archive(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error
}

type Service struct{ repo Repository }

func New(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) List(ctx context.Context, actorID, workspaceID uuid.UUID, filters domain.List) ([]domain.Preset, error) {
	if filters.TeamID == uuid.Nil || (filters.Kind != domain.View && filters.Kind != domain.Template) || filters.Limit < 1 || filters.Limit > 101 {
		return nil, domain.ErrInvalidInput
	}
	return s.repo.List(ctx, actorID, workspaceID, filters)
}

func (s *Service) Create(ctx context.Context, actorID, workspaceID uuid.UUID, input domain.Input) (domain.Preset, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.TeamID == uuid.Nil || (input.Visibility != domain.Personal && input.Visibility != domain.Team) || !validName(input.Name) {
		return domain.Preset{}, domain.ErrInvalidInput
	}
	if err := ValidateConfiguration(input.Kind, input.Configuration); err != nil {
		return domain.Preset{}, err
	}
	return s.repo.Create(ctx, actorID, workspaceID, input)
}

func (s *Service) Update(ctx context.Context, actorID, workspaceID, id uuid.UUID, input domain.Update) (domain.Preset, error) {
	input.Name = strings.TrimSpace(input.Name)
	if id == uuid.Nil || !validName(input.Name) {
		return domain.Preset{}, domain.ErrInvalidInput
	}
	// Updates deliberately rename only. New snapshots use Create so concurrent
	// changes cannot silently replace another person's saved process.
	if len(input.Configuration) != 0 {
		return domain.Preset{}, domain.ErrInvalidInput
	}
	return s.repo.Update(ctx, actorID, workspaceID, id, input)
}

func (s *Service) Archive(ctx context.Context, actorID, workspaceID, id uuid.UUID) error {
	if id == uuid.Nil {
		return domain.ErrInvalidInput
	}
	return s.repo.Archive(ctx, actorID, workspaceID, id)
}

func validName(name string) bool { return len([]rune(name)) >= 1 && len([]rune(name)) <= 100 }
