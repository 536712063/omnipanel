/**
 * Wails v3 运行时绑定层
 *
 * 该模块封装了所有从 Vue 3 前端调用 Go Agent 后端的方法。
 * 在 Wails v3 中, Go 导出的绑定方法通过 `window.go.main.App.<MethodName>` 调用。
 *
 * 架构:
 *   Vue Component → wails/runtime.ts → Wails IPC → Go app.go → gRPC → Plugin Process
 *
 * 错误处理: 所有方法均返回 Promise, 前端通过 try/catch 处理异常。
 */

// Wails v3 全局类型声明
declare global {
  interface Window {
    go: {
      main: {
        App: {
          // License
          GetLicenseStatus(): Promise<LicenseStatus>
          ActivateLicense(licenseKey: string, email: string): Promise<LicenseStatus>

          // 系统
          GetSystemInfo(): Promise<SystemInfo>
          GetProcessList(): Promise<ProcessInfo[]>

          // SDTD (七日杀)
          SDTDInstallServer(installDir: string, steamUser: string, steamPass: string): Promise<string>
          SDTDStartServer(): Promise<StatusResponse>
          SDTDStopServer(): Promise<StatusResponse>
          SDTDGetServerConfig(): Promise<SDTServerConfig>
          SDTDSaveServerConfig(config: Record<string, unknown>): Promise<StatusResponse>
          SDTDSendConsoleCommand(command: string): Promise<string>
          SDTDGetPlayers(): Promise<SDTPlayer[]>
          SDTDCreateBackup(): Promise<BackupInfo>

          // 插件管理
          GetPluginStatus(): Promise<PluginStatus[]>

          // 设置
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
// 类型定义
// ===========================================================================

export interface LicenseStatus {
  valid: boolean
  plan: string
  expiresAt: number
  daysRemaining: number
  fingerprint: string
  features: string[]
  message: string
}

export interface SystemInfo {
  hostname: string
  platform: string
  arch: string
  goVersion: string
  numCPU: number
  numGoroutine: number
  uptime: number
}

export interface ProcessInfo {
  pid: number
  name: string
  cpu: number
  mem: number
}

export interface StatusResponse {
  success: boolean
  message: string
}

export interface SDTServerConfig {
  serverName: string
  serverPort: number
  maxPlayers: number
  gameDifficulty: number
  worldGenSeed: string
  eacEnabled: boolean
  bloodMoonFrequency: number
  serverDescription: string
  serverPassword: string
  maxSpawnedZombies: number
  maxSpawnedAnimals: number
  dropOnDeath: number
  gameMode: string
}

export interface SDTPlayer {
  name: string
  steamId: string
  level: number
  online: boolean
  playTime?: string
  score?: number
}

export interface BackupInfo {
  id: string
  name: string
  size: string
}

export interface PluginStatus {
  name: string
  running: boolean
  version: string
  uptime: string
}

export interface AppSettings {
  theme: string
  language: string
  sidebarOpen: boolean
  notifications: boolean
}

// ===========================================================================
// Wails Runtime API 封装
// ===========================================================================

const app = () => window.go?.main?.App

function checkWails(): boolean {
  if (!app()) {
    console.warn('[OmniPanel] Wails runtime 不可用 — 请确保在 Wails 环境中运行')
    return false
  }
  return true
}

// ---------- License ----------

export async function getLicenseStatus(): Promise<LicenseStatus> {
  if (!checkWails()) return mockLicenseStatus()
  return app().GetLicenseStatus()
}

export async function activateLicense(key: string, email: string): Promise<LicenseStatus> {
  if (!checkWails()) return { valid: true, plan: 'pro', expiresAt: 0, daysRemaining: 365, fingerprint: 'mock', features: ['all'], message: 'dev mode' }
  return app().ActivateLicense(key, email)
}

// ---------- 系统 ----------

export async function getSystemInfo(): Promise<SystemInfo> {
  if (!checkWails()) return mockSystemInfo()
  return app().GetSystemInfo()
}

export async function getProcessList(): Promise<ProcessInfo[]> {
  if (!checkWails()) return []
  return app().GetProcessList()
}

// ---------- SDTD (七日杀) ----------

export async function sdtInstallServer(
  installDir: string,
  steamUser: string,
  steamPass: string
): Promise<string> {
  if (!checkWails()) return 'mock-task-id'
  return app().SDTDInstallServer(installDir, steamUser, steamPass)
}

export async function sdtStartServer(): Promise<StatusResponse> {
  if (!checkWails()) return { success: true, message: '模拟启动' }
  return app().SDTDStartServer()
}

export async function sdtStopServer(): Promise<StatusResponse> {
  if (!checkWails()) return { success: true, message: '模拟停止' }
  return app().SDTDStopServer()
}

export async function sdtGetServerConfig(): Promise<SDTServerConfig> {
  if (!checkWails()) return mockSDTConfig()
  return app().SDTDGetServerConfig()
}

export async function sdtSaveServerConfig(config: Record<string, unknown>): Promise<StatusResponse> {
  if (!checkWails()) return { success: true, message: '配置已保存 (开发模式)' }
  return app().SDTDSaveServerConfig(config)
}

export async function sdtSendCommand(command: string): Promise<string> {
  if (!checkWails()) return `[模拟] 命令 "${command}" 已发送`
  return app().SDTDSendConsoleCommand(command)
}

export async function sdtGetPlayers(): Promise<SDTPlayer[]> {
  if (!checkWails()) return mockPlayers()
  return app().SDTDGetPlayers()
}

export async function sdtCreateBackup(): Promise<BackupInfo> {
  if (!checkWails()) return { id: 'mock', name: 'mock_saves.zip', size: '0B' }
  return app().SDTDCreateBackup()
}

// ---------- 插件管理 ----------

export async function getPluginStatus(): Promise<PluginStatus[]> {
  if (!checkWails()) return mockPluginStatus()
  return app().GetPluginStatus()
}

// ---------- 设置 ----------

export async function getTheme(): Promise<string> {
  if (!checkWails()) return 'dark'
  return app().GetTheme()
}

export async function setTheme(theme: string): Promise<void> {
  if (!checkWails()) return
  return app().SetTheme(theme)
}

export async function getSettings(): Promise<AppSettings> {
  if (!checkWails()) return { theme: 'dark', language: 'zh-CN', sidebarOpen: true, notifications: true }
  return app().GetSettings()
}

export async function saveSettings(settings: Record<string, unknown>): Promise<void> {
  if (!checkWails()) return
  return app().SaveSettings(settings)
}

// ===========================================================================
// WebSocket 实时流 (用于七日杀控制台 / Docker 日志 / 终端输出)
// ===========================================================================

export class RealtimeStream {
  private ws: WebSocket | null = null
  private url: string
  private onMessage: (data: unknown) => void
  private onError: (err: Event) => void
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 10

  constructor(
    url: string,
    onMessage: (data: unknown) => void,
    onError?: (err: Event) => void
  ) {
    this.url = url
    this.onMessage = onMessage
    this.onError = onError || (() => {})
  }

  connect(): void {
    try {
      // 在 Wails 环境中, WebSocket 连接到本地 Agent
      const wsUrl = `ws://localhost:3001/ws?session=${encodeURIComponent(this.url)}`
      this.ws = new WebSocket(wsUrl)

      this.ws.onopen = () => {
        console.log('[OmniPanel] WebSocket 已连接:', this.url)
        this.reconnectAttempts = 0
      }

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          this.onMessage(data)
        } catch {
          this.onMessage(event.data)
        }
      }

      this.ws.onerror = (err) => {
        console.error('[OmniPanel] WebSocket 错误:', err)
        this.onError(err)
      }

      this.ws.onclose = () => {
        console.log('[OmniPanel] WebSocket 已断开, 尝试重连...')
        this.scheduleReconnect()
      }
    } catch (err) {
      console.error('[OmniPanel] WebSocket 连接失败:', err)
      this.scheduleReconnect()
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('[OmniPanel] WebSocket 重连次数已达上限, 放弃重连')
      return
    }
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000)
    this.reconnectAttempts++
    this.reconnectTimer = setTimeout(() => this.connect(), delay)
  }

  send(data: unknown): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data))
    }
  }

  disconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
    }
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }
}

