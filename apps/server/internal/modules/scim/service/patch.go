package scimservice

import (
	"bytes"
	"context"
	"encoding/json"
	"regexp"
	"strings"

	domain "github.com/complexus-tech/projects-api/internal/modules/scim/domain"
	"github.com/google/uuid"
)

func inputFor(user domain.User) (domain.Input, error) {
	var profile domain.Profile
	if err := json.Unmarshal(user.Profile, &profile); err != nil {
		return domain.Input{}, err
	}
	schemas := []string{domain.UserSchema}
	if profile.Enterprise != nil {
		schemas = append(schemas, domain.EnterpriseSchema)
	}
	return domain.Input{Schemas: schemas, UserName: user.UserName, ExternalID: user.ExternalID, Active: &user.Active, DisplayName: profile.DisplayName, Name: profile.Name, Emails: profile.Emails, Enterprise: profile.Enterprise}, nil
}

// Patch merges a bounded set of supported attributes, then commits the complete
// validated profile and membership change once with an optimistic version fence.
func (s *Service) Patch(ctx context.Context, scope domain.Scope, id uuid.UUID, patch domain.Patch) (domain.User, error) {
	if len(patch.Schemas) != 1 || patch.Schemas[0] != domain.PatchSchema || len(patch.Operations) < 1 || len(patch.Operations) > 50 {
		return domain.User{}, invalid("invalid PatchOp schema or operation count")
	}
	old, err := s.repo.Get(ctx, scope, id)
	if err != nil {
		return domain.User{}, err
	}
	input, err := inputFor(old)
	if err != nil {
		return domain.User{}, err
	}
	body, err := json.Marshal(input)
	if err != nil {
		return domain.User{}, err
	}
	var values map[string]json.RawMessage
	if err = json.Unmarshal(body, &values); err != nil {
		return domain.User{}, err
	}
	for _, op := range patch.Operations {
		if err = applyPatch(values, op); err != nil {
			return domain.User{}, err
		}
	}
	body, err = json.Marshal(values)
	if err != nil {
		return domain.User{}, err
	}
	input = domain.Input{}
	if err = decodeKnown(body, &input); err != nil {
		return domain.User{}, err
	}
	if input.Enterprise != nil && len(input.Schemas) == 1 {
		input.Schemas = append(input.Schemas, domain.EnterpriseSchema)
	}
	mutation, err := validate(input, false)
	if err != nil {
		return domain.User{}, err
	}
	user, err := s.repo.Update(ctx, scope, id, old.Version, mutation, false)
	if err == nil {
		s.reconcile(ctx, domain.AdminScope{WorkspaceID: scope.WorkspaceID, ActorID: scope.IssuerID})
	}
	return user, err
}
func decodeKnown(body []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return invalid("unsupported or invalid attribute")
	}
	return nil
}
func canonicalPath(path string) string {
	paths := []string{"userName", "externalId", "active", "displayName", "name", "name.formatted", "name.givenName", "name.familyName", "name.middleName", "name.honorificPrefix", "name.honorificSuffix", "emails", domain.EnterpriseSchema, domain.EnterpriseSchema + ":employeeNumber", domain.EnterpriseSchema + ":costCenter", domain.EnterpriseSchema + ":organization", domain.EnterpriseSchema + ":division", domain.EnterpriseSchema + ":department", domain.EnterpriseSchema + ":manager"}
	for _, candidate := range paths {
		if strings.EqualFold(candidate, path) {
			return candidate
		}
	}
	return ""
}
func applyPatch(values map[string]json.RawMessage, op domain.Operation) error {
	action := strings.ToLower(op.Op)
	if action != "add" && action != "replace" && action != "remove" {
		return invalid("unsupported patch operation")
	}
	if op.Path == "" {
		if action == "remove" {
			return invalid("remove requires an attribute path")
		}
		var object map[string]json.RawMessage
		if json.Unmarshal(op.Value, &object) != nil || object == nil {
			return invalid("pathless patch requires an object")
		}
		for key, value := range object {
			if strings.EqualFold(key, "schemas") {
				continue
			}
			if err := applyPatch(values, domain.Operation{Op: action, Path: key, Value: value}); err != nil {
				return err
			}
		}
		return nil
	}
	if strings.EqualFold(op.Path, "id") || strings.HasPrefix(strings.ToLower(op.Path), "meta") {
		return &domain.Invalid{Type: "mutability", Detail: "Read-only attribute cannot be changed"}
	}
	if matched, err := patchEmails(values, op, action); matched {
		return err
	}
	path := canonicalPath(strings.TrimPrefix(op.Path, domain.UserSchema+":"))
	if path == "" {
		return &domain.Invalid{Type: "invalidPath", Detail: "Unsupported patch attribute path"}
	}
	if action == "remove" && (path == "userName" || path == "active") {
		return invalid("required attribute cannot be removed")
	}
	if action != "remove" && (len(op.Value) == 0 || !json.Valid(op.Value)) {
		return invalid("patch value is required")
	}
	parent, child := "", ""
	if strings.HasPrefix(path, "name.") {
		parent = "name"
		child = strings.TrimPrefix(path, "name.")
	}
	if strings.HasPrefix(path, domain.EnterpriseSchema+":") {
		parent = domain.EnterpriseSchema
		child = strings.TrimPrefix(path, domain.EnterpriseSchema+":")
	}
	if parent != "" {
		object := map[string]json.RawMessage{}
		if len(values[parent]) > 0 {
			if err := json.Unmarshal(values[parent], &object); err != nil {
				return invalid("invalid complex attribute")
			}
		}
		if object == nil {
			object = map[string]json.RawMessage{}
		}
		if action == "remove" {
			delete(object, child)
		} else {
			object[child] = op.Value
		}
		body, err := json.Marshal(object)
		if err != nil {
			return err
		}
		values[parent] = body
		return nil
	}
	if action == "remove" {
		delete(values, path)
	} else if action == "add" && path == "emails" && len(values[path]) > 0 {
		var existing, added []domain.Email
		if json.Unmarshal(values[path], &existing) != nil || json.Unmarshal(op.Value, &added) != nil {
			return invalid("invalid emails value")
		}
		body, err := json.Marshal(append(existing, added...))
		if err != nil {
			return err
		}
		values[path] = body
	} else {
		values[path] = op.Value
	}
	return nil
}

