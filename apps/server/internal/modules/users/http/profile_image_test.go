package usershttp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUploadProfileImageRejectsOversizedMultipartRequest(t *testing.T) {
	t.Parallel()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("image", "oversized.png")
	require.NoError(t, err)
	_, err = part.Write(bytes.Repeat([]byte{'x'}, validate.MaxProfileImageSize+(2<<20)))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	request := httptest.NewRequest(http.MethodPost, "/users/profile/image", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()
	handler := &Handlers{
		log: logger.NewWithText(io.Discard, slog.LevelError, "profile-image-test"),
	}
	ctx := platformauth.SetUserID(context.Background(), uuid.New())

	require.NoError(t, handler.UploadProfileImage(ctx, recorder, request))
	require.Equal(t, http.StatusRequestEntityTooLarge, recorder.Code)

	var response web.Response
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.NotNil(t, response.Error)
	require.Equal(t, "request_too_large", response.Error.Code)
}
