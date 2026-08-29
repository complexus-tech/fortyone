package api

import (
	"fmt"

	"github.com/complexus-tech/projects-api/internal/eventconsumer"
	"github.com/complexus-tech/projects-api/internal/platform/http/mux"
	"github.com/complexus-tech/projects-api/pkg/mailer"
)

type Runtime struct {
	RouteAdder mux.RouteAdder
	Consumer   *eventconsumer.Consumer
}

func BuildRuntime(
	cfg mux.Config,
	dependencies Dependencies,
	websiteURL string,
	emailService mailer.Service,
) (Runtime, error) {
	if dependencies.DatabasePool == nil {
		return Runtime{}, fmt.Errorf("bootstrap database pool is required")
	}
	if dependencies.VerificationTokens == nil {
		return Runtime{}, fmt.Errorf("bootstrap verification token manager is required")
	}
	if dependencies.InvitationTokens == nil {
		return Runtime{}, fmt.Errorf("bootstrap invitation token manager is required")
	}
	if dependencies.CredentialVault == nil {
		return Runtime{}, fmt.Errorf("bootstrap credential vault is required")
	}
	if dependencies.DeveloperCredentialTokens == nil {
		return Runtime{}, fmt.Errorf("bootstrap developer credential token manager is required")
	}
	if dependencies.DeveloperOAuthPlatform == nil {
		return Runtime{}, fmt.Errorf("bootstrap developer OAuth platform is required")
	}
	if dependencies.DeveloperAPIOAuth == nil {
		return Runtime{}, fmt.Errorf("bootstrap public API OAuth service is required")
	}
	if dependencies.DeveloperOAuthApplications == nil {
		return Runtime{}, fmt.Errorf("bootstrap OAuth application management service is required")
	}
	svcs := buildServices(cfg, dependencies)
	if err := svcs.validate(); err != nil {
		return Runtime{}, fmt.Errorf("bootstrap service validation failed: %w", err)
	}

	streamConsumer := eventconsumer.New(
		cfg.Redis,
		cfg.Log,
		websiteURL,
		svcs.notifications,
		emailService,
		svcs.stories,
		svcs.objectives,
		svcs.users,
		svcs.states,
		svcs.github,
		svcs.feedback,
		cfg.TasksService,
	)

	return Runtime{
		RouteAdder: NewWithServices(svcs),
		Consumer:   streamConsumer,
	}, nil
}
