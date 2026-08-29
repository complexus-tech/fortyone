package oauthstate

import (
	"bytes"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTokenRoundTripIsCanonicalAndDigestOnly(t *testing.T) {
	t.Parallel()

	source := bytes.NewReader(bytes.Repeat([]byte{0x5a}, TokenSize))
	token, err := New(source)
	require.NoError(t, err)
	require.Len(t, token.String(), 43)

	parsed, err := Parse(token.String())
	require.NoError(t, err)
	require.Equal(t, token.String(), parsed.String())
	require.Equal(t, token.Digest(), parsed.Digest())
	require.NotEqual(t, []byte(token.String()), token.Digest())
}

func TestParseRejectsMalformedWrongLengthAndNonCanonicalTokens(t *testing.T) {
	t.Parallel()

	valid := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{1}, TokenSize))
	for _, candidate := range []string{
		"",
		"not-base64!",
		base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{1}, TokenSize-1)),
		valid + "=",
	} {
		_, err := Parse(candidate)
		require.ErrorIs(t, err, ErrInvalidToken, "candidate %q", candidate)
	}
}

func TestNewFailsClosedWithoutRandomSource(t *testing.T) {
	t.Parallel()

	_, err := New(nil)
	require.ErrorIs(t, err, ErrRandomSource)

	_, err = New(errorReader{})
	require.Error(t, err)
	require.False(t, errors.Is(err, ErrInvalidToken))
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, errors.New("entropy unavailable")
}
