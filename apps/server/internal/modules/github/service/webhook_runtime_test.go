package github

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	githubshared "github.com/complexus-tech/projects-api/internal/modules/github/shared"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
)

type webhookInstallationStub struct {
	installation githubshared.WebhookInstallation
	err          error
	calls        int
}

func (stub *webhookInstallationStub) GetAuthorizedWebhookInstallation(
	_ context.Context,
	_, _ int64,
) (githubshared.WebhookInstallation, error) {
	stub.calls++
	return stub.installation, stub.err
}

func (stub *webhookInstallationStub) GetCurrentWebhookInstallation(
	_ context.Context,
	_, _ uuid.UUID,
	_ int64,
) (githubshared.WebhookInstallation, error) {
	stub.calls++
	return stub.installation, stub.err
}

type githubWebhookQueueStub struct{ payloads []tasks.GitHubWebhookPayload }

func (stub *githubWebhookQueueStub) EnqueueGitHubWebhook(
	_ context.Context,
	payload tasks.GitHubWebhookPayload,
) error {
	stub.payloads = append(stub.payloads, payload)
	return nil
}

func TestGitHubWebhookVerifierAuthenticatesExactBytesBeforeParsing(t *testing.T) {
	t.Parallel()
	installation := githubshared.WebhookInstallation{
		ID: uuid.New(), WorkspaceID: uuid.New(), ExternalInstallationID: 41,
		InstallationGeneration: uuid.New(), RepositoryID: uuid.New(), ExternalRepositoryID: 42,
	}
	repository := &webhookInstallationStub{installation: installation}
	queue := &githubWebhookQueueStub{}
	runtime, err := NewWebhookRuntime(repository, queue, Config{
		WebhookSecret:        "github-signing-secret",
		WebhookPayloadSecret: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	})
	if err != nil {
		t.Fatalf("NewWebhookRuntime() error = %v", err)
	}
	body := []byte(`{"installation":{"id":41},"repository":{"id":42}}`)
	request := signedGitHubRequest(body, "delivery-1", "issues")
	request.Body = append(request.Body, ' ')
	if _, err := runtime.Registration.Verifier.Verify(t.Context(), request); !errors.Is(err, webhooks.ErrUnauthenticated) {
		t.Fatalf("Verify(tampered) error = %v, want ErrUnauthenticated", err)
	}
	if repository.calls != 0 {
		t.Fatalf("repository calls before signature verification = %d", repository.calls)
	}

	request = signedGitHubRequest(body, "delivery-1", "issues")
	verified, err := runtime.Registration.Verifier.Verify(t.Context(), request)
	if err != nil {
		t.Fatalf("Verify(valid) error = %v", err)
	}
	if verified.DeliveryID != "delivery-1" || verified.InstallationID != installation.ID ||
		verified.InstallationGeneration != installation.InstallationGeneration || repository.calls != 1 {
		t.Fatalf("Verify(valid) = %#v, repository calls = %d", verified, repository.calls)
	}
}

func TestGitHubWebhookVerifierRejectsAmbiguousHeadersAndMethods(t *testing.T) {
	t.Parallel()
	body := []byte(`{"installation":{"id":41},"repository":{"id":42}}`)
	installation := githubshared.WebhookInstallation{
		ID: uuid.New(), WorkspaceID: uuid.New(), ExternalInstallationID: 41,
		InstallationGeneration: uuid.New(), RepositoryID: uuid.New(), ExternalRepositoryID: 42,
	}

	tests := []struct {
		name   string
		mutate func(*webhooks.SignedRequest)
	}{
		{
			name: "non post method",
			mutate: func(request *webhooks.SignedRequest) {
				request.Method = "GET"
			},
		},
		{
			name: "duplicate signature",
			mutate: func(request *webhooks.SignedRequest) {
				request.Headers["X-Hub-Signature-256"] = append(request.Headers["X-Hub-Signature-256"], request.Headers["X-Hub-Signature-256"][0])
			},
		},
		{
			name: "duplicate delivery id",
			mutate: func(request *webhooks.SignedRequest) {
				request.Headers["X-GitHub-Delivery"] = []string{"delivery-1", "delivery-2"}
			},
		},
		{
			name: "duplicate event",
			mutate: func(request *webhooks.SignedRequest) {
				request.Headers["X-GitHub-Event"] = []string{"issues", "push"}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := &webhookInstallationStub{installation: installation}
			verifier := &githubWebhookVerifier{installations: repository, secret: "github-signing-secret"}
			request := signedGitHubRequest(body, "delivery-1", "issues")
			test.mutate(&request)
			if _, err := verifier.Verify(t.Context(), request); err == nil {
				t.Fatal("Verify() error = nil, want rejection")
			}
			if repository.calls != 0 {
				t.Fatalf("repository calls = %d, want 0", repository.calls)
			}
		})
	}
}

func TestGitHubWebhookVerifierClassifiesAuthorizationStoreOutage(t *testing.T) {
	t.Parallel()
	const sensitiveCause = "database connection secret"
	repository := &webhookInstallationStub{err: errors.New(sensitiveCause)}
	verifier := &githubWebhookVerifier{installations: repository, secret: "github-signing-secret"}
	body := []byte(`{"installation":{"id":41},"repository":{"id":42}}`)

	_, err := verifier.Verify(t.Context(), signedGitHubRequest(body, "delivery-1", "issues"))
	if !errors.Is(err, webhooks.ErrVerificationUnavailable) || strings.Contains(err.Error(), sensitiveCause) {
		t.Fatalf("Verify() error = %q", err)
	}
}

