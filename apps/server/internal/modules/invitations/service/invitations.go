package invitations

import (
	"context"
	"errors"
	"fmt"
	"time"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	workspacedomain "github.com/complexus-tech/projects-api/internal/modules/workspaces/domain"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
)

const (
	invitationLifetime        = 7 * 24 * time.Hour
	maximumBulkInvitations    = 50
	invitationOutboxBatchSize = 50
	invitationOutboxMaxTries  = 8
	invitationOutboxStaleFor  = 10 * time.Minute
)

// Repository owns invitation persistence and the cohesive acceptance unit of
// work. Raw transactions and sqlc types never cross this consumer-owned port.
type Repository interface {
	CreateBulkInvitations(context.Context, uuid.UUID, []NewWorkspaceInvitation) ([]CoreWorkspaceInvitation, error)
	GetInvitation(context.Context, InvitationTokenLookup) (CoreWorkspaceInvitation, error)
	ListInvitations(context.Context, uuid.UUID, uuid.UUID, time.Time) ([]CoreWorkspaceInvitation, error)
	RevokeInvitation(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, time.Time) error
	ListInvitationsByEmail(context.Context, string, time.Time) ([]CoreWorkspaceInvitation, error)
	AcceptInvitation(context.Context, AcceptInvitationCommand) (CoreWorkspaceInvitation, error)
}

type InvitationOutboxRepository interface {
	ClaimInvitationOutboxEvents(context.Context, int, time.Time, time.Time) ([]CoreInvitationOutboxEvent, error)
	CompleteInvitationOutboxEvent(context.Context, uuid.UUID, uuid.UUID, time.Time) error
	RetryInvitationOutboxEvent(context.Context, uuid.UUID, uuid.UUID, string, time.Time, time.Time, bool) error
}

// UserService is the read-only user boundary needed to snapshot inviter data.
type UserService interface {
	GetUser(context.Context, uuid.UUID) (usersdomain.User, error)
}

// WorkspaceService is the read-only workspace and authorization boundary.
type WorkspaceService interface {
	Get(context.Context, uuid.UUID, uuid.UUID) (workspacedomain.Workspace, error)
}

// Service provides invitation business use cases. The clock and token manager
// are explicit dependencies so expiry and security behavior are deterministic.
type Service struct {
	repo       Repository
	tokens     *InvitationTokenManager
	users      UserService
	workspaces WorkspaceService
	now        func() time.Time
}

func New(
	repo Repository,
	tokens *InvitationTokenManager,
	users UserService,
	workspaces WorkspaceService,
) *Service {
	return &Service{
		repo:       repo,
		tokens:     tokens,
		users:      users,
		workspaces: workspaces,
		now:        time.Now,
	}
}

func (s *Service) requireWorkspaceAdmin(
	ctx context.Context,
	workspaceID uuid.UUID,
	actorID uuid.UUID,
) (workspacedomain.Workspace, error) {
	workspace, err := s.workspaces.Get(ctx, workspaceID, actorID)
	if err != nil {
		if errors.Is(err, workspacedomain.ErrNotFound) {
			return workspacedomain.Workspace{}, authorization.ErrWorkspaceAdminRequired
		}
		return workspacedomain.Workspace{}, fmt.Errorf("authorize workspace administrator: %w", err)
	}
	if err := authorization.RequireWorkspaceAdmin(authorization.WorkspaceRole(workspace.UserRole)); err != nil {
		return workspacedomain.Workspace{}, err
	}
	return workspace, nil
}
