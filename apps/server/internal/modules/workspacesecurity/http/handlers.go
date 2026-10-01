package workspacesecurityhttp

import (
	"context"
	"errors"
	"net/http"

	domain "github.com/complexus-tech/projects-api/internal/modules/workspacesecurity/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/internal/platform/pagination"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type Service interface {
	Policy(context.Context, domain.Scope) (domain.Policy, error)
	UpdatePolicy(context.Context, domain.Scope, domain.PolicyUpdate) (domain.Policy, error)
	Sessions(context.Context, domain.Scope, *uuid.UUID, bool) (domain.SessionList, error)
	RevokeSession(context.Context, domain.Scope, uuid.UUID, string) error
	RevokeMember(context.Context, domain.Scope, uuid.UUID, string) error
	Audit(context.Context, domain.Scope, domain.AuditFilter) ([]domain.AuditEvent, error)
	RecordAuditExport(context.Context, domain.Scope, int) error
}
type Handlers struct {
	service Service
	cursors pagination.CursorCodec[auditCursor]
}

func New(service Service, secret string) (*Handlers, error) {
	key, err := pagination.DeriveSigningKey("active", []byte(secret), "workspace-security.audit")
	if err != nil {
		return nil, err
	}
	codec, err := pagination.NewCursorCodec[auditCursor](key)
	if err != nil {
		return nil, err
	}
	return &Handlers{service: service, cursors: codec}, nil
}
func identities(ctx context.Context) (domain.Scope, error) {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return domain.Scope{}, domain.ErrForbidden
	}
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return domain.Scope{}, domain.ErrForbidden
	}
	return domain.Scope{ActorID: actorID, WorkspaceID: workspace.ID}, nil
}
func respond(ctx context.Context, w http.ResponseWriter, result any, err error) error {
	w.Header().Set("Cache-Control", "private, no-store")
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, domain.ErrForbidden):
			status = http.StatusForbidden
		case errors.Is(err, domain.ErrNotFound):
			status = http.StatusNotFound
		case errors.Is(err, domain.ErrInvalid), errors.Is(err, domain.ErrLimit):
			status = http.StatusBadRequest
		case errors.Is(err, domain.ErrConflict):
			status = http.StatusConflict
		}
		return web.RespondError(ctx, w, err, status)
	}
	return web.Respond(ctx, w, result, http.StatusOK)
}
func (h *Handlers) Policy(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := identities(ctx)
	if err != nil {
		return respond(ctx, w, nil, err)
	}
	result, err := h.service.Policy(ctx, scope)
	return respond(ctx, w, result, err)
}
func (h *Handlers) UpdatePolicy(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := identities(ctx)
	if err != nil {
		return respond(ctx, w, nil, err)
	}
	var input domain.PolicyUpdate
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	result, err := h.service.UpdatePolicy(ctx, scope, input)
	return respond(ctx, w, result, err)
}
func (h *Handlers) Sessions(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := identities(ctx)
	if err != nil {
		return respond(ctx, w, nil, err)
	}
	var userID *uuid.UUID
	var revoked bool
	for key, values := range r.URL.Query() {
		if len(values) != 1 || (key != "userId" && key != "includeRevoked") {
			return respond(ctx, w, nil, domain.ErrInvalid)
		}
	}
	if value := r.URL.Query().Get("userId"); value != "" {
		id, err := uuid.Parse(value)
		if err != nil || id == uuid.Nil {
			return respond(ctx, w, nil, domain.ErrInvalid)
		}
		userID = &id
	}
	if r.URL.Query().Has("includeRevoked") {
		switch r.URL.Query().Get("includeRevoked") {
		case "true":
			revoked = true
		case "false":
		default:
			return respond(ctx, w, nil, domain.ErrInvalid)
		}
	}
	result, err := h.service.Sessions(ctx, scope, userID, revoked)
	if session, ok := auth.GetBrowserSession(ctx); ok {
		for index := range result.Items {
			result.Items[index].Current = result.Items[index].ID == session.SessionID
		}
	}
	return respond(ctx, w, result, err)
}
func (h *Handlers) revoke(ctx context.Context, w http.ResponseWriter, r *http.Request, member bool) error {
	scope, err := identities(ctx)
	if err != nil {
		return respond(ctx, w, nil, err)
	}
	param := "sessionId"
	if member {
		param = "userId"
	}
	id, err := uuid.Parse(web.Params(r, param))
	if err != nil || id == uuid.Nil {
		return respond(ctx, w, nil, domain.ErrInvalid)
	}
	var input domain.RevokeInput
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if member {
		err = h.service.RevokeMember(ctx, scope, id, input.Reason)
	} else {
		err = h.service.RevokeSession(ctx, scope, id, input.Reason)
	}
	return respond(ctx, w, nil, err)
}
func (h *Handlers) RevokeSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return h.revoke(ctx, w, r, false)
}
func (h *Handlers) RevokeMember(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return h.revoke(ctx, w, r, true)
}
