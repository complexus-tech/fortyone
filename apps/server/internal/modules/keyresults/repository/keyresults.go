package keyresultsrepository

import (
	"context"
	"errors"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository errors
var (
	ErrNotFound = errors.New("key result not found")
)

// Repository defines the repository for key results
type Repository interface {
	Create(ctx context.Context, kr *CoreKeyResult) (uuid.UUID, error)
	Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, updates map[string]any) error
	Delete(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (CoreKeyResult, error)
	List(ctx context.Context, objectiveId uuid.UUID, workspaceId uuid.UUID) ([]CoreKeyResult, error)
	ListPaginated(ctx context.Context, filters CoreKeyResultFilters) (CoreKeyResultListResponse, error)
	AddContributors(ctx context.Context, keyResultID uuid.UUID, contributorIDs []uuid.UUID) error
	UpdateContributors(ctx context.Context, keyResultID uuid.UUID, contributorIDs []uuid.UUID) error
	GetContributors(ctx context.Context, keyResultID uuid.UUID) ([]uuid.UUID, error)
}

type repo struct {
	db  *sqlx.DB
	log *logger.Logger
}

// New creates a new key results repository
func New(log *logger.Logger, db *sqlx.DB) *repo {
	return &repo{
		db:  db,
		log: log,
	}
}
