// SDTD (Seven Days To Die) Plugin
//
// 七日杀专用服务器管理插件, 作为独立 OS 进程运行。
// 通过 hashicorp/go-plugin 与 Agent 主进程通信。
//
// 功能:
//   - SteamCMD 一键部署/更新服务端
//   - 启动/停止/重启游戏进程
//   - 实时控制台交互
//   - serverconfig.xml 读写与验证
//   - 玩家管理 (踢出/封禁/传送/给予物品)
//   - 管理员管理 (saveserveradmin.xml)
//   - 存档自动备份与恢复
//   - MOD 管理
//   - 实时 TPS/FPS/内存监控
package main

import (
	"context"
	"fmt"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/omnipanel/omnipanel/plugins/sdk"
)

// SDTDPlugin 实现 sdk.BasePlugin 接口
type SDTDPlugin struct {
	config    *SDTDConfig
	serverCmd *exec.Cmd
	logger    hclog.Logger
	cancel    context.CancelFunc
}

// SDTDConfig 插件本地配置
type SDTDConfig struct {
	InstallDir   string `json:"install_dir"`
	ServerConfig string `json:"server_config_path"` // serverconfig.xml 路径
	AdminConfig  string `json:"admin_config_path"`  // saveserveradmin.xml 路径
	SavesDir     string `json:"saves_dir"`
	ModsDir      string `json:"mods_dir"`
	BackupDir    string `json:"backup_dir"`
	SteamCMDPath string `json:"steamcmd_path"`
}

func (p *SDTDPlugin) Name() string        { return "sdt" }
func (p *SDTDPlugin) Version() string     { return "1.0.0" }
func (p *SDTDPlugin) Description() string { return "七日杀 (7 Days to Die) 专用服务器管理插件" }

func (p *SDTDPlugin) HealthCheck(ctx context.Context) error {
	if p.serverCmd != nil && p.serverCmd.Process != nil {
		if p.serverCmd.ProcessState != nil && p.serverCmd.ProcessState.Exited() {
			return fmt.Errorf("server process has exited")
		}
	}
	return nil
}

func (p *SDTDPlugin) Shutdown(ctx context.Context) error {
	p.logger.Info("SDTD plugin shutting down")
	if p.cancel != nil {
		p.cancel()
	}
	if p.serverCmd != nil && p.serverCmd.Process != nil {
		if err := p.serverCmd.Process.Signal(syscall.SIGTERM); err != nil {
			p.logger.Warn("failed to send SIGTERM to game server", "error", err)
			p.serverCmd.Process.Kill()
		}

		done := make(chan error, 1)
		go func() { done <- p.serverCmd.Wait() }()
		select {
		case <-done:
		case <-time.After(30 * time.Second):
			p.serverCmd.Process.Kill()
			p.logger.Warn("game server did not exit gracefully, force killed")
		}
	}
	return nil
}

// ServeSDTDPlugin 启动 SDTD 插件 gRPC 服务
func ServeSDTDPlugin() {
	base := &SDTDPlugin{
		config: &SDTDConfig{
			InstallDir:   "/opt/omnipanel/sdt",
			ServerConfig: "/opt/omnipanel/sdt/serverconfig.xml",
			AdminConfig:  "/opt/omnipanel/sdt/saveserveradmin.xml",
			SavesDir:     "/opt/omnipanel/sdt/Saves",
			ModsDir:      "/opt/omnipanel/sdt/Mods",
			BackupDir:    "/opt/omnipanel/sdt/Backups",
			SteamCMDPath: "/usr/games/steamcmd",
		},
	}

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   base.Name(),
		Level:  hclog.LevelFromString("INFO"),
		Output: hclog.DefaultOutput,
	})
	base.logger = logger

	sdk.ServePlugin(base, &SDTDGRPCPlugin{impl: base, ctx: ctx})
}

func main() {
	ServeSDTDPlugin()
}
