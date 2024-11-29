package payment

import (
	"context"
	"fmt"
	"strings"
)

type NotificationService struct {
	emailService *EmailService
}

type NotificationTemplate struct {
	Subject string
	Body    string
}

var (
	paymentSuccessTemplate = NotificationTemplate{
		Subject: "Payment Successful - Enterprise Features Activated",
		Body: `Your payment has been processed successfully. The following items have been activated for your organization:

Features: %s
Bundles: %s
Duration: %s

You can start using these features immediately. For assistance, please refer to our documentation or contact support.`,
	}

	paymentFailureTemplate = NotificationTemplate{
		Subject: "Payment Processing Issue",
		Body: `We encountered an issue processing your payment for enterprise features:

Requested Features: %s
Requested Bundles: %s

%s

Please try again or contact support if the issue persists.`,
	}
)

func NewNotificationService(emailService *EmailService) *NotificationService {
	return &NotificationService{
		emailService: emailService,
	}
}

func (s *NotificationService) NotifyPaymentSuccess(ctx context.Context, orgEmail string, purchaseType string) error {
	metadata := ctx.Value("payment_metadata").(struct {
		Features []string
		Bundles  []string
		Duration string
	})

	subject := paymentSuccessTemplate.Subject
	body := fmt.Sprintf(paymentSuccessTemplate.Body,
		formatList(metadata.Features),
		formatList(metadata.Bundles),
		metadata.Duration,
	)

	return s.emailService.SendEmail(ctx, EmailMessage{
		To:      orgEmail,
		Subject: subject,
		Body:    body,
	})
}

func (s *NotificationService) NotifyPaymentFailure(ctx context.Context, orgEmail string, purchaseType string, reason string) error {
	metadata := ctx.Value("payment_metadata").(struct {
		Features []string
		Bundles  []string
	})

	subject := paymentFailureTemplate.Subject
	body := fmt.Sprintf(paymentFailureTemplate.Body,
		formatList(metadata.Features),
		formatList(metadata.Bundles),
		reason,
	)

	return s.emailService.SendEmail(ctx, EmailMessage{
		To:      orgEmail,
		Subject: subject,
		Body:    body,
	})
}

func formatList(items []string) string {
	if len(items) == 0 {
		return "None"
	}
	return "- " + strings.Join(items, "\n- ")
}
