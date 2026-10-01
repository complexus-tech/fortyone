//go:build integration

package customfieldsrepository_test

import (
	"errors"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/bootstrap/customfieldsadapter"
	domain "github.com/complexus-tech/projects-api/internal/modules/customfields/domain"
	repository "github.com/complexus-tech/projects-api/internal/modules/customfields/repository"
	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	auth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type fixture struct {
	pool                    *pgxpool.Pool
	repository              *repository.Repository
	scope                   domain.Scope
	member, guest, inactive uuid.UUID
	story, secondStory      uuid.UUID
}

func seed(t *testing.T) fixture {
	t.Helper()
	postgres := testkit.NewPostgres(t)
	f := fixture{pool: postgres.Pool, repository: repository.New(postgres.Pool), scope: domain.Scope{ActorID: uuid.New(), WorkspaceID: uuid.New(), TeamID: uuid.New()}, member: uuid.New(), guest: uuid.New(), inactive: uuid.New(), story: uuid.New(), secondStory: uuid.New()}
	exec(t, f.pool, `INSERT INTO workspaces (workspace_id,name,slug) VALUES ($1,'Custom fields',$2)`, f.scope.WorkspaceID, uuid.NewString())
	exec(t, f.pool, `INSERT INTO teams (team_id,workspace_id,name,code,color,is_private) VALUES ($1,$2,'Private sales','SALE','#000000',TRUE)`, f.scope.TeamID, f.scope.WorkspaceID)
	for _, entry := range []struct {
		id     uuid.UUID
		role   string
		active bool
	}{{f.scope.ActorID, "admin", true}, {f.member, "member", true}, {f.guest, "guest", true}, {f.inactive, "member", false}} {
		exec(t, f.pool, `INSERT INTO users (user_id,username,email,full_name,is_active) VALUES ($1,$2,$3,$4,$5)`, entry.id, entry.id.String(), entry.id.String()+"@example.test", "Person "+entry.role, entry.active)
		exec(t, f.pool, `INSERT INTO workspace_members (workspace_id,user_id,role) VALUES ($1,$2,CAST($3 AS user_role))`, f.scope.WorkspaceID, entry.id, entry.role)
		exec(t, f.pool, `INSERT INTO team_members (team_id,user_id) VALUES ($1,$2)`, f.scope.TeamID, entry.id)
	}
	for index, id := range []uuid.UUID{f.story, f.secondStory} {
		exec(t, f.pool, `INSERT INTO stories (id,workspace_id,team_id,title,created_at,completed_at) VALUES ($1,$2,$3,'Deal',CAST($4 AS timestamptz),CAST($5 AS timestamptz))`, id, f.scope.WorkspaceID, f.scope.TeamID, time.Date(2026, 1, index+1, 9, 0, 0, 0, time.UTC), time.Date(2026, 2, index+1, 9, 0, 0, 0, time.UTC))
	}
	return f
}

func exec(t *testing.T, pool *pgxpool.Pool, query string, args ...any) {
	t.Helper()
	if _, err := pool.Exec(t.Context(), query, args...); err != nil {
		t.Fatalf("fixture SQL: %v", err)
	}
}
func text(value string) *string { return &value }
func create(t *testing.T, f fixture, input domain.Definition) domain.Field {
	t.Helper()
	field, err := f.repository.Create(t.Context(), f.scope, input)
	if err != nil {
		t.Fatalf("create field: %v", err)
	}
	return field
}
func auditCount(t *testing.T, f fixture, storyID uuid.UUID) int64 {
	t.Helper()
	var count int64
	if err := f.pool.QueryRow(t.Context(), `SELECT COUNT(*) FROM story_custom_field_audit WHERE story_id=$1`, storyID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func TestCustomFieldsExactValuesAtomicAuditAndActorFences(t *testing.T) {
	f := seed(t)
	amount := create(t, f, domain.Definition{Name: "Deal value", Type: domain.Money, Currency: text("USD")})
	stage := create(t, f, domain.Definition{Name: "Stage", Type: domain.Select, Options: []domain.OptionInput{{Name: "Won"}, {Name: "Open"}}})
	owner := create(t, f, domain.Definition{Name: "Owner", Type: domain.Person})
	closeDate := create(t, f, domain.Definition{Name: "Close date", Type: domain.Date})
	memberScope := f.scope
	memberScope.ActorID = f.member
	guestScope := f.scope
	guestScope.ActorID = f.guest
	if _, err := f.repository.Create(t.Context(), memberScope, domain.Definition{Name: "Unsafe", Type: domain.Text}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("member definition create = %v", err)
	}
	if _, err := f.repository.Patch(t.Context(), guestScope, f.story, domain.ValuePatch{Values: []domain.Value{{FieldID: amount.ID, Value: text("1")}}}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("guest write = %v", err)
	}
	initial, err := f.repository.Patch(t.Context(), memberScope, f.story, domain.ValuePatch{Values: []domain.Value{{FieldID: amount.ID, Value: text("9007199254740993.01")}, {FieldID: stage.ID, Value: text(stage.Options[0].ID.String())}, {FieldID: owner.ID, Value: text(f.member.String())}, {FieldID: closeDate.ID, Value: text("2026-03-01")}}})
	if err != nil || initial.Version != 1 || len(initial.Values) != 4 {
		t.Fatalf("initial values = %#v, %v", initial, err)
	}
	if auditCount(t, f, f.story) != 4 {
		t.Fatal("initial audit was not atomic")
	}
	if _, err := f.repository.Patch(t.Context(), memberScope, f.story, domain.ValuePatch{Values: []domain.Value{{FieldID: amount.ID, Value: text("1")}, {FieldID: stage.ID, Value: text(uuid.NewString())}}}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("invalid option = %v", err)
	}
	current, err := f.repository.Snapshot(t.Context(), guestScope, f.story)
	if err != nil || current.Version != 1 || auditCount(t, f, f.story) != 4 {
		t.Fatalf("failed patch changed state = %#v,%v", current, err)
	}
	stale := int64(0)
	if _, err := f.repository.Patch(t.Context(), f.scope, f.story, domain.ValuePatch{ExpectedVersion: &stale, Values: []domain.Value{{FieldID: amount.ID, Value: text("0")}}}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("stale version = %v", err)
	}
	if _, err := f.repository.Patch(t.Context(), f.scope, f.story, domain.ValuePatch{Values: []domain.Value{{FieldID: owner.ID, Value: text(f.inactive.String())}}}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("inactive person = %v", err)
	}
	if _, err := f.repository.Patch(t.Context(), f.scope, f.secondStory, domain.ValuePatch{Values: []domain.Value{{FieldID: amount.ID, Value: text("0")}}}); err != nil {
		t.Fatal(err)
	}
	report, err := f.repository.Report(t.Context(), f.scope, domain.ReportInput{FieldID: amount.ID, Aggregation: "sum", GroupBy: "none", DateBasis: "created"})
	if err != nil || report.TotalCount != 2 || report.ValuedCount != 2 || len(report.Rows) != 1 || report.Rows[0].Value != "9007199254740993.01" || report.Currency == nil || *report.Currency != "USD" {
		t.Fatalf("exact report = %#v, %v", report, err)
	}
	if _, err := f.repository.Patch(t.Context(), f.scope, f.secondStory, domain.ValuePatch{Values: []domain.Value{{FieldID: amount.ID, Value: nil}}}); err != nil {
		t.Fatal(err)
	}
	report, err = f.repository.Report(t.Context(), f.scope, domain.ReportInput{FieldID: amount.ID, Aggregation: "count", GroupBy: "none", DateBasis: "created"})
	if err != nil || report.ValuedCount != 1 || report.MissingCount != 1 || report.Rows[0].Value != "1" {
		t.Fatalf("missing vs zero report = %#v, %v", report, err)
	}
	start, end := "2026-03-01", "2026-03-31"
	report, err = f.repository.Report(t.Context(), f.scope, domain.ReportInput{FieldID: amount.ID, Aggregation: "sum", GroupBy: "month", DateBasis: closeDate.ID.String(), StartDate: &start, EndDate: &end})
	if err != nil || report.TotalCount != 1 || report.Rows[0].Key != "2026-03" {
		t.Fatalf("custom date basis = %#v, %v", report, err)
	}
	foreign := f.scope
	foreign.WorkspaceID = uuid.New()
	if _, err := f.repository.Report(t.Context(), foreign, domain.ReportInput{FieldID: amount.ID, Aggregation: "sum", GroupBy: "none", DateBasis: "created"}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("foreign report = %v", err)
	}
	outsider := uuid.New()
	exec(t, f.pool, `INSERT INTO users (user_id,username,email,full_name,is_active) VALUES ($1,$2,$3,'Outsider',TRUE)`, outsider, outsider.String(), outsider.String()+"@example.test")
	exec(t, f.pool, `INSERT INTO workspace_members (workspace_id,user_id,role) VALUES ($1,$2,'member')`, f.scope.WorkspaceID, outsider)
	outsideScope := f.scope
	outsideScope.ActorID = outsider
	if _, err := f.repository.Snapshot(t.Context(), outsideScope, f.story); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("private team outsider = %v", err)
	}
	batch, err := f.repository.Batch(t.Context(), outsideScope, []uuid.UUID{f.story, f.secondStory})
	if err != nil || len(batch.Items) != 0 {
		t.Fatalf("private team batch leaked values = %#v,%v", batch, err)
	}
	batch, err = f.repository.Batch(t.Context(), guestScope, []uuid.UUID{f.story, f.secondStory})
	if err != nil || len(batch.Items) != 2 {
		t.Fatalf("batch = %#v,%v", batch, err)
	}
}

func TestArchivedDefinitionsAndOptionsRetainStoredValues(t *testing.T) {
	f := seed(t)
	stage := create(t, f, domain.Definition{Name: "Stage", Type: domain.Select, Options: []domain.OptionInput{{Name: "Won"}, {Name: "Open"}}})
	selected := stage.Options[0].ID.String()
	if _, err := f.repository.Patch(t.Context(), f.scope, f.story, domain.ValuePatch{Values: []domain.Value{{FieldID: stage.ID, Value: &selected}}}); err != nil {
		t.Fatal(err)
	}
	updated, err := f.repository.Update(t.Context(), f.scope, stage.ID, domain.Definition{Name: "Stage", Type: domain.Select, Options: []domain.OptionInput{{ID: stage.Options[1].ID, Name: "In progress"}}})
	if err != nil || len(updated.Options) != 2 {
		t.Fatalf("option archive = %#v,%v", updated, err)
	}
	if _, err := f.repository.Patch(t.Context(), f.scope, f.secondStory, domain.ValuePatch{Values: []domain.Value{{FieldID: stage.ID, Value: &selected}}}); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("new archived selection = %v", err)
	}
	if err := f.repository.Archive(t.Context(), f.scope, stage.ID); err != nil {
		t.Fatal(err)
	}
	snapshot, err := f.repository.Snapshot(t.Context(), f.scope, f.story)
	if err != nil || len(snapshot.Fields) != 1 || snapshot.Fields[0].ArchivedAt == nil || snapshot.Values[0].Value == nil || *snapshot.Values[0].Value != selected {
		t.Fatalf("archived value lost = %#v,%v", snapshot, err)
	}
}

func TestStoryCreationRollsBackFieldsStoryAndOutboxTogether(t *testing.T) {
	f := seed(t)
	amount := create(t, f, domain.Definition{Name: "Value", Type: domain.Money, Currency: text("USD")})
	stories := storyrepository.NewMutationRepository(nil, f.pool, storyrepository.WithCustomFieldCreation(customfieldsadapter.CreationBinder))
	actor, err := auth.NewHumanActor(f.scope.ActorID).WithWorkspace(f.scope.WorkspaceID)
	if err != nil {
		t.Fatal(err)
	}
	scope := storydomain.MutationScope{Actor: actor, WorkspaceID: f.scope.WorkspaceID, ActivityUser: &f.scope.ActorID}
	command := func(value string) storydomain.CreateStoryCommand {
		id, now := uuid.New(), time.Now().UTC()
		return storydomain.CreateStoryCommand{Scope: scope, Story: storydomain.Story{ID: id, Workspace: f.scope.WorkspaceID, Team: f.scope.TeamID, Title: "Atomic deal", Reporter: &f.scope.ActorID, Priority: "High", AutoSchedulingStatus: "off", CreatedAt: now, UpdatedAt: now}, CustomFieldValues: []storydomain.CustomFieldValue{{FieldID: amount.ID, Value: &value}}, Event: storydomain.MutationEvent{ID: uuid.New(), WorkspaceID: f.scope.WorkspaceID, StoryID: id, Type: storydomain.MutationEventStoryCreated, Actor: actor, Payload: []byte(`{"title":"Atomic deal"}`), OccurredAt: now}}
	}
	invalid := command("1e3")
	if _, err := stories.CreateStoryMutation(t.Context(), invalid); !errors.Is(err, storydomain.ErrInvalidMutation) {
		t.Fatalf("invalid atomic create = %v", err)
	}
	for _, entry := range []struct{ table, column string }{{"stories", "id"}, {"story_custom_field_values", "story_id"}, {"story_mutation_events", "story_id"}} {
		var count int64
		if err := f.pool.QueryRow(t.Context(), "SELECT COUNT(*) FROM "+entry.table+" WHERE "+entry.column+"=$1", invalid.Story.ID).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("%s committed across field failure", entry.table)
		}
	}
	valid := command("123.45")
	created, err := stories.CreateStoryMutation(t.Context(), valid)
	if err != nil || !created.Created {
		t.Fatalf("atomic create = %#v,%v", created, err)
	}
	snapshot, err := f.repository.Snapshot(t.Context(), f.scope, created.Story.ID)
	if err != nil || len(snapshot.Values) != 1 || *snapshot.Values[0].Value != "123.45" || auditCount(t, f, created.Story.ID) != 1 {
		t.Fatalf("atomic creation values = %#v,%v", snapshot, err)
	}
}
