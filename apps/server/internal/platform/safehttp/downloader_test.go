package safehttp

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"testing"
	"time"
)

func TestNewDownloaderRejectsUnboundedConfiguration(t *testing.T) {
	t.Parallel()
	for _, config := range []Config{
		{Timeout: 31 * time.Second},
		{Timeout: time.Second, TLSHandshakeTimeout: 2 * time.Second},
		{MaxResponseBytes: maximumDownloadBytes + 1},
		{MaxResponseHeaderBytes: (64 << 10) + 1},
	} {
		if _, err := NewDownloader(config); !errors.Is(err, ErrUnsupportedRequest) {
			t.Fatalf("NewDownloader(%+v) error = %v, want ErrUnsupportedRequest", config, err)
		}
	}
}

func TestDownloaderRejectsPrivateTargetsBeforeDial(t *testing.T) {
	t.Parallel()
	resolver := downloadResolver{addresses: []netip.Addr{netip.MustParseAddr("127.0.0.1")}}
	dialed := false
	downloader, err := NewDownloader(Config{
		Resolver: resolver,
		DialContext: func(context.Context, string, string) (net.Conn, error) {
			dialed = true
			return nil, errors.New("unexpected dial")
		},
	})
	if err != nil {
		t.Fatalf("NewDownloader() error = %v", err)
	}
	if _, err := downloader.Download(context.Background(), "https://images.example.com/avatar.png"); !errors.Is(err, ErrUnsafeAddress) {
		t.Fatalf("Download() error = %v, want ErrUnsafeAddress", err)
	}
	if dialed {
		t.Fatal("private target reached the dialer")
	}
}

type downloadResolver struct {
	addresses []netip.Addr
}

func (resolver downloadResolver) LookupNetIP(context.Context, string, string) ([]netip.Addr, error) {
	return append([]netip.Addr(nil), resolver.addresses...), nil
}
