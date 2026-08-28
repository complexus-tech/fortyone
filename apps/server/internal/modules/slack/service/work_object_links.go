package slack

import (
	"net/url"
	"strings"

	"github.com/google/uuid"
)

// ParseFortyOneStoryURL accepts only canonical production story routes under a
// single FortyOne workspace subdomain. This rejects look-alike hosts, API/docs
// routes, credentials, ports, and encoded path separators before any lookup.
func ParseFortyOneStoryURL(rawURL string) (FortyOneStoryLink, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || !strings.EqualFold(parsed.Scheme, "https") || parsed.User != nil || parsed.Port() != "" {
		return FortyOneStoryLink{}, ErrInvalidFortyOneStoryURL
	}

	host := strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
	const domainSuffix = ".fortyone.app"
	if !strings.HasSuffix(host, domainSuffix) {
		return FortyOneStoryLink{}, ErrInvalidFortyOneStoryURL
	}
	workspaceSlug := strings.TrimSuffix(host, domainSuffix)
	if strings.Contains(workspaceSlug, ".") || !workspaceSlugPattern.MatchString(workspaceSlug) {
		return FortyOneStoryLink{}, ErrInvalidFortyOneStoryURL
	}

	escapedPath := strings.ToLower(parsed.EscapedPath())
	if strings.Contains(escapedPath, "%2f") || strings.Contains(escapedPath, "%5c") {
		return FortyOneStoryLink{}, ErrInvalidFortyOneStoryURL
	}
	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(segments) != 2 || segments[0] != "work" {
		return FortyOneStoryLink{}, ErrInvalidFortyOneStoryURL
	}
	storyReference := strings.ToUpper(strings.TrimSpace(segments[1]))
	if !storyReferencePattern.MatchString(storyReference) {
		return FortyOneStoryLink{}, ErrInvalidFortyOneStoryURL
	}

	postedURL := *parsed
	postedURL.Scheme = "https"
	postedURL.Host = host
	canonicalURL := url.URL{
		Scheme: "https",
		Host:   host,
		Path:   "/work/" + storyReference,
	}
	return FortyOneStoryLink{
		PostedURL:      postedURL.String(),
		CanonicalURL:   canonicalURL.String(),
		WorkspaceSlug:  workspaceSlug,
		StoryReference: storyReference,
	}, nil
}

// ParseFortyOneRequestURL accepts only canonical production request routes
// under one FortyOne workspace subdomain. Team and request identities must be
// UUIDs, and encoded path data cannot alter the route boundary.
func ParseFortyOneRequestURL(rawURL string) (FortyOneRequestLink, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || !strings.EqualFold(parsed.Scheme, "https") || parsed.User != nil || parsed.Port() != "" {
		return FortyOneRequestLink{}, ErrInvalidFortyOneRequestURL
	}

	host := strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
	const domainSuffix = ".fortyone.app"
	if !strings.HasSuffix(host, domainSuffix) {
		return FortyOneRequestLink{}, ErrInvalidFortyOneRequestURL
	}
	workspaceSlug := strings.TrimSuffix(host, domainSuffix)
	if strings.Contains(workspaceSlug, ".") || !workspaceSlugPattern.MatchString(workspaceSlug) {
		return FortyOneRequestLink{}, ErrInvalidFortyOneRequestURL
	}

	if parsed.EscapedPath() != parsed.Path {
		return FortyOneRequestLink{}, ErrInvalidFortyOneRequestURL
	}
	segments := strings.Split(strings.TrimSuffix(parsed.Path, "/"), "/")
	if len(segments) != 5 || segments[0] != "" || segments[1] != "teams" || segments[3] != "requests" {
		return FortyOneRequestLink{}, ErrInvalidFortyOneRequestURL
	}
	teamID, err := uuid.Parse(segments[2])
	if err != nil || teamID == uuid.Nil || segments[2] != teamID.String() {
		return FortyOneRequestLink{}, ErrInvalidFortyOneRequestURL
	}
	requestID, err := uuid.Parse(segments[4])
	if err != nil || requestID == uuid.Nil || segments[4] != requestID.String() {
		return FortyOneRequestLink{}, ErrInvalidFortyOneRequestURL
	}

	postedURL := *parsed
	postedURL.Scheme = "https"
	postedURL.Host = host
	canonicalURL := url.URL{
		Scheme: "https",
		Host:   host,
		Path:   "/teams/" + teamID.String() + "/requests/" + requestID.String(),
	}
	return FortyOneRequestLink{
		PostedURL:     postedURL.String(),
		CanonicalURL:  canonicalURL.String(),
		WorkspaceSlug: workspaceSlug,
		TeamID:        teamID,
		RequestID:     requestID,
	}, nil
}

