//go:build integration

package storiesrepository

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type storyMutationFixture struct {
	workspaceID        uuid.UUID
	foreignWorkspaceID uuid.UUID
	teamID             uuid.UUID
	otherTeamID        uuid.UUID
	foreignTeamID      uuid.UUID
	statusID           uuid.UUID
	otherStatusID      uuid.UUID
	foreignStatusID    uuid.UUID
	actorID            uuid.UUID
	assigneeID         uuid.UUID
	foreignActorID     uuid.UUID
}

func seedStoryMutationFixture(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
) storyMutationFixture {
	t.Helper()

	fixture := storyMutationFixture{
		workspaceID: uuid.New(), foreignWorkspaceID: uuid.New(),
		teamID: uuid.New(), otherTeamID: uuid.New(), foreignTeamID: uuid.New(),
		statusID: uuid.New(), otherStatusID: uuid.New(), foreignStatusID: uuid.New(),
		actorID: uuid.New(), assigneeID: uuid.New(), foreignActorID: uuid.New(),
	}
	insertMutationUser(t, ctx, pool, fixture.actorID, true)
	insertMutationUser(t, ctx, pool, fixture.assigneeID, true)
	insertMutationUser(t, ctx, pool, fixture.foreignActorID, true)
	insertMutationWorkspace(t, ctx, pool, fixture.workspaceID, fixture.actorID, "primary")
	insertMutationWorkspace(t, ctx, pool, fixture.foreignWorkspaceID, fixture.foreignActorID, "foreign")
	insertMutationTeam(t, ctx, pool, fixture.teamID, fixture.workspaceID, "MUT")
	insertMutationTeam(t, ctx, pool, fixture.otherTeamID, fixture.workspaceID, "OTH")
	insertMutationTeam(t, ctx, pool, fixture.foreignTeamID, fixture.foreignWorkspaceID, "FRN")
	insertMutationStatus(t, ctx, pool, fixture.statusID, fixture.workspaceID, fixture.teamID)
	insertMutationStatus(t, ctx, pool, fixture.otherStatusID, fixture.workspaceID, fixture.otherTeamID)
	insertMutationStatus(t, ctx, pool, fixture.foreignStatusID, fixture.foreignWorkspaceID, fixture.foreignTeamID)

	for _, userID := range []uuid.UUID{fixture.actorID, fixture.assigneeID} {
		mustMutationExec(
			t, ctx, pool,
			"INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'member')",
			fixture.workspaceID, userID,
		)
		mustMutationExec(
			t, ctx, pool,
			"INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)",
			fixture.teamID, userID,
		)
	}
	mustMutationExec(
		t, ctx, pool,
		"INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'member')",
		fixture.foreignWorkspaceID, fixture.foreignActorID,
	)
	mustMutationExec(
		t, ctx, pool,
		"INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)",
		fixture.foreignTeamID, fixture.foreignActorID,
	)
	return fixture
}

func mutationScopeForFixture(t *testing.T, fixture storyMutationFixture) storydomain.MutationScope {
	t.Helper()

	actor, err := platformauth.NewHumanActor(fixture.actorID).WithWorkspace(fixture.workspaceID)
	if err != nil {
		t.Fatalf("scope story mutation actor: %v", err)
	}
	actorID := fixture.actorID
	return storydomain.MutationScope{
		Actor: actor, WorkspaceID: fixture.workspaceID, ActivityUser: &actorID,
	}
}

