package workpresets

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/workpresets/domain"
)

func TestConfigurationRejectsUnknownAndUnboundedState(t *testing.T) {
	tests := []struct {
		name string
		kind domain.Kind
		raw  string
	}{
		{"unknown schema version", domain.Template, `{"version":2,"title":"Bug","priority":"High","checklist":[]}`},
		{"credential-bearing arbitrary key", domain.Template, `{"version":1,"title":"Bug","priority":"High","checklist":[],"apiToken":"private-value"}`},
		{"unknown nested filter", domain.View, `{"version":1,"layout":"list","filters":{"accessToken":"private-value"},"viewOptions":{"groupBy":"status","orderBy":"created","orderDirection":"desc","displayColumns":[]}}`},
		{"invalid layout", domain.View, `{"version":1,"layout":"html","filters":{},"viewOptions":{"groupBy":"status","orderBy":"created","orderDirection":"desc","displayColumns":[]}}`},
		{"duplicate custom field", domain.Template, `{"version":1,"priority":"High","customFieldValues":[{"fieldId":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","value":"1"},{"fieldId":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","value":"2"}]}`},
		{"invalid date", domain.View, `{"version":1,"layout":"list","filters":{"endDate":"2026-02-31"},"viewOptions":{"groupBy":"status","orderBy":"created","orderDirection":"desc","displayColumns":[]}}`},
		{"oversize body", domain.Template, strings.Repeat(" ", maximumConfigurationBytes+1)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateConfiguration(test.kind, json.RawMessage(test.raw))
			if !errors.Is(err, domain.ErrInvalidInput) {
				t.Fatalf("got %v", err)
			}
			if strings.Contains(err.Error(), "private-value") {
				t.Fatal("validation reflected a secret")
			}
		})
	}
}

func TestConfigurationPreservesTypedSnapshots(t *testing.T) {
	tests := []struct {
		kind domain.Kind
		raw  string
	}{
		{domain.View, `{"version":1,"layout":"kanban","filters":{"assignedToMe":true,"operators":{"assigneeIds":"isAnyOf"}},"viewOptions":{"groupBy":"assignee","orderBy":"deadline","orderDirection":"asc","displayColumns":["Status","Labels"],"selectedCustomFieldIds":["aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"]}}`},
		{domain.Template, `{"version":1,"title":"Incident review","description":"Review checklist","priority":"High","checklist":["Collect logs","Write review"],"customFieldValues":[{"fieldId":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","value":null}]}`},
	}
	for _, test := range tests {
		if err := ValidateConfiguration(test.kind, json.RawMessage(test.raw)); err != nil {
			t.Fatal(err)
		}
	}
}
