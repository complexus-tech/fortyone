package keyresults

import (
	"context"
	"time"

	keyresultsdomain "github.com/complexus-tech/projects-api/internal/modules/keyresults/domain"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
)

var (
	ErrInvalid          = keyresultsdomain.ErrInvalid
	ErrForbidden        = keyresultsdomain.ErrForbidden
	ErrNotFound         = keyresultsdomain.ErrNotFound
	ErrInvalidReference = keyresultsdomain.ErrInvalidReference
	ErrVersionConflict  = keyresultsdomain.ErrVersionConflict
)

type Repository interface {
	CreateBatch(context.Context, keyresultsdomain.CreateCommand) ([]keyresultsdomain.KeyResult, error)
	Update(context.Context, keyresultsdomain.UpdateCommand) (keyresultsdomain.MutationResult, error)
	Delete(context.Context, keyresultsdomain.DeleteCommand) error
	Get(context.Context, keyresultsdomain.GetQuery) (keyresultsdomain.KeyResult, error)
	List(context.Context, keyresultsdomain.ObjectiveListQuery) ([]keyresultsdomain.KeyResult, error)
	ListPaginated(context.Context, keyresultsdomain.PaginatedListQuery) (keyresultsdomain.ListResponse, error)
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
	if eventPublisher == nil {
		return func(service *Service) { service.publisher = nil }
	}
	return WithEventPublisher(eventPublisher)
}

// WithEventPublisher keeps the service boundary owned by its caller. Production
// uses publisher.Publisher through the compatibility option above, while tests
// and future integration adapters can implement only the single method needed
// by this use case.
func WithEventPublisher(eventPublisher EventPublisher) Option {
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
