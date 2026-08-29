package webhooks

import (
	"context"
	"errors"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/google/uuid"
)

func TestBoundPayloadCodecRejectsIdentitySubstitution(t *testing.T) {
	t.Parallel()
	provider := integrations.ProviderKey("gitlab")
	codec, err := NewBoundPayloadCodec(provider, "gitlab-webhook.v1.", "0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatalf("NewBoundPayloadCodec() error = %v", err)
	}
	binding := PayloadBinding{
		Provider:               provider,
		DeliveryID:             "delivery-1",
		WorkspaceID:            uuid.New(),
		InstallationID:         uuid.New(),
		InstallationGeneration: uuid.New(),
	}
	sealed, err := codec.Seal(context.Background(), binding, []byte(`{"object_kind":"issue"}`))
	if err != nil {
		t.Fatalf("Seal() error = %v", err)
	}
	record := Record{Envelope: Envelope{
		Provider:               binding.Provider,
		DeliveryID:             binding.DeliveryID,
		WorkspaceID:            binding.WorkspaceID,
		InstallationID:         binding.InstallationID,
		InstallationGeneration: binding.InstallationGeneration,
	}}
	opened, err := codec.Open(record, sealed)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if string(opened) != `{"object_kind":"issue"}` {
		t.Fatalf("Open() = %q", opened)
	}

	record.InstallationGeneration = uuid.New()
	if _, err := codec.Open(record, sealed); err == nil {
		t.Fatal("Open() accepted ciphertext under a different installation generation")
	}
	if _, err := codec.Open(Record{Envelope: Envelope{Provider: "github"}}, sealed); err == nil {
		t.Fatal("Open() accepted ciphertext under a different provider")
	}
}

func TestBoundPayloadCodecValidatesConfigurationAndContext(t *testing.T) {
	t.Parallel()
	if _, err := NewBoundPayloadCodec("", "prefix.", "secret"); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("NewBoundPayloadCodec() error = %v, want ErrNotConfigured", err)
	}
	codec, err := NewBoundPayloadCodec("gitlab", "prefix.", "0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatalf("NewBoundPayloadCodec() error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := codec.Seal(ctx, PayloadBinding{}, []byte("payload")); !errors.Is(err, context.Canceled) {
		t.Fatalf("Seal() error = %v, want context.Canceled", err)
	}
}
