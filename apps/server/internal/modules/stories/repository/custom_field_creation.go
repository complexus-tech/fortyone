package storiesrepository

import (
	"context"
	"fmt"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// CustomFieldCreation is a callback-scoped participant. It exposes only the
// story-creation invariant; the bootstrap adapter binds its owning query set
// to the story transaction without exposing that transaction to the service.
type CustomFieldCreation interface {
	Apply(context.Context, storydomain.MutationScope, uuid.UUID, uuid.UUID, []storydomain.CustomFieldValue) error
	Close()
}

type CustomFieldCreationBinder func(pgx.Tx) CustomFieldCreation

func WithCustomFieldCreation(binder CustomFieldCreationBinder) Option {
	return func(repository *repo) { repository.customFieldCreationBinder = binder }
}

func (r *repo) applyCreationFields(ctx context.Context, tx pgx.Tx, command storydomain.CreateStoryCommand) error {
	if len(command.CustomFieldValues) == 0 {
		return nil
	}
	if r.customFieldCreationBinder == nil {
		return fmt.Errorf("%w: custom field transaction participant is unavailable", storydomain.ErrInvalidMutation)
	}
	participant := r.customFieldCreationBinder(tx)
	if participant == nil {
		return fmt.Errorf("%w: custom field transaction participant is unavailable", storydomain.ErrInvalidMutation)
	}
	defer participant.Close()
	return participant.Apply(ctx, command.Scope, command.Story.ID, command.Story.Team, command.CustomFieldValues)
}
