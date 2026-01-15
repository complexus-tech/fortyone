package documentsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/documents"
	"github.com/google/uuid"
)

// AppDocumentList represents a document in the application layer.
type AppDocumentsList struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"title"`
	Description string    `json:"description"`
}

// toAppDocuments converts a list of core documents to a list of application documents.
func toAppDocuments(documents []documents.CoreDocument) []AppDocumentsList {
	appdocuments := make([]AppDocumentsList, len(documents))
	for i, document := range documents {
		appdocuments[i] = AppDocumentsList{
			ID:          document.ID,
			Name:        document.Name,
			Description: document.Description,
		}
	}
	return appdocuments
}
