package invitations

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
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
}

// Service provides invitation operations
type Service struct {
	repo   Repository
	logger *logger.Logger
}

// New constructs a new invitations service instance
func New(repo Repository, logger *logger.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
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