/**
 * 创建 SDTD 控制台实时日志流
 */
export function createSDTConsoleStream(
  onLine: (line: { timestamp: string; level: string; message: string }) => void
): RealtimeStream {
  return new RealtimeStream('sdt-console', (data) => {
    if (typeof data === 'object' && data !== null) {
      onLine(data as { timestamp: string; level: string; message: string })
    }
  })
}

/**
 * 创建 SDTD 监控数据流 (TPS/FPS/内存)
 */
export function createSDTMetricsStream(
  onMetrics: (metrics: { tps: number; fps: number; memoryMb: number; onlinePlayers: number }) => void
): RealtimeStream {
  return new RealtimeStream('sdt-metrics', (data) => {
    if (typeof data === 'object' && data !== null) {
      onMetrics(data as { tps: number; fps: number; memoryMb: number; onlinePlayers: number })
    }
  })
}

// ===========================================================================
// 开发模式 Mock 数据 (当 Wails Runtime 不可用时)
// ===========================================================================

function mockLicenseStatus(): LicenseStatus {
  return {
    valid: true, plan: 'pro', expiresAt: Date.now() / 1000 + 86400 * 365,
    daysRemaining: 365, fingerprint: 'dev-fingerprint',
    features: ['all'], message: '开发模式 - 所有功能已解锁'
  }
}

function mockSystemInfo(): SystemInfo {
  return {
    hostname: 'dev-machine', platform: 'linux', arch: 'amd64',
    goVersion: 'go1.22', numCPU: 16, numGoroutine: 42, uptime: 86400
  }
}

function mockSDTConfig(): SDTServerConfig {
  return {
    serverName: 'OmniPanel 七日杀开发服',
    serverPort: 26900, maxPlayers: 16, gameDifficulty: 4,
    worldGenSeed: 'DevSeed2026', eacEnabled: true,
    bloodMoonFrequency: 7, serverDescription: '开发测试服务器',
    serverPassword: '', maxSpawnedZombies: 60, maxSpawnedAnimals: 50,
    dropOnDeath: 1, gameMode: 'GameModeSurvival'
  }
}

function mockPlayers(): SDTPlayer[] {
  return [
    { name: 'DevPlayer1', steamId: '76561198000000001', level: 50, online: true },
    { name: 'DevPlayer2', steamId: '76561198000000002', level: 30, online: true },
    { name: 'DevPlayer3', steamId: '76561198000000003', level: 75, online: false }
  ]
}

function mockPluginStatus(): PluginStatus[] {
  return [
    { name: 'docker', running: true, version: '1.0.0', uptime: '2h30m' },
    { name: 'ssh', running: true, version: '1.0.0', uptime: '2h30m' },
    { name: 'frp', running: true, version: '1.0.0', uptime: '2h30m' },
    { name: 'sdt', running: true, version: '1.0.0', uptime: '2h29m' }
  ]
}
