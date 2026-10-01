//go:build integration

package teamsrepository

import (
	"context"
	"errors"
	"testing"
	"time"

	teamsdomain "github.com/complexus-tech/projects-api/internal/modules/teams/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
)

func TestTeamWorkNamingPreservesPrivacyAndWorkspaceScope(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	workspace := insertTeamTestWorkspace(t, ctx, postgres.Pool, "naming")
	otherWorkspace := insertTeamTestWorkspace(t, ctx, postgres.Pool, "other-naming")
	actor := insertTeamTestUser(t, ctx, postgres.Pool, "naming-admin", true)
	insertTeamTestWorkspaceMember(t, ctx, postgres.Pool, workspace, actor, "admin")
	teamID := insertTeamTestTeam(t, ctx, postgres.Pool, workspace, "Sales", true)
	repository := newTeamIntegrationRepository(postgres.Pool)
	deal := "deal"
	updated, err := repository.Update(ctx, teamID, teamsdomain.Team{
		Workspace: workspace, StoryTerm: &deal, StoryTermSet: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.StoryTerm == nil || *updated.StoryTerm != deal || !updated.IsPrivate {
		t.Fatalf("naming update lost privacy or term: %#v", updated)
	}
	updated, err = repository.Update(ctx, teamID, teamsdomain.Team{Workspace: workspace, Name: "Revenue"})
	if err != nil {
		t.Fatal(err)
	}
	if updated.StoryTerm == nil || *updated.StoryTerm != deal || !updated.IsPrivate {
		t.Fatalf("omitted naming/privacy changed: %#v", updated)
	}
	if _, err := repository.Update(ctx, teamID, teamsdomain.Team{
		Workspace: otherWorkspace, StoryTermSet: true,
	}); !errors.Is(err, teamsdomain.ErrNotFound) {
		t.Fatalf("cross-workspace update = %v", err)
	}
	read, err := repository.GetByID(ctx, teamID, workspace, actor)
	if err != nil || read.StoryTerm == nil || *read.StoryTerm != deal {
		t.Fatalf("scoped read = %#v, %v", read, err)
	}
	reset, err := repository.Update(ctx, teamID, teamsdomain.Team{Workspace: workspace, StoryTermSet: true})
	if err != nil || reset.StoryTerm != nil || !reset.IsPrivate {
		t.Fatalf("reset = %#v, %v", reset, err)
	}
	invalid := "deals"
	if _, err := repository.Update(ctx, teamID, teamsdomain.Team{Workspace: workspace, StoryTerm: &invalid, StoryTermSet: true}); err == nil {
		t.Fatal("database accepted an unsupported plural work term")
	}
}
