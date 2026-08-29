package safehttp

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"
)

const maximumDownloadBytes = int64(8 << 20)

type Downloader struct {
	resolver               Resolver
	dialContext            DialContextFunc
	timeout                time.Duration
	tlsHandshakeTimeout    time.Duration
	responseHeaderTimeout  time.Duration
	maxResponseBytes       int64
	maxResponseHeaderBytes int64
}

type Download struct {
	Body        []byte
	ContentType string
	ResolvedIP  netip.Addr
	Duration    time.Duration
}

func NewDownloader(config Config) (*Downloader, error) {
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
	if maxBody > maximumDownloadBytes || maxHeaders > 64<<10 {
		return nil, fmt.Errorf("%w: unsafe response limit configuration", ErrUnsupportedRequest)
	}
	return &Downloader{
		resolver:               resolver,
		dialContext:            dialContext,
		timeout:                timeout,
		tlsHandshakeTimeout:    tlsTimeout,
		responseHeaderTimeout:  headerTimeout,
		maxResponseBytes:       maxBody,
		maxResponseHeaderBytes: maxHeaders,
	}, nil
}

// Download retrieves one bounded HTTPS resource. It never uses environment
// proxies, follows redirects, reuses connections, accepts IP-literal targets,
// or performs an unvalidated second DNS lookup.
func (downloader *Downloader) Download(ctx context.Context, rawURL string) (Download, error) {
	if downloader == nil {
		return Download{}, fmt.Errorf("%w: downloader is required", ErrUnsupportedRequest)
	}
	target, err := Resolve(ctx, downloader.resolver, rawURL)
	if err != nil {
		return Download{}, err
	}

	transport := &http.Transport{
		Proxy:                  nil,
		DisableKeepAlives:      true,
		ForceAttemptHTTP2:      true,
		TLSHandshakeTimeout:    downloader.tlsHandshakeTimeout,
		ResponseHeaderTimeout:  downloader.responseHeaderTimeout,
		MaxResponseHeaderBytes: downloader.maxResponseHeaderBytes,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: target.Hostname,
		},
		DialContext: func(dialCtx context.Context, network, _ string) (net.Conn, error) {
			if network != "tcp" && network != "tcp4" && network != "tcp6" {
				return nil, fmt.Errorf("%w: network %q", ErrUnsupportedRequest, network)
			}
			return downloader.dialContext(dialCtx, "tcp", target.dialAddress())
		},
	}
	defer transport.CloseIdleConnections()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target.URL.String(), nil)
	if err != nil {
		return Download{}, fmt.Errorf("build safe download request: %w", err)
	}
	request.Host = target.Hostname
	client := &http.Client{
		Transport: transport,
		Timeout:   downloader.timeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return ErrRedirectDenied
		},
	}
	started := time.Now()
	response, err := client.Do(request)
	if err != nil {
		return Download{ResolvedIP: target.Addresses[0], Duration: time.Since(started)}, fmt.Errorf("safe download request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return Download{ResolvedIP: target.Addresses[0], Duration: time.Since(started)}, fmt.Errorf("%w: status %d", ErrUnexpectedStatus, response.StatusCode)
	}
	if response.ContentLength > downloader.maxResponseBytes {
		return Download{ResolvedIP: target.Addresses[0], Duration: time.Since(started)}, ErrResponseTooLarge
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, downloader.maxResponseBytes+1))
	if err != nil {
		return Download{ResolvedIP: target.Addresses[0], Duration: time.Since(started)}, fmt.Errorf("read safe download response: %w", err)
	}
	if int64(len(body)) > downloader.maxResponseBytes {
		return Download{ResolvedIP: target.Addresses[0], Duration: time.Since(started)}, ErrResponseTooLarge
	}
	return Download{
		Body:        body,
		ContentType: strings.TrimSpace(strings.Split(response.Header.Get("Content-Type"), ";")[0]),
		ResolvedIP:  target.Addresses[0],
		Duration:    time.Since(started),
	}, nil
}
