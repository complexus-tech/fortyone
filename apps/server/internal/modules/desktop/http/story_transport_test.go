package desktophttp

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDesktopCreatePreservesTheFullProjectsStoryRequest(t *testing.T) {
	service := &mutationStories{}
	h := &Handlers{Config: Config{Stories: service}}
	actor, workspace, team, parent, objective, status, assignee, sprint, keyResult, label := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	payload, err := json.Marshal(map[string]any{
		"title": "Full app story", "teamId": team, "parentId": parent, "objectiveId": objective,
		"statusId": status, "assigneeId": assignee, "sprintId": sprint, "keyResultId": keyResult,
		"priority": "High", "description": "Text", "descriptionHTML": "<p>Text</p>",
		"estimateValue": 3, "estimatedDurationMinutes": 60, "minimumFocusBlockMinutes": 15,
		"autoSchedulingEnabled": true, "labelIds": []uuid.UUID{label}, "customFieldValues": []any{},
		"startDate": "2026-10-15", "endDate": "2026-10-16", "idempotencyKey": "normal-form-key",
	})
	require.NoError(t, err)
	operation := Operation{ID: uuid.New(), Kind: "story.create", ResourceID: uuid.New(), Payload: payload}
	apply, err := h.prepareMutation(operation, workspace, actor, "acme")
	require.NoError(t, err)
	_, err = apply(t.Context())
	require.NoError(t, err)
	require.Equal(t, team, service.created.Team)
	require.Equal(t, &actor, service.created.Reporter)
	require.Equal(t, &parent, service.created.Parent)
	require.Equal(t, &objective, service.created.Objective)
	require.Equal(t, &status, service.created.Status)
	require.Equal(t, &assignee, service.created.Assignee)
	require.Equal(t, &sprint, service.created.Sprint)
	require.Equal(t, &keyResult, service.created.KeyResult)
	require.Equal(t, []uuid.UUID{label}, service.created.LabelIDs)
	require.Equal(t, int16(3), *service.created.EstimateValue)
	require.Equal(t, 60, *service.created.EstimatedDurationMinutes)
	require.Equal(t, 15, *service.created.MinimumFocusBlockMinutes)
	require.True(t, service.created.AutoSchedulingEnabled)
	require.Equal(t, "2026-10-15", service.created.StartDate.Format(time.DateOnly))
	require.Equal(t, "2026-10-16", service.created.EndDate.Format(time.DateOnly))
	require.Equal(t, "desktop:"+actor.String()+":"+operation.ID.String(), *service.created.CreationKey)
}

func TestDesktopPatchSharesProjectsNullableDatesSchedulingAndReferenceValidation(t *testing.T) {
	patch, err := decodeStoryPatch(json.RawMessage(`{"estimateValue":3,"estimatedDurationMinutes":60,"minimumFocusBlockMinutes":15,"autoSchedulingEnabled":true,"autoSchedulingLocked":false,"parentId":null,"objectiveId":null,"sprintId":null,"keyResultId":null,"startDate":"2026-10-15","endDate":null,"reconcileDescriptionMedia":false}`))
	require.NoError(t, err)
	duration, present := patch.EstimatedDurationMinutes.Value()
	require.True(t, present)
	require.Equal(t, 60, *duration)
	start, present := patch.StartDate.Value()
	require.True(t, present)
	require.Equal(t, "2026-10-15", start.Format(time.DateOnly))
	end, present := patch.EndDate.Value()
	require.True(t, present)
	require.Nil(t, end)
	for _, raw := range []string{`{"statusId":null}`, `{"estimatedDurationMinutes":"60"}`, `{"id":"resource-metadata"}`, `{"teamId":"resource-metadata"}`} {
		_, err := decodeStoryPatch(json.RawMessage(raw))
		require.Error(t, err, raw)
	}
}

func TestDesktopDescriptionReconciliationUsesTheNormalReferenceParser(t *testing.T) {
	service := &mutationStories{}
	handler := &Handlers{Config: Config{Stories: service}}
	story := uuid.New()
	version := "2026-10-15T12:00:00.123456Z"
	// Stable media URLs use the existing handler's exact workspace/story shape.
	html := "<p>Plain description edits also reconcile safely</p>"
	payload, err := json.Marshal(map[string]any{"descriptionHTML": html, "reconcileDescriptionMedia": true})
	require.NoError(t, err)
	apply, err := handler.prepareMutation(Operation{ID: uuid.New(), Kind: "story.update", ResourceID: story, BaseVersion: &version, Payload: payload}, uuid.New(), uuid.New(), "acme")
	require.NoError(t, err)
	_, err = apply(t.Context())
	require.NoError(t, err)
	require.True(t, service.reconciled)
	require.Empty(t, service.mediaReferences)
	for _, raw := range []string{`{"reconcileDescriptionMedia":true}`, `{"descriptionHTML":null,"reconcileDescriptionMedia":true}`, `{"descriptionHTML":"<img data-attachment-id='invalid' src='/wrong'>","reconcileDescriptionMedia":true}`} {
		_, err = handler.prepareMutation(Operation{ID: uuid.New(), Kind: "story.update", ResourceID: story, BaseVersion: &version, Payload: json.RawMessage(raw)}, uuid.New(), uuid.New(), "acme")
		require.Error(t, err, raw)
	}
}
