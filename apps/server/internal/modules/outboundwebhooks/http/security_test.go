package outboundwebhookshttp

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func TestManagementAccessRejectsMachinePrincipalsBeforeWorkspaceResolution(t *testing.T) {
	t.Parallel()

	actor, err := platformauth.NewActor(
		uuid.New(),
		platformauth.PrincipalServiceAccount,
		uuid.New(),
		platformauth.MustScopeSet(platformauth.ScopeWebhooksManage),
		platformauth.UnrestrictedTeamAccess(),
	)
	if err != nil {
		t.Fatalf("create service-account actor: %v", err)
	}
	ctx, err := platformauth.SetActor(context.Background(), actor)
	if err != nil {
		t.Fatalf("set actor: %v", err)
	}

	if _, err := humanAccessFromContext(ctx); err != errAccessDenied {
		t.Fatalf("humanAccessFromContext error = %v, want %v", err, errAccessDenied)
	}
}

func TestEndpointMetadataNeverExposesOwnerOrStoredSecretMaterial(t *testing.T) {
	t.Parallel()

	endpoint := outboundwebhooksdomain.Endpoint{
		ID: uuid.New(), WorkspaceID: uuid.New(), OwnerPrincipalID: uuid.New(),
		Name: "Production", URL: "https://example.com/webhooks/fortyone",
		Status: outboundwebhooksdomain.EndpointActive, SecretGeneration: 2,
		SubscriptionGeneration: 3,
		Subscriptions:          []outboundwebhooksdomain.EventType{outboundwebhooksdomain.EventStoryCreated},
		CreatedAt:              time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	encoded, err := json.Marshal(endpointModel(endpoint))
	if err != nil {
		t.Fatalf("marshal endpoint model: %v", err)
	}
	body := string(encoded)
	if strings.Contains(body, endpoint.OwnerPrincipalID.String()) ||
		strings.Contains(strings.ToLower(body), "secret_payload") ||
		strings.Contains(strings.ToLower(body), "envelope") {
		t.Fatalf("endpoint metadata exposed protected state: %s", body)
	}
	if !strings.Contains(body, `"secretGeneration":2`) {
		t.Fatalf("endpoint metadata omitted safe generation: %s", body)
	}
}

func TestShowOnceResponsesDisableCaching(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	setShowOnceHeaders(recorder)

	if got := recorder.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
	if got := recorder.Header().Get("Pragma"); got != "no-cache" {
		t.Fatalf("Pragma = %q, want no-cache", got)
	}
}

func TestManagementRoutesKeepTheFullAuthorizationChain(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("routes.go")
	if err != nil {
		t.Fatalf("read routes: %v", err)
	}
	for _, route := range []string{
		`app.Get("/workspaces/{workspaceSlug}/webhook-endpoints"`,
		`app.Post("/workspaces/{workspaceSlug}/webhook-endpoints"`,
		`app.Put("/workspaces/{workspaceSlug}/webhook-endpoints/{endpointId}/subscriptions"`,
		`app.Post("/workspaces/{workspaceSlug}/webhook-endpoints/{endpointId}/rotate-secret"`,
		`app.Post("/workspaces/{workspaceSlug}/webhook-endpoints/{endpointId}/disable"`,
	} {
		line := routeLine(string(source), route)
		if line == "" {
			t.Fatalf("management route %q is not registered", route)
		}
		for _, middleware := range []string{"auth", "workspace", "adminOnly", "webhookScope"} {
			if !strings.Contains(line, middleware) {
				t.Fatalf("route %q is missing %s: %s", route, middleware, line)
			}
		}
		if !strings.Contains(route, `app.Get`) && !strings.Contains(line, "mutationLimit") {
			t.Fatalf("mutation route %q is missing mutationLimit: %s", route, line)
		}
	}
}

func routeLine(source, route string) string {
	for _, line := range strings.Split(source, "\n") {
		if strings.Contains(line, route) {
			return line
		}
	}
	return ""
}
