package outboundwebhooksservice

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSignUsesStableIDTimestampAndExactBody(t *testing.T) {
	t.Parallel()
	deliveryID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	timestamp := time.Unix(1_787_920_000, 0).UTC()
	body := []byte("{\"value\":\"line one\\r\\nline two\"}")
	material := bytes.Repeat([]byte{0x52}, webhookSecretBytes)
	secret := []byte(webhookSecretPrefix + base64.StdEncoding.EncodeToString(material))
	headers, err := Sign(deliveryID, timestamp, body, secret)
	if err != nil {
		t.Fatalf("sign delivery: %v", err)
	}
	if headers.WebhookID != deliveryID.String() || headers.WebhookTimestamp != strconv.FormatInt(timestamp.Unix(), 10) {
		t.Fatalf("signature headers = %+v", headers)
	}
	parts := strings.Split(headers.WebhookSignature, ",")
	if len(parts) != 2 || parts[0] != SignatureVersion {
		t.Fatalf("signature header = %q", headers.WebhookSignature)
	}
	mac := hmac.New(sha256.New, material)
	_, _ = mac.Write([]byte(headers.WebhookID + "." + headers.WebhookTimestamp + "."))
	_, _ = mac.Write(body)
	want := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[1]), []byte(want)) {
		t.Fatalf("signature = %q, want %q", parts[1], want)
	}

	changed, err := Sign(deliveryID, timestamp, append(append([]byte(nil), body...), '\n'), secret)
	if err != nil {
		t.Fatalf("sign changed body: %v", err)
	}
	if changed.WebhookSignature == headers.WebhookSignature {
		t.Fatal("signature did not bind exact body bytes")
	}
}
