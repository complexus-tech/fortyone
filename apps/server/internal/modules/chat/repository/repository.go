package chatrepository

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	chatsql "github.com/complexus-tech/projects-api/internal/modules/chat/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool              *pgxpool.Pool
	queries           chatsql.Querier
	retireAttachments func(context.Context, uuid.UUID, []uuid.UUID) error
}

type Option func(*Repository)

// WithAttachmentRetirement composes orphan checks and durable object retirement
// in the same unit of work that removes authorized attachment references.
func WithAttachmentRetirement(retire func(context.Context, uuid.UUID, []uuid.UUID) error) Option {
	return func(r *Repository) { r.retireAttachments = retire }
}

func New(_ *logger.Logger, pool *pgxpool.Pool, options ...Option) *Repository {
	r := &Repository{pool: pool}
	if pool != nil {
		r.queries = chatsql.New(pool)
	}
	for _, option := range options {
		option(r)
	}
	return r
}

func (r *Repository) query(ctx context.Context) chatsql.Querier {
	if tx, ok := platformdatabase.UnitTransaction(ctx); ok {
		return chatsql.New(tx)
	}
	return r.queries
}

func scopedError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}

func (r *Repository) authorize(ctx context.Context, scope domain.ActorScope, id uuid.UUID, lock bool) (string, error) {
	if scope.Actor.WorkspaceID != scope.WorkspaceID || !scope.Actor.IsUserActor() {
		return "", domain.ErrForbidden
	}
	q := r.query(ctx)
	var kind string
	var teamID *uuid.UUID
	if lock {
		if _, err := q.LockChatWorkspace(ctx, chatsql.LockChatWorkspaceParams{WorkspaceID: scope.WorkspaceID}); err != nil {
			return "", scopedError(err)
		}
		row, err := q.LockConversation(ctx, chatsql.LockConversationParams{WorkspaceID: scope.WorkspaceID, ConversationID: id, UserID: scope.Actor.PrincipalID})
		if err != nil {
			return "", scopedError(err)
		}
		kind, teamID = row.Kind, row.TeamID
	} else {
		row, err := q.AuthorizeConversation(ctx, chatsql.AuthorizeConversationParams{WorkspaceID: scope.WorkspaceID, ConversationID: id, UserID: scope.Actor.PrincipalID})
		if err != nil {
			return "", scopedError(err)
		}
		kind, teamID = row.Kind, row.TeamID
	}
	if (teamID != nil && !scope.Actor.TeamAccess.Allows(*teamID)) || (kind == "dm" && !scope.Actor.TeamAccess.IsUnrestricted()) {
		return "", domain.ErrNotFound
	}
	return kind, nil
}

func decode[T any](raw []byte) (T, error) {
	var value T
	if err := json.Unmarshal(raw, &value); err != nil {
		return value, fmt.Errorf("decode chat projection: %w", err)
	}
	return value, nil
}

func (r *Repository) identity(ctx context.Context, scope domain.ActorScope, id uuid.UUID) (chatsql.GetMessageIdentityRow, error) {
	row, err := r.query(ctx).GetMessageIdentity(ctx, chatsql.GetMessageIdentityParams{WorkspaceID: scope.WorkspaceID, MessageID: id})
	return row, scopedError(err)
}

func (r *Repository) message(ctx context.Context, scope domain.ActorScope, id uuid.UUID) (domain.Message, error) {
	raw, err := r.query(ctx).GetMessage(ctx, chatsql.GetMessageParams{WorkspaceID: scope.WorkspaceID, MessageID: id, UserID: scope.Actor.PrincipalID})
	if err != nil {
		return domain.Message{}, scopedError(err)
	}
	return decode[domain.Message](raw)
}

type cursor struct {
	ConversationID uuid.UUID  `json:"c,omitempty"`
	ThreadRootID   *uuid.UUID `json:"r,omitempty"`
	Sequence       int64      `json:"s,omitempty"`
	Time           *time.Time `json:"t,omitempty"`
	ID             *uuid.UUID `json:"i,omitempty"`
	ResourceScope  string     `json:"f,omitempty"`
}

func encodeCursor(value cursor) string {
	raw, _ := json.Marshal(value)
	return base64.RawURLEncoding.EncodeToString(raw)
}

func decodeCursor(value string) (cursor, error) {
	if len(value) > 1024 {
		return cursor{}, domain.ErrInvalid
	}
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return cursor{}, domain.ErrInvalid
	}
	var result cursor
	if json.Unmarshal(raw, &result) != nil {
		return cursor{}, domain.ErrInvalid
	}
	return result, nil
}

func sequenceCursor(value string, conversation uuid.UUID, root *uuid.UUID) (int64, error) {
	decoded, err := decodeCursor(value)
	if err != nil || decoded.ConversationID != conversation || !sameID(decoded.ThreadRootID, root) || decoded.Sequence < 1 {
		return 0, domain.ErrInvalid
	}
	return decoded.Sequence, nil
}

func sameID(a, b *uuid.UUID) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func cursorString(value cursor) *string { encoded := encodeCursor(value); return &encoded }
