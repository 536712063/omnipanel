/**
 * Wails v3 运行时绑定层 — OmniPanel 全模块 API
 *
 * 架构:
 *   Vue Component → runtime.ts → Wails IPC → Go app.go → Service
 *
 * Wails Events 用于流式数据 (AI 流式回复、DTD 控制台日志、Cloud 上传进度)
 */
declare global {
  interface Window {
    go: {
      main: {
        App: {
          GetLicenseStatus(): Promise<LicenseStatus>
          ActivateLicense(licenseKey: string, email: string): Promise<LicenseStatus>
          GetSystemInfo(): Promise<SystemInfo>
          GetProcessList(): Promise<ProcessInfo[]>

          AIChat(req: AIChatRequest): Promise<AIChatResponse>
          AIChatStream(req: AIChatRequest): Promise<void>
          AIUploadFile(name: string, data: number[]): Promise<AIContentPart>
          AIGetHistory(sessionID: string): Promise<AIChatMessage[]>

          CloudListFiles(provider: string, path: string): Promise<CloudFileInfo[]>
          CloudCopyFile(provider: string, src: string, dst: string): Promise<void>
          CloudMoveFile(provider: string, src: string, dst: string): Promise<void>
          CloudRenameFile(provider: string, path: string, name: string): Promise<void>
          CloudDeleteFile(provider: string, path: string): Promise<void>
          CloudMkdir(provider: string, path: string): Promise<void>
          CloudUploadFile(provider: string, local: string, remote: string): Promise<string>
          CloudDownloadFile(provider: string, remote: string, local: string): Promise<string>
          CloudGetPreview(provider: string, path: string): Promise<CloudPreviewInfo>
          CloudOAuthURL(provider: string): Promise<string>
          CloudHandleOAuthCallback(provider: string, code: string, state: string): Promise<CloudOAuthToken>
          CloudAddProvider(cfg: CloudProviderConfig): Promise<void>
          CloudListProviders(): Promise<string[]>
          CloudSyncFromMachine(req: CloudMachineSyncReq): Promise<string>

          DTDAddServer(cfg: DTDServerConfig): Promise<void>
          DTDListServers(): Promise<DTDServerConfig[]>
          DTDConnect(serverID: string): Promise<void>
          DTDDisconnect(serverID: string): Promise<void>
          DTDSendCommand(serverID: string, cmd: string): Promise<void>
          DTDGetPlayers(serverID: string): Promise<DTDOnlinePlayer[]>
          DTDGetPlayerEvents(serverID: string, limit: number): Promise<DTDPlayerEvent[]>
          DTDGetConsoleMessages(serverID: string, limit: number): Promise<DTDConsoleMessage[]>
          DTDParseLogFile(serverID: string, logPath: string): Promise<void>
          DTDGetMapData(serverID: string): Promise<DTDMapData | null>
          DTDGetBloodMoonInfo(serverID: string): Promise<DTDBloodMoonInfo>
          DTDAddScheduledTask(task: DTDScheduledTask): Promise<DTDScheduledTask>
          DTDRemoveScheduledTask(taskID: string): void
          DTDGetScheduledTasks(): Promise<DTDScheduledTask[]>
          DTDDeployToMachine(req: DTDRemoteDeployReq): Promise<DTDRemoteCmdResult>
          DTDStartRemoteServer(machineID: string, dir: string): Promise<DTDRemoteCmdResult>

          GitAddLocalRepo(path: string): Promise<GitRepo>
          GitCloneRepo(req: GitCloneRequest): Promise<GitRepo>
          GitListRepos(): Promise<GitRepo[]>
          GitRemoveRepo(id: string): void
          GitBranches(repoID: string): Promise<GitBranch[]>
          GitLog(repoID: string, limit: number): Promise<GitCommit[]>
          GitStatus(repoID: string): Promise<GitStatusItem[]>
          GitPull(repoID: string): Promise<void>
          GitPush(repoID: string): Promise<void>
          GitCommit(repoID: string, message: string): Promise<string>
          GitCheckout(repoID: string, branch: string): Promise<void>
          GitCreateBranch(repoID: string, name: string): Promise<void>
          GitGetRepoStats(repoID: string): Promise<Record<string, unknown>>

          I18nExtract(req: I18nExtractReq): Promise<I18nExtractResult>
          I18nGenerateLocaleFile(items: I18nItem[], locale: string, path: string): Promise<void>
          I18nPreviewTranslation(file: string, items: I18nItem[]): Promise<string>
          I18nApplyTranslationFile(file: string, items: I18nItem[]): Promise<void>
          I18nBatchApplyTranslation(result: I18nExtractResult): Promise<void>

          BrowserGetInfo(): Promise<BrowserInfo>
          BrowserOpenExternalURL(url: string): Promise<void>
          BrowserValidateURL(url: string): Promise<void>

          PluginList(): Promise<PluginManifest[]>
          PluginExecute(id: string, ctx: Record<string, unknown>): Promise<Record<string, unknown>>
          PluginEnable(id: string): void
          PluginDisable(id: string): void

          ConfigGetSyncConfig(): Promise<CloudSyncStatus>
          ConfigSaveSyncConfig(cfg: CloudSyncConfig): Promise<void>
          ConfigPullFromCloud(): Promise<Record<string, unknown>>
          ConfigPushToCloud(): Promise<void>
          ConfigGetLocal(): Promise<Record<string, unknown>>
          ConfigSetLocal(data: Record<string, unknown>): Promise<void>

          GetPluginStatus(): Promise<PluginStatus[]>
          GetTheme(): Promise<string>
          SetTheme(theme: string): Promise<void>
          GetSettings(): Promise<AppSettings>
          SaveSettings(settings: Record<string, unknown>): Promise<void>
        }
      }
    }
  }
}

