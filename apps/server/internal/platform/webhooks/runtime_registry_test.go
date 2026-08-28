package webhooks

import (
	"context"
	"errors"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
)

func TestRuntimeRegistryRequiresDeclaredWebhookCapability(t *testing.T) {
	t.Parallel()
	descriptor := testProviderDescriptor()
	descriptor.Capabilities = []integrations.Capability{{
		Key: integrations.CapabilityMessagingDelivery, MajorVersion: 1,
	}}
	catalog, err := integrations.NewRegistry(descriptor)
	if err != nil {
		t.Fatalf("create provider catalog: %v", err)
	}
	_, err = NewRuntimeRegistry(catalog, completeRuntimeRegistration())
	if !errors.Is(err, integrations.ErrCapabilityUnsupported) {
		t.Fatalf("runtime registration error = %v, want unsupported capability", err)
	}
}

func TestRuntimeRegistryRejectsDuplicatesAndIncompleteRuntime(t *testing.T) {
	t.Parallel()
	catalog, err := integrations.NewRegistry(testProviderDescriptor())
	if err != nil {
		t.Fatalf("create provider catalog: %v", err)
	}
	registration := completeRuntimeRegistration()
	if _, err := NewRuntimeRegistry(catalog, registration, registration); err == nil {
		t.Fatal("duplicate runtime registration succeeded")
	}
	registration.Dispatcher = nil
	if _, err := NewRuntimeRegistry(catalog, registration); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("incomplete runtime error = %v, want ErrNotConfigured", err)
	}
}

func TestRuntimeRegistryReturnsStableMissingRuntimeError(t *testing.T) {
	t.Parallel()
	catalog, err := integrations.NewRegistry(testProviderDescriptor())
	if err != nil {
		t.Fatalf("create provider catalog: %v", err)
	}
	registry, err := NewRuntimeRegistry(catalog)
	if err != nil {
		t.Fatalf("create empty runtime registry: %v", err)
	}
	if _, err := registry.require(testProvider); !errors.Is(err, ErrRuntimeNotFound) {
		t.Fatalf("missing runtime error = %v, want ErrRuntimeNotFound", err)
	}
}

func completeRuntimeRegistration() RuntimeRegistration {
	return RuntimeRegistration{
		Provider: testProvider,
		Verifier: verifierFunc(func(context.Context, SignedRequest) (VerifiedDelivery, error) {
			return testVerifiedDelivery(), nil
		}),
		Protector: protectorFunc(func(context.Context, PayloadBinding, []byte) (string, error) {
			return "ciphertext", nil
		}),
		Dispatcher: dispatcherFunc(func(context.Context, Task) error { return nil }),
	}
}
