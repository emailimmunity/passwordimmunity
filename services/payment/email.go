package payment

import (
	"context"
	"fmt"
	"net/smtp"
)

type EmailMessage struct {
	To      string
	Subject string
	Body    string
}

type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type EmailService struct {
	config EmailConfig
}

func NewEmailService(config EmailConfig) *EmailService {
	return &EmailService{
		config: config,
	}
}

func (s *EmailService) SendEmail(ctx context.Context, msg EmailMessage) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	headers := make(map[string]string)
	headers["From"] = s.config.From
	headers["To"] = msg.To
	headers["Subject"] = msg.Subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=\"utf-8\""

	message := ""
	for key, value := range headers {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	message += "\r\n" + msg.Body

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	return smtp.SendMail(addr, auth, s.config.From, []string{msg.To}, []byte(message))
}
