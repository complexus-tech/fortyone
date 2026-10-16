package tasks

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const TypeAgentInvocation = "agent-invocation:process"
const TypeAgentInvocationRecovery = "agent-invocation:recover"

type AgentInvocationPayload struct {
	RunID uuid.UUID `json:"runId"`
}

func (s *Service) EnqueueAgentInvocation(runID uuid.UUID) error {
	payload, err := json.Marshal(AgentInvocationPayload{RunID: runID})
	if err != nil {
		return err
	}
	_, err = s.asynqClient.Enqueue(asynq.NewTask(TypeAgentInvocation, payload),
		asynq.Queue("default"), asynq.MaxRetry(3), asynq.Timeout(5*time.Minute), asynq.Unique(10*time.Second))
	if err == asynq.ErrDuplicateTask {
		return nil
	}
	return err
}
