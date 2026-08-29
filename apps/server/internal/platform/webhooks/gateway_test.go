package webhooks

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGatewayPersistsBeforeDispatchAndQueuesOnlyReceiptIdentity(t *testing.T) {
	t.Parallel()
	var order []string
	raw := []byte("{\n  \"event\": \"created\"\n}")
	verified := testVerifiedDelivery()
	inbox := &inboxStub{
		register: func(_ context.Context, envelope Envelope, encrypted string, expires time.Time) (Record, bool, error) {
			order = append(order, "persist")
			if encrypted != "ciphertext" {
				t.Fatalf("encrypted payload = %q, want ciphertext", encrypted)
			}
			if !expires.Equal(testNow.Add(defaultRetention)) {
				t.Fatalf("payload expiry = %s", expires)
			}
			return recordFor(envelope, StatusPending), true, nil
		},
		markQueued: func(_ context.Context, _ uuid.UUID, queuedAt time.Time) error {
			order = append(order, "mark")
			if !queuedAt.Equal(testNow) {
				t.Fatalf("queued at = %s, want %s", queuedAt, testNow)
			}
			return nil
		},
	}
	gateway := newGatewayFixture(t, inbox,
		verifierFunc(func(_ context.Context, request SignedRequest) (VerifiedDelivery, error) {
			order = append(order, "verify")
			if request.ReceivedAt != testNow || string(request.Body) != string(raw) {
				t.Fatalf("verifier request changed: %#v", request)
			}
			request.Body[0] = 'X'
			return verified, nil
		}),
		protectorFunc(func(_ context.Context, binding PayloadBinding, payload []byte) (string, error) {
			order = append(order, "protect")
			if string(payload) != string(raw) {
				t.Fatalf("payload was mutated by verifier: %q", payload)
			}
			if binding.DeliveryID != verified.DeliveryID || binding.InstallationGeneration != verified.InstallationGeneration {
				t.Fatalf("unexpected payload binding: %#v", binding)
			}
			return "ciphertext", nil
		}),
		dispatcherFunc(func(_ context.Context, task Task) error {
			order = append(order, "dispatch")
			if task.InboxID == uuid.Nil || task.Provider != testProvider {
				t.Fatalf("unexpected queue task: %#v", task)
			}
			return nil
		}),
	)
	receipt, err := gateway.Receive(context.Background(), testProvider, SignedRequest{
		Method: "POST", RequestTarget: "/integrations/slack/events", Body: raw,
	})
	if err != nil {
		t.Fatalf("receive webhook: %v", err)
	}
	if !receipt.Created || !receipt.Queued || receipt.Status != StatusPending {
		t.Fatalf("unexpected receipt: %#v", receipt)
	}
	if !slices.Equal(order, []string{"verify", "protect", "persist", "dispatch", "mark"}) {
		t.Fatalf("operation order = %v", order)
	}
}

func TestGatewayDoesNotRedispatchTerminalDuplicate(t *testing.T) {
	t.Parallel()
	verified := testVerifiedDelivery()
	dispatched := false
	inbox := &inboxStub{
		register: func(_ context.Context, envelope Envelope, _ string, _ time.Time) (Record, bool, error) {
			return recordFor(envelope, StatusCompleted), false, nil
		},
	}
	gateway := newGatewayFixture(t, inbox,
		verifierFunc(func(context.Context, SignedRequest) (VerifiedDelivery, error) { return verified, nil }),
		protectorFunc(func(context.Context, PayloadBinding, []byte) (string, error) { return "ciphertext", nil }),
		dispatcherFunc(func(context.Context, Task) error { dispatched = true; return nil }),
	)
	receipt, err := gateway.Receive(context.Background(), testProvider, SignedRequest{Body: []byte("{}")})
	if err != nil {
		t.Fatalf("receive terminal duplicate: %v", err)
	}
	if receipt.Created || receipt.Queued || receipt.Status != StatusCompleted || dispatched {
		t.Fatalf("terminal duplicate was redispatched: receipt=%#v dispatched=%v", receipt, dispatched)
	}
}

