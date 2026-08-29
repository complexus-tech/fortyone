package usershttp

import (
	"context"
	"errors"

	"github.com/complexus-tech/projects-api/pkg/cache"
)

// takeOAuthState consumes current state atomically and falls back to the
// read-only legacy namespace for callbacks started before deployment.
func takeOAuthState(
	ctx context.Context,
	service *cache.Service,
	currentKey string,
	legacyKey string,
	destination any,
) error {
	err := service.Take(ctx, currentKey, destination)
	if errors.Is(err, cache.ErrNotFound) {
		return service.Take(ctx, legacyKey, destination)
	}
	return err
}
