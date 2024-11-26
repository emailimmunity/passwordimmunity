package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

type SessionService interface {
	CreateSession(ctx context.Context, userID uuid.UUID, deviceInfo string) (*models.Session, error)
	ValidateSession(ctx context.Context, sessionID string) (*models.Session, error)
	RevokeSession(ctx context.Context, sessionID string) error
	RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error
}

type sessionService struct {
	repo repository.Repository
}

func NewSessionService(repo repository.Repository) SessionService {
	return &sessionService{repo: repo}
}

func generateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (s *sessionService) CreateSession(ctx context.Context, userID uuid.UUID, deviceInfo string) (*models.Session, error) {
	token, err := generateSessionToken()
	if err != nil {
		return nil, err
	}

	session := &models.Session{
		UserID:     userID,
		Token:      token,
		DeviceInfo: deviceInfo,
		ExpiresAt:  time.Now().Add(24 * time.Hour), // 24-hour session
		LastUsed:   time.Now(),
	}

	// Create audit log
	metadata := createBasicMetadata("session_created", "New session created")
	metadata["device_info"] = deviceInfo
	if err := s.createAuditLog(ctx, AuditEventUserLogin, userID, uuid.Nil, metadata); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *sessionService) ValidateSession(ctx context.Context, sessionID string) (*models.Session, error) {
	session, err := s.repo.GetSessionByToken(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session == nil {
		return nil, ErrSessionNotFound
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	// Update last used timestamp
	session.LastUsed = time.Now()
	if err := s.repo.UpdateSession(ctx, session); err != nil {
		return nil, err
	}

	// Create audit log for session validation
	metadata := createBasicMetadata("session_validated", "Session validated successfully")
	metadata["device_info"] = session.DeviceInfo
	if err := s.createAuditLog(ctx, AuditEventSessionValidated, session.UserID, uuid.Nil, metadata); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *sessionService) RevokeSession(ctx context.Context, sessionID string) error {
	session, err := s.repo.GetSessionByToken(ctx, sessionID)
	if err != nil {
		return err
	}

	if session == nil {
		return ErrSessionNotFound
	}

	// Mark session as revoked
	session.RevokedAt = time.Now()
	if err := s.repo.UpdateSession(ctx, session); err != nil {
		return err
	}

	// Create audit log for session revocation
	metadata := createBasicMetadata("session_revoked", "Session revoked")
	metadata["device_info"] = session.DeviceInfo
	if err := s.createAuditLog(ctx, AuditEventSessionRevoked, session.UserID, uuid.Nil, metadata); err != nil {
		return err
	}

	return nil
}

func (s *sessionService) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	sessions, err := s.repo.GetUserSessions(ctx, userID)
	if err != nil {
		return err
	}

	revokedAt := time.Now()
	for _, session := range sessions {
		session.RevokedAt = revokedAt
		if err := s.repo.UpdateSession(ctx, session); err != nil {
			return err
		}
	}

	// Create audit log for bulk session revocation
	metadata := createBasicMetadata("all_sessions_revoked", "All user sessions revoked")
	metadata["session_count"] = len(sessions)
	if err := s.createAuditLog(ctx, AuditEventAllSessionsRevoked, userID, uuid.Nil, metadata); err != nil {
		return err
	}

	return nil
}

func (s *sessionService) cleanupExpiredSessions(ctx context.Context) error {
	deletedCount, err := s.repo.DeleteExpiredSessions(ctx, time.Now())
	if err != nil {
		return err
	}

	if deletedCount > 0 {
		// Create audit log for session cleanup
		metadata := createBasicMetadata("sessions_cleaned", "Expired sessions cleaned up")
		metadata["deleted_count"] = deletedCount
		if err := s.createAuditLog(ctx, AuditEventSessionsCleanup, uuid.Nil, uuid.Nil, metadata); err != nil {
			return err
		}
	}

	return nil
}

func (s *sessionService) RevokeSession(ctx context.Context, sessionID string) error {
	session, err := s.repo.GetSessionByToken(ctx, sessionID)
	if err != nil {
		return err
	}

	if session == nil {
		return ErrSessionNotFound
	}

	// Mark session as revoked
	session.RevokedAt = time.Now()
	if err := s.repo.UpdateSession(ctx, session); err != nil {
		return err
	}

	// Create audit log for session revocation
	metadata := createBasicMetadata("session_revoked", "Session revoked")
	metadata["device_info"] = session.DeviceInfo
	if err := s.createAuditLog(ctx, AuditEventSessionRevoked, session.UserID, uuid.Nil, metadata); err != nil {
		return err
	}

	return nil
}

func (s *sessionService) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	sessions, err := s.repo.GetUserSessions(ctx, userID)
	if err != nil {
		return err
	}

	revokedAt := time.Now()
	for _, session := range sessions {
		session.RevokedAt = revokedAt
		if err := s.repo.UpdateSession(ctx, session); err != nil {
			return err
		}
	}

	// Create audit log for bulk session revocation
	metadata := createBasicMetadata("all_sessions_revoked", "All user sessions revoked")
	metadata["session_count"] = len(sessions)
	if err := s.createAuditLog(ctx, AuditEventAllSessionsRevoked, userID, uuid.Nil, metadata); err != nil {
		return err
	}

	return nil
}

// Helper function to clean up expired sessions
func (s *sessionService) cleanupExpiredSessions(ctx context.Context) error {
	// TODO: Implement session cleanup
	// 1. Delete expired sessions
	// 2. Create audit logs
	return nil
}

func (s *sessionService) RevokeSession(ctx context.Context, sessionID string) error {
	session, err := s.repo.GetSessionByToken(ctx, sessionID)
	if err != nil {
		return err
	}

	if session == nil {
		return ErrSessionNotFound
	}

	// Mark session as revoked
	session.RevokedAt = time.Now()
	if err := s.repo.UpdateSession(ctx, session); err != nil {
		return err
	}

	// Create audit log for session revocation
	metadata := createBasicMetadata("session_revoked", "Session revoked")
	metadata["device_info"] = session.DeviceInfo
	if err := s.createAuditLog(ctx, AuditEventSessionRevoked, session.UserID, uuid.Nil, metadata); err != nil {
		return err
	}

	return nil
}

func (s *sessionService) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	// TODO: Implement bulk session revocation
	// 1. Mark all user sessions as revoked
	// 2. Create audit log
	return nil
}

// Helper function to clean up expired sessions
func (s *sessionService) cleanupExpiredSessions(ctx context.Context) error {
	// TODO: Implement session cleanup
	// 1. Delete expired sessions
	// 2. Create audit logs
	return nil
}

func (s *sessionService) RevokeSession(ctx context.Context, sessionID string) error {
	// TODO: Implement session revocation
	// 1. Mark session as revoked
	// 2. Create audit log
	return nil
}

func (s *sessionService) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	// TODO: Implement bulk session revocation
	// 1. Mark all user sessions as revoked
	// 2. Create audit log
	return nil
}

// Helper function to clean up expired sessions
func (s *sessionService) cleanupExpiredSessions(ctx context.Context) error {
	// TODO: Implement session cleanup
	// 1. Delete expired sessions
	// 2. Create audit logs
	return nil
}