func TestGitHubWebhookDispatcherCarriesOnlyDurableIdentity(t *testing.T) {
	t.Parallel()
	queue := &githubWebhookQueueStub{}
	dispatcher := githubWebhookDispatcher{queue: queue}
	inboxID := uuid.New()
	if err := dispatcher.Enqueue(t.Context(), webhooks.Task{Provider: githubWebhookProvider, InboxID: inboxID}); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}
	if len(queue.payloads) != 1 || queue.payloads[0].InboxID != inboxID {
		t.Fatalf("queued payloads = %#v", queue.payloads)
	}
}

func TestGitHubWebhookWorkerRuntimeDoesNotRequireIngressSigningSecret(t *testing.T) {
	t.Parallel()
	queue := &githubWebhookQueueStub{}
	runtime, err := NewWebhookWorkerRuntime(queue, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatalf("NewWebhookWorkerRuntime() error = %v", err)
	}
	if runtime.Payloads == nil || runtime.Dispatcher == nil {
		t.Fatalf("NewWebhookWorkerRuntime() = %#v", runtime)
	}
	if err := runtime.Dispatcher.Enqueue(t.Context(), webhooks.Task{
		Provider: githubWebhookProvider,
		InboxID:  uuid.New(),
	}); err != nil {
		t.Fatalf("worker dispatcher enqueue error = %v", err)
	}
	if len(queue.payloads) != 1 {
		t.Fatalf("queued payloads = %#v", queue.payloads)
	}
}

func TestGitHubWebhookActorIsWorkspaceAndInstallationBound(t *testing.T) {
	t.Parallel()
	workspaceID := uuid.New()
	installationID := uuid.New()
	actorID := uuid.New()
	service := &Service{cfg: Config{GitHubUserID: actorID}}

	ctx, err := service.webhookActorContext(t.Context(), webhooks.Record{Envelope: webhooks.Envelope{
		WorkspaceID:    workspaceID,
		InstallationID: installationID,
	}})
	if err != nil {
		t.Fatalf("webhookActorContext() error = %v", err)
	}
	actor, err := platformauth.GetActor(ctx)
	if err != nil {
		t.Fatalf("GetActor() error = %v", err)
	}
	if actor.PrincipalID != actorID || actor.Kind != platformauth.PrincipalSystem ||
		actor.CredentialID != installationID || actor.WorkspaceID != workspaceID ||
		!actor.Scopes.ContainsAll(platformauth.ScopeStoriesRead, platformauth.ScopeStoriesWrite) {
		t.Fatalf("webhook actor = %#v", actor)
	}
}

func TestCurrentWebhookGrantFencesReinstalledAuthorization(t *testing.T) {
	t.Parallel()
	record := webhooks.Record{Envelope: webhooks.Envelope{
		ExternalAccountID:      "41",
		WorkspaceID:            uuid.New(),
		InstallationID:         uuid.New(),
		InstallationGeneration: uuid.New(),
	}}
	payload := webhookEnvelope{}
	payload.Installation.ID = 41
	payload.Repository.ID = 42
	repository := &webhookInstallationStub{err: githubshared.ErrWebhookInstallationNotFound}
	service := &Service{webhookInstallations: repository}
	current, err := service.currentWebhookGrant(t.Context(), record, payload)
	if err != nil || current {
		t.Fatalf("currentWebhookGrant(stale) = (%t, %v)", current, err)
	}
	repository.err = nil
	repository.installation = githubshared.WebhookInstallation{
		ID: record.InstallationID, WorkspaceID: record.WorkspaceID,
		InstallationGeneration: record.InstallationGeneration,
		ExternalInstallationID: 41,
		ExternalRepositoryID:   42,
	}
	current, err = service.currentWebhookGrant(t.Context(), record, payload)
	if err != nil || !current {
		t.Fatalf("currentWebhookGrant(current) = (%t, %v)", current, err)
	}
	repository.installation.ExternalRepositoryID = 43
	current, err = service.currentWebhookGrant(t.Context(), record, payload)
	if err != nil || current {
		t.Fatalf("currentWebhookGrant(repository mismatch) = (%t, %v)", current, err)
	}
}

func signedGitHubRequest(body []byte, deliveryID, eventType string) webhooks.SignedRequest {
	mac := hmac.New(sha256.New, []byte("github-signing-secret"))
	_, _ = mac.Write(body)
	return webhooks.SignedRequest{
		Method: "POST",
		Headers: webhooks.Headers{
			"X-GitHub-Delivery":   {deliveryID},
			"X-GitHub-Event":      {eventType},
			"X-Hub-Signature-256": {"sha256=" + hex.EncodeToString(mac.Sum(nil))},
		},
		Body: body,
	}
}

func FuzzGitHubWebhookSignatureAuthenticatesExactBytes(f *testing.F) {
	f.Add([]byte(`{"installation":{"id":41},"repository":{"id":42}}`))
	f.Add([]byte("line one\r\nline two"))
	f.Fuzz(func(t *testing.T, body []byte) {
		mac := hmac.New(sha256.New, []byte("github-signing-secret"))
		_, _ = mac.Write(body)
		signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
		if !verifyGitHubWebhookSignature("github-signing-secret", body, signature) {
			t.Fatal("valid exact-byte signature was rejected")
		}

		tampered := append(append([]byte(nil), body...), 0)
		if verifyGitHubWebhookSignature("github-signing-secret", tampered, signature) {
			t.Fatal("signature authenticated a modified body")
		}
	})
}
