package figma

import (
	"encoding/json"
	"testing"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/stretchr/testify/require"
)

func TestParseURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		url     string
		fileKey string
		nodeID  *string
		wantErr bool
	}{
		{name: "frame", url: "https://www.figma.com/design/AbCdEf/Product?node-id=43-2&t=ignored", fileKey: "AbCdEf", nodeID: pointer("43:2")},
		{name: "file", url: "https://figma.com/file/AbCdEf/Product", fileKey: "AbCdEf"},
		{name: "figjam board", url: "https://www.figma.com/board/BoardKey/Workshop?node-id=1-2", fileKey: "BoardKey", nodeID: pointer("1:2")},
		{name: "lookalike host", url: "https://figma.com.example.org/design/key/file", wantErr: true},
		{name: "unsupported route", url: "https://www.figma.com/community/file/123", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			artifact, err := ParseURL(test.url)
			if test.wantErr {
				require.ErrorIs(t, err, ErrInvalidFigmaURL)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.fileKey, artifact.FileKey)
			require.Equal(t, test.nodeID, artifact.NodeID)
			require.NotContains(t, artifact.CanonicalURL, "t=")
		})
	}
}

func TestCredentialRoundTrip(t *testing.T) {
	t.Parallel()
	token := Token{AccessToken: "access", RefreshToken: "refresh", TokenType: "bearer"}
	sealed, err := encryptToken("secret", token)
	require.NoError(t, err)
	require.NotContains(t, sealed, token.AccessToken)
	opened, err := decryptToken("secret", sealed)
	require.NoError(t, err)
	require.Equal(t, token, opened)
}

func TestStoryURL(t *testing.T) {
	t.Parallel()
	story := stories.CoreSingleStory{TeamCode: "prd", SequenceID: 42}
	require.Equal(t, "https://acme.fortyone.app/work/PRD-42", (&Service{config: Config{WebsiteURL: "https://fortyone.app"}}).storyURL("acme", story))
	require.Equal(t, "http://localhost:3000/acme/work/PRD-42", (&Service{config: Config{WebsiteURL: "http://localhost:3000"}}).storyURL("acme", story))
}

func TestWebhookIDAcceptsStringAndNumber(t *testing.T) {
	t.Parallel()
	for _, payload := range []string{`{"webhook_id":"434"}`, `{"webhook_id":434}`} {
		var event WebhookEvent
		require.NoError(t, json.Unmarshal([]byte(payload), &event))
		require.Equal(t, int64(434), int64(event.WebhookID))
	}
}

func pointer(value string) *string { return &value }
