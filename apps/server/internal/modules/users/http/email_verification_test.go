package usershttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func TestSendEmailVerificationReturnsTooManyRequestsWhenTokenCreationIsRateLimited(t *testing.T) {
	handler := newEmailVerificationTestHandler(&emailVerificationRateLimitRepo{})
	body := bytes.NewBufferString(`{"email":"existing-user@example.com","isMobile":false}`)
	request := httptest.NewRequest(http.MethodPost, "/users/verify/email", body)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	if err := handler.SendEmailVerification(context.Background(), recorder, request); err != nil {
		t.Fatalf("send email verification returned error: %v", err)
	}

	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusTooManyRequests, recorder.Code, recorder.Body.String())
	}

	var response web.Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Error == nil || response.Error.Message != users.ErrTooManyAttempts.Error() {
		t.Fatalf("expected too many attempts response, got %#v", response.Error)
	}
}

func newEmailVerificationTestHandler(repo users.Repository) *Handlers {
	log := logger.NewWithText(io.Discard, slog.LevelError, "users-http-test")
	verificationRepo, ok := repo.(users.VerificationTokenRepository)
	if !ok {
		panic("test repository must implement users.VerificationTokenRepository")
	}
	manager, err := users.NewVerificationTokenManager(users.VerificationTokenConfig{
		Current: users.VerificationTokenKey{
			ID:     "test-v1",
			Secret: "test-verification-token-key-with-32-bytes",
		},
	})
	if err != nil {
		panic(err)
	}
	handler := New(
		users.New(log, repo, nil, users.WithVerificationTokens(manager, verificationRepo)),
		nil,
		"test-secret",
		"",
		"http://localhost:3000",
		nil,
		log,
		"",
		nil,
		nil,
		nil,
	)
	handler.verificationRateLimits = emailVerificationRateLimitStore{}
	return handler
}

type emailVerificationRateLimitRepo struct{}

func (emailVerificationRateLimitRepo) ResolveActiveBrowserSessionVersion(context.Context, uuid.UUID) (int64, bool, error) {
	return 1, true, nil
}

func (emailVerificationRateLimitRepo) GetUser(ctx context.Context, userID uuid.UUID) (users.CoreUser, error) {
	return users.CoreUser{}, users.ErrNotFound
}

func (emailVerificationRateLimitRepo) GetUserByEmail(ctx context.Context, email string) (users.CoreUser, error) {
	return users.CoreUser{}, users.ErrNotFound
}

func (emailVerificationRateLimitRepo) GetUserByEmailAnyStatus(ctx context.Context, email string) (users.CoreUser, error) {
	return users.CoreUser{}, errors.New("existing-user lookup should not block token creation")
}

func (emailVerificationRateLimitRepo) GetUsersByIDs(ctx context.Context, userIDs []uuid.UUID) ([]users.CoreUser, error) {
	return nil, nil
}

func (emailVerificationRateLimitRepo) UpdateUser(ctx context.Context, userID uuid.UUID, updates users.CoreUpdateUser) (users.CoreUser, error) {
	return users.CoreUser{}, nil
}

func (emailVerificationRateLimitRepo) ReactivateUserForVerifiedSignIn(ctx context.Context, input users.VerifiedSignInReactivation) (users.CoreUser, error) {
	return users.CoreUser{}, nil
}

func (emailVerificationRateLimitRepo) DeleteUser(ctx context.Context, userID uuid.UUID, deactivatedAt time.Time) error {
	return nil
}

func (emailVerificationRateLimitRepo) UpdateUserWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) error {
	return nil
}

func (emailVerificationRateLimitRepo) List(ctx context.Context, workspaceID uuid.UUID, filter users.CoreListUsersFilter) ([]users.CoreUser, error) {
	return nil, nil
}

func (emailVerificationRateLimitRepo) Create(ctx context.Context, user users.CoreUser) (users.CoreUser, error) {
	return users.CoreUser{}, nil
}

func (emailVerificationRateLimitRepo) CreateVerificationToken(ctx context.Context, token users.NewVerificationToken) (users.CoreVerificationToken, error) {
	return users.CoreVerificationToken{}, users.ErrTooManyAttempts
}

func (emailVerificationRateLimitRepo) ConsumeVerificationToken(ctx context.Context, input users.ConsumeVerificationTokenInput) (users.CoreVerificationToken, error) {
	return users.CoreVerificationToken{}, users.ErrInvalidToken
}

func (emailVerificationRateLimitRepo) InvalidateTokens(ctx context.Context, email string) error {
	return nil
}

