package migrations

import (
	"strings"
	"testing"
)

func TestAttachmentThumbnailRetirementPreservesOriginalAndDerivativeIdentity(t *testing.T) {
	data, err := FS.ReadFile("000233_attachment_thumbnails.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(data)
	for _, expected := range []string{
		"attachment_id uuid NOT NULL REFERENCES public.attachments(attachment_id) ON DELETE CASCADE",
		"workspace_id uuid NOT NULL REFERENCES public.workspaces(workspace_id) ON DELETE CASCADE",
		"UNIQUE (attachment_id, source_version)",
		"AFTER DELETE ON public.attachment_thumbnails",
		"VALUES (OLD.thumbnail_id, OLD.workspace_id",
		"ON CONFLICT (attachment_id) DO NOTHING",
	} {
		if !strings.Contains(sql, expected) {
			t.Errorf("missing durable thumbnail lifecycle contract %q", expected)
		}
	}
	if strings.Contains(sql, "VALUES (OLD.attachment_id,") {
		t.Fatal("derivative cleanup must not overwrite original attachment cleanup")
	}
}
