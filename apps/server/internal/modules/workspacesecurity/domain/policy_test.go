package workspacesecuritydomain

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNormalizePolicyExactDomains(t *testing.T) {
	policy, err := NormalizePolicy(PolicyUpdate{AllowedDomains: []string{" EXAMPLE.COM ", "bücher.example", "example.com"}, AllowGuests: true, MaxSessionAgeHours: 24})
	if err != nil || !reflect.DeepEqual(policy.AllowedDomains, []string{"example.com", "xn--bcher-kva.example"}) {
		t.Fatalf("normalized = %#v, %v", policy, err)
	}
	for _, domain := range []string{"*.example.com", "https://example.com", "user@example.com", "example", "example.com.", "a..example.com", "-a.example.com", "x_example.com"} {
		if _, err := NormalizePolicy(PolicyUpdate{AllowedDomains: []string{domain}}); !errors.Is(err, ErrInvalid) {
			t.Errorf("domain %q accepted: %v", domain, err)
		}
	}
	if !EmailAllowed("person@EXAMPLE.COM", policy.AllowedDomains) || EmailAllowed("person@sub.example.com", policy.AllowedDomains) || EmailAllowed("Person <person@example.com>", policy.AllowedDomains) {
		t.Fatal("email policy did not require exact account domain")
	}
}
func TestTenantSessionPolicyPreservesAuthenticationAge(t *testing.T) {
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	initial := SessionIdentity{ID: uuid.New(), AuthenticatedAt: now.Add(-2 * time.Hour), ExpiresAt: now.Add(24 * time.Hour)}
	state := AccessState{Email: "person@example.com", Role: "member", Policy: Policy{AllowGuests: true}}
	if err := CheckAccess(state, initial, now); err != nil {
		t.Fatal(err)
	}
	state.Policy.MaxSessionAgeHours = 1
	initial.ExpiresAt = now.Add(90 * 24 * time.Hour)
	if err := CheckAccess(state, initial, now); !errors.Is(err, ErrForbidden) {
		t.Fatal("cookie renewal reset true authentication age")
	}
	initial.AuthenticatedAt = now.Add(-30 * time.Minute)
	if err := CheckAccess(state, initial, now); err != nil {
		t.Fatal(err)
	}
	cutoff := now.Add(-20 * time.Minute)
	state.RevokedBefore = &cutoff
	if err := CheckAccess(state, initial, now); !errors.Is(err, ErrForbidden) {
		t.Fatal("tenant epoch did not revoke prior authentication")
	}
	initial.AuthenticatedAt = now.Add(-10 * time.Minute)
	if err := CheckAccess(state, initial, now); err != nil {
		t.Fatal(err)
	}
	state.RevokedAt = &cutoff
	if err := CheckAccess(state, initial, now); !errors.Is(err, ErrForbidden) {
		t.Fatal("single session revocation ignored")
	}
}
func TestLegacySessionRequiresLoginOnlyWhenPolicyNeedsTimestamp(t *testing.T) {
	state := AccessState{Email: "person@example.com", Role: "member", Policy: Policy{AllowGuests: true}}
	if err := CheckAccess(state, SessionIdentity{}, time.Now()); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	state.RevokedBefore = &now
	if err := CheckAccess(state, SessionIdentity{}, now); !errors.Is(err, ErrForbidden) {
		t.Fatal("legacy session bypassed revocation")
	}
	state.RevokedBefore = nil
	state.Policy.MaxSessionAgeHours = 24
	if err := CheckAccess(state, SessionIdentity{}, now); !errors.Is(err, ErrForbidden) {
		t.Fatal("legacy session bypassed max age")
	}
	state.Policy.MaxSessionAgeHours = 0
	state.Policy.AllowGuests = false
	state.Role = "guest"
	if err := CheckAccess(state, SessionIdentity{}, now); !errors.Is(err, ErrForbidden) {
		t.Fatal("guest policy bypassed")
	}
}
