package web

import (
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"sort"
	"strings"
)

const maxAllowedOrigins = 16

var ErrInvalidAllowedOrigin = errors.New("invalid CORS allowed origin")

type subdomainOrigin struct {
	scheme string
	domain string
}

// OriginPolicy is an immutable credentialed-origin allowlist. Entries are
// either exact origins or an explicitly configured leading subdomain pattern
// such as https://*.fortyone.app. A subdomain pattern never matches its apex,
// another scheme, an explicit port, or a lookalike domain.
type OriginPolicy struct {
	origins    map[string]struct{}
	subdomains map[subdomainOrigin]struct{}
}

func NewOriginPolicy(raw string) (OriginPolicy, error) {
	parts := strings.Split(raw, ",")
	origins := make(map[string]struct{}, len(parts))
	subdomains := make(map[subdomainOrigin]struct{})
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "*") {
			subdomain, err := canonicalSubdomainOrigin(part)
			if err != nil {
				return OriginPolicy{}, err
			}
			subdomains[subdomain] = struct{}{}
			if len(origins)+len(subdomains) > maxAllowedOrigins {
				return OriginPolicy{}, fmt.Errorf("%w: at most %d origins are allowed", ErrInvalidAllowedOrigin, maxAllowedOrigins)
			}
			continue
		}
		origin, err := canonicalOrigin(part)
		if err != nil {
			return OriginPolicy{}, err
		}
		origins[origin] = struct{}{}
		if len(origins)+len(subdomains) > maxAllowedOrigins {
			return OriginPolicy{}, fmt.Errorf("%w: at most %d origins are allowed", ErrInvalidAllowedOrigin, maxAllowedOrigins)
		}
	}
	if len(origins)+len(subdomains) == 0 {
		return OriginPolicy{}, fmt.Errorf("%w: at least one origin is required", ErrInvalidAllowedOrigin)
	}
	return OriginPolicy{origins: origins, subdomains: subdomains}, nil
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
	if _, allowed := p.origins[canonical]; allowed {
		return raw
	}
	parsed, err := url.Parse(canonical)
	if err != nil || parsed.Port() != "" {
		return ""
	}
	hostname := strings.ToLower(parsed.Hostname())
	for subdomain := range p.subdomains {
		if parsed.Scheme == subdomain.scheme && strings.HasSuffix(hostname, "."+subdomain.domain) {
			return raw
		}
	}
	return ""
}

func (p OriginPolicy) Origins() []string {
	origins := make([]string, 0, len(p.origins)+len(p.subdomains))
	for origin := range p.origins {
		origins = append(origins, origin)
	}
	for subdomain := range p.subdomains {
		origins = append(origins, subdomain.scheme+"://*."+subdomain.domain)
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

func canonicalSubdomainOrigin(raw string) (subdomainOrigin, error) {
	if strings.Count(raw, "*") != 1 {
		return subdomainOrigin{}, fmt.Errorf("%w: only one leading subdomain wildcard is supported", ErrInvalidAllowedOrigin)
	}
	scheme, host, found := strings.Cut(raw, "://")
	if !found || !strings.HasPrefix(host, "*.") {
		return subdomainOrigin{}, fmt.Errorf("%w: wildcard must be the complete leading hostname label", ErrInvalidAllowedOrigin)
	}
	canonical, err := canonicalOrigin(scheme + "://wildcard." + strings.TrimPrefix(host, "*."))
	if err != nil {
		return subdomainOrigin{}, err
	}
	parsed, err := url.Parse(canonical)
	if err != nil || parsed.Port() != "" {
		return subdomainOrigin{}, fmt.Errorf("%w: subdomain wildcard origins cannot contain a port", ErrInvalidAllowedOrigin)
	}
	domain := strings.TrimPrefix(strings.ToLower(parsed.Hostname()), "wildcard.")
	if domain == parsed.Hostname() || !strings.Contains(domain, ".") {
		return subdomainOrigin{}, fmt.Errorf("%w: subdomain wildcard must contain a multi-label domain", ErrInvalidAllowedOrigin)
	}
	if _, err := netip.ParseAddr(domain); err == nil {
		return subdomainOrigin{}, fmt.Errorf("%w: subdomain wildcard cannot target an IP address", ErrInvalidAllowedOrigin)
	}
	return subdomainOrigin{scheme: parsed.Scheme, domain: domain}, nil
}

func canonicalOrigin(raw string) (string, error) {
	if raw == "*" || strings.EqualFold(raw, "null") {
		return "", fmt.Errorf("%w: global wildcards and opaque origins are not supported", ErrInvalidAllowedOrigin)
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
