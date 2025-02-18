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

// CoreVerificationToken represents a verification token in the application layer
type CoreVerificationToken struct {
	ID        uuid.UUID  `json:"id"`
	Token     string     `json:"token"`
	Email     string     `json:"email"`
	UserID    *uuid.UUID `json:"userId"`
	ExpiresAt time.Time  `json:"expiresAt"`
	UsedAt    *time.Time `json:"usedAt"`
	TokenType string     `json:"tokenType"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// Token types
const (
	TokenTypeLogin        = "login"        // For existing user login
	TokenTypeRegistration = "registration" // For new user registration
)

// Service errors
var (
	ErrNotFound           = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrEmailTaken         = errors.New("email is already taken")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenUsed          = errors.New("token has already been used")
	ErrTooManyAttempts    = errors.New("too many attempts")
	ErrInvalidToken       = errors.New("invalid token")
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
	}

	if !errors.Is(err, ErrNotFound) {
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

// CreateVerificationToken creates a new verification token
func (s *Service) CreateVerificationToken(ctx context.Context, email, tokenType string) (CoreVerificationToken, error) {
	s.log.Info(ctx, "business.core.users.CreateVerificationToken")
	ctx, span := web.AddSpan(ctx, "business.core.users.CreateVerificationToken")
	defer span.End()

	token, err := s.repo.CreateVerificationToken(ctx, email, tokenType)
	if err != nil {
		span.RecordError(err)
		return CoreVerificationToken{}, err
	}

	span.AddEvent("verification token created", trace.WithAttributes(
		attribute.String("email", email),
		attribute.String("token_type", tokenType),
	))

	return token, nil
}

// GetVerificationToken retrieves and validates a verification token
func (s *Service) GetVerificationToken(ctx context.Context, token string) (CoreVerificationToken, error) {
	s.log.Info(ctx, "business.core.users.GetVerificationToken")
	ctx, span := web.AddSpan(ctx, "business.core.users.GetVerificationToken")
	defer span.End()

	verificationToken, err := s.repo.GetVerificationToken(ctx, token)
	if err != nil {
		span.RecordError(err)
		return CoreVerificationToken{}, err
	}

	// Check if token is expired
	if time.Now().After(verificationToken.ExpiresAt) {
		return CoreVerificationToken{}, ErrTokenExpired
	}

	// Check if token is already used
	if verificationToken.UsedAt != nil {
		return CoreVerificationToken{}, ErrTokenUsed
	}

	return verificationToken, nil
}

// MarkTokenUsed marks a verification token as used
func (s *Service) MarkTokenUsed(ctx context.Context, tokenID uuid.UUID) error {
	s.log.Info(ctx, "business.core.users.MarkTokenUsed")
	ctx, span := web.AddSpan(ctx, "business.core.users.MarkTokenUsed")
	defer span.End()

	if err := s.repo.MarkTokenUsed(ctx, tokenID); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("token marked as used", trace.WithAttributes(
		attribute.String("token_id", tokenID.String()),
	))

	return nil
}

// InvalidateTokens invalidates all unused tokens for an email
func (s *Service) InvalidateTokens(ctx context.Context, email string) error {
	s.log.Info(ctx, "business.core.users.InvalidateTokens")
	ctx, span := web.AddSpan(ctx, "business.core.users.InvalidateTokens")
	defer span.End()

	if err := s.repo.InvalidateTokens(ctx, email); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("tokens invalidated", trace.WithAttributes(
		attribute.String("email", email),
	))

	return nil
}

// GetValidTokenCount gets the count of valid tokens for an email within a duration
func (s *Service) GetValidTokenCount(ctx context.Context, email string, duration time.Duration) (int, error) {
	s.log.Info(ctx, "business.core.users.GetValidTokenCount")
	ctx, span := web.AddSpan(ctx, "business.core.users.GetValidTokenCount")
	defer span.End()

	count, err := s.repo.GetValidTokenCount(ctx, email, duration)
	if err != nil {
		span.RecordError(err)
		return 0, err
	}

	span.AddEvent("valid token count retrieved", trace.WithAttributes(
		attribute.String("email", email),
		attribute.Int("count", count),
	))

	return count, nil
}
