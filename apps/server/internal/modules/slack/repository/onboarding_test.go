package slackrepository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestHasSlackUserOnboardingReceiptValidatesIdentity(t *testing.T) {
	t.Parallel()
	repo := &Repo{}

	_, err := repo.HasSlackUserOnboardingReceipt(context.Background(), uuid.Nil, "T123", "U123")
	require.ErrorContains(t, err, "workspace is required")

	_, err = repo.HasSlackUserOnboardingReceipt(context.Background(), uuid.New(), " ", "U123")
	require.ErrorContains(t, err, "slack team is required")

	_, err = repo.HasSlackUserOnboardingReceipt(context.Background(), uuid.New(), "T123", " ")
	require.ErrorContains(t, err, "slack user is required")
}
