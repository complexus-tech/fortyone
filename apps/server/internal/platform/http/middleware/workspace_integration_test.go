//go:build integration

package mid

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	workspacedomain "github.com/complexus-tech/projects-api/internal/modules/workspaces/domain"
	workspacesrepository "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func TestWorkspaceMiddlewareObservesRoleChangesAndRevocationImmediately(t *testing.T) {
	t.Parallel()

	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()

	userID := uuid.New()
	workspaceID := uuid.New()
	suffix := uuid.NewString()
	workspaceSlug := "middleware-revocation-" + suffix
	if _, err := postgres.Pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email, full_name)
		VALUES ($1, $2, $3, $4)
	`, userID, "middleware-"+suffix, "middleware-"+suffix+"@example.com", "Middleware test user"); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if _, err := postgres.Pool.Exec(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug)
		VALUES ($1, $2, $3)
	`, workspaceID, "Middleware revocation test", workspaceSlug); err != nil {
		t.Fatalf("insert workspace: %v", err)
	}
	if _, err := postgres.Pool.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, 'admin')
	`, workspaceID, userID); err != nil {
		t.Fatalf("insert workspace membership: %v", err)
	}

	log := logger.NewWithText(io.Discard, slog.LevelError, "workspace-middleware-integration-test")
	resolver := repositoryWorkspaceResolver{repository: workspacesrepository.New(postgres.Pool)}
	var observed WorkspaceInfo
	var observedActor platformauth.Actor
	nextCalls := 0
	handler := Workspace(log, resolver)(func(requestCtx context.Context, writer http.ResponseWriter, _ *http.Request) error {
		nextCalls++
		var err error
		observed, err = GetWorkspace(requestCtx)
		if err != nil {
			return err
		}
		observedActor, err = platformauth.GetActor(requestCtx)
		if err != nil {
			return err
		}
		return web.Respond(requestCtx, writer, nil, http.StatusNoContent)
	})

	request := func() *httptest.ResponseRecorder {
		t.Helper()
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/workspaces/"+workspaceSlug, nil)
		req.SetPathValue("workspaceSlug", workspaceSlug)
		requestCtx := platformauth.SetUserID(context.Background(), userID)
		if err := handler(requestCtx, recorder, req); err != nil {
			t.Fatalf("run workspace middleware: %v", err)
		}
		return recorder
	}

	if recorder := request(); recorder.Code != http.StatusNoContent {
		t.Fatalf("initial status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if observed.UserRole != "admin" {
		t.Fatalf("initial role = %q, want admin", observed.UserRole)
	}
	if observedActor.WorkspaceID != workspaceID {
		t.Fatalf("actor workspace = %s, want %s", observedActor.WorkspaceID, workspaceID)
	}

	if _, err := postgres.Pool.Exec(ctx, `
		UPDATE workspace_members
		SET role = 'guest'
		WHERE workspace_id = $1 AND user_id = $2
	`, workspaceID, userID); err != nil {
		t.Fatalf("demote workspace member: %v", err)
	}
	if recorder := request(); recorder.Code != http.StatusNoContent {
		t.Fatalf("post-demotion status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if observed.UserRole != "guest" {
		t.Fatalf("post-demotion role = %q, want guest", observed.UserRole)
	}

	if _, err := postgres.Pool.Exec(ctx, `
		DELETE FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2
	`, workspaceID, userID); err != nil {
		t.Fatalf("revoke workspace membership: %v", err)
	}
	if recorder := request(); recorder.Code != http.StatusNotFound {
		t.Fatalf("post-revocation status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
	if nextCalls != 2 {
		t.Fatalf("downstream calls = %d, want 2", nextCalls)
	}
}

type workspaceAccessRepository interface {
	ResolveCurrentMembership(context.Context, string, uuid.UUID) (workspacedomain.CurrentMembership, error)
	RecordAccess(context.Context, uuid.UUID, uuid.UUID) error
}

type repositoryWorkspaceResolver struct {
	repository workspaceAccessRepository
}

func (resolver repositoryWorkspaceResolver) ResolveCurrentWorkspace(
	ctx context.Context,
	slug string,
	userID uuid.UUID,
) (WorkspaceInfo, error) {
	membership, err := resolver.repository.ResolveCurrentMembership(ctx, slug, userID)
	if err != nil {
		if errors.Is(err, workspacedomain.ErrNotFound) {
			return WorkspaceInfo{}, ErrWorkspaceAccessDenied
		}
		return WorkspaceInfo{}, err
	}
	return WorkspaceInfo{
		ID:       membership.WorkspaceID,
		Name:     membership.Name,
		Slug:     membership.Slug,
		UserRole: membership.Role,
	}, nil
}

func (resolver repositoryWorkspaceResolver) RecordWorkspaceAccess(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID uuid.UUID,
) error {
	return resolver.repository.RecordAccess(ctx, workspaceID, userID)
}
