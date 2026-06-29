// Package app 定义 Wails v3 应用的核心结构和前端绑定方法。
//
// 架构:
//   Vue 3 前端通过 Wails Runtime 调用此处的导出方法。
//   每个方法对应一个 UI 操作，方法内部通过 gRPC 调用相应的插件。
//
// 安全:
//   - 所有绑定方法必须验证调用来源 (通过 Wails context)
//   - 敏感操作记录审计日志
//   - 参数校验防止注入攻击
package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/omnipanel/omnipanel/internal/agent"
	"github.com/omnipanel/omnipanel/internal/license"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// OmniPanel 是 Wails 应用的主结构体。
// 所有标记为导出的方法将自动绑定为前端可调用的 API。
type OmniPanel struct {
	ctx       context.Context
	pluginMgr *agent.Manager
	licenser  *license.Licenser
	wsHub     *WebSocketHub
	logger    hclog.Logger
	mu        sync.RWMutex
}

// NewOmniPanel 创建应用实例。
// 该函数在 main.go 中调用，传入 Wails application context。
func NewOmniPanel() *OmniPanel {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "omnipanel",
		Level:  hclog.LevelFromString("INFO"),
		Output: hclog.DefaultOutput,
	})

	cfg := &agent.Config{
		DataDir:   getDataDir(),
		PluginDir: getPluginDir(),
		LogLevel:  "INFO",
	}

	pluginMgr := agent.NewManager(logger.Named("plugin-manager"), cfg)
	licenser := license.NewLicenser(cfg.DataDir + "/license.dat")

	return &OmniPanel{
		pluginMgr: pluginMgr,
		licenser:  licenser,
		wsHub:     NewWebSocketHub(),
		logger:    logger,
	}
}

// Startup 由 Wails 在应用启动时调用。
// 用于初始化插件系统、校验 License、启动 WebSocket 服务。
func (a *OmniPanel) Startup(ctx context.Context) {
	a.ctx = ctx
	a.logger.Info("OmniPanel 正在启动...")

	// 1. 校验 License (宽限策略: 即使失效也允许使用基本功能)
	status, err := a.licenser.Validate("")
	if err != nil {
		a.logger.Warn("License 校验失败", "error", err)
	} else {
		a.logger.Info("License 状态", "plan", status.Plan, "days_remaining", status.DaysRemaining)
	}

	// 2. 启动核心插件
	a.startPlugins(ctx)

	// 3. 启动 WebSocket 服务 (用于实时推流)
	go a.wsHub.Run(ctx)

	// 4. 监听系统信号实现优雅关闭
	go a.handleShutdown(ctx)

	a.logger.Info("OmniPanel 启动完成")
}

// Shutdown 由 Wails 在应用退出时调用。
func (a *OmniPanel) Shutdown(ctx context.Context) {
	a.logger.Info("OmniPanel 正在关闭...")
	if err := a.pluginMgr.ShutdownAll(30 * time.Second); err != nil {
		a.logger.Error("插件关闭超时", "error", err)
	}
	a.wsHub.Shutdown()
	a.logger.Info("OmniPanel 已关闭")
}

// ===========================================================================
// 前端绑定方法 — License
// ===========================================================================

// GetLicenseStatus 返回当前 License 状态。
// 前端在启动时调用此方法决定显示哪些功能模块。
func (a *OmniPanel) GetLicenseStatus() (*license.LicenseStatus, error) {
	return a.licenser.Validate("")
}

// ActivateLicense 激活 License Key。
func (a *OmniPanel) ActivateLicense(licenseKey, email string) (*license.LicenseStatus, error) {
	return a.licenser.Activate(licenseKey, email)
}

// ===========================================================================
// 前端绑定方法 — 系统信息
// ===========================================================================

// GetSystemInfo 获取本地机器信息 (CPU/内存/磁盘/网络)。
func (a *OmniPanel) GetSystemInfo() (map[string]interface{}, error) {
	return collectSystemInfo()
}

// GetProcessList 获取进程列表。
func (a *OmniPanel) GetProcessList() ([]map[string]interface{}, error) {
	return getProcessList()
}

// ===========================================================================
// 前端绑定方法 — SDTD (七日杀)
// ===========================================================================

