package githubhttp

import (
	"net/http"
	"strings"
	"testing"
)

func TestGitHubSetupCallbackParametersAreBoundedAndUnambiguous(t *testing.T) {
	t.Parallel()

	request, err := http.NewRequest(http.MethodGet, "/integrations/github/setup?installation_id=41&state=opaque", nil)
	if err != nil {
		t.Fatal(err)
	}
	installationID, state, err := githubSetupCallbackParameters(request)
	if err != nil || installationID != 41 || state != "opaque" {
		t.Fatalf("callback = (%d, %q, %v)", installationID, state, err)
	}

	for _, test := range []struct {
		name   string
		target string
	}{
		{name: "repeated installation", target: "/integrations/github/setup?installation_id=41&installation_id=42&state=opaque"},
		{name: "repeated state", target: "/integrations/github/setup?installation_id=41&state=first&state=second"},
		{name: "empty state", target: "/integrations/github/setup?installation_id=41&state="},
		{name: "zero installation", target: "/integrations/github/setup?installation_id=0&state=opaque"},
		{name: "overflowing installation", target: "/integrations/github/setup?installation_id=99999999999999999999&state=opaque"},
		{name: "oversized state", target: "/integrations/github/setup?installation_id=41&state=" + strings.Repeat("x", 4097)},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request, err := http.NewRequest(http.MethodGet, test.target, nil)
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err := githubSetupCallbackParameters(request); err == nil {
				t.Fatalf("invalid callback accepted for %q", test.target)
			}
		})
	}
}
