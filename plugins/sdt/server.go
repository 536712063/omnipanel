package main

import (
	"bufio"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ===========================================================================
// XML 数据模型 -- 映射 serverconfig.xml 结构
// ===========================================================================

type ServerConfigXML struct {
	XMLName xml.Name `xml:"ServerSettings"`

	// 基本设置
	ServerName          string `xml:"property[@name='ServerName'][@value]"`
	ServerDescription   string `xml:"property[@name='ServerDescription'][@value]"`
	ServerPassword      string `xml:"property[@name='ServerPassword'][@value]"`
	ServerMaxPlayerCount int   `xml:"property[@name='ServerMaxPlayerCount'][@value]"`
	ServerPort          int    `xml:"property[@name='ServerPort'][@value]"`
	ServerVisibility    int    `xml:"property[@name='ServerVisibility'][@value]"`

	// 游戏设置
	GameDifficulty      int    `xml:"property[@name='GameDifficulty'][@value]"`
	GameMode            string `xml:"property[@name='GameMode'][@value]"`
	WorldGenSeed        string `xml:"property[@name='WorldGenSeed'][@value]"`
	WorldGenSize        int    `xml:"property[@name='WorldGenSize'][@value]"`
	DayNightLength      int    `xml:"property[@name='DayNightLength'][@value]"`
	MaxSpawnedZombies   int    `xml:"property[@name='MaxSpawnedZombies'][@value]"`
	MaxSpawnedAnimals   int    `xml:"property[@name='MaxSpawnedAnimals'][@value]"`
	EACEnabled          bool   `xml:"property[@name='EACEnabled'][@value]"`
	BloodMoonFrequency  int    `xml:"property[@name='BloodMoonFrequency'][@value]"`
	AirDropFrequency    int    `xml:"property[@name='AirDropFrequency'][@value]"`
	DropOnDeath         int    `xml:"property[@name='DropOnDeath'][@value]"`
	DropOnQuit          int    `xml:"property[@name='DropOnQuit'][@value]"`

	// Telnet
	TelnetEnabled  bool   `xml:"property[@name='TelnetEnabled'][@value]"`
	TelnetPort     int    `xml:"property[@name='TelnetPort'][@value]"`
	TelnetPassword string `xml:"property[@name='TelnetPassword'][@value]"`
}

// ===========================================================================
// 配置读取/写入
// ===========================================================================

func (p *SDTDPlugin) readServerConfig() (*ServerConfigXML, error) {
	data, err := os.ReadFile(p.config.ServerConfig)
	if err != nil {
		return nil, fmt.Errorf("无法读取 serverconfig.xml: %w", err)
	}

	var cfg ServerConfigXML
	if err := xml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("XML 解析失败: %w", err)
	}
	return &cfg, nil
}

func (p *SDTDPlugin) writeServerConfig(cfg *ServerConfigXML) error {
	if err := os.MkdirAll(filepath.Dir(p.config.ServerConfig), 0755); err != nil {
		return err
	}

	// 备份旧配置
	backup := p.config.ServerConfig + ".bak"
	if _, err := os.Stat(p.config.ServerConfig); err == nil {
		data, _ := os.ReadFile(p.config.ServerConfig)
		os.WriteFile(backup, data, 0644)
	}

	data, err := xml.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("XML 序列化失败: %w", err)
	}

	xmlHeader := []byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	full := append(xmlHeader, data...)
	return os.WriteFile(p.config.ServerConfig, full, 0644)
}

func (p *SDTDPlugin) validateConfig(cfg *ServerConfigXML) ([]string, []string) {
	var errors, warnings []string

	if cfg.ServerPort < 1 || cfg.ServerPort > 65535 {
		errors = append(errors, "服务器端口必须在 1-65535 之间")
	}
	if cfg.ServerMaxPlayerCount < 1 || cfg.ServerMaxPlayerCount > 64 {
		errors = append(errors, "最大玩家数必须在 1-64 之间")
	}
	if cfg.WorldGenSize < 1024 || cfg.WorldGenSize > 16384 {
		errors = append(errors, "世界大小必须在 1024-16384 之间")
	}
	if cfg.MaxSpawnedZombies > 200 {
		warnings = append(warnings, "僵尸生成数超过 200 可能影响服务器性能")
	}
	if cfg.DayNightLength < 10 {
		warnings = append(warnings, "昼夜周期过短可能导致频繁的血月")
	}

	return errors, warnings
}

