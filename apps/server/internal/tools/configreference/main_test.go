package main

import (
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

func TestParseConfigWalksNamedAndAnonymousStructs(t *testing.T) {
	t.Parallel()

	root := fstest.MapFS{
		"config.go": &fstest.MapFile{Data: []byte(`package fixture

import "time"

type HTTPConfig struct {
	Timeout time.Duration ` + "`" + `default:"5s" env:"APP_HTTP_TIMEOUT"` + "`" + `
}

type Config struct {
	HTTP HTTPConfig
	Auth struct {
		Secret string ` + "`" + `env:"APP_AUTH_SECRET"` + "`" + `
	}
}
`)},
	}

	bindings, err := parseConfig(root, "config.go", "API")
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}
	if len(bindings) != 2 {
		t.Fatalf("got %d bindings, want 2", len(bindings))
	}
	if bindings[0].Name != "APP_AUTH_SECRET" || bindings[0].Field != "Config.Auth.Secret" {
		t.Fatalf("unexpected anonymous binding: %+v", bindings[0])
	}
	if bindings[1].Name != "APP_HTTP_TIMEOUT" || bindings[1].Type != "time.Duration" || bindings[1].Default != "5s" {
		t.Fatalf("unexpected named binding: %+v", bindings[1])
	}
}

func TestMergeAndValidateRejectsExampleDrift(t *testing.T) {
	t.Parallel()

	bindings := []binding{{Name: "APP_REQUIRED", Process: "API"}}
	tests := map[string]map[string]exampleValue{
		"missing": {},
		"unknown": {
			"APP_REQUIRED": {},
			"APP_UNKNOWN":  {},
		},
	}
	for name, examples := range tests {
		name, examples := name, examples
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := mergeAndValidate(bindings, examples); err == nil {
				t.Fatal("expected example drift error")
			}
		})
	}
}

func TestMergeAndValidateRejectsDuplicateProcessBinding(t *testing.T) {
	t.Parallel()

	bindings := []binding{
		{Name: "APP_DUPLICATE", Process: "API"},
		{Name: "APP_DUPLICATE", Process: "API"},
	}
	if _, err := mergeAndValidate(bindings, map[string]exampleValue{"APP_DUPLICATE": {}}); err == nil {
		t.Fatal("expected duplicate binding error")
	}
}

func TestRenderReferenceRedactsSensitiveDefaults(t *testing.T) {
	t.Parallel()

	variables := []variable{{
		Name: "APP_AUTH_SECRET_KEY",
		Bindings: []binding{{
			Name:       "APP_AUTH_SECRET_KEY",
			Process:    "API",
			Field:      "Config.Auth.SecretKey",
			Type:       "string",
			Default:    "must-not-appear",
			SourcePath: "config.go",
			Line:       10,
		}},
	}}

	rendered := string(renderReference(variables))
	if strings.Contains(rendered, "must-not-appear") {
		t.Fatal("generated reference leaked a sensitive default")
	}
	if !strings.Contains(rendered, "development placeholder") {
		t.Fatal("generated reference did not describe the redacted default")
	}
}

func TestSensitiveClassificationAvoidsIdentifiersAndQuotas(t *testing.T) {
	t.Parallel()

	tests := map[string]bool{
		"APP_AUTH_SECRET_KEY":                       true,
		"APP_AZURE_STORAGE_ACCOUNT_KEY":             true,
		"APP_INVITATION_TOKEN_HMAC_PREVIOUS_KEYS":   true,
		"APP_INVITATION_TOKEN_HMAC_KEY_ID":          false,
		"GOOGLE_DRIVE_PICKER_API_KEY":               false,
		"OPENAI_ASSISTANT_WORKSPACE_TOKENS_PER_DAY": false,
	}
	for name, want := range tests {
		if got := isSensitive(name); got != want {
			t.Errorf("isSensitive(%q) = %t, want %t", name, got, want)
		}
	}
}

func TestVariableCategoryClassifiesGoogleDriveAsProviderIntegration(t *testing.T) {
	t.Parallel()

	if got := variableCategory("GOOGLE_DRIVE_CLIENT_ID"); got != "Provider integrations" {
		t.Fatalf("variableCategory(GOOGLE_DRIVE_CLIENT_ID) = %q, want Provider integrations", got)
	}
}

func TestParseExamplesCarriesOnlyAdjacentComments(t *testing.T) {
	t.Parallel()

	root := fstest.MapFS{
		examplePath: &fstest.MapFile{Data: []byte("# First note\nAPP_FIRST=value\n\n# Second note\nAPP_SECOND=\n")},
	}
	examples, err := parseExamples(fs.FS(root))
	if err != nil {
		t.Fatalf("parse examples: %v", err)
	}
	if examples["APP_FIRST"].Notes != "First note" || examples["APP_SECOND"].Notes != "Second note" {
		t.Fatalf("unexpected notes: %+v", examples)
	}
}
