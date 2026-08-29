package testkit

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	// ProviderBodyLimitBytes bounds request capture and stub responses. Provider
	// tests that need larger payloads should exercise a streaming-specific test
	// double beside that adapter rather than weakening the shared default.
	ProviderBodyLimitBytes = 1 << 20
	providerClientTimeout  = 5 * time.Second
)

var (
	ErrSigningHeaderInvalid    = errors.New("HMAC signature header is invalid")
	ErrSigningSecretRequired   = errors.New("HMAC signing secret is required")
	ErrSignedRequestRequired   = errors.New("request to sign is required")
	ErrProviderBodyTooLarge    = errors.New("provider test body exceeds the shared limit")
	ErrProviderContextRequired = errors.New("provider request context is required")
)

// HMACSHA256Signer signs provider fixtures without exposing its key through
// formatted diagnostics. Prefix supports provider wire formats such as
// "sha256=" or "v0="; callers may sign a canonical payload separately from
// the request body when a provider protocol requires it.
type HMACSHA256Signer struct {
	header string
	prefix string
	secret []byte
}

// NewHMACSHA256Signer constructs a reusable test signer. The secret is copied
// so later fixture mutation cannot change signatures unexpectedly.
func NewHMACSHA256Signer(header string, prefix string, secret []byte) (HMACSHA256Signer, error) {
	header = strings.TrimSpace(header)
	if !isHTTPToken(header) {
		return HMACSHA256Signer{}, ErrSigningHeaderInvalid
	}
	if len(secret) == 0 {
		return HMACSHA256Signer{}, ErrSigningSecretRequired
	}
	return HMACSHA256Signer{
		header: http.CanonicalHeaderKey(header),
		prefix: prefix,
		secret: append([]byte(nil), secret...),
	}, nil
}

// Signature returns a lowercase hexadecimal HMAC-SHA-256 signature in the
// configured provider wire format.
func (s HMACSHA256Signer) Signature(payload []byte) (string, error) {
	if !isHTTPToken(s.header) {
		return "", ErrSigningHeaderInvalid
	}
	if len(s.secret) == 0 {
		return "", ErrSigningSecretRequired
	}
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write(payload)
	return s.prefix + hex.EncodeToString(mac.Sum(nil)), nil
}

// Sign sets the configured signature header using payload as the canonical
// signed bytes. It never includes the secret or payload in returned errors.
func (s HMACSHA256Signer) Sign(request *http.Request, payload []byte) error {
	if request == nil {
		return ErrSignedRequestRequired
	}
	signature, err := s.Signature(payload)
	if err != nil {
		return err
	}
	if request.Header == nil {
		request.Header = make(http.Header)
	}
	request.Header.Set(s.header, signature)
	return nil
}

// String deliberately redacts the signing key in normal diagnostics.
func (s HMACSHA256Signer) String() string {
	return fmt.Sprintf("HMACSHA256Signer{header:%q, prefix:%q, secret:<redacted>}", s.header, s.prefix)
}

// GoString deliberately redacts the signing key in %#v diagnostics.
func (s HMACSHA256Signer) GoString() string {
	return s.String()
}

// NewSignedProviderRequest creates a bounded request and signs the exact body
// bytes. URL parse failures are intentionally returned without the target URL
// because provider URLs can contain credentials in userinfo or query values.
func NewSignedProviderRequest(
	ctx context.Context,
	method string,
	target string,
	body []byte,
	signer HMACSHA256Signer,
) (*http.Request, error) {
	if ctx == nil {
		return nil, ErrProviderContextRequired
	}
	if len(body) > ProviderBodyLimitBytes {
		return nil, ErrProviderBodyTooLarge
	}
	bodyCopy := append([]byte(nil), body...)
	request, err := http.NewRequestWithContext(ctx, method, target, bytes.NewReader(bodyCopy))
	if err != nil {
		return nil, errors.New("create signed provider request: invalid method or target URL")
	}
	if err := signer.Sign(request, bodyCopy); err != nil {
		return nil, fmt.Errorf("sign provider request: %w", err)
	}
	return request, nil
}

// ProviderRequest is a bounded request captured by ProviderServer. Its String
// methods intentionally omit headers, query values, and body contents so a
// failed assertion cannot accidentally print bearer tokens or payloads.
type ProviderRequest struct {
	Method string
	Path   string
	Query  url.Values
	Header http.Header
	Body   []byte
}

func (r ProviderRequest) String() string {
	return fmt.Sprintf(
		"ProviderRequest{method:%q, path_bytes:%d, query_keys:%d, headers:%d, body_bytes:%d}",
		r.Method,
		len(r.Path),
		len(r.Query),
		len(r.Header),
		len(r.Body),
	)
}

