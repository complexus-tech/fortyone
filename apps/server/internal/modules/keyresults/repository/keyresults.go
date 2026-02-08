package keyresultsrepository

import (
	"context"

	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository errors
var (
	ErrNotFound = keyresults.ErrNotFound
)

// Repository defines the repository for key results
type Repository interface {
	Create(ctx context.Context, kr *keyresults.CoreKeyResult) (uuid.UUID, error)
	Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, updates map[string]any) error
	Delete(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error
	Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (keyresults.CoreKeyResult, error)
	List(ctx context.Context, objectiveId uuid.UUID, workspaceId uuid.UUID) ([]keyresults.CoreKeyResult, error)
	ListPaginated(ctx context.Context, filters keyresults.CoreKeyResultFilters) (keyresults.CoreKeyResultListResponse, error)
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