// ===========================================================================
// 服务端生命周期管理
// ===========================================================================

func (p *SDTDPlugin) startServer(ctx context.Context) error {
	if p.serverCmd != nil && p.serverCmd.Process != nil {
		return fmt.Errorf("服务器已在运行中")
	}

	executable := filepath.Join(p.config.InstallDir, "7DaysToDieServer.x86_64")
	if _, err := os.Stat(executable); os.IsNotExist(err) {
		return fmt.Errorf("服务端可执行文件未找到: %s — 请先运行安装", executable)
	}

	ctxWithCancel, cancel := context.WithCancel(ctx)
	p.cancel = cancel

	cmd := exec.CommandContext(ctxWithCancel, executable,
		"-configfile="+p.config.ServerConfig,
		"-logfile="+filepath.Join(p.config.InstallDir, "logs", "7dtd_output.log"),
		"-quit",
		"-batchmode",
		"-nographics",
		"-dedicated",
	)
	cmd.Dir = p.config.InstallDir
	cmd.Env = append(os.Environ(),
		"LD_LIBRARY_PATH="+p.config.InstallDir+"/linux64:"+os.Getenv("LD_LIBRARY_PATH"),
	)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	p.serverCmd = cmd

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动服务端失败: %w", err)
	}
	p.logger.Info("游戏服务端已启动", "pid", cmd.Process.Pid)

	go p.monitorProcess(ctxWithCancel, cmd)
	go p.streamOutput(stdout, "stdout")
	go p.streamOutput(stderr, "stderr")

	return nil
}

func (p *SDTDPlugin) stopServer() error {
	if p.serverCmd == nil || p.serverCmd.Process == nil {
		return fmt.Errorf("服务器未在运行")
	}

	// 先尝试优雅关闭 (发送 "shutdown" 命令)
	p.sendCommand("shutdown")

	done := make(chan error, 1)
	go func() { done <- p.serverCmd.Wait() }()

	select {
	case <-done:
		p.logger.Info("游戏服务端已优雅关闭")
	case <-time.After(30 * time.Second):
		p.serverCmd.Process.Kill()
		p.logger.Warn("游戏服务端超时, 已强制终止")
	}

	p.serverCmd = nil
	p.cancel = nil
	return nil
}

func (p *SDTDPlugin) restartServer(ctx context.Context) error {
	if err := p.stopServer(); err != nil {
		p.logger.Warn("停止服务端时出错, 继续重启", "error", err)
	}
	time.Sleep(3 * time.Second)
	return p.startServer(ctx)
}

func (p *SDTDPlugin) isServerRunning() bool {
	if p.serverCmd == nil || p.serverCmd.Process == nil {
		return false
	}
	if p.serverCmd.ProcessState != nil && p.serverCmd.ProcessState.Exited() {
		return false
	}
	return true
}

func (p *SDTDPlugin) monitorProcess(ctx context.Context, cmd *exec.Cmd) {
	<-ctx.Done()
	if cmd.Process != nil {
		cmd.Process.Signal(syscall.SIGTERM)
	}
}

func (p *SDTDPlugin) streamOutput(pipe io.ReadCloser, source string) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		p.logger.Info(scanner.Text(), "source", source)
	}
}

// ===========================================================================
// 控制台命令
// ===========================================================================

func (p *SDTDPlugin) sendCommand(command string) (string, error) {
	cfg, err := p.readServerConfig()
	if err != nil {
		return "", err
	}

	if !cfg.TelnetEnabled {
		return "", fmt.Errorf("Telnet 未启用, 请在 serverconfig.xml 中设置 TelnetEnabled=true")
	}

	// 实际实现: 通过 TCP 连接到 Telnet 端口发送命令
	// 此处展示核心逻辑
	conn, err := net.DialTimeout("tcp",
		fmt.Sprintf("127.0.0.1:%d", cfg.TelnetPort),
		10*time.Second,
	)
	if err != nil {
		return "", fmt.Errorf("Telnet 连接失败: %w", err)
	}
	defer conn.Close()

	// 发送认证
	fmt.Fprintf(conn, "%s\r\n", cfg.TelnetPassword)
	time.Sleep(500 * time.Millisecond)

	// 发送命令
	fmt.Fprintf(conn, "%s\r\n", command)
	time.Sleep(1 * time.Second)

	// 读取响应
	var buf [4096]byte
	n, _ := conn.Read(buf[:])
	return strings.TrimSpace(string(buf[:n])), nil
}

