package dataexport

import (
	"context"
	"encoding/json"
	"fmt"

	exportdomain "github.com/complexus-tech/projects-api/internal/modules/dataexport/domain"
	"github.com/google/uuid"
)

type Repository interface {
	Snapshot(context.Context, exportdomain.Scope) (exportdomain.Snapshot, error)
}

type AuditRecorder interface {
	RecordExport(context.Context, uuid.UUID, uuid.UUID, int) error
}

type Service struct {
	repository Repository
	audit      AuditRecorder
}

func New(repository Repository, audit AuditRecorder) *Service {
	return &Service{repository: repository, audit: audit}
}

func (service *Service) Export(ctx context.Context, scope exportdomain.Scope) (exportdomain.Envelope, error) {
	if err := scope.Validate(); err != nil {
		return exportdomain.Envelope{}, err
	}
	snapshot, err := service.repository.Snapshot(ctx, scope)
	if err != nil {
		return exportdomain.Envelope{}, err
	}
	encoded, err := json.Marshal(snapshot.Envelope)
	if err != nil {
		return exportdomain.Envelope{}, fmt.Errorf("encode work export: %w", err)
	}
	if len(encoded) > exportdomain.MaximumBytes {
		return exportdomain.Envelope{}, exportdomain.ErrTooLarge
	}
	if service.audit != nil {
		if err := service.audit.RecordExport(ctx, scope.ActorID, scope.WorkspaceID, snapshot.TaskCount); err != nil {
			return exportdomain.Envelope{}, fmt.Errorf("record work export: %w", err)
		}
	}
	return snapshot.Envelope, nil
}
