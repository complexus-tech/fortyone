package invitations

import (
	"context"
	"errors"

	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/google/uuid"
)

func (s *Service) GetInvitation(ctx context.Context, rawToken string) (CoreWorkspaceInvitation, error) {
	ctx, span := startInvitationSpan(ctx, "business.core.invitations.GetInvitation")
	defer span.End()
	if s.tokens == nil {
		return CoreWorkspaceInvitation{}, errors.New("invitation token security is not configured")
	}
	lookup, err := s.tokens.Lookup(rawToken)
	if err != nil {
		return CoreWorkspaceInvitation{}, ErrInvitationNotFound
	}
	invitation, err := s.repo.GetInvitation(ctx, lookup)
	if err != nil {
		return CoreWorkspaceInvitation{}, err
	}
	now := s.now().UTC()
	if !invitation.ExpiresAt.After(now) {
		return CoreWorkspaceInvitation{}, ErrInvitationExpired
	}
	if invitation.UsedAt != nil {
		return CoreWorkspaceInvitation{}, ErrInvitationUsed
	}
	return invitation, nil
}

func (s *Service) ListInvitations(
	ctx context.Context,
	workspaceID uuid.UUID,
	actorID uuid.UUID,
) ([]CoreWorkspaceInvitation, error) {
	ctx, span := startInvitationSpan(ctx, "business.core.invitations.ListInvitations")
	defer span.End()
	if _, err := s.requireWorkspaceAdmin(ctx, workspaceID, actorID); err != nil {
		return nil, err
	}
	return s.repo.ListInvitations(ctx, workspaceID, actorID, s.now().UTC())
}

func (s *Service) ListUserInvitations(ctx context.Context, email string) ([]CoreWorkspaceInvitation, error) {
	normalizedEmail, err := validate.Email(email)
	if err != nil {
		return nil, err
	}
	return s.repo.ListInvitationsByEmail(ctx, normalizedEmail, s.now().UTC())
}
