package users

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"mime/multipart"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	platformclock "github.com/complexus-tech/projects-api/internal/platform/clock"
	"github.com/complexus-tech/projects-api/internal/platform/workschedule"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	apptracing "github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/google/uuid"
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
	ErrNotFound              = usersdomain.ErrNotFound
	ErrEmailTaken            = usersdomain.ErrEmailTaken
	ErrTokenExpired          = usersdomain.ErrTokenExpired
	ErrTokenUsed             = usersdomain.ErrTokenUsed
	ErrTooManyAttempts       = usersdomain.ErrTooManyAttempts
	ErrInvalidToken          = usersdomain.ErrInvalidToken
	ErrUserNotFound          = usersdomain.ErrUserNotFound
	ErrInvalidCredentials    = usersdomain.ErrInvalidCredentials
	ErrEmailAlreadyExists    = usersdomain.ErrEmailAlreadyExists
	ErrUsernameAlreadyExists = usersdomain.ErrUsernameAlreadyExists
	ErrTokenNotFound         = usersdomain.ErrTokenNotFound
	ErrTokenCollision        = usersdomain.ErrTokenCollision
	ErrVerificationDisabled  = usersdomain.ErrVerificationDisabled
	ErrWorkspaceNotFound     = usersdomain.ErrWorkspaceNotFound
	ErrMemoryNotFound        = usersdomain.ErrMemoryNotFound
	ErrIdentityStoreMissing  = usersdomain.ErrIdentityStoreMissing
)

// AttachmentsService interface for profile image operations
type AttachmentsService interface {
	UploadProfileImage(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, userID uuid.UUID) (string, error)
	UploadProfileImageFromURL(ctx context.Context, imageURL string, userID uuid.UUID) (string, error)
	DeleteProfileImage(ctx context.Context, avatarURL string) error
	ResolveProfileImageURL(ctx context.Context, avatar string, expiry time.Duration) (string, error)
}

// Service provides user-related operations.
type Service struct {
	repo               Repository
	log                *logger.Logger
	tasksService       *tasks.Service
	verificationTokens *VerificationTokenManager
	verificationRepo   VerificationTokenRepository
	clock              platformclock.Clock
}

// Option configures an optional users service capability.
type Option func(*Service)

// WithClock supplies the application decision clock used by account-state
// mutations. Production defaults to the system clock; tests should inject a
// deterministic clock when exact persistence timestamps matter.
func WithClock(clock platformclock.Clock) Option {
	return func(service *Service) {
		if clock != nil {
			service.clock = clock
		}
	}
}

// WithVerificationTokens enables verification code issuance and consumption.
func WithVerificationTokens(
	manager *VerificationTokenManager,
	repository VerificationTokenRepository,
) Option {
	return func(service *Service) {
		service.verificationTokens = manager
		service.verificationRepo = repository
	}
}

// New constructs a new users service instance with the provided repository.
func New(log *logger.Logger, repo Repository, tasksService *tasks.Service, options ...Option) *Service {
	service := &Service{
		repo:         repo,
		log:          log,
		tasksService: tasksService,
		clock:        platformclock.System{},
	}
	for _, option := range options {
		if option != nil {
			option(service)
		}
	}
	return service
}

// GetUser returns a user by ID.
func (s *Service) GetUser(ctx context.Context, userID uuid.UUID) (CoreUser, error) {
	s.log.Info(ctx, "business.core.users.GetUser")
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.users.GetUser")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID.String()))

	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return CoreUser{}, err
	}

	return user, nil
}

// ResolveActiveBrowserSessionVersion returns the monotonic first-party session
// epoch for an active account. It is deliberately narrower than GetUser so the
// authentication hot path does not hydrate profile or workspace fields.
func (s *Service) ResolveActiveBrowserSessionVersion(ctx context.Context, userID uuid.UUID) (int64, bool, error) {
	if userID == uuid.Nil {
		return 0, false, nil
	}
	return s.repo.ResolveActiveBrowserSessionVersion(ctx, userID)
}

// GetUserByEmail returns a user by email.
func (s *Service) GetUserByEmail(ctx context.Context, email string) (CoreUser, error) {
	s.log.Info(ctx, "business.core.users.GetUserByEmail")
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.users.GetUserByEmail")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", email))

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		span.RecordError(err)
		return CoreUser{}, err
	}

	return user, nil
}

// GetUserByEmailAnyStatus returns a user by email regardless of activation state.
func (s *Service) GetUserByEmailAnyStatus(ctx context.Context, email string) (CoreUser, error) {
	s.log.Info(ctx, "business.core.users.GetUserByEmailAnyStatus")
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.users.GetUserByEmailAnyStatus")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", email))

	user, err := s.repo.GetUserByEmailAnyStatus(ctx, email)
	if err != nil {
		span.RecordError(err)
		return CoreUser{}, err
	}

	return user, nil
}

