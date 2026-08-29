package slack

import (
	"fmt"
	"net"
	"net/url"
	"path"
	"strings"
)

func (s *Service) buildWorkspaceIntegrationURL(workspaceSlug string) string {
	link := buildWorkspaceURL(s.cfg.WebsiteURL, workspaceSlug, "settings", "workspace", "integrations", "slack")
	if link == "" {
		return "/"
	}
	return link
}

func (s *Service) buildAccountIntegrationURL(workspaceSlug string) string {
	link := buildWorkspaceURL(s.cfg.WebsiteURL, workspaceSlug, "settings", "integrations", "slack")
	if link == "" {
		return "/"
	}
	return link
}

func (s *Service) safeSlackAccountLinkReturnURL(workspaceSlug, returnURL string) string {
	fallback := s.buildAccountIntegrationURL(workspaceSlug)
	candidate, err := url.Parse(strings.TrimSpace(returnURL))
	if err != nil || candidate.Scheme == "" || candidate.Host == "" || candidate.User != nil {
		return fallback
	}
	expected, err := url.Parse(buildWorkspaceURL(s.cfg.WebsiteURL, workspaceSlug))
	if err != nil || candidate.Scheme != expected.Scheme || candidate.Host != expected.Host {
		return fallback
	}
	return candidate.String()
}

func (s *Service) slackAccountLinkRedirect(returnURL, status string) string {
	parsed, err := url.Parse(strings.TrimSpace(returnURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return returnURL
	}
	query := parsed.Query()
	query.Set("slack_link_status", strings.TrimSpace(status))
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func buildWorkspaceURL(websiteURL, workspaceSlug string, routeSegments ...string) string {
	baseURL, err := url.Parse(strings.TrimRight(strings.TrimSpace(websiteURL), "/"))
	if err != nil {
		return ""
	}
	if strings.TrimSpace(baseURL.Hostname()) == "" || strings.TrimSpace(workspaceSlug) == "" {
		return ""
	}

	cleanSegments := make([]string, 0, len(routeSegments))
	for _, segment := range routeSegments {
		if trimmed := strings.TrimSpace(segment); trimmed != "" {
			cleanSegments = append(cleanSegments, trimmed)
		}
	}

	host := baseURL.Hostname()
	if isLocalWebsiteHost(host) {
		baseURL.Path = path.Join(append([]string{"/", workspaceSlug}, cleanSegments...)...)
		return baseURL.String()
	}

	baseURL.Path = path.Join(append([]string{"/"}, cleanSegments...)...)
	if !strings.HasPrefix(host, workspaceSlug+".") {
		if port := baseURL.Port(); port != "" {
			baseURL.Host = fmt.Sprintf("%s.%s:%s", workspaceSlug, host, port)
		} else {
			baseURL.Host = fmt.Sprintf("%s.%s", workspaceSlug, host)
		}
	}

	return baseURL.String()
}

func buildStoryReference(teamCode string, sequenceID int, fallbackID string) string {
	if storyCode := buildStoryCode(teamCode, sequenceID); storyCode != "" {
		return storyCode
	}
	return strings.TrimSpace(fallbackID)
}

func buildStoryCode(teamCode string, sequenceID int) string {
	normalizedCode := strings.ToUpper(strings.TrimSpace(teamCode))
	if normalizedCode == "" || sequenceID <= 0 {
		return ""
	}
	return fmt.Sprintf("%s-%d", normalizedCode, sequenceID)
}

func buildTaskURL(websiteURL, workspaceSlug, storyReference string) string {
	if strings.TrimSpace(storyReference) == "" {
		return ""
	}
	return buildWorkspaceURL(websiteURL, workspaceSlug, "work", storyReference)
}

func buildRequestURL(websiteURL, workspaceSlug, teamID, requestID string) string {
	if strings.TrimSpace(teamID) == "" || strings.TrimSpace(requestID) == "" {
		return ""
	}
	return buildWorkspaceURL(websiteURL, workspaceSlug, "teams", teamID, "requests", requestID)
}

func isLocalWebsiteHost(host string) bool {
	return strings.EqualFold(host, "localhost") || strings.EqualFold(host, "0.0.0.0") || net.ParseIP(host) != nil
}
