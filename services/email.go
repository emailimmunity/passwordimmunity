package services

import (
	"context"
	"time"
	"html/template"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

type EmailService interface {
	SendEmail(ctx context.Context, email models.Email) error
	SendTemplatedEmail(ctx context.Context, template string, data interface{}, recipients []string) error
	GetEmailTemplate(ctx context.Context, templateName string) (*models.EmailTemplate, error)
	CreateEmailTemplate(ctx context.Context, template models.EmailTemplate) error
	UpdateEmailTemplate(ctx context.Context, template models.EmailTemplate) error
	GetEmailHistory(ctx context.Context, userID uuid.UUID) ([]models.EmailHistory, error)
}

type emailService struct {
	repo        repository.Repository
	templates   map[string]*template.Template
	smtpConfig  models.SMTPConfig
}

func NewEmailService(
	repo repository.Repository,
	smtpConfig models.SMTPConfig,
) EmailService {
	return &emailService{
		repo:       repo,
		templates:  make(map[string]*template.Template),
		smtpConfig: smtpConfig,
	}
}

func (s *emailService) SendEmail(ctx context.Context, email models.Email) error {
	email.ID = uuid.New()
	email.Status = "pending"
	email.CreatedAt = time.Now()
	email.UpdatedAt = time.Now()

	if err := s.repo.CreateEmail(ctx, &email); err != nil {
		return err
	}

	// Send email asynchronously
	go s.sendEmailAsync(context.Background(), email)

	return nil
}

func (s *emailService) SendTemplatedEmail(ctx context.Context, templateName string, data interface{}, recipients []string) error {
	template, err := s.GetEmailTemplate(ctx, templateName)
	if err != nil {
		return err
	}

	// Parse template
	tmpl, err := s.parseTemplate(template)
	if err != nil {
		return err
	}

	// Execute template
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return err
	}

	// Send to each recipient
	for _, recipient := range recipients {
		email := models.Email{
			To:       recipient,
			Subject:  template.Subject,
			Body:     body.String(),
			IsHTML:   template.IsHTML,
		}

		if err := s.SendEmail(ctx, email); err != nil {
			return err
		}
	}

	return nil
}

func (s *emailService) GetEmailTemplate(ctx context.Context, templateName string) (*models.EmailTemplate, error) {
	return s.repo.GetEmailTemplate(ctx, templateName)
}

func (s *emailService) CreateEmailTemplate(ctx context.Context, template models.EmailTemplate) error {
	template.ID = uuid.New()
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	// Validate template
	if _, err := s.parseTemplate(&template); err != nil {
		return err
	}

	return s.repo.CreateEmailTemplate(ctx, &template)
}

func (s *emailService) UpdateEmailTemplate(ctx context.Context, template models.EmailTemplate) error {
	template.UpdatedAt = time.Now()

	// Validate template
	if _, err := s.parseTemplate(&template); err != nil {
		return err
	}

	return s.repo.UpdateEmailTemplate(ctx, &template)
}

func (s *emailService) GetEmailHistory(ctx context.Context, userID uuid.UUID) ([]models.EmailHistory, error) {
	return s.repo.GetEmailHistory(ctx, userID)
}

func (s *emailService) sendEmailAsync(ctx context.Context, email models.Email) {
	// Configure SMTP client
	dialer := gomail.NewDialer(
		s.smtpConfig.Host,
		s.smtpConfig.Port,
		s.smtpConfig.Username,
		s.smtpConfig.Password,
	)

	// Create message
	m := gomail.NewMessage()
	m.SetHeader("From", s.smtpConfig.From)
	m.SetHeader("To", email.To)
	m.SetHeader("Subject", email.Subject)

	if email.IsHTML {
		m.SetBody("text/html", email.Body)
	} else {
		m.SetBody("text/plain", email.Body)
	}

	// Send email
	if err := dialer.DialAndSend(m); err != nil {
		s.updateEmailStatus(ctx, &email, "failed", err.Error())
		return
	}

	s.updateEmailStatus(ctx, &email, "sent", "")
}

func (s *emailService) updateEmailStatus(ctx context.Context, email *models.Email, status string, error string) {
	email.Status = status
	email.Error = error
	email.UpdatedAt = time.Now()
	if status == "sent" {
		email.SentAt = &time.Time{}
		*email.SentAt = time.Now()
	}

	if err := s.repo.UpdateEmail(ctx, email); err != nil {
		log.Printf("Failed to update email status: %v", err)
	}
}

func (s *emailService) parseTemplate(template *models.EmailTemplate) (*template.Template, error) {
	// Check cache first
	if tmpl, ok := s.templates[template.Name]; ok {
		return tmpl, nil
	}

	// Parse template
	tmpl, err := template.New(template.Name).Parse(template.Body)
	if err != nil {
		return nil, err
	}

	// Cache template
	s.templates[template.Name] = tmpl

	return tmpl, nil
}
