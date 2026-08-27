package linksrepository

import (
	"os"
	"strings"
	"testing"
)

func TestLinkMutationQueriesScopeWorkspaceAndRejectNoOp(t *testing.T) {
	data, err := os.ReadFile("commands.go")
	if err != nil {
		t.Fatalf("read commands.go: %v", err)
	}
	source := string(data)

	for _, contract := range []struct {
		name  string
		start string
		end   string
		parts []string
	}{
		{name: "create", start: "func (r *repo) CreateLink(", end: "func (r *repo) UpdateLink(", parts: []string{"FROM stories s", "s.workspace_id = :workspace_id", "links.ErrNotFound"}},
		{name: "update", start: "func (r *repo) UpdateLink(", end: "func (r *repo) DeleteLink(", parts: []string{"s.workspace_id = :workspace_id", "result.RowsAffected()", "links.ErrNotFound"}},
		{name: "delete", start: "func (r *repo) DeleteLink(", end: "", parts: []string{"USING stories s", "s.workspace_id = :workspace_id", "result.RowsAffected()", "links.ErrNotFound"}},
	} {
		t.Run(contract.name, func(t *testing.T) {
			body := sourceFromMarkers(t, source, contract.start, contract.end)
			for _, part := range contract.parts {
				if !strings.Contains(body, part) {
					t.Fatalf("%s mutation is missing %q", contract.name, part)
				}
			}
		})
	}
}

func sourceFromMarkers(t *testing.T, source, start, end string) string {
	t.Helper()
	startIndex := strings.Index(source, start)
	if startIndex < 0 {
		t.Fatalf("source is missing %q", start)
	}
	if end == "" {
		return source[startIndex:]
	}
	endIndex := strings.Index(source[startIndex+len(start):], end)
	if endIndex < 0 {
		t.Fatalf("source after %q is missing %q", start, end)
	}
	return source[startIndex : startIndex+len(start)+endIndex]
}
