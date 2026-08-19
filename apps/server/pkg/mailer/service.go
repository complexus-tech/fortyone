package mailer

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"gopkg.in/gomail.v2"
)

type Service interface {
	Send(ctx context.Context, email Email) error
	SendTemplated(ctx context.Context, email TemplatedEmail) error
}

type Config struct {
	Host            string
	Port            int
	Username        string
	Password        string
	FromAddress     string
	FromName        string
	MayaFromAddress string
	MayaFromName    string
	Environment     string
	BaseDir         string
}

type SenderProfile string

const SenderProfileMaya SenderProfile = "maya"

type Email struct {
	To            []string
	Subject       string
	Body          string
	PlainTextBody string
	IsHTML        bool
	Attachments   []Attachment
	Sender        SenderProfile
	ReplyTo       string
	MessageID     string
	InReplyTo     string
	References    []string
}

type Attachment struct {
	Filename string
	Data     []byte
	MimeType string
}

type TemplatedEmail struct {
	To            []string
	Template      string
	Data          any
	Subject       string
	PlainTextBody string
	Sender        SenderProfile
	ReplyTo       string
	MessageID     string
	InReplyTo     string
	References    []string
}

type service struct {
	dialer       *gomail.Dialer
	config       Config
	log          *logger.Logger
	templates    map[string]*template.Template
	baseTemplate *template.Template
}

const (
	defaultCompanyName     = "FortyOne"
	defaultLogoURL         = "https://fortyone.app/images/logo.png"
	defaultMayaFromAddress = "maya@fortyone.app"
	defaultSubject         = "FortyOne"
)

func NewService(cfg Config, log *logger.Logger) (Service, error) {
	dialer := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: cfg.Environment != "production"}

	if cfg.BaseDir == "" {
		return nil, fmt.Errorf("email template base directory is required")
	}

	baseTemplatePath := filepath.Join(cfg.BaseDir, "templates/layouts/base.html")
	baseTemplate, err := template.New("").Funcs(template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("January 2, 2006")
		},
		"safeHTML": func(value string) template.HTML {
			return template.HTML(value)
		},
		"emailStyle": emailStyle,
	}).ParseFiles(baseTemplatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base template: %w", err)
	}

	templates := make(map[string]*template.Template)
	contentTemplatePattern := filepath.Join(cfg.BaseDir, "templates/*/*.html")
	contentTemplatePaths, err := filepath.Glob(contentTemplatePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to find content templates: %w", err)
	}

	for _, templatePath := range contentTemplatePaths {
		if templatePath == baseTemplatePath || filepath.Dir(templatePath) == filepath.Join(cfg.BaseDir, "templates/layouts") {
			continue
		}

		relPath, err := filepath.Rel(filepath.Join(cfg.BaseDir, "templates"), templatePath)
		if err != nil {
			log.Error(context.Background(), "failed to get relative path", "path", templatePath, "error", err)
			continue
		}

		templateName := strings.TrimSuffix(relPath, filepath.Ext(relPath))

		tmpl, err := baseTemplate.Clone()
		if err != nil {
			return nil, fmt.Errorf("failed to clone base template: %w", err)
		}

		_, err = tmpl.ParseFiles(templatePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", templatePath, err)
		}

		templates[templateName] = tmpl
	}

	return &service{
		dialer:       dialer,
		config:       cfg,
		log:          log,
		templates:    templates,
		baseTemplate: baseTemplate,
	}, nil
}

