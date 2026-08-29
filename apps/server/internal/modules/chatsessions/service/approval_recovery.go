package chatsessions

import chatsessionsdomain "github.com/complexus-tech/projects-api/internal/modules/chatsessions/domain"

const (
	historicalAttachmentPlaceholder = chatsessionsdomain.HistoricalAttachmentPlaceholder
	skippedApprovalOutputMessage    = chatsessionsdomain.SkippedApprovalOutputMessage
)

type CompletedApprovalOutputLookup = chatsessionsdomain.CompletedApprovalOutputLookup
type DurableApprovalReceipt = chatsessionsdomain.DurableApprovalReceipt
type DurableApprovalReceiptLookup = chatsessionsdomain.DurableApprovalReceiptLookup

func RecoverDurableApprovalOutputs(current []any, lookup CompletedApprovalOutputLookup) ([]any, error) {
	return chatsessionsdomain.RecoverDurableApprovalOutputs(current, lookup)
}

func RecoverDurableApprovalReceipts(current []any, lookup DurableApprovalReceiptLookup) ([]any, error) {
	return chatsessionsdomain.RecoverDurableApprovalReceipts(current, lookup)
}

func RecoverCompletedApprovalOutputsForReservation(current, incoming []any, lookup CompletedApprovalOutputLookup) ([]any, error) {
	return chatsessionsdomain.RecoverCompletedApprovalOutputsForReservation(current, incoming, lookup)
}

func ReconcileCompletedApprovalReservation(current, incoming []any, lookup CompletedApprovalOutputLookup) ([]any, []any, error) {
	return chatsessionsdomain.ReconcileCompletedApprovalReservation(current, incoming, lookup)
}