// ===========================================================================
// 通用类型
// ===========================================================================

export interface LicenseStatus { valid: boolean; plan: string; expiresAt: number; daysRemaining: number; fingerprint: string; features: string[]; message: string }
export interface SystemInfo { hostname: string; platform: string; arch: string; goVersion: string; numCPU: number; numGoroutine: number; uptime: number }
export interface ProcessInfo { pid: number; name: string; cpu: number; mem: number }
export interface PluginStatus { name: string; running: boolean; version: string; uptime: string }
export interface AppSettings { theme: string; language: string; sidebarOpen: boolean; notifications: boolean }

// ===========================================================================
// AI 类型
// ===========================================================================

export type AIRole = 'user' | 'assistant' | 'system'
export interface AIContentPart { type: string; text?: string; image_url?: string; file_name?: string; file_type?: string; file_data?: number[] }
export interface AIChatMessage { id: string; role: AIRole; content: AIContentPart[]; created_at: string; is_error?: boolean }
export interface AIChatRequest { session_id: string; messages: AIChatMessage[]; model?: string; temperature?: number; max_tokens?: number }
export interface AIStreamEvent { type: string; message_id: string; delta?: string; error?: string }
export interface AIChatResponse { message_id: string; content: string; done: boolean }

// ===========================================================================
// Cloud 类型
// ===========================================================================

export interface CloudFileInfo { name: string; path: string; size: number; is_dir: boolean; modified_at: string; thumbnail_url?: string; mime_type?: string }
export interface CloudProgressEvent { type: string; context_id: string; path: string; transferred: number; total: number; speed: number; error?: string }
export interface CloudPreviewInfo { type: string; data?: number[]; url?: string; mime_type: string; extension: string; size: number }
export interface CloudProviderConfig { name: string; type: string; base_url: string; token: string; username: string; password: string; client_id: string; client_secret: string; redirect_uri: string }
export interface CloudOAuthToken { access_token: string; refresh_token: string; expires_at: string; provider: string }
export interface CloudMachineSyncReq { machine_id: string; remote_file: string; cloud_path: string; provider_name: string }

// ===========================================================================
// DTD (7DTD) 类型
// ===========================================================================

export interface DTDServerConfig { id: string; name: string; host: string; game_port: number; telnet_port: number; telnet_pass: string; rcon_port: number; rcon_pass: string; save_dir: string; log_path: string; blood_moon_day: number; installed_dir: string }
export interface DTDOnlinePlayer { id: string; name: string; platform_id?: string; status: string; joined_at: string; position?: DTDPosition }
export interface DTDPlayerEvent { timestamp: string; type: string; player_id: string; player_name: string }
export interface DTDConsoleMessage { timestamp: string; level: string; message: string }
export interface DTDBloodMoonInfo { current_day: number; next_blood_moon: number; remaining_days: number; remaining_hours: number }
export interface DTDScheduledTask { id: string; name: string; type: string; schedule: string; interval: number; enabled: boolean; payload: string; last_run_at?: string; next_run_at?: string }
export interface DTDPosition { id?: string; name: string; x: number; y: number; z: number }
export interface DTDMapData { spawn_points: DTDPosition[]; players: DTDPosition[]; traders: DTDPosition[]; claims: DTDPosition[]; world_size: number; world_name: string }
export interface DTDRemoteDeployReq { machine_id: string; install_dir: string; server_config: DTDServerConfig }
export interface DTDRemoteCmdResult { success: boolean; output: string; error?: string }

