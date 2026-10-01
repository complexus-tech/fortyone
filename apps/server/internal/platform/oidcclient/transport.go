package oidcclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/safehttp"
)

// Every discovery, token, and JWKS request revalidates DNS and pins the dial.
// Provider-controlled metadata cannot redirect the API to internal services.
type publicTransport struct{}

func (publicTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.Method != http.MethodGet && request.Method != http.MethodPost {
		return nil, errors.New("unsupported OIDC request")
	}
	target, err := safehttp.Resolve(request.Context(), net.DefaultResolver, request.URL.String())
	if err != nil {
		return nil, err
	}
	dialer := &net.Dialer{Timeout: 3 * time.Second}
	transport := &http.Transport{Proxy: nil, DisableKeepAlives: true, ForceAttemptHTTP2: true, TLSHandshakeTimeout: 5 * time.Second, ResponseHeaderTimeout: 5 * time.Second, MaxResponseHeaderBytes: 32 << 10, TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12, ServerName: target.Hostname}, DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, net.JoinHostPort(target.Addresses[0].String(), target.Port))
	}}
	defer transport.CloseIdleConnections()
	response, err := transport.RoundTrip(request.Clone(request.Context()))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, (1<<20)+1))
	if err != nil {
		return nil, err
	}
	if len(body) > 1<<20 {
		return nil, errors.New("OIDC response exceeds size limit")
	}
	response.Body = io.NopCloser(bytes.NewReader(body))
	response.ContentLength = int64(len(body))
	return response, nil
}
func publicHTTPClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second, Transport: publicTransport{}, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("OIDC redirect denied") }}
}
