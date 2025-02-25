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
	// Email invitations
	CreateInvitation(ctx context.Context, invitation CoreWorkspaceInvitation) (CoreWorkspaceInvitation, error)
	GetInvitation(ctx context.Context, token string) (CoreWorkspaceInvitation, error)
	ListInvitations(ctx context.Context, workspaceID uuid.UUID) ([]CoreWorkspaceInvitation, error)
	RevokeInvitation(ctx context.Context, invitationID uuid.UUID) error
	MarkInvitationUsed(ctx context.Context, invitationID uuid.UUID) error

	// Invitation links
	CreateInvitationLink(ctx context.Context, link CoreWorkspaceInvitationLink) (CoreWorkspaceInvitationLink, error)
	GetInvitationLink(ctx context.Context, token string) (CoreWorkspaceInvitationLink, error)
	ListInvitationLinks(ctx context.Context, workspaceID uuid.UUID) ([]CoreWorkspaceInvitationLink, error)
	RevokeInvitationLink(ctx context.Context, linkID uuid.UUID) error
	IncrementLinkUsage(ctx context.Context, linkID uuid.UUID) error

	// Transaction support
	BeginTx(ctx context.Context) (*sql.Tx, error)
}

// Service provides invitation operations
type Service struct {
	repo Repository
	log  *logger.Logger
}

// New constructs a new invitations service instance
func New(log *logger.Logger, repo Repository) *Service {
	return &Service{
		repo: repo,
		log:  log,
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

// CreateInvitation creates a new workspace invitation
func (s *Service) CreateInvitation(ctx context.Context, workspaceID, inviterID uuid.UUID, email, role string, teamIDs []uuid.UUID) (CoreWorkspaceInvitation, error) {
	s.log.Info(ctx, "business.core.invitations.CreateInvitation")
	ctx, span := web.AddSpan(ctx, "business.core.invitations.CreateInvitation")
	defer span.End()

	token, err := generateToken()
	if err != nil {
		return CoreWorkspaceInvitation{}, err
	}

	invitation := CoreWorkspaceInvitation{
		WorkspaceID: workspaceID,
		InviterID:   inviterID,
		Email:       email,
		Role:        role,
		Token:       token,
		TeamIDs:     teamIDs,
		ExpiresAt:   time.Now().Add(24 * time.Hour), // 24 hour expiry
	}

	result, err := s.repo.CreateInvitation(ctx, invitation)
	if err != nil {
		span.RecordError(err)
		return CoreWorkspaceInvitation{}, err
	}

	span.AddEvent("invitation created", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.String("inviter_id", inviterID.String()),
		attribute.String("email", email),
	))

	return result, nil
}

// GetInvitation retrieves a workspace invitation by token
func (s *Service) GetInvitation(ctx context.Context, token string) (CoreWorkspaceInvitation, error) {
	s.log.Info(ctx, "business.core.invitations.GetInvitation")
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
	s.log.Info(ctx, "business.core.invitations.ListInvitations")
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
	s.log.Info(ctx, "business.core.invitations.RevokeInvitation")
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

// CreateInvitationLink creates a new workspace invitation link
func (s *Service) CreateInvitationLink(ctx context.Context, workspaceID, creatorID uuid.UUID, role string, teamIDs []uuid.UUID, maxUses *int, expiresAt *time.Time) (CoreWorkspaceInvitationLink, error) {
	s.log.Info(ctx, "business.core.invitations.CreateInvitationLink")
	ctx, span := web.AddSpan(ctx, "business.core.invitations.CreateInvitationLink")
	defer span.End()

	token, err := generateToken()
	if err != nil {
		return CoreWorkspaceInvitationLink{}, err
	}

	link := CoreWorkspaceInvitationLink{
		WorkspaceID: workspaceID,
		CreatorID:   creatorID,
		Token:       token,
		Role:        role,
		TeamIDs:     teamIDs,
		MaxUses:     maxUses,
		ExpiresAt:   expiresAt,
		IsActive:    true,
	}

	result, err := s.repo.CreateInvitationLink(ctx, link)
	if err != nil {
		span.RecordError(err)
		return CoreWorkspaceInvitationLink{}, err
	}

	span.AddEvent("invitation link created", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.String("creator_id", creatorID.String()),
	))

	return result, nil
}

// GetInvitationLink retrieves a workspace invitation link by token
func (s *Service) GetInvitationLink(ctx context.Context, token string) (CoreWorkspaceInvitationLink, error) {
	s.log.Info(ctx, "business.core.invitations.GetInvitationLink")
	ctx, span := web.AddSpan(ctx, "business.core.invitations.GetInvitationLink")
	defer span.End()

	link, err := s.repo.GetInvitationLink(ctx, token)
	if err != nil {
		span.RecordError(err)
		return CoreWorkspaceInvitationLink{}, err
	}

	// Check if link is active
	if !link.IsActive {
		return CoreWorkspaceInvitationLink{}, ErrInvitationRevoked
	}

	// Check if link has expired
	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		return CoreWorkspaceInvitationLink{}, ErrInvitationExpired
	}

	// Check if max uses reached
	if link.MaxUses != nil && link.UsedCount >= *link.MaxUses {
		return CoreWorkspaceInvitationLink{}, ErrMaxUsesReached
	}

	return link, nil
}

