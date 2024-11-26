package payment

import (
    "context"
    "fmt"
    "net/smtp"
    "os"

    "github.com/rs/zerolog/log"
)

type defaultEmailService struct{}

func (s *defaultEmailService) SendEmail(ctx context.Context, to string, subject string, body string) error {
    logger := log.With().
        Str("to", to).
        Str("subject", subject).
        Logger()

    smtpHost := os.Getenv("SMTP_HOST")
    smtpPort := os.Getenv("SMTP_PORT")
    smtpUser := os.Getenv("SMTP_USER")
    smtpPass := os.Getenv("SMTP_PASS")
    from := os.Getenv("SMTP_FROM")

    if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" || from == "" {
        logger.Error().Msg("missing SMTP configuration")
        return fmt.Errorf("missing SMTP configuration")
    }

    msg := fmt.Sprintf("From: %s\r\n"+
        "To: %s\r\n"+
        "Subject: %s\r\n"+
        "\r\n"+
        "%s\r\n", from, to, subject, body)

    auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
    addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

    if err := smtp.SendMail(addr, auth, from, []string{to}, []byte(msg)); err != nil {
        logger.Error().Err(err).Msg("failed to send email")
        return err
    }

    logger.Info().Msg("email sent successfully")
    return nil
}
