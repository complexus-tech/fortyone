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

	accepted, err := service.billingRedirect("https://app.fortyone.example/settings/billing?tab=plan", "workspace")
	if err != nil || accepted.Host != origin.Host {
		t.Fatalf("accepted redirect = %v, %v", accepted, err)
	}
	checkoutRedirect, err := service.checkoutSuccessRedirect(accepted.String(), "workspace")
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
		if _, err := service.billingRedirect(candidate, "workspace"); !errors.Is(err, ErrInvalidBillingRedirect) {
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

func TestBillingRedirectAllowsOnlyAuthenticatedHostedWorkspace(t *testing.T) {
	for _, originURL := range []string{"https://fortyone.app", "https://www.fortyone.app", "https://app.fortyone.app"} {
		t.Run(originURL, func(t *testing.T) {
			origin, err := parseBillingOrigin(originURL)
			if err != nil {
				t.Fatal(err)
			}
			service := &Service{redirectOrigin: origin}
			for _, path := range []string{"/my-work", "/settings/workspace/billing"} {
				candidate := "https://complexus.fortyone.app" + path
				if _, err := service.billingRedirect(candidate, "complexus"); err != nil {
					t.Fatal(err)
				}
				success, err := service.checkoutSuccessRedirect(candidate, "complexus")
				if err != nil || !strings.Contains(success, "session_id={CHECKOUT_SESSION_ID}") {
					t.Fatalf("success=%s error=%v", success, err)
				}
			}
			for _, candidate := range []string{
				"https://other.fortyone.app/settings/workspace/billing",
				"https://complexus.fortyone.app.attacker.example/",
				"https://complexus.fortyone.app@attacker.example/",
				"https://complexus.fortyone.app:444/",
				"http://complexus.fortyone.app/",
				"https://complexus.fortyone.app/#fragment",
			} {
				if _, err := service.billingRedirect(candidate, "complexus"); !errors.Is(err, ErrInvalidBillingRedirect) {
					t.Errorf("accepted %s", candidate)
				}
			}
		})
	}
}

func TestBillingRedirectDoesNotExpandLocalOrCustomOrigins(t *testing.T) {
	for _, originURL := range []string{"http://localhost:3000", "https://custom.example", "https://fortyone.app.attacker.example"} {
		origin, err := parseBillingOrigin(originURL)
		if err != nil {
			t.Fatal(err)
		}
		service := &Service{redirectOrigin: origin}
		if _, err := service.billingRedirect("https://complexus.fortyone.app/", "complexus"); !errors.Is(err, ErrInvalidBillingRedirect) {
			t.Errorf("origin %s expanded unexpectedly", originURL)
		}
	}
}
