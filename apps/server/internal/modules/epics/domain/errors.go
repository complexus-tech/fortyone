package epicsdomain

import "errors"

// ErrNotImplemented is explicit because the current database has no epic
// aggregate or story relationship. Returning synthetic rows would make the
// API claim that non-existent durable records are real.
var ErrNotImplemented = errors.New("epics are not implemented")
