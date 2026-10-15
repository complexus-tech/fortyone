package desktophttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const storySnapshotLimit = 100
const documentSnapshotLimit = 50
const maximumSnapshotBytes = 20 << 20

type Snapshot struct {
	SchemaVersion       int                    `json:"schemaVersion"`
	SyncedAt            time.Time              `json:"syncedAt"`
	WorkspaceID         uuid.UUID              `json:"workspaceId"`
	UserID              uuid.UUID              `json:"userId"`
	Role                string                 `json:"role"`
	Stories             []Story                `json:"stories"`
	Documents           []Document             `json:"documents"`
	Comments            []Comment              `json:"comments"`
	Teams               []Team                 `json:"teams"`
	Statuses            []Status               `json:"statuses"`
	Complete            map[string]bool        `json:"complete"`
	Removed             map[string][]uuid.UUID `json:"removed"`
	NextStoryOffset     *int                   `json:"nextStoryOffset"`
	NextDocumentOffset  *int                   `json:"nextDocumentOffset"`
	Scope               string                 `json:"scope"`
	CommentsCompleteFor []uuid.UUID            `json:"commentsCompleteFor"`
}

// Sync returns bounded actor-visible data. Only an explicitly complete
// collection can replace a complete local collection. Known IDs are rechecked
// individually so revoked/deleted items can be evicted even in large workspaces.
// This is a replacement protocol, not a cursor pretending to be a change feed.
func (h *Handlers) Sync(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Cache-Control", "no-store")
	workspace, userID, err := requestIdentity(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	knownStories, err := parseKnownIDs(r, "storyIds")
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	knownDocuments, err := parseKnownIDs(r, "documentIds")
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	storyLimit, err := parseSnapshotLimit(r, "storyLimit", storySnapshotLimit)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	documentLimit, err := parseSnapshotLimit(r, "documentLimit", documentSnapshotLimit)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	storyOffset, err := parseOffset(r, "storyOffset")
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	documentOffset, err := parseOffset(r, "documentOffset")
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	scope := r.URL.Query().Get("scope")
	if scope != "" && scope != "selected" {
		return web.RespondError(ctx, w, errInvalidOperation, http.StatusBadRequest)
	}
	selected := scope == "selected"
	myWork := r.URL.Query().Get("myWork") == "true"
	if value := r.URL.Query().Get("myWork"); value != "" && value != "true" && value != "false" {
		return web.RespondError(ctx, w, errInvalidOperation, http.StatusBadRequest)
	}
	if myWork && !selected {
		scope = "myWork"
	}
	result := Snapshot{SchemaVersion: 1, Scope: scope, SyncedAt: time.Now().UTC(), WorkspaceID: workspace.ID, UserID: userID, Role: workspace.UserRole, Stories: []Story{}, Documents: []Document{}, Comments: []Comment{}, Teams: []Team{}, Statuses: []Status{}, Complete: map[string]bool{"stories": false, "documents": false, "comments": false, "teams": true, "statuses": true}, Removed: map[string][]uuid.UUID{"stories": {}, "documents": {}}, CommentsCompleteFor: []uuid.UUID{}}
	teamList, err := h.Teams.List(ctx, workspace.ID, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, operationStatus(err))
	}
	for _, team := range teamList {
		result.Teams = append(result.Teams, Team{ID: team.ID, Name: team.Name, Code: team.Code, Color: team.Color})
	}
	statuses, err := h.Statuses.List(ctx, workspace.ID, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, operationStatus(err))
	}
	for _, status := range statuses {
		result.Statuses = append(result.Statuses, Status{ID: status.ID, Name: status.Name, TeamID: status.Team, Category: status.Category, Color: status.Color, OrderIndex: status.OrderIndex, IsDefault: status.IsDefault})
	}
	showSubStories := true
	var assignedToMe *bool
	if myWork {
		assignedToMe = &myWork
	}
	var listedStories []stories.CoreStoryList
	if !selected {
		listedStories, err = h.Stories.List(ctx, workspace.ID, stories.CoreStoryFilters{Limit: storyLimit + 1, Offset: storyOffset, ShowSubStories: &showSubStories, AssignedToMe: assignedToMe})
	}
	if err != nil {
		return web.RespondError(ctx, w, err, operationStatus(err))
	}
	result.Complete["stories"] = !selected && !myWork && storyOffset == 0 && len(listedStories) <= storyLimit
	if len(listedStories) > storyLimit {
		next := storyOffset + storyLimit
		result.NextStoryOffset = &next
	}
	storyIDs := append([]uuid.UUID{}, knownStories...)
	for _, story := range listedStories[:min(len(listedStories), storyLimit)] {
		storyIDs = append(storyIDs, story.ID)
	}
	for _, id := range uniqueIDs(storyIDs) {
		story, err := h.Stories.Get(ctx, id, workspace.ID)
		if errors.Is(err, stories.ErrNotFound) || errors.Is(err, stories.ErrStoryReadForbidden) {
			result.Removed["stories"] = append(result.Removed["stories"], id)
			continue
		}
		if err != nil {
			return web.RespondError(ctx, w, err, operationStatus(err))
		}
		result.Stories = append(result.Stories, mapStory(story))
		comments, more, err := h.Stories.GetComments(ctx, id, workspace.ID, 1, 50)
		if err != nil {
			return web.RespondError(ctx, w, err, operationStatus(err))
		}
		for _, comment := range comments {
			result.Comments = append(result.Comments, mapComment(comment))
		}
		if !more {
			result.CommentsCompleteFor = append(result.CommentsCompleteFor, id)
		}
	}
	result.Complete["comments"] = result.Complete["stories"] && len(result.CommentsCompleteFor) == len(result.Stories)
	limit := documentLimit + 1
	var listedDocuments []documents.CoreDocumentSummary
	if !selected {
		listedDocuments, err = h.Documents.List(ctx, documents.CoreListInput{WorkspaceID: workspace.ID, UserID: userID, Scope: "all", Limit: &limit, Offset: documentOffset})
	}
	if err != nil {
		return web.RespondError(ctx, w, err, operationStatus(err))
	}
	result.Complete["documents"] = !selected && documentOffset == 0 && len(listedDocuments) <= documentLimit
	if len(listedDocuments) > documentLimit {
		next := documentOffset + documentLimit
		result.NextDocumentOffset = &next
	}
	documentIDs := append([]uuid.UUID{}, knownDocuments...)
	for _, document := range listedDocuments[:min(len(listedDocuments), documentLimit)] {
		documentIDs = append(documentIDs, document.ID)
	}
	for _, id := range uniqueIDs(documentIDs) {
		document, err := h.Documents.GetForOffline(ctx, workspace.ID, userID, id)
		if errors.Is(err, documents.ErrNotFound) || errors.Is(err, documents.ErrForbidden) {
			result.Removed["documents"] = append(result.Removed["documents"], id)
			continue
		}
		if err != nil {
			return web.RespondError(ctx, w, err, operationStatus(err))
		}
		result.Documents = append(result.Documents, mapDocument(document))
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return err
	}
	if len(encoded) > maximumSnapshotBytes {
		return web.RespondError(ctx, w, errors.New("download is too large; reduce selected documents"), http.StatusRequestEntityTooLarge)
	}
	return web.Respond(ctx, w, result, http.StatusOK)
}
func parseKnownIDs(r *http.Request, key string) ([]uuid.UUID, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return nil, nil
	}
	if len(value) > 3700 {
		return nil, errInvalidOperation
	}
	pieces := strings.Split(value, ",")
	if len(pieces) > 100 {
		return nil, errInvalidOperation
	}
	ids := make([]uuid.UUID, 0, len(pieces))
	for _, piece := range pieces {
		id, err := uuid.Parse(piece)
		if err != nil || id == uuid.Nil {
			return nil, errInvalidOperation
		}
		ids = append(ids, id)
	}
	return uniqueIDs(ids), nil
}
func uniqueIDs(values []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(values))
	result := make([]uuid.UUID, 0, len(values))
	for _, id := range values {
		if _, exists := seen[id]; !exists {
			seen[id] = struct{}{}
			result = append(result, id)
		}
	}
	return result
}

