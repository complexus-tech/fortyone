package scimhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	domain "github.com/complexus-tech/projects-api/internal/modules/scim/domain"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type Service interface {
	Authenticate(context.Context, string, string) (domain.Scope, error)
	Status(context.Context, domain.AdminScope) (domain.Status, error)
	Mint(context.Context, domain.AdminScope, domain.MintInput) (domain.Minted, error)
	Revoke(context.Context, domain.AdminScope, uuid.UUID) error
	RetrySeats(context.Context, domain.AdminScope) error
	List(context.Context, domain.Scope, domain.Page) ([]domain.User, int64, error)
	Get(context.Context, domain.Scope, uuid.UUID) (domain.User, error)
	Create(context.Context, domain.Scope, domain.Input) (domain.User, error)
	Replace(context.Context, domain.Scope, uuid.UUID, domain.Input) (domain.User, error)
	Patch(context.Context, domain.Scope, uuid.UUID, domain.Patch) (domain.User, error)
	Delete(context.Context, domain.Scope, uuid.UUID) error
}
type Handlers struct {
	service   Service
	publicURL string
}

func New(service Service, publicURL string) *Handlers {
	return &Handlers{service: service, publicURL: strings.TrimRight(publicURL, "/")}
}
func raw(w http.ResponseWriter, status int, body any) error {
	w.Header().Set("Content-Type", "application/scim+json")
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if status == http.StatusNoContent {
		w.WriteHeader(status)
		return nil
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return err
	}
	w.WriteHeader(status)
	_, err = w.Write(encoded)
	return err
}
func fail(w http.ResponseWriter, err error) error {
	status, kind, detail := http.StatusInternalServerError, "", "SCIM request could not be completed"
	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		status = http.StatusUnauthorized
		detail = "A valid workspace SCIM bearer token is required"
		w.Header().Set("WWW-Authenticate", `Bearer realm="SCIM"`)
	case errors.Is(err, domain.ErrForbidden):
		status = http.StatusForbidden
		detail = "This provisioning operation is not allowed"
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
		detail = "SCIM resource not found"
	case errors.Is(err, domain.ErrConflict):
		status = http.StatusConflict
		kind = "uniqueness"
		detail = "A resource with that identifier already exists"
	case errors.Is(err, domain.ErrChanged):
		status = http.StatusPreconditionFailed
		detail = "Resource changed; retrieve it and retry"
	case errors.Is(err, domain.ErrInvalidInput):
		status = http.StatusBadRequest
		kind = "invalidValue"
		detail = err.Error()
	}
	var classified *domain.Invalid
	if errors.As(err, &classified) {
		kind = classified.Type
	}
	var input *queryError
	if errors.As(err, &input) {
		kind = input.kind
	}
	body := map[string]any{"schemas": []string{domain.ErrorSchema}, "status": strconv.Itoa(status), "detail": detail}
	if kind != "" {
		body["scimType"] = kind
	}
	return raw(w, status, body)
}
func (h *Handlers) authenticate(ctx context.Context, r *http.Request) (domain.Scope, error) {
	values := r.Header.Values("Authorization")
	if len(values) != 1 {
		return domain.Scope{}, domain.ErrUnauthorized
	}
	fields := strings.Fields(values[0])
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
		return domain.Scope{}, domain.ErrUnauthorized
	}
	return h.service.Authenticate(ctx, web.Params(r, "workspaceSlug"), fields[1])
}
func decode(w http.ResponseWriter, r *http.Request, target any) error {
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || (media != "application/scim+json" && media != "application/json") {
		return fmt.Errorf("%w: use application/scim+json", domain.ErrInvalidInput)
	}
	body, err := web.ReadBoundedBody(w, r, 65536)
	if err != nil {
		return fmt.Errorf("%w: resource exceeds the supported request bounds", domain.ErrInvalidInput)
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(target); err != nil {
		return fmt.Errorf("%w: invalid or unsupported resource attributes", domain.ErrInvalidInput)
	}
	if err = decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: expected one resource", domain.ErrInvalidInput)
	}
	return nil
}
func resourceID(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(web.Params(r, "resourceId"))
	if err != nil || id == uuid.Nil {
		return uuid.Nil, domain.ErrNotFound
	}
	return id, nil
}
func (h *Handlers) location(r *http.Request, id uuid.UUID) string {
	return h.publicURL + "/scim/v2/" + url.PathEscape(web.Params(r, "workspaceSlug")) + "/Users/" + id.String()
}
func (h *Handlers) representation(r *http.Request, user domain.User) (map[string]any, error) {
	var profile domain.Profile
	if err := json.Unmarshal(user.Profile, &profile); err != nil {
		return nil, err
	}
	schemas := []string{domain.UserSchema}
	if profile.Enterprise != nil {
		schemas = append(schemas, domain.EnterpriseSchema)
	}
	result := map[string]any{"schemas": schemas, "id": user.ID.String(), "userName": user.UserName, "active": user.Active, "meta": map[string]any{"resourceType": "User", "created": user.CreatedAt, "lastModified": user.UpdatedAt, "location": h.location(r, user.ID)}}
	if user.ExternalID != nil {
		result["externalId"] = *user.ExternalID
	}
	if profile.DisplayName != "" {
		result["displayName"] = profile.DisplayName
	}
	if profile.Name != nil {
		result["name"] = profile.Name
	}
	if len(profile.Emails) > 0 {
		result["emails"] = profile.Emails
	}
	if profile.Enterprise != nil {
		result[domain.EnterpriseSchema] = profile.Enterprise
	}
	return project(result, r.URL.Query())
}
func (h *Handlers) Users(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := h.authenticate(ctx, r)
	if err != nil {
		return fail(w, err)
	}
	if r.Method == http.MethodGet {
		page, err := parsePage(r.URL.Query())
		if err != nil {
			return fail(w, err)
		}
		users, total, err := h.service.List(ctx, scope, page)
		if err != nil {
			return fail(w, err)
		}
		resources := make([]map[string]any, 0, len(users))
		for _, user := range users {
			representation, err := h.representation(r, user)
			if err != nil {
				return fail(w, err)
			}
			resources = append(resources, representation)
		}
		return raw(w, http.StatusOK, map[string]any{"schemas": []string{domain.ListSchema}, "totalResults": total, "startIndex": page.StartIndex, "itemsPerPage": len(resources), "Resources": resources})
	}
	if _, err := parseProjection(r.URL.Query(), false); err != nil {
		return fail(w, err)
	}
	var input domain.Input
	if err := decode(w, r, &input); err != nil {
		return fail(w, err)
	}
	user, err := h.service.Create(ctx, scope, input)
	if err != nil {
		return fail(w, err)
	}
	representation, err := h.representation(r, user)
	if err != nil {
		return fail(w, err)
	}
	w.Header().Set("Location", h.location(r, user.ID))
	return raw(w, http.StatusCreated, representation)
}
func (h *Handlers) User(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := h.authenticate(ctx, r)
	if err != nil {
		return fail(w, err)
	}
	if _, err := parseProjection(r.URL.Query(), false); err != nil {
		return fail(w, err)
	}
	id, err := resourceID(r)
	if err != nil {
		return fail(w, err)
	}
	var user domain.User
	switch r.Method {
	case http.MethodGet:
		user, err = h.service.Get(ctx, scope, id)
	case http.MethodPut:
		var input domain.Input
		if err = decode(w, r, &input); err == nil {
			user, err = h.service.Replace(ctx, scope, id, input)
		}
	case http.MethodPatch:
		var patch domain.Patch
		if err = decode(w, r, &patch); err == nil {
			user, err = h.service.Patch(ctx, scope, id, patch)
		}
	case http.MethodDelete:
		err = h.service.Delete(ctx, scope, id)
		if err == nil {
			return raw(w, http.StatusNoContent, nil)
		}
	}
	if err != nil {
		return fail(w, err)
	}
	representation, err := h.representation(r, user)
	if err != nil {
		return fail(w, err)
	}
	return raw(w, http.StatusOK, representation)
}
