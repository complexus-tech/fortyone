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
	Host        string
	Port        int
	Username    string
	Password    string
	FromAddress string
	FromName    string
	Environment string
	BaseDir     string
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
	Data     any
	Subject  string
	ReplyTo  string
}

type service struct {
	dialer       *gomail.Dialer
	config       Config
	log          *logger.Logger
	templates    map[string]*template.Template
	baseTemplate *template.Template
}

const (
	defaultCompanyName = "FortyOne"
	defaultLogoURL     = "https://fortyone.app/images/logo.png"
	defaultSubject     = "FortyOne"
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
	msg := gomail.NewMessage()
	fromName := s.config.FromName
	if fromName == "" {
		fromName = s.config.FromAddress
	}

	msg.SetHeader("From", fmt.Sprintf("%s <%s>", fromName, s.config.FromAddress))
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
		To:      templateEmail.To,
		Subject: subject,
		Body:    buf.String(),
		IsHTML:  true,
		ReplyTo: templateEmail.ReplyTo,
	}

	return s.Send(ctx, email)
}
