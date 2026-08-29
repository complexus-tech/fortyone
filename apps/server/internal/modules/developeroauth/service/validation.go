package developeroauth

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"sort"
	"strings"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
)

// ScopePolicy defines the permissions that one exact OAuth resource may issue.
// A policy belongs to an audience: an MCP token and an API token must never be
// accepted interchangeably even when they were granted to the same client.
type ScopePolicy struct {
	Supported       []string
	Required        []string
	RequireExplicit bool
}

type scopePolicy struct {
	supported       map[string]struct{}
	required        []string
	requireExplicit bool
}

// MCPResourceScopePolicy preserves the existing MCP consent contract.
func MCPResourceScopePolicy() ScopePolicy {
	return ScopePolicy{
		Supported: []string{
			developeroauthdomain.ScopeMCPAccess,
			developeroauthdomain.ScopeOfflineAccess,
		},
		Required: []string{
			developeroauthdomain.ScopeMCPAccess,
			developeroauthdomain.ScopeOfflineAccess,
		},
	}
}

// PublicAPIResourceScopePolicy is the least-privilege catalog for the current
// /api/v1 contract. It intentionally excludes permissions for unpublished API
// mutations. Callers must request at least one capability in addition to the
// refresh-token permission.
func PublicAPIResourceScopePolicy() ScopePolicy {
	return ScopePolicy{
		Supported: []string{
			developeroauthdomain.ScopeOfflineAccess,
			string(platformauth.ScopeWorkspacesRead),
			string(platformauth.ScopeTeamsRead),
			string(platformauth.ScopeStoriesRead),
			string(platformauth.ScopeStoriesWrite),
			string(platformauth.ScopeCommentsRead),
			string(platformauth.ScopeLabelsRead),
			string(platformauth.ScopeSprintsRead),
			string(platformauth.ScopeObjectivesRead),
			string(platformauth.ScopeWebhooksManage),
		},
		Required:        []string{developeroauthdomain.ScopeOfflineAccess},
		RequireExplicit: true,
	}
}

// PublicAPIApplicationActorScopePolicy is intentionally narrower than the
// delegated-user catalog. The first application-actor release can perform only
// idempotent story creation; membership-backed reads and webhook management
// remain user-only until their repositories have installation-aware policies.
func PublicAPIApplicationActorScopePolicy() ScopePolicy {
	return ScopePolicy{
		Supported:       []string{string(platformauth.ScopeStoriesWrite)},
		RequireExplicit: true,
	}
}

func newScopePolicy(config ScopePolicy) (scopePolicy, error) {
	if config.Supported == nil && config.Required == nil && !config.RequireExplicit {
		config = MCPResourceScopePolicy()
	}
	policy := scopePolicy{
		supported:       make(map[string]struct{}, len(config.Supported)),
		requireExplicit: config.RequireExplicit,
	}
	for _, raw := range config.Supported {
		scope := strings.TrimSpace(raw)
		if scope == "" || scope != raw || strings.ContainsAny(scope, "\t\r\n ") {
			return scopePolicy{}, errors.New("OAuth supported scopes must be non-empty canonical tokens")
		}
		if _, duplicate := policy.supported[scope]; duplicate {
			return scopePolicy{}, fmt.Errorf("OAuth supported scope %q is duplicated", scope)
		}
		policy.supported[scope] = struct{}{}
	}
	if len(policy.supported) == 0 {
		return scopePolicy{}, errors.New("OAuth scope policy must support at least one scope")
	}
	seenRequired := make(map[string]struct{}, len(config.Required))
	for _, raw := range config.Required {
		scope := strings.TrimSpace(raw)
		if _, supported := policy.supported[scope]; !supported {
			return scopePolicy{}, fmt.Errorf("OAuth required scope %q is not supported", raw)
		}
		if _, duplicate := seenRequired[scope]; duplicate {
			return scopePolicy{}, fmt.Errorf("OAuth required scope %q is duplicated", scope)
		}
		seenRequired[scope] = struct{}{}
		policy.required = append(policy.required, scope)
	}
	sort.Strings(policy.required)
	return policy, nil
}

