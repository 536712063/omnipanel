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
	"github.com/omnipanel/omnipanel/internal/ai"
	"github.com/omnipanel/omnipanel/internal/browser"
	"github.com/omnipanel/omnipanel/internal/cloud"
	"github.com/omnipanel/omnipanel/internal/common"
	"github.com/omnipanel/omnipanel/internal/dtd"
	"github.com/omnipanel/omnipanel/internal/git"
	"github.com/omnipanel/omnipanel/internal/i18n"
	"github.com/omnipanel/omnipanel/internal/license"
	"github.com/omnipanel/omnipanel/internal/plugins"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type OmniPanel struct {
	ctx          context.Context
	pluginMgr    *agent.Manager
	licenser     *license.Licenser
	wsHub        *WebSocketHub
	logger       hclog.Logger

	aiSvc        *ai.Service
	cloudSvc     *cloud.Service
	dtdSvc       *dtd.Service
	gitSvc       *git.Service
	i18nSvc      *i18n.Service
	browserSvc   *browser.Service
	appLogger    *common.Logger
	pluginMarket *plugins.Marketplace
	configSync   *common.ConfigSync

	mu sync.RWMutex
}

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

func (a *OmniPanel) Startup(ctx context.Context) {
	a.ctx = ctx
	a.logger.Info("OmniPanel 正在启动...")

	status, err := a.licenser.Validate("")
	if err != nil {
		a.logger.Warn("License 校验失败", "error", err)
	} else {
		a.logger.Info("License 状态", "plan", status.Plan, "days_remaining", status.DaysRemaining)
	}

	a.initializeServices(ctx)
	a.startPlugins(ctx)
	go a.wsHub.Run(ctx)
	go a.handleShutdown(ctx)

	a.logger.Info("OmniPanel 启动完成")
}

func (a *OmniPanel) Shutdown(ctx context.Context) {
	a.logger.Info("OmniPanel 正在关闭...")
	if a.dtdSvc != nil {
		a.dtdSvc.Stop()
	}
	if err := a.pluginMgr.ShutdownAll(30 * time.Second); err != nil {
		a.logger.Error("插件关闭超时", "error", err)
	}
	a.wsHub.Shutdown()
	a.logger.Info("OmniPanel 已关闭")
}

func (a *OmniPanel) initializeServices(ctx context.Context) {
	wailsEmitter := &WailsEventsEmitter{ctx: ctx}

	a.appLogger, _ = common.NewLogger(getDataDir() + "/logs")

	aiProvider := ai.NewOpenAICompatibleProvider(ai.ProviderConfig{
		Name:    "openai",
		BaseURL: os.Getenv("OPENAI_BASE_URL"),
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		Model:   "gpt-4o",
	})
	a.aiSvc = ai.NewService(aiProvider, wailsEmitter)

	a.cloudSvc = cloud.NewService(wailsEmitter, cloud.NewMultiMachineAdapter(nil, nil))

	dtdBridge := &dtdBridgeAdapter{}
	a.dtdSvc = dtd.NewService(wailsEmitter, dtdBridge)

	a.gitSvc = git.NewService()
	a.i18nSvc = i18n.NewService()
	a.browserSvc = browser.NewService()

	a.pluginMarket = plugins.NewMarketplace(getDataDir() + "/plugins")
	_ = a.pluginMarket.LoadFromDir()

	localStore := common.NewLocalConfigStore(getDataDir() + "/config.json")
	a.configSync = common.NewConfigSync(localStore, nil, common.CloudSyncConfig{})
}

type WailsEventsEmitter struct {
	ctx context.Context
}

func (e *WailsEventsEmitter) EmitEvent(eventName string, optionalData ...interface{}) {
	app := application.Get(e.ctx)
	if app != nil {
		app.EmitEvent(eventName, optionalData...)
	}
}

type dtdBridgeAdapter struct{}

func (b *dtdBridgeAdapter) ExecuteCommand(ctx context.Context, machineID, command string) (string, error) {
	return "", fmt.Errorf("multi-machine bridge not configured")
}
func (b *dtdBridgeAdapter) UploadToMachine(ctx context.Context, machineID, localPath, remotePath string) error {
	return fmt.Errorf("multi-machine bridge not configured")
}
func (b *dtdBridgeAdapter) DownloadFromMachine(ctx context.Context, machineID, remotePath, localPath string) error {
	return fmt.Errorf("multi-machine bridge not configured")
}

