package github

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/url"
	"path"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) appAPIConfigDiagnostics() []any {
	return []any{
		"app_id_configured", s.cfg.AppID != 0,
		"app_slug_configured", strings.TrimSpace(s.cfg.AppSlug) != "",
		"private_key_base64_present", strings.TrimSpace(s.cfg.PrivateKeyBase64) != "",
		"private_key_loaded", s.privateKey != nil,
		"private_key_load_error", s.privateKeyLoadError,
		"redirect_url_configured", strings.TrimSpace(s.cfg.RedirectURL) != "",
		"webhook_secret_configured", strings.TrimSpace(s.cfg.WebhookSecret) != "",
	}
}

func buildLinkedTaskComment(storyURL, teamCode string, sequenceID int) string {
	normalizedTeamCode := strings.ToUpper(strings.TrimSpace(teamCode))
	taskKey := fmt.Sprintf("#%d", sequenceID)
	if normalizedTeamCode != "" {
		taskKey = fmt.Sprintf("%s-%d", normalizedTeamCode, sequenceID)
	}
	return fmt.Sprintf(
		"Linked to FortyOne task [%s](%s).",
		taskKey,
		storyURL,
	)
}

func githubString(value string) *string {
	return &value
}

func storyCommentMarker(storyID uuid.UUID) string {
	return fmt.Sprintf("`%s`", storyID.String())
}

func buildStoryReference(teamCode string, sequenceID int, fallbackID string) string {
	normalizedCode := strings.ToUpper(strings.TrimSpace(teamCode))
	if normalizedCode != "" && sequenceID > 0 {
		return fmt.Sprintf("%s-%d", normalizedCode, sequenceID)
	}
	return strings.TrimSpace(fallbackID)
}

func storyURLFromWebsite(websiteURL, workspaceSlug, storyReference string) (string, error) {
	baseURL, err := url.Parse(strings.TrimRight(websiteURL, "/"))
	if err != nil {
		return "", err
	}

	if workspaceSlug == "" {
		return "", errors.New("workspace slug is required")
	}
	if strings.TrimSpace(storyReference) == "" {
		return "", errors.New("story reference is required")
	}

	baseURL.Path = path.Join("/", "work", storyReference)

	host := baseURL.Hostname()
	if host == "" {
		return "", errors.New("website host is required")
	}

	if isLocalWebsiteHost(host) {
		baseURL.Path = path.Join("/", workspaceSlug, "work", storyReference)
		return baseURL.String(), nil
	}

	if !strings.HasPrefix(host, workspaceSlug+".") {
		if port := baseURL.Port(); port != "" {
			baseURL.Host = fmt.Sprintf("%s.%s:%s", workspaceSlug, host, port)
		} else {
			baseURL.Host = fmt.Sprintf("%s.%s", workspaceSlug, host)
		}
	}

	return baseURL.String(), nil
}

func isLocalWebsiteHost(host string) bool {
	return strings.EqualFold(host, "localhost") || strings.EqualFold(host, "0.0.0.0") || net.ParseIP(host) != nil
}

func loadPrivateKey(privateKeyBase64 string) (*rsa.PrivateKey, error) {
	pemBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(privateKeyBase64))
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode private key: %w", err)
	}
	defer clear(pemBytes)
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("invalid github private key: no PEM block found after base64 decoding")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return validateGitHubPrivateKey(key)
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("github private key is not RSA")
	}
	return validateGitHubPrivateKey(rsaKey)
}

func validateGitHubPrivateKey(key *rsa.PrivateKey) (*rsa.PrivateKey, error) {
	if key == nil || key.N == nil || key.N.BitLen() < 2048 {
		return nil, errors.New("github private key must be at least 2048 bits")
	}
	if err := key.Validate(); err != nil {
		return nil, errors.New("github private key is invalid")
	}
	return key, nil
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
