package secretbox

import (
	"bytes"
	"testing"
)

func TestBoxRoundTrip(t *testing.T) {
	t.Parallel()

	box, err := newWithReader("test-secret", bytes.NewReader(make([]byte, 64)))
	if err != nil {
		t.Fatalf("newWithReader() error = %v", err)
	}
	sealed, err := box.Seal([]byte(`{"accessToken":"xoxb-secret"}`))
	if err != nil {
		t.Fatalf("Seal() error = %v", err)
	}
	if sealed == `{"accessToken":"xoxb-secret"}` {
		t.Fatal("Seal() returned plaintext")
	}
	opened, err := box.Open(sealed)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if opened.Version != CurrentVersion() {
		t.Fatalf("Open() version = %d, want %d", opened.Version, CurrentVersion())
	}
	if got, want := string(opened.Plaintext), `{"accessToken":"xoxb-secret"}`; got != want {
		t.Fatalf("Open() plaintext = %q, want %q", got, want)
	}
}

func TestBoxOpensLegacyPlaintext(t *testing.T) {
	t.Parallel()

	box, err := New("test-secret")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	opened, err := box.Open("xoxb-legacy")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if opened.Version != 0 || string(opened.Plaintext) != "xoxb-legacy" {
		t.Fatalf("Open() = %#v, want legacy plaintext", opened)
	}
}

func TestBoxRejectsTampering(t *testing.T) {
	t.Parallel()

	box, err := newWithReader("test-secret", bytes.NewReader(make([]byte, 64)))
	if err != nil {
		t.Fatalf("newWithReader() error = %v", err)
	}
	sealed, err := box.Seal([]byte("secret"))
	if err != nil {
		t.Fatalf("Seal() error = %v", err)
	}
	tampered := sealed[:len(sealed)-1] + "A"
	if _, err := box.Open(tampered); err == nil {
		t.Fatal("Open() error = nil, want tamper rejection")
	}
}
