package subscriptionsdomain

import "github.com/google/uuid"

type WebhookClaimDisposition string

const (
	WebhookClaimAcquired         WebhookClaimDisposition = "acquired"
	WebhookClaimAlreadyProcessed WebhookClaimDisposition = "already_processed"
	WebhookClaimInProgress       WebhookClaimDisposition = "in_progress"
)

type WebhookProcessingResult string

const (
	WebhookResultHandled WebhookProcessingResult = "handled"
	WebhookResultIgnored WebhookProcessingResult = "ignored"
)

type WebhookFailureCode string

const WebhookFailureHandler WebhookFailureCode = "handler_failed"

type WebhookClaim struct {
	Disposition WebhookClaimDisposition
	LeaseToken  uuid.UUID
	Attempt     int
}

type WebhookOutcome struct {
	Result      WebhookProcessingResult
	WorkspaceID *uuid.UUID
}
