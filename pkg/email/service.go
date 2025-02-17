package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"gopkg.in/gomail.v2"
)

type Service interface {
	SendEmail(ctx context.Context, email Email) error
	SendTemplatedEmail(ctx context.Context, templateEmail TemplatedEmail) error
}

type Config struct {
	Host        string
	Port        int
	Username    string
	Password    string
	FromAddress string
	FromName    string
	Environment string
}

type Email struct {
	To          []string
	Subject     string
	Body        string
	IsHTML      bool
	Attachments []Attachment
	ReplyTo     string
}

type Attachment struct {
	Filename string
	Data     []byte
	MimeType string
}

type TemplatedEmail struct {
	To       []string
	Template string
	Data     interface{}
	ReplyTo  string
}

type service struct {
	dialer *gomail.Dialer
	config Config
	log    *logger.Logger
}

func NewService(cfg Config, log *logger.Logger) (Service, error) {
	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: cfg.Environment != "production"}

	s := &service{
		dialer: dialer,
		config: cfg,
		log:    log,
	}

	return s, nil
}

func (s *service) SendEmail(ctx context.Context, email Email) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromAddress))
	msg.SetHeader("To", email.To...)
	msg.SetHeader("Subject", email.Subject)

	if email.ReplyTo != "" {
		msg.SetHeader("Reply-To", email.ReplyTo)
	}

	if email.IsHTML {
		msg.SetBody("text/html", email.Body)
	} else {
		msg.SetBody("text/plain", email.Body)
	}

	for _, attachment := range email.Attachments {
		msg.Attach(attachment.Filename,
			gomail.SetCopyFunc(func(w io.Writer) error {
				_, err := w.Write(attachment.Data)
				return err
			}),
			gomail.SetHeader(map[string][]string{
				"Content-Type": {attachment.MimeType},
			}),
		)
	}

	if err := s.dialer.DialAndSend(msg); err != nil {
		s.log.Error(ctx, "failed to send email", "error", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.log.Info(ctx, "email sent successfully", "to", email.To)
	return nil
}

func (s *service) SendTemplatedEmail(ctx context.Context, templateEmail TemplatedEmail) error {
	// TODO: Implement template rendering
	return fmt.Errorf("template email not implemented yet")
}
