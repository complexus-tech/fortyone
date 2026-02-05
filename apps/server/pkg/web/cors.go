package web

import (
	"net/http"
	"net/url"
	"strings"
)

const baseDomain = "fortyone.app"

// AllowedOrigin returns the validated CORS origin or an empty string.
func AllowedOrigin(r *http.Request) string {
	return allowedOrigin(r.Header.Get("Origin"))
}

func allowedOrigin(origin string) string {
	if origin == "" {
		return ""
	}

	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}

	host := strings.ToLower(parsed.Hostname())
	if host == "" {
		return ""
	}

	if isLocalhost(host) {
		return origin
	}

	if host == baseDomain || strings.HasSuffix(host, "."+baseDomain) {
		return origin
	}

	return ""
}

func isLocalhost(host string) bool {
	if host == "localhost" || host == "127.0.0.1" {
		return true
	}

	return strings.HasSuffix(host, ".localhost")
}