var emailPath = regexp.MustCompile(`(?i)^emails\[(type|primary)\s+eq\s+("[^"\r\n]{1,32}"|true|false)\](?:\.(value|type|primary))?$`)

func patchEmails(values map[string]json.RawMessage, op domain.Operation, action string) (bool, error) {
	if !strings.HasPrefix(strings.ToLower(op.Path), "emails[") {
		return false, nil
	}
	match := emailPath.FindStringSubmatch(op.Path)
	if len(match) != 4 {
		return true, &domain.Invalid{Type: "invalidPath", Detail: "Unsupported email value path"}
	}
	var emails []domain.Email
	if body := values["emails"]; len(body) > 0 {
		if json.Unmarshal(body, &emails) != nil {
			return true, invalid("Invalid email attribute")
		}
	}
	selectedType := ""
	selectedPrimary := false
	if strings.EqualFold(match[1], "type") {
		if json.Unmarshal([]byte(match[2]), &selectedType) != nil {
			return true, invalid("Email type filter requires a string")
		}
	} else {
		if match[2] != "true" && match[2] != "false" {
			return true, invalid("Email primary filter requires a boolean")
		}
		selectedPrimary = match[2] == "true"
	}
	matched := false
	result := make([]domain.Email, 0, len(emails))
	for _, email := range emails {
		selected := strings.EqualFold(email.Type, selectedType)
		if strings.EqualFold(match[1], "primary") {
			selected = email.Primary == selectedPrimary
		}
		if !selected {
			result = append(result, email)
			continue
		}
		matched = true
		if action == "remove" {
			if match[3] != "" && !strings.EqualFold(match[3], "value") {
				switch strings.ToLower(match[3]) {
				case "type":
					email.Type = ""
				case "primary":
					email.Primary = false
				}
				result = append(result, email)
			}
			continue
		}
		switch strings.ToLower(match[3]) {
		case "value":
			if json.Unmarshal(op.Value, &email.Value) != nil {
				return true, invalid("Email value must be a string")
			}
		case "type":
			if json.Unmarshal(op.Value, &email.Type) != nil {
				return true, invalid("Email type must be a string")
			}
		case "primary":
			if json.Unmarshal(op.Value, &email.Primary) != nil {
				return true, invalid("Email primary must be a boolean")
			}
		default:
			if decodeKnown(op.Value, &email) != nil {
				return true, invalid("Email value must be an object")
			}
		}
		result = append(result, email)
	}
	if !matched {
		return true, &domain.Invalid{Type: "noTarget", Detail: "Email path did not match an existing value"}
	}
	body, err := json.Marshal(result)
	if err != nil {
		return true, err
	}
	values["emails"] = body
	return true, nil
}
