package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"time"

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
	BaseDir     string // Base directory for email templates
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
	Template string      // Path to template relative to BaseDir
	Data     interface{} // Data to be passed to the template
	ReplyTo  string
}

type service struct {
	dialer    *gomail.Dialer
	config    Config
	log       *logger.Logger
	templates *template.Template
}

func NewService(cfg Config, log *logger.Logger) (Service, error) {
	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: cfg.Environment != "production"}

	// Ensure base directory is set
	if cfg.BaseDir == "" {
		return nil, fmt.Errorf("email template base directory is required")
	}

	// Create a new template with functions
	templates := template.New("").Funcs(template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("January 2, 2006")
		},
	})

	// Load all templates
	templatePattern := filepath.Join(cfg.BaseDir, "templates/**/*.html")
	templates, err := templates.ParseGlob(templatePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to parse email templates: %w", err)
	}

	s := &service{
		dialer:    dialer,
		config:    cfg,
		log:       log,
		templates: templates,
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
	// Add default data that all templates need
	data := map[string]interface{}{
		"Year":    time.Now().Year(),
		"LogoURL": "https://complexus.app/images/logo.png",
	}

	// If templateEmail.Data is provided, merge it with our default data
	if templateEmail.Data != nil {
		switch d := templateEmail.Data.(type) {
		case map[string]interface{}:
			for k, v := range d {
				data[k] = v
			}
		default:
			// For non-map data, add it under a "Data" key
			data["Data"] = templateEmail.Data
		}
	}

	// Render the template
	var buf bytes.Buffer
	templateName := templateEmail.Template
	if err := s.templates.ExecuteTemplate(&buf, templateName, data); err != nil {
		s.log.Error(ctx, "failed to render email template", "error", err, "template", templateName)
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Get subject from data if provided
	subject := "Complexus"
	if subj, ok := data["Subject"].(string); ok {
		subject = subj
	}

	// Create the email
	email := Email{
		To:      templateEmail.To,
		Subject: subject,
		Body:    buf.String(),
		IsHTML:  true,
		ReplyTo: templateEmail.ReplyTo,
	}

	// Send the email
	return s.SendEmail(ctx, email)
}
