package sprintshttp

import (
	"context"
	"errors"
	"net/http"

	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func respondSprintError(ctx context.Context, w http.ResponseWriter, err error) error {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, sprints.ErrInvalid), errors.Is(err, sprints.ErrInvalidReference):
		status = http.StatusBadRequest
	case errors.Is(err, sprints.ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, sprints.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, sprints.ErrVersionConflict):
		status = http.StatusConflict
	}
	return web.RespondError(ctx, w, err, status)
}

func sprintPathID(ctx context.Context, w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := web.UUIDPathParameter(r, "sprintId")
	if err != nil {
		_ = web.RespondError(ctx, w, err, http.StatusBadRequest)
		return uuid.Nil, false
	}
	return id, true
}
