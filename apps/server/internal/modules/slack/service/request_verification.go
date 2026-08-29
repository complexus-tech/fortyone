package slack

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (s *Service) VerifyRequest(rawBody []byte, headers http.Header) error {
	return verifySlackSignature(
		s.cfg.SigningSecret,
		s.clock.Now(),
		rawBody,
		headers.Get("X-Slack-Request-Timestamp"),
		headers.Get("X-Slack-Signature"),
	)
}

func verifySlackSignature(secret string, now time.Time, rawBody []byte, timestamp, signature string) error {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return ErrSlackSigningSecretNotConfigured
	}

	timestamp = strings.TrimSpace(timestamp)
	signature = strings.TrimSpace(signature)
	if timestamp == "" || signature == "" {
		return ErrSlackInvalidSignature
	}

	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return ErrSlackInvalidSignature
	}
	if math.Abs(float64(now.UTC().Unix()-ts)) > 300 {
		return ErrSlackRequestExpired
	}

	base := "v0:" + timestamp + ":" + string(rawBody)
	h := hmac.New(sha256.New, []byte(secret))
	_, _ = h.Write([]byte(base))
	expected := "v0=" + hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return ErrSlackInvalidSignature
	}
	return nil
}
