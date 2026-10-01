//go:build integration

package customfieldsrepository_test

import (
	"math/big"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/customfields/domain"
	"github.com/google/uuid"
)

func TestReportsPreserveAbsentAggregatesAndExactValuedZero(t *testing.T) {
	f := seed(t)
	amount := create(t, f, domain.Definition{Name: "Deal amount", Type: domain.Money, Currency: text("USD")})
	valuedStatus, emptyStatus := uuid.New(), uuid.New()
	for _, statusID := range []uuid.UUID{valuedStatus, emptyStatus} {
		exec(t, f.pool, `
			INSERT INTO statuses (status_id, name, category, workspace_id, team_id)
			VALUES ($1, $2, 'started', $3, $4)
		`, statusID, statusID.String(), f.scope.WorkspaceID, f.scope.TeamID)
	}
	exec(t, f.pool, `UPDATE stories SET status_id = $1 WHERE id = $2 OR id = $3`, valuedStatus, f.story, f.secondStory)
	exec(t, f.pool, `
		INSERT INTO stories (id, workspace_id, team_id, title, status_id)
		VALUES ($1, $2, $3, 'Deal without an amount', $4)
	`, uuid.New(), f.scope.WorkspaceID, f.scope.TeamID, emptyStatus)
	for _, entry := range []struct {
		storyID uuid.UUID
		value   string
	}{{f.story, "1250.50"}, {f.secondStory, "249.50"}} {
		if _, err := f.repository.Patch(t.Context(), f.scope, entry.storyID, domain.ValuePatch{
			Values: []domain.Value{{FieldID: amount.ID, Value: text(entry.value)}},
		}); err != nil {
			t.Fatalf("store deal amount: %v", err)
		}
	}
	assertReport := func(aggregation, expected string, valuedCount int64) {
		t.Helper()
		report, err := f.repository.Report(t.Context(), f.scope, domain.ReportInput{
			FieldID: amount.ID, Aggregation: aggregation, GroupBy: "status", DateBasis: "created",
		})
		if err != nil {
			t.Fatalf("build %s report: %v", aggregation, err)
		}
		if report.TotalCount != 3 || report.ValuedCount != valuedCount || report.MissingCount != 3-valuedCount || len(report.Rows) != 2 {
			t.Fatalf("%s coverage = %#v", aggregation, report)
		}
		for _, row := range report.Rows {
			value := expected
			count := valuedCount
			if row.Key == emptyStatus.String() {
				count = 0
				value = ""
				if aggregation == "sum" || aggregation == "count" {
					value = "0"
				}
			}
			if row.Count != count {
				t.Errorf("%s row count = %d, want %d", aggregation, row.Count, count)
			}
			if value == "" {
				if row.Value != "" {
					t.Errorf("%s empty group value = %q, want absent string", aggregation, row.Value)
				}
				continue
			}
			got, valid := new(big.Rat).SetString(row.Value)
			want, _ := new(big.Rat).SetString(value)
			if !valid || got.Cmp(want) != 0 {
				t.Errorf("%s exact aggregate = %q, want %s", aggregation, row.Value, value)
			}
		}
	}
	assertReport("average", "750.00", 2)
	assertReport("min", "249.50", 2)
	assertReport("max", "1250.50", 2)
	for _, storyID := range []uuid.UUID{f.story, f.secondStory} {
		if _, err := f.repository.Patch(t.Context(), f.scope, storyID, domain.ValuePatch{
			Values: []domain.Value{{FieldID: amount.ID, Value: nil}},
		}); err != nil {
			t.Fatalf("clear deal amount: %v", err)
		}
	}
	for _, aggregation := range []string{"average", "min", "max"} {
		assertReport(aggregation, "", 0)
	}
	for _, aggregation := range []string{"sum", "count"} {
		assertReport(aggregation, "0", 0)
	}
	if _, err := f.repository.Patch(t.Context(), f.scope, f.story, domain.ValuePatch{
		Values: []domain.Value{{FieldID: amount.ID, Value: text("0")}},
	}); err != nil {
		t.Fatalf("store actual zero: %v", err)
	}
	assertReport("average", "0", 1)
}
