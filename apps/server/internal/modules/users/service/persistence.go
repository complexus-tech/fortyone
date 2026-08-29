package users

import (
	"context"
	"time"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	"github.com/google/uuid"
)

// ExternalIdentityRepository is the persistence port for OAuth identities.
type ExternalIdentityRepository interface {
	ResolveExternalIdentity(ctx context.Context, input CoreExternalIdentityInput) (CoreExternalIdentityResult, error)
}

// Repository provides access to user accounts and their private capabilities.
type Repository interface {
	ResolveActiveBrowserSessionVersion(ctx context.Context, userID uuid.UUID) (int64, bool, error)
	GetUser(ctx context.Context, userID uuid.UUID) (CoreUser, error)
	GetUserByEmail(ctx context.Context, email string) (CoreUser, error)
	GetUserByEmailAnyStatus(ctx context.Context, email string) (CoreUser, error)
	GetUsersByIDs(ctx context.Context, userIDs []uuid.UUID) ([]CoreUser, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, updates CoreUpdateUser) (CoreUser, error)
	ReactivateUserForVerifiedSignIn(ctx context.Context, input usersdomain.VerifiedSignInReactivation) (CoreUser, error)
	DeleteUser(ctx context.Context, userID uuid.UUID, deactivatedAt time.Time) error
	UpdateUserWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) error
	List(ctx context.Context, workspaceID uuid.UUID, filter CoreListUsersFilter) ([]CoreUser, error)
	Create(ctx context.Context, user CoreUser) (CoreUser, error)
	GetAutomationPreferences(ctx context.Context, userID, workspaceID uuid.UUID) (CoreAutomationPreferences, error)
	UpdateAutomationPreferences(ctx context.Context, userID, workspaceID uuid.UUID, updates CoreUpdateAutomationPreferences) error
	AddUserMemory(ctx context.Context, memory NewUserMemoryItem) (CoreUserMemoryItem, error)
	UpdateUserMemory(ctx context.Context, id uuid.UUID, scope UserMemoryScope, update UpdateUserMemoryItem) error
	DeleteUserMemory(ctx context.Context, id uuid.UUID, scope UserMemoryScope) error
	ListUserMemories(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID) ([]CoreUserMemoryItem, error)
}

// VerificationTokenRepository is the narrow persistence port for verification
// codes. Keeping the capability explicit prevents authentication callers from
// depending on the broader account repository.
type VerificationTokenRepository interface {
	CreateVerificationToken(ctx context.Context, token NewVerificationToken) (CoreVerificationToken, error)
	ConsumeVerificationToken(ctx context.Context, input ConsumeVerificationTokenInput) (CoreVerificationToken, error)
	InvalidateTokens(ctx context.Context, email string) error
}