// ===========================================================================
// SteamCMD 安装/更新
// ===========================================================================

func (p *SDTDPlugin) installViaSteamCMD(ctx context.Context, username, password string) (<-chan string, error) {
	outputCh := make(chan string, 100)

	if _, err := os.Stat(p.config.SteamCMDPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("SteamCMD 未找到: %s", p.config.SteamCMDPath)
	}

	if err := os.MkdirAll(p.config.InstallDir, 0755); err != nil {
		return nil, fmt.Errorf("无法创建安装目录: %w", err)
	}

	go func() {
		defer close(outputCh)

		updateScript := filepath.Join(p.config.InstallDir, "update_script.txt")
		os.WriteFile(updateScript, []byte(fmt.Sprintf(
			`@ShutdownOnFailedCommand 1
@NoPromptForPassword 1
login %s %s
force_install_dir %s
app_update 294420 validate
quit
`, username, password, p.config.InstallDir)), 0644)

		cmd := exec.CommandContext(ctx, p.config.SteamCMDPath,
			"+runscript", updateScript,
		)
		cmd.Dir = p.config.InstallDir

		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		cmd.Start()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				line := scanner.Text()
				outputCh <- line
				p.logger.Info("steamcmd", "output", line)
			}
		}()

		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				line := scanner.Text()
				outputCh <- "ERR: " + line
				p.logger.Error("steamcmd", "error", line)
			}
		}()

		wg.Wait()
		cmd.Wait()
		p.logger.Info("SteamCMD 安装完成")
	}()

	return outputCh, nil
}

// ===========================================================================
// 存档备份
// ===========================================================================

func (p *SDTDPlugin) createBackup() (string, error) {
	if err := os.MkdirAll(p.config.BackupDir, 0755); err != nil {
		return "", err
	}

	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("saves_%s.zip", timestamp)
	backupPath := filepath.Join(p.config.BackupDir, backupName)

	// 使用系统 zip 命令打包存档
	cmd := exec.Command("zip", "-r", backupPath, p.config.SavesDir)
	cmd.Dir = p.config.InstallDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("备份失败: %w, 输出: %s", err, string(output))
	}

	p.logger.Info("存档备份已创建", "path", backupPath)
	return backupName, nil
}

func (p *SDTDPlugin) restoreBackup(backupID string) error {
	if !p.isServerRunning() {
		return fmt.Errorf("恢复存档前请确保服务器已关闭")
	}

	backupPath := filepath.Join(p.config.BackupDir, backupID)
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("备份文件不存在: %s", backupPath)
	}

	// 备份当前存档
	currentBackup, _ := p.createBackup()
	p.logger.Info("已备份当前存档", "backup", currentBackup)

	// 清空当前存档
	os.RemoveAll(p.config.SavesDir)

	// 解压备份
	cmd := exec.Command("unzip", "-o", backupPath, "-d", p.config.InstallDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("恢复备份失败: %w, 输出: %s", err, string(output))
	}

	p.logger.Info("存档已恢复", "backup", backupID)
	return nil
}

// ===========================================================================
// MOD 管理
// ===========================================================================

func (p *SDTDPlugin) listMods() ([]ModInfo, error) {
	modsDir := filepath.Join(p.config.InstallDir, "Mods")
	if _, err := os.Stat(modsDir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(modsDir)
	if err != nil {
		return nil, err
	}

	var mods []ModInfo
	for _, entry := range entries {
		if entry.IsDir() {
			mods = append(mods, ModInfo{
				Name:    entry.Name(),
				Enabled: true,
				Path:    filepath.Join(modsDir, entry.Name()),
			})
		}
	}
	return mods, nil
}

type ModInfo struct {
	Name    string
	Version string
	Author  string
	Enabled bool
	Path    string
}
