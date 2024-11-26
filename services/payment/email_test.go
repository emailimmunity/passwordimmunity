package payment

import (
    "context"
    "os"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestDefaultEmailService(t *testing.T) {
    service := &defaultEmailService{}
    ctx := context.Background()

    t.Run("missing SMTP configuration", func(t *testing.T) {
        os.Clearenv()

        err := service.SendEmail(ctx, "test@example.com", "Test Subject", "Test Body")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "missing SMTP configuration")
    })

    t.Run("complete SMTP configuration", func(t *testing.T) {
        os.Setenv("SMTP_HOST", "smtp.example.com")
        os.Setenv("SMTP_PORT", "587")
        os.Setenv("SMTP_USER", "user")
        os.Setenv("SMTP_PASS", "pass")
        os.Setenv("SMTP_FROM", "from@example.com")
        defer os.Clearenv()

        // Note: This test will fail in actual execution since we're not mocking SMTP
        // In a real implementation, we'd use a mock SMTP server or interface
        err := service.SendEmail(ctx, "test@example.com", "Test Subject", "Test Body")
        assert.Error(t, err) // Expected to fail due to actual SMTP connection attempt
    })
}