// ===========================================================================
// License
// ===========================================================================

func (a *OmniPanel) GetLicenseStatus() (*license.LicenseStatus, error) {
	return a.licenser.Validate("")
}

func (a *OmniPanel) ActivateLicense(licenseKey, email string) (*license.LicenseStatus, error) {
	return a.licenser.Activate(licenseKey, email)
}

// ===========================================================================
// System
// ===========================================================================

func (a *OmniPanel) GetSystemInfo() (map[string]interface{}, error) {
	return collectSystemInfo()
}

func (a *OmniPanel) GetProcessList() ([]map[string]interface{}, error) {
	return getProcessList()
}

// ===========================================================================
// AI Assistant
// ===========================================================================

func (a *OmniPanel) AIChat(chatReq ai.ChatRequest) (ai.ChatResponse, error) {
	return a.aiSvc.SendMessage(chatReq)
}

func (a *OmniPanel) AIChatStream(chatReq ai.ChatRequest) error {
	return a.aiSvc.SendMessageStream(chatReq)
}

func (a *OmniPanel) AIUploadFile(name string, data []byte) (ai.ContentPart, error) {
	return a.aiSvc.UploadFile(name, data)
}

func (a *OmniPanel) AIGetHistory(sessionID string) []ai.ChatMessage {
	return a.aiSvc.GetSessionHistory(sessionID)
}

// ===========================================================================
// Cloud Storage
// ===========================================================================

func (a *OmniPanel) CloudListFiles(providerName, path string) ([]cloud.FileInfo, error) {
	return a.cloudSvc.ListFiles(providerName, path)
}

func (a *OmniPanel) CloudCopyFile(providerName, srcPath, dstPath string) error {
	return a.cloudSvc.CopyFile(providerName, srcPath, dstPath)
}

func (a *OmniPanel) CloudMoveFile(providerName, srcPath, dstPath string) error {
	return a.cloudSvc.MoveFile(providerName, srcPath, dstPath)
}

func (a *OmniPanel) CloudRenameFile(providerName, path, newName string) error {
	return a.cloudSvc.RenameFile(providerName, path, newName)
}

func (a *OmniPanel) CloudDeleteFile(providerName, path string) error {
	return a.cloudSvc.DeleteFile(providerName, path)
}

func (a *OmniPanel) CloudMkdir(providerName, path string) error {
	return a.cloudSvc.Mkdir(providerName, path)
}

func (a *OmniPanel) CloudUploadFile(providerName, localPath, remotePath string) (string, error) {
	return a.cloudSvc.UploadFile(providerName, localPath, remotePath)
}

func (a *OmniPanel) CloudDownloadFile(providerName, remotePath, localPath string) (string, error) {
	return a.cloudSvc.DownloadFile(providerName, remotePath, localPath)
}

func (a *OmniPanel) CloudGetPreview(providerName, path string) (cloud.PreviewInfo, error) {
	return a.cloudSvc.GetPreview(providerName, path)
}

func (a *OmniPanel) CloudOAuthURL(providerName string) (string, error) {
	return a.cloudSvc.OAuthURL(providerName)
}

func (a *OmniPanel) CloudHandleOAuthCallback(providerName, code, state string) (*cloud.OAuthToken, error) {
	return a.cloudSvc.HandleOAuthCallback(providerName, code, state)
}

func (a *OmniPanel) CloudAddProvider(cfg cloud.ProviderConfig) error {
	return a.cloudSvc.AddProviderFromConfig(cfg)
}

func (a *OmniPanel) CloudListProviders() []string {
	return a.cloudSvc.ListProviders()
}

func (a *OmniPanel) CloudSyncFromMachine(req cloud.MultiMachineSyncRequest) (string, error) {
	return a.cloudSvc.SyncFromMachine(req)
}

// ===========================================================================
// 7DTD Server Management
// ===========================================================================

func (a *OmniPanel) DTDAddServer(cfg dtd.ServerConfig) error {
	return a.dtdSvc.AddServer(cfg)
}

func (a *OmniPanel) DTDListServers() []dtd.ServerConfig {
	return a.dtdSvc.ListServers()
}

func (a *OmniPanel) DTDConnect(serverID string) error {
	return a.dtdSvc.Connect(serverID)
}

func (a *OmniPanel) DTDDisconnect(serverID string) error {
	return a.dtdSvc.Disconnect(serverID)
}

