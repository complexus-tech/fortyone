package users

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Repository provides access to the users storage.
type Repository interface {
	Login(ctx context.Context, email, password string) (CoreUser, error)
}

// Service provides user-related operations.
type Service struct {
	repo Repository
	log  *logger.Logger
}

// New constructs a new users service instance with the provided repository.
func New(log *logger.Logger, repo Repository) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

// Login returns a user.
func (s *Service) Login(ctx context.Context, email, password string) (CoreUser, error) {
	s.log.Info(ctx, "business.core.users.login")
	ctx, span := web.AddSpan(ctx, "business.core.users.Login")
	defer span.End()

	user, err := s.repo.Login(ctx, email, password)
	if err != nil {
		span.RecordError(err)
		return CoreUser{}, err
	}
	span.AddEvent("user retrieved.", trace.WithAttributes(
		attribute.String("user.email", user.Email),
	))
	return user, nil
}
