package workpresetshttp

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/workpresets/domain"
	workpresets "github.com/complexus-tech/projects-api/internal/modules/workpresets/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type repositoryStub struct {
	workpresets.Repository
	items            []domain.Preset
	calls            int
	actor, workspace uuid.UUID
	filters          domain.List
}

func (s *repositoryStub) List(_ context.Context, actor, workspace uuid.UUID, filters domain.List) ([]domain.Preset, error) {
	s.calls++
	s.actor, s.workspace, s.filters = actor, workspace, filters
	return s.items, nil
}

type workspaceResolver struct{ id uuid.UUID }

func (s workspaceResolver) ResolveCurrentWorkspace(context.Context, string, uuid.UUID) (mid.WorkspaceInfo, error) {
	return mid.WorkspaceInfo{ID: s.id, Slug: "acme", UserRole: "member"}, nil
}

func (s workspaceResolver) RecordWorkspaceAccess(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func request(t *testing.T, handler web.Handler, actor, workspace uuid.UUID, url, body string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, url, strings.NewReader(body))
	r.SetPathValue("workspaceSlug", "acme")
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	log := logger.NewWithText(io.Discard, slog.LevelError, "test")
	bound := mid.Workspace(log, workspaceResolver{workspace})(handler)
	if err := bound(platformauth.SetUserID(t.Context(), actor), w, r); err != nil {
		t.Fatal(err)
	}
	return w
}

func TestListCursorBindsActorWorkspaceTeamKindAndExpiry(t *testing.T) {
	actor, workspace, team := uuid.New(), uuid.New(), uuid.New()
	repo := &repositoryStub{items: []domain.Preset{{ID: uuid.New(), CreatedAt: time.Now()}, {ID: uuid.New(), CreatedAt: time.Now()}}}
	h, err := New(workpresets.New(repo), "test-cursor-secret-with-at-least-32-bytes")
	if err != nil {
		t.Fatal(err)
	}
	url := "/?teamId=" + team.String() + "&kind=view&limit=1"
	w := request(t, h.List, actor, workspace, url, "")
	if w.Code != http.StatusOK || repo.actor != actor || repo.workspace != workspace || repo.filters.Limit != 2 {
		t.Fatalf("incorrect list contract: %d %+v", w.Code, repo)
	}
	var page struct {
		Data struct {
			Items      []domain.Preset `json:"items"`
			NextCursor string          `json:"nextCursor"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Data.Items) != 1 || page.Data.NextCursor == "" {
		t.Fatalf("missing bounded next page: %s", w.Body.String())
	}
	w = request(t, h.List, actor, workspace, url+"&cursor="+page.Data.NextCursor, "")
	if w.Code != http.StatusOK || repo.filters.Before == nil || repo.filters.Before.ID != repo.items[0].ID {
		t.Fatal("valid cursor did not forward position")
	}
	for _, test := range []struct {
		name             string
		actor, workspace uuid.UUID
		url              string
	}{
		{"different actor", uuid.New(), workspace, url},
		{"different workspace", actor, uuid.New(), url},
		{"different team", actor, workspace, "/?teamId=" + uuid.NewString() + "&kind=view&limit=1"},
		{"different kind", actor, workspace, "/?teamId=" + team.String() + "&kind=template&limit=1"},
	} {
		t.Run(test.name, func(t *testing.T) {
			before := repo.calls
			w := request(t, h.List, test.actor, test.workspace, test.url+"&cursor="+page.Data.NextCursor, "")
			if w.Code != http.StatusBadRequest || repo.calls != before {
				t.Fatal("cross-scope cursor reached repository")
			}
		})
	}
	expired, err := h.cursors.Encode(cursor{ActorID: actor, WorkspaceID: workspace, TeamID: team, Kind: domain.View, ExpiresAt: time.Now().Add(-time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{"&cursor=" + expired, "&cursor=" + page.Data.NextCursor + "x", "&limit=2", "&secret=value"} {
		before := repo.calls
		if w := request(t, h.List, actor, workspace, url+suffix, ""); w.Code != http.StatusBadRequest || repo.calls != before {
			t.Fatalf("invalid query reached repository: %s", suffix)
		}
	}
}

func TestCreateRejectsUnknownAndOversizedSnapshotBeforePersistence(t *testing.T) {
	repo := &repositoryStub{}
	h, err := New(workpresets.New(repo), "test-cursor-secret-with-at-least-32-bytes")
	if err != nil {
		t.Fatal(err)
	}
	actor, workspace, team := uuid.New(), uuid.New(), uuid.New()
	for _, configuration := range []string{
		`{"version":1,"layout":"list","filters":{"secret":"value"},"viewOptions":{}}`,
		`{"version":1,"title":"Task","priority":"none","secret":"value"}`,
		`{"version":1,"title":"` + strings.Repeat("x", 70000) + `","priority":"none"}`,
	} {
		body := `{"teamId":"` + team.String() + `","kind":"template","visibility":"team","name":"Test","configuration":` + configuration + `}`
		if w := request(t, h.Create, actor, workspace, "/", body); w.Code != http.StatusBadRequest {
			t.Fatalf("invalid snapshot accepted: %d", w.Code)
		}
	}
}
