package scimhttp

import (
	"context"
	"net/http"
	"strings"

	domain "github.com/complexus-tech/projects-api/internal/modules/scim/domain"
	"github.com/complexus-tech/projects-api/pkg/web"
)

func attribute(name, kind string, multi, required bool) map[string]any {
	return map[string]any{"name": name, "type": kind, "multiValued": multi, "required": required, "mutability": "readWrite", "returned": "default", "uniqueness": "none"}
}
func userSchema() map[string]any {
	username := attribute("userName", "string", false, true)
	username["uniqueness"] = "server"
	username["caseExact"] = false
	name := attribute("name", "complex", false, false)
	name["subAttributes"] = []any{attribute("formatted", "string", false, false), attribute("givenName", "string", false, false), attribute("familyName", "string", false, false), attribute("middleName", "string", false, false), attribute("honorificPrefix", "string", false, false), attribute("honorificSuffix", "string", false, false)}
	email := attribute("emails", "complex", true, false)
	email["subAttributes"] = []any{attribute("value", "string", false, true), attribute("type", "string", false, false), attribute("primary", "boolean", false, false)}
	return map[string]any{"schemas": []string{"urn:ietf:params:scim:schemas:core:2.0:Schema"}, "id": domain.UserSchema, "name": "User", "description": "Workspace members", "attributes": []any{username, attribute("externalId", "string", false, false), attribute("active", "boolean", false, false), attribute("displayName", "string", false, false), name, email}}
}
func enterpriseSchema() map[string]any {
	manager := attribute("manager", "complex", false, false)
	reference := attribute("$ref", "reference", false, false)
	reference["referenceTypes"] = []string{"User"}
	display := attribute("displayName", "string", false, false)
	display["mutability"] = "readOnly"
	manager["subAttributes"] = []any{attribute("value", "string", false, false), reference, display}
	return map[string]any{"schemas": []string{"urn:ietf:params:scim:schemas:core:2.0:Schema"}, "id": domain.EnterpriseSchema, "name": "EnterpriseUser", "description": "Tenant profile metadata", "attributes": []any{attribute("employeeNumber", "string", false, false), attribute("costCenter", "string", false, false), attribute("organization", "string", false, false), attribute("division", "string", false, false), attribute("department", "string", false, false), manager}}
}
func resourceType() map[string]any {
	return map[string]any{"schemas": []string{"urn:ietf:params:scim:schemas:core:2.0:ResourceType"}, "id": "User", "name": "User", "endpoint": "/Users", "schema": domain.UserSchema, "schemaExtensions": []any{map[string]any{"schema": domain.EnterpriseSchema, "required": false}}}
}
func (h *Handlers) Discovery(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if _, err := h.authenticate(ctx, r); err != nil {
		return fail(w, err)
	}
	if len(r.URL.Query()) > 0 {
		return fail(w, domain.ErrInvalidInput)
	}
	path := r.URL.Path
	id := web.Params(r, "resourceId")
	switch {
	case strings.HasSuffix(path, "/ServiceProviderConfig"):
		return raw(w, http.StatusOK, map[string]any{"schemas": []string{"urn:ietf:params:scim:schemas:core:2.0:ServiceProviderConfig"}, "patch": map[string]any{"supported": true}, "bulk": map[string]any{"supported": false, "maxOperations": 0, "maxPayloadSize": 0}, "filter": map[string]any{"supported": true, "maxResults": 100}, "changePassword": map[string]any{"supported": false}, "sort": map[string]any{"supported": false}, "etag": map[string]any{"supported": false}, "authenticationSchemes": []any{map[string]any{"type": "oauthbearertoken", "name": "Workspace bearer token", "description": "Scoped SCIM credential", "primary": true}}})
	case strings.Contains(path, "/ResourceTypes"):
		if id != "" {
			if id != "User" {
				return fail(w, domain.ErrNotFound)
			}
			return raw(w, http.StatusOK, resourceType())
		}
		return raw(w, http.StatusOK, map[string]any{"schemas": []string{domain.ListSchema}, "totalResults": 1, "startIndex": 1, "itemsPerPage": 1, "Resources": []any{resourceType()}})
	case strings.Contains(path, "/Schemas"):
		if id != "" {
			switch id {
			case domain.UserSchema:
				return raw(w, http.StatusOK, userSchema())
			case domain.EnterpriseSchema:
				return raw(w, http.StatusOK, enterpriseSchema())
			default:
				return fail(w, domain.ErrNotFound)
			}
		}
		return raw(w, http.StatusOK, map[string]any{"schemas": []string{domain.ListSchema}, "totalResults": 2, "startIndex": 1, "itemsPerPage": 2, "Resources": []any{userSchema(), enterpriseSchema()}})
	}
	return fail(w, domain.ErrNotFound)
}
