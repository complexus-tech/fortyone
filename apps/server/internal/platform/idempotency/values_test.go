package idempotency

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func TestParseKeyBoundsAndRedaction(t *testing.T) {
	t.Parallel()

	raw := "0123456789abcdef0123456789abcdef"
	key, err := ParseKey(raw)
	if err != nil {
		t.Fatalf("ParseKey() error = %v", err)
	}
	for _, format := range []string{"%v", "%+v", "%#v", "%s", "%q"} {
		formatted := fmt.Sprintf(format, key)
		if strings.Contains(formatted, raw) || formatted != "[REDACTED]" {
			t.Fatalf("formatted key with %s = %q, want redacted", format, formatted)
		}
	}

	tests := []struct {
		name  string
		value string
	}{
		{name: "too short", value: strings.Repeat("a", MinKeyBytes-1)},
		{name: "too long", value: strings.Repeat("a", MaxKeyBytes+1)},
		{name: "space", value: strings.Repeat("a", MinKeyBytes) + " "},
		{name: "control", value: strings.Repeat("a", MinKeyBytes) + "\n"},
		{name: "unicode", value: strings.Repeat("a", MinKeyBytes) + "é"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := ParseKey(test.value); !errors.Is(err, ErrInvalidKey) {
				t.Fatalf("ParseKey() error = %v, want ErrInvalidKey", err)
			}
		})
	}
}

func TestKeyDigestAndRequestHashUseExactBytes(t *testing.T) {
	t.Parallel()

	key, err := ParseKey("Case-Sensitive-Key-0123456789")
	if err != nil {
		t.Fatalf("ParseKey() error = %v", err)
	}
	wantKey := sha256.Sum256([]byte("Case-Sensitive-Key-0123456789"))
	if got := key.digest(); got != wantKey {
		t.Fatalf("key digest = %x, want SHA-256 digest", got)
	}

	body := []byte("{\"title\":\"exact bytes\"}\n")
	wantRequest := sha256.Sum256(body)
	if got := HashRequest(body); got != wantRequest {
		t.Fatalf("request hash = %x, want SHA-256 digest", got)
	}
	if HashRequest(body) == HashRequest(body[:len(body)-1]) {
		t.Fatal("request hashing must include trailing bytes")
	}
}

func TestRequestBodyLimit(t *testing.T) {
	t.Parallel()

	if err := validateRequestBody(make([]byte, MaxRequestBodyBytes)); err != nil {
		t.Fatalf("validateRequestBody() boundary error = %v", err)
	}
	if err := validateRequestBody(make([]byte, MaxRequestBodyBytes+1)); !errors.Is(err, ErrRequestTooLarge) {
		t.Fatalf("validateRequestBody() error = %v, want ErrRequestTooLarge", err)
	}
}

func TestScopeDerivesOptionalWorkspaceFromActor(t *testing.T) {
	t.Parallel()

	operation, err := ParseOperation("stories.create")
	if err != nil {
		t.Fatalf("ParseOperation() error = %v", err)
	}
	actor := auth.NewHumanActor(uuid.New())
	workspaceID := uuid.New()
	actor, err = actor.WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("WithWorkspace() error = %v", err)
	}

	scope, err := NewScope(actor, MethodPost, operation)
	if err != nil {
		t.Fatalf("NewScope() error = %v", err)
	}
	if scope.principalID != actor.PrincipalID || scope.workspaceID != workspaceID || scope.method != MethodPost {
		t.Fatalf("scope does not preserve actor and route identity: %#v", scope)
	}
}

func TestOAuthApplicationScopeUsesStableInstallationIdentity(t *testing.T) {
	t.Parallel()

	operation, err := ParseOperation("stories.create")
	if err != nil {
		t.Fatalf("ParseOperation() error = %v", err)
	}
	principalID := uuid.New()
	installationID := uuid.New()
	actor, err := auth.NewActor(
		principalID,
		auth.PrincipalOAuthApplication,
		installationID,
		auth.MustScopeSet(auth.ScopeStoriesWrite),
		auth.UnrestrictedTeamAccess(),
	)
	if err != nil {
		t.Fatalf("NewActor() error = %v", err)
	}
	workspaceID := uuid.New()
	actor, err = actor.WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("WithWorkspace() error = %v", err)
	}

	scope, err := NewScope(actor, MethodPost, operation)
	if err != nil {
		t.Fatalf("NewScope() error = %v", err)
	}
	if scope.principalID != installationID {
		t.Fatalf("scope identity = %s, want installation %s", scope.principalID, installationID)
	}
	if scope.principalID == principalID {
		t.Fatal("scope must not use the OAuth application principal as its retry identity")
	}
}

func TestOAuthApplicationScopeRejectsMissingInstallationIdentity(t *testing.T) {
	t.Parallel()

	operation, err := ParseOperation("stories.create")
	if err != nil {
		t.Fatalf("ParseOperation() error = %v", err)
	}
	actor, err := auth.NewActor(
		uuid.New(),
		auth.PrincipalOAuthApplication,
		uuid.Nil,
		auth.MustScopeSet(auth.ScopeStoriesWrite),
		auth.UnrestrictedTeamAccess(),
	)
	if err != nil {
		t.Fatalf("NewActor() error = %v", err)
	}

	if _, err := NewScope(actor, MethodPost, operation); !errors.Is(err, ErrInvalidScope) {
		t.Fatalf("NewScope() error = %v, want ErrInvalidScope", err)
	}
}

func TestResponseIsBoundedAndCannotRepresentHeaders(t *testing.T) {
	t.Parallel()

	response, err := NewResponse(201, []byte(`{"id":"story"}`), "application/json; charset=utf-8")
	if err != nil {
		t.Fatalf("NewResponse() error = %v", err)
	}
	body := response.Body()
	body[0] = 'X'
	if string(response.Body()) != `{"id":"story"}` {
		t.Fatal("response body accessor exposed mutable receipt storage")
	}
	responseType := reflect.TypeOf(response)
	for _, forbidden := range []string{"Header", "Headers", "SetCookie", "Cookies"} {
		if _, exists := responseType.FieldByName(forbidden); exists {
			t.Fatalf("Response unexpectedly exposes %s replay metadata", forbidden)
		}
	}

	if _, err := NewResponse(201, make([]byte, MaxResponseBodyBytes+1), "application/json"); !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("oversized response error = %v, want ErrInvalidResponse", err)
	}
	if _, err := NewResponse(201, nil, "application/json\r\nSet-Cookie: secret=1"); !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("header injection response error = %v, want ErrInvalidResponse", err)
	}
}

func TestConfigBounds(t *testing.T) {
	t.Parallel()

	if err := DefaultConfig().validate(); err != nil {
		t.Fatalf("DefaultConfig().validate() error = %v", err)
	}
	for _, config := range []Config{
		{},
		{LeaseDuration: MinLeaseDuration - 1, RetentionDuration: MinRetentionDuration},
		{LeaseDuration: MinLeaseDuration, RetentionDuration: MinRetentionDuration - 1},
	} {
		if err := config.validate(); !errors.Is(err, ErrInvalidConfig) {
			t.Fatalf("Config.validate() error = %v, want ErrInvalidConfig", err)
		}
	}
}