func seedStoryMutationServiceAccount(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	fixture storyMutationFixture,
	createdAt time.Time,
) (storydomain.MutationScope, uuid.UUID) {
	t.Helper()

	principalID, credentialID := uuid.New(), uuid.New()
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	mustMutationExec(
		t, ctx, pool,
		`INSERT INTO principals (
			principal_id, workspace_id, kind, name, workspace_role, status,
			created_by_user_id, created_at, updated_at
		) VALUES ($1, $2, 'service_account', 'Story integration', 'member', 'active', $3, $4, $4)`,
		principalID, fixture.workspaceID, fixture.actorID, createdAt,
	)
	mustMutationExec(
		t, ctx, pool,
		`INSERT INTO api_credentials (
			credential_id, workspace_id, principal_id, kind, name, lookup_prefix,
			secret_digest, token_version, digest_key_id, digest_key_version,
			expires_at, created_by_user_id, created_at
		) VALUES (
			$1, $2, $3, 'service_account_key', 'Story integration key', repeat('a', 12),
			$4, 1, 'integration', 1, $5, $6, $7
		)`,
		credentialID, fixture.workspaceID, principalID, make([]byte, 32),
		expiresAt, fixture.actorID, createdAt,
	)
	mustMutationExec(
		t, ctx, pool,
		"INSERT INTO api_credential_scopes (credential_id, scope) VALUES ($1, 'stories:write')",
		credentialID,
	)
	mustMutationExec(
		t, ctx, pool,
		"INSERT INTO api_credential_team_restrictions (credential_id, workspace_id, team_id) VALUES ($1, $2, $3)",
		credentialID, fixture.workspaceID, fixture.teamID,
	)
	actor, err := platformauth.NewActor(
		principalID,
		platformauth.PrincipalServiceAccount,
		credentialID,
		platformauth.MustScopeSet(platformauth.ScopeStoriesWrite),
		platformauth.UnrestrictedTeamAccess(),
	)
	if err != nil {
		t.Fatalf("construct story integration service account: %v", err)
	}
	actor, err = actor.WithWorkspace(fixture.workspaceID)
	if err != nil {
		t.Fatalf("bind story integration service account: %v", err)
	}
	return storydomain.MutationScope{Actor: actor, WorkspaceID: fixture.workspaceID}, credentialID
}

func storyCreateMutationCommand(
	t *testing.T,
	fixture storyMutationFixture,
	storyID uuid.UUID,
	occurredAt time.Time,
) storydomain.CreateStoryCommand {
	t.Helper()

	scope := mutationScopeForFixture(t, fixture)
	statusID, reporterID, assigneeID := fixture.statusID, fixture.actorID, fixture.assigneeID
	payload := mustMutationJSON(t, map[string]any{
		"story_id": storyID, "workspace_id": fixture.workspaceID,
	})
	activity := storydomain.MutationActivity{
		ID: uuid.New(), StoryID: storyID, UserID: fixture.actorID,
		Type: "create", Field: "story", CurrentValue: "Created story",
		OldValue: json.RawMessage(`null`), NewValue: mustMutationJSON(t, "Created story"),
		WorkspaceID: fixture.workspaceID, CreatedAt: occurredAt,
	}
	return storydomain.CreateStoryCommand{
		Scope: scope,
		Story: storydomain.Story{
			ID: storyID, Title: "Mutation story " + storyID.String(),
			Status: &statusID, Assignee: &assigneeID, Reporter: &reporterID,
			Priority: "High", AutoSchedulingStatus: "off",
			Team: fixture.teamID, Workspace: fixture.workspaceID,
			CreatedAt: occurredAt, UpdatedAt: occurredAt,
		},
		Event: storydomain.MutationEvent{
			ID: uuid.New(), WorkspaceID: fixture.workspaceID, StoryID: storyID,
			Type: storydomain.MutationEventStoryCreated, Actor: scope.Actor,
			Payload: payload, OccurredAt: occurredAt,
		},
		Activity: &activity,
	}
}

