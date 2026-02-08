package users

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"mime/multipart"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Token types
const (
	TokenTypeLogin        = "login"        // For existing user login
	TokenTypeRegistration = "registration" // For new user registration
)

// Service errors
var (
	ErrNotFound              = errors.New("we couldn't find your account")
	ErrEmailTaken            = errors.New("the email address is already registered")
	ErrTokenExpired          = errors.New("the sign-in link has expired - please request a new one")
	ErrTokenUsed             = errors.New("the sign-in link has already been used")
	ErrTooManyAttempts       = errors.New("too many sign-in attempts - please wait a few minutes and try again")
	ErrInvalidToken          = errors.New("the sign-in link is invalid - please request a new one")
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrTokenNotFound         = errors.New("token not found")
	ErrWorkspaceNotFound     = errors.New("workspace not found")
)

// Repository provides access to the users storage.
type Repository interface {
	GetUser(ctx context.Context, userID uuid.UUID) (CoreUser, error)
	GetUserByEmail(ctx context.Context, email string) (CoreUser, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, updates CoreUpdateUser) (CoreUser, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	UpdateUserWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) error
	List(ctx context.Context, workspaceID uuid.UUID, teamID *uuid.UUID) ([]CoreUser, error)
	Create(ctx context.Context, user CoreUser) (CoreUser, error)
	// Verification token methods
	CreateVerificationToken(ctx context.Context, email, tokenType string, expiresAt time.Time) (CoreVerificationToken, error)
	GetVerificationToken(ctx context.Context, token string) (CoreVerificationToken, error)
	MarkTokenUsed(ctx context.Context, tokenID uuid.UUID) error
	InvalidateTokens(ctx context.Context, email string) error
	GetValidTokenCount(ctx context.Context, email string, duration time.Duration) (int, error)
	UpdateUserWorkspaceWithTx(ctx context.Context, tx *sqlx.Tx, userID, workspaceID uuid.UUID) error
	// Automation preferences methods
	GetAutomationPreferences(ctx context.Context, userID, workspaceID uuid.UUID) (CoreAutomationPreferences, error)
	UpdateAutomationPreferences(ctx context.Context, userID, workspaceID uuid.UUID, updates CoreUpdateAutomationPreferences) error
	// User Memory methods
	AddUserMemory(ctx context.Context, memory NewUserMemoryItem) (CoreUserMemoryItem, error)
	UpdateUserMemory(ctx context.Context, id uuid.UUID, update UpdateUserMemoryItem) error
	DeleteUserMemory(ctx context.Context, id uuid.UUID) error
	ListUserMemories(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID) ([]CoreUserMemoryItem, error)
}

// AttachmentsService interface for profile image operations
type AttachmentsService interface {
	UploadProfileImage(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, userID uuid.UUID) (string, error)
	UploadProfileImageFromURL(ctx context.Context, imageURL string, userID uuid.UUID) (string, error)
	DeleteProfileImage(ctx context.Context, avatarURL string) error
	ResolveProfileImageURL(ctx context.Context, avatar string, expiry time.Duration) (string, error)
}

// Service provides user-related operations.
type Service struct {
	repo         Repository
	log          *logger.Logger
	tasksService *tasks.Service
}

// New constructs a new users service instance with the provided repository.
func New(log *logger.Logger, repo Repository, tasksService *tasks.Service) *Service {
	return &Service{
		repo:         repo,
		log:          log,
		tasksService: tasksService,
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
		Timezone:    newUser.Timezone,
		IsActive:    true,
		LastLoginAt: time.Now(),
	}

	user, err = s.repo.Create(ctx, user)
	if err != nil {
		span.RecordError(err)
		return CoreUser{}, err
	}

	// Enqueue onboarding task
	_, err = s.tasksService.EnqueueUserOnboardingStart(tasks.UserOnboardingStartPayload{
		UserID:   user.ID.String(),
		Email:    user.Email,
		FullName: user.FullName,
	})
	if err != nil {
		span.RecordError(err)
		s.log.Error(ctx, "Error enqueuing onboarding task: %v", err)
	}
	return user, nil
}

