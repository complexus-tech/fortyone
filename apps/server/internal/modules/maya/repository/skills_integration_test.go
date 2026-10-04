//go:build integration

package mayarepository

import (
	"context"
	"errors"
	"testing"
	"time"

	mayadomain "github.com/complexus-tech/projects-api/internal/modules/maya/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
)

func TestPersonalSkillsStayOwnerScopedAndPreventLostUpdates(t *testing.T) {
	t.Parallel()
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	workspaceID, userID := insertRealtimeOwner(t, ctx, postgres.Pool, true)
	otherWorkspaceID, otherUserID := insertRealtimeOwner(t, ctx, postgres.Pool, true)
	if _, err := postgres.Pool.Exec(ctx, `INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'member')`, workspaceID, otherUserID); err != nil {
		t.Fatalf("insert second member: %v", err)
	}
	repo := New(postgres.Pool)
	scope := mayadomain.SkillScope{WorkspaceID: workspaceID, UserID: userID}
	content := mayadomain.SkillContent{Name: "Triage", Description: "Review incoming work", Instructions: "Find unassigned tasks and suggest the next step."}
	skill, err := repo.CreateSkill(ctx, scope, content)
	if err != nil {
		t.Fatalf("create skill: %v", err)
	}
	if _, err := repo.CreateSkill(ctx, scope, mayadomain.SkillContent{Name: "TRIAGE", Instructions: "Review"}); !errors.Is(err, mayadomain.ErrSkillNameTaken) {
		t.Fatalf("duplicate name error = %v", err)
	}
	for _, otherScope := range []mayadomain.SkillScope{
		{WorkspaceID: workspaceID, UserID: otherUserID},
		{WorkspaceID: otherWorkspaceID, UserID: userID},
		{WorkspaceID: otherWorkspaceID, UserID: otherUserID},
	} {
		skills, err := repo.ListSkills(ctx, otherScope)
		if err != nil || len(skills) != 0 {
			t.Fatalf("other owner's list = %+v, error = %v", skills, err)
		}
		if _, err := repo.UpdateSkill(ctx, otherScope, skill.ID, content, skill.UpdatedAt); !errors.Is(err, mayadomain.ErrSkillNotFound) {
			t.Fatalf("other owner's edit error = %v", err)
		}
		if err := repo.DeleteSkill(ctx, otherScope, skill.ID); !errors.Is(err, mayadomain.ErrSkillNotFound) {
			t.Fatalf("other owner's delete error = %v", err)
		}
	}
	content.Instructions = "Check incoming tasks and cite each task."
	updated, err := repo.UpdateSkill(ctx, scope, skill.ID, content, skill.UpdatedAt)
	if err != nil || updated.Instructions != content.Instructions || !updated.UpdatedAt.After(skill.UpdatedAt) {
		t.Fatalf("update = %+v, error = %v", updated, err)
	}
	if _, err := repo.UpdateSkill(ctx, scope, skill.ID, content, skill.UpdatedAt); !errors.Is(err, mayadomain.ErrSkillChanged) {
		t.Fatalf("stale edit error = %v", err)
	}
	skills, err := repo.ListSkills(ctx, scope)
	if err != nil || len(skills) != 1 || skills[0].ID != skill.ID {
		t.Fatalf("own list = %+v, error = %v", skills, err)
	}
	if err := repo.DeleteSkill(ctx, scope, skill.ID); err != nil {
		t.Fatalf("delete own skill: %v", err)
	}
	if _, err := repo.CreateSkill(ctx, scope, content); err != nil {
		t.Fatalf("recreate skill: %v", err)
	}
	if _, err := postgres.Pool.Exec(ctx, `DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`, workspaceID, userID); err != nil {
		t.Fatalf("remove membership: %v", err)
	}
	skills, err = repo.ListSkills(ctx, scope)
	if err != nil || len(skills) != 0 {
		t.Fatalf("removed member retained skills = %+v, error = %v", skills, err)
	}
}

func TestSkillSchemaRejectsUnjoinedOwner(t *testing.T) {
	t.Parallel()
	postgres := testkit.NewPostgres(t)
	workspaceID, _ := insertRealtimeOwner(t, t.Context(), postgres.Pool, true)
	_, err := New(postgres.Pool).CreateSkill(t.Context(), mayadomain.SkillScope{WorkspaceID: workspaceID, UserID: uuid.New()}, mayadomain.SkillContent{Name: "Triage", Instructions: "Review"})
	if err == nil {
		t.Fatal("skill was created for an unjoined owner")
	}
}
