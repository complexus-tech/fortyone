package epics

import (
	"context"

	epicsdomain "github.com/complexus-tech/projects-api/internal/modules/epics/domain"
	"github.com/google/uuid"
)

var ErrNotImplemented = epicsdomain.ErrNotImplemented

type Service struct{}

func New() *Service {
	return &Service{}
}

// List fails explicitly until epics have a durable workspace-scoped model.
func (*Service) List(context.Context, uuid.UUID) error {
	return ErrNotImplemented
}
