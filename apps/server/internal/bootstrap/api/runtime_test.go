package api

import (
	"strings"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/http/mux"
)

func TestBuildRuntimeRequiresNativeDatabasePool(t *testing.T) {
	_, err := BuildRuntime(mux.Config{}, Dependencies{}, "", nil)
	if err == nil || !strings.Contains(err.Error(), "database pool is required") {
		t.Fatalf("error = %v, want missing database pool", err)
	}
}
