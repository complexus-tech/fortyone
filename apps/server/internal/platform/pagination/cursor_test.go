package pagination

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

type cursorFixture struct {
	CreatedAt string `json:"createdAt"`
	ID        string `json:"id"`
	Filter    string `json:"filter"`
}

func TestCursorCodecRoundTripAndTamperProtection(t *testing.T) {
	t.Parallel()

	codec := mustCursorCodec(t, SigningKey{ID: "current", Secret: []byte(strings.Repeat("a", 32))})
	want := cursorFixture{CreatedAt: "2026-08-28T10:00:00Z", ID: "story-42", Filter: "assigned"}

	token, err := codec.Encode(want)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	got, err := codec.Decode(token)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if got != want {
		t.Fatalf("Decode() = %#v, want %#v", got, want)
	}

	tampered := token[:len(token)-1] + "A"
	if _, err := codec.Decode(tampered); !errors.Is(err, ErrCursorSignature) {
		t.Fatalf("tampered Decode() error = %v, want ErrCursorSignature", err)
	}
}

func TestCursorCodecSupportsBoundedKeyRotation(t *testing.T) {
	t.Parallel()

	oldKey := SigningKey{ID: "old", Secret: []byte(strings.Repeat("o", 32))}
	newKey := SigningKey{ID: "new", Secret: []byte(strings.Repeat("n", 32))}
	oldCodec := mustCursorCodec(t, oldKey)
	oldToken, err := oldCodec.Encode(cursorFixture{ID: "story-42"})
	if err != nil {
		t.Fatalf("old Encode() error = %v", err)
	}

	rotated := mustCursorCodec(t, newKey, oldKey)
	if _, err := rotated.Decode(oldToken); err != nil {
		t.Fatalf("Decode(old token) error = %v", err)
	}
	newToken, err := rotated.Encode(cursorFixture{ID: "story-43"})
	if err != nil {
		t.Fatalf("new Encode() error = %v", err)
	}
	if !strings.HasPrefix(newToken, "v1.new.") {
		t.Fatalf("new token prefix = %q, want active key id", newToken)
	}

	withoutOldKey := mustCursorCodec(t, newKey)
	if _, err := withoutOldKey.Decode(oldToken); !errors.Is(err, ErrCursorSignature) {
		t.Fatalf("Decode(retired key token) error = %v, want ErrCursorSignature", err)
	}
}

func TestCursorCodecRejectsInvalidConfigurationAndInput(t *testing.T) {
	t.Parallel()

	if _, err := NewCursorCodec[cursorFixture](SigningKey{ID: "bad id", Secret: []byte(strings.Repeat("a", 32))}); !errors.Is(err, ErrInvalidCursorKey) {
		t.Fatalf("NewCursorCodec() error = %v, want ErrInvalidCursorKey", err)
	}
	if _, err := NewCursorCodec[cursorFixture](SigningKey{ID: "short", Secret: []byte("short")}); !errors.Is(err, ErrInvalidCursorKey) {
		t.Fatalf("NewCursorCodec() error = %v, want ErrInvalidCursorKey", err)
	}

	codec := mustCursorCodec(t, SigningKey{ID: "current", Secret: []byte(strings.Repeat("a", 32))})
	for _, token := range []string{"", "v2.current.payload.signature", "v1.current.invalid.signature", strings.Repeat("x", maximumCursorBytes+1)} {
		if _, err := codec.Decode(token); err == nil {
			t.Fatalf("Decode(%q) error = nil", token)
		}
	}
}

func TestDeriveSigningKeyIsPurposeSeparated(t *testing.T) {
	t.Parallel()
	root := []byte("test-root-secret")
	webhooks, err := DeriveSigningKey("v1", root, "outbound-webhooks")
	if err != nil {
		t.Fatalf("DeriveSigningKey(webhooks) error = %v", err)
	}
	stories, err := DeriveSigningKey("v1", root, "stories")
	if err != nil {
		t.Fatalf("DeriveSigningKey(stories) error = %v", err)
	}
	if bytes.Equal(webhooks.Secret, stories.Secret) || bytes.Equal(webhooks.Secret, root) {
		t.Fatal("derived cursor signing keys are not purpose-separated")
	}
	if _, err := DeriveSigningKey("v1", root, " outbound-webhooks"); !errors.Is(err, ErrInvalidCursorKey) {
		t.Fatalf("DeriveSigningKey(invalid purpose) error = %v", err)
	}
}

func TestCursorCodecRejectsUnknownPayloadFields(t *testing.T) {
	t.Parallel()

	type narrowCursor struct {
		ID string `json:"id"`
	}
	wideCodec := mustCursorCodec(t, SigningKey{ID: "current", Secret: []byte(strings.Repeat("a", 32))})
	token, err := wideCodec.Encode(cursorFixture{ID: "story-42", Filter: "assigned"})
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	narrowCodec, err := NewCursorCodec[narrowCursor](SigningKey{ID: "current", Secret: []byte(strings.Repeat("a", 32))})
	if err != nil {
		t.Fatalf("NewCursorCodec() error = %v", err)
	}
	if _, err := narrowCodec.Decode(token); !errors.Is(err, ErrInvalidCursor) {
		t.Fatalf("Decode() error = %v, want ErrInvalidCursor", err)
	}
}

func FuzzCursorCodecDecode(f *testing.F) {
	codec, err := NewCursorCodec[cursorFixture](SigningKey{ID: "current", Secret: []byte(strings.Repeat("a", 32))})
	if err != nil {
		f.Fatalf("NewCursorCodec() error = %v", err)
	}
	valid, err := codec.Encode(cursorFixture{ID: "story-42"})
	if err != nil {
		f.Fatalf("Encode() error = %v", err)
	}
	f.Add(valid)
	f.Add("")
	f.Add("v1.current.payload.signature")

	f.Fuzz(func(t *testing.T, token string) {
		_, _ = codec.Decode(token)
	})
}

func mustCursorCodec(t *testing.T, active SigningKey, previous ...SigningKey) CursorCodec[cursorFixture] {
	t.Helper()
	codec, err := NewCursorCodec[cursorFixture](active, previous...)
	if err != nil {
		t.Fatalf("NewCursorCodec() error = %v", err)
	}
	return codec
}
