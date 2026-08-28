// Package clock defines the application-wide source of decision time.
package clock

import "time"

// Clock supplies the current time to behavior that must be deterministic in
// tests. Consumers should depend on this interface instead of time.Now.
type Clock interface {
	Now() time.Time
}

// System reads wall-clock time.
type System struct{}

// Now returns the current local wall-clock time.
func (System) Now() time.Time {
	return time.Now()
}
