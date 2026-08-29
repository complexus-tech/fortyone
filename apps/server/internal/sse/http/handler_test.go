package ssehttp

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const (
	testBrowserSessionToken = "opaque-browser-session"
	testWorkspaceSlug       = "engineering"
)

type sessionResolution struct {
	userID uuid.UUID
	ok     bool
	err    error
}

type controllableSessionResolver struct {
	mu          sync.Mutex
	resolutions []sessionResolution
	calls       int
	tokens      []string
}

func (resolver *controllableSessionResolver) Resolve(
	_ context.Context,
	r *http.Request,
) (uuid.UUID, bool, error) {
	resolver.mu.Lock()
	defer resolver.mu.Unlock()

	cookie, err := r.Cookie("fortyone_session")
	if err == nil {
		resolver.tokens = append(resolver.tokens, cookie.Value)
	} else {
		resolver.tokens = append(resolver.tokens, "")
	}

	index := resolver.calls
	resolver.calls++
	if index >= len(resolver.resolutions) {
		return uuid.Nil, false, errors.New("unexpected browser-session resolution")
	}
	result := resolver.resolutions[index]
	return result.userID, result.ok, result.err
}

func (resolver *controllableSessionResolver) observed() (int, []string) {
	resolver.mu.Lock()
	defer resolver.mu.Unlock()
	return resolver.calls, append([]string(nil), resolver.tokens...)
}

type sessionResolverFunc func(context.Context, *http.Request) (uuid.UUID, bool, error)

func (resolve sessionResolverFunc) Resolve(
	ctx context.Context,
	r *http.Request,
) (uuid.UUID, bool, error) {
	return resolve(ctx, r)
}

type workspaceResolverFunc func(context.Context, string, uuid.UUID) (mid.WorkspaceInfo, error)

func (resolve workspaceResolverFunc) ResolveCurrentWorkspace(
	ctx context.Context,
	slug string,
	userID uuid.UUID,
) (mid.WorkspaceInfo, error) {
	return resolve(ctx, slug, userID)
}

func (workspaceResolverFunc) RecordWorkspaceAccess(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func TestServeStreamRevalidatesBrowserSessionBeforeEveryDataEvent(t *testing.T) {
	userID := uuid.New()
	workspaceID := uuid.New()
	resolver := &controllableSessionResolver{resolutions: []sessionResolution{
		{userID: userID, ok: true},
		{err: platformauth.ErrInvalidBrowserSession},
	}}
	messages := make(chan []byte, 2)
	messages <- []byte(`{"visibility":"allowed"}`)
	messages <- []byte(`{"visibility":"revoked"}`)

	recorder := httptest.NewRecorder()
	handler := newTestStreamHandler(t, resolver, workspaceID, make(chan time.Time))
	require.NoError(t, handler.serveStream(
		context.Background(),
		recorder,
		newStreamRequest(context.Background()),
		userID,
		workspaceID,
		testWorkspaceSlug,
		context.Background(),
		messages,
	))

	body := recorder.Body.String()
	require.Contains(t, body, `data: {"visibility":"allowed"}`)
	require.NotContains(t, body, `{"visibility":"revoked"}`)
	require.NotContains(t, body, platformauth.ErrInvalidBrowserSession.Error())
	calls, tokens := resolver.observed()
	require.Equal(t, 2, calls)
	require.Equal(t, []string{testBrowserSessionToken, testBrowserSessionToken}, tokens)
}

func TestServeStreamClosesOnDeactivationBeforePendingDataEvent(t *testing.T) {
	userID := uuid.New()
	workspaceID := uuid.New()
	resolver := &controllableSessionResolver{resolutions: []sessionResolution{
		{err: platformauth.ErrInvalidBrowserSession},
	}}
	messages := make(chan []byte, 1)
	messages <- []byte(`{"private":"must-not-leak"}`)

	recorder := httptest.NewRecorder()
	handler := newTestStreamHandler(t, resolver, workspaceID, make(chan time.Time))
	require.NoError(t, handler.serveStream(
		context.Background(),
		recorder,
		newStreamRequest(context.Background()),
		userID,
		workspaceID,
		testWorkspaceSlug,
		context.Background(),
		messages,
	))

	require.NotContains(t, recorder.Body.String(), `{"private":"must-not-leak"}`)
	calls, tokens := resolver.observed()
	require.Equal(t, 1, calls)
	require.Equal(t, []string{testBrowserSessionToken}, tokens)
}

func TestServeStreamClosesOnVersionMismatchAtNextIdleKeepalive(t *testing.T) {
	userID := uuid.New()
	workspaceID := uuid.New()
	resolver := &controllableSessionResolver{resolutions: []sessionResolution{
		{err: platformauth.ErrInvalidBrowserSession},
	}}
	keepAlive := make(chan time.Time, 1)
	keepAlive <- time.Unix(1, 0)

	recorder := httptest.NewRecorder()
	handler := newTestStreamHandler(t, resolver, workspaceID, keepAlive)
	require.NoError(t, handler.serveStream(
		context.Background(),
		recorder,
		newStreamRequest(context.Background()),
		userID,
		workspaceID,
		testWorkspaceSlug,
		context.Background(),
		make(chan []byte),
	))

	require.NotContains(t, recorder.Body.String(), ":keep-alive")
	calls, tokens := resolver.observed()
	require.Equal(t, 1, calls)
	require.Equal(t, []string{testBrowserSessionToken}, tokens)
}

func TestServeStreamClosesWithoutLeakingResolverFailure(t *testing.T) {
	workspaceID := uuid.New()
	resolverFailure := errors.New("database connection contains private diagnostics")
	resolver := &controllableSessionResolver{resolutions: []sessionResolution{
		{err: resolverFailure},
	}}
	messages := make(chan []byte, 1)
	messages <- []byte(`{"private":"must-not-leak"}`)

	recorder := httptest.NewRecorder()
	handler := newTestStreamHandler(t, resolver, workspaceID, make(chan time.Time))
	require.NoError(t, handler.serveStream(
		context.Background(),
		recorder,
		newStreamRequest(context.Background()),
		uuid.New(),
		workspaceID,
		testWorkspaceSlug,
		context.Background(),
		messages,
	))

	body := recorder.Body.String()
	require.NotContains(t, body, `{"private":"must-not-leak"}`)
	require.NotContains(t, body, resolverFailure.Error())
}

func TestServeStreamCancellationDuringRevalidationPreventsPayload(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	workspaceID := uuid.New()
	started := make(chan struct{})
	resolver := sessionResolverFunc(func(ctx context.Context, _ *http.Request) (uuid.UUID, bool, error) {
		close(started)
		<-ctx.Done()
		return uuid.Nil, false, ctx.Err()
	})
	messages := make(chan []byte, 1)
	messages <- []byte(`{"private":"must-not-leak"}`)

	recorder := httptest.NewRecorder()
	handler := newTestStreamHandler(t, resolver, workspaceID, make(chan time.Time))
	done := make(chan error, 1)
	go func() {
		done <- handler.serveStream(
			ctx,
			recorder,
			newStreamRequest(ctx),
			uuid.New(),
			workspaceID,
			testWorkspaceSlug,
			context.Background(),
			messages,
		)
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("browser-session revalidation did not start")
	}
	cancel()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("stream did not close after request cancellation")
	}

	require.NotContains(t, recorder.Body.String(), `{"private":"must-not-leak"}`)
}

