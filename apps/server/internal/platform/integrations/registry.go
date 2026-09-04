package integrations

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var (
	ErrProviderNotFound      = errors.New("integration provider is not registered")
	ErrCapabilityUnsupported = errors.New("integration capability is not supported")

	providerKeyPattern   = regexp.MustCompile(`^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$`)
	capabilityKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:[._-][a-z0-9]+)*$`)
	configKeyPattern     = regexp.MustCompile(`^[A-Z][A-Z0-9]*(?:_[A-Z0-9]+)+$`)
)

// Registry is an immutable catalog constructed explicitly by bootstrap.
type Registry struct {
	providers    map[ProviderKey]Descriptor
	capabilities map[ProviderKey]map[CapabilityKey]uint16
	orderedKeys  []ProviderKey
}

// NewRegistry validates and copies every descriptor. Duplicate identifiers or
// ambiguous capability declarations fail startup rather than using last-write
// wins behavior.
func NewRegistry(descriptors ...Descriptor) (Registry, error) {
	registry := Registry{
		providers:    make(map[ProviderKey]Descriptor, len(descriptors)),
		capabilities: make(map[ProviderKey]map[CapabilityKey]uint16, len(descriptors)),
		orderedKeys:  make([]ProviderKey, 0, len(descriptors)),
	}

	for _, descriptor := range descriptors {
		if err := validateDescriptor(descriptor); err != nil {
			return Registry{}, err
		}
		if _, exists := registry.providers[descriptor.Key]; exists {
			return Registry{}, fmt.Errorf("duplicate integration provider %q", descriptor.Key)
		}

		capabilities := make(map[CapabilityKey]uint16, len(descriptor.Capabilities))
		for _, capability := range descriptor.Capabilities {
			capabilities[capability.Key] = capability.MajorVersion
		}

		registry.providers[descriptor.Key] = descriptor.clone()
		registry.capabilities[descriptor.Key] = capabilities
		registry.orderedKeys = append(registry.orderedKeys, descriptor.Key)
	}

	sort.Slice(registry.orderedKeys, func(left, right int) bool {
		return registry.orderedKeys[left] < registry.orderedKeys[right]
	})

	return registry, nil
}

// Get returns a defensive copy of a registered descriptor.
func (registry Registry) Get(key ProviderKey) (Descriptor, bool) {
	descriptor, found := registry.providers[key]
	return descriptor.clone(), found
}

// Require returns a provider or a stable not-found classification.
func (registry Registry) Require(key ProviderKey) (Descriptor, error) {
	descriptor, found := registry.Get(key)
	if !found {
		return Descriptor{}, fmt.Errorf("%w: %s", ErrProviderNotFound, key)
	}
	return descriptor, nil
}

// List returns providers in stable key order.
func (registry Registry) List() []Descriptor {
	descriptors := make([]Descriptor, 0, len(registry.orderedKeys))
	for _, key := range registry.orderedKeys {
		descriptor, _ := registry.Get(key)
		descriptors = append(descriptors, descriptor)
	}
	return descriptors
}

// Supports reports whether the provider implements exactly the requested major
// contract version. Callers must not silently use another major version.
func (registry Registry) Supports(provider ProviderKey, capability Capability) bool {
	versions, found := registry.capabilities[provider]
	if !found {
		return false
	}
	version, found := versions[capability.Key]
	return found && version == capability.MajorVersion
}

// RequireCapability distinguishes an unknown provider from an unsupported or
// incompatible capability.
func (registry Registry) RequireCapability(provider ProviderKey, capability Capability) error {
	if _, found := registry.providers[provider]; !found {
		return fmt.Errorf("%w: %s", ErrProviderNotFound, provider)
	}
	if !registry.Supports(provider, capability) {
		return fmt.Errorf(
			"%w: provider=%s capability=%s version=%d",
			ErrCapabilityUnsupported,
			provider,
			capability.Key,
			capability.MajorVersion,
		)
	}
	return nil
}

func validateDescriptor(descriptor Descriptor) error {
	if !providerKeyPattern.MatchString(string(descriptor.Key)) {
		return fmt.Errorf("integration provider key %q is invalid", descriptor.Key)
	}
	if strings.TrimSpace(descriptor.DisplayName) == "" {
		return fmt.Errorf("integration provider %q display name is required", descriptor.Key)
	}
	if !validFamily(descriptor.Family) {
		return fmt.Errorf("integration provider %q has invalid family %q", descriptor.Key, descriptor.Family)
	}
	if len(descriptor.Capabilities) == 0 {
		return fmt.Errorf("integration provider %q must declare at least one capability", descriptor.Key)
	}
	if len(descriptor.AuthStrategies) == 0 {
		return fmt.Errorf("integration provider %q must declare at least one auth strategy", descriptor.Key)
	}
	if runbook := strings.TrimSpace(descriptor.OperatorRunbook); runbook == "" || strings.HasPrefix(runbook, "/") || strings.Contains(runbook, "..") {
		return fmt.Errorf("integration provider %q operator runbook must be a safe repository-relative path", descriptor.Key)
	}

	capabilities := make(map[CapabilityKey]struct{}, len(descriptor.Capabilities))
	for _, capability := range descriptor.Capabilities {
		if !capabilityKeyPattern.MatchString(string(capability.Key)) {
			return fmt.Errorf("integration provider %q capability key %q is invalid", descriptor.Key, capability.Key)
		}
		if capability.MajorVersion == 0 {
			return fmt.Errorf("integration provider %q capability %q version must be positive", descriptor.Key, capability.Key)
		}
		if _, exists := capabilities[capability.Key]; exists {
			return fmt.Errorf("integration provider %q declares capability %q more than once", descriptor.Key, capability.Key)
		}
		capabilities[capability.Key] = struct{}{}
	}

	strategies := make(map[AuthStrategy]struct{}, len(descriptor.AuthStrategies))
	for _, strategy := range descriptor.AuthStrategies {
		if !validAuthStrategy(strategy) {
			return fmt.Errorf("integration provider %q has invalid auth strategy %q", descriptor.Key, strategy)
		}
		if _, exists := strategies[strategy]; exists {
			return fmt.Errorf("integration provider %q declares auth strategy %q more than once", descriptor.Key, strategy)
		}
		strategies[strategy] = struct{}{}
	}

	configuration := make(map[string]struct{}, len(descriptor.Configuration))
	for _, requirement := range descriptor.Configuration {
		if !configKeyPattern.MatchString(requirement.Key) {
			return fmt.Errorf("integration provider %q configuration key %q is invalid", descriptor.Key, requirement.Key)
		}
		if strings.TrimSpace(requirement.Purpose) == "" {
			return fmt.Errorf("integration provider %q configuration %q purpose is required", descriptor.Key, requirement.Key)
		}
		if _, exists := configuration[requirement.Key]; exists {
			return fmt.Errorf("integration provider %q declares configuration %q more than once", descriptor.Key, requirement.Key)
		}
		configuration[requirement.Key] = struct{}{}
	}

	if descriptor.Disconnect.MappingRetentionPeriod < 0 {
		return fmt.Errorf("integration provider %q mapping retention cannot be negative", descriptor.Key)
	}
	if !descriptor.Disconnect.RetainMappingMetadata && descriptor.Disconnect.MappingRetentionPeriod != 0 {
		return fmt.Errorf("integration provider %q cannot retain mappings for a duration when retention is disabled", descriptor.Key)
	}

	return nil
}

func validFamily(family Family) bool {
	switch family {
	case FamilyCodeHost, FamilyMessaging, FamilyCalendar, FamilySupportFeedback, FamilyDesignContext, FamilyCloudContent:
		return true
	default:
		return false
	}
}

func validAuthStrategy(strategy AuthStrategy) bool {
	switch strategy {
	case AuthStrategyAppInstallation, AuthStrategyOAuthInstall, AuthStrategyOAuthLink, AuthStrategyWebhookOnly:
		return true
	default:
		return false
	}
}
