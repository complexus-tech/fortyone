package safehttp

import (
	"context"
	"errors"
	"net/netip"
	"testing"
)

type resolverStub struct {
	addresses []netip.Addr
	err       error
	host      string
}

func (resolver *resolverStub) LookupNetIP(_ context.Context, _ string, host string) ([]netip.Addr, error) {
	resolver.host = host
	return append([]netip.Addr(nil), resolver.addresses...), resolver.err
}

func TestResolveAcceptsOnlyCanonicalPublicHTTPS(t *testing.T) {
	t.Parallel()
	resolver := &resolverStub{addresses: []netip.Addr{
		netip.MustParseAddr("2606:4700:4700::1111"),
		netip.MustParseAddr("1.1.1.1"),
	}}
	target, err := Resolve(context.Background(), resolver, " https://WEBHOOK.Example./events?version=1 ")
	if err != nil {
		t.Fatalf("resolve endpoint: %v", err)
	}
	if resolver.host != "webhook.example" {
		t.Fatalf("resolved host = %q", resolver.host)
	}
	if target.URL.String() != "https://webhook.example/events?version=1" {
		t.Fatalf("canonical URL = %q", target.URL.String())
	}
	if target.Port != "443" || len(target.Addresses) != 2 {
		t.Fatalf("target = %+v", target)
	}
}

func TestResolveRejectsUnsafeEndpointShapes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		url     string
		wantErr error
	}{
		{name: "plaintext", url: "http://example.com/hook", wantErr: ErrInsecureScheme},
		{name: "userinfo", url: "https://token@example.com/hook", wantErr: ErrCredentialsInURL},
		{name: "fragment", url: "https://example.com/hook#secret", wantErr: ErrFragmentInURL},
		{name: "custom port", url: "https://example.com:8443/hook", wantErr: ErrUnsupportedPort},
		{name: "IP literal", url: "https://1.1.1.1/hook", wantErr: ErrIPAddressHost},
		{name: "single label", url: "https://localhost/hook", wantErr: ErrInvalidEndpoint},
		{name: "unicode host", url: "https://éxample.com/hook", wantErr: ErrInvalidEndpoint},
		{name: "empty label", url: "https://example..com/hook", wantErr: ErrInvalidEndpoint},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := Resolve(context.Background(), &resolverStub{addresses: []netip.Addr{netip.MustParseAddr("1.1.1.1")}}, test.url)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Resolve() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestResolveRejectsAnyUnsafeDNSAnswer(t *testing.T) {
	t.Parallel()
	resolver := &resolverStub{addresses: []netip.Addr{
		netip.MustParseAddr("1.1.1.1"),
		netip.MustParseAddr("169.254.169.254"),
	}}
	_, err := Resolve(context.Background(), resolver, "https://example.com/hook")
	if !errors.Is(err, ErrUnsafeAddress) {
		t.Fatalf("Resolve() error = %v, want ErrUnsafeAddress", err)
	}
}

func TestPublicAddressPolicy(t *testing.T) {
	t.Parallel()
	tests := []struct {
		address string
		public  bool
	}{
		{address: "1.1.1.1", public: true},
		{address: "2606:4700:4700::1111", public: true},
		{address: "10.0.0.1", public: false},
		{address: "100.64.0.1", public: false},
		{address: "127.0.0.1", public: false},
		{address: "169.254.169.254", public: false},
		{address: "192.0.2.1", public: false},
		{address: "198.18.0.1", public: false},
		{address: "203.0.113.1", public: false},
		{address: "::1", public: false},
		{address: "fc00::1", public: false},
		{address: "fe80::1", public: false},
		{address: "2001:db8::1", public: false},
	}
	for _, test := range tests {
		test := test
		t.Run(test.address, func(t *testing.T) {
			t.Parallel()
			if got := isPublicAddress(netip.MustParseAddr(test.address)); got != test.public {
				t.Fatalf("isPublicAddress(%s) = %t, want %t", test.address, got, test.public)
			}
		})
	}
}
