package admindomain_test

import (
	"errors"
	"testing"

	admindomain "github.com/complexus-tech/projects-api/internal/modules/admin/domain"
)

func TestWorkspaceStatusParserIsFinite(t *testing.T) {
	tests := []struct {
		input string
		want  admindomain.WorkspaceStatus
	}{
		{input: "", want: admindomain.WorkspaceStatusAll},
		{input: " TRIALING ", want: admindomain.WorkspaceStatusTrialing},
		{input: "PAST_DUE", want: admindomain.WorkspaceStatusPastDue},
	}
	for _, test := range tests {
		got, err := admindomain.ParseWorkspaceStatus(test.input)
		if err != nil || got != test.want {
			t.Fatalf("ParseWorkspaceStatus(%q) = %q, %v; want %q", test.input, got, err, test.want)
		}
	}
	if _, err := admindomain.ParseWorkspaceStatus("paid' OR true"); !errors.Is(err, admindomain.ErrInvalidFilter) {
		t.Fatalf("unsafe status error = %v, want ErrInvalidFilter", err)
	}
}

func TestAuditActionParserIsFinite(t *testing.T) {
	action, err := admindomain.ParseAuditAction(" Subscription.Sync_Failed ")
	if err != nil || action != admindomain.AuditSubscriptionSyncFailed {
		t.Fatalf("parsed action = %q, %v", action, err)
	}
	if _, err := admindomain.ParseAuditAction("subscription.anything"); !errors.Is(err, admindomain.ErrInvalidFilter) {
		t.Fatalf("unknown action error = %v, want ErrInvalidFilter", err)
	}
}

func TestIntegrationFiltersAreFinite(t *testing.T) {
	provider, err := admindomain.ParseIntegrationProvider(" GitHub ")
	if err != nil || provider != admindomain.IntegrationProviderGitHub {
		t.Fatalf("parsed provider = %q, %v", provider, err)
	}
	status, err := admindomain.ParseIntegrationStatus(" NOT_CONNECTED ")
	if err != nil || status != admindomain.IntegrationStatusNotConnected {
		t.Fatalf("parsed status = %q, %v", status, err)
	}
	if _, err := admindomain.ParseIntegrationProvider("github' OR true"); !errors.Is(err, admindomain.ErrInvalidFilter) {
		t.Fatalf("unsafe provider error = %v, want ErrInvalidFilter", err)
	}
	if _, err := admindomain.ParseIntegrationStatus("unknown"); !errors.Is(err, admindomain.ErrInvalidFilter) {
		t.Fatalf("unknown status error = %v, want ErrInvalidFilter", err)
	}
}

func TestAdminNoteTargetsAreRestricted(t *testing.T) {
	if _, err := admindomain.ParseNoteTargetType("subscription"); !errors.Is(err, admindomain.ErrInvalidAction) {
		t.Fatalf("subscription note target error = %v, want ErrInvalidAction", err)
	}
	if target, err := admindomain.ParseNoteTargetType("user"); err != nil || target != admindomain.TargetUser {
		t.Fatalf("user note target = %q, %v", target, err)
	}
}
