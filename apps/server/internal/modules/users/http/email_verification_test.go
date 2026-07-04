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
	"testing"
	"time"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func TestSendEmailVerificationReturnsTooManyRequestsWhenTokenCreationIsRateLimited(t *testing.T) {
	handler := newEmailVerificationTestHandler(&emailVerificationRateLimitRepo{})
	body := bytes.NewBufferString(`{"email":"existing-user@example.com","isMobile":false}`)
	request := httptest.NewRequest(http.MethodPost, "/users/verify/email", body)
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
	return New(users.New(log, repo, nil), nil, "test-secret", "", nil, nil, nil)
}

type emailVerificationRateLimitRepo struct{}

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

func (emailVerificationRateLimitRepo) ActivateUser(ctx context.Context, userID uuid.UUID) (users.CoreUser, error) {
	return users.CoreUser{}, nil
}

func (emailVerificationRateLimitRepo) DeleteUser(ctx context.Context, userID uuid.UUID) error {
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

func (emailVerificationRateLimitRepo) CreateVerificationToken(ctx context.Context, email, tokenType string, expiresAt time.Time) (users.CoreVerificationToken, error) {
	return users.CoreVerificationToken{}, users.ErrTooManyAttempts
}

func (emailVerificationRateLimitRepo) GetVerificationToken(ctx context.Context, token string) (users.CoreVerificationToken, error) {
	return users.CoreVerificationToken{}, users.ErrInvalidToken
}

func (emailVerificationRateLimitRepo) MarkTokenUsed(ctx context.Context, tokenID uuid.UUID) error {
	return nil
}

func (emailVerificationRateLimitRepo) InvalidateTokens(ctx context.Context, email string) error {
	return nil
}

func (emailVerificationRateLimitRepo) GetValidTokenCount(ctx context.Context, email string, duration time.Duration) (int, error) {
	return 0, nil
}

func (emailVerificationRateLimitRepo) UpdateUserWorkspaceWithTx(ctx context.Context, tx *sqlx.Tx, userID, workspaceID uuid.UUID) error {
	return nil
}

func (emailVerificationRateLimitRepo) GetAutomationPreferences(ctx context.Context, userID, workspaceID uuid.UUID) (users.CoreAutomationPreferences, error) {
	return users.CoreAutomationPreferences{}, nil
}

func (emailVerificationRateLimitRepo) UpdateAutomationPreferences(ctx context.Context, userID, workspaceID uuid.UUID, updates users.CoreUpdateAutomationPreferences) error {
	return nil
}

func (emailVerificationRateLimitRepo) AddUserMemory(ctx context.Context, memory users.NewUserMemoryItem) (users.CoreUserMemoryItem, error) {
	return users.CoreUserMemoryItem{}, nil
}

func (emailVerificationRateLimitRepo) UpdateUserMemory(ctx context.Context, id uuid.UUID, update users.UpdateUserMemoryItem) error {
	return nil
}

func (emailVerificationRateLimitRepo) DeleteUserMemory(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (emailVerificationRateLimitRepo) ListUserMemories(ctx context.Context, userID uuid.UUID, workspaceID uuid.UUID) ([]users.CoreUserMemoryItem, error) {
	return nil, nil
}