// ListInvitationLinks returns all active invitation links for a workspace
func (s *Service) ListInvitationLinks(ctx context.Context, workspaceID uuid.UUID) ([]CoreWorkspaceInvitationLink, error) {
	s.log.Info(ctx, "business.core.invitations.ListInvitationLinks")
	ctx, span := web.AddSpan(ctx, "business.core.invitations.ListInvitationLinks")
	defer span.End()

	links, err := s.repo.ListInvitationLinks(ctx, workspaceID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return links, nil
}

// RevokeInvitationLink revokes a workspace invitation link
func (s *Service) RevokeInvitationLink(ctx context.Context, linkID uuid.UUID) error {
	s.log.Info(ctx, "business.core.invitations.RevokeInvitationLink")
	ctx, span := web.AddSpan(ctx, "business.core.invitations.RevokeInvitationLink")
	defer span.End()

	if err := s.repo.RevokeInvitationLink(ctx, linkID); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("invitation link revoked", trace.WithAttributes(
		attribute.String("link_id", linkID.String()),
	))

	return nil
}

// UseInvitationLink increments the usage count for an invitation link
func (s *Service) UseInvitationLink(ctx context.Context, linkID uuid.UUID) error {
	s.log.Info(ctx, "business.core.invitations.UseInvitationLink")
	ctx, span := web.AddSpan(ctx, "business.core.invitations.UseInvitationLink")
	defer span.End()

	if err := s.repo.IncrementLinkUsage(ctx, linkID); err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("invitation link used", trace.WithAttributes(
		attribute.String("link_id", linkID.String()),
	))

	return nil
}

// CreateBulkInvitations creates multiple workspace invitations
func (s *Service) CreateBulkInvitations(ctx context.Context, workspaceID, inviterID uuid.UUID, requests []InvitationRequest) ([]CoreWorkspaceInvitation, error) {
	s.log.Info(ctx, "business.core.invitations.CreateBulkInvitations")
	ctx, span := web.AddSpan(ctx, "business.core.invitations.CreateBulkInvitations")
	defer span.End()

	invitations := make([]CoreWorkspaceInvitation, 0, len(requests))

	for _, req := range requests {
		token, err := generateToken()
		if err != nil {
			return nil, err
		}

		invitation := CoreWorkspaceInvitation{
			WorkspaceID: workspaceID,
			InviterID:   inviterID,
			Email:       req.Email,
			Role:        req.Role,
			Token:       token,
			TeamIDs:     req.TeamIDs,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
		}

		result, err := s.repo.CreateInvitation(ctx, invitation)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}

		invitations = append(invitations, result)
	}

	return invitations, nil
}
