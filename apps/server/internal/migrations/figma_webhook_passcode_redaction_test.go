package migrations

import (
	"strings"
	"testing"
)

func TestFigmaWebhookPasscodeRedactionMigrationIsForwardOnly(t *testing.T) {
	t.Parallel()

	forward, err := FS.ReadFile("000165_redact_figma_webhook_passcodes.up.sql")
	if err != nil {
		t.Fatalf("read Figma passcode redaction migration: %v", err)
	}
	for _, contract := range []string{
		"SET payload = payload - 'passcode'",
		"WHERE payload ? 'passcode'",
		"figma_webhook_events_payload_no_passcode_check",
		"CHECK (NOT (payload ? 'passcode'))",
	} {
		if !strings.Contains(string(forward), contract) {
			t.Errorf("Figma passcode redaction migration is missing %q", contract)
		}
	}

	rollback, err := FS.ReadFile("000165_redact_figma_webhook_passcodes.down.sql")
	if err != nil {
		t.Fatalf("read Figma passcode redaction rollback: %v", err)
	}
	if !strings.Contains(string(rollback), "cannot be reversed after Figma webhook events exist") {
		t.Fatal("Figma passcode redaction rollback must fail closed after event retention begins")
	}
}
