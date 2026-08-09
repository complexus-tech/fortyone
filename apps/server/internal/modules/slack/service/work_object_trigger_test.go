package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/stretchr/testify/require"
)

func TestHandleEventsPresentsEntityDetailsWithoutDurableQueue(t *testing.T) {
	t.Parallel()

	fixture := newWorkObjectEditFixture(t)
	queue := &eventQueueStub{}
	inbox := &eventInboxCapture{}
	WithEventRuntime(queue, inbox)(fixture.service)

	var calls atomic.Int32
	fixture.service.webClient.client = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		calls.Add(1)
		require.Equal(t, "/api/entity.presentDetails", request.URL.Path)
		return slackWebAPIResponse(`{"ok":true}`), nil
	})}

	response, err := fixture.service.HandleEvents(
		context.Background(),
		[]byte(entityDetailsEventBody("Ev-entity-fast", fixture.stories.story.ID.String())),
	)

	require.NoError(t, err)
	require.Empty(t, response.Challenge)
	require.EqualValues(t, 1, calls.Load())
	require.Zero(t, inbox.registrations)
	require.Empty(t, queue.payloads)
}

func TestHandleEventsKeepsLinkSharedOnDurableQueue(t *testing.T) {
	t.Parallel()

	fixture := newWorkObjectEditFixture(t)
	queue := &eventQueueStub{}
	inbox := &eventInboxCapture{}
	service := newTestService(
		fixture.repo,
		&mockRequestStore{},
		fixture.stories,
		Config{SecretKey: "event-secret"},
	)
	WithEventRuntime(queue, inbox)(service)

	_, err := service.HandleEvents(context.Background(), []byte(
		`{"type":"event_callback","team_id":"T123","event_id":"Ev-link-durable","event":{"type":"link_shared","user":"U123","channel":"C123","message_ts":"1754700000.123","links":[{"domain":"fortyone.app","url":"https://acme.fortyone.app/work/WEB-123"}]}}`,
	))

	require.NoError(t, err)
	require.Equal(t, 1, inbox.registrations)
	require.Len(t, queue.payloads, 1)
	require.Equal(t, "Ev-link-durable", queue.payloads[0].EventID)
}

func TestHandleEventsTreatsInvalidEntityTriggerAsTerminal(t *testing.T) {
	t.Parallel()

	fixture := newWorkObjectEditFixture(t)
	queue := &eventQueueStub{}
	inbox := &eventInboxCapture{}
	WithEventRuntime(queue, inbox)(fixture.service)

	var calls atomic.Int32
	fixture.service.webClient.client = &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		calls.Add(1)
		return slackWebAPIResponse(`{"ok":false,"error":"invalid_trigger_id"}`), nil
	})}

	_, err := fixture.service.HandleEvents(
		context.Background(),
		[]byte(entityDetailsEventBody("Ev-entity-invalid-trigger", fixture.stories.story.ID.String())),
	)

	require.NoError(t, err, "a spent trigger must be acknowledged instead of retried")
	require.EqualValues(t, 1, calls.Load())
	require.Zero(t, inbox.registrations)
	require.Empty(t, queue.payloads)
}

func TestHandleEventsBoundsEntityDetailsToTriggerDeadline(t *testing.T) {
	t.Parallel()

	fixture := newWorkObjectEditFixture(t)
	fixture.service.workObjectTriggerTimeout = 40 * time.Millisecond
	queue := &eventQueueStub{}
	inbox := &eventInboxCapture{}
	WithEventRuntime(queue, inbox)(fixture.service)

	deadline := make(chan time.Time, 1)
	fixture.service.webClient.client = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		value, ok := request.Context().Deadline()
		if ok {
			deadline <- value
		}
		<-request.Context().Done()
		return nil, request.Context().Err()
	})}

	startedAt := time.Now()
	_, err := fixture.service.HandleEvents(
		context.Background(),
		[]byte(entityDetailsEventBody("Ev-entity-deadline", fixture.stories.story.ID.String())),
	)
	elapsed := time.Since(startedAt)

	require.NoError(t, err, "deadline expiry is terminal for the one-use trigger")
	require.Less(t, elapsed, 500*time.Millisecond)
	select {
	case requestDeadline := <-deadline:
		require.LessOrEqual(t, requestDeadline.Sub(startedAt), 100*time.Millisecond)
	case <-time.After(time.Second):
		t.Fatal("entity.presentDetails was not attempted")
	}
	require.Zero(t, inbox.registrations)
	require.Empty(t, queue.payloads)
}

func TestHandleEventsBoundsEntityDetailsInstallationLookup(t *testing.T) {
	t.Parallel()

	fixture := newWorkObjectEditFixture(t)
	fixture.service.workObjectTriggerTimeout = 40 * time.Millisecond
	blockingRepo := &blockingSlackWorkspaceRepo{
		mockRepo: fixture.repo,
		started:  make(chan struct{}, 1),
		release:  make(chan struct{}),
	}
	fixture.service.repo = blockingRepo

	startedAt := time.Now()
	_, err := fixture.service.HandleEvents(
		context.Background(),
		[]byte(entityDetailsEventBody("Ev-entity-install-deadline", fixture.stories.story.ID.String())),
	)
	elapsed := time.Since(startedAt)

	require.NoError(t, err, "installation lookup expiry is terminal for the one-use trigger")
	require.Less(t, elapsed, 500*time.Millisecond)
	select {
	case <-blockingRepo.started:
	case <-time.After(time.Second):
		t.Fatal("entity details installation lookup was not attempted")
	}
}

