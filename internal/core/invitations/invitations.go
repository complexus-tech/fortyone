package invitations

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/teams"
	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/internal/core/workspaces"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Repository provides access to the invitations storage
type Repository interface {
	// Transaction support
	BeginTx(ctx context.Context) (*sql.Tx, error)
	CreateBulkInvitations(ctx context.Context, tx *sql.Tx, invitations []CoreWorkspaceInvitation) ([]CoreWorkspaceInvitation, error)
	GetInvitation(ctx context.Context, token string) (CoreWorkspaceInvitation, error)
	ListInvitations(ctx context.Context, workspaceID uuid.UUID) ([]CoreWorkspaceInvitation, error)
	RevokeInvitation(ctx context.Context, invitationID uuid.UUID) error
	MarkInvitationUsed(ctx context.Context, invitationID uuid.UUID) error
	ListInvitationsByEmail(ctx context.Context, email string) ([]CoreWorkspaceInvitation, error)
}

// Service provides invitation operations
type Service struct {
	repo       Repository
	logger     *logger.Logger
	publisher  *publisher.Publisher
	users      *users.Service
	workspaces *workspaces.Service
	teams      *teams.Service
}

// New constructs a new invitations service instance
func New(repo Repository, logger *logger.Logger, publisher *publisher.Publisher, users *users.Service, workspaces *workspaces.Service, teams *teams.Service) *Service {
	return &Service{
		repo:       repo,
		logger:     logger,
		publisher:  publisher,
		users:      users,
		workspaces: workspaces,
		teams:      teams,
	}
}

