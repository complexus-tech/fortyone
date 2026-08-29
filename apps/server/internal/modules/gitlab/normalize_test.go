package gitlab

import (
	"errors"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/codehost"
)

func TestNormalizeWebhookMapsIssueAndIssueComment(t *testing.T) {
	t.Parallel()
	adapter := &Adapter{}
	issueBody := []byte(`{
        "object_kind":"issue","user":{"id":19,"username":"joseph"},
        "project":{"id":42,"name":"fortyone","path_with_namespace":"complexus/fortyone","web_url":"https://gitlab.example/complexus/fortyone","default_branch":"main","namespace":"complexus"},
        "object_attributes":{"id":501,"iid":7,"title":"Typed integration","description":"Neutral contract","state":"opened","url":"https://gitlab.example/complexus/fortyone/-/issues/7","action":"open"}
    }`)
	event, err := adapter.NormalizeWebhook(t.Context(), "delivery-1", "Issue Hook", issueBody)
	if err != nil {
		t.Fatalf("NormalizeWebhook(issue) error = %v", err)
	}
	if event.Kind != codehost.EventWorkItemChanged || event.Action != "open" ||
		event.WorkItem == nil || event.WorkItem.Number != 7 || event.WorkItem.Repository.ExternalID != "42" {
		t.Fatalf("NormalizeWebhook(issue) = %#v", event)
	}

	noteBody := []byte(`{
        "object_kind":"note","user":{"id":19,"username":"joseph"},
        "project":{"id":42,"name":"fortyone","path_with_namespace":"complexus/fortyone","web_url":"https://gitlab.example/complexus/fortyone","default_branch":"main","namespace":"complexus"},
        "object_attributes":{"id":9001,"note":"Looks good","noteable_type":"Issue","url":"https://gitlab.example/complexus/fortyone/-/issues/7#note_9001","created_at":"2026-08-28T12:00:00Z"},
        "issue":{"id":501,"iid":7,"title":"Typed integration","description":"Neutral contract","state":"opened","url":"https://gitlab.example/complexus/fortyone/-/issues/7"}
    }`)
	event, err = adapter.NormalizeWebhook(t.Context(), "delivery-2", "Note Hook", noteBody)
	if err != nil {
		t.Fatalf("NormalizeWebhook(note) error = %v", err)
	}
	if event.Kind != codehost.EventCommentCreated || event.Action != "create" ||
		event.Comment == nil || event.Comment.ExternalID != "9001" || event.Comment.WorkItem.Number != 7 {
		t.Fatalf("NormalizeWebhook(note) = %#v", event)
	}
}

func TestNormalizeWebhookRejectsUnprovenCapabilities(t *testing.T) {
	t.Parallel()
	adapter := &Adapter{}
	if _, err := adapter.NormalizeWebhook(t.Context(), "delivery", "Merge Request Hook", []byte(`{}`)); !errors.Is(err, codehost.ErrCapabilityUnsupported) {
		t.Fatalf("NormalizeWebhook(merge request) error = %v", err)
	}
	mergeRequestNote := []byte(`{
        "object_kind":"note","user":{"id":19},
        "project":{"id":42,"name":"fortyone","path_with_namespace":"complexus/fortyone","web_url":"https://gitlab.example/complexus/fortyone"},
        "object_attributes":{"id":9001,"noteable_type":"MergeRequest"}
    }`)
	if _, err := adapter.NormalizeWebhook(t.Context(), "delivery", "Note Hook", mergeRequestNote); !errors.Is(err, codehost.ErrCapabilityUnsupported) {
		t.Fatalf("NormalizeWebhook(merge-request note) error = %v", err)
	}
}
