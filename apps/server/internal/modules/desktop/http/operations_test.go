package desktophttp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type mutationStories struct {
	StoryService
	updatedAt       time.Time
	patch           stories.StoryPatch
	created         stories.CoreNewStory
	mediaReferences []uuid.UUID
	reconciled      bool
	err             error
}

func (s *mutationStories) Create(_ context.Context, input stories.CoreNewStory, _ uuid.UUID) (stories.CoreSingleStory, error) {
	s.created = input
	return stories.CoreSingleStory{ID: uuid.New()}, s.err
}

func (s *mutationStories) UpdatePatchWithMediaReconciliationIfUnchanged(_ context.Context, _, _ uuid.UUID, expected time.Time, patch stories.StoryPatch, references []uuid.UUID) error {
	s.updatedAt, s.patch, s.mediaReferences, s.reconciled = expected, patch, references, true
	return s.err
}

func (s *mutationStories) UpdatePatchIfUnchanged(_ context.Context, _, _ uuid.UUID, expected time.Time, patch stories.StoryPatch) error {
	s.updatedAt, s.patch = expected, patch
	return s.err
}

type mutationDocuments struct {
	DocumentService
	current documents.CoreDocument
	update  documents.CoreUpdateInput
	updates int
}

func (s *mutationDocuments) Get(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (documents.CoreDocument, error) {
	return s.current, nil
}
func (s *mutationDocuments) Update(_ context.Context, input documents.CoreUpdateInput) (documents.CoreDocument, error) {
	s.update = input
	s.updates++
	return s.current, nil
}

func TestStoryMutationUsesClientBaseAndRejectsTargetFields(t *testing.T) {
	repository := &mutationStories{err: stories.ErrStoryChanged}
	h := &Handlers{Config: Config{Stories: repository}}
	base := "2026-10-14T09:10:11.123456Z"
	operation := Operation{ID: uuid.New(), Kind: "story.update", ResourceID: uuid.New(), BaseVersion: &base, Payload: json.RawMessage(`{"title":"Offline change","assigneeId":null}`)}
	apply, err := h.prepareMutation(operation, uuid.New(), uuid.New(), "acme")
	require.NoError(t, err)
	_, err = apply(t.Context())
	require.ErrorIs(t, err, stories.ErrStoryChanged)
	require.Equal(t, http.StatusConflict, operationStatus(err))
	expected, err := time.Parse(time.RFC3339Nano, base)
	require.NoError(t, err)
	require.Equal(t, expected, repository.updatedAt)
	for _, payload := range []string{`{}`, `{"id":"bad"}`, `{"teamId":"bad"}`, `{"title":null}`, `{"priority":"invalid"}`, `{"unexpected":true}`} {
		operation.Payload = json.RawMessage(payload)
		_, err := h.prepareMutation(operation, uuid.New(), uuid.New(), "acme")
		require.Error(t, err, payload)
	}
	operation.BaseVersion = nil
	_, err = h.prepareMutation(operation, uuid.New(), uuid.New(), "acme")
	require.Error(t, err)
}

func TestDocumentMutationPreservesBaseRevisionAndRejectsRestoredOrCollaborativeState(t *testing.T) {
	base := "3"
	for _, test := range []struct {
		name          string
		epoch         int64
		collaborative bool
		conflict      bool
	}{
		{"unchanged epoch", 4, false, false},
		{"restored epoch", 5, false, true},
		{"collaborative", 4, true, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			service := &mutationDocuments{current: documents.CoreDocument{ID: uuid.New(), Revision: 8, CollaborationEpoch: test.epoch, Collaborative: test.collaborative}}
			h := &Handlers{Config: Config{Documents: service}}
			apply, err := h.prepareMutation(Operation{Kind: "document.update", ResourceID: service.current.ID, BaseVersion: &base, Payload: json.RawMessage(`{"title":"Local title","collaborationEpoch":4}`)}, uuid.New(), uuid.New(), "acme")
			require.NoError(t, err)
			_, err = apply(t.Context())
			if test.conflict {
				require.ErrorIs(t, err, documents.ErrConflict)
				require.Zero(t, service.updates)
			} else {
				require.NoError(t, err)
				require.Equal(t, int64(3), service.update.ExpectedRevision, "never substitute latest server revision for offline base")
			}
		})
	}
}

