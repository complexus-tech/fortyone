package usersdomain

import "github.com/google/uuid"

type SubscriberDeletion struct {
	ID           uuid.UUID
	Email        string
	LeaseToken   uuid.UUID
	AttemptCount int
}

type AccountSubscriber struct {
	UserID         uuid.UUID
	Email          string
	FullName       string
	CleanupPending bool
}

type SubscriberUpdate struct {
	Email      string
	UserID     uuid.UUID
	ListIDs    []int64
	Attributes map[string]string
}
