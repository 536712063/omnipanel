// Package sdk 定义了所有 OmniPanel 插件必须实现的统一接口。
//
// 架构原则:
//   - 每个插件作为独立 OS 进程运行，崩溃不影响主进程
//   - Agent 主进程通过 gRPC + hashicorp/go-plugin 与插件通信
//   - 插件启动时向 Agent 注册自身能力和健康状态
//   - 所有插件必须实现 Panic Recovery 和 Graceful Shutdown
package sdk

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// ---------------------------------------------------------------------------
// 插件基础接口 —— 所有插件必须实现
// ---------------------------------------------------------------------------

// BasePlugin 定义插件生命周期和元数据接口。
type BasePlugin interface {
	Name() string
	Version() string
	Description() string
	HealthCheck(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// GRPCPlugin 是 hashicorp/go-plugin 要求的接口适配层。
// 每个插件包都需要实现此接口来注册自己的 gRPC 服务。
type GRPCPlugin interface {
	plugin.Plugin
	GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error
}

// ---------------------------------------------------------------------------
// 通用配置
// ---------------------------------------------------------------------------

type Config struct {
	DataDir    string `json:"data_dir"`
	PluginDir  string `json:"plugin_dir"`
	LogLevel   string `json:"log_level"`
	MaxLogSize int64  `json:"max_log_size"`
}

// ---------------------------------------------------------------------------
// 插件管理器 (Agent 主进程侧)
// ---------------------------------------------------------------------------

type PluginManager struct {
	plugins map[string]*plugin.Client
	logger  hclog.Logger
	config  *Config
}

func NewPluginManager(logger hclog.Logger, cfg *Config) *PluginManager {
	return &PluginManager{
		plugins: make(map[string]*plugin.Client),
		logger:  logger,
		config:  cfg,
	}
}

// ---------------------------------------------------------------------------
// 优雅关闭工具
// ---------------------------------------------------------------------------

// GracefulShutdown 为插件进程提供统一的信号处理机制。
// 插件应在 main() 中调用此函数:
//
//	func main() {
//	    sdk.ServePlugin(&MyPlugin{}, &MyGRPCPlugin{})
//	}
//
//	func ServePlugin(base BasePlugin, grpcPlugin GRPCPlugin) { ... }
func ServePlugin(base BasePlugin, grpcImpl GRPCPlugin) {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP,
	)
	defer cancel()

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   base.Name(),
		Level:  hclog.LevelFromString("INFO"),
		Output: hclog.DefaultOutput,
	})

	logger.Info("plugin starting", "name", base.Name(), "version", base.Version())

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			base.Name(): grpcImpl,
		},
		GRPCServer: plugin.DefaultGRPCServer,
		Logger:     logger,
	})

	go func() {
		<-ctx.Done()
		logger.Info("plugin received shutdown signal")
		if err := base.Shutdown(context.Background()); err != nil {
			logger.Error("plugin shutdown error", "error", err)
		}
	}()

	logger.Info("plugin stopped", "name", base.Name())
}

// HandshakeConfig 是所有插件与 Agent 之间的握手协议。
// 版本号变更意味着不兼容的接口变更。
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "OMNIPANEL_PLUGIN",
	MagicCookieValue: "a8f5f167f44f4964e6c998dee827110c",
}

// GRPCPluginAdapter 是 hashicorp/go-plugin 的通用适配器。
// 用于 Agent 主进程创建插件客户端时引用对应的 plugin.Plugin 实现。
// 每个插件定义自己的 GRPCPlugin 实现, Agent 通过此适配器获取。
type GRPCPluginAdapter struct {
	plugin.Plugin
	Impl GRPCPlugin
}

func (a *GRPCPluginAdapter) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	if a.Impl != nil {
		return a.Impl.GRPCServer(broker, s)
	}
	return nil
}

func (a *GRPCPluginAdapter) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return nil, nil
}
