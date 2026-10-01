package stories

import (
	"context"
	"errors"
	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type importReceiptRepository interface {
	RecordImportReceipt(context.Context, storydomain.ImportReceipt) error
	ListImportReceipts(context.Context, auth.Actor, string, string, *string, int, int) ([]storydomain.ImportReceipt, error)
}

func (s *Service) RecordImportReceipt(ctx context.Context, receipt storydomain.ImportReceipt) error {
	store, ok := s.repo.(importReceiptRepository)
	if !ok {
		return errors.New("import receipts unavailable")
	}
	return store.RecordImportReceipt(ctx, receipt)
}
func (s *Service) ListImportReceipts(ctx context.Context, actorID, workspaceID uuid.UUID, provider, digest string, namespace *string, offset int) ([]storydomain.ImportReceipt, error) {
	actor, err := auth.GetActor(ctx)
	if err != nil || actor.PrincipalID != actorID || actor.WorkspaceID != workspaceID {
		return nil, ErrStoryMutationForbidden
	}
	store, ok := s.repo.(importReceiptRepository)
	if !ok {
		return nil, errors.New("import receipts unavailable")
	}
	return store.ListImportReceipts(ctx, actor, provider, digest, namespace, 500, offset)
}
