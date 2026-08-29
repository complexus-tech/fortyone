package gitlab

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/codehost"
)

const defaultRequestTimeout = 15 * time.Second

type TokenKind string

const (
	TokenOAuthBearer TokenKind = "oauth_bearer"  // #nosec G101 -- Provider token-kind enum, not a credential value.
	TokenPrivate     TokenKind = "private_token" // #nosec G101 -- Provider token-kind enum, not a credential value.
)

// AccessToken is short-lived adapter input. Callers should retrieve it from
// the credential vault at request time and must never log or persist Value.
type AccessToken struct {
	Kind  TokenKind
	Value string
}

func (AccessToken) String() string { return "[REDACTED]" }

func (AccessToken) GoString() string { return "[REDACTED]" }

func (AccessToken) LogValue() slog.Value { return slog.StringValue("[REDACTED]") }

type TokenSource interface {
	AccessToken(ctx context.Context, installation codehost.InstallationRef) (AccessToken, error)
}

type TokenSourceFunc func(ctx context.Context, installation codehost.InstallationRef) (AccessToken, error)

func (source TokenSourceFunc) AccessToken(ctx context.Context, installation codehost.InstallationRef) (AccessToken, error) {
	return source(ctx, installation)
}

type Config struct {
	BaseURL              string
	HTTPClient           *http.Client
	Tokens               TokenSource
	WebhookSigningToken  string
	WebhookPayloadSecret string
	WebhookReplayWindow  time.Duration
	Now                  func() time.Time
}

type Adapter struct {
	baseURL *url.URL
	client  *http.Client
	tokens  TokenSource
}

func NewAdapter(config Config) (*Adapter, error) {
	baseURL, err := normalizeAPIBaseURL(config.BaseURL)
	if err != nil {
		return nil, err
	}
	if config.Tokens == nil {
		return nil, errors.Join(codehost.ErrAuthentication, errors.New("GitLab token source is required"))
	}
	client := cloneHTTPClient(config.HTTPClient)
	if client.Timeout <= 0 {
		client.Timeout = defaultRequestTimeout
	}
	if client.Timeout > 30*time.Second {
		return nil, errors.Join(codehost.ErrInvalidInput, errors.New("GitLab request timeout exceeds 30 seconds"))
	}
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return errors.New("GitLab API redirects are disabled")
	}
	return &Adapter{baseURL: baseURL, client: client, tokens: config.Tokens}, nil
}

func normalizeAPIBaseURL(raw string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil ||
		parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, errors.Join(codehost.ErrInvalidInput, errors.New("GitLab API base URL must be an HTTPS origin without credentials, query, or fragment"))
	}
	parsed.Path = strings.TrimSuffix(path.Clean("/"+strings.TrimSpace(parsed.Path)), "/")
	if !strings.HasSuffix(parsed.Path, "/api/v4") {
		parsed.Path = strings.TrimSuffix(parsed.Path, "/") + "/api/v4"
	}
	parsed.RawPath = ""
	return parsed, nil
}

func normalizeInstanceURL(raw string) (string, error) {
	apiURL, err := normalizeAPIBaseURL(raw)
	if err != nil {
		return "", err
	}
	apiURL.Path = strings.TrimSuffix(apiURL.Path, "/api/v4")
	return strings.TrimSuffix(apiURL.String(), "/"), nil
}

func cloneHTTPClient(client *http.Client) *http.Client {
	if client == nil {
		return &http.Client{}
	}
	cloned := *client
	return &cloned
}