func (r ProviderRequest) GoString() string {
	return r.String()
}

// ProviderResponse is the response emitted by a ProviderServer responder.
// StatusCode zero defaults to 200 OK.
type ProviderResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// ProviderResponder maps a captured, bounded request to a bounded response.
type ProviderResponder func(ProviderRequest) ProviderResponse

// ProviderServer is a concurrency-safe httptest provider with bounded capture,
// defensive snapshots, automatic cleanup, and a timeout-bound HTTP client.
type ProviderServer struct {
	server *httptest.Server
	client *http.Client

	mu       sync.RWMutex
	requests []ProviderRequest
}

// NewProviderServer starts a provider stub and registers cleanup with t. A nil
// responder captures requests and returns 204 No Content.
func NewProviderServer(t testing.TB, responder ProviderResponder) *ProviderServer {
	t.Helper()

	provider := &ProviderServer{}
	provider.server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		captured, ok := captureProviderRequest(writer, request)
		if !ok {
			return
		}
		provider.mu.Lock()
		provider.requests = append(provider.requests, captured)
		provider.mu.Unlock()

		if responder == nil {
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		writeProviderResponse(writer, responder(cloneProviderRequest(captured)))
	}))
	client := provider.server.Client()
	client.Timeout = providerClientTimeout
	provider.client = client
	t.Cleanup(func() {
		provider.server.CloseClientConnections()
		provider.server.Close()
	})
	return provider
}

// URL returns the provider base URL.
func (s *ProviderServer) URL() string {
	return s.server.URL
}

// Client returns a shallow copy of the timeout-bound provider client. Its
// transport remains owned by the server and is closed through test cleanup.
func (s *ProviderServer) Client() *http.Client {
	client := *s.client
	return &client
}

// Requests returns a deep snapshot that callers may inspect or mutate without
// racing with the provider or changing recorded evidence.
func (s *ProviderServer) Requests() []ProviderRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	requests := make([]ProviderRequest, len(s.requests))
	for index, request := range s.requests {
		requests[index] = cloneProviderRequest(request)
	}
	return requests
}

func captureProviderRequest(writer http.ResponseWriter, request *http.Request) (ProviderRequest, bool) {
	body, err := io.ReadAll(io.LimitReader(request.Body, ProviderBodyLimitBytes+1))
	if err != nil {
		http.Error(writer, "provider test server could not read request", http.StatusBadRequest)
		return ProviderRequest{}, false
	}
	if len(body) > ProviderBodyLimitBytes {
		http.Error(writer, "provider test request body is too large", http.StatusRequestEntityTooLarge)
		return ProviderRequest{}, false
	}
	return ProviderRequest{
		Method: request.Method,
		Path:   request.URL.EscapedPath(),
		Query:  request.URL.Query(),
		Header: request.Header.Clone(),
		Body:   body,
	}, true
}

func writeProviderResponse(writer http.ResponseWriter, response ProviderResponse) {
	statusCode := response.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	if statusCode < 200 || statusCode > 599 || len(response.Body) > ProviderBodyLimitBytes || !validProviderHeaders(response.Header) {
		http.Error(writer, "provider test responder returned an invalid response", http.StatusInternalServerError)
		return
	}
	for key, values := range response.Header {
		for _, value := range values {
			writer.Header().Add(key, value)
		}
	}
	writer.WriteHeader(statusCode)
	_, _ = writer.Write(response.Body)
}

func validProviderHeaders(headers http.Header) bool {
	for key, values := range headers {
		if !isHTTPToken(key) {
			return false
		}
		for _, value := range values {
			if strings.ContainsAny(value, "\r\n") {
				return false
			}
		}
	}
	return true
}

func cloneProviderRequest(request ProviderRequest) ProviderRequest {
	return ProviderRequest{
		Method: request.Method,
		Path:   request.Path,
		Query:  cloneURLValues(request.Query),
		Header: request.Header.Clone(),
		Body:   append([]byte(nil), request.Body...),
	}
}

func cloneURLValues(values url.Values) url.Values {
	clone := make(url.Values, len(values))
	for key, entries := range values {
		clone[key] = append([]string(nil), entries...)
	}
	return clone
}

func isHTTPToken(value string) bool {
	if value == "" {
		return false
	}
	for index := range len(value) {
		character := value[index]
		if (character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') ||
			strings.ContainsRune("!#$%&'*+-.^_`|~", rune(character)) {
			continue
		}
		return false
	}
	return true
}
