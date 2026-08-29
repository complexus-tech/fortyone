package chatsessions

import chatsessionsdomain "github.com/complexus-tech/projects-api/internal/modules/chatsessions/domain"

const (
	MutationApprovalUncertainOutputMessage = chatsessionsdomain.MutationApprovalUncertainOutputMessage
	MutationApprovalSkippedOutputMessage   = chatsessionsdomain.MutationApprovalSkippedOutputMessage
)

type MutationApprovalRetryIntent = chatsessionsdomain.MutationApprovalRetryIntent
type MutationApprovalRetryPreparer = chatsessionsdomain.MutationApprovalRetryPreparer

func PrepareMutationApprovalRetries(
	current []any,
	incoming []any,
	prepare MutationApprovalRetryPreparer,
) ([]any, error) {
	return chatsessionsdomain.PrepareMutationApprovalRetries(current, incoming, prepare)
}

func approvedMutationFingerprint(toolName string, input any) (string, error) {
	return chatsessionsdomain.ApprovedMutationFingerprint(toolName, input)
}