// GetUsersByIDs returns users by ID regardless of membership or active status.
func (s *Service) GetUsersByIDs(ctx context.Context, userIDs []uuid.UUID) ([]CoreUser, error) {
	s.log.Info(ctx, "business.core.users.GetUsersByIDs")
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.users.GetUsersByIDs")
	defer span.End()

	if len(userIDs) == 0 {
		return []CoreUser{}, nil
	}

	users, err := s.repo.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return users, nil
}

// Register creates a new user account.
func (s *Service) Register(ctx context.Context, newUser CoreNewUser) (CoreUser, error) {
	s.log.Info(ctx, "business.core.users.Register")
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.users.Register")
	defer span.End()

	span.SetAttributes(attribute.String("user.email", newUser.Email))

	// Check if email is already taken
	_, err := s.repo.GetUserByEmailAnyStatus(ctx, newUser.Email)
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

func (s *Service) AuthenticateExternalIdentity(ctx context.Context, input CoreExternalIdentityInput) (CoreUser, error) {
	identityRepo, ok := s.repo.(ExternalIdentityRepository)
	if !ok {
		return CoreUser{}, ErrIdentityStoreMissing
	}

	result, err := identityRepo.ResolveExternalIdentity(ctx, input)
	if err != nil {
		return CoreUser{}, err
	}

	if result.Created && s.tasksService != nil {
		if _, enqueueErr := s.tasksService.EnqueueUserOnboardingStart(tasks.UserOnboardingStartPayload{
			UserID:   result.User.ID.String(),
			Email:    result.User.Email,
			FullName: result.User.FullName,
		}); enqueueErr != nil {
			s.log.Error(ctx, "error enqueuing external identity user onboarding", "error", enqueueErr, "user_id", result.User.ID)
		}
	}

	return result.User, nil
}

// ReactivateUserForVerifiedSignIn reactivates only an account whose durable
// policy permits a verified authentication to do so. Administrator-disabled
// and conservatively backfilled legacy accounts fail with the same generic
// invalid-credentials error used by authentication endpoints.
func (s *Service) ReactivateUserForVerifiedSignIn(ctx context.Context, userID uuid.UUID) (CoreUser, error) {
	s.log.Info(ctx, "business.core.users.ReactivateUserForVerifiedSignIn")
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.users.ReactivateUserForVerifiedSignIn")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID.String()))

	user, err := s.repo.ReactivateUserForVerifiedSignIn(ctx, usersdomain.VerifiedSignInReactivation{
		UserID: userID, SignedInAt: s.clock.Now().UTC(),
	})
	if err != nil {
		span.RecordError(err)
		return CoreUser{}, err
	}

	return user, nil
}

// UpdateUser updates a user's profile.
func (s *Service) UpdateUser(ctx context.Context, userID uuid.UUID, updates CoreUpdateUser) error {
	s.log.Info(ctx, "business.core.users.UpdateUser")
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.users.UpdateUser")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID.String()))
	if updates.WorkSchedule != nil {
		if len(updates.WorkSchedule.WorkingDays) > 0 {
			if err := workschedule.ValidateWorkingDays(updates.WorkSchedule.WorkingDays); err != nil {
				return err
			}
		}
		hasStart := updates.WorkSchedule.WorkingStartMinute != nil
		hasEnd := updates.WorkSchedule.WorkingEndMinute != nil
		if hasStart != hasEnd {
			return workschedule.ErrInvalidWorkingHours
		}
		if hasStart {
			if err := workschedule.ValidateHours(*updates.WorkSchedule.WorkingStartMinute, *updates.WorkSchedule.WorkingEndMinute); err != nil {
				return err
			}
		}
	}

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
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.users.DeleteUser")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID.String()))

	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	if err := s.repo.DeleteUser(ctx, userID, s.clock.Now().UTC()); err != nil {
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
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.users.UpdateUserWorkspace")
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
func (s *Service) List(ctx context.Context, workspaceID uuid.UUID, filter CoreListUsersFilter) ([]CoreUser, error) {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.users.List")
	defer span.End()

	users, err := s.repo.List(ctx, workspaceID, filter)
	if err != nil {
		return nil, fmt.Errorf("getting users list: %w", err)
	}

	return users, nil
}

// CreateVerificationToken creates a new verification token.
func (s *Service) CreateVerificationToken(
	ctx context.Context,
	email string,
	tokenType string,
	expiresAt time.Time,
) (CoreVerificationToken, error) {
	s.log.Info(ctx, "business.core.users.CreateVerificationToken")
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.users.CreateVerificationToken")
	defer span.End()

	if s.verificationTokens == nil || s.verificationRepo == nil {
		span.RecordError(ErrVerificationDisabled)
		return CoreVerificationToken{}, ErrVerificationDisabled
	}

	for range verificationTokenCollisionRetries {
		plaintext, newToken, err := s.verificationTokens.issue(email, tokenType, expiresAt)
		if err != nil {
			span.RecordError(err)
			return CoreVerificationToken{}, err
		}

		created, err := s.verificationRepo.CreateVerificationToken(ctx, newToken)
		if errors.Is(err, ErrTokenCollision) {
			continue
		}
		if err != nil {
			span.RecordError(err)
			return CoreVerificationToken{}, err
		}
		created.Token = plaintext
		return created, nil
	}

	span.RecordError(ErrTokenCollision)
	return CoreVerificationToken{}, ErrTokenCollision
}

// ConsumeVerificationToken atomically validates and consumes one code. The
// approved purposes are explicit so a digest minted for one flow cannot be
// replayed in another flow.
func (s *Service) ConsumeVerificationToken(
	ctx context.Context,
	email string,
	token string,
	tokenTypes ...string,
) (CoreVerificationToken, error) {
	s.log.Info(ctx, "business.core.users.ConsumeVerificationToken")
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.users.ConsumeVerificationToken")
	defer span.End()

	if s.verificationTokens == nil || s.verificationRepo == nil {
		span.RecordError(ErrVerificationDisabled)
		return CoreVerificationToken{}, ErrVerificationDisabled
	}

	input, err := s.verificationTokens.consumption(email, token, tokenTypes...)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) {
			return CoreVerificationToken{}, ErrInvalidToken
		}
		span.RecordError(err)
		return CoreVerificationToken{}, err
	}

	verificationToken, err := s.verificationRepo.ConsumeVerificationToken(ctx, input)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) {
			return CoreVerificationToken{}, ErrInvalidToken
		}
		span.RecordError(err)
		return CoreVerificationToken{}, err
	}

	return verificationToken, nil
}

