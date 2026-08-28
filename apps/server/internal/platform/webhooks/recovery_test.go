package webhooks

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/google/uuid"
)

func TestRecoverContinuesAfterDispatchFailureAndReleasesExactGeneration(t *testing.T) {
	t.Parallel()
	firstID := uuid.New()
	secondID := uuid.New()
	failedTask := firstID
	var released []Record
	inbox := &inboxStub{
		claim: func(_ context.Context, provider integrations.ProviderKey, policy RecoveryPolicy, now time.Time) ([]Record, error) {
			if provider != testProvider || now != testNow || policy.ClaimLimit != 100 {
				t.Fatalf("unexpected recovery claim: %s %#v %s", provider, policy, now)
			}
			return []Record{
				{ID: firstID, RecoveryGeneration: 7},
				{ID: secondID, RecoveryGeneration: 11},
			}, nil
		},
		release: func(_ context.Context, id uuid.UUID, generation int32, _ time.Time) error {
			released = append(released, Record{ID: id, RecoveryGeneration: generation})
			return nil
		},
		markQueued: func(_ context.Context, id uuid.UUID, _ time.Time) error {
			if id != secondID {
				t.Fatalf("marked failed task queued: %s", id)
			}
			return nil
		},
	}
	gateway := newGatewayFixture(t, inbox,
		verifierFunc(func(context.Context, SignedRequest) (VerifiedDelivery, error) { return testVerifiedDelivery(), nil }),
		protectorFunc(func(context.Context, PayloadBinding, []byte) (string, error) { return "ciphertext", nil }),
		dispatcherFunc(func(_ context.Context, task Task) error {
			if task.InboxID == failedTask {
				return errors.New("queue unavailable")
			}
			return nil
		}),
	)
	report, err := gateway.Recover(context.Background(), testProvider, DefaultRecoveryPolicy())
	if !errors.Is(err, ErrDispatchUnavailable) {
		t.Fatalf("recover error = %v, want ErrDispatchUnavailable", err)
	}
	if report != (RecoveryReport{Claimed: 2, Dispatched: 1, Released: 1}) {
		t.Fatalf("recovery report = %#v", report)
	}
	if len(released) != 1 || released[0].ID != firstID || released[0].RecoveryGeneration != 7 {
		t.Fatalf("released recovery claims = %#v", released)
	}
}

func TestRecoveryPolicyRejectsUnsafeBounds(t *testing.T) {
	t.Parallel()
	tests := []RecoveryPolicy{
		{},
		{ClaimLimit: 501, MaxAttempts: 1, PendingAge: time.Second, FailedAge: time.Second, ProcessingLease: time.Second, RecoveryBaseDelay: time.Second},
		{ClaimLimit: 1, MaxAttempts: 101, PendingAge: time.Second, FailedAge: time.Second, ProcessingLease: time.Second, RecoveryBaseDelay: time.Second},
		{ClaimLimit: 1, MaxAttempts: 1, PendingAge: time.Millisecond, FailedAge: time.Second, ProcessingLease: time.Second, RecoveryBaseDelay: time.Second},
	}
	for _, policy := range tests {
		if err := policy.Validate(); !errors.Is(err, ErrInvalidRequest) {
			t.Fatalf("policy %#v error = %v, want ErrInvalidRequest", policy, err)
		}
	}
}
