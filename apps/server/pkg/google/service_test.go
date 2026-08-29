package google

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/oauth2"
)

func TestCalendarTokenScopesUsesGrantedScopes(t *testing.T) {
	t.Parallel()

	token := (&oauth2.Token{}).WithExtra(map[string]any{
		"scope": "openid email https://www.googleapis.com/auth/calendar.events.readonly",
	})

	scopes := calendarTokenScopes(token, []string{scopeCalendarFreeBusy})

	if len(scopes) != 3 {
		t.Fatalf("unexpected scope count: %#v", scopes)
	}
	if !hasAnyScope(scopes, scopeCalendarEventsReadonly) {
		t.Fatalf("expected granted calendar event read scope: %#v", scopes)
	}
}

func TestCalendarTokenScopesFallsBackWhenGoogleOmitsScopeField(t *testing.T) {
	t.Parallel()

	scopes := calendarTokenScopes(&oauth2.Token{}, []string{scopeCalendarFreeBusy})

	if len(scopes) != 1 || scopes[0] != scopeCalendarFreeBusy {
		t.Fatalf("unexpected fallback scopes: %#v", scopes)
	}
}

func TestCalendarIdentityFromTokenRejectsMissingIDToken(t *testing.T) {
	t.Parallel()

	service := &Service{}
	_, err := service.calendarIdentityFromToken(context.Background(), &oauth2.Token{})
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected invalid token, got %v", err)
	}
}

func TestCalendarOAuthRequestsOwnedEventWriteScope(t *testing.T) {
	t.Parallel()
	service, err := NewService(Config{
		ClientIDs: []string{"client-id"}, ClientSecret: "client-secret",
		RedirectURL: "https://example.com/auth/callback", CalendarRedirectURL: "https://example.com/calendar/callback",
	})
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}
	if service.calendarOAuth == nil || !hasAnyScope(service.calendarOAuth.Scopes, scopeCalendarEventsOwned) {
		t.Fatalf("expected calendar OAuth to request owned-event write access: %#v", service.calendarOAuth)
	}
}
