package fortyone

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const maxRetryDrainBytes = 32 * 1024

// RetryPolicy controls retries for GET and HEAD only. A zero-value policy uses
// three total attempts with bounded full-jitter backoff. Set MaxAttempts to 1
// to disable retries.
type RetryPolicy struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

type normalizedRetryPolicy struct {
	maxAttempts int
	baseDelay   time.Duration
	maxDelay    time.Duration
	now         func() time.Time
	random      func() float64
	sleep       func(context.Context, time.Duration) error
}

func (policy RetryPolicy) normalized() (normalizedRetryPolicy, error) {
	if policy.MaxAttempts == 0 {
		policy.MaxAttempts = 3
	}
	if policy.BaseDelay == 0 {
		policy.BaseDelay = 250 * time.Millisecond
	}
	if policy.MaxDelay == 0 {
		policy.MaxDelay = 5 * time.Second
	}
	if policy.MaxAttempts < 1 || policy.MaxAttempts > 10 {
		return normalizedRetryPolicy{}, errors.New("retry max attempts must be from 1 through 10")
	}
	if policy.BaseDelay <= 0 || policy.MaxDelay < policy.BaseDelay || policy.MaxDelay > time.Minute {
		return normalizedRetryPolicy{}, errors.New("retry delays must be positive, ordered, and no greater than one minute")
	}
	return normalizedRetryPolicy{
		maxAttempts: policy.MaxAttempts,
		baseDelay:   policy.BaseDelay,
		maxDelay:    policy.MaxDelay,
		now:         time.Now,
		random:      rand.Float64,
		sleep:       sleepContext,
	}, nil
}

type retryTransport struct {
	next   http.RoundTripper
	policy normalizedRetryPolicy
}

func newRetryTransport(next http.RoundTripper, policy normalizedRetryPolicy) http.RoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}
	return &retryTransport{next: next, policy: policy}
}

func (transport *retryTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if (request.Method != http.MethodGet && request.Method != http.MethodHead) || request.Body != nil {
		return transport.next.RoundTrip(request)
	}
	for attempt := 1; ; attempt++ {
		response, err := transport.next.RoundTrip(request.Clone(request.Context()))
		if err == nil {
			if attempt >= transport.policy.maxAttempts || !retryableStatus(response.StatusCode) {
				return response, nil
			}
			delay, retry := transport.retryDelay(response, attempt)
			if !retry {
				return response, nil
			}
			discardResponse(response)
			if err := transport.policy.sleep(request.Context(), delay); err != nil {
				return nil, err
			}
			continue
		}
		if response != nil {
			discardResponse(response)
		}
		if attempt >= transport.policy.maxAttempts || request.Context().Err() != nil || !retryableNetworkError(err) {
			return nil, err
		}
		if err := transport.policy.sleep(request.Context(), transport.backoff(attempt)); err != nil {
			return nil, err
		}
	}
}

func (transport *retryTransport) retryDelay(response *http.Response, attempt int) (time.Duration, bool) {
	value := strings.TrimSpace(response.Header.Get("Retry-After"))
	if value == "" {
		return transport.backoff(attempt), true
	}
	if seconds, err := strconv.ParseUint(value, 10, 31); err == nil {
		delay := time.Duration(seconds) * time.Second
		return delay, delay <= transport.policy.maxDelay
	}
	when, err := http.ParseTime(value)
	if err != nil {
		return transport.backoff(attempt), true
	}
	delay := when.Sub(transport.policy.now())
	if delay < 0 {
		delay = 0
	}
	return delay, delay <= transport.policy.maxDelay
}

func (transport *retryTransport) backoff(attempt int) time.Duration {
	ceiling := transport.policy.baseDelay
	for step := 1; step < attempt && ceiling < transport.policy.maxDelay; step++ {
		if ceiling > transport.policy.maxDelay/2 {
			ceiling = transport.policy.maxDelay
			break
		}
		ceiling *= 2
	}
	if ceiling > transport.policy.maxDelay {
		ceiling = transport.policy.maxDelay
	}
	random := transport.policy.random()
	if random < 0 {
		random = 0
	}
	if random > 1 {
		random = 1
	}
	return time.Duration(random * float64(ceiling))
}

func retryableStatus(status int) bool {
	return status == http.StatusTooManyRequests || status == http.StatusServiceUnavailable
}

func retryableNetworkError(err error) bool {
	var networkError net.Error
	return errors.As(err, &networkError)
}

func discardResponse(response *http.Response) {
	if response == nil || response.Body == nil {
		return
	}
	_, _ = io.CopyN(io.Discard, response.Body, maxRetryDrainBytes)
	_ = response.Body.Close()
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return fmt.Errorf("wait to retry FortyOne API request: %w", ctx.Err())
	case <-timer.C:
		return nil
	}
}