func TestServeStreamClosesWhenWorkspaceMembershipIsRevokedBeforePayload(t *testing.T) {
	userID := uuid.New()
	workspaceID := uuid.New()
	resolver := &controllableSessionResolver{resolutions: []sessionResolution{
		{userID: userID, ok: true},
	}}
	messages := make(chan []byte, 1)
	messages <- []byte(`{"workspace":"must-not-leak"}`)

	recorder := httptest.NewRecorder()
	handler := newTestStreamHandler(t, resolver, workspaceID, make(chan time.Time))
	handler.WorkspaceAccess = workspaceResolverFunc(func(
		_ context.Context,
		slug string,
		resolvedUserID uuid.UUID,
	) (mid.WorkspaceInfo, error) {
		require.Equal(t, testWorkspaceSlug, slug)
		require.Equal(t, userID, resolvedUserID)
		return mid.WorkspaceInfo{}, mid.ErrWorkspaceAccessDenied
	})

	require.NoError(t, handler.serveStream(
		context.Background(),
		recorder,
		newStreamRequest(context.Background()),
		userID,
		workspaceID,
		testWorkspaceSlug,
		context.Background(),
		messages,
	))

	require.NotContains(t, recorder.Body.String(), `{"workspace":"must-not-leak"}`)
}

func TestServeStreamEmitsDataWhenBrowserSessionRemainsCurrent(t *testing.T) {
	userID := uuid.New()
	workspaceID := uuid.New()
	resolver := &controllableSessionResolver{resolutions: []sessionResolution{
		{userID: userID, ok: true},
	}}
	messages := make(chan []byte, 1)
	messages <- []byte(`{"visibility":"allowed"}`)
	recorder := &writeFailingRecorder{failOn: `{"visibility":"allowed"}`}

	handler := newTestStreamHandler(t, resolver, workspaceID, make(chan time.Time))
	require.NoError(t, handler.serveStream(
		context.Background(),
		recorder,
		newStreamRequest(context.Background()),
		userID,
		workspaceID,
		testWorkspaceSlug,
		context.Background(),
		messages,
	))

	require.Contains(t, recorder.body.String(), `data: {"visibility":"allowed"}`)
}

func newTestStreamHandler(
	t *testing.T,
	resolver interface {
		Resolve(context.Context, *http.Request) (uuid.UUID, bool, error)
	},
	workspaceID uuid.UUID,
	keepAlive <-chan time.Time,
) *Handler {
	t.Helper()
	var logOutput bytes.Buffer
	return &Handler{
		Log:             logger.NewWithJSON(&logOutput, slog.LevelDebug, "sse-http-test"),
		BrowserSessions: resolver,
		WorkspaceAccess: workspaceResolverFunc(func(_ context.Context, slug string, _ uuid.UUID) (mid.WorkspaceInfo, error) {
			return mid.WorkspaceInfo{ID: workspaceID, Slug: slug}, nil
		}),
		keepAlive: keepAlive,
	}
}

func newStreamRequest(ctx context.Context) *http.Request {
	request := httptest.NewRequest(http.MethodGet, "/notifications/subscribe", nil).WithContext(ctx)
	request.AddCookie(&http.Cookie{Name: "fortyone_session", Value: testBrowserSessionToken})
	return request
}

type writeFailingRecorder struct {
	header http.Header
	body   strings.Builder
	failOn string
}

func (recorder *writeFailingRecorder) Header() http.Header {
	if recorder.header == nil {
		recorder.header = make(http.Header)
	}
	return recorder.header
}

func (recorder *writeFailingRecorder) Write(payload []byte) (int, error) {
	written, _ := recorder.body.Write(payload)
	if strings.Contains(string(payload), recorder.failOn) {
		return written, errors.New("test stream complete")
	}
	return written, nil
}

func (*writeFailingRecorder) WriteHeader(int) {}

func (*writeFailingRecorder) Flush() {}
