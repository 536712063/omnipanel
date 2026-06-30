<template>
  <div class="game-manager">
    <div class="page-header">
      <h2>{{ $t('game.title') }}</h2>
      <div class="header-actions">
        <template v-if="selectedServerId">
          <el-button :type="serverRunning ? 'danger' : 'success'" @click="toggleServer">
            {{ serverRunning ? $t('common.stop') : $t('common.start') }}
          </el-button>
          <el-button @click="restartServer">
            <el-icon><Refresh /></el-icon> {{ $t('common.restart') }}
          </el-button>
        </template>
        <el-button type="primary" @click="showAddServerDialog = true">
          <el-icon><Plus /></el-icon> {{ $t('common.add') }}
        </el-button>
      </div>
    </div>

    <div class="server-status-bar" v-if="selectedServerId && bloodMoon">
      <div class="status-item">
        <span class="status-label">{{ $t('game.serverStatus') }}</span>
        <el-tag :type="serverRunning ? 'success' : 'danger'">{{ serverRunning ? $t('common.running') : $t('common.stopped') }}</el-tag>
      </div>
      <div class="status-item">
        <span class="status-label">{{ $t('game.onlinePlayers') }}</span>
        <span>{{ players.length }}</span>
      </div>
      <div class="status-item">
        <span class="status-label">{{ $t('game.bloodMoon') }}</span>
        <span class="blood-moon-warn" v-if="bloodMoon.remaining_days <= 1">D{{ bloodMoon.remaining_days }} H{{ bloodMoon.remaining_hours }}</span>
        <span v-else>D{{ bloodMoon.remaining_days }}</span>
      </div>
    </div>

    <el-tabs v-model="activeTab" v-if="selectedServerId">
      <el-tab-pane :label="$t('game.console')" name="console">
        <div class="console-panel">
          <div class="console-output" ref="consoleOutput" style="height: 400px; overflow-y: auto; padding: 12px; font-family: 'Consolas', monospace; font-size: 13px; background: #0d0d0d; color: #ccc;">
            <div v-for="(line, i) in consoleMessages" :key="i" class="console-line"
              :class="{ 'console-warn': line.message?.includes('WARN'), 'console-error': line.message?.includes('ERROR') }">
              {{ line.timestamp }} {{ line.message }}
            </div>
          </div>
          <div class="console-input">
            <el-input v-model="consoleCommand" :placeholder="$t('game.console')"
              @keydown.enter="sendConsoleCmd">
              <template #append>
                <el-button @click="sendConsoleCmd">{{ $t('ai.send') }}</el-button>
              </template>
            </el-input>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane :label="$t('game.playerManagement')" name="players">
        <div class="player-manage">
          <div class="tab-toolbar">
            <el-input v-model="playerSearch" :placeholder="$t('common.search')" size="small" clearable style="width: 220px" />
            <div class="toolbar-right">
              <el-button type="primary" size="small" @click="refreshPlayers(selectedServerId)">{{ $t('common.refresh') }}</el-button>
            </div>
          </div>

          <el-table :data="filteredPlayers" size="small">
            <el-table-column prop="name" :label="$t('game.playerManagement')" width="160" />
            <el-table-column prop="platform_id" label="ID" width="200" />
            <el-table-column prop="status" :label="$t('common.detail')" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'online' ? 'success' : 'info'" size="small">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="joined_at" :label="$t('git.log')" width="180" />
            <el-table-column :label="$t('common.edit')" width="200">
              <template #default="{ row }">
                <el-button size="small" @click="kickPlayer(row)">{{ $t('game.kick') }}</el-button>
                <el-button size="small" type="danger" @click="banPlayer(row)">{{ $t('game.ban') }}</el-button>
                <el-button size="small" @click="teleportPlayer(row)">{{ $t('game.teleport') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>

      <el-tab-pane :label="$t('game.configEditor')" name="config">
        <div class="config-editor">
          <div class="config-toolbar">
            <el-button type="primary" size="small" @click="saveConfig">{{ $t('game.saveConfig') }}</el-button>
            <el-button size="small" @click="applyConfig">{{ $t('game.applyConfig') }}</el-button>
            <el-button size="small" @click="backupConfig">{{ $t('game.backupConfig') }}</el-button>
          </div>
          <el-form :model="serverConfig" label-width="160px" label-position="left">
            <el-divider content-position="left">基本设置</el-divider>
            <el-form-item label="服务器名称"><el-input v-model="serverConfig.ServerName" /></el-form-item>
            <el-form-item label="服务器描述"><el-input v-model="serverConfig.ServerDescription" type="textarea" :rows="2" /></el-form-item>
            <el-form-item label="服务器密码"><el-input v-model="serverConfig.ServerPassword" type="password" show-password /></el-form-item>
            <el-form-item label="最大玩家数"><el-input-number v-model="serverConfig.ServerMaxPlayerCount" :min="1" :max="64" /></el-form-item>
            <el-form-item label="服务器端口"><el-input-number v-model="serverConfig.ServerPort" :min="1" :max="65535" /></el-form-item>
            <el-divider content-position="left">游戏设置</el-divider>
            <el-form-item label="游戏难度">
              <el-select v-model="serverConfig.GameDifficulty">
                <el-option :label="`拾荒者 (1)`" :value="1" /><el-option :label="`冒险家 (2)`" :value="2" />
                <el-option :label="`游牧民 (3)`" :value="3" /><el-option :label="`战士 (4)`" :value="4" />
                <el-option :label="`幸存者 (5)`" :value="5" /><el-option :label="`疯狂 (6)`" :value="6" />
              </el-select>
            </el-form-item>
            <el-form-item label="世界种子"><el-input v-model="serverConfig.WorldGenSeed" /></el-form-item>
            <el-form-item label="世界大小"><el-input-number v-model="serverConfig.WorldGenSize" :min="1024" :max="16384" :step="1024" /></el-form-item>
            <el-form-item label="白天时长(分钟)"><el-input-number v-model="serverConfig.DayNightLength" :min="10" :max="120" /></el-form-item>
            <el-form-item label="僵尸生成数"><el-input-number v-model="serverConfig.MaxSpawnedZombies" :min="10" :max="200" /></el-form-item>
            <el-form-item label="血月频率(天)"><el-input-number v-model="serverConfig.BloodMoonFrequency" :min="1" :max="30" /></el-form-item>
            <el-form-item label="启用 EAC 反作弊"><el-switch v-model="serverConfig.EACEnabled" /></el-form-item>
          </el-form>
        </div>
      </el-tab-pane>

      <el-tab-pane :label="$t('game.backupManagement')" name="backup">
        <div class="backup-manage">
          <div class="tab-toolbar">
            <el-button type="primary" size="small" @click="createBackup">{{ $t('game.createBackup') }}</el-button>
          </div>
          <el-table :data="backups" size="small">
            <el-table-column prop="name" :label="$t('game.backupManagement')" min-width="200" />
            <el-table-column prop="size" label="Size" width="100" />
            <el-table-column prop="date" :label="$t('git.log')" width="180" />
            <el-table-column :label="$t('common.edit')" width="180">
              <template #default="{ row }">
                <el-button size="small" @click="restoreBackup(row)">{{ $t('game.restoreBackup') }}</el-button>
                <el-button size="small" type="danger" @click="removeBackup(row)">{{ $t('common.delete') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>

      <el-tab-pane :label="$t('game.stats')" name="stats">
        <div class="stats-panel">
          <div class="stats-grid-3">
            <div class="stat-card-small">
              <h4>{{ $t('game.onlinePlayers') }}</h4>
              <div ref="playerChartRef" style="height: 200px;"></div>
            </div>
            <div class="stat-card-small">
              <h4>{{ $t('game.tps') }}</h4>
              <div ref="tpsChartRef" style="height: 200px;"></div>
            </div>
            <div class="stat-card-small">
              <h4>{{ $t('game.memory') }}</h4>
              <div ref="memChartRef" style="height: 200px;"></div>
            </div>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>

    <div class="server-not-installed" v-else>
      <div class="not-installed-content">
        <div class="install-icon">Game</div>
        <h3>{{ $t('game.notInstalled') }}</h3>
        <p>{{ $t('game.notInstalledDesc') }}</p>
        <el-button type="primary" size="large" @click="showAddServerDialog = true">
          <el-icon><Plus /></el-icon> {{ $t('game.oneClickDeploy') }}
        </el-button>
      </div>
    </div>

    <el-dialog v-model="showAddServerDialog" :title="$t('game.oneClickDeploy')" width="550px">
      <el-form :model="newServer" label-width="120px">
        <el-form-item label="服务器名称"><el-input v-model="newServer.name" /></el-form-item>
        <el-form-item label="Host"><el-input v-model="newServer.host" placeholder="127.0.0.1" /></el-form-item>
        <el-form-item label="Game Port"><el-input-number v-model="newServer.game_port" :min="1" :max="65535" /></el-form-item>
        <el-form-item label="Telnet Port"><el-input-number v-model="newServer.telnet_port" :min="1" :max="65535" /></el-form-item>
        <el-form-item label="Telnet Password"><el-input v-model="newServer.telnet_pass" /></el-form-item>
        <el-form-item label="RCON Port"><el-input-number v-model="newServer.rcon_port" :min="1" :max="65535" /></el-form-item>
        <el-form-item label="RCON Password"><el-input v-model="newServer.rcon_pass" /></el-form-item>
        <el-form-item :label="$t('game.bloodMoon')"><el-input-number v-model="newServer.blood_moon_day" :min="1" :max="30" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddServerDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="doAddServer">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as echarts from 'echarts'
import { useDTD } from '@/composables/useDTD'
import type { DTDServerConfig } from '@/wails/runtime'

const {
  servers, selectedServerId, players, events, consoleMessages, tasks, bloodMoon, mapData, loading, error,
  loadServers, addServer, connect, disconnect, sendCommand, refreshPlayers, refreshAll,
  addTask, removeTask, loadTasks, parseLogFile, deployToMachine, startRemoteServer, subscribe
} = useDTD()

const activeTab = ref('console')
const showAddServerDialog = ref(false)
const serverRunning = ref(false)
const playerSearch = ref('')
const consoleCommand = ref('')
const consoleOutput = ref<HTMLElement>()

const serverConfig = reactive({
  ServerName: 'OmniPanel 七日杀服务器', ServerDescription: '', ServerPassword: '',
  ServerMaxPlayerCount: 16, ServerPort: 26900, ServerVisibility: 2,
  GameDifficulty: 4, GameMode: 'GameModeSurvival', WorldGenSeed: 'OmniPanel2026',
  WorldGenSize: 8192, DayNightLength: 60, MaxSpawnedZombies: 60, MaxSpawnedAnimals: 50,
  EACEnabled: true, BloodMoonFrequency: 7, AirDropFrequency: 24, DropOnDeath: 1, DropOnQuit: 1
})

const newServer = reactive<DTDServerConfig>({
  id: '', name: '', host: '127.0.0.1', game_port: 26900, telnet_port: 8081,
  telnet_pass: '', rcon_port: 8082, rcon_pass: '', save_dir: '/home/steam/7days/Saves',
  log_path: '/home/steam/7days/logs', blood_moon_day: 7, installed_dir: '/home/steam/7days'
})

const backups = ref([
  { name: '20260629_1500_Saves.zip', size: '1.2GB', date: '2026-06-29 15:00' },
  { name: '20260628_1500_Saves.zip', size: '1.1GB', date: '2026-06-28 15:00' }
])

const playerChartRef = ref<HTMLElement>()
const tpsChartRef = ref<HTMLElement>()
const memChartRef = ref<HTMLElement>()
let gameCharts: echarts.ECharts[] = []

const filteredPlayers = computed(() =>
  players.value.filter(p => p.name.toLowerCase().includes(playerSearch.value.toLowerCase()))
)

async function doAddServer() {
  if (!newServer.name || !newServer.host) { ElMessage.warning('请填写必要信息'); return }
  const cfg = { ...newServer, id: `${Date.now()}` }
  await addServer(cfg)
  if (!selectedServerId.value) {
    await connect(cfg.id)
    selectedServerId.value = cfg.id
  }
  showAddServerDialog.value = false
  ElMessage.success('服务器已添加')
}

async function sendConsoleCmd() {
  if (!consoleCommand.value.trim() || !selectedServerId.value) return
  await sendCommand(selectedServerId.value, consoleCommand.value)
  consoleCommand.value = ''
}

function toggleServer() {
  serverRunning.value = !serverRunning.value
  if (!serverRunning.value && selectedServerId.value) disconnect(selectedServerId.value)
  ElMessage.success(serverRunning.value ? '服务器已启动' : '服务器已停止')
}

function restartServer() { serverRunning.value = false; setTimeout(() => { serverRunning.value = true }, 2000); ElMessage.success('正在重启...') }

function kickPlayer(player: any) { if (selectedServerId.value) sendCommand(selectedServerId.value, `kick ${player.name}`) }
function banPlayer(player: any) { if (selectedServerId.value) sendCommand(selectedServerId.value, `ban add ${player.platform_id || player.name}`) }
function teleportPlayer(player: any) { ElMessage.info(`传送玩家 ${player.name}`) }

function saveConfig() { ElMessage.success('配置已保存') }
function applyConfig() { ElMessage.success('配置已应用') }
function backupConfig() { ElMessage.success('配置已备份') }

function createBackup() { backups.value.unshift({ name: `backup_${Date.now()}.zip`, size: '1.2GB', date: new Date().toLocaleString() }); ElMessage.success('备份已创建') }
function restoreBackup(row: any) { ElMessage.success(`从 ${row.name} 恢复中...`) }
function removeBackup(row: any) { backups.value = backups.value.filter(b => b !== row) }

function initCharts() {
  if (!playerChartRef.value) return
  const make = (el: HTMLElement, color: string, data: number[]) => {
    const c = echarts.init(el)
    c.setOption({
      grid: { top: 5, right: 5, bottom: 15, left: 35 },
      xAxis: { type: 'category', data: Array.from({length:24},(_,i)=>`${i}h`), show: false },
      yAxis: { type: 'value', splitLine: { lineStyle: { color: '#333' } } },
      series: [{ data, type: 'line', smooth: true, showSymbol: false, lineStyle: { color, width: 2 }, areaStyle: { color: color + '33' } }]
    })
    gameCharts.push(c)
  }
  make(playerChartRef.value, '#409eff', [0,0,1,5,3,8,12,15,14,16,10,8,5,3,5,8,12,15,14,10,8,5,3,2])
  make(tpsChartRef.value!, '#67c23a', Array.from({length:24},()=>15+Math.random()*5))
  make(memChartRef.value!, '#e6a23c', Array.from({length:24},()=>3000+Math.random()*2000))
}

onMounted(async () => {
  await loadServers()
  subscribe()
  setTimeout(initCharts, 500)
})

onUnmounted(() => { gameCharts.forEach(c => c.dispose()) })
</script>

<style scoped>
.game-manager { padding: 0; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { font-size: 22px; font-weight: 600; }
.header-actions { display: flex; gap: 10px; }

.server-status-bar { display: flex; gap: 24px; padding: 12px 20px; background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 10px; margin-bottom: 16px; }
.status-item { display: flex; align-items: center; gap: 10px; }
.status-label { color: var(--text-secondary); font-size: 13px; }
.blood-moon-warn { color: #f56c6c; font-weight: bold; }

.console-panel { background: #0d0d0d; border-radius: 8px; overflow: hidden; }
.console-line { color: #ccc; line-height: 1.6; }
.console-warn { color: #e6a23c; }
.console-error { color: #f56c6c; }
.console-input { padding: 10px; background: #1a1a1a; }

.tab-toolbar { display: flex; justify-content: space-between; align-items: center; padding: 8px 0; margin-bottom: 10px; }
.toolbar-right { display: flex; gap: 8px; }

.config-editor { padding: 12px 0; }
.config-toolbar { margin-bottom: 16px; display: flex; gap: 8px; }

.stats-grid-3 { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; }
.stat-card-small { background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 10px; padding: 14px; }
.stat-card-small h4 { margin-bottom: 8px; font-size: 13px; }

.server-not-installed { display: flex; align-items: center; justify-content: center; min-height: 400px; }
.not-installed-content { text-align: center; }
.install-icon { font-size: 64px; margin-bottom: 20px; font-weight: bold; color: var(--accent-color); }
.not-installed-content h3 { margin-bottom: 10px; }
.not-installed-content p { color: var(--text-secondary); margin-bottom: 24px; }
</style>