func (a *OmniPanel) DTDSendCommand(serverID, cmd string) error {
	return a.dtdSvc.SendCommand(serverID, cmd)
}

func (a *OmniPanel) DTDGetPlayers(serverID string) []dtd.OnlinePlayer {
	return a.dtdSvc.GetPlayers(serverID)
}

func (a *OmniPanel) DTDGetPlayerEvents(serverID string, limit int) []dtd.PlayerEvent {
	return a.dtdSvc.GetPlayerEvents(serverID, limit)
}

func (a *OmniPanel) DTDGetConsoleMessages(serverID string, limit int) []dtd.ConsoleMessage {
	return a.dtdSvc.GetConsoleMessages(serverID, limit)
}

func (a *OmniPanel) DTDParseLogFile(serverID, logPath string) error {
	return a.dtdSvc.ParseLogFile(serverID, logPath)
}

func (a *OmniPanel) DTDGetMapData(serverID string) (*dtd.MapData, error) {
	return a.dtdSvc.GetMapData(serverID)
}

func (a *OmniPanel) DTDGetBloodMoonInfo(serverID string) (*dtd.BloodMoonInfo, error) {
	return a.dtdSvc.GetBloodMoonInfo(serverID)
}

func (a *OmniPanel) DTDAddScheduledTask(task dtd.ScheduledTask) (*dtd.ScheduledTask, error) {
	return a.dtdSvc.AddScheduledTask(task)
}

func (a *OmniPanel) DTDRemoveScheduledTask(taskID string) {
	a.dtdSvc.RemoveScheduledTask(taskID)
}

func (a *OmniPanel) DTDGetScheduledTasks() []dtd.ScheduledTask {
	return a.dtdSvc.GetScheduledTasks()
}

func (a *OmniPanel) DTDDeployToMachine(req dtd.RemoteDeployRequest) (*dtd.RemoteCommandResult, error) {
	return a.dtdSvc.DeployToMachine(req)
}

func (a *OmniPanel) DTDStartRemoteServer(machineID, installedDir string) (*dtd.RemoteCommandResult, error) {
	return a.dtdSvc.StartRemoteServer(machineID, installedDir)
}

// ===========================================================================
// Git Repository Management
// ===========================================================================

func (a *OmniPanel) GitAddLocalRepo(path string) (*git.Repo, error) {
	return a.gitSvc.AddLocalRepo(path)
}

func (a *OmniPanel) GitCloneRepo(req git.CloneRequest) (*git.Repo, error) {
	return a.gitSvc.CloneRepo(req)
}

func (a *OmniPanel) GitListRepos() []git.Repo {
	return a.gitSvc.ListRepos()
}

func (a *OmniPanel) GitRemoveRepo(id string) {
	a.gitSvc.RemoveRepo(id)
}

func (a *OmniPanel) GitBranches(repoID string) ([]git.Branch, error) {
	return a.gitSvc.Branches(repoID)
}

func (a *OmniPanel) GitLog(repoID string, limit int) ([]git.Commit, error) {
	return a.gitSvc.Log(repoID, limit)
}

func (a *OmniPanel) GitStatus(repoID string) ([]git.StatusItem, error) {
	return a.gitSvc.Status(repoID)
}

func (a *OmniPanel) GitPull(repoID string) error {
	return a.gitSvc.Pull(a.ctx, repoID)
}

func (a *OmniPanel) GitPush(repoID string) error {
	return a.gitSvc.Push(a.ctx, repoID)
}

func (a *OmniPanel) GitCommit(repoID, message string) (string, error) {
	return a.gitSvc.Commit(repoID, message)
}

func (a *OmniPanel) GitCheckout(repoID, branchName string) error {
	return a.gitSvc.Checkout(repoID, branchName)
}

func (a *OmniPanel) GitCreateBranch(repoID, name string) error {
	return a.gitSvc.CreateBranch(repoID, name)
}

func (a *OmniPanel) GitGetRepoStats(repoID string) (map[string]interface{}, error) {
	return a.gitSvc.GetRepoStats(repoID)
}

// ===========================================================================
// I18n Tool
// ===========================================================================

func (a *OmniPanel) I18nExtract(req i18n.ExtractRequest) (*i18n.ExtractResult, error) {
	return a.i18nSvc.Extract(req)
}