// ParseFortyOneObjectiveURL accepts only the canonical team-scoped objective
// route under one FortyOne workspace subdomain.
func ParseFortyOneObjectiveURL(rawURL string) (FortyOneObjectiveLink, error) {
	parsed, workspaceSlug, segments, err := parseFortyOneTeamScopedURL(rawURL, ErrInvalidFortyOneObjectiveURL, []string{"teams", "", "objectives", ""})
	if err != nil {
		return FortyOneObjectiveLink{}, err
	}
	teamID, objectiveID, err := parseFortyOneTeamScopedIDs(segments, ErrInvalidFortyOneObjectiveURL)
	if err != nil {
		return FortyOneObjectiveLink{}, err
	}
	canonicalURL := url.URL{
		Scheme: "https",
		Host:   parsed.Hostname(),
		Path:   "/teams/" + teamID.String() + "/objectives/" + objectiveID.String(),
	}
	postedURL := canonicalPostedURL(parsed, parsed.Hostname())
	return FortyOneObjectiveLink{
		PostedURL:     postedURL.String(),
		CanonicalURL:  canonicalURL.String(),
		WorkspaceSlug: workspaceSlug,
		TeamID:        teamID,
		ObjectiveID:   objectiveID,
	}, nil
}

// ParseFortyOneSprintURL accepts only the canonical sprint stories route under
// one FortyOne workspace subdomain.
func ParseFortyOneSprintURL(rawURL string) (FortyOneSprintLink, error) {
	parsed, workspaceSlug, segments, err := parseFortyOneTeamScopedURL(rawURL, ErrInvalidFortyOneSprintURL, []string{"teams", "", "sprints", "", "stories"})
	if err != nil {
		return FortyOneSprintLink{}, err
	}
	teamID, sprintID, err := parseFortyOneTeamScopedIDs(segments, ErrInvalidFortyOneSprintURL)
	if err != nil {
		return FortyOneSprintLink{}, err
	}
	canonicalURL := url.URL{
		Scheme: "https",
		Host:   parsed.Hostname(),
		Path:   "/teams/" + teamID.String() + "/sprints/" + sprintID.String() + "/stories",
	}
	postedURL := canonicalPostedURL(parsed, parsed.Hostname())
	return FortyOneSprintLink{
		PostedURL:     postedURL.String(),
		CanonicalURL:  canonicalURL.String(),
		WorkspaceSlug: workspaceSlug,
		TeamID:        teamID,
		SprintID:      sprintID,
	}, nil
}

func parseFortyOneTeamScopedURL(rawURL string, invalidURL error, expected []string) (url.URL, string, []string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || !strings.EqualFold(parsed.Scheme, "https") || parsed.User != nil || parsed.Port() != "" {
		return url.URL{}, "", nil, invalidURL
	}
	host := strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
	const domainSuffix = ".fortyone.app"
	if !strings.HasSuffix(host, domainSuffix) {
		return url.URL{}, "", nil, invalidURL
	}
	workspaceSlug := strings.TrimSuffix(host, domainSuffix)
	if strings.Contains(workspaceSlug, ".") || !workspaceSlugPattern.MatchString(workspaceSlug) {
		return url.URL{}, "", nil, invalidURL
	}
	if parsed.EscapedPath() != parsed.Path {
		return url.URL{}, "", nil, invalidURL
	}
	pathSegments := strings.Split(strings.TrimSuffix(parsed.Path, "/"), "/")
	if len(pathSegments) != len(expected)+1 || pathSegments[0] != "" {
		return url.URL{}, "", nil, invalidURL
	}
	for index, expectedSegment := range expected {
		if expectedSegment != "" && pathSegments[index+1] != expectedSegment {
			return url.URL{}, "", nil, invalidURL
		}
	}
	parsed.Scheme = "https"
	parsed.Host = host
	return *parsed, workspaceSlug, pathSegments[1:], nil
}

func parseFortyOneTeamScopedIDs(segments []string, invalidURL error) (uuid.UUID, uuid.UUID, error) {
	if len(segments) < 4 {
		return uuid.Nil, uuid.Nil, invalidURL
	}
	teamID, err := uuid.Parse(segments[1])
	if err != nil || teamID == uuid.Nil || segments[1] != teamID.String() {
		return uuid.Nil, uuid.Nil, invalidURL
	}
	entityID, err := uuid.Parse(segments[3])
	if err != nil || entityID == uuid.Nil || segments[3] != entityID.String() {
		return uuid.Nil, uuid.Nil, invalidURL
	}
	return teamID, entityID, nil
}

func canonicalPostedURL(parsed url.URL, host string) url.URL {
	parsed.Scheme = "https"
	parsed.Host = host
	return parsed
}
