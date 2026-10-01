package workautomations

import (
	"context"
	"strings"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/workautomations/domain"
	"github.com/google/uuid"
)

type Repository interface {
	List(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) ([]domain.Automation, error)
	Get(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (domain.Automation, error)
	Create(context.Context, uuid.UUID, uuid.UUID, domain.Input, *time.Time) (domain.Automation, error)
	Pause(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, bool, *time.Time) (domain.Automation, error)
	Archive(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error
	Runs(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) ([]domain.Run, error)
}

type Service struct{ repository Repository }

func New(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) List(ctx context.Context, actor, workspace, team uuid.UUID) ([]domain.Automation, error) {
	if team == uuid.Nil {
		return nil, domain.ErrInvalidInput
	}
	return s.repository.List(ctx, actor, workspace, team)
}

func (s *Service) Create(ctx context.Context, actor, workspace uuid.UUID, input domain.Input) (domain.Automation, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.TeamID == uuid.Nil || len([]rune(input.Name)) < 1 || len([]rune(input.Name)) > 100 {
		return domain.Automation{}, domain.ErrInvalidInput
	}
	var next *time.Time
	switch input.Kind {
	case "rule":
		if _, err := validateRule(input.Configuration); err != nil {
			return domain.Automation{}, err
		}
	case "recurrence":
		config, err := validateRecurrence(input.Configuration)
		if err != nil {
			return domain.Automation{}, err
		}
		due, err := NextOccurrence(config.Schedule, time.Now())
		if err != nil {
			return domain.Automation{}, err
		}
		next = &due
	default:
		return domain.Automation{}, domain.ErrInvalidInput
	}
	return s.repository.Create(ctx, actor, workspace, input, next)
}

func (s *Service) Pause(ctx context.Context, actor, workspace, id uuid.UUID, paused bool) (domain.Automation, error) {
	if id == uuid.Nil {
		return domain.Automation{}, domain.ErrInvalidInput
	}
	current, err := s.repository.Get(ctx, actor, workspace, id)
	if err != nil {
		return domain.Automation{}, err
	}
	if !current.CanEdit {
		return domain.Automation{}, domain.ErrNotFound
	}
	next := current.NextRunAt
	if !paused && current.Kind == "recurrence" {
		config, err := validateRecurrence(current.Configuration)
		if err != nil {
			return domain.Automation{}, err
		}
		due, err := NextOccurrence(config.Schedule, time.Now())
		if err != nil {
			return domain.Automation{}, err
		}
		next = &due
	}
	return s.repository.Pause(ctx, actor, workspace, id, paused, next)
}

func (s *Service) Archive(ctx context.Context, actor, workspace, id uuid.UUID) error {
	if id == uuid.Nil {
		return domain.ErrInvalidInput
	}
	return s.repository.Archive(ctx, actor, workspace, id)
}

func (s *Service) Runs(ctx context.Context, actor, workspace, id uuid.UUID) ([]domain.Run, error) {
	if id == uuid.Nil {
		return nil, domain.ErrInvalidInput
	}
	if _, err := s.repository.Get(ctx, actor, workspace, id); err != nil {
		return nil, err
	}
	return s.repository.Runs(ctx, actor, workspace, id)
}
