package documentsgrp

import (
	"context"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/documents"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Handlers struct {
	documents *documents.Service
	// audit  *audit.Service
}

// New constructs a new documents handlers instance.
func New(documents *documents.Service) *Handlers {
	return &Handlers{
		documents: documents,
	}
}

// List returns a list of documents.
func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	documents, err := h.documents.List(ctx)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppDocuments(documents), http.StatusOK)
	return nil
}
