package idempotency

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"
)

func FuzzParseKeyAndDigest(f *testing.F) {
	f.Add("0123456789abcdef")
	f.Add("request-key_0123456789abcdef")
	f.Add("short")
	f.Add(strings.Repeat("x", MaxKeyBytes+1))

	f.Fuzz(func(t *testing.T, raw string) {
		key, err := ParseKey(raw)
		if err != nil {
			return
		}
		if len(raw) < MinKeyBytes || len(raw) > MaxKeyBytes {
			t.Fatalf("ParseKey() accepted %d bytes", len(raw))
		}
		for index := range len(raw) {
			if raw[index] < 0x21 || raw[index] > 0x7e {
				t.Fatalf("ParseKey() accepted non-visible byte at %d", index)
			}
		}
		if got, want := key.digest(), sha256.Sum256([]byte(raw)); got != want {
			t.Fatalf("key digest = %x, want %x", got, want)
		}
		if formatted := fmt.Sprintf("%v", key); formatted != "[REDACTED]" || strings.Contains(formatted, raw) {
			t.Fatalf("formatted key was not redacted")
		}
	})
}

func FuzzHashRequestUsesExactBytes(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte("{}"))
	f.Add([]byte("{\"title\":\"story\"}\n"))

	f.Fuzz(func(t *testing.T, body []byte) {
		before := append([]byte(nil), body...)
		got := HashRequest(body)
		want := sha256.Sum256(body)
		if got != want {
			t.Fatalf("HashRequest() = %x, want %x", got, want)
		}
		if string(body) != string(before) {
			t.Fatal("HashRequest() mutated request bytes")
		}
	})
}
