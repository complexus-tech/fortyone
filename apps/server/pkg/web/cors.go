package web

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

const maxAllowedOrigins = 16

var ErrInvalidAllowedOrigin = errors.New("invalid CORS allowed origin")

// OriginPolicy is an immutable exact-origin allowlist. It intentionally does
// not support wildcards or parent-domain matching because a forgotten or
// compromised sibling subdomain must not inherit credentialed API access.
type OriginPolicy struct {
	origins map[string]struct{}
}

func NewOriginPolicy(raw string) (OriginPolicy, error) {
	parts := strings.Split(raw, ",")
	origins := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		origin, err := canonicalOrigin(part)
		if err != nil {
			return OriginPolicy{}, err
		}
		origins[origin] = struct{}{}
		if len(origins) > maxAllowedOrigins {
			return OriginPolicy{}, fmt.Errorf("%w: at most %d origins are allowed", ErrInvalidAllowedOrigin, maxAllowedOrigins)
		}
	}
	if len(origins) == 0 {
		return OriginPolicy{}, fmt.Errorf("%w: at least one origin is required", ErrInvalidAllowedOrigin)
	}
	return OriginPolicy{origins: origins}, nil
}

func (p OriginPolicy) AllowedOrigin(r *http.Request) string {
	raw := strings.TrimSpace(r.Header.Get("Origin"))
	if raw == "" {
		return ""
	}
	canonical, err := canonicalOrigin(raw)
	if err != nil {
		return ""
	}
	if _, allowed := p.origins[canonical]; !allowed {
		return ""
	}
	return raw
}

func (p OriginPolicy) Origins() []string {
	origins := make([]string, 0, len(p.origins))
	for origin := range p.origins {
		origins = append(origins, origin)
	}
	sort.Strings(origins)
	return origins
}

func (p OriginPolicy) ValidateHTTPS() error {
	for _, origin := range p.Origins() {
		parsed, err := url.Parse(origin)
		if err != nil || parsed.Scheme != "https" {
			return fmt.Errorf("%w: production origins must use https", ErrInvalidAllowedOrigin)
		}
	}
	return nil
}

func canonicalOrigin(raw string) (string, error) {
	if raw == "*" || strings.EqualFold(raw, "null") {
		return "", fmt.Errorf("%w: wildcards and opaque origins are not supported", ErrInvalidAllowedOrigin)
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil {
		return "", ErrInvalidAllowedOrigin
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", fmt.Errorf("%w: origin scheme must be http or https", ErrInvalidAllowedOrigin)
	}
	if parsed.Path != "" && parsed.Path != "/" {
		return "", fmt.Errorf("%w: origins cannot contain a path", ErrInvalidAllowedOrigin)
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("%w: origins cannot contain a query or fragment", ErrInvalidAllowedOrigin)
	}
	host := strings.ToLower(parsed.Host)
	if strings.TrimSpace(parsed.Hostname()) == "" || strings.Contains(host, "*") {
		return "", ErrInvalidAllowedOrigin
	}
	return scheme + "://" + host, nil
}
