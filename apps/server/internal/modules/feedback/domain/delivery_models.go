package feedbackdomain

import (
	"context"

	"github.com/google/uuid"
)

// ContributorDeliveryStore is the feedback-owned persistence boundary used by
// the asynchronous mail worker. Claiming is deliberately separate from
// sending: a replay may safely observe an already-completed or suppressed
// delivery without sending the message again.
type ContributorDeliveryStore interface {
	ClaimContributorDelivery(context.Context, uuid.UUID) (CoreClaimedContributorDelivery, bool, error)
	MarkContributorDeliverySent(context.Context, uuid.UUID) error
	MarkContributorDeliveryFailed(context.Context, CoreContributorDeliveryFailure) error
	ListRecoverableContributorDeliveries(context.Context, int) ([]CoreRecoverableContributorDelivery, error)
}

// CoreClaimedContributorDelivery is the delivery worker's safe, bounded
// projection. It intentionally excludes contributor identity metadata that is
// not required to send the message and keeps unsubscribe verification as a
// digest rather than exposing the underlying token.
type CoreClaimedContributorDelivery struct {
	ID             uuid.UUID
	RecipientEmail string
	DisplayName    string
	PortalName     string
	PortalSlug     string
	Subject        string
	Message        string
	DestinationURL string
	TokenHash      []byte
}

// CoreRecoverableContributorDelivery contains only the durable identity needed
// to validate and re-enqueue an interrupted delivery.
type CoreRecoverableContributorDelivery struct {
	DeliveryID uuid.UUID
	TokenHash  []byte
}

// CoreContributorDeliveryFailure preserves finite retry semantics at the
// persistence boundary instead of passing status strings from the worker.
type CoreContributorDeliveryFailure struct {
	DeliveryID uuid.UUID
	Reason     string
	Terminal   bool
}
