package invitationsrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/core/invitations"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type repo struct {
	db  *sqlx.DB
	log *logger.Logger
}

func New(log *logger.Logger, db *sqlx.DB) *repo {
	return &repo{
		db:  db,
		log: log,
	}
}

// CreateInvitation creates a new workspace invitation
func (r *repo) CreateInvitation(ctx context.Context, invitation invitations.CoreWorkspaceInvitation) (invitations.CoreWorkspaceInvitation, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.invitations.CreateInvitation")
	defer span.End()

	var result dbWorkspaceInvitation
	query := `
		INSERT INTO workspace_invitations (
			workspace_id,
			inviter_id,
			email,
			role,
			token,
			expires_at
		)
		VALUES (
			:workspace_id,
			:inviter_id,
			:email,
			:role,
			:token,
			:expires_at
		)
		RETURNING
			invitation_id,
			workspace_id,
			inviter_id,
			email,
			role,
			token,
			expires_at,
			used_at,
			created_at,
			updated_at
	`

	params := map[string]interface{}{
		"workspace_id": invitation.WorkspaceID,
		"inviter_id":   invitation.InviterID,
		"email":        invitation.Email,
		"role":         invitation.Role,
		"token":        invitation.Token,
		"expires_at":   invitation.ExpiresAt,
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
		errMsg := fmt.Sprintf("failed to create invitation: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create invitation"), trace.WithAttributes(attribute.String("error", errMsg)))
		return invitations.CoreWorkspaceInvitation{}, err
	}

	return toCoreInvitation(result), nil
}

// GetInvitation retrieves a workspace invitation by token
func (r *repo) GetInvitation(ctx context.Context, token string) (invitations.CoreWorkspaceInvitation, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.invitations.GetInvitation")
	defer span.End()

	var result dbWorkspaceInvitation
	query := `
		SELECT
			invitation_id,
			workspace_id,
			inviter_id,
			email,
			role,
			token,
			expires_at,
			used_at,
			created_at,
			updated_at
		FROM workspace_invitations
		WHERE token = :token
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
			invitation_id,
			workspace_id,
			inviter_id,
			email,
			role,
			token,
			expires_at,
			used_at,
			created_at,
			updated_at
		FROM workspace_invitations
		WHERE 
			workspace_id = :workspace_id
			AND used_at IS NULL
			AND expires_at > NOW()
		ORDER BY created_at DESC
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

// CreateInvitationLink creates a new workspace invitation link
func (r *repo) CreateInvitationLink(ctx context.Context, link invitations.CoreWorkspaceInvitationLink) (invitations.CoreWorkspaceInvitationLink, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.invitations.CreateInvitationLink")
	defer span.End()

	var result dbWorkspaceInvitationLink
	query := `
		INSERT INTO workspace_invitation_links (
			workspace_id,
			creator_id,
			token,
			role,
			max_uses,
			expires_at,
			is_active
		)
		VALUES (
			:workspace_id,
			:creator_id,
			:token,
			:role,
			:max_uses,
			:expires_at,
			:is_active
		)
		RETURNING
			invitation_link_id,
			workspace_id,
			creator_id,
			token,
			role,
			max_uses,
			used_count,
			expires_at,
			is_active,
			created_at,
			updated_at
	`

	params := map[string]interface{}{
		"workspace_id": link.WorkspaceID,
		"creator_id":   link.CreatorID,
		"token":        link.Token,
		"role":         link.Role,
		"max_uses":     link.MaxUses,
		"expires_at":   link.ExpiresAt,
		"is_active":    link.IsActive,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return invitations.CoreWorkspaceInvitationLink{}, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &result, params); err != nil {
		errMsg := fmt.Sprintf("failed to create invitation link: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to create invitation link"), trace.WithAttributes(attribute.String("error", errMsg)))
		return invitations.CoreWorkspaceInvitationLink{}, err
	}

	return toCoreInvitationLink(result), nil
}

// GetInvitationLink retrieves a workspace invitation link by token
func (r *repo) GetInvitationLink(ctx context.Context, token string) (invitations.CoreWorkspaceInvitationLink, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.invitations.GetInvitationLink")
	defer span.End()

	var result dbWorkspaceInvitationLink
	query := `
		SELECT
			invitation_link_id,
			workspace_id,
			creator_id,
			token,
			role,
			max_uses,
			used_count,
			expires_at,
			is_active,
			created_at,
			updated_at
		FROM workspace_invitation_links
		WHERE token = :token
	`

	params := map[string]interface{}{
		"token": token,
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return invitations.CoreWorkspaceInvitationLink{}, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &result, params); err != nil {
		if err == sql.ErrNoRows {
			return invitations.CoreWorkspaceInvitationLink{}, invitations.ErrInvitationNotFound
		}
		errMsg := fmt.Sprintf("failed to get invitation link: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get invitation link"), trace.WithAttributes(attribute.String("error", errMsg)))
		return invitations.CoreWorkspaceInvitationLink{}, err
	}

	return toCoreInvitationLink(result), nil
}

// ListInvitationLinks returns all active invitation links for a workspace
func (r *repo) ListInvitationLinks(ctx context.Context, workspaceID uuid.UUID) ([]invitations.CoreWorkspaceInvitationLink, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.invitations.ListInvitationLinks")
	defer span.End()

	var results []dbWorkspaceInvitationLink
	query := `
		SELECT
			invitation_link_id,
			workspace_id,
			creator_id,
			token,
			role,
			max_uses,
			used_count,
			expires_at,
			is_active,
			created_at,
			updated_at
		FROM workspace_invitation_links
		WHERE 
			workspace_id = :workspace_id
			AND is_active = true
			AND (expires_at IS NULL OR expires_at > NOW())
			AND (max_uses IS NULL OR used_count < max_uses)
		ORDER BY created_at DESC
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
		errMsg := fmt.Sprintf("failed to list invitation links: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to list invitation links"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, err
	}

	return toCoreInvitationLinks(results), nil
}

// RevokeInvitationLink marks an invitation link as inactive
func (r *repo) RevokeInvitationLink(ctx context.Context, linkID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.invitations.RevokeInvitationLink")
	defer span.End()

	query := `
		UPDATE workspace_invitation_links
		SET 
			is_active = false,
			updated_at = NOW()
		WHERE invitation_link_id = :link_id
		AND is_active = true
	`

	params := map[string]interface{}{
		"link_id": linkID,
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
		errMsg := fmt.Sprintf("failed to revoke invitation link: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to revoke invitation link"), trace.WithAttributes(attribute.String("error", errMsg)))
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

// IncrementLinkUsage increments the used_count for an invitation link
func (r *repo) IncrementLinkUsage(ctx context.Context, linkID uuid.UUID) error {
	ctx, span := web.AddSpan(ctx, "business.repository.invitations.IncrementLinkUsage")
	defer span.End()

	query := `
		UPDATE workspace_invitation_links
		SET 
			used_count = used_count + 1,
			updated_at = NOW()
		WHERE 
			invitation_link_id = :link_id
			AND is_active = true
			AND (expires_at IS NULL OR expires_at > NOW())
			AND (max_uses IS NULL OR used_count < max_uses)
	`

	params := map[string]interface{}{
		"link_id": linkID,
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
		errMsg := fmt.Sprintf("failed to increment link usage: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to increment link usage"), trace.WithAttributes(attribute.String("error", errMsg)))
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
