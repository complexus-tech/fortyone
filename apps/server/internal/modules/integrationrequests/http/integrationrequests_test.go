package integrationrequestshttp

import (
	"database/sql"
	"net/http"
	"testing"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
)

func TestRequestErrorStatusMasksInaccessibleRequestsAsNotFound(t *testing.T) {
	t.Parallel()

	if status := requestErrorStatus(sql.ErrNoRows); status != http.StatusNotFound {
		t.Fatalf("inaccessible request status = %d, want %d", status, http.StatusNotFound)
	}
}

func TestRequestErrorStatusReturnsConflictForIdempotencyReuse(t *testing.T) {
	t.Parallel()

	if status := requestErrorStatus(integrationrequests.ErrIdempotencyConflict); status != http.StatusConflict {
		t.Fatalf("idempotency conflict status = %d, want %d", status, http.StatusConflict)
	}
}
