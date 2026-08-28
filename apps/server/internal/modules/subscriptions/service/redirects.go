package subscriptions

import (
	"fmt"
	"net/url"
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

func (service *Service) billingRedirect(rawURL string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" {
		return nil, ErrInvalidBillingRedirect
	}
	if service.redirectOrigin != nil {
		if parsed.Scheme != service.redirectOrigin.Scheme || !strings.EqualFold(parsed.Host, service.redirectOrigin.Host) {
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

func (service *Service) checkoutSuccessRedirect(rawURL string) (string, error) {
	redirect, err := service.billingRedirect(rawURL)
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
