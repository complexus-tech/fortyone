package brevo

import (
	"context"
	"fmt"

	brevo "github.com/getbrevo/brevo-go/lib"
)

// EmailRecipient represents an email recipient
type EmailRecipient struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// EmailSender represents the email sender
type EmailSender struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// EmailAttachment represents an email attachment
type EmailAttachment struct {
	Content string `json:"content"` // Base64 encoded content
	Name    string `json:"name"`
}

// SendTransactionalEmailRequest represents the request to send a transactional email
type SendTransactionalEmailRequest struct {
	Sender      EmailSender       `json:"sender"`
	To          []EmailRecipient  `json:"to"`
	Cc          []EmailRecipient  `json:"cc,omitempty"`
	Bcc         []EmailRecipient  `json:"bcc,omitempty"`
	Subject     string            `json:"subject"`
	HtmlContent string            `json:"htmlContent,omitempty"`
	TextContent string            `json:"textContent,omitempty"`
	ReplyTo     *EmailRecipient   `json:"replyTo,omitempty"`
	Attachments []EmailAttachment `json:"attachments,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
}

// SendTransactionalEmailResponse represents the response from sending a transactional email
type SendTransactionalEmailResponse struct {
	MessageID string `json:"messageId"`
}

// SendTransactionalEmail sends a transactional email using Brevo.
func (s *Service) SendTransactionalEmail(ctx context.Context, req SendTransactionalEmailRequest) (*SendTransactionalEmailResponse, error) {
	s.log.Info(ctx, "Sending transactional email via Brevo",
		"subject", req.Subject,
		"recipient_count", len(req.To))

	// Convert our request to Brevo format
	brevoEmail := brevo.SendSmtpEmail{
		Sender: &brevo.SendSmtpEmailSender{
			Email: req.Sender.Email,
			Name:  req.Sender.Name,
		},
		Subject:     req.Subject,
		HtmlContent: req.HtmlContent,
		TextContent: req.TextContent,
		Tags:        req.Tags,
	}

	// Convert recipients
	var brevoTo []brevo.SendSmtpEmailTo
	for _, recipient := range req.To {
		brevoTo = append(brevoTo, brevo.SendSmtpEmailTo{
			Email: recipient.Email,
			Name:  recipient.Name,
		})
	}
	brevoEmail.To = brevoTo

	// Convert CC recipients if any
	if len(req.Cc) > 0 {
		var brevoCc []brevo.SendSmtpEmailCc
		for _, recipient := range req.Cc {
			brevoCc = append(brevoCc, brevo.SendSmtpEmailCc{
				Email: recipient.Email,
				Name:  recipient.Name,
			})
		}
		brevoEmail.Cc = brevoCc
	}

	// Convert BCC recipients if any
	if len(req.Bcc) > 0 {
		var brevoBcc []brevo.SendSmtpEmailBcc
		for _, recipient := range req.Bcc {
			brevoBcc = append(brevoBcc, brevo.SendSmtpEmailBcc{
				Email: recipient.Email,
				Name:  recipient.Name,
			})
		}
		brevoEmail.Bcc = brevoBcc
	}

	// Convert ReplyTo if provided
	if req.ReplyTo != nil {
		brevoEmail.ReplyTo = &brevo.SendSmtpEmailReplyTo{
			Email: req.ReplyTo.Email,
			Name:  req.ReplyTo.Name,
		}
	}

	// Convert attachments if any
	if len(req.Attachments) > 0 {
		var brevoAttachments []brevo.SendSmtpEmailAttachment
		for _, attachment := range req.Attachments {
			brevoAttachments = append(brevoAttachments, brevo.SendSmtpEmailAttachment{
				Content: attachment.Content,
				Name:    attachment.Name,
			})
		}
		brevoEmail.Attachment = brevoAttachments
	}

	// Convert headers if any
	if len(req.Headers) > 0 {
		headers := make(map[string]any)
		for k, v := range req.Headers {
			headers[k] = v
		}
		brevoEmail.Headers = headers
	}

	// Send the email
	result, response, err := s.client.TransactionalEmailsApi.SendTransacEmail(ctx, brevoEmail)
	if err != nil {
		s.log.Error(ctx, "Failed to send transactional email via Brevo",
			"error", err,
			"subject", req.Subject,
			"response_status", response.Status)
		return nil, fmt.Errorf("brevo: failed to send transactional email: %w", err)
	}

	s.log.Info(ctx, "Successfully sent transactional email via Brevo",
		"subject", req.Subject,
		"message_id", result.MessageId)

	return &SendTransactionalEmailResponse{
		MessageID: result.MessageId,
	}, nil
}
