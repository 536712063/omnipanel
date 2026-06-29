package agent

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/omnipanel/omnipanel/plugins/sdk"
)

// PluginSpec 描述一个可加载的插件
type PluginSpec struct {
	Name        string
	BinaryPath  string
	Required    bool // 必需插件，启动失败则 Agent 退出
	AutoRestart bool // 崩溃后自动重启
	MaxRetries  int  // 最大重试次数
}

// PluginClient 封装一个已启动插件的运行时状态
type PluginClient struct {
	Spec      PluginSpec
	Client    *plugin.Client
	grpcConn  interface{} // 实际类型 *grpc.ClientConn, plug 间循环引用, 用 interface{} 包装
	PID       int
	StartTime time.Time
	Restarts  int
	mu        sync.RWMutex
}

// Manager 管理所有插件进程的生命周期
type Manager struct {
	logger  hclog.Logger
	plugins map[string]*PluginClient
	config  *sdk.Config
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewManager(logger hclog.Logger, cfg *sdk.Config) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		logger:  logger,
		plugins: make(map[string]*PluginClient),
		config:  cfg,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// LoadPlugin 启动一个插件进程并建立 gRPC 连接。
// 返回 grpc.ClientConnInterface 供上层调用。
func (m *Manager) LoadPlugin(ctx context.Context, spec PluginSpec) (*plugin.Client, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if existing, ok := m.plugins[spec.Name]; ok {
		if existing.Client.Exited() {
			m.logger.Warn("plugin previously exited, re-creating", "name", spec.Name)
			delete(m.plugins, spec.Name)
		} else {
			return existing.Client, nil
		}
	}

	binaryPath := spec.BinaryPath
	if binaryPath == "" {
		binaryPath = filepath.Join(m.config.PluginDir, fmt.Sprintf("omnipanel-plugin-%s", spec.Name))
	}

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin binary not found: %s", binaryPath)
	}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  sdk.HandshakeConfig,
		Plugins:          map[string]plugin.Plugin{spec.Name: &sdk.GRPCPluginAdapter{}},
		Cmd:              exec.Command(binaryPath),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           m.logger.Named(spec.Name),
		Managed:          true,
		AutoMTLS:         true,
	})

	m.logger.Info("starting plugin", "name", spec.Name, "binary", binaryPath)

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("failed to connect to plugin %s: %w", spec.Name, err)
	}

	pc := &PluginClient{
		Spec:      spec,
		Client:    client,
		StartTime: time.Now(),
	}

	m.plugins[spec.Name] = pc

	// 启动健康监控 goroutine
	go m.monitorPlugin(ctx, spec.Name)

	_ = rpcClient // 由上层调用者使用 grpc.Dial 建立具体连接

	return client, nil
}

// monitorPlugin 监控插件进程退出，支持自动重启
func (m *Manager) monitorPlugin(ctx context.Context, name string) {
	m.mu.RLock()
	pc, ok := m.plugins[name]
	m.mu.RUnlock()
	if !ok {
		return
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if pc.Client.Exited() {
				m.logger.Warn("plugin exited unexpectedly", "name", name)
				if pc.Spec.AutoRestart && pc.Restarts < pc.Spec.MaxRetries {
					pc.Restarts++
					m.logger.Info("auto-restarting plugin",
						"name", name,
						"attempt", pc.Restarts,
						"max", pc.Spec.MaxRetries,
					)
					m.mu.Lock()
					delete(m.plugins, name)
					m.mu.Unlock()
					if _, err := m.LoadPlugin(ctx, pc.Spec); err != nil {
						m.logger.Error("failed to restart plugin", "name", name, "error", err)
					}
				} else {
					m.logger.Error("plugin reached max retries, giving up",
						"name", name,
						"restarts", pc.Restarts,
					)
					return
				}
			}
		}
	}
}

// ShutdownAll 优雅关闭所有插件进程
func (m *Manager) ShutdownAll(timeout time.Duration) error {
	m.cancel()
	m.mu.Lock()
	defer m.mu.Unlock()

	var wg sync.WaitGroup
	errCh := make(chan error, len(m.plugins))

	for name, pc := range m.plugins {
		wg.Add(1)
		go func(pluginName string, client *plugin.Client) {
			defer wg.Done()
			m.logger.Info("shutting down plugin", "name", pluginName)
			client.Kill()
		}(name, pc.Client)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		m.logger.Info("all plugins shut down successfully")
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("plugin shutdown timeout after %v", timeout)
	case err := <-errCh:
		return err
	}
}

// ListPlugins 返回所有已加载插件的状态快照
func (m *Manager) ListPlugins() map[string]*PluginClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]*PluginClient, len(m.plugins))
	for k, v := range m.plugins {
		result[k] = v
	}
	return result
}
