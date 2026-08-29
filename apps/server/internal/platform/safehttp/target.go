package safehttp

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"slices"
	"strings"
)

const defaultHTTPSPort = "443"

// Resolver is intentionally the narrow part of net.Resolver used by the
// egress policy. Keeping it injectable makes DNS-rebinding policy testable
// without weakening production validation.
type Resolver interface {
	LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error)
}

type Target struct {
	URL       *url.URL
	Hostname  string
	Port      string
	Addresses []netip.Addr
}

func (target Target) dialAddress() string {
	return net.JoinHostPort(target.Addresses[0].String(), target.Port)
}

// Resolve validates the endpoint and resolves it immediately. Callers may use
// this for create-time feedback, but delivery code must call Resolve again for
// every network attempt and dial one of the returned addresses directly.
func Resolve(ctx context.Context, resolver Resolver, rawURL string) (Target, error) {
	if resolver == nil {
		return Target{}, fmt.Errorf("%w: resolver is required", ErrInvalidEndpoint)
	}
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return Target{}, fmt.Errorf("%w: parse: %v", ErrInvalidEndpoint, err)
	}
	if parsed.Scheme != "https" {
		return Target{}, ErrInsecureScheme
	}
	if parsed.User != nil {
		return Target{}, ErrCredentialsInURL
	}
	if parsed.Fragment != "" {
		return Target{}, ErrFragmentInURL
	}
	if parsed.RawQuery == "" && parsed.ForceQuery {
		return Target{}, fmt.Errorf("%w: empty forced query", ErrInvalidEndpoint)
	}
	hostname := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
	if !isValidDNSHostname(hostname) {
		return Target{}, fmt.Errorf("%w: hostname is required", ErrInvalidEndpoint)
	}
	if net.ParseIP(hostname) != nil {
		return Target{}, ErrIPAddressHost
	}
	port := parsed.Port()
	if port == "" {
		port = defaultHTTPSPort
	}
	if port != defaultHTTPSPort {
		return Target{}, ErrUnsupportedPort
	}

	addresses, err := resolver.LookupNetIP(ctx, "ip", hostname)
	if err != nil {
		return Target{}, fmt.Errorf("resolve safe http endpoint: %w", err)
	}
	if len(addresses) == 0 {
		return Target{}, ErrNoAddresses
	}

	normalized := make([]netip.Addr, 0, len(addresses))
	seen := make(map[netip.Addr]struct{}, len(addresses))
	for _, address := range addresses {
		address = address.Unmap()
		if !isPublicAddress(address) {
			return Target{}, fmt.Errorf("%w: DNS answer contained a prohibited address", ErrUnsafeAddress)
		}
		if _, duplicate := seen[address]; duplicate {
			continue
		}
		seen[address] = struct{}{}
		normalized = append(normalized, address)
	}
	if len(normalized) == 0 {
		return Target{}, ErrNoAddresses
	}
	slices.SortFunc(normalized, func(left, right netip.Addr) int {
		return strings.Compare(left.String(), right.String())
	})

	canonical := *parsed
	canonical.Scheme = "https"
	canonical.Host = hostname
	if port != defaultHTTPSPort {
		canonical.Host = net.JoinHostPort(hostname, port)
	}
	return Target{
		URL:       &canonical,
		Hostname:  hostname,
		Port:      port,
		Addresses: normalized,
	}, nil
}

func isValidDNSHostname(hostname string) bool {
	if hostname == "" || len(hostname) > 253 {
		return false
	}
	labels := strings.Split(hostname, ".")
	if len(labels) < 2 {
		return false
	}
	for _, label := range labels {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for index := range len(label) {
			character := label[index]
			if (character < 'a' || character > 'z') && (character < '0' || character > '9') && character != '-' {
				return false
			}
		}
	}
	return true
}

var prohibitedPrefixes = []netip.Prefix{
	// IPv4 special-purpose, private, link-local, documentation, benchmark,
	// multicast, and reserved ranges. Rejecting the complete DNS answer avoids
	// selecting a public record while a rebinding/private record is also live.
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("10.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("127.0.0.0/8"),
	netip.MustParsePrefix("169.254.0.0/16"),
	netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.168.0.0/16"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("224.0.0.0/4"),
	netip.MustParsePrefix("240.0.0.0/4"),
	// IPv6 unspecified/loopback, discard, low-prefix special use,
	// documentation, 6to4, unique-local, link-local, and multicast ranges.
	netip.MustParsePrefix("::/128"),
	netip.MustParsePrefix("::1/128"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("100::/64"),
	netip.MustParsePrefix("2001::/23"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"),
	netip.MustParsePrefix("fc00::/7"),
	netip.MustParsePrefix("fe80::/10"),
	netip.MustParsePrefix("ff00::/8"),
}

func isPublicAddress(address netip.Addr) bool {
	if !address.IsValid() || !address.IsGlobalUnicast() || address.IsPrivate() || address.IsLoopback() || address.IsUnspecified() || address.IsLinkLocalUnicast() || address.IsLinkLocalMulticast() || address.IsMulticast() {
		return false
	}
	for _, prefix := range prohibitedPrefixes {
		if prefix.Contains(address) {
			return false
		}
	}
	return true
}
