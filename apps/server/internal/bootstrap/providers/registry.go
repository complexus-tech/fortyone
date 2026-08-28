// Package providers composes the first-party integration descriptor registry.
// It contains metadata wiring only; typed factories remain in the API/worker
// bootstrap that owns their concrete dependencies.
package providers

import (
	calendarprovider "github.com/complexus-tech/projects-api/internal/modules/calendar"
	figmaprovider "github.com/complexus-tech/projects-api/internal/modules/figma"
	githubprovider "github.com/complexus-tech/projects-api/internal/modules/github"
	gitlabprovider "github.com/complexus-tech/projects-api/internal/modules/gitlab"
	slackprovider "github.com/complexus-tech/projects-api/internal/modules/slack"
	"github.com/complexus-tech/projects-api/internal/platform/integrations"
)

// BuiltInRegistry returns every provider compiled into this server. Startup
// must fail if descriptors conflict or become invalid.
func BuiltInRegistry() (integrations.Registry, error) {
	return integrations.NewRegistry(
		githubprovider.ProviderDescriptor(),
		gitlabprovider.ProviderDescriptor(),
		slackprovider.ProviderDescriptor(),
		figmaprovider.ProviderDescriptor(),
		calendarprovider.GoogleProviderDescriptor(),
		calendarprovider.MicrosoftProviderDescriptor(),
	)
}