// VerificationRateLimitKey returns an opaque key for public-auth abuse
// controls without exposing the underlying identity in cache keyspace.
func (s *Service) VerificationRateLimitKey(scope, identityType, identity string) (string, error) {
	if s.verificationTokens == nil {
		return "", ErrVerificationDisabled
	}
	return s.verificationTokens.RateLimitKey(scope, identityType, identity)
}

// InvalidateTokens invalidates all unused tokens for an email.
func (s *Service) InvalidateTokens(ctx context.Context, email string) error {
	s.log.Info(ctx, "business.core.users.InvalidateTokens")
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.users.InvalidateTokens")
	defer span.End()

	normalizedEmail, err := validate.Email(email)
	if err != nil {
		return err
	}

	if s.verificationRepo == nil {
		return ErrVerificationDisabled
	}

	if err := s.verificationRepo.InvalidateTokens(ctx, normalizedEmail); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

// GetAutomationPreferences retrieves a user's automation preferences for a workspace.
func (s *Service) GetAutomationPreferences(ctx context.Context, userID, workspaceID uuid.UUID) (CoreAutomationPreferences, error) {
	s.log.Info(ctx, "business.core.users.GetAutomationPreferences")
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.users.GetAutomationPreferences")
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
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.users.UpdateAutomationPreferences")
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
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.users.UploadProfileImage")
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
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.users.DeleteProfileImage")
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
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.users.AddUserMemory")
	defer span.End()

	newItem, err := s.repo.AddUserMemory(ctx, memory)
	if err != nil {
		return CoreUserMemoryItem{}, fmt.Errorf("adding user memory: %w", err)
	}

	return newItem, nil
}

// UpdateUserMemory updates a memory item.
func (s *Service) UpdateUserMemory(ctx context.Context, id uuid.UUID, scope UserMemoryScope, update UpdateUserMemoryItem) error {
	s.log.Info(ctx, "business.core.users.UpdateUserMemory")
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.users.UpdateUserMemory")
	defer span.End()

	if err := s.repo.UpdateUserMemory(ctx, id, scope, update); err != nil {
		return fmt.Errorf("updating user memory: %w", err)
	}

	return nil
}

// DeleteUserMemory deletes a memory item.
func (s *Service) DeleteUserMemory(ctx context.Context, id uuid.UUID, scope UserMemoryScope) error {
	s.log.Info(ctx, "business.core.users.DeleteUserMemory")
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.users.DeleteUserMemory")
	defer span.End()

	if err := s.repo.DeleteUserMemory(ctx, id, scope); err != nil {
		return fmt.Errorf("deleting user memory: %w", err)
	}

	return nil
}

// ListUserMemories retrieves all memory items for a user in a workspace.
func (s *Service) ListUserMemories(ctx context.Context, userID, workspaceID uuid.UUID) ([]CoreUserMemoryItem, error) {
	s.log.Info(ctx, "business.core.users.ListUserMemories")
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.core.users.ListUserMemories")
	defer span.End()

	items, err := s.repo.ListUserMemories(ctx, userID, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("listing user memories: %w", err)
	}

	return items, nil
}