// UpdateUser updates a user's profile.
func (s *Service) UpdateUser(ctx context.Context, userID uuid.UUID, updates CoreUpdateUser) error {
	s.log.Info(ctx, "business.core.users.UpdateUser")
	ctx, span := web.AddSpan(ctx, "business.core.users.UpdateUser")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID.String()))

	user, err := s.repo.UpdateUser(ctx, userID, updates)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Enqueue subscriber update task
	_, err = s.tasksService.EnqueueSubscriberUpdate(tasks.SubscriberUpdatePayload{
		Email:    user.Email,
		FullName: user.FullName,
	})
	if err != nil {
		span.RecordError(err)
		s.log.Error(ctx, "Error enqueuing subscriber update task: %v", err)
	}

	return nil
}

// DeleteUser deletes a user account.
func (s *Service) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	s.log.Info(ctx, "business.core.users.DeleteUser")
	ctx, span := web.AddSpan(ctx, "business.core.users.DeleteUser")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID.String()))

	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	if err := s.repo.DeleteUser(ctx, userID); err != nil {
		span.RecordError(err)
		return err
	}

	// Enqueue subscriber delete task
	_, err = s.tasksService.EnqueueSubscriberDelete(tasks.SubscriberDeletePayload{
		Email: user.Email,
	})
	if err != nil {
		span.RecordError(err)
		s.log.Error(ctx, "Error enqueuing subscriber delete task: %v", err)
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
func (s *Service) List(ctx context.Context, workspaceID uuid.UUID, teamID *uuid.UUID) ([]CoreUser, error) {
	ctx, span := web.AddSpan(ctx, "business.core.users.List")
	defer span.End()

	users, err := s.repo.List(ctx, workspaceID, teamID)
	if err != nil {
		return nil, fmt.Errorf("getting users list: %w", err)
	}

	return users, nil
}

// CreateVerificationToken creates a new verification token.
func (s *Service) CreateVerificationToken(ctx context.Context, email, tokenType string, expiresAt time.Time) (CoreVerificationToken, error) {
	s.log.Info(ctx, "business.core.users.CreateVerificationToken")
	ctx, span := web.AddSpan(ctx, "business.core.users.CreateVerificationToken")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", email))

	token, err := s.repo.CreateVerificationToken(ctx, email, tokenType, expiresAt)
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

// UpdateUserWorkspaceWithTx updates the user's last used workspace using an existing transaction.
func (s *Service) UpdateUserWorkspaceWithTx(ctx context.Context, tx *sqlx.Tx, userID, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "business.core.users.UpdateUserWorkspaceWithTx")
	ctx, span := web.AddSpan(ctx, "business.core.users.UpdateUserWorkspaceWithTx")
	defer span.End()

	if err := s.repo.UpdateUserWorkspaceWithTx(ctx, tx, userID, workspaceID); err != nil {
		s.log.Error(ctx, "Error updating user workspace with transaction: %v", err)
		span.RecordError(err)
		return fmt.Errorf("failed to update user last_used_workspace_id: %w", err)
	}

	span.AddEvent("user workspace updated", trace.WithAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("workspace.id", workspaceID.String()),
	))

	return nil
}

// GetAutomationPreferences retrieves a user's automation preferences for a workspace.
func (s *Service) GetAutomationPreferences(ctx context.Context, userID, workspaceID uuid.UUID) (CoreAutomationPreferences, error) {
	s.log.Info(ctx, "business.core.users.GetAutomationPreferences")
	ctx, span := web.AddSpan(ctx, "business.core.users.GetAutomationPreferences")
	defer span.End()

	preferences, err := s.repo.GetAutomationPreferences(ctx, userID, workspaceID)
	if err != nil {
		s.log.Error(ctx, "Error getting automation preferences: %v", err)
		span.RecordError(err)
		return CoreAutomationPreferences{}, fmt.Errorf("failed to get automation preferences: %w", err)
	}

	span.AddEvent("automation preferences retrieved", trace.WithAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("workspace.id", workspaceID.String()),
	))

	return preferences, nil
}

// UpdateAutomationPreferences updates a user's automation preferences for a workspace.
func (s *Service) UpdateAutomationPreferences(ctx context.Context, userID, workspaceID uuid.UUID, updates CoreUpdateAutomationPreferences) error {
	s.log.Info(ctx, "business.core.users.UpdateAutomationPreferences")
	ctx, span := web.AddSpan(ctx, "business.core.users.UpdateAutomationPreferences")
	defer span.End()

	if err := s.repo.UpdateAutomationPreferences(ctx, userID, workspaceID, updates); err != nil {
		s.log.Error(ctx, "Error updating automation preferences: %v", err)
		span.RecordError(err)
		return fmt.Errorf("failed to update automation preferences: %w", err)
	}

	span.AddEvent("automation preferences updated", trace.WithAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("workspace.id", workspaceID.String()),
	))

	return nil
}

