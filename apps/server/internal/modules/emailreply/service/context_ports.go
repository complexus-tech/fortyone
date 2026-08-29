package emailreply

import (
	"context"
	"time"

	emailreplydomain "github.com/complexus-tech/projects-api/internal/modules/emailreply/domain"
	"github.com/google/uuid"
)

// ContextStore is the persistence boundary used to rebuild and reauthorize an
// email agent's scope. It deliberately exposes use-case reads instead of a
// database handle so the service cannot issue ad-hoc SQL.
type ContextStore interface {
	ActorScope(context.Context, uuid.UUID, uuid.UUID) (emailreplydomain.ActorScope, bool, error)
	LoadTarget(context.Context, uuid.UUID, emailreplydomain.TargetKind, uuid.UUID) (emailreplydomain.TargetSnapshot, bool, error)
	ListStoryChoices(context.Context, uuid.UUID, uuid.UUID, int) ([]emailreplydomain.Choice, error)
	StoryStatusExists(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (bool, error)
	StoryAssigneeExists(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	CurrentVersion(context.Context, uuid.UUID, emailreplydomain.ActionKind, uuid.UUID) (time.Time, bool, error)
	CurrentProposalState(context.Context, uuid.UUID, emailreplydomain.ActionKind, uuid.UUID) (emailreplydomain.ProposalState, bool, error)
	TargetTeam(context.Context, uuid.UUID, emailreplydomain.ActionKind, uuid.UUID) (uuid.UUID, bool, error)
}
