package safehttp

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTimeout            = 10 * time.Second
	defaultConnectTimeout     = 3 * time.Second
	defaultTLSHandshake       = 5 * time.Second
	defaultHeaderTimeout      = 5 * time.Second
	defaultMaxResponseBytes   = int64(64 << 10)
	defaultMaxResponseHeaders = int64(16 << 10)
)

type DialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)

type Config struct {
	Resolver               Resolver
	DialContext            DialContextFunc
	Timeout                time.Duration
	TLSHandshakeTimeout    time.Duration
	ResponseHeaderTimeout  time.Duration
	MaxResponseBytes       int64
	MaxResponseHeaderBytes int64
}

type Client struct {
	resolver               Resolver
	dialContext            DialContextFunc
	timeout                time.Duration
	tlsHandshakeTimeout    time.Duration
	responseHeaderTimeout  time.Duration
	maxResponseBytes       int64
	maxResponseHeaderBytes int64
}

type Result struct {
	StatusCode     int
	ResolvedIP     netip.Addr
	Duration       time.Duration
	ResponseBytes  int64
	ResponseDigest [sha256.Size]byte
	RetryAfter     time.Duration
	Truncated      bool
}

func New(config Config) (*Client, error) {
	resolver := config.Resolver
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	dialContext := config.DialContext
	if dialContext == nil {
		dialer := &net.Dialer{Timeout: defaultConnectTimeout, KeepAlive: -1}
		dialContext = dialer.DialContext
	}
	timeout := positiveOrDefault(config.Timeout, defaultTimeout)
	tlsTimeout := positiveOrDefault(config.TLSHandshakeTimeout, defaultTLSHandshake)
	headerTimeout := positiveOrDefault(config.ResponseHeaderTimeout, defaultHeaderTimeout)
	maxBody := positiveInt64OrDefault(config.MaxResponseBytes, defaultMaxResponseBytes)
	maxHeaders := positiveInt64OrDefault(config.MaxResponseHeaderBytes, defaultMaxResponseHeaders)
	if timeout > 30*time.Second || tlsTimeout > timeout || headerTimeout > timeout {
		return nil, fmt.Errorf("%w: unsafe timeout configuration", ErrUnsupportedRequest)
	}
	if maxBody > 1<<20 || maxHeaders > 64<<10 {
		return nil, fmt.Errorf("%w: unsafe response limit configuration", ErrUnsupportedRequest)
	}
	return &Client{
		resolver:               resolver,
		dialContext:            dialContext,
		timeout:                timeout,
		tlsHandshakeTimeout:    tlsTimeout,
		responseHeaderTimeout:  headerTimeout,
		maxResponseBytes:       maxBody,
		maxResponseHeaderBytes: maxHeaders,
	}, nil
}

// Validate resolves an endpoint using the same policy used by Do. It provides
// fast feedback at endpoint creation, but is not a substitute for Do's
// attempt-time resolution and IP-pinned connection.
func (client *Client) Validate(ctx context.Context, rawURL string) (string, error) {
	if client == nil {
		return "", fmt.Errorf("%w: client is required", ErrUnsupportedRequest)
	}
	target, err := Resolve(ctx, client.resolver, rawURL)
	if err != nil {
		return "", err
	}
	return target.URL.String(), nil
}

// Do sends one HTTPS request without proxies, redirects, connection reuse, or
// a second DNS lookup. DNS is re-resolved and fully validated for every call;
// the TCP connection is pinned to the validated address while TLS verifies the
// original hostname. Response content is never retained.
func (client *Client) Do(ctx context.Context, request *http.Request) (Result, error) {
	if client == nil || request == nil || request.URL == nil {
		return Result{}, fmt.Errorf("%w: request and client are required", ErrUnsupportedRequest)
	}
	if request.Method != http.MethodPost {
		return Result{}, fmt.Errorf("%w: only POST is allowed", ErrUnsupportedRequest)
	}
	if request.Body == nil {
		return Result{}, fmt.Errorf("%w: request body is required", ErrUnsupportedRequest)
	}
	target, err := Resolve(ctx, client.resolver, request.URL.String())
	if err != nil {
		return Result{}, err
	}

	transport := &http.Transport{
		Proxy:                  nil,
		DisableKeepAlives:      true,
		ForceAttemptHTTP2:      true,
		TLSHandshakeTimeout:    client.tlsHandshakeTimeout,
		ResponseHeaderTimeout:  client.responseHeaderTimeout,
		MaxResponseHeaderBytes: client.maxResponseHeaderBytes,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: target.Hostname,
		},
		DialContext: func(dialCtx context.Context, network, _ string) (net.Conn, error) {
			if network != "tcp" && network != "tcp4" && network != "tcp6" {
				return nil, fmt.Errorf("%w: network %q", ErrUnsupportedRequest, network)
			}
			return client.dialContext(dialCtx, "tcp", target.dialAddress())
		},
	}
	defer transport.CloseIdleConnections()

	cloned := request.Clone(ctx)
	cloned.URL = target.URL
	cloned.Host = target.Hostname
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   client.timeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return ErrRedirectDenied
		},
	}
	started := time.Now()
	// Resolve validated every address and the transport ignores the URL's dial
	// target in favor of the pinned approved IP while retaining TLS hostname
	// verification. This is the SSRF boundary, not an unvalidated outbound URL.
	response, err := httpClient.Do(cloned) // #nosec G704 -- resolved URL plus IP-pinned dialer prevents DNS rebinding and private-address access.
	if err != nil {
		return Result{ResolvedIP: target.Addresses[0], Duration: time.Since(started)}, fmt.Errorf("safe http request: %w", err)
	}
	defer response.Body.Close()

	hasher := sha256.New()
	read, copyErr := io.CopyN(hasher, response.Body, client.maxResponseBytes+1)
	if copyErr != nil && !errors.Is(copyErr, io.EOF) {
		return Result{StatusCode: response.StatusCode, ResolvedIP: target.Addresses[0], Duration: time.Since(started)}, fmt.Errorf("read safe http response: %w", copyErr)
	}
	result := Result{
		StatusCode:    response.StatusCode,
		ResolvedIP:    target.Addresses[0],
		Duration:      time.Since(started),
		ResponseBytes: read,
		RetryAfter:    parseRetryAfter(response.Header.Get("Retry-After"), time.Now()),
	}
	copy(result.ResponseDigest[:], hasher.Sum(nil))
	if read > client.maxResponseBytes {
		result.Truncated = true
		result.ResponseBytes = client.maxResponseBytes
		return result, ErrResponseTooLarge
	}
	return result, nil
}

func parseRetryAfter(value string, now time.Time) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	if seconds, err := strconv.ParseInt(value, 10, 32); err == nil {
		if seconds <= 0 || seconds > int64((24*time.Hour)/time.Second) {
			return 0
		}
		return time.Duration(seconds) * time.Second
	}
	when, err := http.ParseTime(value)
	if err != nil || !when.After(now) {
		return 0
	}
	delay := when.Sub(now)
	if delay > 24*time.Hour {
		return 0
	}
	return delay
}

func positiveOrDefault(value, fallback time.Duration) time.Duration {
	if value <= 0 {
		return fallback
	}
	return value
}

func positiveInt64OrDefault(value, fallback int64) int64 {
	if value <= 0 {
		return fallback
	}
	return value
}
