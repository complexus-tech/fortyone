package api

import (
	"testing"

	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
)

func TestBuildGitHubWebhookGatewayPreservesOptionalIntegration(t *testing.T) {
	t.Parallel()
	gateway, inbox, payloads, err := buildGitHubWebhookGateway(nil, nil, nil, github.Config{})
	if err != nil || gateway != nil || inbox != nil || payloads != nil {
		t.Fatalf("buildGitHubWebhookGateway(unconfigured) = (%v, %v, %v, %v)", gateway, inbox, payloads, err)
	}
}
