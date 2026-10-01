package customfieldsadapter

import (
	"context"
	"errors"
	"fmt"

	fielddomain "github.com/complexus-tech/projects-api/internal/modules/customfields/domain"
	fieldrepository "github.com/complexus-tech/projects-api/internal/modules/customfields/repository"
	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func CreationBinder(tx pgx.Tx) storyrepository.CustomFieldCreation {
	return &creation{participant: fieldrepository.NewCreationParticipant(tx)}
}

type creation struct {
	participant *fieldrepository.CreationParticipant
}

func (p *creation) Close() { p.participant.Close() }
func (p *creation) Apply(ctx context.Context, scope storydomain.MutationScope, storyID, teamID uuid.UUID, inputs []storydomain.CustomFieldValue) error {
	if !scope.Actor.IsUserActor() {
		return storydomain.ErrMutationForbidden
	}
	values := make([]fielddomain.Value, 0, len(inputs))
	for _, input := range inputs {
		values = append(values, fielddomain.Value{FieldID: input.FieldID, Value: input.Value})
	}
	err := p.participant.Apply(ctx, fielddomain.Scope{ActorID: scope.Actor.PrincipalID, WorkspaceID: scope.WorkspaceID, TeamID: teamID}, storyID, values)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, fielddomain.ErrForbidden), errors.Is(err, fielddomain.ErrNotFound):
		return storydomain.ErrMutationForbidden
	case errors.Is(err, fielddomain.ErrConflict):
		return storydomain.ErrMutationConflict
	case errors.Is(err, fielddomain.ErrInvalid), errors.Is(err, fielddomain.ErrLimit):
		return fmt.Errorf("%w: custom field values do not match the current team definitions", storydomain.ErrInvalidMutation)
	default:
		return err
	}
}
