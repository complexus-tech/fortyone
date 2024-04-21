package documentsrepo

import (
	"context"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/documents"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

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

func (r *repo) List(ctx context.Context) ([]documents.CoreDocument, error) {

	p := []dbDocument{
		{ID: uuid.New(), Name: "Document 1", Description: "This is document 1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "Document 2", Description: "This is document 2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "Document 3", Description: "This is document 3", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	return toCoreDocuments(p), nil
}
