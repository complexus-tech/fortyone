//go:build integration && chatload

package api

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Model the active Messages screen's first loaded page. Each authorized hint
// coalesces for 150ms, then refetches its conversation list, detail, history and
// latest page, matching realtime.ts/queries.ts/use-chat.ts. Other platform users
// remain connected to SSE without mounting these active query observers.
type loadReconciliation struct {
	runner       *loadRunner
	hints        []chan struct{}
	active       atomic.Bool
	received     atomic.Int64
	cycles       atomic.Int64
	heldCycles   atomic.Int64
	synchronized bool
	closing      chan struct{}
	once         sync.Once
	loops        sync.WaitGroup
	requests     sync.WaitGroup
}

func newLoadReconciliation(t *testing.T, runner *loadRunner, users int) *loadReconciliation {
	t.Helper()
	percentage := 20
	if raw := os.Getenv("CHAT_LOAD_ACTIVE_CHAT_PERCENT"); raw != "" {
		var err error
		percentage, err = strconv.Atoi(raw)
		require.NoError(t, err)
	}
	require.GreaterOrEqual(t, percentage, 0)
	require.LessOrEqual(t, percentage, 100)
	replay := &loadReconciliation{
		runner:       runner,
		hints:        make([]chan struct{}, users*percentage/100),
		closing:      make(chan struct{}),
		synchronized: os.Getenv("CHAT_LOAD_SYNCHRONIZED_CHAT_START") == "true",
	}
	replay.active.Store(!replay.synchronized)
	for user := range replay.hints {
		replay.hints[user] = make(chan struct{}, 1)
		replay.loops.Add(1)
		go replay.run(user)
	}
	return replay
}

func (r *loadReconciliation) start() {
	if !r.synchronized {
		return
	}
	r.active.Store(true)
	for user := range r.hints {
		r.connected(user)
	}
}

// Initial query observers mount as each browser arrives during the ramp.
// Reopened streams also reconcile, matching EventSource's onopen behavior.
func (r *loadReconciliation) connected(user int) {
	if user >= len(r.hints) || !r.active.Load() {
		return
	}
	select {
	case r.hints[user] <- struct{}{}:
	default:
	}
}

func (r *loadReconciliation) hint(user int) {
	if user >= len(r.hints) || !r.active.Load() {
		return
	}
	r.received.Add(1)
	select {
	case r.hints[user] <- struct{}{}:
	default:
	}
}

func (r *loadReconciliation) stop() {
	r.once.Do(func() {
		r.active.Store(false)
		close(r.closing)
		r.loops.Wait()
		r.requests.Wait()
	})
}

func (r *loadReconciliation) run(user int) {
	defer r.loops.Done()
	for {
		select {
		case <-r.closing:
			return
		case <-r.hints[user]:
		}
		timer := time.NewTimer(150 * time.Millisecond)
		select {
		case <-r.closing:
			timer.Stop()
			return
		case <-timer.C:
		}
		// All hints inside this coalescing window invalidate the same room.
		select {
		case <-r.hints[user]:
		default:
		}
		room := r.runner.rooms[r.runner.people[user].Team].String()
		reads := []struct{ kind, path string }{
			{"reactive_conversations", "conversations?limit=50"},
			{"reactive_conversation", "conversations/" + room},
			{"reactive_history", "conversations/" + room + "/messages?limit=50"},
			{"reactive_latest", "conversations/" + room + "/messages?limit=50"},
		}
		r.cycles.Add(1)
		if r.runner.steady.Load() {
			r.heldCycles.Add(1)
		}
		for _, read := range reads {
			r.requests.Add(1)
			go func() {
				defer r.requests.Done()
				if err := r.runner.request(user, http.MethodGet, r.runner.path(read.path), read.kind, nil, nil); err != nil {
					r.runner.metrics.operationError(fmt.Errorf("reactive reconciliation: %w", err))
				}
			}()
		}
	}
}
