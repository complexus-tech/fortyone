package emailreply

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/complexus-tech/projects-api/pkg/emailthread"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
)

func (processor *Processor) deliverReply(
	ctx context.Context,
	delivery OutboundDelivery,
	thread Thread,
	reply resolvedReply,
	authorizedTeamIDs []uuid.UUID,
) error {
	htmlBody, err := processor.renderer.RenderHTML(reply.Copy)
	if err != nil {
		return err
	}
	var replyContext emailthread.ThreadContext
	if len(thread.Context) > 0 {
		if err := json.Unmarshal(thread.Context, &replyContext); err != nil {
			return fmt.Errorf("decode reply footer context: %w", err)
		}
	}
	htmlBody, err = renderMayaReplyHTML(reply.Copy.Subject, htmlBody, replyContext.WorkspaceSlug)
	if err != nil {
		return err
	}
	eventID := strings.TrimPrefix(delivery.IdempotencyKey, "email-reply:")
	messageID := deterministicReplyMessageID(thread.ID, eventID, reply.Key)
	if len(delivery.ProviderPayload) == 0 {
		replyToken, tokenErr := processor.threads.NewReplyToken(ctx, thread)
		if tokenErr != nil {
			_ = processor.failOutboundDelivery(ctx, delivery.ID, tokenErr)
			return tokenErr
		}
		plainPayload, encodeErr := json.Marshal(emailDeliveryPayload{
			To: []string{thread.RecipientEmail}, Subject: reply.Copy.Subject, HTML: htmlBody, PlainText: reply.Copy.PlainText,
			ReplyToken: replyToken, MessageID: messageID, InReplyTo: thread.LatestInternetMessageID,
			References: emailReplyReferences(thread), Kind: reply.Kind, HistoryIdempotencyKey: "outbound:" + eventID,
			AuthorizationVersion: 1, AuthorizedTeamIDs: append([]uuid.UUID(nil), authorizedTeamIDs...),
		})
		if encodeErr != nil {
			_ = processor.failOutboundDelivery(ctx, delivery.ID, encodeErr)
			return fmt.Errorf("encode Maya email delivery: %w", encodeErr)
		}
		sealed, sealErr := processor.inbound.SealProcessorState(plainPayload)
		if sealErr != nil {
			_ = processor.failOutboundDelivery(ctx, delivery.ID, sealErr)
			return sealErr
		}
		frozenPayload, encodeErr := json.Marshal(emailDeliveryEnvelope{Sealed: sealed})
		if encodeErr != nil {
			_ = processor.failOutboundDelivery(ctx, delivery.ID, encodeErr)
			return fmt.Errorf("encode sealed Maya email delivery: %w", encodeErr)
		}
		if persistErr := processor.store.SetOutboundDeliveryContentAndProviderPayload(ctx, delivery.ID, reply.Copy.PlainText, frozenPayload); persistErr != nil {
			_ = processor.failOutboundDelivery(ctx, delivery.ID, persistErr)
			return persistErr
		}
		delivery.ProviderPayload = frozenPayload
	}
	return processor.sendClaimedDelivery(ctx, delivery, thread)
}

func (processor *Processor) claimReplyDelivery(
	ctx context.Context,
	externalWorkspaceID, eventID string,
	receiptID uuid.UUID,
	thread Thread,
) (OutboundDelivery, bool, error) {
	expiresAt := processor.now().UTC().Add(emailReplyDeliveryLifetime)
	userID := thread.UserID
	return processor.store.StartOutboundDelivery(ctx, OutboundDeliveryInput{
		Provider: Provider, WorkspaceID: thread.WorkspaceID, UserID: &userID,
		ExternalWorkspaceID: externalWorkspaceID, ExternalRecipientUserID: thread.UserID.String(),
		InboundEventID: &receiptID, IdempotencyKey: "email-reply:" + eventID,
		ExternalChannelID: thread.RecipientEmail, ExternalThreadID: thread.ExternalThreadID,
		Purpose: emailReplyPurpose, ExpiresAt: &expiresAt,
	})
}

