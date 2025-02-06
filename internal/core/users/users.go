package users

import (
	"context"
	"errors"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

// Service errors
var (
	ErrNotFound           = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Repository provides access to the users storage.
type Repository interface {
	GetUser(ctx context.Context, userID uuid.UUID) (CoreUser, error)
	GetUserByEmail(ctx context.Context, email string) (CoreUser, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, updates CoreUpdateUser) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	UpdateUserWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) error
	List(ctx context.Context, workspaceID uuid.UUID) ([]CoreUser, error)
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

// GetUser returns a user by ID.
func (s *Service) GetUser(ctx context.Context, userID uuid.UUID) (CoreUser, error) {
	s.log.Info(ctx, "business.core.users.GetUser")
	ctx, span := web.AddSpan(ctx, "business.core.users.GetUser")
	defer span.End()

	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return CoreUser{}, err
	}

	span.AddEvent("user retrieved.", trace.WithAttributes(
		attribute.String("user.id", userID.String()),
	))

	return user, nil
}

// UpdateUser updates a user's profile.
func (s *Service) UpdateUser(ctx context.Context, userID uuid.UUID, updates CoreUpdateUser) error {
	s.log.Info(ctx, "business.core.users.UpdateUser")
	ctx, span := web.AddSpan(ctx, "business.core.users.UpdateUser")
	defer span.End()

	if err := s.repo.UpdateUser(ctx, userID, updates); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("user updated.", trace.WithAttributes(
		attribute.String("user.id", userID.String()),
	))

	return nil
}

// DeleteUser soft deletes a user.
func (s *Service) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	s.log.Info(ctx, "business.core.users.DeleteUser")
	ctx, span := web.AddSpan(ctx, "business.core.users.DeleteUser")
	defer span.End()

	if err := s.repo.DeleteUser(ctx, userID); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("user deleted.", trace.WithAttributes(
		attribute.String("user.id", userID.String()),
	))

	return nil
}

// UpdateUserWorkspace updates a user's last used workspace.
func (s *Service) UpdateUserWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "business.core.users.UpdateUserWorkspace")
	ctx, span := web.AddSpan(ctx, "business.core.users.UpdateUserWorkspace")
	defer span.End()

	if err := s.repo.UpdateUserWorkspace(ctx, userID, workspaceID); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("user workspace updated.", trace.WithAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("workspace.id", workspaceID.String()),
	))

	return nil
}

// Login authenticates a user and returns their profile.
func (s *Service) Login(ctx context.Context, email, password string) (CoreUser, error) {
	s.log.Info(ctx, "business.core.users.Login")
	ctx, span := web.AddSpan(ctx, "business.core.users.Login")
	defer span.End()

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		span.RecordError(err)
		return CoreUser{}, ErrInvalidCredentials
	}

	if err := checkHash(password, user.Password); err != nil {
		span.RecordError(err)
		return CoreUser{}, ErrInvalidCredentials
	}

	span.AddEvent("user logged in.", trace.WithAttributes(
		attribute.String("user.email", email),
	))

	return user, nil
}

func checkHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func generateHash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

// List returns a list of users for a workspace.
func (s *Service) List(ctx context.Context, workspaceID uuid.UUID) ([]CoreUser, error) {
	s.log.Info(ctx, "business.core.users.List")
	ctx, span := web.AddSpan(ctx, "business.core.users.List")
	defer span.End()

	users, err := s.repo.List(ctx, workspaceID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("users retrieved.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
	))
	return users, nil
}
