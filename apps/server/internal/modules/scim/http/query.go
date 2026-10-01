package scimhttp

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	domain "github.com/complexus-tech/projects-api/internal/modules/scim/domain"
)

type queryError struct{ kind, detail string }

func (e *queryError) Error() string      { return e.detail }
func (e *queryError) Unwrap() error      { return domain.ErrInvalidInput }
func badQuery(kind, detail string) error { return &queryError{kind: kind, detail: detail} }

var equalityFilter = regexp.MustCompile(`(?i)^\s*(userName|externalId|id|active)\s+eq\s+(.+)\s*$`)

func parsePage(query url.Values) (domain.Page, error) {
	page := domain.Page{StartIndex: 1, Count: 100}
	if _, err := parseProjection(query, true); err != nil {
		return page, err
	}
	for _, item := range []struct {
		key   string
		value *int
	}{{"startIndex", &page.StartIndex}, {"count", &page.Count}} {
		if query.Has(item.key) {
			number, err := strconv.ParseInt(query.Get(item.key), 10, 32)
			if err != nil {
				return page, badQuery("invalidValue", "Pagination values must be integers")
			}
			*item.value = int(number)
		}
	}
	if page.StartIndex < 1 {
		page.StartIndex = 1
	}
	if page.Count < 0 {
		page.Count = 0
	}
	if page.Count > 100 {
		page.Count = 100
	}
	if filter := query.Get("filter"); filter != "" {
		if len(filter) > 1024 {
			return page, badQuery("invalidFilter", "Filter is too long")
		}
		match := equalityFilter.FindStringSubmatch(filter)
		if len(match) != 3 {
			return page, badQuery("invalidFilter", "Use an equality filter on userName, externalId, id, or active")
		}
		for _, field := range []string{"userName", "externalId", "id", "active"} {
			if strings.EqualFold(field, match[1]) {
				page.Filter.Field = field
			}
		}
		if page.Filter.Field == "active" {
			value, err := strconv.ParseBool(strings.TrimSpace(match[2]))
			if err != nil || (strings.TrimSpace(match[2]) != "true" && strings.TrimSpace(match[2]) != "false") {
				return page, badQuery("invalidFilter", "active filter requires true or false")
			}
			page.Filter.Active = &value
		} else {
			if err := json.Unmarshal([]byte(strings.TrimSpace(match[2])), &page.Filter.Value); err != nil || len(page.Filter.Value) > 255 {
				return page, badQuery("invalidFilter", "Equality filter requires a quoted string")
			}
		}
	}
	return page, nil
}
func parseProjection(query url.Values, list bool) (map[string]bool, error) {
	for key, values := range query {
		allowed := key == "attributes" || key == "excludedAttributes" || (list && (key == "startIndex" || key == "count" || key == "filter"))
		if !allowed || len(values) != 1 {
			return nil, badQuery("invalidValue", "Unsupported or repeated query parameter")
		}
	}
	if query.Has("attributes") && query.Has("excludedAttributes") {
		return nil, badQuery("invalidValue", "Choose attributes or excludedAttributes")
	}
	key := "attributes"
	if query.Has("excludedAttributes") {
		key = "excludedAttributes"
	}
	if len(query.Get(key)) > 2048 {
		return nil, badQuery("invalidValue", "Attribute projection is too long")
	}
	fields := map[string]bool{}
	if !query.Has(key) {
		return fields, nil
	}
	for _, path := range strings.Split(query.Get(key), ",") {
		path = strings.TrimSpace(path)
		canonical := ""
		for _, known := range projectionPaths() {
			if strings.EqualFold(path, known) {
				canonical = known
			}
		}
		if canonical == "" {
			return nil, badQuery("invalidValue", fmt.Sprintf("Unsupported attribute projection: %s", path))
		}
		fields[canonical] = true
	}
	return fields, nil
}

func projectionPaths() []string {
	roots := []string{"schemas", "id", "userName", "externalId", "active", "displayName", "name", "emails", "meta", domain.EnterpriseSchema}
	for parent, children := range map[string][]string{
		"name":   {"formatted", "givenName", "familyName", "middleName", "honorificPrefix", "honorificSuffix"},
		"emails": {"value", "type", "primary"},
		"meta":   {"resourceType", "created", "lastModified", "location"},
	} {
		for _, child := range children {
			roots = append(roots, parent+"."+child)
		}
	}
	for _, child := range []string{"employeeNumber", "costCenter", "organization", "division", "department", "manager", "manager.value", "manager.$ref", "manager.displayName"} {
		roots = append(roots, domain.EnterpriseSchema+":"+child)
	}
	return roots
}
func projectionPath(path string) (string, []string) {
	if strings.HasPrefix(path, domain.EnterpriseSchema+":") {
		return domain.EnterpriseSchema, strings.Split(strings.TrimPrefix(path, domain.EnterpriseSchema+":"), ".")
	}
	pieces := strings.Split(path, ".")
	return pieces[0], pieces[1:]
}
func copyProjected(source any, paths [][]string, exclude bool) any {
	switch value := source.(type) {
	case []any:
		result := make([]any, 0, len(value))
		for _, item := range value {
			result = append(result, copyProjected(item, paths, exclude))
		}
		return result
	case map[string]any:
		result := map[string]any{}
		for key, item := range value {
			direct := false
			children := [][]string{}
			for _, path := range paths {
				if len(path) > 0 && path[0] == key {
					if len(path) == 1 {
						direct = true
					} else {
						children = append(children, path[1:])
					}
				}
			}
			if exclude {
				if direct {
					continue
				}
				if len(children) > 0 {
					result[key] = copyProjected(item, children, true)
				} else {
					result[key] = item
				}
			} else {
				if direct {
					result[key] = item
				} else if len(children) > 0 {
					result[key] = copyProjected(item, children, false)
				}
			}
		}
		return result
	default:
		return source
	}
}
func project(resource map[string]any, query url.Values) (map[string]any, error) {
	fields, err := parseProjection(query, query.Has("filter") || query.Has("startIndex") || query.Has("count"))
	if err != nil {
		return nil, err
	}
	if !query.Has("attributes") && !query.Has("excludedAttributes") {
		return resource, nil
	}
	exclude := query.Has("excludedAttributes")
	for key, value := range resource {
		if key == "schemas" || key == "id" {
			continue
		}
		if fields[key] {
			if exclude {
				delete(resource, key)
			}
			continue
		}
		paths := [][]string{}
		for path := range fields {
			parent, children := projectionPath(path)
			if parent == key && len(children) > 0 {
				paths = append(paths, children)
			}
		}
		if len(paths) == 0 {
			if !exclude {
				delete(resource, key)
			}
			continue
		}
		// Convert typed profile structs to JSON values before selecting nested fields.
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		var raw any
		if err = json.Unmarshal(encoded, &raw); err != nil {
			return nil, err
		}
		resource[key] = copyProjected(raw, paths, exclude)
	}
	return resource, nil
}
