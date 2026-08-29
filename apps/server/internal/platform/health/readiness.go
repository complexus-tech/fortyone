package health

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

const defaultCheckTimeout = 2 * time.Second

// Phase describes whether the process is able to accept new traffic.
type Phase string

const (
	PhaseStarting Phase = "starting"
	PhaseReady    Phase = "ready"
	PhaseDraining Phase = "draining"
	PhaseFailed   Phase = "failed"
)

// Dependency is a required service that must be reachable before the API is
// ready. Names are emitted by the readiness endpoint and therefore must be
// stable, low-cardinality identifiers rather than connection details.
type Dependency struct {
	Name  string
	Check func(context.Context) error
}

// Report is the safe, operator-facing readiness result.
type Report struct {
	Status string            `json:"status"`
	Phase  Phase             `json:"phase"`
	Checks map[string]string `json:"checks"`
}

// Readiness combines process lifecycle state with required dependency checks.
// The lifecycle phase is lock-free because it is read on every readiness probe.
type Readiness struct {
	phase        atomic.Uint32
	checkTimeout time.Duration
	dependencies []Dependency
}

// NewReadiness validates and constructs process readiness state. The returned
// state starts in PhaseStarting and cannot report ready until MarkReady is
// called by the process supervisor.
func NewReadiness(checkTimeout time.Duration, dependencies ...Dependency) (*Readiness, error) {
	if checkTimeout == 0 {
		checkTimeout = defaultCheckTimeout
	}
	if checkTimeout < 0 {
		return nil, errors.New("readiness check timeout must be positive")
	}

	seen := make(map[string]struct{}, len(dependencies))
	checked := make([]Dependency, 0, len(dependencies))
	for _, dependency := range dependencies {
		name := strings.TrimSpace(dependency.Name)
		if name == "" {
			return nil, errors.New("readiness dependency name is required")
		}
		if dependency.Check == nil {
			return nil, fmt.Errorf("readiness dependency %q has no check", name)
		}
		if _, exists := seen[name]; exists {
			return nil, fmt.Errorf("readiness dependency %q is registered more than once", name)
		}
		seen[name] = struct{}{}
		checked = append(checked, Dependency{Name: name, Check: dependency.Check})
	}

	readiness := &Readiness{
		checkTimeout: checkTimeout,
		dependencies: checked,
	}
	readiness.setPhase(PhaseStarting)
	return readiness, nil
}

// MarkReady allows dependency checks to determine readiness.
func (r *Readiness) MarkReady() {
	if r != nil {
		r.setPhase(PhaseReady)
	}
}

// MarkDraining immediately removes the process from service before shutdown.
func (r *Readiness) MarkDraining() {
	if r != nil {
		r.setPhase(PhaseDraining)
	}
}

// MarkFailed records a terminal supervisor failure.
func (r *Readiness) MarkFailed() {
	if r != nil {
		r.setPhase(PhaseFailed)
	}
}

// Phase returns the current lifecycle phase.
func (r *Readiness) Phase() Phase {
	if r == nil {
		return PhaseFailed
	}
	return decodePhase(r.phase.Load())
}

// Report checks required dependencies only while the supervisor is accepting
// traffic. Dependency error text is intentionally not exposed to clients.
func (r *Readiness) Report(ctx context.Context) Report {
	if r == nil {
		return Report{Status: "not_ready", Phase: PhaseFailed, Checks: map[string]string{}}
	}
	phase := r.Phase()
	report := Report{
		Status: "not_ready",
		Phase:  phase,
		Checks: make(map[string]string, len(r.dependencies)),
	}
	if phase != PhaseReady {
		for _, dependency := range r.dependencies {
			report.Checks[dependency.Name] = "not_checked"
		}
		return report
	}

	if ctx == nil {
		ctx = context.Background()
	}
	checkCtx, cancel := context.WithTimeout(ctx, r.checkTimeout)
	defer cancel()

	ready := true
	for _, dependency := range r.dependencies {
		if err := dependency.Check(checkCtx); err != nil {
			report.Checks[dependency.Name] = "unavailable"
			ready = false
			continue
		}
		report.Checks[dependency.Name] = "ready"
	}
	if ready {
		report.Status = "ready"
	}
	return report
}

func (r *Readiness) setPhase(phase Phase) {
	r.phase.Store(encodePhase(phase))
}

func encodePhase(phase Phase) uint32 {
	switch phase {
	case PhaseStarting:
		return 0
	case PhaseReady:
		return 1
	case PhaseDraining:
		return 2
	case PhaseFailed:
		return 3
	default:
		return 3
	}
}

func decodePhase(phase uint32) Phase {
	switch phase {
	case 0:
		return PhaseStarting
	case 1:
		return PhaseReady
	case 2:
		return PhaseDraining
	default:
		return PhaseFailed
	}
}