func (s *service) Send(ctx context.Context, email Email) error {
	msg, err := s.buildMessage(email)
	if err != nil {
		return err
	}

	if err := s.dialer.DialAndSend(msg); err != nil {
		s.log.Error(ctx, "failed to send email", "error", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.log.Info(ctx, "email sent successfully", "to", email.To)
	return nil
}

func (s *service) buildMessage(email Email) (*gomail.Message, error) {
	msg := gomail.NewMessage()
	fromAddress, fromName := s.senderForProfile(email.Sender)

	msg.SetAddressHeader("From", fromAddress, fromName)
	msg.SetHeader("To", email.To...)
	msg.SetHeader("Subject", email.Subject)
	if messageID := strings.TrimSpace(email.MessageID); messageID != "" {
		if err := validateEmailHeaderValue("Message-ID", messageID); err != nil {
			return nil, err
		}
		msg.SetHeader("Message-ID", messageID)
	}

	if replyTo := strings.TrimSpace(email.ReplyTo); replyTo != "" {
		if err := validateEmailHeaderValue("Reply-To", replyTo); err != nil {
			return nil, err
		}
		msg.SetHeader("Reply-To", replyTo)
	}

	if inReplyTo := strings.TrimSpace(email.InReplyTo); inReplyTo != "" {
		if err := validateEmailHeaderValue("In-Reply-To", inReplyTo); err != nil {
			return nil, err
		}
		msg.SetHeader("In-Reply-To", inReplyTo)
	}

	if len(email.References) > 0 {
		references := make([]string, 0, len(email.References))
		for _, reference := range email.References {
			reference = strings.TrimSpace(reference)
			if reference == "" {
				continue
			}
			if err := validateEmailHeaderValue("References", reference); err != nil {
				return nil, err
			}
			references = append(references, reference)
		}
		if len(references) > 0 {
			msg.SetHeader("References", strings.Join(references, " "))
		}
	}

	if email.IsHTML {
		if email.PlainTextBody != "" {
			msg.SetBody("text/plain", email.PlainTextBody)
			msg.AddAlternative("text/html", email.Body)
		} else {
			msg.SetBody("text/html", email.Body)
		}
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

	return msg, nil
}

func validateEmailHeaderValue(name, value string) error {
	if strings.ContainsAny(value, "\r\n") {
		return fmt.Errorf("%s header contains a line break", name)
	}
	return nil
}

func (s *service) senderForProfile(profile SenderProfile) (string, string) {
	fromAddress := s.config.FromAddress
	fromName := s.config.FromName
	if profile == SenderProfileMaya {
		fromAddress = defaultMayaFromAddress
		if configuredAddress := strings.TrimSpace(s.config.MayaFromAddress); configuredAddress != "" {
			fromAddress = configuredAddress
		}
		if configuredName := strings.TrimSpace(s.config.MayaFromName); configuredName != "" {
			fromName = configuredName
		} else {
			fromName = "Maya, AI Agent"
		}
	}
	if fromName == "" {
		fromName = fromAddress
	}
	return fromAddress, fromName
}

func (s *service) SendTemplated(ctx context.Context, templateEmail TemplatedEmail) error {
	data := map[string]any{
		"Year":        time.Now().Year(),
		"LogoURL":     defaultLogoURL,
		"CompanyName": defaultCompanyName,
	}

	if templateEmail.Data != nil {
		switch d := templateEmail.Data.(type) {
		case map[string]any:
			for k, v := range d {
				data[k] = v
			}
		default:
			data["Data"] = templateEmail.Data
		}
	}

	tmpl, ok := s.templates[templateEmail.Template]
	if !ok {
		s.log.Error(ctx, "template not found", "template", templateEmail.Template)
		return fmt.Errorf("template not found: %s", templateEmail.Template)
	}

	subject := templateEmail.Subject
	if subject == "" {
		if subj, ok := data["Subject"].(string); ok && subj != "" {
			subject = subj
		}
	}
	if subject == "" {
		subject = defaultSubject
	}
	data["Subject"] = subject

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		s.log.Error(ctx, "failed to render email template", "error", err, "template", templateEmail.Template)
		return fmt.Errorf("failed to render email template: %w", err)
	}

	email := Email{
		To:            templateEmail.To,
		Subject:       subject,
		Body:          buf.String(),
		PlainTextBody: templateEmail.PlainTextBody,
		IsHTML:        true,
		Sender:        templateEmail.Sender,
		ReplyTo:       templateEmail.ReplyTo,
		MessageID:     templateEmail.MessageID,
		InReplyTo:     templateEmail.InReplyTo,
		References:    templateEmail.References,
	}

	return s.Send(ctx, email)
}
