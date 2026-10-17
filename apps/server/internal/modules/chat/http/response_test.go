package chathttp

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestUnexpectedResponseLogsClassificationWithoutDatabasePayload(t *testing.T) {
	var output bytes.Buffer
	h := handlers{log: logger.NewWithText(&output, slog.LevelError, "test")}
	err := fmt.Errorf("private message: %w", &pgconn.PgError{Code: "40001", Message: "private SQL", Detail: "private body"})
	w := httptest.NewRecorder()
	ctx := logger.WithRequestID(context.Background(), "test-correlation")
	require.NoError(t, h.respond(ctx, w, nil, err, http.StatusCreated))
	require.Equal(t, http.StatusInternalServerError, w.Code)
	require.Contains(t, output.String(), "40001")
	require.Contains(t, output.String(), "test-correlation")
	require.NotContains(t, output.String(), "private")
	require.NotContains(t, w.Body.String(), "40001")
	require.NotContains(t, w.Body.String(), "private")
}

func TestExpectedDomainResponseDoesNotLogServerFailure(t *testing.T) {
	var output bytes.Buffer
	h := handlers{log: logger.NewWithText(&output, slog.LevelError, "test")}
	w := httptest.NewRecorder()
	require.NoError(t, h.respond(context.Background(), w, nil, domain.ErrForbidden, http.StatusOK))
	require.Equal(t, http.StatusForbidden, w.Code)
	require.Empty(t, output.String())
}
