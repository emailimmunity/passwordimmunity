package payment

import (
	"context"
	"testing"
	"time"
)

func TestEmailService(t *testing.T) {
	config := EmailConfig{
		Host:     "smtp.test.com",
		Port:     587,
		Username: "test",
		Password: "test",
		From:     "noreply@test.com",
	}

	service := NewEmailService(config)

	tests := []struct {
		name    string
		msg     EmailMessage
		ctx     context.Context
		wantErr bool
	}{
		{
			name: "valid email",
			msg: EmailMessage{
				To:      "test@example.com",
				Subject: "Test Subject",
				Body:    "Test Body",
			},
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name: "cancelled context",
			msg: EmailMessage{
				To:      "test@example.com",
				Subject: "Test Subject",
				Body:    "Test Body",
			},
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: true,
		},
		{
			name: "timeout context",
			msg: EmailMessage{
				To:      "test@example.com",
				Subject: "Test Subject",
				Body:    "Test Body",
			},
			ctx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
				defer cancel()
				time.Sleep(time.Millisecond * 2)
				return ctx
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SendEmail(tt.ctx, tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmailConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  EmailConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: EmailConfig{
				Host:     "smtp.test.com",
				Port:     587,
				Username: "test",
				Password: "test",
				From:     "noreply@test.com",
			},
			wantErr: false,
		},
		{
			name: "missing host",
			config: EmailConfig{
				Port:     587,
				Username: "test",
				Password: "test",
				From:     "noreply@test.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewEmailService(tt.config)
			if tt.wantErr && service != nil {
				t.Error("Expected nil service for invalid config")
			}
		})
	}
}
