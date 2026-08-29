package testkit

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const maxObservationRunes = 512

var (
	// ErrEventuallyDeadlineRequired prevents an accidentally unbounded poll.
	ErrEventuallyDeadlineRequired = errors.New("eventually context must have a deadline")
	// ErrEventuallyIntervalInvalid prevents a busy loop or disabled timer.
	ErrEventuallyIntervalInvalid = errors.New("eventually interval must be positive")
	// ErrEventuallyObserverRequired identifies a missing condition callback.
	ErrEventuallyObserverRequired = errors.New("eventually observer is required")
)

// Observation is one safe, human-readable view of an eventually consistent
// condition. Detail should identify state, counts, or safe fingerprints; it
// must not contain credentials or raw customer/provider payloads.
type Observation struct {
	Complete bool
	Detail   string
}

// Observer evaluates an eventually consistent condition. Returning an error
// stops polling immediately; transient states should instead be represented by
// an incomplete Observation with safe diagnostic detail.
type Observer func(context.Context) (Observation, error)

// Eventually polls immediately and then at interval until the observation is
// complete or the required context deadline/cancellation ends the wait. It does
// not expose a general sleep primitive and always reports the last bounded
// observation when the condition is not met.
func Eventually(ctx context.Context, interval time.Duration, observe Observer) error {
	if ctx == nil {
		return ErrEventuallyDeadlineRequired
	}
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		return ErrEventuallyDeadlineRequired
	}
	if interval <= 0 {
		return ErrEventuallyIntervalInvalid
	}
	if observe == nil {
		return ErrEventuallyObserverRequired
	}

	attempts := 0
	lastObservation := "<none>"
	for {
		if err := ctx.Err(); err != nil {
			return &EventuallyError{
				Cause:           err,
				Attempts:        attempts,
				LastObservation: lastObservation,
			}
		}

		observation, err := observe(ctx)
		attempts++
		if err != nil {
			return fmt.Errorf("observe eventually condition: %w", err)
		}
		lastObservation = boundedObservation(observation.Detail)
		if observation.Complete {
			return nil
		}

		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return &EventuallyError{
				Cause:           ctx.Err(),
				Attempts:        attempts,
				LastObservation: lastObservation,
			}
		case <-timer.C:
		}
	}
}

// EventuallyError describes a bounded poll that ended before completion.
type EventuallyError struct {
	Cause           error
	Attempts        int
	LastObservation string
}

func (e *EventuallyError) Error() string {
	return fmt.Sprintf(
		"eventually condition not met after %d observations: %v; last observation: %s",
		e.Attempts,
		e.Cause,
		e.LastObservation,
	)
}

func (e *EventuallyError) Unwrap() error {
	return e.Cause
}

func boundedObservation(detail string) string {
	detail = strings.Join(strings.Fields(detail), " ")
	if detail == "" {
		return "<empty>"
	}

	runes := []rune(detail)
	if len(runes) <= maxObservationRunes {
		return detail
	}
	return string(runes[:maxObservationRunes]) + "..."
}
