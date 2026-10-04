package workpresets

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/workpresets/domain"
)

func TestViewMetadataValidation(t *testing.T) {
	for _, test := range []struct {
		name        string
		description *string
		icon        *string
		valid       bool
	}{
		{"legacy absent metadata", nil, nil, true},
		{"Unicode description at limit", metadataString(strings.Repeat("🗓", 2000)), metadataString("calendar"), true},
		{"too long description", metadataString(strings.Repeat("a", 2001)), nil, false},
		{"untrusted icon URL", nil, metadataString("https://example.com/icon.svg"), false},
		{"untrusted icon markup", nil, metadataString("<svg>"), false},
		{"unknown icon", nil, metadataString("unknown"), false},
	} {
		t.Run(test.name, func(t *testing.T) {
			value := domain.ViewConfiguration{Version: 1, Layout: "list", Description: test.description, Icon: test.icon, ViewOptions: domain.ViewOptions{GroupBy: "status", OrderBy: "created", OrderDirection: "desc"}}
			raw, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			err = ValidateConfiguration(domain.View, raw)
			if test.valid && err != nil {
				t.Fatalf("metadata rejected: %v", err)
			}
			if !test.valid && !errors.Is(err, domain.ErrInvalidInput) {
				t.Fatalf("invalid metadata accepted: %v", err)
			}
		})
	}
	for _, icon := range []string{"list", "kanban", "calendar", "clock", "star", "analytics", "goal", "tags", "code", "team", "book", "workflow", "roadmap", "docs"} {
		value := domain.ViewConfiguration{Version: 1, Layout: "kanban", Icon: &icon, ViewOptions: domain.ViewOptions{GroupBy: "status", OrderBy: "created", OrderDirection: "desc"}}
		raw, _ := json.Marshal(value)
		if err := ValidateConfiguration(domain.View, raw); err != nil {
			t.Fatalf("supported icon %s rejected: %v", icon, err)
		}
	}
}

func metadataString(value string) *string { return &value }

func TestMyWorkViewScopeValidation(t *testing.T) {
	for _, test := range []struct {
		name, scope string
		valid       bool
	}{
		{"relative upcoming", `{"kind":"my-work","tab":"upcoming"}`, true},
		{"full scope", `{"kind":"my-work","tab":"collaborating","category":"paused","overdue":false,"createdAfter":"2026-10-01","createdBefore":"2026-10-31"}`, true},
		{"nullable optional state", `{"kind":"my-work","tab":"all","category":null,"createdAfter":null,"createdBefore":null}`, true},
		{"unknown scope kind", `{"kind":"team","tab":"all"}`, false},
		{"unknown tab", `{"kind":"my-work","tab":"archived"}`, false},
		{"unknown category", `{"kind":"my-work","tab":"all","category":"deleted"}`, false},
		{"invalid date", `{"kind":"my-work","tab":"all","createdBefore":"2026-02-31"}`, false},
		{"unexpected nested field", `{"kind":"my-work","tab":"all","userId":"private-value"}`, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			raw := json.RawMessage(`{"version":1,"layout":"list","filters":{},"viewOptions":{"groupBy":"status","orderBy":"created","orderDirection":"desc","displayColumns":[]},"scope":` + test.scope + `}`)
			err := ValidateConfiguration(domain.View, raw)
			if test.valid && err != nil {
				t.Fatalf("valid scope rejected: %v", err)
			}
			if !test.valid && !errors.Is(err, domain.ErrInvalidInput) {
				t.Fatalf("invalid scope accepted: %v", err)
			}
		})
	}
}
