package usershttp

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/google"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestVerifiedSignInRechecksPolicyForPreviouslyActiveAccount(t *testing.T) {
	t.Parallel()

	repository := &signInReactivationRepository{
		user:            users.CoreUser{ID: uuid.New(), IsActive: true},
		reactivationErr: users.ErrInvalidCredentials,
	}
	log := logger.NewWithJSON(&bytes.Buffer{}, slog.LevelError, "sign-in-test")
	handler := &Handlers{users: users.New(log, repository, nil)}

	_, err := handler.reactivateUserForSignIn(t.Context(), repository.user)

	require.ErrorIs(t, err, users.ErrInvalidCredentials)
	require.Equal(t, repository.user.ID, repository.signIn.UserID)
	require.False(t, repository.signIn.SignedInAt.IsZero())
}

func TestGoogleReturningUserGetsFreshSessionWithoutRevivingOldCookie(t *testing.T) {
	t.Parallel()

	redisServer := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	log := logger.NewWithJSON(&bytes.Buffer{}, slog.LevelError, "sign-in-test")
	cacheService := cache.New(client, log)
	repository := &signInReactivationRepository{
		user: users.CoreUser{
			ID: uuid.New(), Email: "returning@example.com", FullName: "Returning User",
		},
		version: 7,
	}
	handler := &Handlers{
		users: users.New(log, repository, nil), cache: cacheService,
		cookieDomain: ".fortyone.app",
	}
	oldCookie := &http.Cookie{Name: sessionCookieName, Value: "old-session"}
	require.NoError(t, cacheService.Set(t.Context(), cache.AuthSessionCacheKey(oldCookie.Value),
		platformauth.BrowserSession{UserID: repository.user.ID, Version: 6}, time.Hour))
	request := httptest.NewRequest(http.MethodGet, "https://api.fortyone.app/auth/google/callback", nil)
	request.AddCookie(oldCookie)
	recorder := httptest.NewRecorder()

	user, err := handler.authenticateWithGoogleIdentity(t.Context(), recorder, request, google.Identity{
		Email: repository.user.Email, EmailVerified: true,
	})

	require.NoError(t, err)
	require.Equal(t, repository.user.ID, user.ID)
	require.True(t, user.IsActive)
	require.Nil(t, user.LastUsedWorkspaceID)
	newCookie := requireSingleSessionCookie(t, recorder)
	require.NotEqual(t, oldCookie.Value, newCookie.Value)
	require.Equal(t, "fortyone.app", newCookie.Domain)

	resolver := mid.NewBrowserSessionResolver(cacheService, repository)
	_, accepted, err := resolver.Resolve(t.Context(), request)
	require.ErrorIs(t, err, platformauth.ErrInvalidBrowserSession)
	require.False(t, accepted)
	newRequest := httptest.NewRequest(http.MethodGet, "https://api.fortyone.app/auth/me", nil)
	newRequest.AddCookie(newCookie)
	userID, accepted, err := resolver.Resolve(t.Context(), newRequest)
	require.NoError(t, err)
	require.True(t, accepted)
	require.Equal(t, user.ID, userID)
}

type signInReactivationRepository struct {
	users.Repository
	user            users.CoreUser
	version         int64
	signIn          users.VerifiedSignInReactivation
	reactivationErr error
}

func (repository *signInReactivationRepository) GetUserByEmailAnyStatus(context.Context, string) (users.CoreUser, error) {
	return repository.user, nil
}

func (repository *signInReactivationRepository) ReactivateUserForVerifiedSignIn(
	_ context.Context,
	input users.VerifiedSignInReactivation,
) (users.CoreUser, error) {
	repository.signIn = input
	if repository.reactivationErr != nil {
		return users.CoreUser{}, repository.reactivationErr
	}
	if input.UserID != repository.user.ID {
		return users.CoreUser{}, errors.New("unexpected sign-in account")
	}
	repository.user.IsActive = true
	repository.user.LastLoginAt = input.SignedInAt
	return repository.user, nil
}

func (repository *signInReactivationRepository) ResolveActiveBrowserSessionVersion(
	_ context.Context,
	userID uuid.UUID,
) (int64, bool, error) {
	if userID != repository.user.ID {
		return 0, false, nil
	}
	return repository.version, repository.user.IsActive, nil
}
