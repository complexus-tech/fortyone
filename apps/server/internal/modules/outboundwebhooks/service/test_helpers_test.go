package outboundwebhooksservice

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/internal/platform/safehttp"
	"github.com/google/uuid"
)

type testClock struct {
	values []time.Time
	index  int
}

func (clock *testClock) Now() time.Time {
	if len(clock.values) == 0 {
		return time.Time{}
	}
	index := clock.index
	if index >= len(clock.values) {
		index = len(clock.values) - 1
	} else {
		clock.index++
	}
	return clock.values[index]
}

type testIDs struct {
	values []uuid.UUID
	err    error
	index  int
}

func (ids *testIDs) NewUUID() (uuid.UUID, error) {
	if ids.err != nil {
		return uuid.Nil, ids.err
	}
	if ids.index >= len(ids.values) {
		return uuid.Nil, io.EOF
	}
	value := ids.values[ids.index]
	ids.index++
	return value, nil
}

func newTestSecretManager(t *testing.T) *SecretManager {
	t.Helper()
	vault, err := credentialvault.New(credentialvault.Config{
		Active: credentialvault.KeyRef{ID: "outbound-test", Version: 1},
		Keys: []credentialvault.Key{{
			Ref:      credentialvault.KeyRef{ID: "outbound-test", Version: 1},
			Material: bytes.Repeat([]byte{0x4f}, 32),
		}},
	})
	if err != nil {
		t.Fatalf("create test credential vault: %v", err)
	}
	manager, err := newSecretManager(vault, bytes.NewReader(bytes.Repeat([]byte{0x61}, 512)))
	if err != nil {
		t.Fatalf("create test secret manager: %v", err)
	}
	return manager
}

type deliveryRepositoryStub struct {
	delivery    outboundwebhooksdomain.ClaimedDelivery
	claimErr    error
	completed   []outboundwebhooksdomain.DeliveryAttempt
	completeErr error
	recovered   int64
	recoverErr  error
}

func (repository *deliveryRepositoryStub) ClaimNextDelivery(
	_ context.Context,
	leaseToken uuid.UUID,
	_, leaseExpiresAt time.Time,
) (outboundwebhooksdomain.ClaimedDelivery, error) {
	if repository.claimErr != nil {
		return outboundwebhooksdomain.ClaimedDelivery{}, repository.claimErr
	}
	delivery := repository.delivery
	delivery.LeaseToken = leaseToken
	delivery.LeaseExpiresAt = leaseExpiresAt
	return delivery, nil
}

func (repository *deliveryRepositoryStub) CompleteAttempt(
	_ context.Context,
	attempt outboundwebhooksdomain.DeliveryAttempt,
	_, _ uuid.UUID,
) error {
	repository.completed = append(repository.completed, attempt)
	return repository.completeErr
}

func (repository *deliveryRepositoryStub) RecoverExpiredLeases(context.Context, time.Time) (int64, error) {
	return repository.recovered, repository.recoverErr
}

type httpClientStub struct {
	result  safehttp.Result
	err     error
	body    []byte
	headers http.Header
	url     string
}

func (client *httpClientStub) Do(_ context.Context, request *http.Request) (safehttp.Result, error) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return safehttp.Result{}, err
	}
	client.body = body
	client.headers = request.Header.Clone()
	client.url = request.URL.String()
	return client.result, client.err
}

type failingSecretVault struct{ err error }

func (vault failingSecretVault) Seal(credentialvault.Context, []byte) (string, error) {
	return "", vault.err
}

func (vault failingSecretVault) Open(credentialvault.Context, string) (credentialvault.Secret, error) {
	return credentialvault.Secret{}, vault.err
}