// UploadProfileImage uploads a new profile image for a user
func (s *Service) UploadProfileImage(ctx context.Context, userID uuid.UUID, file multipart.File, fileHeader *multipart.FileHeader, attachmentsService AttachmentsService) error {
	s.log.Info(ctx, "business.core.users.uploadProfileImage")
	ctx, span := web.AddSpan(ctx, "business.core.users.UploadProfileImage")
	defer span.End()

	// Get current user to check for existing avatar
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Upload new image
	blobName, err := attachmentsService.UploadProfileImage(ctx, file, fileHeader, userID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Delete old image if exists
	if user.AvatarURL != "" {
		_ = attachmentsService.DeleteProfileImage(ctx, user.AvatarURL)
	}

	// Update user's avatar URL using pointer-based update
	updates := CoreUpdateUser{
		AvatarURL: &blobName,
	}

	_, err = s.repo.UpdateUser(ctx, userID, updates)
	if err != nil {
		span.RecordError(err)
		// Try to cleanup uploaded image since DB update failed
		_ = attachmentsService.DeleteProfileImage(ctx, blobName)
		return err
	}

	span.AddEvent("profile image updated", trace.WithAttributes(
		attribute.String("user_id", userID.String()),
		attribute.String("blob_name", blobName),
	))

	return nil
}

// DeleteProfileImage removes the current profile image
func (s *Service) DeleteProfileImage(ctx context.Context, userID uuid.UUID, attachmentsService AttachmentsService) error {
	s.log.Info(ctx, "business.core.users.deleteProfileImage")
	ctx, span := web.AddSpan(ctx, "business.core.users.DeleteProfileImage")
	defer span.End()

	// Get current user
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Delete from Azure if exists
	if user.AvatarURL != "" {
		_ = attachmentsService.DeleteProfileImage(ctx, user.AvatarURL)
	}

	// Clear avatar URL in database using pointer-based update
	avatarURL := ""
	updates := CoreUpdateUser{
		AvatarURL: &avatarURL,
	}

	_, err = s.repo.UpdateUser(ctx, userID, updates)
	if err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("profile image deleted", trace.WithAttributes(
		attribute.String("user_id", userID.String()),
	))

	return nil
}

// AddUserMemory adds a new memory item for a user.
func (s *Service) AddUserMemory(ctx context.Context, memory NewUserMemoryItem) (CoreUserMemoryItem, error) {
	s.log.Info(ctx, "business.core.users.AddUserMemory")
	ctx, span := web.AddSpan(ctx, "business.core.users.AddUserMemory")
	defer span.End()

	newItem, err := s.repo.AddUserMemory(ctx, memory)
	if err != nil {
		return CoreUserMemoryItem{}, fmt.Errorf("adding user memory: %w", err)
	}

	return newItem, nil
}

// UpdateUserMemory updates a memory item.
func (s *Service) UpdateUserMemory(ctx context.Context, id uuid.UUID, update UpdateUserMemoryItem) error {
	s.log.Info(ctx, "business.core.users.UpdateUserMemory")
	ctx, span := web.AddSpan(ctx, "business.core.users.UpdateUserMemory")
	defer span.End()

	if err := s.repo.UpdateUserMemory(ctx, id, update); err != nil {
		return fmt.Errorf("updating user memory: %w", err)
	}

	return nil
}

// DeleteUserMemory deletes a memory item.
func (s *Service) DeleteUserMemory(ctx context.Context, id uuid.UUID) error {
	s.log.Info(ctx, "business.core.users.DeleteUserMemory")
	ctx, span := web.AddSpan(ctx, "business.core.users.DeleteUserMemory")
	defer span.End()

	if err := s.repo.DeleteUserMemory(ctx, id); err != nil {
		return fmt.Errorf("deleting user memory: %w", err)
	}

	return nil
}

// ListUserMemories retrieves all memory items for a user in a workspace.
func (s *Service) ListUserMemories(ctx context.Context, userID, workspaceID uuid.UUID) ([]CoreUserMemoryItem, error) {
	s.log.Info(ctx, "business.core.users.ListUserMemories")
	ctx, span := web.AddSpan(ctx, "business.core.users.ListUserMemories")
	defer span.End()

	items, err := s.repo.ListUserMemories(ctx, userID, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("listing user memories: %w", err)
	}

	return items, nil
}