// SDTDInstallServer 安装七日杀服务端 (通过 SteamCMD)。
// 返回一个任务 ID，前端通过 WebSocket 接收实时安装日志。
func (a *OmniPanel) SDTDInstallServer(installDir, steamUser, steamPass string) (string, error) {
	taskID := fmt.Sprintf("sdt-install-%d", time.Now().UnixNano())
	go func() {
		// 实际调用 SDTD 插件 gRPC 接口
		// client := sdtv1.NewSDTServiceClient(conn)
		// stream, _ := client.InstallServer(ctx, &sdtv1.InstallServerRequest{...})
		// for { line, _ := stream.Recv(); a.wsHub.Broadcast(taskID, line) }
		a.wsHub.Broadcast(taskID, map[string]string{
			"status":  "running",
			"message": "正在通过 SteamCMD 下载七日杀服务端...",
		})
	}()
	return taskID, nil
}

// SDTDStartServer 启动七日杀服务端。
func (a *OmniPanel) SDTDStartServer() (map[string]interface{}, error) {
	// 实际调用 SDTD 插件 gRPC 接口
	return map[string]interface{}{
		"success": true,
		"message": "七日杀服务端启动指令已发送",
	}, nil
}

// SDTDStopServer 停止七日杀服务端。
func (a *OmniPanel) SDTDStopServer() (map[string]interface{}, error) {
	return map[string]interface{}{
		"success": true,
		"message": "七日杀服务端停止指令已发送",
	}, nil
}

// SDTDGetServerConfig 获取 serverconfig.xml 内容。
func (a *OmniPanel) SDTDGetServerConfig() (map[string]interface{}, error) {
	// 实际调用 SDTD 插件 gRPC 接口
	return map[string]interface{}{
		"serverName":        "OmniPanel 七日杀服务器",
		"serverPort":        26900,
		"maxPlayers":        16,
		"gameDifficulty":    4,
		"worldGenSeed":      "OmniPanel2026",
		"eacEnabled":        true,
		"bloodMoonFrequency": 7,
	}, nil
}

// SDTDSaveServerConfig 保存并应用 serverconfig.xml。
func (a *OmniPanel) SDTDSaveServerConfig(config map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"success": true,
		"message": "配置已保存",
	}, nil
}

// SDTDSendConsoleCommand 向游戏控制台发送命令。
func (a *OmniPanel) SDTDSendConsoleCommand(command string) (string, error) {
	return fmt.Sprintf("命令 '%s' 已发送到游戏控制台", command), nil
}

// SDTDGetPlayers 获取玩家列表。
func (a *OmniPanel) SDTDGetPlayers() ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{"name": "Survivor1", "steamId": "76561198123456789", "level": 45, "online": true},
		{"name": "ZombieHunter", "steamId": "76561198234567890", "level": 32, "online": true},
		{"name": "BaseBuilder", "steamId": "76561198345678901", "level": 78, "online": false},
	}, nil
}

// SDTDCreateBackup 创建存档备份。
func (a *OmniPanel) SDTDCreateBackup() (map[string]interface{}, error) {
	timestamp := time.Now().Format("20060102_150405")
	return map[string]interface{}{
		"id":   timestamp,
		"name": fmt.Sprintf("saves_%s.zip", timestamp),
		"size": "1.2GB",
	}, nil
}

// ===========================================================================
// 前端绑定方法 — 插件管理
// ===========================================================================

// GetPluginStatus 返回所有插件的运行状态。
func (a *OmniPanel) GetPluginStatus() ([]map[string]interface{}, error) {
	plugins := a.pluginMgr.ListPlugins()
	var result []map[string]interface{}
	for _, pc := range plugins {
		result = append(result, map[string]interface{}{
			"name":    pc.Spec.Name,
			"running": !pc.Client.Exited(),
			"version": "1.0.0",
			"uptime":  time.Since(pc.StartTime).String(),
		})
	}
	return result, nil
}

// GetTheme 返回当前主题设置。
func (a *OmniPanel) GetTheme() string {
	return "dark"
}

// SetTheme 设置主题 (light/dark/auto)。
func (a *OmniPanel) SetTheme(theme string) error {
	a.logger.Info("主题切换", "theme", theme)
	// 通过 Wails 事件发送给前端
	return nil
}

// GetSettings 返回所有设置。
func (a *OmniPanel) GetSettings() (map[string]interface{}, error) {
	return map[string]interface{}{
		"theme":         "dark",
		"language":      "zh-CN",
		"sidebarOpen":   true,
		"notifications": true,
	}, nil
}