func (a *OmniPanel) I18nGenerateLocaleFile(items []i18n.ExtractedItem, locale, outputPath string) error {
	return a.i18nSvc.GenerateLocaleFile(items, locale, outputPath)
}

func (a *OmniPanel) I18nPreviewTranslation(filePath string, items []i18n.ExtractedItem) (string, error) {
	return a.i18nSvc.PreviewTranslation(filePath, items)
}

func (a *OmniPanel) I18nApplyTranslationFile(filePath string, items []i18n.ExtractedItem) error {
	return a.i18nSvc.ApplyTranslationFile(filePath, items)
}

func (a *OmniPanel) I18nBatchApplyTranslation(result i18n.ExtractResult) error {
	return a.i18nSvc.BatchApplyTranslation(result)
}

// ===========================================================================
// Built-in Browser
// ===========================================================================

func (a *OmniPanel) BrowserGetInfo() browser.BrowserInfo {
	return a.browserSvc.GetInfo()
}

func (a *OmniPanel) BrowserOpenExternalURL(url string) error {
	return a.browserSvc.OpenExternalURL(a.ctx, url)
}

func (a *OmniPanel) BrowserValidateURL(url string) error {
	return a.browserSvc.ValidateURL(url)
}

// ===========================================================================
// Plugin Marketplace
// ===========================================================================

func (a *OmniPanel) PluginList() []plugins.Manifest {
	return a.pluginMarket.ListPlugins()
}

func (a *OmniPanel) PluginExecute(id string, execCtx map[string]interface{}) (map[string]interface{}, error) {
	return a.pluginMarket.ExecuteAgent(id, execCtx)
}

func (a *OmniPanel) PluginEnable(id string) {
	a.pluginMarket.EnablePlugin(id)
}

func (a *OmniPanel) PluginDisable(id string) {
	a.pluginMarket.DisablePlugin(id)
}

// ===========================================================================
// Config Sync
// ===========================================================================

func (a *OmniPanel) ConfigGetSyncConfig() map[string]interface{} {
	return a.configSync.GetSyncStatus()
}

func (a *OmniPanel) ConfigSaveSyncConfig(cfg common.CloudSyncConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.configSync = common.NewConfigSync(a.configSync.GetLocalStore(), a.configSync.GetRemoteStore(), cfg)
	return nil
}

func (a *OmniPanel) ConfigPullFromCloud() (map[string]interface{}, error) {
	return a.configSync.PullFromCloud()
}

func (a *OmniPanel) ConfigPushToCloud() error {
	return a.configSync.PushToCloud()
}

func (a *OmniPanel) ConfigGetLocal() (map[string]interface{}, error) {
	return a.configSync.GetConfig()
}

func (a *OmniPanel) ConfigSetLocal(data map[string]interface{}) error {
	return a.configSync.SetConfig(data)
}

// ===========================================================================
// Settings / Theme
// ===========================================================================

func (a *OmniPanel) GetTheme() string {
	return "dark"
}

func (a *OmniPanel) SetTheme(theme string) error {
	a.logger.Info("主题切换", "theme", theme)
	return nil
}

func (a *OmniPanel) GetSettings() (map[string]interface{}, error) {
	return map[string]interface{}{
		"theme":         "dark",
		"language":      "zh-CN",
		"sidebarOpen":   true,
		"notifications": true,
	}, nil
}

func (a *OmniPanel) SaveSettings(settings map[string]interface{}) error {
	a.logger.Info("设置已保存")
	return nil
}

// ===========================================================================
// Plugin Status (keep legacy)
// ===========================================================================

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

// ===========================================================================
// Internal Helpers
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
		"hostname":     hostname,
		"platform":     runtime.GOOS,
		"arch":         runtime.GOARCH,
		"goVersion":    runtime.Version(),
		"numCPU":       runtime.NumCPU(),
		"numGoroutine": runtime.NumGoroutine(),
		"uptime":       time.Now().Unix(),
	}, nil
}

func getProcessList() ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{"pid": 1234, "name": "omnipanel-agent", "cpu": 2.3, "mem": 128},
	}, nil
}

// ===========================================================================
// Wails Application Service Creation
// ===========================================================================

