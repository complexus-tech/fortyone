package invitationsrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	invitations "github.com/complexus-tech/projects-api/internal/modules/invitations/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// GetInvitation retrieves a workspace invitation by token
func (r *repo) GetInvitation(ctx context.Context, token string) (invitations.CoreWorkspaceInvitation, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.invitations.GetInvitation")
	defer span.End()

	var result dbWorkspaceInvitation
	query := `
		SELECT
			i.invitation_id,
			i.workspace_id,
			i.inviter_id,
			i.email,
			i.role,
			i.token,
			i.expires_at,
			i.used_at,
			i.created_at,
			i.updated_at,
			w.name AS workspace_name,
			w.slug AS workspace_slug,
			w.color AS workspace_color,
			COALESCE(
				(
					SELECT json_agg(t.team_id)
					FROM workspace_invitation_teams t
					WHERE t.invitation_id = i.invitation_id
				),
				'[]'
			) as team_ids
		FROM workspace_invitations i
		JOIN workspaces w ON i.workspace_id = w.workspace_id
		WHERE i.token = :token
	`

	params := map[string]interface{}{
		"token": token,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return invitations.CoreWorkspaceInvitation{}, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &result, params); err != nil {
		if err == sql.ErrNoRows {
			return invitations.CoreWorkspaceInvitation{}, invitations.ErrInvitationNotFound
		}
		errMsg := fmt.Sprintf("failed to get invitation: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get invitation"), trace.WithAttributes(attribute.String("error", errMsg)))
		return invitations.CoreWorkspaceInvitation{}, err
	}

	return toCoreInvitation(result), nil
}

// ListInvitations returns all pending invitations for a workspace
func (r *repo) ListInvitations(ctx context.Context, workspaceID uuid.UUID) ([]invitations.CoreWorkspaceInvitation, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.invitations.ListInvitations")
	defer span.End()

	var results []dbWorkspaceInvitation
	query := `
		SELECT
			i.invitation_id,
			i.workspace_id,
			i.inviter_id,
			i.email,
			i.role,
			i.expires_at,
			i.used_at,
			i.created_at,
			i.updated_at,
			w.name AS workspace_name,
			w.slug AS workspace_slug,
			w.color AS workspace_color,
			COALESCE(
				(
					SELECT json_agg(t.team_id)
					FROM workspace_invitation_teams t
					WHERE t.invitation_id = i.invitation_id
				),
				'[]'
			) as team_ids
		FROM workspace_invitations i
		JOIN workspaces w ON i.workspace_id = w.workspace_id
		WHERE 
			i.workspace_id = :workspace_id
			AND i.used_at IS NULL
			AND i.expires_at > NOW()
		ORDER BY i.created_at DESC
	`

	params := map[string]interface{}{
		"workspace_id": workspaceID,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &results, params); err != nil {
		errMsg := fmt.Sprintf("failed to list invitations: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to list invitations"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	return toCoreInvitations(results), nil
}

// Add BeginTx method
func (r *repo) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

// ListInvitationsByEmail returns all pending invitations for a user's email
func (r *repo) ListInvitationsByEmail(ctx context.Context, email string) ([]invitations.CoreWorkspaceInvitation, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.invitations.ListInvitationsByEmail")
	defer span.End()

	var results []dbWorkspaceInvitation
	query := `
		SELECT
			i.invitation_id,
			i.workspace_id,
			i.inviter_id,
			i.email,
			i.role,
			i.token,
			i.expires_at,
			i.used_at,
			i.created_at,
			i.updated_at,
			w.name AS workspace_name,
			w.slug AS workspace_slug,
			w.color AS workspace_color,
			COALESCE(
				(
					SELECT json_agg(t.team_id)
					FROM workspace_invitation_teams t
					WHERE t.invitation_id = i.invitation_id
				),
				'[]'
			) as team_ids
		FROM workspace_invitations i
		JOIN workspaces w ON i.workspace_id = w.workspace_id
		WHERE 
			i.email = :email
			AND i.used_at IS NULL
			AND i.expires_at > NOW()
		ORDER BY i.created_at DESC
	`

	params := map[string]interface{}{
		"email": email,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &results, params); err != nil {
		errMsg := fmt.Sprintf("failed to list invitations: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to list invitations"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	return toCoreInvitations(results), nil
}
