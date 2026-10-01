//go:build integration

package customfieldsrepository_test

import (
	"encoding/json"
	"errors"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/customfields/domain"
)

func TestIconsPersistPreserveOmittedAndAuditExplicitReset(t *testing.T) {
	f := seed(t)
	field := create(t, f, domain.Definition{Name: "Amount", Type: domain.Number, Icon: text("star")})
	if field.Icon == nil || *field.Icon != "star" {
		t.Fatal("create icon was not persisted")
	}
	unchanged, err := f.repository.Update(t.Context(), f.scope, field.ID, domain.Definition{Name: "Renamed", Type: domain.Number})
	if err != nil || unchanged.Icon == nil || *unchanged.Icon != "star" {
		t.Fatal("legacy definition update cleared icon", err)
	}
	reset, err := f.repository.Update(t.Context(), f.scope, field.ID, domain.Definition{Name: "Renamed", IconSet: true})
	if err != nil || reset.Icon != nil {
		t.Fatal("explicit null did not reset icon", err)
	}
	var metadata []byte
	if err := f.pool.QueryRow(t.Context(), `SELECT metadata FROM custom_field_definition_audit WHERE field_id=$1 AND operation='custom_field.updated' AND metadata -> 'newIcon' = 'null' ORDER BY created_at DESC,id DESC LIMIT 1`, field.ID).Scan(&metadata); err != nil {
		t.Fatal(err)
	}
	var event struct{ OldIcon, NewIcon *string }
	if err := json.Unmarshal(metadata, &event); err != nil || event.OldIcon == nil || *event.OldIcon != "star" || event.NewIcon != nil {
		t.Fatal("icon reset audit lost previous selection", string(metadata), err)
	}
	member := f.scope
	member.ActorID = f.member
	if _, err := f.repository.Update(t.Context(), member, field.ID, domain.Definition{Name: "Renamed", IconSet: true, Icon: text("globe")}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatal("member changed definition icon", err)
	}
	if _, err := f.repository.Update(t.Context(), f.scope, field.ID, domain.Definition{Name: "Renamed", IconSet: true, Icon: text("unknown")}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatal("unknown icon accepted", err)
	}
	if _, err := f.pool.Exec(t.Context(), `UPDATE custom_field_definition_audit SET metadata='{}' WHERE field_id=$1`, field.ID); err == nil {
		t.Fatal("definition audit was mutable")
	}
	if err := f.repository.Archive(t.Context(), f.scope, field.ID); err != nil {
		t.Fatal(err)
	}
	if err := f.repository.Archive(t.Context(), f.scope, field.ID); err != nil {
		t.Fatal("archive retry was not idempotent", err)
	}
	var count int
	if err := f.pool.QueryRow(t.Context(), `SELECT COUNT(*) FROM custom_field_definition_audit WHERE field_id=$1`, field.ID).Scan(&count); err != nil || count != 4 {
		t.Fatal("unauthorized/invalid mutation or archive retry wrote audit", count, err)
	}
}

func TestExpandedIconsPersistAcrossDefinitionSnapshotsAndExactReports(t *testing.T) {
	f := seed(t)
	field := create(t, f, domain.Definition{Name: "Amount", Type: domain.Number, Icon: text("money")})
	if _, err := f.repository.Patch(t.Context(), f.scope, f.story, domain.ValuePatch{Values: []domain.Value{{FieldID: field.ID, Value: text("1250.50")}}}); err != nil {
		t.Fatal(err)
	}
	for _, icon := range []string{
		"tag", "pin", "objective", "strategy", "roadmap", "home", "health", "approval",
		"chat", "comment", "share", "image", "video", "microphone", "book", "help", "info",
		"analytics", "dashboard", "workflow", "kanban", "sprint", "automation", "history", "lock", "key", "code", "warning",
		"money", "coins", "wallet", "credit-card", "bank", "invoice", "percent", "calculator", "piggy-bank", "target-money",
		"building", "megaphone", "store", "package", "phone", "handshake", "shopping-cart",
	} {
		t.Run(icon, func(t *testing.T) {
			updated, err := f.repository.Update(t.Context(), f.scope, field.ID, domain.Definition{Name: field.Name, IconSet: true, Icon: &icon})
			if err != nil || updated.Icon == nil || *updated.Icon != icon {
				t.Fatalf("expanded definition icon = %#v, %v", updated, err)
			}
			snapshot, err := f.repository.Snapshot(t.Context(), f.scope, f.story)
			if err != nil || len(snapshot.Fields) != 1 || snapshot.Fields[0].Icon == nil || *snapshot.Fields[0].Icon != icon || len(snapshot.Values) != 1 || snapshot.Values[0].Value == nil || *snapshot.Values[0].Value != "1250.50" {
				t.Fatalf("icon edit changed the value snapshot = %#v, %v", snapshot, err)
			}
			report, err := f.repository.Report(t.Context(), f.scope, domain.ReportInput{FieldID: field.ID, Aggregation: "sum", GroupBy: "none", DateBasis: "created"})
			if err != nil || report.Field.Icon == nil || *report.Field.Icon != icon || len(report.Rows) != 1 || report.Rows[0].Value != "1250.50" {
				t.Fatalf("expanded icon exact report = %#v, %v", report, err)
			}
		})
	}
	if _, err := f.pool.Exec(t.Context(), `UPDATE custom_fields SET icon='unknown' WHERE id=$1`, field.ID); err == nil {
		t.Fatal("database accepted an icon outside the catalog")
	}
}
