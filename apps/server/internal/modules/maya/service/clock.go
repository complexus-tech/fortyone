package maya

import "time"

// Clock supplies the decision time for Maya scheduling operations.
// Implementations must be safe for concurrent use.
type Clock interface {
	Now() time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now()
}
