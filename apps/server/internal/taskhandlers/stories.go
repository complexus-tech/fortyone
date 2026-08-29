package taskhandlers

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
)

type StoryScheduleTransitionOutboxProcessor interface {
	DispatchReadyScheduleTransitionOutbox(context.Context) (int, error)
}

func (h *handlers) HandleStoryScheduleTransitionOutboxDispatch(ctx context.Context, _ *asynq.Task) error {
	if h.storyScheduleOutbox == nil {
		return fmt.Errorf("story schedule transition outbox processor is not configured")
	}
	_, err := h.storyScheduleOutbox.DispatchReadyScheduleTransitionOutbox(ctx)
	return err
}
