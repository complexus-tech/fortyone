package documentsgrp

import (
	"context"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/documents"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Handlers struct {
	documents *documents.Service
}

func New(documents *documents.Service) *Handlers {
	return &Handlers{
		documents: documents,
	}
}

func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	_, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	documents, err := h.documents.List(ctx)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppDocuments(documents), http.StatusOK)
	return nil
}
