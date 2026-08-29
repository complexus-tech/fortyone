package web

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseMultipartFormEnforcesWholeRequestLimit(t *testing.T) {
	t.Parallel()

	request := multipartRequest(t, bytes.Repeat([]byte("a"), 1024))
	response := httptest.NewRecorder()

	err := ParseMultipartForm(response, request, 128)

	require.ErrorIs(t, err, ErrRequestBodyTooLarge)
	require.NoError(t, RemoveMultipartForm(request))
}

func TestParseMultipartFormParsesAndCleansUp(t *testing.T) {
	t.Parallel()

	request := multipartRequest(t, []byte("image payload"))
	response := httptest.NewRecorder()

	err := ParseMultipartForm(response, request, 4096)
	require.NoError(t, err)
	file, header, err := request.FormFile("file")
	require.NoError(t, err)
	require.Equal(t, "image.png", header.Filename)
	require.NoError(t, file.Close())
	require.NoError(t, RemoveMultipartForm(request))
}

func TestParseMultipartFormRejectsInvalidLimits(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewReader(nil))
	_, err := request.MultipartReader()
	require.Error(t, err)
	require.Error(t, ParseMultipartForm(httptest.NewRecorder(), request, 0))
	require.Error(t, ParseMultipartForm(nil, request, 1))
	require.NoError(t, RemoveMultipartForm(nil))
}

func multipartRequest(t *testing.T, content []byte) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "image.png")
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	request := httptest.NewRequest(http.MethodPost, "/upload", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}

func TestRemoveMultipartFormReturnsNilWithoutParsedForm(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewReader(nil))
	require.NoError(t, RemoveMultipartForm(request))
}
