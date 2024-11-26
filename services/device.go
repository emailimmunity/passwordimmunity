package services

import (
	"context"
	"time"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

type DeviceType string

const (
	DeviceBrowser DeviceType = "browser"
	DeviceDesktop DeviceType = "desktop"
	DeviceMobile  DeviceType = "mobile"
	DeviceCLI     DeviceType = "cli"
)

type DeviceService interface {
	RegisterDevice(ctx context.Context, userID uuid.UUID, device models.Device) error
	DeregisterDevice(ctx context.Context, deviceID uuid.UUID) error
	GetDevice(ctx context.Context, deviceID uuid.UUID) (*models.Device, error)
	ListDevices(ctx context.Context, userID uuid.UUID) ([]models.Device, error)
	AuthorizeDevice(ctx context.Context, deviceID uuid.UUID) error
	BlockDevice(ctx context.Context, deviceID uuid.UUID) error
	ValidateDeviceAccess(ctx context.Context, deviceID uuid.UUID) error
}

type deviceService struct {
	repo        repository.Repository
	audit       AuditService
	policy      PolicyService
}

func NewDeviceService(
	repo repository.Repository,
	audit AuditService,
	policy PolicyService,
) DeviceService {
	return &deviceService{
		repo:   repo,
		audit:  audit,
		policy: policy,
	}
}

func (s *deviceService) RegisterDevice(ctx context.Context, userID uuid.UUID, device models.Device) error {
	device.ID = uuid.New()
	device.UserID = userID
	device.Status = "pending"
	device.CreatedAt = time.Now()
	device.UpdatedAt = time.Now()

	if err := s.repo.CreateDevice(ctx, &device); err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("device_registered", "Device registration requested")
	metadata["device_type"] = string(device.Type)
	metadata["device_name"] = device.Name
	if err := s.createAuditLog(ctx, "device.registered", userID, uuid.Nil, metadata); err != nil {
		return err
	}

	return nil
}

func (s *deviceService) DeregisterDevice(ctx context.Context, deviceID uuid.UUID) error {
	device, err := s.GetDevice(ctx, deviceID)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteDevice(ctx, deviceID); err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("device_deregistered", "Device deregistered")
	metadata["device_type"] = string(device.Type)
	metadata["device_name"] = device.Name
	if err := s.createAuditLog(ctx, "device.deregistered", device.UserID, uuid.Nil, metadata); err != nil {
		return err
	}

	return nil
}

func (s *deviceService) GetDevice(ctx context.Context, deviceID uuid.UUID) (*models.Device, error) {
	return s.repo.GetDevice(ctx, deviceID)
}

func (s *deviceService) ListDevices(ctx context.Context, userID uuid.UUID) ([]models.Device, error) {
	return s.repo.ListDevices(ctx, userID)
}

func (s *deviceService) AuthorizeDevice(ctx context.Context, deviceID uuid.UUID) error {
	device, err := s.GetDevice(ctx, deviceID)
	if err != nil {
		return err
	}

	device.Status = "authorized"
	device.AuthorizedAt = &time.Time{}
	*device.AuthorizedAt = time.Now()
	device.UpdatedAt = time.Now()

	if err := s.repo.UpdateDevice(ctx, device); err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("device_authorized", "Device authorized")
	metadata["device_type"] = string(device.Type)
	metadata["device_name"] = device.Name
	if err := s.createAuditLog(ctx, "device.authorized", device.UserID, uuid.Nil, metadata); err != nil {
		return err
	}

	return nil
}

func (s *deviceService) BlockDevice(ctx context.Context, deviceID uuid.UUID) error {
	device, err := s.GetDevice(ctx, deviceID)
	if err != nil {
		return err
	}

	device.Status = "blocked"
	device.BlockedAt = &time.Time{}
	*device.BlockedAt = time.Now()
	device.UpdatedAt = time.Now()

	if err := s.repo.UpdateDevice(ctx, device); err != nil {
		return err
	}

	// Create audit log
	metadata := createBasicMetadata("device_blocked", "Device blocked")
	metadata["device_type"] = string(device.Type)
	metadata["device_name"] = device.Name
	if err := s.createAuditLog(ctx, "device.blocked", device.UserID, uuid.Nil, metadata); err != nil {
		return err
	}

	return nil
}

func (s *deviceService) ValidateDeviceAccess(ctx context.Context, deviceID uuid.UUID) error {
	device, err := s.GetDevice(ctx, deviceID)
	if err != nil {
		return err
	}

	if device.Status != "authorized" {
		return errors.New("device not authorized")
	}

	// Check device-specific policies
	if err := s.policy.EnforcePolicy(ctx, device.UserID, PolicyTypeIPRestriction, map[string]interface{}{
		"device_id": deviceID,
		"ip":        device.LastIP,
	}); err != nil {
		return err
	}

	return nil
}
