package okractivitiesrepository

import (
	"context"

	okractivities "github.com/complexus-tech/projects-api/internal/modules/okractivities/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository provides access to the OKR activities storage.
type Repository interface {
	Create(ctx context.Context, na okractivities.CoreNewActivity) error
	GetObjectiveActivities(ctx context.Context, objectiveID uuid.UUID, page, pageSize int) ([]okractivities.CoreActivity, bool, error)
	GetKeyResultActivities(ctx context.Context, keyResultID uuid.UUID, page, pageSize int) ([]okractivities.CoreActivity, bool, error)
}

type repo struct {
	db  *sqlx.DB
	log *logger.Logger
}

func New(log *logger.Logger, db *sqlx.DB) *repo {
	return &repo{
		db:  db,
		log: log,
	}
}
