package workspaceurl

import (
	"fmt"
	"net"
	"net/url"
	"path"
	"strings"
)

const hostedDomain = "fortyone.app"

// Build returns a workspace-aware application URL. FortyOne's production domain
// uses the workspace slug as a subdomain, while local and preview origins keep
// the workspace slug in the path.
func Build(websiteURL, workspaceSlug string, routeSegments ...string) string {
	baseURL, err := url.Parse(strings.TrimRight(strings.TrimSpace(websiteURL), "/"))
	if err != nil ||
		(baseURL.Scheme != "http" && baseURL.Scheme != "https") ||
		baseURL.User != nil ||
		strings.TrimSpace(baseURL.Hostname()) == "" {
		return ""
	}

	workspaceSlug = strings.TrimSpace(workspaceSlug)
	if workspaceSlug == "" {
		return ""
	}

	segments := make([]string, 0, len(routeSegments)+2)
	if basePath := strings.Trim(baseURL.Path, "/"); basePath != "" {
		segments = append(segments, basePath)
	}

	host := strings.ToLower(baseURL.Hostname())
	if host == hostedDomain || strings.HasSuffix(host, "."+hostedDomain) {
		if !validWorkspaceSlug(workspaceSlug) {
			return ""
		}
		baseURL.Host = workspaceHost(baseURL, workspaceSlug)
	} else {
		segments = append(segments, workspaceSlug)
	}

	for _, segment := range routeSegments {
		if segment = strings.Trim(segment, "/ "); segment != "" {
			segments = append(segments, segment)
		}
	}

	baseURL.Path = path.Join(append([]string{"/"}, segments...)...)
	baseURL.RawPath = ""
	baseURL.RawQuery = ""
	baseURL.Fragment = ""
	return baseURL.String()
}

func workspaceHost(baseURL *url.URL, workspaceSlug string) string {
	host := fmt.Sprintf("%s.%s", workspaceSlug, hostedDomain)
	if port := baseURL.Port(); port != "" {
		return net.JoinHostPort(host, port)
	}
	return host
}

func validWorkspaceSlug(value string) bool {
	if value == "" || len(value) > 63 || value[0] == '-' || value[len(value)-1] == '-' {
		return false
	}
	for _, char := range value {
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '-' {
			return false
		}
	}
	return true
}