func TestDispatchSlackWorkObjectEditDetachesAndUsesTriggerOnlyOnce(t *testing.T) {
	t.Parallel()

	fixture := newWorkObjectEditFixture(t)
	fixture.service.workObjectTriggerTimeout = 40 * time.Millisecond
	fixture.payload.View.State.Values = interactionViewStateValues{
		"title": workObjectTextState("title", "Updated within Slack"),
	}

	var calls atomic.Int32
	var startedOnce sync.Once
	started := make(chan struct{})
	done := make(chan struct{}, 1)
	fixture.service.webClient.client = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		calls.Add(1)
		startedOnce.Do(func() { close(started) })
		<-request.Context().Done()
		done <- struct{}{}
		return nil, request.Context().Err()
	})}

	parent, cancel := context.WithCancel(context.Background())
	cancel()
	fixture.service.dispatchSlackWorkObjectEdit(parent, fixture.payload)

	select {
	case <-started:
		// The request-owned context was cancelled, so reaching the provider proves
		// dispatch retained only values and installed its own bounded deadline.
	case <-time.After(time.Second):
		t.Fatal("detached Work Object edit did not reach entity.presentDetails")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Work Object edit did not stop at its trigger deadline")
	}
	time.Sleep(75 * time.Millisecond)
	require.EqualValues(t, 1, calls.Load(), "a failed presentation attempt must not reuse the trigger for edit feedback")
}

func TestDispatchSlackWorkObjectEditPresentsConflictWithinOriginalDeadline(t *testing.T) {
	t.Parallel()

	fixture := newWorkObjectEditFixture(t)
	fixture.service.workObjectTriggerTimeout = 250 * time.Millisecond
	fixture.stories.updateErr = stories.ErrStoryChanged
	fixture.payload.View.State.Values = interactionViewStateValues{
		"title": workObjectTextState("title", "Conflicting title"),
	}

	presented := make(chan SlackEntityDetailsRequest, 1)
	fixture.service.webClient.client = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		var payload SlackEntityDetailsRequest
		if err := decodeJSONBody(request.Body, &payload); err != nil {
			return nil, err
		}
		presented <- payload
		return slackWebAPIResponse(`{"ok":true}`), nil
	})}

	fixture.service.dispatchSlackWorkObjectEdit(context.Background(), fixture.payload)

	select {
	case payload := <-presented:
		require.NotNil(t, payload.Error)
		require.Equal(t, "edit_error", payload.Error.Status)
		require.Contains(t, payload.Error.CustomMessage, "changed while you were editing")
	case <-time.After(time.Second):
		t.Fatal("conflict feedback was not presented within the trigger deadline")
	}
}

func TestEntityDetailsTerminalErrorClassification(t *testing.T) {
	t.Parallel()

	for _, code := range []string{"invalid_trigger_id", "trigger_expired", "trigger_exchanged"} {
		code := code
		t.Run(code, func(t *testing.T) {
			t.Parallel()
			err := &slackEntityDetailsPresentationError{cause: &SlackAPIError{Method: "entity.presentDetails", Code: code}}
			require.True(t, isSlackEntityDetailsTerminalError(err))
			require.True(t, slackEntityDetailsPresentationWasAttempted(err))
		})
	}
	require.True(t, isSlackEntityDetailsTerminalError(context.DeadlineExceeded))
	require.False(t, isSlackEntityDetailsTerminalError(&SlackAPIError{Method: "entity.presentDetails", Code: "internal_error"}))
}

func TestEventProcessorIgnoresLegacyQueuedEntityDetailsTrigger(t *testing.T) {
	t.Parallel()

	repo := newEventRepositoryStub()
	store := newEventStoreStub()
	processor := newTestEventProcessor(
		t,
		repo,
		store,
		&assistantStub{},
		&accessCheckerStub{allowed: true},
		&messageSenderStub{},
	)

	rawBody := strings.Replace(entityDetailsEventBody("Ev-legacy-entity", "story-id"), `"team_id":"T123"`, `"team_id":"T1"`, 1)
	err := processor.Process(context.Background(), []byte(rawBody))

	require.NoError(t, err)
	assertSingleInboundStatus(t, store, "ignored")
	require.Zero(t, repo.getInstallationCalls)
	require.Empty(t, store.outboundInputs)
}

func entityDetailsEventBody(eventID, storyID string) string {
	return fmt.Sprintf(
		`{"type":"event_callback","team_id":"T123","event_id":%q,"event":{"type":"entity_details_requested","user":"U123","channel":"C123","message_ts":"1754700000.123","trigger_id":"trigger-123","external_ref":{"id":%q,"type":"story"},"entity_url":"https://acme.fortyone.app/work/WEB-123"}}`,
		eventID,
		"acme:"+storyID,
	)
}

func slackWebAPIResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func decodeJSONBody(body io.Reader, destination any) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, destination)
}
