package idempotency

import (
	"strings"
	"testing"

	"github.com/complexus-tech/projects-api/internal/migrations"
)

func TestReceiptMigrationStoresOnlyBoundedReplayMetadata(t *testing.T) {
	t.Parallel()

	source, err := migrations.FS.ReadFile("000156_api_idempotency_receipts.up.sql")
	if err != nil {
		t.Fatalf("read receipt migration: %v", err)
	}
	migration := strings.ToLower(string(source))
	for _, required := range []string{
		"key_digest bytea not null",
		"request_hash bytea not null",
		"octet_length(key_digest) = 32",
		"octet_length(request_hash) = 32",
		"octet_length(response_body) <= 65536",
		"unique nulls not distinct",
		"lease_generation > 0",
		"response_content_type",
	} {
		if !strings.Contains(migration, required) {
			t.Errorf("receipt migration is missing %q", required)
		}
	}
	for _, forbidden := range []string{
		"raw_key",
		"request_body",
		"response_headers",
		"set_cookie",
		"set-cookie",
	} {
		if strings.Contains(migration, forbidden) {
			t.Errorf("receipt migration contains forbidden replay material %q", forbidden)
		}
	}
}

func TestReceiptDownMigrationRefusesDataLoss(t *testing.T) {
	t.Parallel()

	source, err := migrations.FS.ReadFile("000156_api_idempotency_receipts.down.sql")
	if err != nil {
		t.Fatalf("read receipt rollback: %v", err)
	}
	migration := strings.ToLower(string(source))
	if !strings.Contains(migration, "if exists (select 1 from public.api_idempotency_receipts)") ||
		!strings.Contains(migration, "cannot be reversed while idempotency receipts exist") {
		t.Fatal("receipt rollback must guard retained receipt data")
	}
}
