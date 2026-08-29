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
	"sync"
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

func TestVerifyIDTokenEnforcesOIDCSecurityClaims(t *testing.T) {
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

	tests := []struct {
		name      string
		mutate    func(*idTokenClaims)
		method    jwt.SigningMethod
		key       any
		keyID     string
		omitKeyID bool
		header    func(map[string]any)
		wantValid bool
	}{
		{
			name:      "valid token",
			wantValid: true,
		},
		{
			name: "wrong audience",
			mutate: func(claims *idTokenClaims) {
				claims.Audience = jwt.ClaimStrings{"another-client"}
			},
		},
		{
			name: "wrong issuer",
			mutate: func(claims *idTokenClaims) {
				claims.Issuer = "https://issuer.example/" + testTenantID
			},
		},
		{
			name: "expired",
			mutate: func(claims *idTokenClaims) {
				claims.ExpiresAt = jwt.NewNumericDate(now.Add(-time.Minute))
			},
		},
		{
			name: "missing expiration",
			mutate: func(claims *idTokenClaims) {
				claims.ExpiresAt = nil
			},
		},
		{
			name: "future issued at",
			mutate: func(claims *idTokenClaims) {
				claims.IssuedAt = jwt.NewNumericDate(now.Add(2 * time.Minute))
			},
		},
		{
			name: "missing issued at",
			mutate: func(claims *idTokenClaims) {
				claims.IssuedAt = nil
			},
		},
		{
			name: "missing subject",
			mutate: func(claims *idTokenClaims) {
				claims.Subject = ""
			},
		},
		{
			name: "invalid tenant identifier",
			mutate: func(claims *idTokenClaims) {
				claims.TenantID = "not-a-tenant-id"
				claims.Issuer = "https://login.microsoftonline.com/not-a-tenant-id/v2.0"
			},
		},
		{
			name: "invalid object identifier",
			mutate: func(claims *idTokenClaims) {
				claims.ObjectID = "not-an-object-id"
			},
		},
		{
			name:      "missing key id",
			omitKeyID: true,
		},
		{
			name: "missing token type",
			header: func(header map[string]any) {
				delete(header, "typ")
			},
		},
		{
			name: "wrong token type",
			header: func(header map[string]any) {
				header["typ"] = "at+jwt"
			},
		},
		{
			name:  "unknown key id",
			keyID: "unknown-key",
		},
		{
			name:   "HMAC algorithm confusion",
			method: jwt.SigningMethodHS256,
			key:    []byte("attacker-controlled-secret"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := validIDTokenClaims(now)
			if tt.mutate != nil {
				tt.mutate(&claims)
			}
			method := tt.method
			if method == nil {
				method = jwt.SigningMethodRS256
			}
			key := tt.key
			if key == nil {
				key = privateKey
			}
			keyID := tt.keyID
			if keyID == "" && !tt.omitKeyID {
				keyID = testKeyID
			}
			rawToken := signClaims(t, method, key, keyID, claims, tt.header)

			identity, err := service.verifyIDToken(context.Background(), rawToken, testNonce)
			if tt.wantValid {
				if err != nil {
					t.Fatalf("verifyIDToken returned an error: %v", err)
				}
				if identity.ObjectID != testObjectID || identity.TenantID != testTenantID {
					t.Fatalf("identity = %+v, want tenant %q and object %q", identity, testTenantID, testObjectID)
				}
				return
			}
			if !errors.Is(err, ErrInvalidToken) {
				t.Fatalf("verifyIDToken error = %v, want ErrInvalidToken", err)
			}
		})
	}
}

