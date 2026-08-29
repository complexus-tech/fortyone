package figma

import (
	"errors"
	"net/url"
	"strings"
)

var ErrInvalidFigmaURL = errors.New("enter a valid Figma file or frame URL")

func ParseURL(raw string) (Artifact, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme != "https" {
		return Artifact{}, ErrInvalidFigmaURL
	}
	host := strings.ToLower(parsed.Hostname())
	if host != "figma.com" && host != "www.figma.com" {
		return Artifact{}, ErrInvalidFigmaURL
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 || !isSupportedFileType(parts[0]) || strings.TrimSpace(parts[1]) == "" {
		return Artifact{}, ErrInvalidFigmaURL
	}
	fileKey := parts[1]
	var nodeID *string
	if value := strings.TrimSpace(parsed.Query().Get("node-id")); value != "" {
		normalized := strings.ReplaceAll(value, "-", ":")
		nodeID = &normalized
	}
	canonical := url.URL{Scheme: "https", Host: "www.figma.com", Path: "/" + parts[0] + "/" + fileKey}
	query := url.Values{}
	if nodeID != nil {
		query.Set("node-id", strings.ReplaceAll(*nodeID, ":", "-"))
	}
	canonical.RawQuery = query.Encode()
	return Artifact{FileKey: fileKey, NodeID: nodeID, OriginalURL: raw, CanonicalURL: canonical.String()}, nil
}

func isSupportedFileType(value string) bool {
	switch strings.ToLower(value) {
	case "design", "file", "proto", "board":
		return true
	default:
		return false
	}
}
