package chatservice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"
	"strings"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/complexus-tech/projects-api/pkg/validate"
)

const DefaultTranscriptionModel = "gpt-4o-mini-transcribe"

type transcriptionHTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type OpenAITranscriber struct {
	apiKey, model, endpoint string
	client                  transcriptionHTTPClient
}

func NewOpenAITranscriber(apiKey, model string) *OpenAITranscriber {
	if strings.TrimSpace(model) == "" {
		model = DefaultTranscriptionModel
	}
	return &OpenAITranscriber{
		apiKey: strings.TrimSpace(apiKey), model: model, endpoint: "https://api.openai.com/v1/audio/transcriptions",
		client: &http.Client{Timeout: 2 * time.Minute, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }},
	}
}

func (p *OpenAITranscriber) Transcribe(ctx context.Context, audio domain.TranscriptionAudio, data []byte) (string, string, error) {
	if p.apiKey == "" {
		return "", "", domain.ErrTranscriptionUnavailable
	}
	if len(data) == 0 || len(data) > validate.MaxVoiceNoteSize || (audio.MediaType != "audio/webm" && audio.MediaType != "audio/mp4") {
		return "", "", domain.ErrInvalid
	}
	var payload bytes.Buffer
	writer := multipart.NewWriter(&payload)
	if err := writer.WriteField("model", p.model); err != nil {
		return "", "", err
	}
	if err := writer.WriteField("response_format", "json"); err != nil {
		return "", "", err
	}
	filename := filepath.Base(audio.FileName)
	// The file extension is required by the transcription API. The source MIME
	// was established by our recording validator, not a browser-supplied header.
	extension := ".webm"
	if audio.MediaType == "audio/mp4" {
		extension = ".m4a"
	}
	filename = strings.TrimSuffix(filename, filepath.Ext(filename)) + extension
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{"name": "file", "filename": filename}))
	header.Set("Content-Type", audio.MediaType)
	part, err := writer.CreatePart(header)
	if err != nil {
		return "", "", err
	}
	if _, err := part.Write(data); err != nil {
		return "", "", err
	}
	if err := writer.Close(); err != nil {
		return "", "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, &payload)
	if err != nil {
		return "", "", err
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+p.apiKey)
	response, err := p.client.Do(request)
	if err != nil {
		return "", "", fmt.Errorf("voice transcription request failed: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", "", fmt.Errorf("voice transcription provider returned status %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, (1<<20)+1))
	if err != nil || len(body) > 1<<20 {
		return "", "", errors.New("invalid voice transcription response size")
	}
	var result struct {
		Text *string `json:"text"`
	}
	if err := json.Unmarshal(body, &result); err != nil || result.Text == nil {
		return "", "", errors.New("invalid voice transcription response")
	}
	return strings.TrimSpace(*result.Text), p.model, nil
}
