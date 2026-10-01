package enterprisessohttp

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/enterprisesso/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type Service interface {
	Get(context.Context, domain.Scope) (*domain.Connection, error)
	Create(context.Context, domain.Scope, domain.Create) (domain.Connection, error)
	Update(context.Context, domain.Scope, domain.Update, domain.SessionProof) (domain.Connection, error)
	Archive(context.Context, domain.Scope, string) error
	Public(context.Context, string) (domain.Connection, error)
	AuthorizationURL(context.Context, domain.Connection, domain.Attempt, string) (string, error)
	Complete(context.Context, domain.Attempt, string) (uuid.UUID, domain.Identity, error)
}
type SessionIssuer interface {
	Issue(context.Context, http.ResponseWriter, *http.Request, uuid.UUID, auth.WorkspaceSSOAuthentication) error
}
type StateStore interface {
	Set(context.Context, string, any, time.Duration) error
	Take(context.Context, string, any) error
}
type Handlers struct {
	service                               Service
	states                                StateStore
	issuer                                SessionIssuer
	callbackURL, websiteURL, cookieDomain string
	production                            bool
}

func New(service Service, states StateStore, issuer SessionIssuer, callbackURL, websiteURL, cookieDomain string, production bool) *Handlers {
	return &Handlers{service: service, states: states, issuer: issuer, callbackURL: callbackURL, websiteURL: strings.TrimRight(websiteURL, "/"), cookieDomain: cookieDomain, production: production}
}
func scope(ctx context.Context) (domain.Scope, error) {
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
func proof(ctx context.Context, workspaceID uuid.UUID) domain.SessionProof {
	session, ok := auth.GetBrowserSession(ctx)
	if !ok || session.WorkspaceSSO == nil || session.WorkspaceSSO.WorkspaceID != workspaceID {
		return domain.SessionProof{}
	}
	sso := session.WorkspaceSSO
	return domain.SessionProof{ConnectionID: sso.ConnectionID, Generation: sso.Generation, AuthenticatedAt: sso.AuthenticatedAt}
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
		case errors.Is(err, domain.ErrConflict):
			status = http.StatusConflict
		case errors.Is(err, domain.ErrInvalid):
			status = http.StatusBadRequest
		case errors.Is(err, domain.ErrProvider):
			status = http.StatusBadGateway
		}
		return web.RespondError(ctx, w, err, status)
	}
	return web.Respond(ctx, w, result, http.StatusOK)
}
func (h *Handlers) Get(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := scope(ctx)
	if err != nil {
		return respond(ctx, w, nil, err)
	}
	connection, err := h.service.Get(ctx, scope)
	if err != nil {
		return respond(ctx, w, nil, err)
	}
	verified := false
	if connection != nil {
		p := proof(ctx, scope.WorkspaceID)
		verified = p.ConnectionID == connection.ID && p.Generation == connection.Generation && !p.AuthenticatedAt.IsZero()
	}
	workspace, _ := mid.GetWorkspace(ctx)
	return respond(ctx, w, struct {
		Connection  *domain.Connection `json:"connection"`
		Verified    bool               `json:"verified"`
		CallbackURL string             `json:"callbackUrl"`
		SignInURL   string             `json:"signInUrl"`
	}{connection, verified, h.callbackURL, strings.TrimSuffix(h.callbackURL, "/callback") + "/" + url.PathEscape(workspace.Slug) + "?returnTo=security"}, nil)
}
func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := scope(ctx)
	if err != nil {
		return respond(ctx, w, nil, err)
	}
	var input domain.Create
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	result, err := h.service.Create(ctx, scope, input)
	return respond(ctx, w, result, err)
}
func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := scope(ctx)
	if err != nil {
		return respond(ctx, w, nil, err)
	}
	var input domain.Update
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	result, err := h.service.Update(ctx, scope, input, proof(ctx, scope.WorkspaceID))
	return respond(ctx, w, result, err)
}
func (h *Handlers) Archive(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := scope(ctx)
	if err != nil {
		return respond(ctx, w, nil, err)
	}
	var input struct {
		Reason string `json:"reason"`
	}
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	return respond(ctx, w, nil, h.service.Archive(ctx, scope, input.Reason))
}

func (h *Handlers) PublicStatus(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	connection, err := h.service.Public(ctx, web.Params(r, "workspaceSlug"))
	if errors.Is(err, domain.ErrNotFound) {
		return respond(ctx, w, struct {
			Enabled  bool `json:"enabled"`
			Required bool `json:"requireSSO"`
		}{}, nil)
	}
	if err != nil {
		return respond(ctx, w, nil, err)
	}
	return respond(ctx, w, struct {
		Enabled  bool `json:"enabled"`
		Required bool `json:"requireSSO"`
	}{connection.Enabled, connection.RequireSSO}, nil)
}
