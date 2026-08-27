package integrationrequestshttp

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

func TestIntegrationRequestMutationRoutesRequireMemberRole(t *testing.T) {
	t.Parallel()

	routes, err := os.Open("routes.go")
	if err != nil {
		t.Fatalf("open integration-request routes: %v", err)
	}
	defer routes.Close()

	mutationRoutes := []string{
		"/integration-requests/accept-all",
		"/integration-requests/decline-all",
		"/integration-requests/{requestId}\"",
		"/integration-requests/{requestId}/comments",
		"/integration-requests/{requestId}/accept",
		"/integration-requests/{requestId}/decline",
	}
	found := make(map[string]bool, len(mutationRoutes))
	scanner := bufio.NewScanner(routes)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "app.Post(") && !strings.Contains(line, "app.Put(") {
			continue
		}
		for _, route := range mutationRoutes {
			if strings.Contains(line, route) {
				if !strings.Contains(line, "memberOnly") {
					t.Fatalf("integration-request mutation route %q is missing memberOnly: %s", route, line)
				}
				found[route] = true
			}
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan integration-request routes: %v", err)
	}
	for _, route := range mutationRoutes {
		if !found[route] {
			t.Fatalf("integration-request mutation route %q was not checked", route)
		}
	}
}
