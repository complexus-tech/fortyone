//go:build integration

package gitlab

import (
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	webhooksrepository "github.com/complexus-tech/projects-api/internal/platform/webhooks/repository"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
)

func TestGitLabWebhookGatewayPersistsAndDeduplicatesOnPostgres18(t *testing.T) {
	t.Parallel()
	postgres := testkit.NewPostgres(t)
	workspaceID := uuid.New()
	ownerID := uuid.New()
	suffix := uuid.NewString()
	if _, err := postgres.Pool.Exec(t.Context(), `
		INSERT INTO users (user_id, username, email, full_name)
		VALUES ($1, $2, $3, 'GitLab proof owner')
	`, ownerID, "gitlab-proof-"+suffix, "gitlab-proof-"+suffix+"@example.com"); err != nil {
		t.Fatalf("insert owner: %v", err)
	}
	if _, err := postgres.Pool.Exec(t.Context(), `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, 'GitLab proof workspace', $2, $3)
	`, workspaceID, "gitlab-proof-"+suffix, ownerID); err != nil {
		t.Fatalf("insert workspace: %v", err)
	}
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	installation := testInstallation()
	installation.WorkspaceID = workspaceID
	resolver := &resolverStub{installation: installation}
	queue := &queueStub{}
	runtime := newTestWebhookRuntime(t, resolver, queue, now)
	catalog, err := integrations.NewRegistry(ProviderDescriptor())
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	runtimes, err := webhooks.NewRuntimeRegistry(catalog, runtime.Registration)
	if err != nil {
		t.Fatalf("NewRuntimeRegistry() error = %v", err)
	}
	inbox := webhooksrepository.New(postgres.Pool)
	gateway, err := webhooks.NewGateway(inbox, runtimes, webhooks.Config{
		MaxBodyBytes: 1 << 20, PayloadRetention: time.Hour, Now: func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("NewGateway() error = %v", err)
	}
	request := signedGitLabRequest([]byte(`{"project":{"id":42}}`), now, "gitlab-delivery-1")
	first, err := gateway.Receive(t.Context(), ProviderKey, request)
	if err != nil || !first.Created || !first.Queued || len(queue.tasks) != 1 {
		t.Fatalf("first Receive() = (%#v, %v), tasks = %#v", first, err, queue.tasks)
	}
	record, process, err := inbox.Start(t.Context(), first.ID, now.Add(time.Second), time.Minute)
	if err != nil || !process || record.ID != first.ID {
		t.Fatalf("Start() = (%#v, %v, %v)", record, process, err)
	}
	if err := inbox.Complete(t.Context(), first.ID, webhooks.StatusCompleted, "gitlab.normalized", now.Add(2*time.Second)); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	second, err := gateway.Receive(t.Context(), ProviderKey, request)
	if err != nil || second.Created || second.Queued || second.ID != first.ID || second.Status != webhooks.StatusCompleted {
		t.Fatalf("duplicate Receive() = (%#v, %v)", second, err)
	}
	if len(queue.tasks) != 1 {
		t.Fatalf("terminal duplicate queued again: %#v", queue.tasks)
	}
}
