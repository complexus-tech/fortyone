package oidcclient

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"golang.org/x/oauth2"
)

func TestOIDCCodeFlowValidatesSignedIdentityAndFreshAuthentication(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key}, (&jose.SignerOptions{}).WithHeader("kid", "test-key"))
	if err != nil {
		t.Fatal(err)
	}
	var issuer string
	var overrides map[string]any
	verifier := oauth2.GenerateVerifier()
	nonce := "test-private-nonce"
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			_ = json.NewEncoder(w).Encode(map[string]any{"issuer": issuer, "authorization_endpoint": issuer + "/authorize", "token_endpoint": issuer + "/token", "jwks_uri": issuer + "/keys", "id_token_signing_alg_values_supported": []string{"RS256"}})
		case "/keys":
			_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &key.PublicKey, KeyID: "test-key", Algorithm: "RS256", Use: "sig"}}})
		case "/token":
			if err := r.ParseForm(); err != nil || r.Form.Get("code_verifier") != verifier || r.Form.Get("code") != "valid-code" {
				http.Error(w, "bad code proof", 400)
				return
			}
			now := time.Now().Unix()
			claims := map[string]any{"iss": issuer, "aud": "client-id", "sub": "stable-subject", "email": "person@example.com", "email_verified": true, "nonce": nonce, "auth_time": now, "iat": now, "exp": now + 3600}
			for key, value := range overrides {
				claims[key] = value
			}
			payload, _ := json.Marshal(claims)
			token, err := signer.Sign(payload)
			if err != nil {
				t.Error(err)
				return
			}
			compact, err := token.CompactSerialize()
			if err != nil {
				t.Error(err)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "private-provider-access-token", "token_type": "Bearer", "id_token": compact})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	issuer = server.URL
	client := &Client{http: server.Client(), validate: func(context.Context, string) error { return nil }}
	credential := Credential{Issuer: issuer, ClientID: "client-id", ClientSecret: "private-client-secret"}
	attempt := Attempt{Nonce: nonce, Verifier: verifier, CreatedAt: time.Now().UTC()}
	redirect := "https://app.example.com/auth/sso/callback"
	authorize, err := client.AuthorizationURL(t.Context(), credential, attempt, "private-state", redirect)
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ := url.Parse(authorize)
	query := parsed.Query()
	if query.Get("code_challenge") != oauth2.S256ChallengeFromVerifier(verifier) || query.Get("code_challenge_method") != "S256" || query.Get("nonce") != nonce || query.Get("prompt") != "login" || query.Get("max_age") != "0" || query.Get("redirect_uri") != redirect {
		t.Fatal("authorization omitted PKCE, nonce, or fresh login request")
	}
	identity, err := client.Exchange(t.Context(), credential, attempt, "valid-code", redirect)
	if err != nil || identity.Subject != "stable-subject" || identity.Email != "person@example.com" || identity.AuthenticatedAt.IsZero() {
		t.Fatalf("verified identity %#v %v", identity, err)
	}
	for name, claims := range map[string]map[string]any{"nonce": {"nonce": "attacker"}, "audience": {"aud": "attacker"}, "issuer": {"iss": "https://attacker.example"}, "expired": {"exp": time.Now().Add(-time.Hour).Unix()}, "unverified_email": {"email_verified": false}, "stale_authentication": {"auth_time": time.Now().Add(-time.Hour).Unix()}, "missing_authentication": {"auth_time": 0}, "future_authentication": {"auth_time": time.Now().Add(time.Hour).Unix()}, "authorized_party": {"aud": []string{"client-id", "other-client"}, "azp": "other-client"}, "missing_azp": {"aud": []string{"client-id", "other-client"}}, "subject": {"sub": ""}} {
		t.Run(name, func(t *testing.T) {
			overrides = claims
			if _, err := client.Exchange(t.Context(), credential, attempt, "valid-code", redirect); !errors.Is(err, ErrProvider) {
				t.Fatalf("invalid signed claims accepted: %v", err)
			}
		})
	}
}
func TestPublicOIDCTransportRejectsUntrustedDestinations(t *testing.T) {
	for _, target := range []string{"http://provider.example/keys", "https://127.0.0.1/keys", "https://provider.example:8443/keys", "https://username:password@provider.example/keys"} {
		request, _ := http.NewRequest(http.MethodGet, target, nil)
		if _, err := (publicTransport{}).RoundTrip(request); err == nil {
			t.Fatalf("unsafe provider destination %q accepted", target)
		}
	}
}
