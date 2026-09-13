package slackdomain

import (
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)

var ErrInternalAlertLeaseLost = errors.New("internal Slack alert lease lost")

// InternalAlert is an operations-only delivery; its source workspace must never
// select the Slack installation or destination channel.
type InternalAlert struct {
	ID         uuid.UUID
	LeaseToken uuid.UUID
	DedupeKey  string
	Kind       string
	Payload    json.RawMessage
	Attempts   int
}