func parseOffset(r *http.Request, name string) (int, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return 0, nil
	}
	offset, err := strconv.Atoi(value)
	if err != nil || offset < 0 || offset > 1_000_000 {
		return 0, errInvalidOperation
	}
	return offset, nil
}

// Comments allows every thread in a downloaded story to be fetched. A page
// never replaces all cached threads; the client replaces after hasMore=false.
func (h *Handlers) Comments(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Cache-Control", "no-store")
	workspace, _, err := requestIdentity(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	id, err := uuid.Parse(web.Params(r, "id"))
	if err != nil || id == uuid.Nil {
		return web.RespondError(ctx, w, errInvalidOperation, http.StatusBadRequest)
	}
	page := 1
	if raw := r.URL.Query().Get("page"); raw != "" {
		page, err = strconv.Atoi(raw)
		if err != nil || page < 1 || page > 20_000 {
			return web.RespondError(ctx, w, errInvalidOperation, http.StatusBadRequest)
		}
	}
	comments, more, err := h.Stories.GetComments(ctx, id, workspace.ID, page, 50)
	if err != nil {
		return web.RespondError(ctx, w, err, operationStatus(err))
	}
	mapped := make([]Comment, len(comments))
	for i, comment := range comments {
		mapped[i] = mapComment(comment)
	}
	return web.Respond(ctx, w, struct {
		StoryID  uuid.UUID `json:"storyId"`
		Comments []Comment `json:"comments"`
		HasMore  bool      `json:"hasMore"`
		Page     int       `json:"page"`
	}{id, mapped, more, page}, http.StatusOK)
}

func parseSnapshotLimit(r *http.Request, name string, maximum int) (int, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return maximum, nil
	}
	limit, err := strconv.Atoi(value)
	if err != nil || limit < 1 || limit > maximum {
		return 0, errInvalidOperation
	}
	return limit, nil
}
