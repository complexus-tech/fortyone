package outboundwebhooksservice

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

const SignatureVersion = "v1"

type SignatureHeaders struct {
	WebhookID        string
	WebhookTimestamp string
	WebhookSignature string
}

// Sign produces the Standard Webhooks-compatible signed content
// "<delivery-id>.<unix-seconds>.<exact-body>". The delivery ID is stable
// across automatic retries while the timestamp and signature are per attempt.
func Sign(deliveryID uuid.UUID, timestamp time.Time, exactBody, encodedSecret []byte) (SignatureHeaders, error) {
	if deliveryID == uuid.Nil || timestamp.IsZero() || len(exactBody) == 0 || len(exactBody) > 256<<10 {
		return SignatureHeaders{}, fmt.Errorf("outbound webhook signature input is invalid")
	}
	material, err := decodeSigningSecret(encodedSecret)
	if err != nil {
		return SignatureHeaders{}, err
	}
	defer clear(material)
	unixTimestamp := strconv.FormatInt(timestamp.UTC().Unix(), 10)
	mac := hmac.New(sha256.New, material)
	_, _ = mac.Write([]byte(deliveryID.String()))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write([]byte(unixTimestamp))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(exactBody)
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return SignatureHeaders{
		WebhookID:        deliveryID.String(),
		WebhookTimestamp: unixTimestamp,
		WebhookSignature: SignatureVersion + "," + signature,
	}, nil
}
