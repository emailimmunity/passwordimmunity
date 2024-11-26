package services

import (
	"context"
	"encoding/json"
	"time"
	"plugin"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/google/uuid"
)

type PluginType string

const (
	PluginTypeAuth      PluginType = "auth"
	PluginTypeStorage   PluginType = "storage"
	PluginTypeNotify    PluginType = "notify"
	PluginTypeIntegration PluginType = "integration"
)

type PluginService interface {
	RegisterPlugin(ctx context.Context, orgID uuid.UUID, pluginPath string) (*models.Plugin, error)
	UnregisterPlugin(ctx context.Context, pluginID uuid.UUID) error
	EnablePlugin(ctx context.Context, pluginID uuid.UUID) error
	DisablePlugin(ctx context.Context, pluginID uuid.UUID) error
	GetPlugin(ctx context.Context, pluginID uuid.UUID) (*models.Plugin, error)
	ListPlugins(ctx context.Context, orgID uuid.UUID) ([]models.Plugin, error)
	ExecutePlugin(ctx context.Context, pluginID uuid.UUID, action string, params map[string]interface{}) (interface{}, error)
}

type pluginService struct {
	repo        repository.Repository
	audit       AuditService
	plugins     map[uuid.UUID]*plugin.Plugin
}

func NewPluginService(repo repository.Repository, audit AuditService) PluginService {
	return &pluginService{
		repo:    repo,
		audit:   audit,
		plugins: make(map[uuid.UUID]*plugin.Plugin),
	}
}

func (s *pluginService) RegisterPlugin(ctx context.Context, orgID uuid.UUID, pluginPath string) (*models.Plugin, error) {
	// Load and validate plugin
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, err
	}

	// Get plugin metadata
	metadataSymbol, err := p.Lookup("Metadata")
	if err != nil {
		return nil, err
	}

	metadata, ok := metadataSymbol.(*models.PluginMetadata)
	if !ok {
		return nil, errors.New("invalid plugin metadata")
	}

	plugin := &models.Plugin{
		ID:          uuid.New(),
		OrgID:       orgID,
		Name:        metadata.Name,
		Type:        string(metadata.Type),
		Version:     metadata.Version,
		Enabled:     false,
		Path:        pluginPath,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreatePlugin(ctx, plugin); err != nil {
		return nil, err
	}

	s.plugins[plugin.ID] = p

	// Create audit log
	metadata := createBasicMetadata("plugin_registered", "Plugin registered")
	metadata["plugin_name"] = plugin.Name
	metadata["plugin_type"] = plugin.Type
	if err := s.createAuditLog(ctx, "plugin.registered", uuid.Nil, orgID, metadata); err != nil {
		return nil, err
	}

	return plugin, nil
}

func (s *pluginService) UnregisterPlugin(ctx context.Context, pluginID uuid.UUID) error {
	plugin, err := s.GetPlugin(ctx, pluginID)
	if err != nil {
		return err
	}

	delete(s.plugins, pluginID)

	// Create audit log
	metadata := createBasicMetadata("plugin_unregistered", "Plugin unregistered")
	metadata["plugin_name"] = plugin.Name
	if err := s.createAuditLog(ctx, "plugin.unregistered", uuid.Nil, plugin.OrgID, metadata); err != nil {
		return err
	}

	return s.repo.DeletePlugin(ctx, pluginID)
}

func (s *pluginService) EnablePlugin(ctx context.Context, pluginID uuid.UUID) error {
	plugin, err := s.GetPlugin(ctx, pluginID)
	if err != nil {
		return err
	}

	plugin.Enabled = true
	plugin.UpdatedAt = time.Now()

	// Create audit log
	metadata := createBasicMetadata("plugin_enabled", "Plugin enabled")
	metadata["plugin_name"] = plugin.Name
	if err := s.createAuditLog(ctx, "plugin.enabled", uuid.Nil, plugin.OrgID, metadata); err != nil {
		return err
	}

	return s.repo.UpdatePlugin(ctx, plugin)
}

func (s *pluginService) DisablePlugin(ctx context.Context, pluginID uuid.UUID) error {
	plugin, err := s.GetPlugin(ctx, pluginID)
	if err != nil {
		return err
	}

	plugin.Enabled = false
	plugin.UpdatedAt = time.Now()

	// Create audit log
	metadata := createBasicMetadata("plugin_disabled", "Plugin disabled")
	metadata["plugin_name"] = plugin.Name
	if err := s.createAuditLog(ctx, "plugin.disabled", uuid.Nil, plugin.OrgID, metadata); err != nil {
		return err
	}

	return s.repo.UpdatePlugin(ctx, plugin)
}

func (s *pluginService) GetPlugin(ctx context.Context, pluginID uuid.UUID) (*models.Plugin, error) {
	return s.repo.GetPlugin(ctx, pluginID)
}

func (s *pluginService) ListPlugins(ctx context.Context, orgID uuid.UUID) ([]models.Plugin, error) {
	return s.repo.ListPlugins(ctx, orgID)
}


func (s *pluginService) ExecutePlugin(ctx context.Context, pluginID uuid.UUID, action string, params map[string]interface{}) (interface{}, error) {
	plugin, err := s.GetPlugin(ctx, pluginID)
	if err != nil {
		return nil, err
	}

	if !plugin.Enabled {
		return nil, errors.New("plugin is disabled")
	}

	p, ok := s.plugins[pluginID]
	if !ok {
		return nil, errors.New("plugin not loaded")
	}

	// Look up and execute the action
	actionSymbol, err := p.Lookup(action)
	if err != nil {
		return nil, err
	}

	actionFunc, ok := actionSymbol.(func(context.Context, map[string]interface{}) (interface{}, error))
	if !ok {
		return nil, errors.New("invalid action function")
	}

	result, err := actionFunc(ctx, params)
	if err != nil {
		return nil, err
	}

	// Create audit log
	metadata := createBasicMetadata("plugin_executed", "Plugin action executed")
	metadata["plugin_name"] = plugin.Name
	metadata["action"] = action
	if err := s.createAuditLog(ctx, "plugin.executed", uuid.Nil, plugin.OrgID, metadata); err != nil {
		return nil, err
	}

	return result, nil
}