// ===========================================================================
// Git 类型
// ===========================================================================

export interface GitRepo { id: string; path: string; name: string; remote_url?: string }
export interface GitCloneRequest { url: string; local_path: string; username?: string; password?: string; ssh_key_path?: string }
export interface GitBranch { name: string; is_current: boolean; is_remote: boolean; hash: string; updated_at: string }
export interface GitCommit { hash: string; short_hash: string; author: string; email: string; message: string; date: string }
export interface GitStatusItem { file: string; status: string }

// ===========================================================================
// I18n 类型
// ===========================================================================

export interface I18nItem { key: string; original: string; translation: string; context?: string; line: number }
export interface I18nExtractReq { source_dir: string; extensions?: string[] }
export interface I18nExtractResult { total_files: number; total_items: number; files: Record<string, I18nItem[]> }

// ===========================================================================
// Browser 类型
// ===========================================================================

export interface BrowserInfo { user_agent: string; platform: string; is_embedded: boolean }

// ===========================================================================
// Plugin / Config 类型
// ===========================================================================

export interface PluginManifest { id: string; name: string; version: string; description: string; author: string; entry: string; type: string; permissions: string[]; config?: Record<string, string> }
export interface CloudSyncConfig { enabled: boolean; provider: string; endpoint: string; token: string; remote_path: string; interval: number }
export interface CloudSyncStatus { enabled: boolean; provider: string; last_sync: string }

// ===========================================================================
// Wails Runtime API
// ===========================================================================

const app = () => window.go?.main?.App
const checkWails = (): boolean => {
  if (!app()) { console.warn('[OmniPanel] Wails runtime 不可用 — 开发模式'); return false }
  return true
}

// ---------- License ----------
export async function getLicenseStatus(): Promise<LicenseStatus> {
  if (!checkWails()) return { valid: true, plan: 'pro', expiresAt: 0, daysRemaining: 365, fingerprint: 'mock', features: ['all'], message: 'dev mode' }
  return app().GetLicenseStatus()
}
export async function activateLicense(key: string, email: string): Promise<LicenseStatus> {
  if (!checkWails()) return { valid: true, plan: 'pro', expiresAt: 0, daysRemaining: 365, fingerprint: 'mock', features: ['all'], message: 'dev mode' }
  return app().ActivateLicense(key, email)
}

// ---------- System ----------
export async function getSystemInfo(): Promise<SystemInfo> {
  if (!checkWails()) return { hostname: 'dev', platform: 'linux', arch: 'amd64', goVersion: 'go1.22', numCPU: 16, numGoroutine: 42, uptime: 86400 }
  return app().GetSystemInfo()
}
export async function getProcessList(): Promise<ProcessInfo[]> {
  if (!checkWails()) return []
  return app().GetProcessList()
}

// ---------- AI ----------
export async function aiChat(req: AIChatRequest) { if (!checkWails()) return mockAiResp(); return app().AIChat(req) }
export async function aiChatStream(req: AIChatRequest) { if (!checkWails()) return; return app().AIChatStream(req) }
export async function aiUploadFile(name: string, data: number[]) { if (!checkWails()) return mockContentPart(name); return app().AIUploadFile(name, data) }
export async function aiGetHistory(sid: string) { if (!checkWails()) return []; return app().AIGetHistory(sid) }

