package slack

import (
	"context"
	"errors"
	"fmt"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
)

type workspaceAccessStore interface {
	GetWorkspaceRole(ctx context.Context, workspaceID, actorID uuid.UUID) (authorization.WorkspaceRole, error)
}

func (s *Service) requireWorkspaceRole(
	ctx context.Context,
	workspaceID, actorID uuid.UUID,
	minimum authorization.WorkspaceRole,
) error {
	if workspaceID == uuid.Nil || actorID == uuid.Nil {
		return slackdomain.ErrForbidden
	}
	access, ok := s.repo.(workspaceAccessStore)
	if !ok {
		return errors.New("slack workspace authorization is not configured")
	}
	role, err := access.GetWorkspaceRole(ctx, workspaceID, actorID)
	if err != nil {
		if errors.Is(err, slackdomain.ErrNotFound) || errors.Is(err, slackdomain.ErrForbidden) {
			return slackdomain.ErrForbidden
		}
		return fmt.Errorf("authorize Slack workspace access: %w", err)
	}
	if err := authorization.RequireMinimumWorkspaceRole(role, minimum); err != nil {
		return errors.Join(slackdomain.ErrForbidden, err)
	}
	return nil
}

func (s *Service) requireWorkspaceMember(ctx context.Context, workspaceID, actorID uuid.UUID) error {
	return s.requireWorkspaceRole(ctx, workspaceID, actorID, authorization.WorkspaceRoleMember)
}

func (s *Service) requireWorkspaceAdmin(ctx context.Context, workspaceID, actorID uuid.UUID) error {
	return s.requireWorkspaceRole(ctx, workspaceID, actorID, authorization.WorkspaceRoleAdmin)
}
