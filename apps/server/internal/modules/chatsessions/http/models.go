package chatsessionshttp

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	chatsessions "github.com/complexus-tech/projects-api/internal/modules/chatsessions/service"
	"github.com/google/uuid"
)

var mutationApprovalFingerprintPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// AppChatSession represents a chat session in the application layer
type AppChatSession struct {
	ID          string    `json:"id"`
	UserID      uuid.UUID `json:"userId"`
	WorkspaceID uuid.UUID `json:"workspaceId"`
	Title       string    `json:"title"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// AppNewChatSession represents a request to create a new chat session
type AppNewChatSession struct {
	ID       string `json:"id" validate:"required,len=16"`
	Title    string `json:"title" validate:"required"`
	Messages []any  `json:"messages"`
}

// AppSaveMessagesRequest represents a request to save messages for a session
type AppSaveMessagesRequest struct {
	ID       string `json:"id" validate:"required,len=16"`
	Messages []any  `json:"messages" validate:"required"`
}

type AppBeginMessageWriteRequest struct {
	Title     string                             `json:"title"`
	Messages  []any                              `json:"messages"`
	Operation chatsessions.MessageWriteOperation `json:"operation"`
	MessageID string                             `json:"messageId,omitempty"`
}

func (r AppBeginMessageWriteRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title is required")
	}
	switch r.Operation {
	case chatsessions.MessageWriteAppend,
		chatsessions.MessageWriteRegenerate,
		chatsessions.MessageWriteApproval:
		return nil
	default:
		return errors.New("operation is not supported")
	}
}

type AppFinalizeMessageWriteRequest struct {
	Messages   []any     `json:"messages"`
	Generation int64     `json:"generation"`
	Token      uuid.UUID `json:"token"`
}

func (r AppFinalizeMessageWriteRequest) Validate() error {
	if r.Generation <= 0 {
		return errors.New("generation must be positive")
	}
	if r.Token == uuid.Nil {
		return errors.New("token must be a non-zero UUID")
	}
	return nil
}

type AppRecoverMutationApprovalOutputRequest struct {
	Fingerprint string `json:"fingerprint"`
}

func (r AppRecoverMutationApprovalOutputRequest) Validate() error {
	if !mutationApprovalFingerprintPattern.MatchString(r.Fingerprint) {
		return errors.New("fingerprint must be a lowercase SHA-256 digest")
	}
	return nil
}

type AppMessageWriteReservation struct {
	Generation int64     `json:"generation"`
	Token      uuid.UUID `json:"token"`
	Messages   []any     `json:"messages,omitempty"`
}

type AppMessageWriteResult struct {
	Applied bool `json:"applied"`
}

// AppUpdateSessionRequest represents a request to update a session
type AppUpdateSessionRequest struct {
	Title string `json:"title" validate:"required"`
}

type GetUserMessageCountResponse struct {
	Count int `json:"count"`
}

type AppClaimMutationApprovalRequest struct {
	Fingerprint string `json:"fingerprint"`
}

func (r AppClaimMutationApprovalRequest) Validate() error {
	if !mutationApprovalFingerprintPattern.MatchString(r.Fingerprint) {
		return errors.New("fingerprint must be a lowercase SHA-256 digest")
	}
	return nil
}

type AppCompleteMutationApprovalRequest struct {
	Fingerprint string          `json:"fingerprint"`
	LeaseToken  uuid.UUID       `json:"leaseToken"`
	Output      json.RawMessage `json:"output"`
}

func (r AppCompleteMutationApprovalRequest) Validate() error {
	if err := validateMutationApprovalLease(r.Fingerprint, r.LeaseToken); err != nil {
		return err
	}
	if !json.Valid(r.Output) {
		return errors.New("output must be valid JSON")
	}
	return nil
}

type AppStartMutationApprovalRequest struct {
	Fingerprint string    `json:"fingerprint"`
	LeaseToken  uuid.UUID `json:"leaseToken"`
}

func (r AppStartMutationApprovalRequest) Validate() error {
	return validateMutationApprovalLease(r.Fingerprint, r.LeaseToken)
}

type AppFailMutationApprovalRequest struct {
	FailureCode string    `json:"failureCode"`
	Fingerprint string    `json:"fingerprint"`
	LeaseToken  uuid.UUID `json:"leaseToken"`
}

type AppReconcileMutationApprovalRequest struct {
	Evidence    chatsessions.MutationApprovalReconciliationEvidence   `json:"evidence"`
	Fingerprint string                                                `json:"fingerprint"`
	Output      json.RawMessage                                       `json:"output,omitempty"`
	Resolution  chatsessions.MutationApprovalReconciliationResolution `json:"resolution"`
}

func (r AppReconcileMutationApprovalRequest) Validate() error {
	if !mutationApprovalFingerprintPattern.MatchString(r.Fingerprint) {
		return errors.New("fingerprint must be a lowercase SHA-256 digest")
	}
	switch r.Resolution {
	case chatsessions.MutationApprovalReconciliationVerifiedCompleted:
		if !json.Valid(r.Output) {
			return errors.New("verified completion requires valid JSON output")
		}
	case chatsessions.MutationApprovalReconciliationVerifiedNotApplied:
		if len(r.Output) > 0 {
			return errors.New("verified-not-applied reconciliation cannot include output")
		}
	default:
		return errors.New("resolution is not supported")
	}
	reference := strings.TrimSpace(r.Evidence.Reference)
	summary := strings.TrimSpace(r.Evidence.Summary)
	if len(reference) == 0 || len(reference) > 255 {
		return errors.New("evidence reference must contain between 1 and 255 characters")
	}
	if len(summary) == 0 || len(summary) > 500 {
		return errors.New("evidence summary must contain between 1 and 500 characters")
	}
	switch r.Evidence.Kind {
	case chatsessions.MutationApprovalReconciliationEvidenceIdempotencyLookup,
		chatsessions.MutationApprovalReconciliationEvidenceWorkspaceProbe:
	default:
		return errors.New("evidence kind is not supported")
	}
	return nil
}

func (r AppFailMutationApprovalRequest) Validate() error {
	if err := validateMutationApprovalLease(r.Fingerprint, r.LeaseToken); err != nil {
		return err
	}
	switch r.FailureCode {
	case chatsessions.MutationApprovalFailureStartUncertain,
		chatsessions.MutationApprovalFailureCompletionUncertain:
		return nil
	default:
		return errors.New("failureCode is not supported")
	}
}

type AppMutationApprovalExecution struct {
	State          chatsessions.MutationApprovalExecutionState `json:"state"`
	Output         json.RawMessage                             `json:"output,omitempty"`
	LeaseToken     *uuid.UUID                                  `json:"leaseToken,omitempty"`
	LeaseExpiresAt *time.Time                                  `json:"leaseExpiresAt,omitempty"`
	FailureCode    string                                      `json:"failureCode,omitempty"`
}

func validateMutationApprovalLease(fingerprint string, leaseToken uuid.UUID) error {
	if !mutationApprovalFingerprintPattern.MatchString(fingerprint) {
		return errors.New("fingerprint must be a lowercase SHA-256 digest")
	}
	if leaseToken == uuid.Nil {
		return errors.New("leaseToken must be a non-zero UUID")
	}
	return nil
}

// Conversion functions
func toAppChatSession(s chatsessions.CoreChatSession) AppChatSession {
	return AppChatSession{
		ID:          s.ID,
		UserID:      s.UserID,
		WorkspaceID: s.WorkspaceID,
		Title:       s.Title,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

func toAppChatSessions(sessions []chatsessions.CoreChatSession) []AppChatSession {
	result := make([]AppChatSession, len(sessions))
	for i, session := range sessions {
		result[i] = toAppChatSession(session)
	}
	return result
}

func toAppMutationApprovalExecution(execution chatsessions.CoreMutationApprovalExecution) AppMutationApprovalExecution {
	return AppMutationApprovalExecution{
		State:          execution.State,
		Output:         execution.Output,
		LeaseToken:     execution.LeaseToken,
		LeaseExpiresAt: execution.LeaseExpiresAt,
		FailureCode:    execution.FailureCode,
	}
}

func toAppMessageWriteReservation(reservation chatsessions.CoreMessageWriteReservation) AppMessageWriteReservation {
	return AppMessageWriteReservation{
		Generation: reservation.Generation,
		Token:      reservation.Token,
		Messages:   reservation.Messages,
	}
}

func toAppMessageWriteResult(result chatsessions.CoreMessageWriteResult) AppMessageWriteResult {
	return AppMessageWriteResult{Applied: result.Applied}
}