// ---------- Cloud ----------
export const cloudListFiles = (p: string, d: string) => checkWails() ? app().CloudListFiles(p, d) : Promise.resolve([])
export const cloudCopyFile = (p: string, s: string, d: string) => checkWails() ? app().CloudCopyFile(p, s, d) : Promise.resolve()
export const cloudMoveFile = (p: string, s: string, d: string) => checkWails() ? app().CloudMoveFile(p, s, d) : Promise.resolve()
export const cloudRenameFile = (p: string, f: string, n: string) => checkWails() ? app().CloudRenameFile(p, f, n) : Promise.resolve()
export const cloudDeleteFile = (p: string, f: string) => checkWails() ? app().CloudDeleteFile(p, f) : Promise.resolve()
export const cloudMkdir = (p: string, d: string) => checkWails() ? app().CloudMkdir(p, d) : Promise.resolve()
export const cloudUploadFile = (p: string, l: string, r: string) => checkWails() ? app().CloudUploadFile(p, l, r) : Promise.resolve('mock-ctx')
export const cloudDownloadFile = (p: string, r: string, l: string) => checkWails() ? app().CloudDownloadFile(p, r, l) : Promise.resolve('mock-ctx')
export const cloudGetPreview = (p: string, f: string) => checkWails() ? app().CloudGetPreview(p, f) : Promise.resolve(mockPreview())
export const cloudOAuthURL = (p: string) => checkWails() ? app().CloudOAuthURL(p) : Promise.resolve('')
export const cloudHandleOAuthCallback = (p: string, c: string, s: string) => checkWails() ? app().CloudHandleOAuthCallback(p, c, s) : Promise.resolve(null)
export const cloudAddProvider = (cfg: CloudProviderConfig) => checkWails() ? app().CloudAddProvider(cfg) : Promise.resolve()
export const cloudListProviders = () => checkWails() ? app().CloudListProviders() : Promise.resolve([])
export const cloudSyncFromMachine = (req: CloudMachineSyncReq) => checkWails() ? app().CloudSyncFromMachine(req) : Promise.resolve('mock-ctx')

// ---------- DTD ----------
export const dtdAddServer = (cfg: DTDServerConfig) => checkWails() ? app().DTDAddServer(cfg) : Promise.resolve()
export const dtdListServers = () => checkWails() ? app().DTDListServers() : Promise.resolve([])
export const dtdConnect = (id: string) => checkWails() ? app().DTDConnect(id) : Promise.resolve()
export const dtdDisconnect = (id: string) => checkWails() ? app().DTDDisconnect(id) : Promise.resolve()
export const dtdSendCommand = (id: string, cmd: string) => checkWails() ? app().DTDSendCommand(id, cmd) : Promise.resolve()
export const dtdGetPlayers = (id: string) => checkWails() ? app().DTDGetPlayers(id) : Promise.resolve([])
export const dtdGetPlayerEvents = (id: string, limit: number) => checkWails() ? app().DTDGetPlayerEvents(id, limit) : Promise.resolve([])
export const dtdGetConsoleMessages = (id: string, limit: number) => checkWails() ? app().DTDGetConsoleMessages(id, limit) : Promise.resolve([])
export const dtdParseLogFile = (id: string, path: string) => checkWails() ? app().DTDParseLogFile(id, path) : Promise.resolve()
export const dtdGetMapData = (id: string) => checkWails() ? app().DTDGetMapData(id) : Promise.resolve(null)
export const dtdGetBloodMoonInfo = (id: string) => checkWails() ? app().DTDGetBloodMoonInfo(id) : Promise.resolve({ current_day: 1, next_blood_moon: 7, remaining_days: 6, remaining_hours: 144 })
export const dtdAddScheduledTask = (task: DTDScheduledTask) => checkWails() ? app().DTDAddScheduledTask(task) : Promise.resolve(task)
export const dtdRemoveScheduledTask = (id: string) => { if (checkWails()) app().DTDRemoveScheduledTask(id) }
export const dtdGetScheduledTasks = () => checkWails() ? app().DTDGetScheduledTasks() : Promise.resolve([])
export const dtdDeployToMachine = (req: DTDRemoteDeployReq) => checkWails() ? app().DTDDeployToMachine(req) : Promise.resolve({ success: false, output: '', error: 'dev mode' })
export const dtdStartRemoteServer = (mid: string, dir: string) => checkWails() ? app().DTDStartRemoteServer(mid, dir) : Promise.resolve({ success: false, output: '', error: 'dev mode' })

// ---------- Git ----------
export const gitAddLocalRepo = (path: string) => checkWails() ? app().GitAddLocalRepo(path) : Promise.resolve(mockGitRepo(path))
export const gitCloneRepo = (req: GitCloneRequest) => checkWails() ? app().GitCloneRepo(req) : Promise.resolve(mockGitRepo(req.local_path))
export const gitListRepos = () => checkWails() ? app().GitListRepos() : Promise.resolve([])
export const gitRemoveRepo = (id: string) => { if (checkWails()) app().GitRemoveRepo(id) }
export const gitBranches = (rid: string) => checkWails() ? app().GitBranches(rid) : Promise.resolve([])
export const gitLog = (rid: string, limit: number) => checkWails() ? app().GitLog(rid, limit) : Promise.resolve([])
export const gitStatus = (rid: string) => checkWails() ? app().GitStatus(rid) : Promise.resolve([])
export const gitPull = (rid: string) => checkWails() ? app().GitPull(rid) : Promise.resolve()
export const gitPush = (rid: string) => checkWails() ? app().GitPush(rid) : Promise.resolve()
export const gitCommit = (rid: string, msg: string) => checkWails() ? app().GitCommit(rid, msg) : Promise.resolve('mock-hash')
export const gitCheckout = (rid: string, branch: string) => checkWails() ? app().GitCheckout(rid, branch) : Promise.resolve()
export const gitCreateBranch = (rid: string, name: string) => checkWails() ? app().GitCreateBranch(rid, name) : Promise.resolve()
export const gitGetRepoStats = (rid: string) => checkWails() ? app().GitGetRepoStats(rid) : Promise.resolve({})

