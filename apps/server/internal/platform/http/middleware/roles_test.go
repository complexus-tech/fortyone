package mid

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/logger"
)

func TestRequireMinimumRoleAdminMatrix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		role       Role
		wantStatus int
		wantNext   bool
	}{
		{name: "guest denied", role: RoleGuest, wantStatus: http.StatusForbidden},
		{name: "member denied", role: RoleMember, wantStatus: http.StatusForbidden},
		{name: "admin allowed", role: RoleAdmin, wantStatus: http.StatusNoContent, wantNext: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			called := false
			handler := RequireMinimumRole(
				logger.NewWithText(io.Discard, slog.LevelError, "workspace-role-test"),
				RoleAdmin,
			)(func(_ context.Context, writer http.ResponseWriter, _ *http.Request) error {
				called = true
				writer.WriteHeader(http.StatusNoContent)
				return nil
			})
			ctx := context.WithValue(context.Background(), workspaceKey, WorkspaceInfo{UserRole: string(test.role)})
			recorder := httptest.NewRecorder()

			if err := handler(ctx, recorder, httptest.NewRequest(http.MethodGet, "/", nil)); err != nil {
				t.Fatalf("admin middleware returned error: %v", err)
			}
			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, test.wantStatus)
			}
			if called != test.wantNext {
				t.Fatalf("next called = %t, want %t", called, test.wantNext)
			}
		})
	}
}
