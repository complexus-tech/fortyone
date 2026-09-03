package workerbootstrap

import (
	"testing"

	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
	"github.com/complexus-tech/projects-api/pkg/tasks"
)

func TestBuildGitHubWebhookRuntimeRequiresWorkerDependencies(t *testing.T) {
	t.Parallel()
	inbox, runtime, err := buildGitHubWebhookRuntime(nil, (*tasks.Service)(nil), github.Config{})
	if err == nil || inbox != nil || runtime.Payloads != nil || runtime.Dispatcher != nil {
		t.Fatalf("buildGitHubWebhookRuntime(unconfigured) = (%v, %#v, %v)", inbox, runtime, err)
	}
}
