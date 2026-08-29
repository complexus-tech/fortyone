package webhooks

import (
	"fmt"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
)

type RuntimeRegistration struct {
	Provider   integrations.ProviderKey
	Verifier   WebhookVerifier
	Protector  PayloadProtector
	Dispatcher Dispatcher
}

type RuntimeRegistry struct {
	runtimes map[integrations.ProviderKey]RuntimeRegistration
}

func NewRuntimeRegistry(catalog integrations.Registry, registrations ...RuntimeRegistration) (RuntimeRegistry, error) {
	runtimes := make(map[integrations.ProviderKey]RuntimeRegistration, len(registrations))
	capability := integrations.Capability{
		Key:          integrations.CapabilityWebhookVerification,
		MajorVersion: 1,
	}
	for _, registration := range registrations {
		if err := catalog.RequireCapability(registration.Provider, capability); err != nil {
			return RuntimeRegistry{}, fmt.Errorf("register webhook runtime: %w", err)
		}
		if registration.Verifier == nil || registration.Protector == nil || registration.Dispatcher == nil {
			return RuntimeRegistry{}, fmt.Errorf("register webhook runtime %q: %w", registration.Provider, ErrNotConfigured)
		}
		if _, exists := runtimes[registration.Provider]; exists {
			return RuntimeRegistry{}, fmt.Errorf("register webhook runtime %q more than once", registration.Provider)
		}
		runtimes[registration.Provider] = registration
	}
	return RuntimeRegistry{runtimes: runtimes}, nil
}

func (registry RuntimeRegistry) require(provider integrations.ProviderKey) (RuntimeRegistration, error) {
	runtime, ok := registry.runtimes[provider]
	if !ok {
		return RuntimeRegistration{}, ErrRuntimeNotFound
	}
	return runtime, nil
}
