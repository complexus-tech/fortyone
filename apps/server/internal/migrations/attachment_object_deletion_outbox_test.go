package migrations

import (
	"strings"
	"testing"
)

func TestAttachmentObjectDeletionOutboxMigrationIsDurableAndCredentialFree(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile("000172_attachment_object_deletion_outbox.up.sql")
	if err != nil {
		t.Fatalf("read attachment deletion outbox migration: %v", err)
	}
	migration := strings.ToLower(string(data))
	for _, contract := range []string{
		"create table public.attachment_object_deletion_outbox",
		"storage_provider varchar(32) not null",
		"container_name varchar(255) not null",
		"blob_name varchar(1024) not null",
		"completed_at timestamptz",
		"claim_token uuid",
		"unique (attachment_id)",
		"status in ('pending', 'processing', 'retrying', 'completed')",
		"attachment_object_deletion_outbox_lifecycle_check",
		"idx_attachment_object_deletion_outbox_due",
		"where status in ('pending', 'retrying')",
		"idx_attachment_object_deletion_outbox_lease",
		"where status = 'processing'",
		"idx_attachment_object_deletion_outbox_completed",
		"where status = 'completed'",
		"application logs, traces, metrics, and errors must not emit this value",
	} {
		if !strings.Contains(migration, contract) {
			t.Errorf("attachment deletion outbox migration is missing contract %q", contract)
		}
	}
	for _, forbidden := range []string{
		"foreign key (",
		"references public.",
		"access_key",
		"account_key",
		"connection_string",
		"credential_payload",
		"client_secret",
		"password",
	} {
		if strings.Contains(migration, forbidden) {
			t.Errorf("attachment deletion outbox migration contains forbidden material %q", forbidden)
		}
	}
}

func TestAttachmentObjectDeletionOutboxRollbackRefusesAnyDeliveryState(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile("000172_attachment_object_deletion_outbox.down.sql")
	if err != nil {
		t.Fatalf("read attachment deletion outbox rollback: %v", err)
	}
	migration := strings.ToLower(string(data))
	for _, contract := range []string{
		"migration 000172 is forward-only",
		"select 1",
		"from public.attachment_object_deletion_outbox",
		"raise exception",
		"using errcode = '55000'",
	} {
		if !strings.Contains(migration, contract) {
			t.Errorf("attachment deletion outbox rollback is missing contract %q", contract)
		}
	}
	if strings.Index(migration, "raise exception") > strings.Index(migration, "drop table") {
		t.Fatal("attachment deletion outbox rollback must guard before dropping the table")
	}
}
