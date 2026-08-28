package github

import (
	"errors"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/codehost"
)

func TestGitHubCodeHostAdapterContract(t *testing.T) {
	t.Parallel()
	service := &Service{}
	for _, capability := range []codehost.Capability{
		codehost.CapabilityInstallationAuth,
		codehost.CapabilityRepositoryCatalog,
		codehost.CapabilityWorkItemWriter,
		codehost.CapabilityCommentWriter,
		codehost.CapabilityWebhookNormalization,
	} {
		if err := service.Capabilities().Require(capability); err != nil {
			t.Fatalf("Require(%q) error = %v", capability, err)
		}
	}
	body := []byte(`{
		"action":"opened","repository":{"id":42,"name":"fortyone","full_name":"complexus/fortyone","html_url":"https://github.com/complexus/fortyone","default_branch":"main","private":true,"owner":{"login":"complexus"}},
		"sender":{"id":19,"login":"joseph"},
		"issue":{"id":501,"number":7,"title":"Typed integration","body":"Neutral contract","state":"open","html_url":"https://github.com/complexus/fortyone/issues/7"}
	}`)
	event, err := service.NormalizeWebhook(t.Context(), "delivery-1", "issues", body)
	if err != nil {
		t.Fatalf("NormalizeWebhook() error = %v", err)
	}
	if event.Provider != githubWebhookProvider || event.Kind != codehost.EventWorkItemChanged ||
		event.WorkItem == nil || event.WorkItem.Number != 7 || event.WorkItem.Repository.ExternalID != "42" {
		t.Fatalf("NormalizeWebhook() = %#v", event)
	}
	if _, err := service.NormalizeWebhook(t.Context(), "delivery-2", "deployment", []byte(`{}`)); !errors.Is(err, codehost.ErrCapabilityUnsupported) {
		t.Fatalf("NormalizeWebhook(unsupported) error = %v, want ErrCapabilityUnsupported", err)
	}
}
