package workspaceshttp

import (
	"errors"
	"net/url"

	"github.com/complexus-tech/projects-api/pkg/web"
)

const maximumWorkspaceSlugLength = 255

func parseSlugAvailabilityQuery(values url.Values) (string, error) {
	slug, present, err := web.OptionalTextQueryParameter(
		values, "slug", maximumWorkspaceSlugLength, maximumWorkspaceSlugLength,
	)
	if err != nil {
		return "", err
	}
	if !present || len(slug) < 3 {
		return "", errors.New("slug must be between 3 and 255 characters")
	}
	if !slugPattern.MatchString(slug) {
		return "", errors.New("slug can only contain lowercase letters, numbers, and hyphens")
	}
	return slug, nil
}
