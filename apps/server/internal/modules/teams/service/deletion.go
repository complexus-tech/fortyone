package teams

import (
	"context"
	"errors"
	"time"

	teamsdomain "github.com/complexus-tech/projects-api/internal/modules/teams/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

// DeletionManager commits the team, its dependent data and durable cleanup
// receipts together. It must recheck administrative membership in storage.
type DeletionManager interface {
	Delete(context.Context, teamsdomain.Deletion) error
}

func (s *Service) ConfigureDeletion(manager DeletionManager) {
	s.deletion = manager
}

func (s *Service) Delete(ctx context.Context, teamID, workspaceID uuid.UUID) error {
	actor, err := platformauth.GetActor(ctx)
	if err != nil {
		return teamsdomain.ErrDeletionForbidden
	}
	command := teamsdomain.Deletion{
		TeamID: teamID, WorkspaceID: workspaceID, Actor: actor, DeletedAt: time.Now().UTC(),
	}
	if err := command.Validate(); err != nil {
		return err
	}
	if s.deletion == nil {
		return errors.New("team deletion is not configured")
	}
	if err := s.deletion.Delete(ctx, command); err != nil {
		s.log.Error(ctx, "team deletion failed", "error", err, "team_id", teamID, "workspace_id", workspaceID)
		return err
	}
	s.log.Info(ctx, "team deleted", "team_id", teamID, "workspace_id", workspaceID)
	return nil
}
