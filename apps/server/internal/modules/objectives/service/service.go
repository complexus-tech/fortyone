package objectives

import (
	"context"
	"time"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/google/uuid"
)

var (
	ErrInvalid          = objectivesdomain.ErrInvalid
	ErrForbidden        = objectivesdomain.ErrForbidden
	ErrNotFound         = objectivesdomain.ErrNotFound
	ErrNameExists       = objectivesdomain.ErrNameExists
	ErrVersionConflict  = objectivesdomain.ErrVersionConflict
	ErrInvalidReference = objectivesdomain.ErrInvalidReference
)

type Repository interface {
	List(context.Context, objectivesdomain.ListQuery) ([]objectivesdomain.Objective, error)
	Get(context.Context, objectivesdomain.GetQuery) (objectivesdomain.Objective, error)
	Create(context.Context, objectivesdomain.CreateCommand) (objectivesdomain.CreateResult, error)
	Update(context.Context, objectivesdomain.UpdateCommand) (objectivesdomain.Objective, error)
	Delete(context.Context, objectivesdomain.DeleteCommand) error
	GetAnalytics(context.Context, objectivesdomain.AnalyticsQuery, time.Time) (objectivesdomain.ObjectiveAnalytics, error)
	GetStrategyMap(context.Context, objectivesdomain.StrategyQuery) (objectivesdomain.StrategyMap, error)
	UpdateStrategy(context.Context, objectivesdomain.StrategyQuery, objectivesdomain.StrategyUpdate) error
	CreateStrategicPillar(context.Context, objectivesdomain.StrategyQuery, objectivesdomain.NewStrategicPillar) (objectivesdomain.StrategicPillar, error)
	UpdateStrategicPillar(context.Context, objectivesdomain.StrategyQuery, uuid.UUID, objectivesdomain.UpdateStrategicPillar) (objectivesdomain.StrategicPillar, error)
	DeleteStrategicPillar(context.Context, objectivesdomain.StrategyQuery, uuid.UUID) error
	AlignObjective(context.Context, objectivesdomain.StrategyQuery, uuid.UUID, *uuid.UUID) error
}

type EventPublisher interface {
	Publish(context.Context, events.Event) error
}

type Service struct {
	repo      Repository
	log       *logger.Logger
	publisher EventPublisher
	now       func() time.Time
}

type Option func(*Service)

func WithPublisher(eventPublisher *publisher.Publisher) Option {
	return func(service *Service) { service.publisher = eventPublisher }
}

func withEventPublisher(eventPublisher EventPublisher) Option {
	return func(service *Service) { service.publisher = eventPublisher }
}

func withClock(now func() time.Time) Option {
	return func(service *Service) { service.now = now }
}

func New(log *logger.Logger, repo Repository, options ...Option) *Service {
	service := &Service{repo: repo, log: log, now: time.Now}
	for _, option := range options {
		option(service)
	}
	return service
}