// generateToken creates a cryptographically secure token
func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// CreateBulkInvitations creates multiple workspace invitations
func (s *Service) CreateBulkInvitations(ctx context.Context, workspaceID, inviterID uuid.UUID, requests []InvitationRequest) ([]CoreWorkspaceInvitation, error) {
	s.logger.Info(ctx, "creating bulk invitations",
		"workspace_id", workspaceID.String(),
		"inviter_id", inviterID.String(),
		"invitation_count", len(requests))

	// Get inviter details
	inviter, err := s.users.GetUser(ctx, inviterID)
	if err != nil {
		s.logger.Error(ctx, "failed to get inviter details", "err", err)
		return nil, fmt.Errorf("failed to get inviter details: %w", err)
	}

	// Get workspace details
	workspace, err := s.workspaces.Get(ctx, workspaceID, inviterID)
	if err != nil {
		s.logger.Error(ctx, "failed to get workspace details", "err", err)
		return nil, fmt.Errorf("failed to get workspace details: %w", err)
	}

	// Start a transaction
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		s.logger.Error(ctx, "failed to begin transaction", "err", err)
		return nil, err
	}
	defer tx.Rollback()

	// Create invitations
	invitations := make([]CoreWorkspaceInvitation, len(requests))
	for i, req := range requests {
		token, err := generateToken()
		if err != nil {
			s.logger.Error(ctx, "failed to generate token", "err", err)
			return nil, err
		}

		invitations[i] = CoreWorkspaceInvitation{
			WorkspaceID: workspaceID,
			InviterID:   inviterID,
			Email:       req.Email,
			Role:        req.Role,
			TeamIDs:     req.TeamIDs,
			Token:       token,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}

	// Save invitations
	results, err := s.repo.CreateBulkInvitations(ctx, tx, invitations)
	if err != nil {
		s.logger.Error(ctx, "failed to create bulk invitations", "err", err)
		return nil, err
	}

	// Publish invitation email events
	for _, invitation := range results {
		event := events.Event{
			Type: events.InvitationEmail,
			Payload: events.InvitationEmailPayload{
				InviterName:   inviter.FullName,
				Email:         invitation.Email,
				Token:         invitation.Token,
				Role:          invitation.Role,
				ExpiresAt:     invitation.ExpiresAt,
				WorkspaceID:   invitation.WorkspaceID,
				WorkspaceName: workspace.Name,
			},
			Timestamp: time.Now(),
			ActorID:   inviterID,
		}

		if err := s.publisher.Publish(ctx, event); err != nil {
			s.logger.Error(ctx, "failed to publish invitation email event",
				"error", err,
				"email", invitation.Email)
			return nil, fmt.Errorf("failed to publish invitation email event: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		s.logger.Error(ctx, "failed to commit transaction", "err", err)
		return nil, err
	}

	s.logger.Info(ctx, "bulk invitations created successfully")
	return results, nil
}

// GetInvitation retrieves a workspace invitation by token
func (s *Service) GetInvitation(ctx context.Context, token string) (CoreWorkspaceInvitation, error) {
	s.logger.Info(ctx, "business.core.invitations.GetInvitation")
	ctx, span := web.AddSpan(ctx, "business.core.invitations.GetInvitation")
	defer span.End()

	invitation, err := s.repo.GetInvitation(ctx, token)
	if err != nil {
		span.RecordError(err)
		return CoreWorkspaceInvitation{}, err
	}

	// Check if invitation has expired
	if time.Now().After(invitation.ExpiresAt) {
		return CoreWorkspaceInvitation{}, ErrInvitationExpired
	}

	// Check if invitation has been used
	if invitation.UsedAt != nil {
		return CoreWorkspaceInvitation{}, ErrInvitationUsed
	}

	return invitation, nil
}

// ListInvitations returns all pending invitations for a workspace
func (s *Service) ListInvitations(ctx context.Context, workspaceID uuid.UUID) ([]CoreWorkspaceInvitation, error) {
	s.logger.Info(ctx, "business.core.invitations.ListInvitations")
	ctx, span := web.AddSpan(ctx, "business.core.invitations.ListInvitations")
	defer span.End()

	invitations, err := s.repo.ListInvitations(ctx, workspaceID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return invitations, nil
}

// RevokeInvitation revokes a workspace invitation
func (s *Service) RevokeInvitation(ctx context.Context, invitationID uuid.UUID) error {
	s.logger.Info(ctx, "business.core.invitations.RevokeInvitation")
	ctx, span := web.AddSpan(ctx, "business.core.invitations.RevokeInvitation")
	defer span.End()

	if err := s.repo.RevokeInvitation(ctx, invitationID); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("invitation revoked", trace.WithAttributes(
		attribute.String("invitation_id", invitationID.String()),
	))

	return nil
}

// ListUserInvitations returns all pending invitations for a user's email
func (s *Service) ListUserInvitations(ctx context.Context, email string) ([]CoreWorkspaceInvitation, error) {
	s.logger.Info(ctx, "listing user invitations", "email", email)

	invitations, err := s.repo.ListInvitationsByEmail(ctx, email)
	if err != nil {
		s.logger.Error(ctx, "failed to list user invitations", "err", err)
		return nil, fmt.Errorf("failed to list user invitations: %w", err)
	}

	return invitations, nil
}

// AcceptInvitation accepts an invitation and adds the user to the workspace and teams
func (s *Service) AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) error {
	s.logger.Info(ctx, "accepting invitation", "token", token, "user_id", userID)

	// Get user details to verify email
	user, err := s.users.GetUser(ctx, userID)
	if err != nil {
		s.logger.Error(ctx, "failed to get user details", "err", err)
		return fmt.Errorf("failed to get user details: %w", err)
	}

	// Get and validate invitation
	invitation, err := s.GetInvitation(ctx, token)
	if err != nil {
		s.logger.Error(ctx, "failed to get invitation", "err", err)
		return err
	}

	// Verify user email matches invitation email
	if user.Email != invitation.Email {
		s.logger.Error(ctx, "user email does not match invitation email",
			"user_email", user.Email,
			"invitation_email", invitation.Email)
		return ErrInvalidInvitee
	}

	// Start transaction
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		s.logger.Error(ctx, "failed to begin transaction", "err", err)
		return err
	}
	defer tx.Rollback()

	// Add member to workspace with specified role
	if err := s.workspaces.AddMember(ctx, invitation.WorkspaceID, userID, invitation.Role); err != nil {
		s.logger.Error(ctx, "failed to add member to workspace", "err", err)
		if err == workspaces.ErrAlreadyWorkspaceMember {
			// Mark invitation as used if the user is already a member
			if markErr := s.repo.MarkInvitationUsed(ctx, invitation.ID); markErr != nil {
				s.logger.Error(ctx, "failed to mark invitation as used", "err", markErr)
			}
			return ErrAlreadyWorkspaceMember
		}
		return fmt.Errorf("failed to add member to workspace: %w", err)
	}

	// Add member to teams
	for _, teamID := range invitation.TeamIDs {
		if err := s.teams.AddMember(ctx, teamID, userID); err != nil {
			s.logger.Error(ctx, "failed to add member to team", "err", err)
			return fmt.Errorf("failed to add member to team: %w", err)
		}
	}

	// Mark invitation as used
	if err := s.repo.MarkInvitationUsed(ctx, invitation.ID); err != nil {
		s.logger.Error(ctx, "failed to mark invitation as used", "err", err)
		return fmt.Errorf("failed to mark invitation as used: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		s.logger.Error(ctx, "failed to commit transaction", "err", err)
		return err
	}

	// Get inviter details
	inviter, err := s.users.GetUser(ctx, invitation.InviterID)
	if err != nil {
		s.logger.Error(ctx, "failed to get inviter details", "err", err)
		// Don't return error here, as the invitation was successfully accepted
	} else {
		// Get workspace details
		workspace, err := s.workspaces.Get(ctx, invitation.WorkspaceID, invitation.InviterID)
		if err != nil {
			s.logger.Error(ctx, "failed to get workspace details", "err", err)
			// Don't return error here, as the invitation was successfully accepted
		} else {
			// Use username as fallback if full name is empty
			inviterName := inviter.FullName
			if inviterName == "" {
				inviterName = inviter.Username
			}

			inviteeName := user.FullName
			if inviteeName == "" {
				inviteeName = user.Username
			}

			// Publish invitation accepted event
			event := events.Event{
				Type: events.InvitationAccepted,
				Payload: events.InvitationAcceptedPayload{
					InviterEmail:  inviter.Email,
					InviterName:   inviterName,
					InviteeName:   inviteeName,
					InviteeEmail:  user.Email,
					Role:          invitation.Role,
					WorkspaceID:   invitation.WorkspaceID,
					WorkspaceName: workspace.Name,
					WorkspaceSlug: workspace.Slug,
				},
				Timestamp: time.Now(),
				ActorID:   userID,
			}

			if err := s.publisher.Publish(ctx, event); err != nil {
				s.logger.Error(ctx, "failed to publish invitation accepted event", "err", err)
				// Don't return error here, as the invitation was successfully accepted
			}
		}
	}

	s.logger.Info(ctx, "invitation accepted successfully")
	return nil
}
