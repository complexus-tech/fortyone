package documentshttp

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
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

func TestToAppDocumentSummaryOmitsDetailFields(t *testing.T) {
	t.Parallel()

	summary := toAppDocumentSummaries([]documents.CoreDocumentSummary{{
		ID:               uuid.New(),
		WorkspaceID:      uuid.New(),
		Title:            "Project brief",
		Visibility:       documents.VisibilityWorkspace,
		CreatedBy:        uuid.New(),
		UpdatedBy:        uuid.New(),
		CanEdit:          true,
		RelatedWorkCount: 2,
	}}, true)

	payload, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("marshal document summary: %v", err)
	}
	var documentsPayload []map[string]json.RawMessage
	if err := json.Unmarshal(payload, &documentsPayload); err != nil {
		t.Fatalf("unmarshal document summary: %v", err)
	}
	if len(documentsPayload) != 1 {
		t.Fatalf("expected one document summary, got %d", len(documentsPayload))
	}
	for _, field := range []string{"contentHtml", "contentText", "sharedWith", "relatedWork"} {
		if _, exists := documentsPayload[0][field]; exists {
			t.Fatalf("document summary must not serialize %q", field)
		}
	}
	for _, field := range []string{
		"id", "workspaceId", "title", "visibility", "createdBy", "updatedBy",
		"createdAt", "updatedAt", "canEdit", "relatedWorkCount",
	} {
		if _, exists := documentsPayload[0][field]; !exists {
			t.Fatalf("document summary must serialize %q", field)
		}
	}
}

func TestDocumentListLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		query   string
		want    *int
		wantErr bool
	}{
		{name: "omitted", query: "", want: nil},
		{name: "positive", query: "?limit=12", want: intPointer(12)},
		{name: "zero", query: "?limit=0", wantErr: true},
		{name: "negative", query: "?limit=-1", wantErr: true},
		{name: "not a number", query: "?limit=recent", wantErr: true},
		{name: "above maximum", query: "?limit=101", wantErr: true},
		{name: "repeated", query: "?limit=10&limit=12", wantErr: true},
		{name: "blank", query: "?limit=", wantErr: true},
		{name: "overflow", query: "?limit=999999999999999999999999", wantErr: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request := httptest.NewRequest("GET", "/documents"+test.query, nil)
			got, err := documentListLimit(request)
			if test.wantErr {
				if err == nil {
					t.Fatal("expected an invalid limit error")
				}
				return
			}
			if err != nil {
				t.Fatalf("parse document list limit: %v", err)
			}
			if test.want == nil {
				if got != nil {
					t.Fatalf("expected no limit, got %d", *got)
				}
				return
			}
			if got == nil || *got != *test.want {
				t.Fatalf("expected limit %d, got %v", *test.want, got)
			}
		})
	}
}

func TestDocumentListQueryIsBoundedAndUnambiguous(t *testing.T) {
	t.Parallel()

	t.Run("trims values", func(t *testing.T) {
		t.Parallel()
		request := httptest.NewRequest("GET", "/documents?search=%20roadmap%20&scope=%20mine%20&limit=20", nil)
		query, err := documentListQuery(request)
		if err != nil {
			t.Fatalf("parse document query: %v", err)
		}
		if query.search != "roadmap" || query.scope != "mine" || query.limit == nil || *query.limit != 20 {
			t.Fatalf("unexpected document query: %#v", query)
		}
	})

	for name, path := range map[string]string{
		"repeated search":  "/documents?search=one&search=two",
		"repeated scope":   "/documents?scope=mine&scope=all",
		"oversized search": "/documents?search=" + strings.Repeat("x", documentSearchMaximumBytes+1),
	} {
		name, path := name, path
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			request := httptest.NewRequest("GET", path, nil)
			if _, err := documentListQuery(request); err == nil {
				t.Fatal("expected an invalid document query")
			}
		})
	}
}

func intPointer(value int) *int {
	return &value
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
