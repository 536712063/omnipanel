import { ref, reactive, onUnmounted } from 'vue'
import type { DTDServerConfig, DTDOnlinePlayer, DTDPlayerEvent, DTDConsoleMessage, DTDScheduledTask, DTDBloodMoonInfo, DTDMapData, DTDRemoteDeployReq, DTDRemoteCmdResult } from '@/wails/runtime'
import * as api from '@/wails/runtime'

export function useDTD() {
  const servers = ref<DTDServerConfig[]>([])
  const selectedServerId = ref('')
  const players = ref<DTDOnlinePlayer[]>([])
  const events = ref<DTDPlayerEvent[]>([])
  const consoleMessages = ref<DTDConsoleMessage[]>([])
  const tasks = ref<DTDScheduledTask[]>([])
  const bloodMoon = ref<DTDBloodMoonInfo | null>(null)
  const mapData = ref<DTDMapData | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function loadServers() {
    servers.value = await api.dtdListServers()
  }

  async function addServer(cfg: DTDServerConfig) {
    await api.dtdAddServer(cfg)
    await loadServers()
  }

  async function connect(serverId: string) {
    loading.value = true
    try {
      await api.dtdConnect(serverId)
      selectedServerId.value = serverId
      await refreshAll()
    } catch (e: any) {
      error.value = e?.message || 'Failed to connect'
    } finally {
      loading.value = false
    }
  }

  async function disconnect(serverId: string) {
    await api.dtdDisconnect(serverId)
    if (selectedServerId.value === serverId) selectedServerId.value = ''
  }

  async function sendCommand(serverId: string, command: string) {
    await api.dtdSendCommand(serverId, command)
  }

  async function refreshPlayers(serverId: string) {
    try { await api.dtdSendCommand(serverId, 'lp') } catch {}
    try { players.value = await api.dtdGetPlayers(serverId) } catch {}
  }

  async function refreshAll() {
    const sid = selectedServerId.value
    if (!sid) return
    try {
      const [p, e, c, b, m] = await Promise.all([
        api.dtdGetPlayers(sid),
        api.dtdGetPlayerEvents(sid, 50),
        api.dtdGetConsoleMessages(sid, 200),
        api.dtdGetBloodMoonInfo(sid),
        api.dtdGetMapData(sid),
      ])
      players.value = p
      events.value = e
      consoleMessages.value = c
      bloodMoon.value = b
      mapData.value = m
    } catch (e: any) {
      error.value = e?.message
    }
  }

  async function addTask(task: Omit<DTDScheduledTask, 'id' | 'last_run_at' | 'next_run_at'>) {
    await api.dtdAddScheduledTask(task as DTDScheduledTask)
    await loadTasks()
  }

  async function removeTask(taskId: string) {
    api.dtdRemoveScheduledTask(taskId)
    await loadTasks()
  }

  async function loadTasks() {
    tasks.value = await api.dtdGetScheduledTasks()
  }

  async function parseLogFile(serverId: string) {
    const cfg = servers.value.find(s => s.id === serverId)
    if (cfg?.log_path) await api.dtdParseLogFile(serverId, cfg.log_path)
  }

  async function deployToMachine(req: DTDRemoteDeployReq): Promise<DTDRemoteCmdResult> {
    return api.dtdDeployToMachine(req)
  }

  async function startRemoteServer(machineId: string, dir: string): Promise<DTDRemoteCmdResult> {
    return api.dtdStartRemoteServer(machineId, dir)
  }

  let unsubConsole: (() => void) | null = null
  function subscribe() {
    const handler = (data: any) => {
      if (data?.server_id === selectedServerId.value && data?.line) {
        consoleMessages.value.push({ timestamp: new Date().toISOString(), level: 'INFO', message: data.line })
        if (consoleMessages.value.length > 500) consoleMessages.value = consoleMessages.value.slice(-500)
      }
    }
    ;(window as any)?.EventsOn?.('dtd:console:message', handler)
    unsubConsole = () => { (window as any)?.EventsOff?.('dtd:console:message', handler) }
  }

  onUnmounted(() => { if (unsubConsole) unsubConsole() })

  return { servers, selectedServerId, players, events, consoleMessages, tasks, bloodMoon, mapData, loading, error,
    loadServers, addServer, connect, disconnect, sendCommand, refreshPlayers, refreshAll,
    addTask, removeTask, loadTasks, parseLogFile, deployToMachine, startRemoteServer, subscribe }
}
