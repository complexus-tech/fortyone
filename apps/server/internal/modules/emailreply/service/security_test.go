package emailreply

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWebhookTokenIsDeterministicAndDomainSeparated(t *testing.T) {
	t.Parallel()

	first, err := DeriveWebhookToken("application-secret")
	require.NoError(t, err)
	second, err := DeriveWebhookToken("application-secret")
	require.NoError(t, err)
	other, err := DeriveWebhookToken("different-secret")
	require.NoError(t, err)

	require.Equal(t, first, second)
	require.NotEqual(t, first, other)
	require.NotEqual(t, "application-secret", first)
	require.Len(t, first, 43)
	require.True(t, VerifyWebhookToken("application-secret", first))
	require.False(t, VerifyWebhookToken("application-secret", other))
	require.False(t, VerifyWebhookToken("", first))
}

func TestPayloadCodecRoundTrip(t *testing.T) {
	t.Parallel()

	codec, err := NewPayloadCodec("application-secret")
	require.NoError(t, err)

	sealed, err := codec.Seal([]byte(`{"subject":"Quarterly plan"}`))
	require.NoError(t, err)
	require.NotContains(t, sealed, "Quarterly plan")

	opened, err := codec.Open(sealed)
	require.NoError(t, err)
	require.JSONEq(t, `{"subject":"Quarterly plan"}`, string(opened))
}

func TestPayloadCodecRejectsOtherPurposeKeyAndPlaintext(t *testing.T) {
	t.Parallel()

	codec, err := NewPayloadCodec("application-secret")
	require.NoError(t, err)
	other, err := NewPayloadCodec("different-secret")
	require.NoError(t, err)

	sealed, err := codec.Seal([]byte("reply"))
	require.NoError(t, err)
	_, err = other.Open(sealed)
	require.Error(t, err)
	_, err = codec.Open("reply")
	require.ErrorContains(t, err, "unencrypted")
}