// ---------- I18n ----------
export const i18nExtract = (req: I18nExtractReq) => checkWails() ? app().I18nExtract(req) : Promise.resolve({ total_files: 0, total_items: 0, files: {} })
export const i18nGenerateLocaleFile = (items: I18nItem[], locale: string, path: string) => checkWails() ? app().I18nGenerateLocaleFile(items, locale, path) : Promise.resolve()
export const i18nPreviewTranslation = (file: string, items: I18nItem[]) => checkWails() ? app().I18nPreviewTranslation(file, items) : Promise.resolve('')
export const i18nApplyTranslationFile = (file: string, items: I18nItem[]) => checkWails() ? app().I18nApplyTranslationFile(file, items) : Promise.resolve()
export const i18nBatchApplyTranslation = (result: I18nExtractResult) => checkWails() ? app().I18nBatchApplyTranslation(result) : Promise.resolve()

// ---------- Browser ----------
export const browserGetInfo = () => checkWails() ? app().BrowserGetInfo() : Promise.resolve({ user_agent: 'OmniPanel/1.0', platform: 'linux', is_embedded: true })
export const browserOpenExternalURL = (url: string) => checkWails() ? app().BrowserOpenExternalURL(url) : Promise.resolve()
export const browserValidateURL = (url: string) => checkWails() ? app().BrowserValidateURL(url) : Promise.resolve()

// ---------- Plugin ----------
export const pluginList = () => checkWails() ? app().PluginList() : Promise.resolve([])
export const pluginExecute = (id: string, ctx: Record<string, unknown>) => checkWails() ? app().PluginExecute(id, ctx) : Promise.resolve({})
export const pluginEnable = (id: string) => { if (checkWails()) app().PluginEnable(id) }
export const pluginDisable = (id: string) => { if (checkWails()) app().PluginDisable(id) }

// ---------- Config Sync ----------
export const configGetSyncConfig = () => checkWails() ? app().ConfigGetSyncConfig() : Promise.resolve({ enabled: false, provider: '', last_sync: '' })
export const configSaveSyncConfig = (cfg: CloudSyncConfig) => checkWails() ? app().ConfigSaveSyncConfig(cfg) : Promise.resolve()
export const configPullFromCloud = () => checkWails() ? app().ConfigPullFromCloud() : Promise.resolve({})
export const configPushToCloud = () => checkWails() ? app().ConfigPushToCloud() : Promise.resolve()
export const configGetLocal = () => checkWails() ? app().ConfigGetLocal() : Promise.resolve({})
export const configSetLocal = (data: Record<string, unknown>) => checkWails() ? app().ConfigSetLocal(data) : Promise.resolve()

// ---------- Settings ----------
export async function getTheme() { if (!checkWails()) return 'dark'; return app().GetTheme() }
export async function setTheme(theme: string) { if (!checkWails()) return; return app().SetTheme(theme) }
export async function getSettings() { if (!checkWails()) return { theme: 'dark', language: 'zh-CN', sidebarOpen: true, notifications: true }; return app().GetSettings() }
export async function saveSettings(s: Record<string, unknown>) { if (!checkWails()) return; return app().SaveSettings(s) }
export async function getPluginStatus() { if (!checkWails()) return []; return app().GetPluginStatus() }

// ===========================================================================
// Mock helpers
// ===========================================================================
function mockAiResp(): AIChatResponse { return { message_id: 'mock', content: '开发模式 AI 回复', done: true } }
function mockContentPart(name: string): AIContentPart { return { type: 'file', file_name: name } }
function mockPreview(): CloudPreviewInfo { return { type: 'unsupported', mime_type: '', extension: '', size: 0 } }
function mockGitRepo(path: string): GitRepo { return { id: 'mock', path, name: path.split('/').pop() || 'repo' } }
