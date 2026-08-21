package workspaceurl

import "testing"

func TestBuild(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		websiteURL string
		workspace  string
		want       string
	}{
		{
			name:       "hosted root domain",
			websiteURL: "https://fortyone.app",
			workspace:  "acme",
			want:       "https://acme.fortyone.app/settings/workspace/integrations/figma",
		},
		{
			name:       "hosted cloud domain",
			websiteURL: "https://cloud.fortyone.app",
			workspace:  "acme",
			want:       "https://acme.fortyone.app/settings/workspace/integrations/figma",
		},
		{
			name:       "existing workspace domain",
			websiteURL: "https://other.fortyone.app",
			workspace:  "acme",
			want:       "https://acme.fortyone.app/settings/workspace/integrations/figma",
		},
		{
			name:       "localhost path routing",
			websiteURL: "http://localhost:3000",
			workspace:  "acme",
			want:       "http://localhost:3000/acme/settings/workspace/integrations/figma",
		},
		{
			name:       "self-hosted path routing",
			websiteURL: "https://projects.example.com/app",
			workspace:  "acme",
			want:       "https://projects.example.com/app/acme/settings/workspace/integrations/figma",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := Build(
				test.websiteURL,
				test.workspace,
				"settings",
				"workspace",
				"integrations",
				"figma",
			)
			if got != test.want {
				t.Fatalf("Build() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestBuildRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		websiteURL string
		workspace  string
	}{
		{name: "invalid URL", websiteURL: "://invalid", workspace: "acme"},
		{name: "unsafe scheme", websiteURL: "javascript://fortyone.app", workspace: "acme"},
		{name: "embedded credentials", websiteURL: "https://user@fortyone.app", workspace: "acme"},
		{name: "empty workspace", websiteURL: "https://fortyone.app"},
		{name: "unsafe workspace", websiteURL: "https://fortyone.app", workspace: "acme.example.com"},
		{name: "uppercase workspace", websiteURL: "https://fortyone.app", workspace: "Acme"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := Build(test.websiteURL, test.workspace, "settings"); got != "" {
				t.Fatalf("Build() = %q, want an empty URL", got)
			}
		})
	}
}
