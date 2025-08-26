package smtp

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/smtp"

	"github.com/SomeHowMicroservice/shm-be/auth/config"
)

type smtpServiceImpl struct {
	cfg  *config.Config
	auth smtp.Auth
}

//go:embed template/auth.html
var emailTemplates embed.FS

func NewSMTPService(cfg *config.Config) SMTPService {
	auth := smtp.PlainAuth("", cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.Host)
	return &smtpServiceImpl{
		cfg:  cfg,
		auth: auth,
	}
}

func (s *smtpServiceImpl) Send(to, subject, body string) error {
	msg := []byte(fmt.Sprintf("Subject: %s\r\nMIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n%s", subject, body))
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTP.Host, s.cfg.SMTP.Port)
	return smtp.SendMail(addr, s.auth, s.cfg.SMTP.Username, []string{to}, msg)
}

func (s *smtpServiceImpl) SendAuthEmail(to, subject, otp string) error {
	tmpl, err := template.ParseFS(emailTemplates, "template/auth.html")
	if err != nil {
		return fmt.Errorf("không thể load template: %w", err)
	}

	var body bytes.Buffer
	data := struct {
		Subject string 
		Otp string
	} {
		Subject: subject,
		Otp: otp,
	}

	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("không thể render template: %w", err)
	}

	return s.Send(to, subject, body.String())
}