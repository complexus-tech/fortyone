package integrations

import (
	"errors"
	"testing"
	"time"
)

const repositoryCatalogV1 CapabilityKey = "codehost.repository_catalog"

func TestRegistryIsDeterministicAndDefensivelyCopied(t *testing.T) {
	t.Parallel()

	github := validDescriptor("github", "GitHub")
	gitlab := validDescriptor("gitlab", "GitLab")
	registry, err := NewRegistry(gitlab, github)
	if err != nil {
		t.Fatalf("construct registry: %v", err)
	}

	providers := registry.List()
	if len(providers) != 2 || providers[0].Key != "github" || providers[1].Key != "gitlab" {
		t.Fatalf("provider order = %#v", providers)
	}

	providers[0].Capabilities[0].MajorVersion = 99
	githubAgain, found := registry.Get("github")
	if !found {
		t.Fatal("GitHub descriptor not found")
	}
	if githubAgain.Capabilities[0].MajorVersion != 1 {
		t.Fatal("caller mutated registry capability metadata")
	}
}

func TestRegistryCapabilityErrorsDistinguishProviderAndVersion(t *testing.T) {
	t.Parallel()

	registry, err := NewRegistry(validDescriptor("github", "GitHub"))
	if err != nil {
		t.Fatalf("construct registry: %v", err)
	}

	if err := registry.RequireCapability("unknown", Capability{Key: repositoryCatalogV1, MajorVersion: 1}); !errors.Is(err, ErrProviderNotFound) {
		t.Fatalf("unknown provider error = %v", err)
	}
	if err := registry.RequireCapability("github", Capability{Key: repositoryCatalogV1, MajorVersion: 2}); !errors.Is(err, ErrCapabilityUnsupported) {
		t.Fatalf("incompatible version error = %v", err)
	}
	if err := registry.RequireCapability("github", Capability{Key: repositoryCatalogV1, MajorVersion: 1}); err != nil {
		t.Fatalf("supported capability: %v", err)
	}
}

func TestRegistryRejectsAmbiguousOrUnsafeMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		descriptor Descriptor
	}{
		{name: "invalid provider key", descriptor: mutate(validDescriptor("github", "GitHub"), func(value *Descriptor) { value.Key = "Git Hub" })},
		{name: "missing display name", descriptor: mutate(validDescriptor("github", "GitHub"), func(value *Descriptor) { value.DisplayName = " " })},
		{name: "unknown family", descriptor: mutate(validDescriptor("github", "GitHub"), func(value *Descriptor) { value.Family = "everything" })},
		{name: "no capabilities", descriptor: mutate(validDescriptor("github", "GitHub"), func(value *Descriptor) { value.Capabilities = nil })},
		{name: "no auth strategies", descriptor: mutate(validDescriptor("github", "GitHub"), func(value *Descriptor) { value.AuthStrategies = nil })},
		{name: "zero capability version", descriptor: mutate(validDescriptor("github", "GitHub"), func(value *Descriptor) { value.Capabilities[0].MajorVersion = 0 })},
		{name: "duplicate capability", descriptor: mutate(validDescriptor("github", "GitHub"), func(value *Descriptor) { value.Capabilities = append(value.Capabilities, value.Capabilities[0]) })},
		{name: "unknown auth strategy", descriptor: mutate(validDescriptor("github", "GitHub"), func(value *Descriptor) { value.AuthStrategies = []AuthStrategy{"password"} })},
		{name: "invalid config key", descriptor: mutate(validDescriptor("github", "GitHub"), func(value *Descriptor) { value.Configuration[0].Key = "github.secret" })},
		{name: "missing config purpose", descriptor: mutate(validDescriptor("github", "GitHub"), func(value *Descriptor) { value.Configuration[0].Purpose = "" })},
		{name: "unsafe runbook", descriptor: mutate(validDescriptor("github", "GitHub"), func(value *Descriptor) { value.OperatorRunbook = "../../secrets" })},
		{name: "retention without metadata", descriptor: mutate(validDescriptor("github", "GitHub"), func(value *Descriptor) {
			value.Disconnect.RetainMappingMetadata = false
			value.Disconnect.MappingRetentionPeriod = time.Hour
		})},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := NewRegistry(test.descriptor); err == nil {
				t.Fatal("expected invalid descriptor error")
			}
		})
	}

	github := validDescriptor("github", "GitHub")
	if _, err := NewRegistry(github, github); err == nil {
		t.Fatal("expected duplicate provider error")
	}
}

func validDescriptor(key ProviderKey, displayName string) Descriptor {
	return Descriptor{
		Key:         key,
		DisplayName: displayName,
		Family:      FamilyCodeHost,
		Capabilities: []Capability{
			{Key: repositoryCatalogV1, MajorVersion: 1},
		},
		AuthStrategies: []AuthStrategy{AuthStrategyAppInstallation},
		Configuration: []ConfigurationRequirement{
			{Key: "GITHUB_CLIENT_SECRET", Required: true, Sensitive: true, Purpose: "OAuth client authentication"},
		},
		Disconnect: DisconnectPolicy{
			RevokeRemoteGrant:      true,
			DeleteWebhook:          true,
			DeleteCredentials:      true,
			RetainMappingMetadata:  true,
			MappingRetentionPeriod: 30 * 24 * time.Hour,
		},
		OperatorRunbook: "docs/integrations/code-hosts.md",
	}
}

func mutate(descriptor Descriptor, change func(*Descriptor)) Descriptor {
	descriptor = descriptor.clone()
	change(&descriptor)
	return descriptor
}
