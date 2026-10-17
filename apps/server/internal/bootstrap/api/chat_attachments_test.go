package api

import (
	"testing"

	chat "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/stretchr/testify/require"
)

func TestChatAttachmentMetadataRequiresAnExplicitValidRecording(t *testing.T) {
	empty, err := parseChatAttachmentMetadata("", "")
	require.NoError(t, err)
	require.Empty(t, empty.Kind)
	require.Nil(t, empty.DurationMS)
	voice, err := parseChatAttachmentMetadata("voice-note", "300000")
	require.NoError(t, err)
	require.Equal(t, "voice-note", voice.Kind)
	require.EqualValues(t, 300000, *voice.DurationMS)
	for _, input := range [][2]string{{"", "5000"}, {"video", "5000"}, {"voice-note", ""}, {"voice-note", "0"}, {"voice-note", "-1"}, {"voice-note", "300001"}, {"voice-note", "1.5"}, {"voice-note", "9999999999999999999999"}} {
		_, err := parseChatAttachmentMetadata(input[0], input[1])
		require.ErrorIs(t, err, chat.ErrInvalid)
	}
}
