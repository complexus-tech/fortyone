package desktophttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type syncStories struct {
	StoryService
	list      []stories.CoreStoryList
	filter    stories.CoreStoryFilters
	listCalls int
	removed   uuid.UUID
}

func (s *syncStories) List(_ context.Context, _ uuid.UUID, filter stories.CoreStoryFilters) ([]stories.CoreStoryList, error) {
	s.filter = filter
	s.listCalls++
	return s.list, nil
}
func (s *syncStories) Get(_ context.Context, id, workspace uuid.UUID) (stories.CoreSingleStory, error) {
	if id == s.removed {
		return stories.CoreSingleStory{}, stories.ErrStoryReadForbidden
	}
	return stories.CoreSingleStory{ID: id, Workspace: workspace}, nil
}
func (s *syncStories) GetComments(context.Context, uuid.UUID, uuid.UUID, int, int) ([]stories.CoreComment, bool, error) {
	return nil, false, nil
}

type syncDocuments struct {
	DocumentService
	listCalls int
	document  documents.CoreDocument
	removed   uuid.UUID
}

func (s *syncDocuments) List(context.Context, documents.CoreListInput) ([]documents.CoreDocumentSummary, error) {
	s.listCalls++
	return nil, nil
}
func (s *syncDocuments) GetForOffline(_ context.Context, _, _, id uuid.UUID) (documents.CoreDocument, error) {
	if id == s.removed {
		return documents.CoreDocument{}, documents.ErrForbidden
	}
	return s.document, nil
}

type syncTeams struct{ TeamReader }

func (*syncTeams) List(context.Context, uuid.UUID, uuid.UUID, ...teams.CoreListTeamsFilter) ([]teams.CoreTeam, error) {
	return nil, nil
}

type syncStatuses struct{ StatusReader }

func (*syncStatuses) List(context.Context, uuid.UUID, uuid.UUID) ([]states.CoreState, error) {
	return nil, nil
}

type workspaceResolver struct{ value mid.WorkspaceInfo }

func (r workspaceResolver) ResolveCurrentWorkspace(context.Context, string, uuid.UUID) (mid.WorkspaceInfo, error) {
	return r.value, nil
}
func (workspaceResolver) RecordWorkspaceAccess(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func runWorkspaceHandler(t *testing.T, handler web.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.SetPathValue("workspaceSlug", "acme")
	recorder := httptest.NewRecorder()
	wrapped := mid.Workspace(nil, workspaceResolver{mid.WorkspaceInfo{ID: uuid.New(), Slug: "acme", UserRole: "member"}})(handler)
	require.NoError(t, wrapped(platformauth.SetUserID(t.Context(), uuid.New()), recorder, request))
	return recorder
}

func TestSelectedSyncRevalidatesKnownIDsWithoutListingAndPreservesCRDTEpoch(t *testing.T) {
	storyID, revokedStory := uuid.New(), uuid.New()
	documentID, revokedDocument := uuid.New(), uuid.New()
	storyService := &syncStories{removed: revokedStory}
	documentService := &syncDocuments{removed: revokedDocument, document: documents.CoreDocument{ID: documentID, CollaborationEpoch: 7, Collaborative: true, CollaborationState: []byte{1, 2, 3}}}
	h := &Handlers{Config: Config{Stories: storyService, Documents: documentService, Teams: &syncTeams{}, Statuses: &syncStatuses{}}}
	response := runWorkspaceHandler(t, h.Sync, "/?scope=selected&storyIds="+storyID.String()+","+revokedStory.String()+"&documentIds="+documentID.String()+","+revokedDocument.String())
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var result struct{ Data Snapshot }
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
	require.Zero(t, storyService.listCalls)
	require.Zero(t, documentService.listCalls)
	require.False(t, result.Data.Complete["stories"])
	require.False(t, result.Data.Complete["documents"])
	require.Equal(t, []uuid.UUID{revokedStory}, result.Data.Removed["stories"])
	require.Equal(t, []uuid.UUID{revokedDocument}, result.Data.Removed["documents"])
	require.Equal(t, []uuid.UUID{storyID}, result.Data.CommentsCompleteFor)
	require.Len(t, result.Data.Documents, 1)
	require.Equal(t, int64(7), result.Data.Documents[0].CollaborationEpoch)
	require.Equal(t, []byte{1, 2, 3}, result.Data.Documents[0].CollaborationState)
}

func TestMyWorkPaginationNeverClaimsWholeWorkspaceReplacement(t *testing.T) {
	storyService := &syncStories{list: make([]stories.CoreStoryList, storySnapshotLimit+1)}
	for i := range storyService.list {
		storyService.list[i].ID = uuid.New()
	}
	h := &Handlers{Config: Config{Stories: storyService, Documents: &syncDocuments{}, Teams: &syncTeams{}, Statuses: &syncStatuses{}}}
	response := runWorkspaceHandler(t, h.Sync, "/?myWork=true&storyOffset=100")
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var result struct{ Data Snapshot }
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))
	require.NotNil(t, storyService.filter.AssignedToMe)
	require.True(t, *storyService.filter.AssignedToMe)
	require.Equal(t, 100, storyService.filter.Offset)
	require.Equal(t, 200, *result.Data.NextStoryOffset)
	require.Equal(t, "myWork", result.Data.Scope)
	require.False(t, result.Data.Complete["stories"])
	require.Len(t, result.Data.Stories, storySnapshotLimit)
}
