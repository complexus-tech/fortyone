package api

import (
	"fmt"

	"github.com/complexus-tech/projects-api/internal/platform/http/mux"
	"github.com/complexus-tech/projects-api/pkg/consumer"
	"github.com/complexus-tech/projects-api/pkg/mailer"
)

type Runtime struct {
	RouteAdder mux.RouteAdder
	Consumer   *consumer.Consumer
}

func BuildRuntime(cfg mux.Config, websiteURL string, emailService mailer.Service) (Runtime, error) {
	svcs := buildServices(cfg)
	if err := svcs.validate(); err != nil {
		return Runtime{}, fmt.Errorf("bootstrap service validation failed: %w", err)
	}

	streamConsumer := consumer.New(
		cfg.Redis,
		cfg.DB,
		cfg.Log,
		websiteURL,
		svcs.notifications,
		emailService,
		svcs.stories,
		svcs.objectives,
		svcs.users,
		svcs.states,
	)

	return Runtime{
		RouteAdder: NewWithServices(svcs),
		Consumer:   streamConsumer,
	}, nil
}
