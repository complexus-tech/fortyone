//go:build integration

package storiesrepository

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	commentsrepository "github.com/complexus-tech/projects-api/internal/modules/comments/repository"
	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	notificationsrepository "github.com/complexus-tech/projects-api/internal/modules/notifications/repository"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	reportsdomain "github.com/complexus-tech/projects-api/internal/modules/reports/domain"
	reportsrepository "github.com/complexus-tech/projects-api/internal/modules/reports/repository"
	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Exercise the actual services and SQLC repositories without HTTP middleware.
// No Redis, email, calendar provider, or production database is involved.
func TestBackgroundStoryOperationsUseExplicitActors(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx := t.Context()
	fixture := seedStoryReadFixture(t, ctx, postgres.Pool)
	log := logger.NewWithText(io.Discard, slog.LevelError, "background-operations-test")
	repository := New(log, postgres.Pool)
	service := stories.New(log, repository, nil, nil)
	service.ConfigureCommentCreator(backgroundCommentCreator{comments.New(commentsrepository.New(log, postgres.Pool))})
	systemID, recipientID := uuid.New(), uuid.New()
	insertStoryReadUser(t, ctx, postgres.Pool, systemID, true)
	insertStoryReadUser(t, ctx, postgres.Pool, recipientID, true)
	mustStoryReadExec(t, ctx, postgres.Pool, "UPDATE users SET is_system = TRUE WHERE user_id = $1", systemID)
	mustStoryReadExec(t, ctx, postgres.Pool, "INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'member')", fixture.workspaceA, recipientID)
	mustStoryReadExec(t, ctx, postgres.Pool, "INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)", fixture.teamA, recipientID)
	systemActor, err := auth.NewActor(systemID, auth.PrincipalSystem, uuid.Nil, auth.MustScopeSet(auth.ScopeFirstParty), auth.UnrestrictedTeamAccess())
	require.NoError(t, err)
	systemActor, err = systemActor.WithWorkspace(fixture.workspaceA)
	require.NoError(t, err)
	systemCtx, err := auth.SetActor(ctx, systemActor)
	require.NoError(t, err)

	t.Run("interactive reads still require an actor", func(t *testing.T) {
		_, err := service.Get(ctx, fixture.visible, fixture.workspaceA)
		require.ErrorIs(t, err, stories.ErrStoryReadForbidden)
		story, err := service.GetForSystem(ctx, systemID, fixture.visible, fixture.workspaceA)
		require.NoError(t, err)
		require.Equal(t, fixture.visible, story.ID)
		_, err = service.GetForSystem(ctx, fixture.actor, fixture.visible, fixture.workspaceA)
		require.Error(t, err, "a human UUID must not grant system access")
		_, err = service.GetForSystem(ctx, systemID, fixture.crossTenant, fixture.workspaceA)
		require.Error(t, err)
	})

	t.Run("assignment reaches the recipient inbox", func(t *testing.T) {
		rules := notifications.NewRules(log, service, nil, nil)
		batch, err := rules.ProcessStoryUpdate(ctx, events.StoryUpdatedPayload{
			StoryID: fixture.visible, WorkspaceID: fixture.workspaceA,
			Updates: map[string]any{"assignee_id": recipientID.String()},
		}, fixture.actor)
		require.NoError(t, err)
		require.Len(t, batch, 1)
		require.NotEmpty(t, batch[0].Title)
		batch[0].DedupeKey = uuid.NewString()
		inbox := notificationsrepository.New(postgres.Pool)
		created, inserted, err := inbox.Create(ctx, batch[0])
		require.NoError(t, err)
		require.True(t, inserted)
		require.True(t, created.InAppEnabled)
		items, err := inbox.List(ctx, notificationsdomain.ListQuery{
			Access: notificationsdomain.WorkspaceAccess{ActorID: recipientID, WorkspaceID: fixture.workspaceA}, Limit: 20,
		})
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.Equal(t, created.ID, items[0].ID)
		for _, actorID := range []uuid.UUID{fixture.foreignUser, fixture.inactive} {
			_, err := service.GetEventStoryTitle(ctx, actorID, fixture.visible, fixture.workspaceA)
			require.ErrorIs(t, err, storydomain.ErrNotFound)
		}
		_, err = service.GetEventStoryTitle(ctx, systemID, fixture.crossTenant, fixture.workspaceA)
		require.ErrorIs(t, err, storydomain.ErrNotFound)
		_, err = service.GetEventStoryTitle(ctx, systemID, fixture.hidden, fixture.workspaceA)
		require.NoError(t, err, "active system identity has an explicit event read")
	})

	t.Run("background status updates also set completion", func(t *testing.T) {
		completedID := uuid.New()
		mustStoryReadExec(t, ctx, postgres.Pool, "INSERT INTO statuses (status_id, name, category, workspace_id, team_id) VALUES ($1, 'Done', 'completed', $2, $3)", completedID, fixture.workspaceA, fixture.teamA)
		err := service.UpdateExternalWithReason(ctx, systemID, fixture.visible, fixture.workspaceA, map[string]any{"status_id": completedID}, "GitHub completed the task")
		require.NoError(t, err)
		story, err := service.GetForSystem(ctx, systemID, fixture.visible, fixture.workspaceA)
		require.NoError(t, err)
		require.NotNil(t, story.CompletedAt)
		err = service.RecordSystemActivity(ctx, storydomain.Activity{
			ID: uuid.New(), StoryID: fixture.visible, WorkspaceID: fixture.workspaceA, UserID: systemID,
			Type: "update", Field: "auto_scheduling_status", CurrentValue: "scheduled", CreatedAt: time.Now().UTC(),
		})
		require.NoError(t, err)
	})

	t.Run("external human comments and provider replies", func(t *testing.T) {
		parent, err := service.CreateCommentExternal(ctx, fixture.actor, fixture.workspaceA, stories.CoreNewComment{
			StoryID: fixture.visible, UserID: fixture.actor, Comment: "Human comment from a background action",
		})
		require.NoError(t, err)
		reply, err := service.CreateCommentExternal(systemCtx, systemID, fixture.workspaceA, stories.CoreNewComment{
			StoryID: fixture.visible, UserID: systemID, Parent: &parent.ID, Comment: "Provider reply",
		})
		require.NoError(t, err)
		require.Equal(t, systemID, reply.UserID)
		require.Equal(t, parent.ID, *reply.Parent)
		_, err = service.CreateCommentExternal(systemCtx, fixture.actor, fixture.workspaceA, stories.CoreNewComment{
			StoryID: fixture.visible, Comment: "Identity mismatch",
		})
		require.ErrorIs(t, err, stories.ErrStoryMutationForbidden)
		_, err = service.CreateCommentExternal(ctx, fixture.foreignUser, fixture.workspaceA, stories.CoreNewComment{
			StoryID: fixture.visible, Comment: "Outside the workspace",
		})
		require.Error(t, err)
		_, err = service.CreateCommentExternal(systemCtx, systemID, fixture.workspaceA, stories.CoreNewComment{
			StoryID: fixture.hidden, Parent: &parent.ID, Comment: "Parent from another story",
		})
		require.ErrorIs(t, err, storydomain.ErrNotFound)
	})

	t.Run("system workload is limited to the requested team", func(t *testing.T) {
		reports := reportsrepository.New(log, postgres.Pool)
		analysis, err := reports.GetSystemTeamWorkload(ctx, systemID, fixture.workspaceA, fixture.teamA)
		require.NoError(t, err)
		for _, team := range analysis.Teams {
			require.Equal(t, fixture.teamA, team.TeamID)
		}
		for _, actorID := range []uuid.UUID{fixture.actor, fixture.inactive, uuid.New()} {
			_, err := reports.GetSystemTeamWorkload(ctx, actorID, fixture.workspaceA, fixture.teamA)
			require.ErrorIs(t, err, reportsdomain.ErrReportsAccessDenied)
		}
		_, err = reports.GetSystemTeamWorkload(ctx, systemID, fixture.workspaceA, fixture.teamB)
		require.ErrorIs(t, err, reportsdomain.ErrReportsAccessDenied)
	})

	t.Run("revoked system users cannot read or create", func(t *testing.T) {
		mustStoryReadExec(t, ctx, postgres.Pool, "UPDATE users SET is_active = FALSE WHERE user_id = $1", systemID)
		_, err := service.GetEventStoryTitle(ctx, systemID, fixture.visible, fixture.workspaceA)
		require.ErrorIs(t, err, storydomain.ErrNotFound)
		_, err = service.GetForSystem(ctx, systemID, fixture.visible, fixture.workspaceA)
		require.Error(t, err)
		_, err = service.CreateCommentExternal(systemCtx, systemID, fixture.workspaceA, stories.CoreNewComment{StoryID: fixture.visible, Comment: "Revoked provider"})
		require.Error(t, err)
	})
}

type backgroundCommentCreator struct{ service *comments.Service }

func (adapter backgroundCommentCreator) CreateComment(ctx context.Context, command stories.CreateCommentCommand) (stories.CoreComment, error) {
	created, err := adapter.service.CreateComment(ctx, comments.CreateCommentCommand{
		WorkspaceID: command.WorkspaceID, StoryID: command.StoryID, ParentID: command.ParentID,
		Actor: command.Actor, Content: command.Content, MentionedUserIDs: command.MentionedUserIDs,
	})
	return stories.CoreComment{ID: created.ID, StoryID: created.StoryID, Parent: created.Parent, UserID: created.UserID, Comment: created.Comment, CreatedAt: created.CreatedAt, UpdatedAt: created.UpdatedAt}, err
}
