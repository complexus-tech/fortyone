package subscriptions

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

const stripeCheckoutSessionIDPlaceholder = "{CHECKOUT_SESSION_ID}"

func parseBillingOrigin(rawURL string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil {
		return nil, ErrInvalidBillingRedirect
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return nil, ErrInvalidBillingRedirect
	}
	return &url.URL{Scheme: parsed.Scheme, Host: parsed.Host}, nil
}

func (service *Service) billingRedirect(rawURL, workspaceSlug string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" {
		return nil, ErrInvalidBillingRedirect
	}
	if service.redirectOrigin != nil {
		if (parsed.Scheme != service.redirectOrigin.Scheme || !strings.EqualFold(parsed.Host, service.redirectOrigin.Host)) && !service.isWorkspaceBillingOrigin(parsed, workspaceSlug) {
			return nil, ErrInvalidBillingRedirect
		}
	} else if parsed.Scheme != "https" {
		return nil, ErrInvalidBillingRedirect
	}
	if parsed.Path == "" {
		parsed.Path = "/"
	}
	if parsed.Opaque != "" {
		return nil, fmt.Errorf("%w: opaque URL", ErrInvalidBillingRedirect)
	}
	return parsed, nil
}

func (service *Service) checkoutSuccessRedirect(rawURL, workspaceSlug string) (string, error) {
	redirect, err := service.billingRedirect(rawURL, workspaceSlug)
	if err != nil {
		return "", err
	}

	query := redirect.Query()
	query.Set("session_id", stripeCheckoutSessionIDPlaceholder)
	redirect.RawQuery = strings.ReplaceAll(
		query.Encode(),
		url.QueryEscape(stripeCheckoutSessionIDPlaceholder),
		stripeCheckoutSessionIDPlaceholder,
	)
	return redirect.String(), nil
}

var billingWorkspaceLabel = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)

// Hosted workspaces use <slug>.fortyone.app, rather than the configured
// login origin. The slug must come from authenticated workspace middleware.
// Local/custom deployments retain exact-origin validation.
func (service *Service) isWorkspaceBillingOrigin(candidate *url.URL, workspaceSlug string) bool {
	if service.redirectOrigin == nil || service.redirectOrigin.Scheme != "https" || service.redirectOrigin.Port() != "" {
		return false
	}
	configuredHost := strings.ToLower(service.redirectOrigin.Hostname())
	if configuredHost != "fortyone.app" && !strings.HasSuffix(configuredHost, ".fortyone.app") {
		return false
	}
	if !billingWorkspaceLabel.MatchString(workspaceSlug) {
		return false
	}
	return candidate.Scheme == "https" && strings.EqualFold(candidate.Host, workspaceSlug+".fortyone.app")
}