func TestGatewayRejectsConflictingDuplicateIdentity(t *testing.T) {
	t.Parallel()
	verified := testVerifiedDelivery()
	inbox := &inboxStub{
		register: func(_ context.Context, envelope Envelope, _ string, _ time.Time) (Record, bool, error) {
			envelope.WorkspaceID = uuid.New()
			return recordFor(envelope, StatusPending), false, nil
		},
	}
	gateway := newGatewayFixture(t, inbox,
		verifierFunc(func(context.Context, SignedRequest) (VerifiedDelivery, error) { return verified, nil }),
		protectorFunc(func(context.Context, PayloadBinding, []byte) (string, error) { return "ciphertext", nil }),
		dispatcherFunc(func(context.Context, Task) error { t.Fatal("conflict must not dispatch"); return nil }),
	)
	_, err := gateway.Receive(context.Background(), testProvider, SignedRequest{Body: []byte("{}")})
	if !errors.Is(err, ErrDeliveryConflict) {
		t.Fatalf("conflicting duplicate error = %v, want ErrDeliveryConflict", err)
	}
}

func TestGatewayClassifiesVerifierErrorsWithoutLeakingProviderDetails(t *testing.T) {
	t.Parallel()
	secret := "signed-payload-secret"
	gateway := newGatewayFixture(t, &inboxStub{},
		verifierFunc(func(context.Context, SignedRequest) (VerifiedDelivery, error) {
			return VerifiedDelivery{}, errors.New("provider rejected " + secret)
		}),
		protectorFunc(func(context.Context, PayloadBinding, []byte) (string, error) { return "ciphertext", nil }),
		dispatcherFunc(func(context.Context, Task) error { return nil }),
	)
	_, err := gateway.Receive(context.Background(), testProvider, SignedRequest{Body: []byte("{}")})
	if !errors.Is(err, ErrVerificationFailed) || strings.Contains(err.Error(), secret) {
		t.Fatalf("verification error = %q", err)
	}
}

func TestGatewayPreservesRetryableVerifierOutagesWithoutLeakingCauses(t *testing.T) {
	t.Parallel()
	secret := "database-password-in-cause"
	gateway := newGatewayFixture(t, &inboxStub{},
		verifierFunc(func(context.Context, SignedRequest) (VerifiedDelivery, error) {
			return VerifiedDelivery{}, fmt.Errorf("%w: %s", ErrVerificationUnavailable, secret)
		}),
		protectorFunc(func(context.Context, PayloadBinding, []byte) (string, error) { return "ciphertext", nil }),
		dispatcherFunc(func(context.Context, Task) error { return nil }),
	)
	_, err := gateway.Receive(context.Background(), testProvider, SignedRequest{Body: []byte("{}")})
	if !errors.Is(err, ErrVerificationUnavailable) || strings.Contains(err.Error(), secret) {
		t.Fatalf("verification outage error = %q", err)
	}
}

func TestGatewayBoundsRequestBeforeProviderWork(t *testing.T) {
	t.Parallel()
	called := false
	gateway := newGatewayFixture(t, &inboxStub{},
		verifierFunc(func(context.Context, SignedRequest) (VerifiedDelivery, error) {
			called = true
			return testVerifiedDelivery(), nil
		}),
		protectorFunc(func(context.Context, PayloadBinding, []byte) (string, error) { return "ciphertext", nil }),
		dispatcherFunc(func(context.Context, Task) error { return nil }),
	)
	_, err := gateway.Receive(context.Background(), testProvider, SignedRequest{Body: make([]byte, defaultMaxBodyBytes+1)})
	if !errors.Is(err, ErrPayloadTooLarge) || called {
		t.Fatalf("oversized error/call = %v/%v", err, called)
	}
}

func TestGatewayAcknowledgesExplicitlyIgnoredDeliveryWithoutRetention(t *testing.T) {
	t.Parallel()
	gateway := newGatewayFixture(t, &inboxStub{},
		verifierFunc(func(context.Context, SignedRequest) (VerifiedDelivery, error) {
			return VerifiedDelivery{}, ErrDeliveryIgnored
		}),
		protectorFunc(func(context.Context, PayloadBinding, []byte) (string, error) {
			t.Fatal("ignored payload protected")
			return "", nil
		}),
		dispatcherFunc(func(context.Context, Task) error { t.Fatal("ignored payload dispatched"); return nil }),
	)
	receipt, err := gateway.Receive(context.Background(), testProvider, SignedRequest{Body: []byte("{}")})
	if err != nil || !receipt.Ignored {
		t.Fatalf("ignored receipt = %#v, error = %v", receipt, err)
	}
}
