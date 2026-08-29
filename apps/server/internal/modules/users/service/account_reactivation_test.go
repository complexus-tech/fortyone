package users

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestReactivateUserForVerifiedSignInUsesOneApplicationUTCInstant(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	location := time.FixedZone("CAT", 2*60*60)
	now := time.Date(2026, time.August, 28, 14, 30, 0, 123000000, location)
	repository := &reactivationRepository{
		result: CoreUser{ID: userID, IsActive: true},
	}
	service := New(
		logger.NewWithJSON(&bytes.Buffer{}, slog.LevelError, "users-reactivation-test"),
		repository,
		nil,
		WithClock(fixedUsersClock{now: now}),
	)

	user, err := service.ReactivateUserForVerifiedSignIn(context.Background(), userID)

	require.NoError(t, err)
	require.Equal(t, userID, user.ID)
	require.Equal(t, VerifiedSignInReactivation{
		UserID: userID, SignedInAt: now.UTC(),
	}, repository.input)
}

func TestReactivateUserForVerifiedSignInPreservesGenericAuthenticationFailure(t *testing.T) {
	t.Parallel()

	repository := &reactivationRepository{err: ErrInvalidCredentials}
	service := New(
		logger.NewWithJSON(&bytes.Buffer{}, slog.LevelError, "users-reactivation-test"),
		repository,
		nil,
		WithClock(fixedUsersClock{now: time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)}),
	)

	_, err := service.ReactivateUserForVerifiedSignIn(context.Background(), uuid.New())

	require.ErrorIs(t, err, ErrInvalidCredentials)
}

type reactivationRepository struct {
	Repository
	input  VerifiedSignInReactivation
	result CoreUser
	err    error
}

func (repository *reactivationRepository) ReactivateUserForVerifiedSignIn(
	_ context.Context,
	input VerifiedSignInReactivation,
) (CoreUser, error) {
	repository.input = input
	return repository.result, repository.err
}

type fixedUsersClock struct {
	now time.Time
}

func (clock fixedUsersClock) Now() time.Time {
	return clock.now
}