func ValidateRedirectURI(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || parsed.Fragment != "" {
		return fmt.Errorf("%w: redirect URI must be absolute and must not contain a fragment", developeroauthdomain.ErrInvalidRedirectURI)
	}
	if parsed.User != nil {
		return fmt.Errorf("%w: redirect URI must not contain user information", developeroauthdomain.ErrInvalidRedirectURI)
	}
	if parsed.Scheme == "https" {
		return nil
	}
	host := parsed.Hostname()
	if parsed.Scheme == "http" && (host == "localhost" || host == "127.0.0.1" || host == "::1") {
		return nil
	}
	return fmt.Errorf("%w: redirect URI must use HTTPS except for loopback clients", developeroauthdomain.ErrInvalidRedirectURI)
}

func validateIssuerAndResource(issuer, resource string) error {
	issuerURL, err := url.Parse(issuer)
	if err != nil || issuerURL.Host == "" || issuerURL.Fragment != "" || issuerURL.RawQuery != "" {
		return errors.New("OAuth issuer must be an absolute origin or path without query or fragment")
	}
	resourceURL, err := url.Parse(resource)
	if err != nil || resourceURL.Host == "" || resourceURL.Fragment != "" || resourceURL.RawQuery != "" {
		return errors.New("OAuth resource must be an absolute URL without query or fragment")
	}
	return nil
}

func (policy scopePolicy) normalize(raw []string) ([]string, error) {
	unique := make(map[string]struct{}, len(raw)+len(policy.required))
	explicit := make(map[string]struct{}, len(raw))
	for _, value := range raw {
		for _, scope := range strings.Fields(value) {
			if _, supported := policy.supported[scope]; !supported {
				return nil, fmt.Errorf("%w: unsupported scope %q", developeroauthdomain.ErrInvalidScope, scope)
			}
			unique[scope] = struct{}{}
			explicit[scope] = struct{}{}
		}
	}
	for _, scope := range policy.required {
		unique[scope] = struct{}{}
		delete(explicit, scope)
	}
	if policy.requireExplicit && len(explicit) == 0 {
		return nil, fmt.Errorf("%w: at least one resource permission is required", developeroauthdomain.ErrInvalidScope)
	}
	result := make([]string, 0, len(unique))
	for scope := range unique {
		result = append(result, scope)
	}
	sort.Strings(result)
	return result, nil
}

func (policy scopePolicy) accepts(scopes []string) bool {
	if len(scopes) == 0 {
		return false
	}
	for _, scope := range scopes {
		if _, supported := policy.supported[scope]; !supported {
			return false
		}
	}
	return scopesAreSubset(policy.required, scopes)
}

func (policy scopePolicy) scopes() []string {
	result := make([]string, 0, len(policy.supported))
	for scope := range policy.supported {
		result = append(result, scope)
	}
	sort.Strings(result)
	return result
}

func validatePKCEChallenge(challenge string) error {
	decoded, err := base64.RawURLEncoding.DecodeString(challenge)
	if err != nil || len(decoded) != 32 || len(challenge) != 43 {
		return fmt.Errorf("%w: S256 challenge must be a 32-byte base64url digest", developeroauthdomain.ErrInvalidPKCE)
	}
	zeroByteSlices([][]byte{decoded})
	return nil
}

func validatePKCEVerifier(verifier string) error {
	if len(verifier) < 43 || len(verifier) > 128 {
		return fmt.Errorf("%w: verifier must contain 43-128 characters", developeroauthdomain.ErrInvalidPKCE)
	}
	for _, character := range verifier {
		if !(character >= 'A' && character <= 'Z') &&
			!(character >= 'a' && character <= 'z') &&
			!(character >= '0' && character <= '9') &&
			!slices.Contains([]rune{'-', '.', '_', '~'}, character) {
			return fmt.Errorf("%w: verifier contains an invalid character", developeroauthdomain.ErrInvalidPKCE)
		}
	}
	return nil
}

func containsExact(values []string, expected string) bool {
	return slices.Contains(values, expected)
}

func scopesAreSubset(candidate, allowed []string) bool {
	for _, scope := range candidate {
		if !slices.Contains(allowed, scope) {
			return false
		}
	}
	return true
}