func TestVerifyIDTokenFiltersUnsuitableJWKSKeys(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate signing key: %v", err)
	}
	now := time.Date(2026, time.August, 21, 12, 0, 0, 0, time.UTC)
	rawToken := signIDToken(t, privateKey, now, testNonce, testTenantID, testObjectID)

	tests := []struct {
		name      string
		configure func(*jwk)
		wantValid bool
	}{
		{
			name: "signature key using RS256",
			configure: func(key *jwk) {
				key.Use = "sig"
				key.Algorithm = "RS256"
			},
			wantValid: true,
		},
		{
			name: "optional metadata omitted",
			configure: func(key *jwk) {
				key.Use = ""
				key.Algorithm = ""
			},
			wantValid: true,
		},
		{
			name: "encryption key",
			configure: func(key *jwk) {
				key.Use = "enc"
				key.Algorithm = "RS256"
			},
		},
		{
			name: "different signing algorithm",
			configure: func(key *jwk) {
				key.Use = "sig"
				key.Algorithm = "RS512"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := publicJWK(&privateKey.PublicKey)
			tt.configure(&key)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(t, w, jwksDocument{Keys: []jwk{key}})
			}))
			defer server.Close()

			service := NewService(Config{ClientID: testClientID, JWKSURL: server.URL, HTTPClient: server.Client()})
			service.now = func() time.Time { return now }
			_, err := service.verifyIDToken(context.Background(), rawToken, testNonce)
			if tt.wantValid {
				if err != nil {
					t.Fatalf("verifyIDToken returned an error: %v", err)
				}
				return
			}
			if !errors.Is(err, ErrInvalidToken) {
				t.Fatalf("verifyIDToken error = %v, want ErrInvalidToken", err)
			}
		})
	}
}

func TestVerifyIDTokenRefreshesJWKSForRotatedKeyID(t *testing.T) {
	oldPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate old signing key: %v", err)
	}
	newPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate new signing key: %v", err)
	}
	now := time.Date(2026, time.August, 21, 12, 0, 0, 0, time.UTC)

	oldJWK := publicJWK(&oldPrivateKey.PublicKey)
	oldJWK.KeyID = "old-key"
	newJWK := publicJWK(&newPrivateKey.PublicKey)
	newJWK.KeyID = "new-key"
	currentKey := oldJWK
	var keyMu sync.RWMutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		keyMu.RLock()
		key := currentKey
		keyMu.RUnlock()
		w.Header().Set("Cache-Control", "public, max-age=3600")
		writeJSON(t, w, jwksDocument{Keys: []jwk{key}})
	}))
	defer server.Close()

	service := NewService(Config{ClientID: testClientID, JWKSURL: server.URL, HTTPClient: server.Client()})
	service.now = func() time.Time { return now }
	oldToken := signClaims(t, jwt.SigningMethodRS256, oldPrivateKey, oldJWK.KeyID, validIDTokenClaims(now))
	if _, err := service.verifyIDToken(context.Background(), oldToken, testNonce); err != nil {
		t.Fatalf("verify old signing key: %v", err)
	}

	keyMu.Lock()
	currentKey = newJWK
	keyMu.Unlock()
	newToken := signClaims(t, jwt.SigningMethodRS256, newPrivateKey, newJWK.KeyID, validIDTokenClaims(now))
	if _, err := service.verifyIDToken(context.Background(), newToken, testNonce); err != nil {
		t.Fatalf("verify rotated signing key: %v", err)
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
	claims := validIDTokenClaims(now)
	claims.Nonce = nonce
	claims.TenantID = tenantID
	claims.ObjectID = objectID
	claims.Issuer = "https://login.microsoftonline.com/" + tenantID + "/v2.0"
	return signClaims(t, jwt.SigningMethodRS256, privateKey, testKeyID, claims)
}

func validIDTokenClaims(now time.Time) idTokenClaims {
	return idTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "https://login.microsoftonline.com/" + testTenantID + "/v2.0",
			Subject:   "pairwise-subject",
			Audience:  jwt.ClaimStrings{testClientID},
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now.Add(-time.Minute)),
		},
		Nonce:             testNonce,
		TenantID:          testTenantID,
		ObjectID:          testObjectID,
		PreferredUsername: "ada@example.com",
	}
}

func signClaims(
	t *testing.T,
	method jwt.SigningMethod,
	key any,
	keyID string,
	claims idTokenClaims,
	configureHeader ...func(map[string]any),
) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Method = method
	token.Header["alg"] = method.Alg()
	if keyID != "" {
		token.Header["kid"] = keyID
	}
	for _, configure := range configureHeader {
		if configure != nil {
			configure(token.Header)
		}
	}
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("sign id token: %v", err)
	}
	return signed
}

func publicJWK(publicKey *rsa.PublicKey) jwk {
	exponent := big.NewInt(int64(publicKey.E)).Bytes()
	return jwk{
		KeyType:   "RSA",
		KeyID:     testKeyID,
		Use:       "sig",
		Algorithm: "RS256",
		N:         base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
		E:         base64.RawURLEncoding.EncodeToString(exponent),
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
