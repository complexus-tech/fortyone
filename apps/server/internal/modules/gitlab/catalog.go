package gitlab

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/complexus-tech/projects-api/internal/platform/codehost"
	"github.com/complexus-tech/projects-api/internal/platform/integrations"
)

type projectResponse struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
	WebURL            string `json:"web_url"`
	DefaultBranch     string `json:"default_branch"`
	Visibility        string `json:"visibility"`
	Archived          bool   `json:"archived"`
	Namespace         struct {
		FullPath string `json:"full_path"`
	} `json:"namespace"`
}

func (adapter *Adapter) Provider() integrations.ProviderKey { return ProviderKey }

func (adapter *Adapter) Capabilities() codehost.Capabilities {
	return codehost.Capabilities{
		codehost.CapabilityInstallationAuth:     true,
		codehost.CapabilityRepositoryCatalog:    true,
		codehost.CapabilityWorkItemWriter:       true,
		codehost.CapabilityCommentWriter:        true,
		codehost.CapabilityWebhookNormalization: true,
	}
}

func (adapter *Adapter) Authorize(ctx context.Context, installation codehost.InstallationRef) error {
	_, err := adapter.do(ctx, installation, http.MethodGet, "user", nil, nil, nil)
	return err
}

func (adapter *Adapter) ListRepositories(
	ctx context.Context,
	installation codehost.InstallationRef,
	cursor codehost.Cursor,
) (codehost.RepositoryPage, error) {
	if err := codehost.ValidateCursor(cursor); err != nil {
		return codehost.RepositoryPage{}, err
	}
	query := url.Values{
		"membership": {"true"},
		"simple":     {"true"},
		"pagination": {"keyset"},
		"order_by":   {"id"},
		"sort":       {"asc"},
		"per_page":   {strconv.Itoa(cursor.Limit)},
	}
	if strings.TrimSpace(cursor.Value) != "" {
		idAfter, parseErr := strconv.ParseInt(cursor.Value, 10, 64)
		if parseErr != nil || idAfter <= 0 {
			return codehost.RepositoryPage{}, errors.Join(codehost.ErrInvalidInput, errors.New("GitLab project cursor is invalid"))
		}
		query.Set("id_after", strconv.FormatInt(idAfter, 10))
	}
	var response []projectResponse
	headers, err := adapter.do(ctx, installation, http.MethodGet, "projects", query, nil, &response)
	if err != nil {
		return codehost.RepositoryPage{}, err
	}
	repositories := make([]codehost.RepositoryRef, 0, len(response))
	for _, project := range response {
		if project.ID <= 0 || strings.TrimSpace(project.PathWithNamespace) == "" {
			continue
		}
		repositories = append(repositories, mapRepository(project))
	}
	nextCursor, err := nextProjectCursor(headers.Get("Link"))
	if err != nil {
		return codehost.RepositoryPage{}, err
	}
	return codehost.RepositoryPage{
		Repositories: repositories,
		NextCursor:   nextCursor,
	}, nil
}

// GitLab project keyset pagination returns the next position in the Link URL
// as id_after. Keep only that validated scalar so callers cannot steer the
// adapter to a provider-supplied or caller-supplied URL.
func nextProjectCursor(linkHeader string) (string, error) {
	linkHeader = strings.TrimSpace(linkHeader)
	if linkHeader == "" {
		return "", nil
	}
	for _, segment := range strings.Split(linkHeader, ",") {
		start := strings.IndexByte(segment, '<')
		end := strings.IndexByte(segment, '>')
		if start < 0 || end <= start || !linkHasRelation(segment[end+1:], "next") {
			continue
		}
		nextURL, err := url.Parse(strings.TrimSpace(segment[start+1 : end]))
		if err != nil {
			return "", fmt.Errorf("parse GitLab project continuation: %w", err)
		}
		value := strings.TrimSpace(nextURL.Query().Get("id_after"))
		idAfter, err := strconv.ParseInt(value, 10, 64)
		if err != nil || idAfter <= 0 {
			return "", errors.New("GitLab project continuation is missing a valid id_after value")
		}
		return strconv.FormatInt(idAfter, 10), nil
	}
	return "", nil
}

func linkHasRelation(parameters, expected string) bool {
	for _, parameter := range strings.Split(parameters, ";") {
		key, value, ok := strings.Cut(strings.TrimSpace(parameter), "=")
		if !ok || !strings.EqualFold(strings.TrimSpace(key), "rel") {
			continue
		}
		for _, relation := range strings.Fields(strings.Trim(strings.TrimSpace(value), `"`)) {
			if strings.EqualFold(relation, expected) {
				return true
			}
		}
	}
	return false
}

func mapRepository(project projectResponse) codehost.RepositoryRef {
	owner := strings.TrimSpace(project.Namespace.FullPath)
	if owner == "" {
		parts := strings.Split(project.PathWithNamespace, "/")
		owner = strings.Join(parts[:max(1, len(parts)-1)], "/")
	}
	return codehost.RepositoryRef{
		ExternalID:    strconv.FormatInt(project.ID, 10),
		Owner:         owner,
		Name:          project.Name,
		FullName:      project.PathWithNamespace,
		WebURL:        project.WebURL,
		DefaultBranch: project.DefaultBranch,
		Private:       project.Visibility != "public",
		Archived:      project.Archived,
	}
}

var (
	_ codehost.Adapter                   = (*Adapter)(nil)
	_ codehost.InstallationAuthenticator = (*Adapter)(nil)
	_ codehost.RepositoryCatalog         = (*Adapter)(nil)
)