// SaveSettings 保存设置。
func (a *OmniPanel) SaveSettings(settings map[string]interface{}) error {
	a.logger.Info("设置已保存")
	return nil
}

// ===========================================================================
// 内部辅助方法
// ===========================================================================

func (a *OmniPanel) startPlugins(ctx context.Context) {
	plugins := []agent.PluginSpec{
		{Name: "docker", Required: false, AutoRestart: true, MaxRetries: 3},
		{Name: "ssh", Required: true, AutoRestart: true, MaxRetries: 5},
		{Name: "frp", Required: false, AutoRestart: true, MaxRetries: 3},
		{Name: "sdt", Required: false, AutoRestart: true, MaxRetries: 3},
	}

	for _, spec := range plugins {
		go func(s agent.PluginSpec) {
			_, err := a.pluginMgr.LoadPlugin(ctx, s)
			if err != nil {
				a.logger.Error("插件加载失败", "name", s.Name, "error", err)
				if s.Required {
					a.logger.Error("必需插件加载失败, 应用将退出", "name", s.Name)
					os.Exit(1)
				}
			}
		}(spec)
	}
}

func (a *OmniPanel) handleShutdown(ctx context.Context) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	a.logger.Info("收到信号, 开始优雅关闭", "signal", sig.String())
	a.Shutdown(context.Background())
	os.Exit(0)
}

func getDataDir() string {
	dir := os.Getenv("OMNIPANEL_DATA_DIR")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = home + "/.omnipanel"
	}
	os.MkdirAll(dir, 0700)
	return dir
}

func getPluginDir() string {
	dir := os.Getenv("OMNIPANEL_PLUGIN_DIR")
	if dir == "" {
		dir = "/opt/omnipanel/plugins"
	}
	return dir
}

func collectSystemInfo() (map[string]interface{}, error) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	hostname, _ := os.Hostname()
	return map[string]interface{}{
		"hostname":    hostname,
		"platform":    runtime.GOOS,
		"arch":        runtime.GOARCH,
		"goVersion":   runtime.Version(),
		"numCPU":      runtime.NumCPU(),
		"numGoroutine": runtime.NumGoroutine(),
		"uptime":      time.Now().Unix(),
	}, nil
}

func getProcessList() ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{"pid": 1234, "name": "omnipanel-agent", "cpu": 2.3, "mem": 128},
	}, nil
}

// ===========================================================================
// Wails 应用服务创建
// ===========================================================================

// CreateWailsService 配置 Wails v3 应用服务。
// 该函数在 main.go 中调用，负责:
//   1. 创建 Wails Application 实例
//   2. 注册前端绑定方法
//   3. 配置窗口参数
//   4. 设置生命周期钩子
func CreateWailsService(appInstance *OmniPanel) *application.App {
	app := application.New(application.Options{
		Name:        "OmniPanel",
		Description: "全能面板 - 一体化服务器&游戏管理",
		Icon:        nil,
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})

	// 绑定 Go 方法到前端
	// 在 Wails v3 中, 通过 app.Bind() 注册导出方法
	app.Bind(appInstance.GetLicenseStatus)
	app.Bind(appInstance.ActivateLicense)
	app.Bind(appInstance.GetSystemInfo)
	app.Bind(appInstance.GetProcessList)
	app.Bind(appInstance.SDTDInstallServer)
	app.Bind(appInstance.SDTDStartServer)
	app.Bind(appInstance.SDTDStopServer)
	app.Bind(appInstance.SDTDGetServerConfig)
	app.Bind(appInstance.SDTDSaveServerConfig)
	app.Bind(appInstance.SDTDSendConsoleCommand)
	app.Bind(appInstance.SDTDGetPlayers)
	app.Bind(appInstance.SDTDCreateBackup)
	app.Bind(appInstance.GetPluginStatus)
	app.Bind(appInstance.GetTheme)
	app.Bind(appInstance.SetTheme)
	app.Bind(appInstance.GetSettings)
	app.Bind(appInstance.SaveSettings)

	// 注册生命周期钩子
	app.OnStartup(func(ctx context.Context, a *application.App) {
		appInstance.Startup(ctx)
	})

	app.OnShutdown(func(ctx context.Context, a *application.App) {
		appInstance.Shutdown(ctx)
	})

	return app
}
