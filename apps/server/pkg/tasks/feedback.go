package tasks

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	TypeFeedbackContributorDelivery         = "feedback:contributor:delivery"
	TypeFeedbackContributorDeliveryRecovery = "feedback:contributor:delivery:recovery"
	TypeFeedbackOutboxDispatch              = "feedback:outbox:dispatch"
)

type FeedbackContributorDeliveryPayload struct {
	DeliveryID       uuid.UUID `json:"deliveryId"`
	UnsubscribeToken string    `json:"unsubscribeToken"`
}

func (s *Service) EnqueueFeedbackContributorDelivery(payload FeedbackContributorDeliveryPayload) error {
	if payload.DeliveryID == uuid.Nil || payload.UnsubscribeToken == "" {
		return fmt.Errorf("tasks: feedback contributor delivery payload is incomplete")
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("tasks: marshal %s payload: %w", TypeFeedbackContributorDelivery, err)
	}
	_, err = s.asynqClient.Enqueue(
		asynq.NewTask(TypeFeedbackContributorDelivery, encoded),
		asynq.Queue("notifications"),
		asynq.MaxRetry(3),
		asynq.TaskID("feedback-contributor-delivery:"+payload.DeliveryID.String()),
	)
	if err == asynq.ErrTaskIDConflict {
		return nil
	}
	if err != nil {
		return fmt.Errorf("tasks: enqueue %s: %w", TypeFeedbackContributorDelivery, err)
	}
	return nil
}
