package storiesrepository

import (
	"testing"

	"github.com/google/uuid"
)

func TestRewriteStoryMediaHTMLUsesDuplicatedStoryResolver(t *testing.T) {
	originalStoryID := uuid.MustParse("edfcab34-c240-46fa-a455-d708c630fa0b")
	duplicatedStoryID := uuid.MustParse("ca2b0120-488c-4d7f-a920-9d184c312326")
	contentHTML := `<p>Before</p><img src="https://api.test/workspaces/acme/stories/edfcab34-c240-46fa-a455-d708c630fa0b/media/03a5c565-06a3-47a6-89ca-e4f797f3532e"><p>After</p>`

	rewritten := rewriteStoryMediaHTML(&contentHTML, originalStoryID, duplicatedStoryID)
	if rewritten == nil {
		t.Fatal("rewritten story media HTML is nil")
	}
	want := `<p>Before</p><img src="https://api.test/workspaces/acme/stories/ca2b0120-488c-4d7f-a920-9d184c312326/media/03a5c565-06a3-47a6-89ca-e4f797f3532e"><p>After</p>`
	if *rewritten != want {
		t.Fatalf("rewritten HTML = %q, want %q", *rewritten, want)
	}
}

func TestRewriteStoryMediaHTMLPreservesNilContent(t *testing.T) {
	if got := rewriteStoryMediaHTML(nil, uuid.New(), uuid.New()); got != nil {
		t.Fatalf("rewritten nil HTML = %q, want nil", *got)
	}
}