func (emailVerificationRateLimitRepo) GetAutomationPreferences(ctx context.Context, userID, workspaceID uuid.UUID) (users.CoreAutomationPreferences, error) {
	return users.CoreAutomationPreferences{}, nil
}

func (emailVerificationRateLimitRepo) UpdateAutomationPreferences(ctx context.Context, userID, workspaceID uuid.UUID, updates users.CoreUpdateAutomationPreferences) error {
	return nil
}

func (emailVerificationRateLimitRepo) GetOnboardingTourProgress(
	ctx context.Context,
	userID, workspaceID uuid.UUID,
	scope users.CoreOnboardingTourScope,
) (users.CoreOnboardingTourProgress, error) {
	return users.CoreOnboardingTourProgress{}, nil
}

func (emailVerificationRateLimitRepo) UpdateOnboardingTourProgress(
	ctx context.Context,
	userID, workspaceID uuid.UUID,
	updates users.CoreUpdateOnboardingTourProgress,
) (users.CoreOnboardingTourProgress, error) {
	return users.CoreOnboardingTourProgress{}, nil
}

func (emailVerificationRateLimitRepo) AddUserMemory(ctx context.Context, memory users.NewUserMemoryItem) (users.CoreUserMemoryItem, error) {
	return users.CoreUserMemoryItem{}, nil
}

func (emailVerificationRateLimitRepo) UpdateUserMemory(ctx context.Context, id uuid.UUID, scope users.UserMemoryScope, update users.UpdateUserMemoryItem) error {
	return nil
}

func (emailVerificationRateLimitRepo) DeleteUserMemory(ctx context.Context, id uuid.UUID, scope users.UserMemoryScope) error {
	return nil
}

func (emailVerificationRateLimitRepo) ListUserMemories(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID) ([]users.CoreUserMemoryItem, error) {
	return nil, nil
}

type emailVerificationRateLimitStore struct{}

func (emailVerificationRateLimitStore) IncrementWithTTL(context.Context, string, time.Duration) (int64, error) {
	return 1, nil
}

func TestVerifyEmailReturnsOneEnumerationSafeTokenError(t *testing.T) {
	t.Parallel()

	errorsToTest := []error{
		users.ErrInvalidToken,
		users.ErrTokenExpired,
		users.ErrTokenUsed,
	}
	var responseBody string
	for _, tokenErr := range errorsToTest {
		repo := &verificationConsumeRepo{consumeErr: tokenErr}
		handler := newEmailVerificationTestHandler(repo)
		request := httptest.NewRequest(
			http.MethodPost,
			"/users/verify/email/confirm",
			bytes.NewBufferString(`{"email":"user@example.com","token":"123456"}`),
		)
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		if err := handler.VerifyEmail(context.Background(), recorder, request); err != nil {
			t.Fatalf("verify email returned error: %v", err)
		}
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
		}
		if responseBody == "" {
			responseBody = recorder.Body.String()
		} else if recorder.Body.String() != responseBody {
			t.Fatalf("token state changed public response: %s != %s", recorder.Body.String(), responseBody)
		}
		if !strings.Contains(recorder.Body.String(), users.ErrInvalidToken.Error()) {
			t.Fatalf("response = %s, want generic invalid-token message", recorder.Body.String())
		}
	}
}

