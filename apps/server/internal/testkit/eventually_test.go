package testkit

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestEventuallyCompletesAfterObservedStateChange(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	attempts := 0
	err := Eventually(ctx, time.Millisecond, func(context.Context) (Observation, error) {
		attempts++
		return Observation{
			Complete: attempts == 3,
			Detail:   "records=" + strconv.Itoa(attempts),
		}, nil
	})
	if err != nil {
		t.Fatalf("wait for observed state: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("observations = %d, want 3", attempts)
	}
}

func TestEventuallyRequiresBoundedInputs(t *testing.T) {
	t.Parallel()

	observer := func(context.Context) (Observation, error) { return Observation{Complete: true}, nil }
	if err := Eventually(context.Background(), time.Millisecond, observer); !errors.Is(err, ErrEventuallyDeadlineRequired) {
		t.Fatalf("missing-deadline error = %v", err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	if err := Eventually(ctx, 0, observer); !errors.Is(err, ErrEventuallyIntervalInvalid) {
		t.Fatalf("invalid-interval error = %v", err)
	}
	if err := Eventually(ctx, time.Millisecond, nil); !errors.Is(err, ErrEventuallyObserverRequired) {
		t.Fatalf("missing-observer error = %v", err)
	}
}

func TestEventuallyReportsLastBoundedObservation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	longDetail := strings.Repeat("queued\n", maxObservationRunes)
	err := Eventually(ctx, time.Hour, func(context.Context) (Observation, error) {
		cancel()
		return Observation{Detail: longDetail}, nil
	})
	if err == nil {
		t.Fatal("canceled eventually wait succeeded")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("eventually error = %v, want context cancellation", err)
	}
	var eventuallyError *EventuallyError
	if !errors.As(err, &eventuallyError) {
		t.Fatalf("eventually error type = %T", err)
	}
	if eventuallyError.Attempts != 1 {
		t.Fatalf("attempts = %d, want 1", eventuallyError.Attempts)
	}
	if strings.ContainsAny(eventuallyError.LastObservation, "\r\n\t") {
		t.Fatalf("last observation was not normalized: %q", eventuallyError.LastObservation)
	}
	if len([]rune(eventuallyError.LastObservation)) > maxObservationRunes+3 {
		t.Fatalf("last observation was not bounded: %d runes", len([]rune(eventuallyError.LastObservation)))
	}
}

func TestEventuallyStopsOnObserverError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("provider is unavailable")
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	err := Eventually(ctx, time.Millisecond, func(context.Context) (Observation, error) {
		return Observation{}, wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("observer error = %v, want %v", err, wantErr)
	}
}
