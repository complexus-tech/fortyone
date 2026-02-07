package documentsrepository

import (
	"context"
	"time"

	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	"github.com/google/uuid"
)

func (r *repo) List(ctx context.Context) ([]documents.CoreDocument, error) {

	p := []dbDocument{
		{ID: uuid.New(), Name: "Document 1", Description: "This is document 1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "Document 2", Description: "This is document 2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "Document 3", Description: "This is document 3", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	return toCoreDocuments(p), nil
}
