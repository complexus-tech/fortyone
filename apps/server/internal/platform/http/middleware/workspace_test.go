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

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type workspaceResolverFake struct {
	resolve func(context.Context, string, uuid.UUID) (WorkspaceInfo, error)
	record  func(context.Context, uuid.UUID, uuid.UUID) error
}

func (fake workspaceResolverFake) ResolveCurrentWorkspace(
	ctx context.Context,
	slug string,
	userID uuid.UUID,
) (WorkspaceInfo, error) {
	return fake.resolve(ctx, slug, userID)
}

func (fake workspaceResolverFake) RecordWorkspaceAccess(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID uuid.UUID,
) error {
	if fake.record == nil {
		return nil
	}
	return fake.record(ctx, workspaceID, userID)
}

func TestWorkspaceMiddlewarePrincipalMatrix(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	principalID := uuid.New()
	credentialID := uuid.New()
	resolver := workspaceResolverFake{
		resolve: func(_ context.Context, slug string, _ uuid.UUID) (WorkspaceInfo, error) {
			return WorkspaceInfo{ID: workspaceID, Name: "Compiler Team", Slug: slug, UserRole: "member"}, nil
		},
	}

	tests := []struct {
		name       string
		kind       platformauth.PrincipalKind
		wantStatus int
	}{
		{name: "human user", kind: platformauth.PrincipalHumanUser, wantStatus: http.StatusNoContent},
		{name: "personal token user", kind: platformauth.PrincipalPersonalToken, wantStatus: http.StatusNoContent},
		{name: "oauth user", kind: platformauth.PrincipalOAuthUser, wantStatus: http.StatusNoContent},
		{name: "service account", kind: platformauth.PrincipalServiceAccount, wantStatus: http.StatusUnauthorized},
		{name: "oauth application", kind: platformauth.PrincipalOAuthApplication, wantStatus: http.StatusUnauthorized},
		{name: "system", kind: platformauth.PrincipalSystem, wantStatus: http.StatusUnauthorized},
		{name: "external contributor", kind: platformauth.PrincipalExternalContributor, wantStatus: http.StatusUnauthorized},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actor, err := platformauth.NewActor(
				principalID,
				test.kind,
				credentialID,
				platformauth.MustScopeSet(platformauth.ScopeWorkspacesRead),
				platformauth.UnrestrictedTeamAccess(),
			)
			if err != nil {
				t.Fatalf("new actor: %v", err)
			}
			requestCtx, err := platformauth.SetActor(context.Background(), actor)
			if err != nil {
				t.Fatalf("set actor: %v", err)
			}

			recorder := runWorkspaceMiddleware(t, requestCtx, resolver, nil)
			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, test.wantStatus)
			}
		})
	}
}

func TestWorkspaceMiddlewareBoundsAccessMetadataWrite(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	deadlineObserved := false
	resolver := workspaceResolverFake{
		resolve: func(context.Context, string, uuid.UUID) (WorkspaceInfo, error) {
			return WorkspaceInfo{ID: workspaceID, Name: "Compiler Team", Slug: "compiler-team", UserRole: "member"}, nil
		},
		record: func(ctx context.Context, _ uuid.UUID, _ uuid.UUID) error {
			deadline, ok := ctx.Deadline()
			deadlineObserved = ok && time.Until(deadline) <= workspaceAccessTimeout
			<-ctx.Done()
			return ctx.Err()
		},
	}

	start := time.Now()
	downstream := false
	recorder := runWorkspaceMiddleware(t, platformauth.SetUserID(context.Background(), userID), resolver, &downstream)
	if recorder.Code != http.StatusNoContent || !downstream {
		t.Fatalf("bounded access response = %d/downstream %t", recorder.Code, downstream)
	}
	if !deadlineObserved {
		t.Fatal("access metadata write did not receive the bounded deadline")
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("access metadata write blocked for %s, want at most one second", elapsed)
	}
}

func TestWorkspaceMiddlewareFailureMatrix(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	workspaceID := uuid.New()
	tests := []struct {
		name           string
		resolveErr     error
		recordErr      error
		wantStatus     int
		wantDownstream bool
	}{
		{name: "membership denied", resolveErr: ErrWorkspaceAccessDenied, wantStatus: http.StatusNotFound},
		{name: "repository unavailable", resolveErr: errors.New("database unavailable"), wantStatus: http.StatusInternalServerError},
		{name: "last access failure is best effort", recordErr: errors.New("write unavailable"), wantStatus: http.StatusNoContent, wantDownstream: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			downstream := false
			resolver := workspaceResolverFake{
				resolve: func(context.Context, string, uuid.UUID) (WorkspaceInfo, error) {
					if test.resolveErr != nil {
						return WorkspaceInfo{}, test.resolveErr
					}
					return WorkspaceInfo{ID: workspaceID, Name: "Compiler Team", Slug: "compiler-team", UserRole: "guest"}, nil
				},
				record: func(context.Context, uuid.UUID, uuid.UUID) error { return test.recordErr },
			}
			recorder := runWorkspaceMiddleware(t, platformauth.SetUserID(context.Background(), userID), resolver, &downstream)
			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, test.wantStatus)
			}
			if downstream != test.wantDownstream {
				t.Fatalf("downstream called = %t, want %t", downstream, test.wantDownstream)
			}
		})
	}
}

func runWorkspaceMiddleware(
	t *testing.T,
	ctx context.Context,
	resolver WorkspaceResolver,
	downstream *bool,
) *httptest.ResponseRecorder {
	t.Helper()
	log := logger.NewWithText(io.Discard, slog.LevelError, "workspace-middleware-test")
	handler := Workspace(log, resolver)(func(ctx context.Context, writer http.ResponseWriter, _ *http.Request) error {
		if downstream != nil {
			*downstream = true
		}
		workspace, err := GetWorkspace(ctx)
		if err != nil {
			return err
		}
		actor, err := platformauth.GetActor(ctx)
		if err != nil {
			return err
		}
		if actor.WorkspaceID != workspace.ID {
			t.Fatalf("actor workspace = %s, want %s", actor.WorkspaceID, workspace.ID)
		}
		return web.Respond(ctx, writer, nil, http.StatusNoContent)
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/workspaces/compiler-team", nil)
	request.SetPathValue("workspaceSlug", "compiler-team")
	if err := handler(ctx, recorder, request); err != nil {
		t.Fatalf("run middleware: %v", err)
	}
	return recorder
}
