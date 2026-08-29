package outboundwebhooksservice

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
)

func TestSecretManagerShowsOnceAndBindsEnvelopeContext(t *testing.T) {
	t.Parallel()
	vault, err := credentialvault.New(credentialvault.Config{
		Active: credentialvault.KeyRef{ID: "test", Version: 1},
		Keys: []credentialvault.Key{{
			Ref:      credentialvault.KeyRef{ID: "test", Version: 1},
			Material: bytes.Repeat([]byte{0x42}, 32),
		}},
	})
	if err != nil {
		t.Fatalf("create vault: %v", err)
	}
	manager, err := newSecretManager(vault, bytes.NewReader(bytes.Repeat([]byte{0x24}, webhookSecretBytes)))
	if err != nil {
		t.Fatalf("create secret manager: %v", err)
	}
	workspaceID, endpointID := uuid.New(), uuid.New()
	secret, envelope, err := manager.Generate(workspaceID, endpointID, 1)
	if err != nil {
		t.Fatalf("generate secret: %v", err)
	}
	if !strings.HasPrefix(secret.Reveal(), webhookSecretPrefix) || secret.String() != "[REDACTED]" {
		t.Fatalf("secret format/redaction is invalid")
	}
	encoded, err := json.Marshal(secret)
	if err != nil {
		t.Fatalf("marshal secret: %v", err)
	}
	if bytes.Contains(encoded, []byte(secret.Reveal())) {
		t.Fatalf("JSON serialization exposed plaintext secret")
	}
	opened, err := manager.Open(workspaceID, endpointID, 1, envelope)
	if err != nil {
		t.Fatalf("open secret: %v", err)
	}
	defer clear(opened)
	if string(opened) != secret.Reveal() {
		t.Fatalf("opened secret differs from issued secret")
	}
	if _, err := manager.Open(uuid.New(), endpointID, 1, envelope); err == nil {
		t.Fatal("cross-workspace envelope opened")
	}
}

func TestDecodeSigningSecretRejectsWrongLengthAndAlphabet(t *testing.T) {
	t.Parallel()
	for _, value := range []string{
		"secret_" + base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, 32)),
		webhookSecretPrefix + "not-base64",
		webhookSecretPrefix + base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, 16)),
	} {
		if _, err := decodeSigningSecret([]byte(value)); err == nil {
			t.Fatalf("decodeSigningSecret(%q) succeeded", value)
		}
	}
}
