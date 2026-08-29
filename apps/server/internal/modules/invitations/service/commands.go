package invitations

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// CreateBulkInvitations persists invitations and token-free email outbox rows
// together. It performs no network I/O inside or after the transaction.
func (s *Service) CreateBulkInvitations(
	ctx context.Context,
	workspaceID uuid.UUID,
	inviterID uuid.UUID,
	requests []InvitationRequest,
) ([]CoreWorkspaceInvitation, error) {
	ctx, span := startInvitationSpan(ctx, "business.core.invitations.CreateBulkInvitations")
	defer span.End()
	span.SetAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.String("inviter_id", inviterID.String()),
		attribute.Int("invitation_count", len(requests)),
	)

	if len(requests) > maximumBulkInvitations {
		return nil, ErrTooManyInvitations
	}
	workspace, err := s.requireWorkspaceAdmin(ctx, workspaceID, inviterID)
	if err != nil {
		return nil, err
	}
	normalized, err := normalizeInvitationRequests(requests)
	if err != nil {
		return nil, err
	}
	if len(normalized) == 0 {
		return []CoreWorkspaceInvitation{}, nil
	}
	if s.tokens == nil {
		return nil, errors.New("invitation token security is not configured")
	}

	inviter, err := s.users.GetUser(ctx, inviterID)
	if err != nil {
		return nil, fmt.Errorf("get invitation sender: %w", err)
	}
	inviterName := strings.TrimSpace(inviter.FullName)
	if inviterName == "" {
		inviterName = inviter.Username
	}

	now := s.now().UTC()
	expiresAt := now.Add(invitationLifetime)
	commands := make([]NewWorkspaceInvitation, 0, len(normalized))
	for _, request := range normalized {
		_, storedToken, err := s.tokens.Issue()
		if err != nil {
			return nil, err
		}
		invitation := CoreWorkspaceInvitation{
			ID:          uuid.New(),
			WorkspaceID: workspaceID,
			InviterID:   inviterID,
			Email:       request.Email,
			Role:        request.Role,
			TeamIDs:     request.TeamIDs,
			ExpiresAt:   expiresAt,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		commands = append(commands, NewWorkspaceInvitation{
			Invitation: invitation,
			Token:      storedToken,
			EmailOutbox: InvitationEmailOutboxPayload{
				InviterName:   inviterName,
				Email:         invitation.Email,
				Role:          invitation.Role,
				ExpiresAt:     invitation.ExpiresAt,
				WorkspaceID:   workspaceID,
				WorkspaceName: workspace.Name,
			},
		})
	}

	return s.repo.CreateBulkInvitations(ctx, inviterID, commands)
}

func normalizeInvitationRequests(requests []InvitationRequest) ([]InvitationRequest, error) {
	result := make([]InvitationRequest, 0, len(requests))
	seenEmails := make(map[string]struct{}, len(requests))
	for _, request := range requests {
		if err := ValidateInvitationRole(request.Role); err != nil {
			return nil, err
		}
		email, err := validate.Email(request.Email)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidInvitationEmail, err)
		}
		if _, exists := seenEmails[email]; exists {
			return nil, ErrDuplicateInvitation
		}
		seenEmails[email] = struct{}{}

		teamIDs := make([]uuid.UUID, 0, len(request.TeamIDs))
		seenTeams := make(map[uuid.UUID]struct{}, len(request.TeamIDs))
		for _, teamID := range request.TeamIDs {
			if teamID == uuid.Nil {
				return nil, ErrInvalidInvitationTeam
			}
			if _, exists := seenTeams[teamID]; exists {
				continue
			}
			seenTeams[teamID] = struct{}{}
			teamIDs = append(teamIDs, teamID)
		}
		result = append(result, InvitationRequest{Email: email, Role: request.Role, TeamIDs: teamIDs})
	}
	return result, nil
}

func (s *Service) RevokeInvitation(
	ctx context.Context,
	workspaceID uuid.UUID,
	actorID uuid.UUID,
	invitationID uuid.UUID,
) error {
	ctx, span := startInvitationSpan(ctx, "business.core.invitations.RevokeInvitation")
	defer span.End()
	if _, err := s.requireWorkspaceAdmin(ctx, workspaceID, actorID); err != nil {
		span.RecordError(err)
		return err
	}
	if err := s.repo.RevokeInvitation(ctx, workspaceID, actorID, invitationID, s.now().UTC()); err != nil {
		span.RecordError(err)
		return err
	}
	span.AddEvent("invitation revoked", trace.WithAttributes(attribute.String("invitation_id", invitationID.String())))
	return nil
}

// AcceptInvitation delegates the complete single-use invariant to one
// repository-owned transaction after local token-format verification.
func (s *Service) AcceptInvitation(ctx context.Context, rawToken string, userID uuid.UUID) error {
	ctx, span := startInvitationSpan(ctx, "business.core.invitations.AcceptInvitation")
	defer span.End()
	span.SetAttributes(attribute.String("user_id", userID.String()))

	if s.tokens == nil {
		return errors.New("invitation token security is not configured")
	}
	lookup, err := s.tokens.Lookup(rawToken)
	if err != nil {
		return ErrInvitationNotFound
	}
	_, err = s.repo.AcceptInvitation(ctx, AcceptInvitationCommand{
		Lookup:     lookup,
		UserID:     userID,
		AcceptedAt: s.now().UTC(),
	})
	return err
}
