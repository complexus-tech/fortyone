package fortyone

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

const (
	testWorkspaceID      = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	testStoryID          = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	testWebhookID        = "11111111-1111-4111-8111-111111111111"
	testWebhookTimestamp = "1787920000"
	testWebhookSecret    = "whsec_UlJSUlJSUlJSUlJSUlJSUlJSUlJSUlJSUlJSUlJSUlI="
	testWebhookSignature = "v1,fmCSZfMbiTh50bHta4Wg4YSYK94JW+6U9d0vvQHQQbk="
	testWebhookBody      = `{"value":"line one\r\nline two"}`
	testAuthorization    = "Bearer f41_pat_v1_test"
)

func TestNewAuthenticatesGeneratedRequests(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if got := request.Header.Get("Authorization"); got != testAuthorization {
			t.Errorf("Authorization = %q, want %q", got, testAuthorization)
		}
		if got := request.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept = %q, want application/json", got)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(writer, `{"data":{"id":"`+testStoryID+`","workspaceId":"`+testWorkspaceID+`","teamId":"dddddddd-dddd-4ddd-8ddd-dddddddddddd","sequenceId":1,"reference":"ENG-1","title":"Typed SDK","priority":"none","autoSchedulingEnabled":false,"autoSchedulingLocked":false,"autoSchedulingStatus":"disabled","labels":[],"createdAt":"2026-08-28T12:00:00Z","updatedAt":"2026-08-28T12:00:00Z"}}`)
	}))
	t.Cleanup(server.Close)

	client, err := New(Config{Token: "f41_pat_v1_test", BaseURL: server.URL, AllowInsecureLoopback: true})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	response, err := client.GetStoryWithResponse(t.Context(), uuid.MustParse(testWorkspaceID), uuid.MustParse(testStoryID))
	if err != nil {
		t.Fatalf("GetStoryWithResponse() error = %v", err)
	}
	if response.JSON200 == nil || response.JSON200.Data.Id.String() != testStoryID {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestGeneratedClientExposesTypedCollaborationReads(t *testing.T) {
	t.Parallel()
	teamID := uuid.New()
	labelID := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/v1/workspaces/"+testWorkspaceID+"/labels" || request.URL.Query().Get("teamId") != teamID.String() {
			t.Errorf("unexpected label request URL: %s", request.URL.String())
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(writer, `{"data":[{"id":"%s","workspaceId":"%s","teamId":"%s","name":"API","color":"#123456","createdAt":"2026-08-28T12:00:00Z","updatedAt":"2026-08-28T12:00:00Z"}],"meta":{"hasMore":false}}`, labelID, testWorkspaceID, teamID)
	}))
	t.Cleanup(server.Close)

	client, err := New(Config{Token: "f41_pat_v1_test", BaseURL: server.URL, AllowInsecureLoopback: true})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	response, err := client.ListLabelsWithResponse(t.Context(), uuid.MustParse(testWorkspaceID), &ListLabelsParams{TeamId: teamID})
	if err != nil {
		t.Fatalf("ListLabelsWithResponse() error = %v", err)
	}
	if response.JSON200 == nil || len(response.JSON200.Data) != 1 || response.JSON200.Data[0].Id != labelID {
		t.Fatalf("unexpected label response: %#v", response)
	}
}

func TestNewRejectsUnsafeConfiguration(t *testing.T) {
	t.Parallel()
	for index, testCase := range []Config{
		{Token: ""},
		{Token: "token with spaces"},
		{Token: "valid", BaseURL: "http://api.example.com"},
		{Token: "valid", BaseURL: "https://user@example.com"},
	} {
		if _, err := New(testCase); err == nil {
			t.Fatalf("test case %d New() error = nil, want validation error", index)
		}
	}
	if rendered := fmt.Sprintf("%+v %#v", Config{Token: "do-not-log"}, Config{Token: "do-not-log"}); strings.Contains(rendered, "do-not-log") {
		t.Fatalf("Config formatting exposed token: %q", rendered)
	}
}

