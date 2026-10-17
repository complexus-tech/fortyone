package validate

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/require"
)

func webmElement(id []byte, content []byte) []byte {
	return append(append(append([]byte{}, id...), byte(0x80|len(content))), content...)
}

func voiceWebm(trackType byte, duration float64, streaming bool) []byte {
	header := webmElement([]byte{0x1a, 0x45, 0xdf, 0xa3}, webmElement([]byte{0x42, 0x82}, []byte("webm")))
	entry := append(webmElement([]byte{0x83}, []byte{trackType}), webmElement([]byte{0x86}, []byte("A_OPUS"))...)
	tracks := webmElement([]byte{0x16, 0x54, 0xae, 0x6b}, webmElement([]byte{0xae}, entry))
	if duration != 0 {
		value := make([]byte, 8)
		binary.BigEndian.PutUint64(value, math.Float64bits(duration))
		tracks = append(webmElement([]byte{0x15, 0x49, 0xa9, 0x66}, webmElement([]byte{0x44, 0x89}, value)), tracks...)
	}
	cluster := webmElement([]byte{0x1f, 0x43, 0xb6, 0x75}, []byte{0xa3, 0x81, 0x00})
	if streaming {
		cluster[4] = 0xff
	}
	segment := webmElement([]byte{0x18, 0x53, 0x80, 0x67}, append(tracks, cluster...))
	if streaming {
		segment[4] = 0xff
	}
	return append(header, segment...)
}

func voiceMP4Box(kind string, content ...[]byte) []byte {
	data := make([]byte, 8)
	copy(data[4:], kind)
	for _, part := range content {
		data = append(data, part...)
	}
	binary.BigEndian.PutUint32(data, uint32(len(data)))
	return data
}

func voiceMP4(handler string, duration uint32) []byte {
	header := make([]byte, 20)
	binary.BigEndian.PutUint32(header[12:], 1000)
	binary.BigEndian.PutUint32(header[16:], duration)
	return bytes.Join([][]byte{
		voiceMP4Box("ftyp", []byte("M4A "+"\x00\x00\x00\x00")),
		voiceMP4Box("moov", voiceMP4Box("trak", voiceMP4Box("mdia", voiceMP4Box("hdlr", make([]byte, 8), []byte(handler)), voiceMP4Box("mdhd", header)))),
		voiceMP4Box("mdat", []byte{1, 2, 3}),
	}, nil)
}

func TestVoiceNoteUsesAudioTrackStructureAndRestoresCursor(t *testing.T) {
	for _, tc := range []struct {
		name, file, mime string
		data             []byte
	}{
		{"webm", "voice.webm", "audio/webm", voiceWebm(2, 2000, false)},
		{"streaming webm", "voice.webm", "audio/webm", voiceWebm(2, 0, true)},
		{"m4a", "voice.m4a", "audio/mp4", voiceMP4("soun", 2000)},
		{"fragmented mp4 duration", "voice.mp4", "audio/mp4", voiceMP4("soun", 0)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file := &memoryMultipartFile{bytes.NewReader(tc.data)}
			_, err := file.Seek(2, io.SeekStart)
			require.NoError(t, err)
			mime, err := VoiceNote(file, &multipart.FileHeader{Filename: tc.file, Size: int64(len(tc.data)), Header: map[string][]string{"Content-Type": {"text/html"}}}, 2000)
			require.NoError(t, err)
			require.Equal(t, tc.mime, mime)
			position, err := file.Seek(0, io.SeekCurrent)
			require.NoError(t, err)
			require.Zero(t, position)
		})
	}
}

func TestVoiceNoteRejectsSpoofsVideoOversizeAndInvalidDurations(t *testing.T) {
	for _, tc := range []struct {
		name, file     string
		data           []byte
		size, duration int64
		want           error
	}{
		{"webm video", "voice.webm", voiceWebm(1, 0, false), 0, 1000, ErrInvalidFileType},
		{"mp4 video", "voice.m4a", voiceMP4("vide", 1000), 0, 1000, ErrInvalidFileType},
		{"fake MIME", "voice.webm", []byte("<html>not audio</html>"), 0, 1000, ErrInvalidFileType},
		{"mismatched extension", "voice.webm", voiceMP4("soun", 1000), 0, 1000, ErrInvalidFileType},
		{"unsupported extension", "voice.ogg", voiceWebm(2, 0, false), 0, 1000, ErrInvalidFileType},
		{"too large", "voice.webm", voiceWebm(2, 0, false), MaxVoiceNoteSize + 1, 1000, ErrFileTooLarge},
		{"empty duration", "voice.webm", voiceWebm(2, 0, false), 0, 0, ErrVoiceNoteDuration},
		{"excess declared duration", "voice.webm", voiceWebm(2, 0, false), 0, 300001, ErrVoiceNoteDuration},
		{"excess webm duration", "voice.webm", voiceWebm(2, 400000, false), 0, 1000, ErrVoiceNoteDuration},
		{"excess mp4 duration", "voice.m4a", voiceMP4("soun", 400000), 0, 1000, ErrVoiceNoteDuration},
		{"invalid container duration", "voice.webm", voiceWebm(2, math.Inf(1), false), 0, 1000, ErrInvalidFileType},
		{"truncated mp4", "voice.m4a", voiceMP4("soun", 2000)[:20], 0, 1000, ErrInvalidFileType},
		{"empty file", "voice.webm", nil, 0, 1000, ErrEmptyFile},
	} {
		t.Run(tc.name, func(t *testing.T) {
			size := tc.size
			if size == 0 {
				size = int64(len(tc.data))
			}
			_, err := VoiceNote(&memoryMultipartFile{bytes.NewReader(tc.data)}, &multipart.FileHeader{Filename: tc.file, Size: size}, tc.duration)
			require.ErrorIs(t, err, tc.want)
		})
	}
}

