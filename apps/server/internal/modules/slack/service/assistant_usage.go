package slack

import (
	"context"
)

func (p *EventProcessor) recordAssistantUsage(parent context.Context, input dailyUsageRecordInput) (dailyUsageSnapshot, error) {
	usageCtx, cancel := context.WithTimeout(context.WithoutCancel(parent), slackStateWriteTimeout)
	defer cancel()
	return p.usageBudget.Record(usageCtx, input, p.dailyWorkspaceTokenLimit)
}
