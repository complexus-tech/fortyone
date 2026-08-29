package keyresults

import (
	"context"

	keyresultsdomain "github.com/complexus-tech/projects-api/internal/modules/keyresults/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func (service *Service) Delete(ctx context.Context, id, workspaceID, userID uuid.UUID) error {
	access, err := service.accessFor(ctx, workspaceID, userID, platformauth.ScopeObjectivesWrite)
	if err != nil {
		return err
	}
	command := keyresultsdomain.DeleteCommand{
		Access: access, KeyResultID: id,
	}
	if err := command.Validate(); err != nil {
		return err
	}
	return service.repo.Delete(ctx, command)
}
