package invitationsrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	invitations "github.com/complexus-tech/projects-api/internal/modules/invitations/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// CreateBulkInvitations creates multiple workspace invitations in a transaction
func (r *repo) CreateBulkInvitations(ctx context.Context, tx *sql.Tx, invites []invitations.CoreWorkspaceInvitation) ([]invitations.CoreWorkspaceInvitation, error) {
	// Revoke any existing pending invitations
	revokeQuery := `
		UPDATE workspace_invitations
		SET 
			used_at = NOW(),
			updated_at = NOW()
		WHERE workspace_id = :workspace_id 
		AND email = :email 
		AND used_at IS NULL 
		AND expires_at > NOW()`

	revokeStmt, err := r.db.PrepareNamedContext(ctx, revokeQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare revoke statement: %w", err)
	}
	defer revokeStmt.Close()

	for _, inv := range invites {
		params := map[string]any{
			"workspace_id": inv.WorkspaceID,
			"email":        inv.Email,
		}

		if _, err := revokeStmt.ExecContext(ctx, params); err != nil {
			return nil, fmt.Errorf("failed to revoke existing invitations: %w", err)
		}
	}

	// Prepare the statement for inserting invitations
	invQuery := `INSERT INTO workspace_invitations (workspace_id, inviter_id, email, role, token, expires_at) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING invitation_id, created_at, updated_at`

	invStmt, err := tx.PrepareContext(ctx, invQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare invitation statement: %w", err)
	}
	defer invStmt.Close()

	// Prepare the statement for inserting team assignments
	teamQuery := `INSERT INTO workspace_invitation_teams (invitation_id, team_id) VALUES ($1, $2)`

	teamStmt, err := tx.PrepareContext(ctx, teamQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare team statement: %w", err)
	}
	defer teamStmt.Close()

	results := make([]invitations.CoreWorkspaceInvitation, 0, len(invites))
	for _, inv := range invites {
		// Insert the invitation
		var result struct {
			ID        uuid.UUID `db:"invitation_id"`
			CreatedAt time.Time `db:"created_at"`
			UpdatedAt time.Time `db:"updated_at"`
		}

		err := tx.QueryRowContext(ctx, invQuery,
			inv.WorkspaceID,
			inv.InviterID,
			inv.Email,
			inv.Role,
			inv.Token,
			inv.ExpiresAt,
		).Scan(&result.ID, &result.CreatedAt, &result.UpdatedAt)

		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				return nil, invitations.ErrDuplicateInvitation
			}
			return nil, fmt.Errorf("failed to insert invitation: %w", err)
		}

		// Insert team assignments
		for _, teamID := range inv.TeamIDs {
			_, err = teamStmt.ExecContext(ctx, result.ID, teamID)
			if err != nil {
				return nil, fmt.Errorf("failed to insert team assignment: %w", err)
			}
		}

		// Update the invitation with the returned values
		inv.ID = result.ID
		inv.CreatedAt = result.CreatedAt
		inv.UpdatedAt = result.UpdatedAt
		results = append(results, inv)
	}

	return results, nil
}

// RevokeInvitation marks an invitation as used to effectively revoke it
func (r *repo) RevokeInvitation(ctx context.Context, invitationID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.invitations.RevokeInvitation")
	defer span.End()

	query := `
		UPDATE workspace_invitations
		SET 
			used_at = NOW(),
			updated_at = NOW()
		WHERE invitation_id = :invitation_id
		AND used_at IS NULL
	`

	params := map[string]interface{}{
		"invitation_id": invitationID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		errMsg := fmt.Sprintf("failed to revoke invitation: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to revoke invitation"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return invitations.ErrInvitationNotFound
	}

	return nil
}

// MarkInvitationUsed marks an invitation as used
func (r *repo) MarkInvitationUsed(ctx context.Context, invitationID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.invitations.MarkInvitationUsed")
	defer span.End()

	query := `
		UPDATE workspace_invitations
		SET 
			used_at = NOW(),
			updated_at = NOW()
		WHERE invitation_id = :invitation_id
		AND used_at IS NULL
	`

	params := map[string]interface{}{
		"invitation_id": invitationID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		errMsg := fmt.Sprintf("failed to mark invitation used: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to mark invitation used"), trace.WithAttributes(attribute.String("error", errMsg)))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return invitations.ErrInvitationNotFound
	}

	return nil
}