func (processor *Processor) sendClaimedDelivery(
	ctx context.Context,
	delivery OutboundDelivery,
	thread Thread,
) error {
	persisted, err := processor.decodeEmailDeliveryPayload(delivery.ProviderPayload)
	if err != nil {
		_ = processor.failOutboundDelivery(ctx, delivery.ID, err)
		return err
	}
	replyTo, err := processor.threads.PrepareReply(ctx, ReplyPreparation{
		Thread: thread, ReplyToken: persisted.ReplyToken, InternetMessageID: persisted.MessageID,
		InReplyTo: persisted.InReplyTo, Subject: persisted.Subject, Content: persisted.PlainText,
		Kind: persisted.Kind, IdempotencyKey: persisted.HistoryIdempotencyKey,
		Context: json.RawMessage(`{"source":"email_reply"}`),
	})
	if err != nil {
		_ = processor.failOutboundDelivery(ctx, delivery.ID, err)
		return err
	}
	persisted.ReplyTo = replyTo
	if delivery.ExpiresAt != nil && !processor.now().UTC().Before(delivery.ExpiresAt.UTC()) {
		err := errors.New("maya email reply delivery expired before send")
		_ = processor.failOutboundDelivery(ctx, delivery.ID, err)
		return err
	}
	if err := processor.mailer.Send(ctx, persisted.Email()); err != nil {
		if failErr := processor.failOutboundDelivery(ctx, delivery.ID, err); failErr != nil {
			return errors.Join(err, failErr)
		}
		return err
	}
	if err := processor.store.CompleteOutboundDelivery(ctx, delivery.ID, persisted.MessageID); err != nil {
		return err
	}
	return nil
}

type emailDeliveryPayload struct {
	To                    []string    `json:"to"`
	Subject               string      `json:"subject"`
	HTML                  string      `json:"html"`
	PlainText             string      `json:"plainText"`
	ReplyToken            string      `json:"replyToken"`
	ReplyTo               string      `json:"-"`
	MessageID             string      `json:"messageId"`
	InReplyTo             string      `json:"inReplyTo"`
	References            []string    `json:"references"`
	Kind                  string      `json:"kind"`
	HistoryIdempotencyKey string      `json:"historyIdempotencyKey"`
	AuthorizationVersion  int         `json:"authorizationVersion"`
	AuthorizedTeamIDs     []uuid.UUID `json:"authorizedTeamIds"`
}

type emailDeliveryEnvelope struct {
	Sealed string `json:"sealed"`
}

func (payload emailDeliveryPayload) Email() mailer.Email {
	return mailer.Email{
		To: payload.To, Subject: payload.Subject, Body: payload.HTML, PlainTextBody: payload.PlainText, IsHTML: true,
		Sender: mailer.SenderProfileMaya, ReplyTo: payload.ReplyTo,
		MessageID: payload.MessageID, InReplyTo: payload.InReplyTo, References: payload.References,
	}
}

func (processor *Processor) decodeEmailDeliveryPayload(raw []byte) (emailDeliveryPayload, error) {
	var envelope emailDeliveryEnvelope
	if len(raw) == 0 || json.Unmarshal(raw, &envelope) != nil || strings.TrimSpace(envelope.Sealed) == "" {
		return emailDeliveryPayload{}, errors.New("persisted Maya email delivery envelope is invalid")
	}
	opened, err := processor.inbound.OpenProcessorState(envelope.Sealed)
	if err != nil {
		return emailDeliveryPayload{}, err
	}
	var payload emailDeliveryPayload
	if json.Unmarshal(opened, &payload) != nil || len(payload.To) != 1 ||
		strings.TrimSpace(payload.Subject) == "" || strings.TrimSpace(payload.HTML) == "" ||
		strings.TrimSpace(payload.PlainText) == "" || strings.TrimSpace(payload.MessageID) == "" ||
		!validReplyToken(payload.ReplyToken) || strings.TrimSpace(payload.Kind) == "" ||
		strings.TrimSpace(payload.HistoryIdempotencyKey) == "" || payload.AuthorizationVersion != 1 {
		return emailDeliveryPayload{}, errors.New("persisted Maya email delivery is invalid")
	}
	for _, teamID := range payload.AuthorizedTeamIDs {
		if teamID == uuid.Nil {
			return emailDeliveryPayload{}, errors.New("persisted Maya email delivery authorization is invalid")
		}
	}
	return payload, nil
}

func (processor *Processor) authorizeFrozenDelivery(raw []byte, current AuthorizedContext) error {
	payload, err := processor.decodeEmailDeliveryPayload(raw)
	if err != nil {
		return err
	}
	allowed := make(map[uuid.UUID]struct{}, len(current.AllowedTeamIDs))
	for _, teamID := range current.AllowedTeamIDs {
		allowed[teamID] = struct{}{}
	}
	for _, requiredTeamID := range payload.AuthorizedTeamIDs {
		if _, ok := allowed[requiredTeamID]; !ok {
			return ErrActionUnauthorized
		}
	}
	return nil
}
