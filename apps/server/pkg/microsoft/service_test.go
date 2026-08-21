package microsoft

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	testClientID = "11111111-1111-4111-8111-111111111111"
	testTenantID = "22222222-2222-4222-8222-222222222222"
	testObjectID = "33333333-3333-4333-8333-333333333333"
	testKeyID    = "microsoft-test-key"
	testNonce    = "nonce-value"
)

func TestAuthCodeURLUsesOIDCStateAndPKCE(t *testing.T) {
	service := NewService(Config{
		ClientID:     testClientID,
		ClientSecret: "client-secret",
		RedirectURL:  "http://localhost:8000/auth/microsoft/callback",
		AuthURL:      "https://login.example/authorize",
		TokenURL:     "https://login.example/token",
	})
	verifier := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"

	rawURL, err := service.AuthCodeURL("state-value", testNonce, verifier)
	if err != nil {
		t.Fatalf("AuthCodeURL returned an error: %v", err)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse authorization URL: %v", err)
	}
	query := parsed.Query()

	expectQueryValue(t, query, "state", "state-value")
	expectQueryValue(t, query, "nonce", testNonce)
	expectQueryValue(t, query, "prompt", "select_account")
	expectQueryValue(t, query, "response_mode", "query")
	expectQueryValue(t, query, "code_challenge_method", "S256")

	challenge := sha256.Sum256([]byte(verifier))
	expectQueryValue(t, query, "code_challenge", base64.RawURLEncoding.EncodeToString(challenge[:]))

	scopes := strings.Fields(query.Get("scope"))
	for _, required := range []string{"openid", "profile", "email", "User.Read"} {
		if !contains(scopes, required) {
			t.Errorf("authorization scope is missing %q: %v", required, scopes)
		}
	}
	if contains(scopes, "Calendars.ReadWrite") {
		t.Errorf("sign-in must not request calendar permissions: %v", scopes)
	}
}

func TestExchangeCodeVerifiesIdentityAndFetchesProfile(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate signing key: %v", err)
	}
	now := time.Date(2026, time.August, 21, 12, 0, 0, 0, time.UTC)
	idToken := signIDToken(t, privateKey, now, testNonce, testTenantID, testObjectID)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			if err := r.ParseForm(); err != nil {
				t.Errorf("parse token form: %v", err)
			}
			if got := r.Form.Get("code"); got != "authorization-code" {
				t.Errorf("authorization code = %q, want %q", got, "authorization-code")
			}
			if got := r.Form.Get("code_verifier"); got != "pkce-verifier" {
				t.Errorf("code verifier = %q, want %q", got, "pkce-verifier")
			}
			writeJSON(t, w, map[string]any{
				"access_token": "graph-access-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
				"id_token":     idToken,
			})
		case "/keys":
			w.Header().Set("Cache-Control", "public, max-age=3600")
			writeJSON(t, w, jwksDocument{Keys: []jwk{publicJWK(&privateKey.PublicKey)}})
		case "/me":
			if got := r.Header.Get("Authorization"); got != "Bearer graph-access-token" {
				t.Errorf("authorization header = %q", got)
			}
			if got := r.URL.Query().Get("$select"); !strings.Contains(got, "mail") {
				t.Errorf("profile selection = %q, expected mail field", got)
			}
			writeJSON(t, w, graphProfile{
				ID:          testObjectID,
				DisplayName: "Ada Lovelace",
				GivenName:   "Ada",
				Surname:     "Lovelace",
				Mail:        "ada@example.com",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	service := NewService(Config{
		ClientID:     testClientID,
		ClientSecret: "client-secret",
		RedirectURL:  "http://localhost:8000/auth/microsoft/callback",
		TokenURL:     server.URL + "/token",
		AuthURL:      server.URL + "/authorize",
		ProfileURL:   server.URL + "/me",
		JWKSURL:      server.URL + "/keys",
		HTTPClient:   server.Client(),
	})
	service.now = func() time.Time { return now }

	identity, err := service.ExchangeCode(context.Background(), "authorization-code", "pkce-verifier", testNonce)
	if err != nil {
		t.Fatalf("ExchangeCode returned an error: %v", err)
	}
	if identity.ObjectID != testObjectID {
		t.Errorf("object ID = %q, want %q", identity.ObjectID, testObjectID)
	}
	if identity.Issuer != "https://login.microsoftonline.com/"+testTenantID+"/v2.0" {
		t.Errorf("issuer = %q", identity.Issuer)
	}
	if identity.Email != "ada@example.com" {
		t.Errorf("email = %q, want graph email", identity.Email)
	}
	if identity.FullName != "Ada Lovelace" {
		t.Errorf("full name = %q, want graph display name", identity.FullName)
	}
}

func TestVerifyIDTokenRejectsNonceMismatch(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate signing key: %v", err)
	}
	now := time.Date(2026, time.August, 21, 12, 0, 0, 0, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, jwksDocument{Keys: []jwk{publicJWK(&privateKey.PublicKey)}})
	}))
	defer server.Close()

	service := NewService(Config{ClientID: testClientID, JWKSURL: server.URL, HTTPClient: server.Client()})
	service.now = func() time.Time { return now }
	token := signIDToken(t, privateKey, now, "different-nonce", testTenantID, testObjectID)

	if _, err := service.verifyIDToken(context.Background(), token, testNonce); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("verifyIDToken error = %v, want ErrInvalidToken", err)
	}
}

func TestGraphProfileMatchesIdentity(t *testing.T) {
	tests := []struct {
		name      string
		tenantID  string
		profileID string
		objectID  string
		want      bool
	}{
		{
			name:      "personal account can use legacy graph identifier",
			tenantID:  personalTenantID,
			profileID: "legacy-msa-graph-id",
			objectID:  testObjectID,
			want:      true,
		},
		{
			name:      "organizational account with matching object ID",
			tenantID:  testTenantID,
			profileID: testObjectID,
			objectID:  testObjectID,
			want:      true,
		},
		{
			name:      "organizational account with mismatched object ID",
			tenantID:  testTenantID,
			profileID: "44444444-4444-4444-8444-444444444444",
			objectID:  testObjectID,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			identity := Identity{TenantID: tt.tenantID, ObjectID: tt.objectID}
			if got := graphProfileMatchesIdentity(tt.profileID, identity); got != tt.want {
				t.Errorf("graphProfileMatchesIdentity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func signIDToken(t *testing.T, privateKey *rsa.PrivateKey, now time.Time, nonce, tenantID, objectID string) string {
	t.Helper()
	claims := idTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "https://login.microsoftonline.com/" + tenantID + "/v2.0",
			Subject:   "pairwise-subject",
			Audience:  jwt.ClaimStrings{testClientID},
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now.Add(-time.Minute)),
		},
		Nonce:             nonce,
		TenantID:          tenantID,
		ObjectID:          objectID,
		PreferredUsername: "ada@example.com",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = testKeyID
	signed, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign id token: %v", err)
	}
	return signed
}

func publicJWK(publicKey *rsa.PublicKey) jwk {
	exponent := big.NewInt(int64(publicKey.E)).Bytes()
	return jwk{
		KeyType: "RSA",
		KeyID:   testKeyID,
		Use:     "sig",
		N:       base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
		E:       base64.RawURLEncoding.EncodeToString(exponent),
	}
}

func expectQueryValue(t *testing.T, query url.Values, key, want string) {
	t.Helper()
	if got := query.Get(key); got != want {
		t.Errorf("query %s = %q, want %q", key, got, want)
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Errorf("encode response: %v", err)
	}
}