func CreateWailsService(appInstance *OmniPanel) *application.App {
	app := application.New(application.Options{
		Name:        "OmniPanel",
		Description: "全能面板 - 一体化服务器&游戏管理",
		Icon:        nil,
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})

	app.Bind(appInstance.GetLicenseStatus)
	app.Bind(appInstance.ActivateLicense)
	app.Bind(appInstance.GetSystemInfo)
	app.Bind(appInstance.GetProcessList)

	app.Bind(appInstance.AIChat)
	app.Bind(appInstance.AIChatStream)
	app.Bind(appInstance.AIUploadFile)
	app.Bind(appInstance.AIGetHistory)

	app.Bind(appInstance.CloudListFiles)
	app.Bind(appInstance.CloudCopyFile)
	app.Bind(appInstance.CloudMoveFile)
	app.Bind(appInstance.CloudRenameFile)
	app.Bind(appInstance.CloudDeleteFile)
	app.Bind(appInstance.CloudMkdir)
	app.Bind(appInstance.CloudUploadFile)
	app.Bind(appInstance.CloudDownloadFile)
	app.Bind(appInstance.CloudGetPreview)
	app.Bind(appInstance.CloudOAuthURL)
	app.Bind(appInstance.CloudHandleOAuthCallback)
	app.Bind(appInstance.CloudAddProvider)
	app.Bind(appInstance.CloudListProviders)
	app.Bind(appInstance.CloudSyncFromMachine)

	app.Bind(appInstance.DTDAddServer)
	app.Bind(appInstance.DTDListServers)
	app.Bind(appInstance.DTDConnect)
	app.Bind(appInstance.DTDDisconnect)
	app.Bind(appInstance.DTDSendCommand)
	app.Bind(appInstance.DTDGetPlayers)
	app.Bind(appInstance.DTDGetPlayerEvents)
	app.Bind(appInstance.DTDGetConsoleMessages)
	app.Bind(appInstance.DTDParseLogFile)
	app.Bind(appInstance.DTDGetMapData)
	app.Bind(appInstance.DTDGetBloodMoonInfo)
	app.Bind(appInstance.DTDAddScheduledTask)
	app.Bind(appInstance.DTDRemoveScheduledTask)
	app.Bind(appInstance.DTDGetScheduledTasks)
	app.Bind(appInstance.DTDDeployToMachine)
	app.Bind(appInstance.DTDStartRemoteServer)

	app.Bind(appInstance.GitAddLocalRepo)
	app.Bind(appInstance.GitCloneRepo)
	app.Bind(appInstance.GitListRepos)
	app.Bind(appInstance.GitRemoveRepo)
	app.Bind(appInstance.GitBranches)
	app.Bind(appInstance.GitLog)
	app.Bind(appInstance.GitStatus)
	app.Bind(appInstance.GitPull)
	app.Bind(appInstance.GitPush)
	app.Bind(appInstance.GitCommit)
	app.Bind(appInstance.GitCheckout)
	app.Bind(appInstance.GitCreateBranch)
	app.Bind(appInstance.GitGetRepoStats)

	app.Bind(appInstance.I18nExtract)
	app.Bind(appInstance.I18nGenerateLocaleFile)
	app.Bind(appInstance.I18nPreviewTranslation)
	app.Bind(appInstance.I18nApplyTranslationFile)
	app.Bind(appInstance.I18nBatchApplyTranslation)

	app.Bind(appInstance.BrowserGetInfo)
	app.Bind(appInstance.BrowserOpenExternalURL)
	app.Bind(appInstance.BrowserValidateURL)

	app.Bind(appInstance.PluginList)
	app.Bind(appInstance.PluginExecute)
	app.Bind(appInstance.PluginEnable)
	app.Bind(appInstance.PluginDisable)

	app.Bind(appInstance.ConfigGetSyncConfig)
	app.Bind(appInstance.ConfigSaveSyncConfig)
	app.Bind(appInstance.ConfigPullFromCloud)
	app.Bind(appInstance.ConfigPushToCloud)
	app.Bind(appInstance.ConfigGetLocal)
	app.Bind(appInstance.ConfigSetLocal)

	app.Bind(appInstance.GetTheme)
	app.Bind(appInstance.SetTheme)
	app.Bind(appInstance.GetSettings)
	app.Bind(appInstance.SaveSettings)
	app.Bind(appInstance.GetPluginStatus)

	app.OnStartup(func(ctx context.Context, a *application.App) {
		appInstance.Startup(ctx)
	})

	app.OnShutdown(func(ctx context.Context, a *application.App) {
		appInstance.Shutdown(ctx)
	})

	return app
}
