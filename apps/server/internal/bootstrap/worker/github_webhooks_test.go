package workerbootstrap

import (
	"testing"

	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
)

func TestBuildGitHubWebhookRuntimePreservesOptionalIntegration(t *testing.T) {
	t.Parallel()
	gateway, inbox, payloads, err := buildGitHubWebhookRuntime(nil, nil, nil, github.Config{})
	if err != nil || gateway != nil || inbox != nil || payloads != nil {
		t.Fatalf("buildGitHubWebhookRuntime(unconfigured) = (%v, %v, %v, %v)", gateway, inbox, payloads, err)
	}
}