func TestOperationCanonicalizationAndStrictEnvelope(t *testing.T) {
	operation := Operation{ID: uuid.New(), ResourceID: uuid.New(), Kind: "story.create", Payload: json.RawMessage(`{"title":"A", "priority":"Low"}`)}
	first, err := canonicalOperation(operation)
	require.NoError(t, err)
	operation.Payload = json.RawMessage(`{ "priority": "Low", "title": "A" }`)
	second, err := canonicalOperation(operation)
	require.NoError(t, err)
	require.Equal(t, string(first), string(second))
	operation.Payload = json.RawMessage(`null`)
	_, err = canonicalOperation(operation)
	require.Error(t, err)
	operation.Payload = json.RawMessage(`[]`)
	_, err = canonicalOperation(operation)
	require.Error(t, err)
	operation.Kind = "document.delete"
	_, err = canonicalOperation(operation)
	require.Error(t, err)
}

func TestAttachmentUploadRejectsUnsafeInputBeforeStorage(t *testing.T) {
	for _, test := range []struct {
		name  string
		input attachmentUpload
		want  error
	}{
		{"valid", attachmentUpload{FileName: "notes.txt", ContentType: "text/plain", Data: base64.StdEncoding.EncodeToString([]byte("notes")), TargetKind: "story"}, nil},
		{"path", attachmentUpload{FileName: "../notes.txt", ContentType: "text/plain", Data: "bm90ZXM=", TargetKind: "story"}, errInvalidOperation},
		{"empty", attachmentUpload{FileName: "notes.txt", ContentType: "text/plain", TargetKind: "story"}, attachments.ErrInvalidFile},
		{"bad target", attachmentUpload{FileName: "notes.txt", ContentType: "text/plain", Data: "bm90ZXM=", TargetKind: "workspace"}, errInvalidOperation},
		{"bad encoding", attachmentUpload{FileName: "notes.txt", ContentType: "text/plain", Data: "%%", TargetKind: "story"}, attachments.ErrInvalidFile},
		{"too large", attachmentUpload{FileName: "notes.txt", ContentType: "text/plain", Data: strings.Repeat("a", base64.StdEncoding.EncodedLen(maximumAttachmentBytes)+1), TargetKind: "story"}, attachments.ErrFileTooLarge},
	} {
		t.Run(test.name, func(t *testing.T) {
			raw, err := json.Marshal(test.input)
			require.NoError(t, err)
			_, _, err = decodeAttachment(raw)
			require.ErrorIs(t, err, test.want)
		})
	}
}

func TestKnownIDsAreBoundedAndDeduplicated(t *testing.T) {
	id := uuid.New().String()
	request := httptest.NewRequest(http.MethodGet, "/?storyIds="+id+","+id, nil)
	ids, err := parseKnownIDs(request, "storyIds")
	require.NoError(t, err)
	require.Len(t, ids, 1)
	for _, value := range []string{uuid.Nil.String(), "invalid", strings.Repeat(id+",", 101)} {
		_, err := parseKnownIDs(httptest.NewRequest(http.MethodGet, "/?storyIds="+value, nil), "storyIds")
		require.Error(t, err)
	}
}

func TestSnapshotLimitsRejectUnboundedRequests(t *testing.T) {
	for _, value := range []string{"0", "-1", "101", "1000000", "oops"} {
		_, err := parseSnapshotLimit(httptest.NewRequest(http.MethodGet, "/?storyLimit="+value, nil), "storyLimit", 100)
		require.Error(t, err, value)
	}
	limit, err := parseSnapshotLimit(httptest.NewRequest(http.MethodGet, "/?documentLimit=1", nil), "documentLimit", 50)
	require.NoError(t, err)
	require.Equal(t, 1, limit)
}