func TestVoiceNoteContainerParsingIsBounded(t *testing.T) {
	data := bytes.Repeat(voiceMP4Box("free"), 4097)
	_, err := VoiceNote(&memoryMultipartFile{bytes.NewReader(data)}, &multipart.FileHeader{Filename: "voice.m4a", Size: int64(len(data))}, 1000)
	require.ErrorIs(t, err, ErrInvalidFileType)
	for _, size := range []int{1, 2, 3, 4, 5, 8, 12} {
		data := bytes.Repeat([]byte{0xff}, size)
		_, err := VoiceNote(&memoryMultipartFile{bytes.NewReader(data)}, &multipart.FileHeader{Filename: "voice.webm", Size: int64(size)}, 1000)
		require.Error(t, err)
	}
}

func FuzzVoiceNoteContainerValidation(f *testing.F) {
	f.Add(voiceWebm(2, 1000, false), true)
	f.Add(voiceWebm(2, 0, true), true)
	f.Add(voiceMP4("soun", 1000), false)
	f.Add([]byte{0xff}, true)
	f.Fuzz(func(t *testing.T, data []byte, webm bool) {
		if len(data) > 1<<20 {
			t.Skip("container fuzz input exceeds the metadata-focused test budget")
		}
		filename := "voice.m4a"
		if webm {
			filename = "voice.webm"
		}
		file := &memoryMultipartFile{bytes.NewReader(data)}
		mime, err := VoiceNote(file, &multipart.FileHeader{Filename: filename, Size: int64(len(data))}, 1000)
		if err == nil {
			require.Contains(t, []string{"audio/webm", "audio/mp4"}, mime)
		}
		position, seekErr := file.Seek(0, io.SeekCurrent)
		require.NoError(t, seekErr)
		require.Zero(t, position, "validation must leave the stored upload intact")
	})
}

func TestChatAudioDoesNotRequireRecordingMetadata(t *testing.T) {
	for _, tc := range []struct {
		name, filename, mime string
		data                 []byte
	}{
		{"webm", "interview.webm", "audio/webm", voiceWebm(2, 600000, false)},
		{"m4a", "interview.m4a", "audio/mp4", voiceMP4("soun", 600000)},
		{"mp4", "interview.mp4", "audio/mp4", voiceMP4("soun", 600000)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file := &memoryMultipartFile{bytes.NewReader(tc.data)}
			header := &multipart.FileHeader{Filename: tc.filename, Size: int64(len(tc.data))}
			mime, err := ChatAudio(file, header)
			require.NoError(t, err)
			require.Equal(t, tc.mime, mime)
			_, err = VoiceNote(file, header, 1000)
			require.ErrorIs(t, err, ErrVoiceNoteDuration, "the recording duration limit remains enforced")
		})
	}
}

func TestChatAudioRejectsVideoSpoofsAndOversizedAudio(t *testing.T) {
	oversized := voiceMP4("soun", 1000)
	padding := make([]byte, MaxVoiceNoteSize-len(oversized))
	oversized = append(oversized, voiceMP4Box("free", padding)...)
	for _, tc := range []struct {
		name, filename string
		data           []byte
		want           error
	}{
		{"oversized audio mp4", "audio.mp4", oversized, ErrFileTooLarge},
		{"video mp4", "video.mp4", voiceMP4("vide", 1000), ErrInvalidFileType},
		{"video webm", "video.webm", voiceWebm(1, 1000, false), ErrInvalidFileType},
		{"spoofed m4a", "audio.m4a", []byte("not an audio container"), ErrInvalidFileType},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ChatAudio(&memoryMultipartFile{bytes.NewReader(tc.data)}, &multipart.FileHeader{Filename: tc.filename, Size: int64(len(tc.data))})
			require.ErrorIs(t, err, tc.want)
		})
	}
}