func TestNewDoesNotForwardCredentialsAcrossRedirects(t *testing.T) {
	t.Parallel()
	var targetCalls atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		targetCalls.Add(1)
	}))
	t.Cleanup(target.Close)
	source := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, target.URL, http.StatusTemporaryRedirect)
	}))
	t.Cleanup(source.Close)

	client, err := New(Config{Token: "f41_pat_v1_test", BaseURL: source.URL, AllowInsecureLoopback: true})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	response, err := client.GetWorkspaceWithResponse(t.Context(), uuid.MustParse(testWorkspaceID))
	if err != nil {
		t.Fatalf("GetWorkspaceWithResponse() error = %v", err)
	}
	if response.StatusCode() != http.StatusTemporaryRedirect || targetCalls.Load() != 0 {
		t.Fatalf("redirect status/calls = %d/%d", response.StatusCode(), targetCalls.Load())
	}
}

func TestRetryTransportRetriesSafeTransientResponse(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	delays := make(chan time.Duration, 1)
	policy, err := (RetryPolicy{MaxAttempts: 2, BaseDelay: 10 * time.Millisecond, MaxDelay: 10 * time.Millisecond}).normalized()
	if err != nil {
		t.Fatalf("normalized() error = %v", err)
	}
	policy.random = func() float64 { return 1 }
	policy.sleep = func(_ context.Context, delay time.Duration) error {
		delays <- delay
		return nil
	}
	transport := newRetryTransport(roundTripFunc(func(request *http.Request) (*http.Response, error) {
		status := http.StatusServiceUnavailable
		if calls.Add(1) == 2 {
			status = http.StatusNoContent
		}
		return &http.Response{StatusCode: status, Header: make(http.Header), Body: http.NoBody, Request: request}, nil
	}), policy)

	response, err := transport.RoundTrip(mustRequest(t, http.MethodGet))
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	if response.StatusCode != http.StatusNoContent || calls.Load() != 2 || <-delays != 10*time.Millisecond {
		t.Fatalf("unexpected retry result: status=%d calls=%d", response.StatusCode, calls.Load())
	}
}

func TestRetryTransportDoesNotRetryWritesOrPermanentNetworkErrors(t *testing.T) {
	t.Parallel()
	policy, err := (RetryPolicy{}).normalized()
	if err != nil {
		t.Fatalf("normalized() error = %v", err)
	}
	var calls atomic.Int32
	transport := newRetryTransport(roundTripFunc(func(request *http.Request) (*http.Response, error) {
		calls.Add(1)
		if request.Method == http.MethodPost {
			return &http.Response{StatusCode: http.StatusServiceUnavailable, Header: make(http.Header), Body: http.NoBody, Request: request}, nil
		}
		return nil, errors.New("invalid certificate")
	}), policy)

	if _, err := transport.RoundTrip(mustRequest(t, http.MethodPost)); err != nil {
		t.Fatalf("POST RoundTrip() error = %v", err)
	}
	if _, err := transport.RoundTrip(mustRequest(t, http.MethodGet)); err == nil {
		t.Fatal("GET RoundTrip() error = nil, want permanent error")
	}
	if calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2", calls.Load())
	}
}

func TestNewAPIErrorMapsSafeEnvelope(t *testing.T) {
	t.Parallel()
	header := http.Header{"Retry-After": {"8"}, "X-Request-Id": {"req_header"}}
	errorResponse := NewAPIError(http.StatusTooManyRequests, header, []byte(`{"error":{"code":"rate_limited","message":"Try again later","requestId":"req_body","fields":[{"field":"limit","code":"exhausted","message":"Too many"}]}}`))
	if errorResponse.Code != "rate_limited" || errorResponse.RequestID != "req_body" || errorResponse.RetryAfterSeconds != 8 || len(errorResponse.Fields) != 1 {
		t.Fatalf("unexpected API error: %#v", errorResponse)
	}
	if strings.Contains(errorResponse.Error(), "exhausted") {
		t.Fatalf("Error() exposed field detail: %q", errorResponse.Error())
	}
}