func storyUpdateMutationCommand(
	t *testing.T,
	fixture storyMutationFixture,
	storyID uuid.UUID,
	expectedUpdatedAt, occurredAt time.Time,
	title string,
) storydomain.UpdateStoryCommand {
	t.Helper()

	scope := mutationScopeForFixture(t, fixture)
	payload := mustMutationJSON(t, map[string]any{
		"story_id": storyID, "workspace_id": fixture.workspaceID,
		"changed_fields": []string{"title"},
	})
	activity := storydomain.MutationActivity{
		ID: uuid.New(), StoryID: storyID, UserID: fixture.actorID,
		Type: "update", Field: "title", CurrentValue: title,
		OldValue: mustMutationJSON(t, "before"), NewValue: mustMutationJSON(t, title),
		WorkspaceID: fixture.workspaceID, CreatedAt: occurredAt,
	}
	return storydomain.UpdateStoryCommand{
		Scope: scope, StoryID: storyID, ExpectedUpdatedAt: expectedUpdatedAt,
		Patch: storydomain.StoryPatch{Title: storydomain.SetField(title)},
		Event: storydomain.MutationEvent{
			ID: uuid.New(), WorkspaceID: fixture.workspaceID, StoryID: storyID,
			Type: storydomain.MutationEventStoryUpdated, Actor: scope.Actor,
			Payload: payload, OccurredAt: occurredAt,
		},
		Activities: []storydomain.MutationActivity{activity},
	}
}

func storyDeleteMutationCommand(
	t *testing.T,
	fixture storyMutationFixture,
	storyID uuid.UUID,
	expectedUpdatedAt, occurredAt time.Time,
) storydomain.DeleteStoryCommand {
	t.Helper()

	scope := mutationScopeForFixture(t, fixture)
	payload := mustMutationJSON(t, map[string]any{
		"story_id": storyID, "workspace_id": fixture.workspaceID,
	})
	activity := storydomain.MutationActivity{
		ID: uuid.New(), StoryID: storyID, UserID: fixture.actorID,
		Type: "delete", Field: "story", CurrentValue: "Deleted story",
		OldValue: mustMutationJSON(t, "Deleted story"), NewValue: json.RawMessage(`null`),
		WorkspaceID: fixture.workspaceID, CreatedAt: occurredAt,
	}
	return storydomain.DeleteStoryCommand{
		Scope: scope, StoryID: storyID, ExpectedUpdatedAt: expectedUpdatedAt,
		Event: storydomain.MutationEvent{
			ID: uuid.New(), WorkspaceID: fixture.workspaceID, StoryID: storyID,
			Type: storydomain.MutationEventStoryDeleted, Actor: scope.Actor,
			Payload: payload, OccurredAt: occurredAt,
		},
		Activity: &activity,
	}
}

func insertMutationUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, active bool) {
	t.Helper()
	suffix := uuid.NewString()
	mustMutationExec(
		t, ctx, pool,
		"INSERT INTO users (user_id, username, email, is_active) VALUES ($1, $2, $3, $4)",
		id, "mutation-"+suffix, "mutation-"+suffix+"@example.com", active,
	)
}

func insertMutationWorkspace(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	id, creatorID uuid.UUID,
	name string,
) {
	t.Helper()
	mustMutationExec(
		t, ctx, pool,
		"INSERT INTO workspaces (workspace_id, name, slug, created_by) VALUES ($1, $2, $3, $4)",
		id, "Mutation "+name, "mutation-"+name+"-"+uuid.NewString(), creatorID,
	)
}

func insertMutationTeam(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	id, workspaceID uuid.UUID,
	code string,
) {
	t.Helper()
	mustMutationExec(
		t, ctx, pool,
		"INSERT INTO teams (team_id, name, workspace_id, code, color) VALUES ($1, $2, $3, $4, '#000000')",
		id, "Mutation "+code, workspaceID, code,
	)
}

func insertMutationStatus(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	id, workspaceID, teamID uuid.UUID,
) {
	t.Helper()
	mustMutationExec(
		t, ctx, pool,
		"INSERT INTO statuses (status_id, name, category, workspace_id, team_id) VALUES ($1, 'Started', 'started', $2, $3)",
		id, workspaceID, teamID,
	)
}

func mustMutationJSON(t *testing.T, value any) json.RawMessage {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("encode story mutation fixture: %v", err)
	}
	return payload
}

func mustMutationExec(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	statement string,
	args ...any,
) {
	t.Helper()
	if _, err := pool.Exec(ctx, statement, args...); err != nil {
		t.Fatalf("seed story mutation fixture: %v\nstatement: %s", err, fmt.Sprintf("%.120s", statement))
	}
}
