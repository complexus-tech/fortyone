package objectiveshttp

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	"github.com/complexus-tech/projects-api/pkg/web"
)

func TestAppUpdateObjectivePreservesOmittedNullAndValue(t *testing.T) {
	t.Parallel()

	request := decodeObjectivePatch(t, `{
		"name":"Customer trust",
		"description":null,
		"startDate":"2026-09-01",
		"health":"On Track",
		"expectedUpdatedAt":"2026-08-28T10:30:00Z"
	}`)
	patch := request.ObjectivePatch()

	name, specified := patch.Name.Value()
	if !specified || name == nil || *name != "Customer trust" {
		t.Fatalf("name field = (%v, %v), want concrete value", name, specified)
	}
	description, specified := patch.Description.Value()
	if !specified || description != nil {
		t.Fatalf("description field = (%v, %v), want explicit clear", description, specified)
	}
	if _, specified := patch.ShortSummary.Value(); specified {
		t.Fatal("omitted shortSummary was marked specified")
	}
	startDate, specified := patch.StartDate.Value()
	if !specified || startDate == nil || startDate.Format(time.DateOnly) != "2026-09-01" {
		t.Fatalf("start date field = (%v, %v), want 2026-09-01", startDate, specified)
	}
	health, specified := patch.Health.Value()
	if !specified || health == nil || *health != objectivesdomain.HealthOnTrack {
		t.Fatalf("health field = (%v, %v), want on track", health, specified)
	}
	if request.ExpectedUpdatedAt == nil || request.ExpectedUpdatedAt.Location() != time.UTC {
		t.Fatalf("expectedUpdatedAt = %v, want UTC timestamp", request.ExpectedUpdatedAt)
	}
}

func TestObjectiveUpdateDecodeIsStrict(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{name: "unknown field", body: `{"name":"Launch","unexpected":true}`},
		{name: "trailing object", body: `{"name":"Launch"} {"name":"Overwrite"}`},
		{name: "empty patch", body: `{}`},
		{name: "null name", body: `{"name":null}`},
		{name: "invalid health", body: `{"health":"maybe"}`},
		{name: "invalid date", body: `{"startDate":"09/01/2026"}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			record := httptest.NewRequest("PUT", "/objectives/id", strings.NewReader(test.body))
			record.Header.Set("Content-Type", "application/json")
			var request AppUpdateObjective
			if err := web.Decode(record, &request); err == nil {
				t.Fatal("Decode() error = nil, want strict rejection")
			}
		})
	}
}

func TestObjectiveUpdateRejectsNonJSONContentType(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest("PUT", "/objectives/id", strings.NewReader(`{"name":"Launch"}`))
	request.Header.Set("Content-Type", "text/plain")
	var update AppUpdateObjective
	if err := web.Decode(request, &update); !errors.Is(err, web.ErrInvalidJSONContentType) {
		t.Fatalf("Decode() error = %v, want ErrInvalidJSONContentType", err)
	}
}

func TestStrategicPillarPatchCanClearDescription(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest("PUT", "/pillars/id", strings.NewReader(`{"description":null}`))
	request.Header.Set("Content-Type", "application/json")
	var update AppUpdateStrategicPillar
	if err := web.Decode(request, &update); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	description, specified := update.Patch().Description.Value()
	if !specified || description != nil {
		t.Fatalf("description = (%v, %v), want explicit clear", description, specified)
	}
}

func decodeObjectivePatch(t *testing.T, body string) AppUpdateObjective {
	t.Helper()
	request := httptest.NewRequest("PUT", "/objectives/id", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	var update AppUpdateObjective
	if err := web.Decode(request, &update); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	return update
}
