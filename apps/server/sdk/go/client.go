package fortyone

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultHTTPTimeout = 30 * time.Second

// Config configures a contract-generated FortyOne API client. Token is never
// added to URLs or errors. BaseURL defaults to the production URL in the
// committed OpenAPI document.
type Config struct {
	Token                 string
	BaseURL               string
	HTTPClient            *http.Client
	Retry                 RetryPolicy
	AllowInsecureLoopback bool
}

func (config Config) String() string {
	return fmt.Sprintf("{Token:[REDACTED] BaseURL:%q Retry:%v AllowInsecureLoopback:%t}", config.BaseURL, config.Retry, config.AllowInsecureLoopback)
}

func (config Config) GoString() string { return config.String() }

// New creates a generated response-aware client with bearer authentication,
// bounded safe-read retries, and a redirect-denying HTTP client.
func New(config Config) (*ClientWithResponses, error) {
	token, err := validateToken(config.Token)
	if err != nil {
		return nil, err
	}
	baseURL, err := normalizeBaseURL(config.BaseURL, config.AllowInsecureLoopback)
	if err != nil {
		return nil, err
	}
	policy, err := config.Retry.normalized()
	if err != nil {
		return nil, err
	}

	httpClient := cloneHTTPClient(config.HTTPClient)
	httpClient.Transport = newRetryTransport(httpClient.Transport, policy)
	client, err := NewClientWithResponses(
		baseURL,
		WithHTTPClient(httpClient),
		WithRequestEditorFn(func(_ context.Context, request *http.Request) error {
			request.Header.Set("Authorization", "Bearer "+token)
			if request.Header.Get("Accept") == "" {
				request.Header.Set("Accept", "application/json")
			}
			return nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("create FortyOne API client: %w", err)
	}
	return client, nil
}

func validateToken(token string) (string, error) {
	if token == "" || token != strings.TrimSpace(token) {
		return "", errors.New("FortyOne API token is missing or malformed")
	}
	for _, character := range token {
		if character <= 0x20 || character >= 0x7f {
			return "", errors.New("FortyOne API token is missing or malformed")
		}
	}
	return token, nil
}

func normalizeBaseURL(raw string, allowInsecureLoopback bool) (string, error) {
	if raw == "" {
		raw = DefaultBaseURL
	}
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || !parsed.IsAbs() || parsed.Host == "" {
		return "", errors.New("FortyOne API base URL must be an absolute URL")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("FortyOne API base URL must not include credentials, a query, or a fragment")
	}
	if parsed.Scheme != "https" && !(allowInsecureLoopback && parsed.Scheme == "http" && isLoopback(parsed.Hostname())) {
		return "", errors.New("FortyOne API base URL must use HTTPS (HTTP is available only for explicitly enabled loopback tests)")
	}
	return strings.TrimRight(parsed.String(), "/"), nil
}

func isLoopback(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func cloneHTTPClient(provided *http.Client) *http.Client {
	if provided == nil {
		return &http.Client{
			Timeout: defaultHTTPTimeout,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}
	clone := *provided
	if clone.Timeout == 0 {
		clone.Timeout = defaultHTTPTimeout
	}
	clone.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &clone
}
