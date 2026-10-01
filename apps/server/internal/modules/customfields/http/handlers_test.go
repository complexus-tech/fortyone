package customfieldshttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/customfields/domain"
	auth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type handlerService struct {
	Service
	called bool
}

func (s *handlerService) Patch(context.Context, domain.Scope, uuid.UUID, domain.ValuePatch) (domain.Snapshot, error) {
	s.called = true
	return domain.Snapshot{Fields: []domain.Field{}, Values: []domain.Value{}}, nil
}

type workspaceResolver struct{ id uuid.UUID }

func (r workspaceResolver) ResolveCurrentWorkspace(context.Context, string, uuid.UUID) (mid.WorkspaceInfo, error) {
	return mid.WorkspaceInfo{ID: r.id, Slug: "test", UserRole: "admin"}, nil
}
func (r workspaceResolver) RecordWorkspaceAccess(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func callHandler(t *testing.T, handler web.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	actorID, workspaceID := uuid.New(), uuid.New()
	ctx, err := auth.SetActor(t.Context(), auth.NewHumanActor(actorID))
	if err != nil {
		t.Fatal(err)
	}
	ctx = auth.SetUserID(ctx, actorID)
	request := httptest.NewRequest(http.MethodPut, "/workspaces/test/stories/"+uuid.NewString()+"/custom-fields", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.SetPathValue("workspaceSlug", "test")
	request.SetPathValue("id", uuid.NewString())
	response := httptest.NewRecorder()
	if err := mid.Workspace(nil, workspaceResolver{id: workspaceID})(handler)(ctx, response, request); err != nil {
		t.Fatal(err)
	}
	return response
}

func TestPatchRejectsNumbersUnknownFieldsAndOversizedRequests(t *testing.T) {
	for _, body := range []string{
		`{"values":[{"fieldId":"` + uuid.NewString() + `","value":123.45}]}`,
		`{"values":[],"unexpected":"SECRET-CUSTOMER-VALUE"}`,
		`{"values":[{"fieldId":"` + uuid.NewString() + `","value":"` + strings.Repeat("x", 4001) + `"}]}`,
	} {
		service := &handlerService{}
		response := callHandler(t, New(service).Patch, body)
		if response.Code != http.StatusBadRequest || service.called {
			t.Fatalf("invalid body response = %d called=%v body=%s", response.Code, service.called, response.Body.String())
		}
		if strings.Contains(response.Body.String(), "SECRET-CUSTOMER-VALUE") {
			t.Fatal("validation reflected a customer value")
		}
	}
}

func TestPatchAcceptsExactStringsAndExplicitNull(t *testing.T) {
	service := &handlerService{}
	response := callHandler(t, New(service).Patch, `{"expectedVersion":0,"values":[{"fieldId":"`+uuid.NewString()+`","value":"9007199254740993.01"},{"fieldId":"`+uuid.NewString()+`","value":null}]}`)
	if response.Code != http.StatusOK || !service.called {
		t.Fatalf("exact string response = %d called=%v body=%s", response.Code, service.called, response.Body.String())
	}
}
