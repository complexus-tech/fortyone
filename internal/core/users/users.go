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
)

// Token types
const (
	TokenTypeLogin        = "login"        // For existing user login
	TokenTypeRegistration = "registration" // For new user registration
)

// Service errors
var (
	ErrNotFound        = errors.New("we couldn't find your account")
	ErrEmailTaken      = errors.New("the email address is already registered")
	ErrTokenExpired    = errors.New("the sign-in link has expired - please request a new one")
	ErrTokenUsed       = errors.New("the sign-in link has already been used")
	ErrTooManyAttempts = errors.New("too many sign-in attempts - please wait a few minutes and try again")
	ErrInvalidToken    = errors.New("the sign-in link is invalid - please request a new one")
)

// Repository provides access to the users storage.
type Repository interface {
	GetUser(ctx context.Context, userID uuid.UUID) (CoreUser, error)
	GetUserByEmail(ctx context.Context, email string) (CoreUser, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, updates CoreUpdateUser) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	UpdateUserWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) error
	List(ctx context.Context, workspaceID uuid.UUID) ([]CoreUser, error)
	Create(ctx context.Context, user CoreUser) (CoreUser, error)
	// Verification token methods
	CreateVerificationToken(ctx context.Context, email, tokenType string) (CoreVerificationToken, error)
	GetVerificationToken(ctx context.Context, token string) (CoreVerificationToken, error)
	MarkTokenUsed(ctx context.Context, tokenID uuid.UUID) error
	InvalidateTokens(ctx context.Context, email string) error
	GetValidTokenCount(ctx context.Context, email string, duration time.Duration) (int, error)
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

	span.SetAttributes(attribute.String("user.id", userID.String()))

	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return CoreUser{}, err
	}

	return user, nil
}

// GetUserByEmail returns a user by email.
func (s *Service) GetUserByEmail(ctx context.Context, email string) (CoreUser, error) {
	s.log.Info(ctx, "business.core.users.GetUserByEmail")
	ctx, span := web.AddSpan(ctx, "business.core.users.GetUserByEmail")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", email))

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		span.RecordError(err)
		return CoreUser{}, err
	}

	return user, nil
}

// Register creates a new user account.
func (s *Service) Register(ctx context.Context, newUser CoreNewUser) (CoreUser, error) {
	s.log.Info(ctx, "business.core.users.Register")
	ctx, span := web.AddSpan(ctx, "business.core.users.Register")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", newUser.Email))

	// Check if email is already taken
	_, err := s.repo.GetUserByEmail(ctx, newUser.Email)
	if err == nil {
		span.RecordError(ErrEmailTaken)
		return CoreUser{}, ErrEmailTaken
	}
	if !errors.Is(err, ErrNotFound) {
		span.RecordError(err)
		return CoreUser{}, err
	}

	// Create user
	user := CoreUser{
		Username:    strings.Split(newUser.Email, "@")[0],
		Email:       newUser.Email,
		FullName:    newUser.FullName,
		AvatarURL:   newUser.AvatarURL,
		IsActive:    true,
		LastLoginAt: time.Now(),
	}

	user, err = s.repo.Create(ctx, user)
	if err != nil {
		span.RecordError(err)
		return CoreUser{}, err
	}

	return user, nil
}

// UpdateUser updates a user's profile.
func (s *Service) UpdateUser(ctx context.Context, userID uuid.UUID, updates CoreUpdateUser) error {
	s.log.Info(ctx, "business.core.users.UpdateUser")
	ctx, span := web.AddSpan(ctx, "business.core.users.UpdateUser")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID.String()))

	if err := s.repo.UpdateUser(ctx, userID, updates); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

// DeleteUser deletes a user account.
func (s *Service) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	s.log.Info(ctx, "business.core.users.DeleteUser")
	ctx, span := web.AddSpan(ctx, "business.core.users.DeleteUser")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID.String()))

	if err := s.repo.DeleteUser(ctx, userID); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

// UpdateUserWorkspace updates the user's last used workspace.
func (s *Service) UpdateUserWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "business.core.users.UpdateUserWorkspace")
	ctx, span := web.AddSpan(ctx, "business.core.users.UpdateUserWorkspace")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("workspace.id", workspaceID.String()),
	)

	if err := s.repo.UpdateUserWorkspace(ctx, userID, workspaceID); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

// List returns a list of users for a workspace.
func (s *Service) List(ctx context.Context, workspaceID uuid.UUID) ([]CoreUser, error) {
	s.log.Info(ctx, "business.core.users.List")
	ctx, span := web.AddSpan(ctx, "business.core.users.List")
	defer span.End()

	span.SetAttributes(attribute.String("workspace.id", workspaceID.String()))

	users, err := s.repo.List(ctx, workspaceID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return users, nil
}

// CreateVerificationToken creates a new verification token.
func (s *Service) CreateVerificationToken(ctx context.Context, email, tokenType string) (CoreVerificationToken, error) {
	s.log.Info(ctx, "business.core.users.CreateVerificationToken")
	ctx, span := web.AddSpan(ctx, "business.core.users.CreateVerificationToken")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", email))

	token, err := s.repo.CreateVerificationToken(ctx, email, tokenType)
	if err != nil {
		span.RecordError(err)
		return CoreVerificationToken{}, err
	}

	return token, nil
}

// GetVerificationToken returns a verification token.
func (s *Service) GetVerificationToken(ctx context.Context, token string) (CoreVerificationToken, error) {
	s.log.Info(ctx, "business.core.users.GetVerificationToken")
	ctx, span := web.AddSpan(ctx, "business.core.users.GetVerificationToken")
	defer span.End()

	verificationToken, err := s.repo.GetVerificationToken(ctx, token)
	if err != nil {
		span.RecordError(err)
		return CoreVerificationToken{}, err
	}

	// Check if token has expired
	if time.Now().After(verificationToken.ExpiresAt) {
		span.RecordError(ErrTokenExpired)
		return CoreVerificationToken{}, ErrTokenExpired
	}

	// Check if token has been used
	if verificationToken.UsedAt != nil {
		span.RecordError(ErrTokenUsed)
		return CoreVerificationToken{}, ErrTokenUsed
	}

	return verificationToken, nil
}

// MarkTokenUsed marks a verification token as used.
func (s *Service) MarkTokenUsed(ctx context.Context, tokenID uuid.UUID) error {
	s.log.Info(ctx, "business.core.users.MarkTokenUsed")
	ctx, span := web.AddSpan(ctx, "business.core.users.MarkTokenUsed")
	defer span.End()

	span.SetAttributes(attribute.String("token.id", tokenID.String()))

	if err := s.repo.MarkTokenUsed(ctx, tokenID); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

// InvalidateTokens invalidates all unused tokens for an email.
func (s *Service) InvalidateTokens(ctx context.Context, email string) error {
	s.log.Info(ctx, "business.core.users.InvalidateTokens")
	ctx, span := web.AddSpan(ctx, "business.core.users.InvalidateTokens")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", email))

	if err := s.repo.InvalidateTokens(ctx, email); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

// GetValidTokenCount returns the number of valid tokens for an email within a duration.
func (s *Service) GetValidTokenCount(ctx context.Context, email string, duration time.Duration) (int, error) {
	s.log.Info(ctx, "business.core.users.GetValidTokenCount")
	ctx, span := web.AddSpan(ctx, "business.core.users.GetValidTokenCount")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", email))

	count, err := s.repo.GetValidTokenCount(ctx, email, duration)
	if err != nil {
		span.RecordError(err)
		return 0, err
	}

	return count, nil
}
