package feedbacksecurity

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDeriveUnsubscribeTokenIsStableScopedAndHashOnly(t *testing.T) {
	deliveryID := uuid.MustParse("9f887dcc-fc4a-4f50-9eca-b50c0424af37")
	token, hash, err := DeriveUnsubscribeToken("auth-secret", deliveryID)
	require.NoError(t, err)
	require.Len(t, token, 43)
	require.Len(t, hash, 32)

	again, againHash, err := DeriveUnsubscribeToken("auth-secret", deliveryID)
	require.NoError(t, err)
	require.Equal(t, token, again)
	require.Equal(t, hash, againHash)

	otherDelivery, _, err := DeriveUnsubscribeToken("auth-secret", uuid.New())
	require.NoError(t, err)
	require.NotEqual(t, token, otherDelivery)
	otherSecret, _, err := DeriveUnsubscribeToken("different-secret", deliveryID)
	require.NoError(t, err)
	require.NotEqual(t, token, otherSecret)
}

func TestDeriveUnsubscribeTokenRejectsMissingKeyMaterial(t *testing.T) {
	_, _, err := DeriveUnsubscribeToken("", uuid.New())
	require.Error(t, err)
	_, _, err = DeriveUnsubscribeToken("secret", uuid.Nil)
	require.Error(t, err)
}
