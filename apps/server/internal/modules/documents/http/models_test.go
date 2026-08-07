package documentshttp

import (
	"testing"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	"github.com/google/uuid"
)

func TestToAppDocumentRequiresWorkspaceMutationPermission(t *testing.T) {
	t.Parallel()

	document := documents.CoreDocument{CanEdit: true}

	if got := toAppDocument(document, true); !got.CanEdit {
		t.Fatal("expected an editable document for a workspace member")
	}
	if got := toAppDocument(document, false); got.CanEdit {
		t.Fatal("expected a read-only document for a workspace guest")
	}
}

func TestDocumentMediaUsesStableResolverURL(t *testing.T) {
	t.Parallel()

	documentID := uuid.MustParse("66db0798-2eef-4dad-bb35-413612ab0fd1")
	attachmentID := uuid.MustParse("f124a762-a767-446c-bbd1-0b3f43dce115")
	stableURL := documentMediaURL("product and design", documentID, attachmentID)
	media := toAppDocumentMedia(attachments.FileInfo{
		ID:       attachmentID,
		Filename: "brief.png",
		MimeType: "image/png",
	}, stableURL)

	if media.URL != "/workspaces/product%20and%20design/documents/66db0798-2eef-4dad-bb35-413612ab0fd1/media/f124a762-a767-446c-bbd1-0b3f43dce115" {
		t.Fatalf("unexpected stable media URL: %s", media.URL)
	}
	if media.URL == "" || media.MimeType != "image/png" {
		t.Fatalf("unexpected media response: %#v", media)
	}
}
