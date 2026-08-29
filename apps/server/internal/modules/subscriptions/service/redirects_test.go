package subscriptions

import (
	"errors"
	"strings"
	"testing"
)

func TestBillingRedirectRequiresConfiguredOrigin(t *testing.T) {
	t.Parallel()
	origin, err := parseBillingOrigin("https://app.fortyone.example/base")
	if err != nil {
		t.Fatalf("parse origin: %v", err)
	}
	service := &Service{redirectOrigin: origin}

	accepted, err := service.billingRedirect("https://app.fortyone.example/settings/billing?tab=plan")
	if err != nil || accepted.Host != origin.Host {
		t.Fatalf("accepted redirect = %v, %v", accepted, err)
	}
	checkoutRedirect, err := service.checkoutSuccessRedirect(accepted.String())
	if err != nil {
		t.Fatalf("checkout redirect: %v", err)
	}
	if !strings.Contains(checkoutRedirect, "session_id={CHECKOUT_SESSION_ID}") || strings.Contains(checkoutRedirect, "%7B") {
		t.Fatalf("checkout placeholder was not preserved literally: %s", checkoutRedirect)
	}
	for _, candidate := range []string{
		"https://attacker.example/callback",
		"https://app.fortyone.example@attacker.example/callback",
		"javascript:alert(1)",
		"https://app.fortyone.example/callback#fragment",
	} {
		if _, err := service.billingRedirect(candidate); !errors.Is(err, ErrInvalidBillingRedirect) {
			t.Errorf("redirect %q error = %v", candidate, err)
		}
	}
}

func TestSupportedPaidLookupKeyIsAnExplicitCatalog(t *testing.T) {
	t.Parallel()
	if !supportedPaidLookupKey("pro_monthly") {
		t.Fatal("pro_monthly unexpectedly rejected")
	}
	if supportedPaidLookupKey("internal_unpublished_price") || supportedPaidLookupKey("free") {
		t.Fatal("unsupported lookup key unexpectedly accepted")
	}
}