func TestVerifyEmailRateLimitStopsBeforeTokenConsumption(t *testing.T) {
	t.Parallel()

	repo := &verificationConsumeRepo{}
	handler := newEmailVerificationTestHandler(repo)
	handler.verificationRateLimits = &recordingVerificationRateLimitStore{
		counts: []int64{verificationConfirmEmailLimit + 1},
	}
	request := httptest.NewRequest(
		http.MethodPost,
		"/users/verify/email/confirm",
		bytes.NewBufferString(`{"email":"user@example.com","token":"123456"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	if err := handler.VerifyEmail(context.Background(), recorder, request); err != nil {
		t.Fatalf("verify email returned error: %v", err)
	}
	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusTooManyRequests)
	}
	if recorder.Header().Get("Retry-After") == "" {
		t.Fatal("rate-limited response omitted Retry-After")
	}
	if recorder.Header().Get("RateLimit-Policy") != `"verification-confirm-email";q=10;w=600` {
		t.Fatalf("RateLimit-Policy = %q", recorder.Header().Get("RateLimit-Policy"))
	}
	if recorder.Header().Get("RateLimit") != `"verification-confirm-email";r=0;t=600` {
		t.Fatalf("RateLimit = %q", recorder.Header().Get("RateLimit"))
	}
	if repo.consumeCalled {
		t.Fatal("rate-limited attempt reached token storage")
	}
}

func TestVerifyEmailReturnsGenericUnauthorizedForAdminBlockedAccount(t *testing.T) {
	t.Parallel()

	repo := &blockedSignInVerificationRepo{userID: uuid.New()}
	handler := newEmailVerificationTestHandler(repo)
	request := httptest.NewRequest(
		http.MethodPost,
		"/users/verify/email/confirm",
		bytes.NewBufferString(`{"email":"user@example.com","token":"123456"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	if err := handler.VerifyEmail(context.Background(), recorder, request); err != nil {
		t.Fatalf("verify email returned error: %v", err)
	}
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusUnauthorized, recorder.Body.String())
	}
	var response web.Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Error == nil || response.Error.Message != users.ErrInvalidCredentials.Error() {
		t.Fatalf("response = %#v, want generic invalid credentials", response.Error)
	}
	for _, sensitiveState := range []string{"admin", "disabled", "reactivation", "policy"} {
		if strings.Contains(strings.ToLower(recorder.Body.String()), sensitiveState) {
			t.Fatalf("response leaked account state %q: %s", sensitiveState, recorder.Body.String())
		}
	}
	if len(recorder.Result().Cookies()) != 0 {
		t.Fatal("blocked sign-in issued a browser session cookie")
	}
	if !repo.reactivationCalled {
		t.Fatal("verified sign-in did not reach the durable reactivation gate")
	}
}

func TestVerificationRateLimitCacheKeysAreOpaque(t *testing.T) {
	t.Parallel()

	handler := newEmailVerificationTestHandler(&emailVerificationRateLimitRepo{})
	store := &recordingVerificationRateLimitStore{}
	handler.verificationRateLimits = store
	recorder := httptest.NewRecorder()

	allowed, err := handler.enforceVerificationRateLimits(
		context.Background(),
		recorder,
		"confirm",
		verificationRateLimitIdentity{kind: "email", value: "private@example.com", limit: 10},
		verificationRateLimitIdentity{kind: "token", value: "123456", limit: 10},
		verificationRateLimitIdentity{kind: "network", value: "203.0.113.42", limit: 10},
	)
	if err != nil || !allowed {
		t.Fatalf("enforce verification rate limits = (%v, %v), want allowed", allowed, err)
	}
	if got := recorder.Header().Values("RateLimit-Policy"); len(got) != 3 {
		t.Fatalf("RateLimit-Policy values = %v, want three independent policies", got)
	}

	for _, key := range store.Keys() {
		for _, secretValue := range []string{"private@example.com", "123456", "203.0.113.42"} {
			if strings.Contains(key, secretValue) {
				t.Fatalf("cache key %q leaked %q", key, secretValue)
			}
		}
	}
}

type verificationConsumeRepo struct {
	emailVerificationRateLimitRepo
	consumeErr    error
	consumeCalled bool
}

type blockedSignInVerificationRepo struct {
	emailVerificationRateLimitRepo
	userID             uuid.UUID
	reactivationCalled bool
}

func (repository *blockedSignInVerificationRepo) ConsumeVerificationToken(
	context.Context,
	users.ConsumeVerificationTokenInput,
) (users.CoreVerificationToken, error) {
	return users.CoreVerificationToken{}, nil
}

func (repository *blockedSignInVerificationRepo) GetUserByEmailAnyStatus(
	context.Context,
	string,
) (users.CoreUser, error) {
	return users.CoreUser{ID: repository.userID, Email: "user@example.com", IsActive: false}, nil
}

func (repository *blockedSignInVerificationRepo) ReactivateUserForVerifiedSignIn(
	_ context.Context,
	input users.VerifiedSignInReactivation,
) (users.CoreUser, error) {
	repository.reactivationCalled = true
	if input.UserID != repository.userID {
		return users.CoreUser{}, errors.New("unexpected account")
	}
	return users.CoreUser{}, users.ErrInvalidCredentials
}

func (r *verificationConsumeRepo) ConsumeVerificationToken(context.Context, users.ConsumeVerificationTokenInput) (users.CoreVerificationToken, error) {
	r.consumeCalled = true
	if r.consumeErr != nil {
		return users.CoreVerificationToken{}, r.consumeErr
	}
	return users.CoreVerificationToken{}, users.ErrInvalidToken
}

type recordingVerificationRateLimitStore struct {
	mu     sync.Mutex
	keys   []string
	counts []int64
}

func (s *recordingVerificationRateLimitStore) IncrementWithTTL(_ context.Context, key string, _ time.Duration) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.keys = append(s.keys, key)
	if len(s.counts) == 0 {
		return 1, nil
	}
	count := s.counts[0]
	s.counts = s.counts[1:]
	return count, nil
}

func (s *recordingVerificationRateLimitStore) Keys() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.keys...)
}
