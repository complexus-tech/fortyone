package keyresults

import (
	"context"
	"fmt"

	keyresultsdomain "github.com/complexus-tech/projects-api/internal/modules/keyresults/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func (service *Service) Create(
	ctx context.Context,
	draft CoreNewKeyResult,
	workspaceID uuid.UUID,
) (CoreKeyResult, error) {
	created, err := service.CreateBatch(ctx, []CoreNewKeyResult{draft}, workspaceID)
	if err != nil {
		return CoreKeyResult{}, err
	}
	if len(created) != 1 {
		return CoreKeyResult{}, fmt.Errorf("create key result: %w", ErrInvalid)
	}
	return created[0], nil
}

func (service *Service) CreateBatch(
	ctx context.Context,
	drafts []CoreNewKeyResult,
	workspaceID uuid.UUID,
) ([]CoreKeyResult, error) {
	if len(drafts) == 0 {
		return []CoreKeyResult{}, nil
	}
	access, err := service.accessFor(ctx, workspaceID, drafts[0].CreatedBy, platformauth.ScopeObjectivesWrite)
	if err != nil {
		return nil, err
	}
	values := make([]keyresultsdomain.NewKeyResult, len(drafts))
	for index, draft := range drafts {
		if draft.CreatedBy != uuid.Nil && draft.CreatedBy != access.ActorID {
			return nil, ErrForbidden
		}
		draft.CreatedBy = access.ActorID
		values[index] = draft
	}
	command, err := (keyresultsdomain.CreateCommand{
		Access: access, KeyResults: values,
	}).Normalize()
	if err != nil {
		return nil, err
	}
	created, err := service.repo.CreateBatch(ctx, command)
	if err != nil {
		return nil, fmt.Errorf("create key results: %w", err)
	}
	return created, nil
}
