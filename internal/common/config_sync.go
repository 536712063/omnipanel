package common

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"
)

type ConfigStore interface {
	Load(ctx context.Context) (map[string]interface{}, error)
	Save(ctx context.Context, data map[string]interface{}) error
}

type LocalConfigStore struct {
	path string
}

func NewLocalConfigStore(path string) *LocalConfigStore {
	return &LocalConfigStore{path: path}
}

func (s *LocalConfigStore) Load(ctx context.Context) (map[string]interface{}, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]interface{}{}, nil
		}
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return map[string]interface{}{}, nil
	}
	return result, nil
}

func (s *LocalConfigStore) Save(ctx context.Context, data map[string]interface{}) error {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, bytes, 0644)
}

type CloudSyncConfig struct {
	Enabled    bool   `json:"enabled"`
	Provider   string `json:"provider"`
	Endpoint   string `json:"endpoint"`
	Token      string `json:"token"`
	RemotePath string `json:"remote_path"`
	Interval   int    `json:"interval"`
}

type ConfigSync struct {
	mu         sync.RWMutex
	local      ConfigStore
	remote     ConfigStore
	cfg        CloudSyncConfig
	lastSyncAt time.Time
}

func NewConfigSync(local ConfigStore, remote ConfigStore, cfg CloudSyncConfig) *ConfigSync {
	return &ConfigSync{
		local:  local,
		remote: remote,
		cfg:    cfg,
	}
}

func (s *ConfigSync) GetConfig() (map[string]interface{}, error) {
	return s.local.Load(context.Background())
}

func (s *ConfigSync) SetConfig(data map[string]interface{}) error {
	return s.local.Save(context.Background(), data)
}

func (s *ConfigSync) PullFromCloud() (map[string]interface{}, error) {
	if s.remote == nil {
		return nil, nil
	}
	data, err := s.remote.Load(context.Background())
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.lastSyncAt = time.Now()
	s.mu.Unlock()
	return data, nil
}

func (s *ConfigSync) PushToCloud() error {
	if s.remote == nil {
		return nil
	}
	data, err := s.local.Load(context.Background())
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.lastSyncAt = time.Now()
	s.mu.Unlock()
	return s.remote.Save(context.Background(), data)
}

func (s *ConfigSync) GetSyncStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return map[string]interface{}{
		"enabled":    s.cfg.Enabled,
		"provider":   s.cfg.Provider,
		"last_sync":  s.lastSyncAt.Format(time.RFC3339),
	}
}

func (s *ConfigSync) StartAutoSync(ctx context.Context) {
	if !s.cfg.Enabled || s.cfg.Interval <= 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(time.Duration(s.cfg.Interval) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = s.PushToCloud()
			}
		}
	}()
}

func (s *ConfigSync) GetLocalStore() ConfigStore {
	return s.local
}

func (s *ConfigSync) GetRemoteStore() ConfigStore {
	return s.remote
}
