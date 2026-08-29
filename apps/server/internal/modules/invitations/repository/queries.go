package invitationsrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	invitationsdomain "github.com/complexus-tech/projects-api/internal/modules/invitations/domain"
	invitationsql "github.com/complexus-tech/projects-api/internal/modules/invitations/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *repo) GetInvitation(
	ctx context.Context,
	lookup invitationsdomain.TokenLookup,
) (invitationsdomain.WorkspaceInvitation, error) {
	row, err := r.queries.GetInvitationByToken(ctx, tokenLookupParams(lookup))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return invitationsdomain.WorkspaceInvitation{}, invitationsdomain.ErrInvitationNotFound
		}
		return invitationsdomain.WorkspaceInvitation{}, fmt.Errorf("get invitation by token digest: %w", err)
	}
	return invitationFromGet(row), nil
}

func (r *repo) ListInvitations(
	ctx context.Context,
	workspaceID uuid.UUID,
	actorID uuid.UUID,
	now time.Time,
) ([]invitationsdomain.WorkspaceInvitation, error) {
	var result []invitationsdomain.WorkspaceInvitation
	err := r.withinTransaction(ctx, func(queries invitationsql.Querier) error {
		if err := lockActiveWorkspaceAdmin(ctx, queries, workspaceID, actorID); err != nil {
			return err
		}

		rows, err := queries.ListWorkspaceInvitations(ctx, invitationsql.ListWorkspaceInvitationsParams{
			WorkspaceID: workspaceID,
			Now:         now.UTC(),
		})
		if err != nil {
			return fmt.Errorf("list workspace invitations: %w", err)
		}
		result = make([]invitationsdomain.WorkspaceInvitation, 0, len(rows))
		for _, row := range rows {
			result = append(result, invitationFromWorkspaceList(row))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *repo) ListInvitationsByEmail(
	ctx context.Context,
	email string,
	now time.Time,
) ([]invitationsdomain.WorkspaceInvitation, error) {
	rows, err := r.queries.ListInvitationsByEmail(ctx, invitationsql.ListInvitationsByEmailParams{
		Email: email,
		Now:   now,
	})
	if err != nil {
		return nil, fmt.Errorf("list invitations by email: %w", err)
	}
	result := make([]invitationsdomain.WorkspaceInvitation, 0, len(rows))
	for _, row := range rows {
		result = append(result, invitationFromEmailList(row))
	}
	return result, nil
}

func tokenLookupParams(lookup invitationsdomain.TokenLookup) invitationsql.GetInvitationByTokenParams {
	var version *int16
	if lookup.Version > 0 {
		value := lookup.Version
		version = &value
	}
	return invitationsql.GetInvitationByTokenParams{
		TokenKeyID:   lookup.KeyID,
		TokenVersion: version,
		TokenDigest:  lookup.Digest,
		LegacyToken:  lookup.LegacyToken,
	}
}

func lockTokenLookupParams(lookup invitationsdomain.TokenLookup) invitationsql.LockInvitationByTokenParams {
	params := tokenLookupParams(lookup)
	return invitationsql.LockInvitationByTokenParams(params)
}
