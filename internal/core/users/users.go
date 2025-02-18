package users

import (
	"context"
	"errors"
	"strings"
	"time"

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
	ErrInvalidPassword    = errors.New("invalid password")
	ErrEmailTaken         = errors.New("email is already taken")
)

// Repository provides access to the users storage.
type Repository interface {
	GetUser(ctx context.Context, userID uuid.UUID) (CoreUser, error)
	GetUserByEmail(ctx context.Context, email string) (CoreUser, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, updates CoreUpdateUser) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	UpdateUserWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) error
	List(ctx context.Context, workspaceID uuid.UUID) ([]CoreUser, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error
	Create(ctx context.Context, user CoreUser) (CoreUser, error)
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

// ResetPassword updates a user's password after verifying their current password.
func (s *Service) ResetPassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	s.log.Info(ctx, "business.core.users.ResetPassword")
	ctx, span := web.AddSpan(ctx, "business.core.users.ResetPassword")
	defer span.End()

	// Get the user to verify current password
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Verify current password
	if err := checkHash(currentPassword, user.Password); err != nil {
		span.RecordError(err)
		return ErrInvalidPassword
	}

	// Hash new password
	hashedPassword, err := generateHash(newPassword)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Update password
	if err := s.repo.UpdatePassword(ctx, userID, hashedPassword); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("user password reset.", trace.WithAttributes(
		attribute.String("user.id", userID.String()),
	))

	return nil
}

// Register creates a new user account.
func (s *Service) Register(ctx context.Context, newUser CoreNewUser) (CoreUser, error) {
	s.log.Info(ctx, "business.core.users.Register")
	ctx, span := web.AddSpan(ctx, "business.core.users.Register")
	defer span.End()

	// Check if email is already taken
	_, err := s.repo.GetUserByEmail(ctx, newUser.Email)
	if err == nil {
		span.RecordError(ErrEmailTaken)
		return CoreUser{}, ErrEmailTaken
	} else if err != nil && !errors.Is(err, ErrNotFound) {
		span.RecordError(err)
		return CoreUser{}, err
	}

	// Generate username from email
	username := strings.Split(newUser.Email, "@")[0]

	// Hash the password
	hashedPassword, err := generateHash(newUser.Password)
	if err != nil {
		span.RecordError(err)
		return CoreUser{}, err
	}

	now := time.Now()
	user := CoreUser{
		ID:          uuid.New(),
		Username:    username,
		Email:       newUser.Email,
		Password:    hashedPassword,
		FullName:    newUser.FullName,
		AvatarURL:   newUser.AvatarURL,
		IsActive:    true,
		LastLoginAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	createdUser, err := s.repo.Create(ctx, user)
	if err != nil {
		span.RecordError(err)
		return CoreUser{}, err
	}

	span.AddEvent("user registered.", trace.WithAttributes(
		attribute.String("user.id", createdUser.ID.String()),
		attribute.String("user.email", createdUser.Email),
	))

	return createdUser, nil
}

// GetUserByEmail returns a user by email.
func (s *Service) GetUserByEmail(ctx context.Context, email string) (CoreUser, error) {
	s.log.Info(ctx, "business.core.users.GetUserByEmail")
	ctx, span := web.AddSpan(ctx, "business.core.users.GetUserByEmail")
	defer span.End()

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		span.RecordError(err)
		return CoreUser{}, err
	}

	span.AddEvent("user retrieved by email", trace.WithAttributes(
		attribute.String("user.email", email),
	))

	return user, nil
}
