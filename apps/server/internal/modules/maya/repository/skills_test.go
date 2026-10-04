package mayarepository

import (
	"context"
	"errors"
	"testing"
	"time"

	mayadomain "github.com/complexus-tech/projects-api/internal/modules/maya/domain"
	mayasql "github.com/complexus-tech/projects-api/internal/modules/maya/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestSkillWritesStayOwnerScopedAndReportStaleVersion(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name   string
		exists bool
		want   error
	}{
		{name: "stale own skill", exists: true, want: mayadomain.ErrSkillChanged},
		{name: "missing or another owner", want: mayadomain.ErrSkillNotFound},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			queries := &skillQueryStub{exists: test.exists}
			scope := mayadomain.SkillScope{WorkspaceID: uuid.New(), UserID: uuid.New()}
			id, version := uuid.New(), time.Now().UTC()
			_, err := newWithQueries(queries).UpdateSkill(t.Context(), scope, id, mayadomain.SkillContent{}, version)
			if !errors.Is(err, test.want) {
				t.Fatalf("update error = %v, want %v", err, test.want)
			}
			if queries.update.ID != id || queries.update.WorkspaceID != scope.WorkspaceID || queries.update.UserID != scope.UserID || !queries.update.ExpectedUpdatedAt.Equal(version) {
				t.Fatalf("update params = %+v", queries.update)
			}
			if queries.lookup.ID != id || queries.lookup.WorkspaceID != scope.WorkspaceID || queries.lookup.UserID != scope.UserID {
				t.Fatalf("version lookup crossed scope: %+v", queries.lookup)
			}
		})
	}
}

func TestSkillDeleteCannotDeleteAnotherOwner(t *testing.T) {
	t.Parallel()
	queries := &skillQueryStub{}
	scope := mayadomain.SkillScope{WorkspaceID: uuid.New(), UserID: uuid.New()}
	id := uuid.New()
	err := newWithQueries(queries).DeleteSkill(t.Context(), scope, id)
	if !errors.Is(err, mayadomain.ErrSkillNotFound) || queries.deleted.ID != id || queries.deleted.UserID != scope.UserID || queries.deleted.WorkspaceID != scope.WorkspaceID {
		t.Fatalf("delete error = %v, params = %+v", err, queries.deleted)
	}
}

func TestSkillDuplicateNameHasUsefulError(t *testing.T) {
	t.Parallel()
	err := skillWriteError("create", &pgconn.PgError{Code: "23505", ConstraintName: "maya_skills_owner_name_idx"})
	if !errors.Is(err, mayadomain.ErrSkillNameTaken) {
		t.Fatalf("duplicate error = %v", err)
	}
	other := &pgconn.PgError{Code: "23505", ConstraintName: "unrelated"}
	if !errors.Is(skillWriteError("create", other), other) {
		t.Fatal("unrelated persistence error was discarded")
	}
}

type skillQueryStub struct {
	mayasql.Querier
	exists  bool
	update  mayasql.UpdateMayaSkillParams
	lookup  mayasql.MayaSkillExistsParams
	deleted mayasql.DeleteMayaSkillParams
}

func (queries *skillQueryStub) UpdateMayaSkill(_ context.Context, params mayasql.UpdateMayaSkillParams) (mayasql.UpdateMayaSkillRow, error) {
	queries.update = params
	return mayasql.UpdateMayaSkillRow{}, pgx.ErrNoRows
}

func (queries *skillQueryStub) MayaSkillExists(_ context.Context, params mayasql.MayaSkillExistsParams) (bool, error) {
	queries.lookup = params
	return queries.exists, nil
}

func (queries *skillQueryStub) DeleteMayaSkill(_ context.Context, params mayasql.DeleteMayaSkillParams) (int64, error) {
	queries.deleted = params
	return 0, nil
}
