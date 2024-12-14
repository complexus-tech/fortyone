package users

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

// Repository provides access to the users storage.
type Repository interface {
	GetByEmail(ctx context.Context, email string) (CoreUser, error)
	List(ctx context.Context, workspaceId uuid.UUID) ([]CoreUser, error)
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

	user, err := s.repo.GetByEmail(ctx, email)

	if err != nil {
		span.RecordError(err)
		return CoreUser{}, err
	}

	if err := checkHash(password, user.Password); err != nil {
		span.RecordError(err)
		return CoreUser{}, err
	}

	span.AddEvent("user retrieved.", trace.WithAttributes(
		attribute.String("user.email", user.Email),
	))
	return user, nil
}

func checkHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}

func generateHash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

func (s *Service) List(ctx context.Context, workspaceId uuid.UUID) ([]CoreUser, error) {
	s.log.Info(ctx, "business.core.users.list")
	ctx, span := web.AddSpan(ctx, "business.core.users.List")
	defer span.End()

	users, err := s.repo.List(ctx, workspaceId)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("users retrieved.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceId.String()),
	))
	return users, nil
}
