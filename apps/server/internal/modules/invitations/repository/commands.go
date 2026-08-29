package invitationsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	invitationsdomain "github.com/complexus-tech/projects-api/internal/modules/invitations/domain"
	invitationsql "github.com/complexus-tech/projects-api/internal/modules/invitations/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *repo) CreateBulkInvitations(
	ctx context.Context,
	actorID uuid.UUID,
	commands []invitationsdomain.NewWorkspaceInvitation,
) ([]invitationsdomain.WorkspaceInvitation, error) {
	created := make([]invitationsdomain.WorkspaceInvitation, 0, len(commands))
	err := r.withinTransaction(ctx, func(queries invitationsql.Querier) error {
		if err := lockInvitationWorkspaceAdmins(ctx, queries, actorID, commands); err != nil {
			return err
		}

		locks := append([]invitationsdomain.NewWorkspaceInvitation(nil), commands...)
		sort.Slice(locks, func(i, j int) bool {
			left, right := locks[i].Invitation, locks[j].Invitation
			if left.WorkspaceID != right.WorkspaceID {
				return left.WorkspaceID.String() < right.WorkspaceID.String()
			}
			return strings.ToLower(left.Email) < strings.ToLower(right.Email)
		})
		var previousLock string
		for _, command := range locks {
			invitation := command.Invitation
			lockKey := invitation.WorkspaceID.String() + "\x00" + strings.ToLower(invitation.Email)
			if lockKey == previousLock {
				continue
			}
			if err := queries.LockInvitationRecipient(ctx, invitationsql.LockInvitationRecipientParams{
				WorkspaceID: invitation.WorkspaceID,
				Email:       invitation.Email,
			}); err != nil {
				return fmt.Errorf("lock invitation recipient: %w", err)
			}
			previousLock = lockKey
		}
		for _, command := range commands {
			invitation, err := createInvitation(ctx, queries, command)
			if err != nil {
				return err
			}
			created = append(created, invitation)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func createInvitation(
	ctx context.Context,
	queries invitationsql.Querier,
	command invitationsdomain.NewWorkspaceInvitation,
) (invitationsdomain.WorkspaceInvitation, error) {
	invitation := command.Invitation
	if len(command.Token.Digest) != 32 || len(command.Token.Nonce) != 32 ||
		command.Token.KeyID == "" || command.Token.Version <= 0 {
		return invitationsdomain.WorkspaceInvitation{}, errors.New("invitation token persistence metadata is invalid")
	}
	revokedAt := invitation.CreatedAt.UTC()
	if _, err := queries.RevokePendingInvitationsForEmail(ctx, invitationsql.RevokePendingInvitationsForEmailParams{
		RevokedAt:   &revokedAt,
		WorkspaceID: invitation.WorkspaceID,
		Email:       invitation.Email,
	}); err != nil {
		return invitationsdomain.WorkspaceInvitation{}, fmt.Errorf("revoke superseded invitations: %w", err)
	}

	keyID, version := command.Token.KeyID, command.Token.Version
	row, err := queries.CreateInvitation(ctx, invitationsql.CreateInvitationParams{
		InvitationID: invitation.ID,
		WorkspaceID:  invitation.WorkspaceID,
		InviterID:    invitation.InviterID,
		Email:        invitation.Email,
		Role:         invitationsql.UserRole(invitation.Role),
		TokenDigest:  command.Token.Digest,
		TokenNonce:   command.Token.Nonce,
		TokenKeyID:   &keyID,
		TokenVersion: &version,
		ExpiresAt:    invitation.ExpiresAt.UTC(),
		CreatedAt:    invitation.CreatedAt.UTC(),
	})
	if err != nil {
		if platformdatabase.Classify(err) == platformdatabase.ErrorClassUniqueViolation {
			return invitationsdomain.WorkspaceInvitation{}, invitationsdomain.ErrDuplicateInvitation
		}
		return invitationsdomain.WorkspaceInvitation{}, fmt.Errorf("create invitation: %w", err)
	}
	invitation.ID = row.InvitationID
	invitation.CreatedAt = row.CreatedAt
	invitation.UpdatedAt = row.UpdatedAt

	for _, teamID := range invitation.TeamIDs {
		rows, err := queries.AddInvitationTeam(ctx, invitationsql.AddInvitationTeamParams{
			InvitationID: invitation.ID,
			TeamID:       teamID,
			WorkspaceID:  invitation.WorkspaceID,
		})
		if err != nil {
			return invitationsdomain.WorkspaceInvitation{}, fmt.Errorf("add invitation team: %w", err)
		}
		if rows != 1 {
			return invitationsdomain.WorkspaceInvitation{}, invitationsdomain.ErrInvalidInvitationTeam
		}
	}

	payload, err := json.Marshal(command.EmailOutbox)
	if err != nil {
		return invitationsdomain.WorkspaceInvitation{}, fmt.Errorf("marshal invitation email outbox: %w", err)
	}
	readyAt := invitation.CreatedAt.UTC()
	if _, err := queries.InsertInvitationOutboxEvent(ctx, invitationsql.InsertInvitationOutboxEventParams{
		OutboxID:       uuid.New(),
		InvitationID:   invitation.ID,
		WorkspaceID:    invitation.WorkspaceID,
		ActorID:        invitation.InviterID,
		EventType:      string(events.InvitationEmail),
		EventPayload:   payload,
		IdempotencyKey: "invitation-email:" + invitation.ID.String(),
		ReadyAt:        &readyAt,
	}); err != nil {
		return invitationsdomain.WorkspaceInvitation{}, fmt.Errorf("create invitation email outbox: %w", err)
	}

	return invitation, nil
}

func (r *repo) RevokeInvitation(
	ctx context.Context,
	workspaceID uuid.UUID,
	actorID uuid.UUID,
	invitationID uuid.UUID,
	revokedAt time.Time,
) error {
	revokedAt = revokedAt.UTC()
	return r.withinTransaction(ctx, func(queries invitationsql.Querier) error {
		if err := lockActiveWorkspaceAdmin(ctx, queries, workspaceID, actorID); err != nil {
			return err
		}

		rows, err := queries.RevokeInvitation(ctx, invitationsql.RevokeInvitationParams{
			RevokedAt:    &revokedAt,
			InvitationID: invitationID,
			WorkspaceID:  workspaceID,
		})
		if err != nil {
			return fmt.Errorf("revoke invitation: %w", err)
		}
		if rows != 1 {
			return invitationsdomain.ErrInvitationNotFound
		}
		return nil
	})
}

func (r *repo) AcceptInvitation(
	ctx context.Context,
	command invitationsdomain.AcceptCommand,
) (invitationsdomain.WorkspaceInvitation, error) {
	var (
		accepted      invitationsdomain.WorkspaceInvitation
		alreadyMember bool
	)
	err := r.withinTransaction(ctx, func(queries invitationsql.Querier) error {
		row, err := queries.LockInvitationByToken(ctx, lockTokenLookupParams(command.Lookup))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return invitationsdomain.ErrInvitationNotFound
			}
			return fmt.Errorf("lock invitation for acceptance: %w", err)
		}
		accepted = invitationFromLock(row)
		if accepted.UsedAt != nil {
			return invitationsdomain.ErrInvitationUsed
		}
		if !accepted.ExpiresAt.After(command.AcceptedAt) {
			return invitationsdomain.ErrInvitationExpired
		}

		matches, err := queries.ActiveInviteeMatchesInvitation(ctx, invitationsql.ActiveInviteeMatchesInvitationParams{
			UserID: command.UserID,
			Email:  accepted.Email,
		})
		if err != nil {
			return fmt.Errorf("validate invitation recipient: %w", err)
		}
		if !matches {
			return invitationsdomain.ErrInvitationNotFound
		}

		alreadyMember, err = queries.WorkspaceMembershipExists(ctx, invitationsql.WorkspaceMembershipExistsParams{
			WorkspaceID: accepted.WorkspaceID,
			UserID:      command.UserID,
		})
		if err != nil {
			return fmt.Errorf("check existing workspace membership: %w", err)
		}
		if !alreadyMember {
			rows, err := queries.AddWorkspaceMembership(ctx, invitationsql.AddWorkspaceMembershipParams{
				WorkspaceID: accepted.WorkspaceID,
				UserID:      command.UserID,
				Role:        invitationsql.UserRole(accepted.Role),
				AcceptedAt:  command.AcceptedAt.UTC(),
			})
			if err != nil {
				return fmt.Errorf("add invitation workspace membership: %w", err)
			}
			if rows != 1 {
				alreadyMember = true
			}
		}

		if alreadyMember {
			return consumeInvitation(ctx, queries, accepted.ID, command.AcceptedAt)
		}

		if _, err := queries.AddInvitationTeamMemberships(ctx, invitationsql.AddInvitationTeamMembershipsParams{
			UserID:       command.UserID,
			AcceptedAt:   command.AcceptedAt.UTC(),
			WorkspaceID:  accepted.WorkspaceID,
			InvitationID: accepted.ID,
		}); err != nil {
			return fmt.Errorf("add invitation team memberships: %w", err)
		}
		workspaceID := accepted.WorkspaceID
		rows, err := queries.UpdateInviteeLastWorkspace(ctx, invitationsql.UpdateInviteeLastWorkspaceParams{
			WorkspaceID: &workspaceID,
			AcceptedAt:  command.AcceptedAt.UTC(),
			UserID:      command.UserID,
		})
		if err != nil {
			return fmt.Errorf("update invitee workspace: %w", err)
		}
		if rows != 1 {
			return invitationsdomain.ErrInvitationNotFound
		}
		if err := consumeInvitation(ctx, queries, accepted.ID, command.AcceptedAt); err != nil {
			return err
		}

		details, err := queries.GetInvitationAcceptedEventDetails(ctx, invitationsql.GetInvitationAcceptedEventDetailsParams{
			InviteeID:    command.UserID,
			InvitationID: accepted.ID,
		})
		if err != nil {
			return fmt.Errorf("snapshot invitation acceptance event: %w", err)
		}
		payload, err := json.Marshal(events.InvitationAcceptedPayload{
			InviterEmail:  details.InviterEmail,
			InviterName:   details.InviterName,
			InviteeName:   details.InviteeName,
			InviteeEmail:  details.InviteeEmail,
			Role:          accepted.Role,
			WorkspaceID:   accepted.WorkspaceID,
			WorkspaceName: details.WorkspaceName,
			WorkspaceSlug: details.WorkspaceSlug,
		})
		if err != nil {
			return fmt.Errorf("marshal invitation acceptance outbox: %w", err)
		}
		readyAt := command.AcceptedAt.UTC()
		if _, err := queries.InsertInvitationOutboxEvent(ctx, invitationsql.InsertInvitationOutboxEventParams{
			OutboxID:       uuid.New(),
			InvitationID:   accepted.ID,
			WorkspaceID:    accepted.WorkspaceID,
			ActorID:        command.UserID,
			EventType:      string(events.InvitationAccepted),
			EventPayload:   payload,
			IdempotencyKey: "invitation-accepted:" + accepted.ID.String(),
			ReadyAt:        &readyAt,
		}); err != nil {
			return fmt.Errorf("create invitation acceptance outbox: %w", err)
		}
		return nil
	})
	if err != nil {
		return invitationsdomain.WorkspaceInvitation{}, err
	}
	if alreadyMember {
		return invitationsdomain.WorkspaceInvitation{}, invitationsdomain.ErrAlreadyWorkspaceMember
	}
	return accepted, nil
}

func consumeInvitation(
	ctx context.Context,
	queries invitationsql.Querier,
	invitationID uuid.UUID,
	acceptedAt time.Time,
) error {
	acceptedAt = acceptedAt.UTC()
	rows, err := queries.ConsumeInvitation(ctx, invitationsql.ConsumeInvitationParams{
		AcceptedAt:   &acceptedAt,
		InvitationID: invitationID,
	})
	if err != nil {
		return fmt.Errorf("consume invitation: %w", err)
	}
	if rows != 1 {
		return invitationsdomain.ErrInvitationUsed
	}
	return nil
}
