package documentsrepository

import (
	"context"
	"errors"
	"strings"
	"testing"

	documentdomain "github.com/complexus-tech/projects-api/internal/modules/documents/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestRepositoryFailsClosedWhenNativePoolIsMissing(t *testing.T) {
	t.Parallel()

	repository := New(nil)
	_, err := repository.Get(context.Background(), uuid.New(), uuid.New(), uuid.New())
	if !errors.Is(err, documentdomain.ErrNotConfigured) {
		t.Fatalf("Get() error = %v, want ErrNotConfigured", err)
	}
}

func TestMapNotFoundDoesNotLeakAdapterError(t *testing.T) {
	t.Parallel()

	if err := mapNotFound("hidden document", pgx.ErrNoRows); !errors.Is(err, documentdomain.ErrNotFound) {
		t.Fatalf("mapNotFound() error = %v, want ErrNotFound", err)
	}
	want := errors.New("database unavailable")
	if err := mapNotFound("read document", want); !errors.Is(err, want) {
		t.Fatalf("mapNotFound() error = %v, want wrapped cause", err)
	}
}

func TestDuplicateTitleUsesPostgreSQLCompatibleRuneLimit(t *testing.T) {
	t.Parallel()

	title := strings.Repeat("界", 300)
	got := duplicateTitle(title)
	if len([]rune(got)) != 255 || !strings.HasPrefix(got, "Copy of ") {
		t.Fatalf("duplicateTitle() rune length = %d", len([]rune(got)))
	}
}
