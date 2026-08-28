package clock_test

import (
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/clock"
)

func TestSystemImplementsClock(t *testing.T) {
	t.Parallel()

	var source clock.Clock = clock.System{}
	before := time.Now()
	now := source.Now()
	after := time.Now()
	if now.Before(before) || now.After(after) {
		t.Fatalf("system clock returned %v outside [%v, %v]", now, before, after)
	}
}