func TestStoryPagerUsesOpaqueCursor(t *testing.T) {
	t.Parallel()
	next := "opaque-next"
	fake := &fakeStoryPageClient{responses: []*ListStoriesResponse{
		{HTTPResponse: &http.Response{StatusCode: http.StatusOK}, JSON200: &ComponentsResourcesStoryPageResponse{Data: []ComponentsResourcesStory{{Id: uuid.MustParse(testStoryID)}}, Meta: ComponentsCommonPageMeta{HasMore: true, NextCursor: &next}}},
		{HTTPResponse: &http.Response{StatusCode: http.StatusOK}, JSON200: &ComponentsResourcesStoryPageResponse{Meta: ComponentsCommonPageMeta{HasMore: false}}},
	}}
	pager, err := NewStoryPager(fake, uuid.MustParse(testWorkspaceID), StoryPaginationOptions{Limit: 1})
	if err != nil {
		t.Fatalf("NewStoryPager() error = %v", err)
	}
	if _, ok, err := pager.NextPage(t.Context()); err != nil || !ok {
		t.Fatalf("first NextPage() = ok %v, error %v", ok, err)
	}
	if _, ok, err := pager.NextPage(t.Context()); err != nil || !ok {
		t.Fatalf("second NextPage() = ok %v, error %v", ok, err)
	}
	if fake.cursors[0] != "" || fake.cursors[1] != next {
		t.Fatalf("cursors = %#v", fake.cursors)
	}
}

func TestStoryPagerRejectsRepeatedCursor(t *testing.T) {
	t.Parallel()
	repeated := "same-cursor"
	fake := &fakeStoryPageClient{responses: []*ListStoriesResponse{
		{HTTPResponse: &http.Response{StatusCode: http.StatusOK}, JSON200: &ComponentsResourcesStoryPageResponse{Meta: ComponentsCommonPageMeta{HasMore: true, NextCursor: &repeated}}},
		{HTTPResponse: &http.Response{StatusCode: http.StatusOK}, JSON200: &ComponentsResourcesStoryPageResponse{Meta: ComponentsCommonPageMeta{HasMore: true, NextCursor: &repeated}}},
	}}
	pager, err := NewStoryPager(fake, uuid.MustParse(testWorkspaceID), StoryPaginationOptions{})
	if err != nil {
		t.Fatalf("NewStoryPager() error = %v", err)
	}
	if _, _, err := pager.NextPage(t.Context()); err != nil {
		t.Fatalf("first NextPage() error = %v", err)
	}
	if _, _, err := pager.NextPage(t.Context()); !errors.Is(err, ErrInvalidPaginationCursor) {
		t.Fatalf("second NextPage() error = %v, want invalid cursor", err)
	}
}

func TestWebhookVerifierSharedFixtureAndFailures(t *testing.T) {
	t.Parallel()
	verifier, err := NewWebhookVerifier(testWebhookSecret, 0)
	if err != nil {
		t.Fatalf("NewWebhookVerifier() error = %v", err)
	}
	verifier.now = func() time.Time { return time.Unix(1787920000, 0) }
	header := http.Header{
		"Webhook-Id":        {testWebhookID},
		"Webhook-Timestamp": {testWebhookTimestamp},
		"Webhook-Signature": {"v1,AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA= " + testWebhookSignature},
	}
	verified, err := verifier.Verify([]byte(testWebhookBody), header)
	if err != nil || verified.ID.String() != testWebhookID {
		t.Fatalf("Verify() = %#v, %v", verified, err)
	}

	header.Set("Webhook-Signature", testWebhookSignature)
	_, err = verifier.Verify([]byte(`{"value":"changed"}`), header)
	var verificationError *WebhookVerificationError
	if !errors.As(err, &verificationError) || verificationError.Code != WebhookInvalidSignature {
		t.Fatalf("tampered Verify() error = %v", err)
	}
	verifier.now = func() time.Time { return time.Unix(1787920301, 0) }
	_, err = verifier.Verify([]byte(testWebhookBody), header)
	if !errors.As(err, &verificationError) || verificationError.Code != WebhookStaleTimestamp {
		t.Fatalf("stale Verify() error = %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func mustRequest(t *testing.T, method string) *http.Request {
	t.Helper()
	request, err := http.NewRequestWithContext(t.Context(), method, "https://api.fortyone.app/test", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext() error = %v", err)
	}
	return request
}

type fakeStoryPageClient struct {
	responses []*ListStoriesResponse
	cursors   []string
}

func (fake *fakeStoryPageClient) ListStoriesWithResponse(_ context.Context, _ ComponentsCommonWorkspaceId, params *ListStoriesParams, _ ...RequestEditorFn) (*ListStoriesResponse, error) {
	cursor := ""
	if params.Cursor != nil {
		cursor = *params.Cursor
	}
	fake.cursors = append(fake.cursors, cursor)
	response := fake.responses[0]
	fake.responses = fake.responses[1:]
	return response, nil
}
