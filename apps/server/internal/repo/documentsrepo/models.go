package documentsrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/documents"
	"github.com/google/uuid"
)

type dbDocument struct {
	ID          uuid.UUID  `db:"id"`
	Name        string     `db:"name"`
	Description string     `db:"description"`
	Owner       *uuid.UUID `db:"owner"`
	StartDate   *time.Time `db:"start_date"`
	EndDate     *time.Time `db:"end_date"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
}

func toCoreDocument(p dbDocument) documents.CoreDocument {
	return documents.CoreDocument{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Owner:       p.Owner,
		StartDate:   p.StartDate,
		EndDate:     p.EndDate,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toCoreDocuments(do []dbDocument) []documents.CoreDocument {
	documents := make([]documents.CoreDocument, len(do))
	for i, o := range do {
		documents[i] = toCoreDocument(o)
	}
	return documents
}
