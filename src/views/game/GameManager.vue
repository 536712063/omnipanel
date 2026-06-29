<template>
  <div class="game-manager">
    <div class="page-header">
      <h2>🎮 七日杀 V3.0 服务端管理面板</h2>
      <div class="header-actions">
        <el-button type="primary" @click="installServer" v-if="!serverInstalled">
          <el-icon><Download /></el-icon> 一键部署服务端
        </el-button>
        <template v-else>
          <el-button :type="serverRunning ? 'danger' : 'success'" @click="toggleServer">
            {{ serverRunning ? '停止' : '启动' }}
          </el-button>
          <el-button @click="restartServer">
            <el-icon><Refresh /></el-icon> 重启
          </el-button>
        </template>
      </div>
    </div>

    <div class="server-status-bar" v-if="serverInstalled">
      <div class="status-item">
        <span class="status-label">服务状态</span>
        <el-tag :type="serverRunning ? 'success' : 'danger'">{{ serverRunning ? '运行中' : '已停止' }}</el-tag>
      </div>
      <div class="status-item">
        <span class="status-label">在线玩家</span>
        <span>{{ onlinePlayers }}/{{ maxPlayers }}</span>
      </div>
      <div class="status-item">
        <span class="status-label">TPS</span>
        <span :class="tps > 15 ? 'status-online' : 'status-warning'">{{ tps.toFixed(1) }}</span>
      </div>
      <div class="status-item">
        <span class="status-label">FPS</span>
        <span>{{ fps }}</span>
      </div>
      <div class="status-item">
        <span class="status-label">内存</span>
        <span>{{ serverMemory }}MB</span>
      </div>
    </div>

    <el-tabs v-model="activeTab" v-if="serverInstalled">
      <el-tab-pane label="控制台" name="console">
        <div class="console-panel">
          <div class="console-output" ref="consoleOutput">
            <div v-for="(line, i) in consoleLogs" :key="i" class="console-line"
                 :class="{ 'console-warn': line.includes('WARN'), 'console-error': line.includes('ERROR') }">
              {{ line }}
            </div>
          </div>
          <div class="console-input">
            <el-input v-model="consoleCommand" placeholder="输入控制台命令... (例: help, listplayers, kick <name>)"
                      @keydown.enter="sendConsoleCommand">
              <template #append>
                <el-button @click="sendConsoleCommand">发送</el-button>
              </template>
            </el-input>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="玩家管理" name="players">
        <div class="player-manage">
          <div class="tab-toolbar">
            <el-input v-model="playerSearch" placeholder="搜索玩家..." size="small" clearable style="width: 220px" />
            <div class="toolbar-right">
              <el-button type="primary" size="small" @click="refreshPlayers">刷新</el-button>
            </div>
          </div>

          <el-table :data="filteredPlayers" size="small">
            <el-table-column prop="name" label="玩家名称" width="150" />
            <el-table-column prop="steamId" label="Steam ID" width="180">
              <template #default="{ row }">
                <span style="font-family: monospace; font-size: 12px;">{{ row.steamId }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="playTime" label="游戏时长" width="100" />
            <el-table-column prop="level" label="等级" width="70" />
            <el-table-column prop="score" label="积分" width="80" />
            <el-table-column prop="status" label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.online ? 'success' : 'info'" size="small">{{ row.online ? '在线' : '离线' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="240" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click="kickPlayer(row)">踢出</el-button>
                <el-button size="small" :type="row.banned ? 'success' : 'danger'" @click="toggleBan(row)">
                  {{ row.banned ? '解封' : '封禁' }}
                </el-button>
                <el-button size="small" @click="teleportPlayer(row)">传送</el-button>
                <el-button size="small" @click="giveItem(row)">给予物品</el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="sub-section">
            <h4>白名单管理</h4>
            <div class="whitelist-controls">
              <el-input v-model="newWhitelist" placeholder="输入 Steam ID" size="small" style="width: 250px" />
              <el-button size="small" type="primary" @click="addWhitelist">添加</el-button>
              <el-switch v-model="whitelistEnabled" active-text="白名单已启用" style="margin-left: 20px" />
            </div>
            <div class="whitelist-tags" style="margin-top: 10px;">
              <el-tag v-for="(id, i) in whitelist" :key="i" closable @close="removeWhitelist(i)" size="small" style="margin-right: 6px; margin-bottom: 6px;">
                {{ id }}
              </el-tag>
            </div>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="配置编辑器" name="config">
        <div class="config-editor">
          <div class="config-toolbar">
            <el-button type="primary" size="small" @click="saveConfig">保存配置</el-button>
            <el-button size="small" @click="applyConfig">应用配置</el-button>
            <el-button size="small" @click="backupConfig">备份配置</el-button>
          </div>

          <el-scrollbar height="calc(100vh - 300px)">
            <el-form :model="serverConfig" label-width="180px" label-position="left">
              <el-divider content-position="left">基本设置</el-divider>
              <el-form-item label="服务器名称">
                <el-input v-model="serverConfig.ServerName" />
              </el-form-item>
              <el-form-item label="服务器描述">
                <el-input v-model="serverConfig.ServerDescription" type="textarea" :rows="2" />
              </el-form-item>
              <el-form-item label="服务器密码">
                <el-input v-model="serverConfig.ServerPassword" type="password" show-password />
              </el-form-item>
              <el-form-item label="最大玩家数">
                <el-input-number v-model="serverConfig.ServerMaxPlayerCount" :min="1" :max="64" />
              </el-form-item>
              <el-form-item label="服务器端口">
                <el-input-number v-model="serverConfig.ServerPort" :min="1" :max="65535" />
              </el-form-item>
              <el-form-item label="服务器可见性">
                <el-select v-model="serverConfig.ServerVisibility">
                  <el-option label="不公开" :value="0" />
                  <el-option label="仅好友" :value="1" />
                  <el-option label="公开" :value="2" />
                </el-select>
              </el-form-item>

              <el-divider content-position="left">游戏设置</el-divider>
              <el-form-item label="游戏难度">
                <el-select v-model="serverConfig.GameDifficulty">
                  <el-option label="拾荒者 (1)" :value="1" />
                  <el-option label="冒险家 (2)" :value="2" />
                  <el-option label="游牧民 (3)" :value="3" />
                  <el-option label="战士 (4)" :value="4" />
                  <el-option label="幸存者 (5)" :value="5" />
                  <el-option label="疯狂 (6)" :value="6" />
                </el-select>
              </el-form-item>
              <el-form-item label="游戏模式">
                <el-select v-model="serverConfig.GameMode">
                  <el-option label="生存模式" value="GameModeSurvival" />
                  <el-option label="创造模式" value="GameModeCreative" />
                </el-select>
              </el-form-item>
              <el-form-item label="世界种子">
                <el-input v-model="serverConfig.WorldGenSeed" />
              </el-form-item>
              <el-form-item label="世界大小">
                <el-input-number v-model="serverConfig.WorldGenSize" :min="1024" :max="16384" :step="1024" />
              </el-form-item>
              <el-form-item label="白天时长(分钟)">
                <el-input-number v-model="serverConfig.DayNightLength" :min="10" :max="120" />
              </el-form-item>
              <el-form-item label="僵尸生成数">
                <el-input-number v-model="serverConfig.MaxSpawnedZombies" :min="10" :max="200" />
              </el-form-item>
              <el-form-item label="动物生成数">
                <el-input-number v-model="serverConfig.MaxSpawnedAnimals" :min="10" :max="100" />
              </el-form-item>
              <el-form-item label="启用 EAC 反作弊">
                <el-switch v-model="serverConfig.EACEnabled" />
              </el-form-item>
              <el-form-item label="血月频率(天)">
                <el-input-number v-model="serverConfig.BloodMoonFrequency" :min="1" :max="30" />
              </el-form-item>
              <el-form-item label="空投频率(小时)">
                <el-input-number v-model="serverConfig.AirDropFrequency" :min="1" :max="168" />
              </el-form-item>
              <el-form-item label="掉落模式">
                <el-select v-model="serverConfig.DropOnDeath">
                  <el-option label="全部掉落 (Everything)" :value="3" />
                  <el-option label="仅背包 (Backpack)" :value="2" />
                  <el-option label="仅工具带 (Toolbelt)" :value="1" />
                  <el-option label="不删除 (Delete)" :value="0" />
                </el-select>
              </el-form-item>
              <el-form-item label="掉落保留">
                <el-select v-model="serverConfig.DropOnQuit">
                  <el-option label="不保留" :value="0" />
                  <el-option label="保留物品" :value="1" />
                  <el-option label="保留全部" :value="2" />
                </el-select>
              </el-form-item>

              <el-divider content-position="left">管理员设置</el-divider>
              <div class="admin-list">
                <div class="admin-item" v-for="(admin, i) in admins" :key="i">
                  <el-input v-model="admin.steamId" placeholder="Steam ID" size="small" style="width: 200px" />
                  <el-select v-model="admin.level" size="small" style="width: 120px; margin-left: 10px">
                    <el-option label="管理员 (0)" :value="0" />
                    <el-option label="内部人员 (1)" :value="1" />
                    <el-option label="主管理员 (2)" :value="2" />
                    <el-option label="服主 (3)" :value="3" />
                  </el-select>
                  <el-button size="small" type="danger" @click="removeAdmin(i)" style="margin-left: 10px">删除</el-button>
                </div>
                <el-button size="small" @click="addAdmin">+ 添加管理员</el-button>
              </div>
            </el-form>
          </el-scrollbar>
        </div>
      </el-tab-pane>

      <el-tab-pane label="公告&商店" name="shop">
        <div class="shop-panel">
          <div class="sub-section">
            <h4>公告系统</h4>
            <div class="announcement-controls">
              <el-input v-model="newAnnouncement" placeholder="输入公告内容..." size="small" style="width: 400px" />
              <el-button size="small" type="primary" @click="sendAnnouncement">立即发送</el-button>
              <el-select v-model="announceInterval" size="small" style="width: 150px; margin-left: 20px">
                <el-option label="每5分钟" :value="5" />
                <el-option label="每15分钟" :value="15" />
                <el-option label="每30分钟" :value="30" />
                <el-option label="每小时" :value="60" />
              </el-select>
              <el-switch v-model="autoAnnounce" active-text="定时公告" />
            </div>
            <div class="announce-history" style="margin-top: 12px">
              <h5>公告记录</h5>
              <div v-for="a in announcements" :key="a.id" class="announce-item">
                <span class="announce-time">{{ a.time }}</span>
                <span class="announce-text">{{ a.text }}</span>
              </div>
            </div>
          </div>

          <div class="sub-section" style="margin-top: 24px">
            <h4>游戏商店</h4>
            <el-table :data="shopItems" size="small">
              <el-table-column prop="name" label="物品名称" width="150" />
              <el-table-column prop="itemId" label="物品ID" width="120">
                <template #default="{ row }">
                  <span style="font-family: monospace; font-size: 12px;">{{ row.itemId }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="price" label="价格(积分)" width="110" />
              <el-table-column prop="quantity" label="数量" width="80" />
              <el-table-column prop="category" label="分类" width="100" />
              <el-table-column label="操作">
                <template #default="{ row }">
                  <el-button size="small" @click="editShopItem(row)">编辑</el-button>
                  <el-button size="small" type="danger" @click="deleteShopItem(row)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
            <el-button size="small" type="primary" style="margin-top: 10px" @click="addShopItem">添加商品</el-button>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="MOD 管理" name="mods">
        <div class="mod-manage">
          <div class="tab-toolbar">
            <el-button type="primary" size="small" @click="installMod">
              <el-icon><Plus /></el-icon> 安装 MOD
            </el-button>
          </div>
          <el-table :data="installedMods" size="small">
            <el-table-column prop="name" label="MOD 名称" min-width="180" />
            <el-table-column prop="version" label="版本" width="80" />
            <el-table-column prop="author" label="作者" width="120" />
            <el-table-column prop="enabled" label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.enabled ? 'success' : 'info'" size="small">{{ row.enabled ? '启用' : '禁用' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="180">
              <template #default="{ row }">
                <el-button size="small" @click="toggleMod(row)">{{ row.enabled ? '禁用' : '启用' }}</el-button>
                <el-button size="small" type="danger" @click="uninstallMod(row)">卸载</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>

      <el-tab-pane label="备份管理" name="backup">
        <div class="backup-manage">
          <div class="tab-toolbar">
            <el-button type="primary" size="small" @click="createBackup">立即备份</el-button>
            <el-switch v-model="autoBackup" active-text="自动备份" style="margin-left: 20px" />
            <el-select v-model="backupInterval" size="small" style="width: 120px; margin-left: 10px" v-if="autoBackup">
              <el-option label="每小时" :value="1" />
              <el-option label="每6小时" :value="6" />
              <el-option label="每12小时" :value="12" />
              <el-option label="每天" :value="24" />
            </el-select>
          </div>

          <el-table :data="backups" size="small">
            <el-table-column prop="name" label="备份名称" min-width="200" />
            <el-table-column prop="size" label="大小" width="100" />
            <el-table-column prop="date" label="创建时间" width="180" />
            <el-table-column label="操作" width="180">
              <template #default="{ row }">
                <el-button size="small" @click="restoreBackup(row)">恢复</el-button>
                <el-button size="small" type="danger" @click="deleteBackup(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>

      <el-tab-pane label="统计监控" name="stats">
        <div class="stats-panel">
          <div class="stats-grid-3">
            <div class="stat-card-small">
              <h4>在线人数</h4>
              <div ref="playerChartRef" style="height: 200px;"></div>
            </div>
            <div class="stat-card-small">
              <h4>TPS 监控</h4>
              <div ref="tpsChartRef" style="height: 200px;"></div>
            </div>
            <div class="stat-card-small">
              <h4>内存使用</h4>
              <div ref="memChartRef" style="height: 200px;"></div>
            </div>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>

    <div class="server-not-installed" v-else>
      <div class="not-installed-content">
        <div class="install-icon">🎮</div>
        <h3>七日杀专用服务器未安装</h3>
        <p>点击上方按钮通过 SteamCMD 一键部署 7 Days to Die Dedicated Server</p>
        <el-button type="primary" size="large" @click="installServer">
          <el-icon><Download /></el-icon> 一键部署服务端
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as echarts from 'echarts'

const activeTab = ref('console')
const serverInstalled = ref(true)
const serverRunning = ref(true)
const onlinePlayers = ref(5)
const maxPlayers = ref(16)
const tps = ref(18.5)
const fps = ref(55)
const serverMemory = ref(4096)
const playerSearch = ref('')
const consoleCommand = ref('')
const newAnnouncement = ref('')
const announceInterval = ref(15)
const autoAnnounce = ref(false)
const whitelistEnabled = ref(false)
const newWhitelist = ref('')
const autoBackup = ref(false)
const backupInterval = ref(6)

const consoleLogs = ref([
  '[2026-06-29 15:30:00] Game server started',
  '[2026-06-29 15:30:05] Loading world...',
  '[2026-06-29 15:30:12] World loaded: Navezgane',
  '[2026-06-29 15:30:15] Server is ready. Listening on port 26900',
  '[2026-06-29 15:30:20] Player "Survivor1" joined',
  '[2026-06-29 15:31:00] Player "ZombieHunter" joined',
  '[2026-06-29 15:35:00] BloodMoon starting...'
])

const players = ref([
  { name: 'Survivor1', steamId: '76561198123456789', playTime: '120h', level: 45, score: 12500, online: true, banned: false },
  { name: 'ZombieHunter', steamId: '76561198234567890', playTime: '85h', level: 32, score: 8900, online: true, banned: false },
  { name: 'BaseBuilder', steamId: '76561198345678901', playTime: '200h', level: 78, score: 25000, online: true, banned: false },
  { name: 'LootMaster', steamId: '76561198456789012', playTime: '50h', level: 18, score: 3200, online: false, banned: false },
  { name: 'NightStalker', steamId: '76561198567890123', playTime: '30h', level: 12, score: 1500, online: true, banned: true }
])

const filteredPlayers = computed(() =>
  players.value.filter(p => p.name.toLowerCase().includes(playerSearch.value.toLowerCase()))
)

const whitelist = ref(['76561198123456789', '76561198234567890'])

const announcements = ref([
  { id: '1', time: '15:30', text: '欢迎来到 OmniPanel 七日杀服务器!' },
  { id: '2', time: '15:35', text: '血月即将来临,请做好准备!' }
])

const admins = ref([
  { steamId: '76561198000000001', level: 3 },
  { steamId: '76561198000000002', level: 0 }
])

const serverConfig = reactive({
  ServerName: 'OmniPanel 七日杀服务器',
  ServerDescription: '欢迎来到 OmniPanel 管理的七日杀服务器!',
  ServerPassword: '',
  ServerMaxPlayerCount: 16,
  ServerPort: 26900,
  ServerVisibility: 2,
  GameDifficulty: 4,
  GameMode: 'GameModeSurvival',
  WorldGenSeed: 'OmniPanel2026',
  WorldGenSize: 8192,
  DayNightLength: 60,
  MaxSpawnedZombies: 60,
  MaxSpawnedAnimals: 50,
  EACEnabled: true,
  BloodMoonFrequency: 7,
  AirDropFrequency: 24,
  DropOnDeath: 1,
  DropOnQuit: 1
})

const shopItems = ref([
  { name: '急救包', itemId: 'medicalFirstAidBandage', price: 100, quantity: 1, category: '医疗' },
  { name: '弹药 - 9mm', itemId: 'ammo9mmBullet', price: 50, quantity: 50, category: '弹药' },
  { name: '瓶装水', itemId: 'drinkJarBoiledWater', price: 30, quantity: 1, category: '食物' }
])

const installedMods = ref([
  { name: 'SMX UI Mod', version: '1.2', author: 'Sirillion', enabled: true },
  { name: 'Server Tools', version: '3.1', author: 'Sorrows', enabled: true },
  { name: 'QuickStack', version: '2.0', author: 'Demono', enabled: false }
])

const backups = ref([
  { name: '20260629_1500_Saves.zip', size: '1.2GB', date: '2026-06-29 15:00' },
  { name: '20260628_1500_Saves.zip', size: '1.1GB', date: '2026-06-28 15:00' },
  { name: '20260627_1500_Saves.zip', size: '1.0GB', date: '2026-06-27 15:00' }
])

const playerChartRef = ref()
const tpsChartRef = ref()
const memChartRef = ref()
let gameCharts: echarts.ECharts[] = []

function installServer() {
  ElMessageBox.confirm('将使用 SteamCMD 下载并安装七日杀专用服务器,继续吗？', '确认安装')
    .then(() => {
      serverInstalled.value = true
      ElMessage.success('七日杀服务端安装成功!')
    }).catch(() => {})
}

function toggleServer() {
  serverRunning.value = !serverRunning.value
  if (serverRunning.value) {
    consoleLogs.value.push(`[${new Date().toLocaleTimeString()}] Server started`)
    ElMessage.success('服务器已启动')
  } else {
    consoleLogs.value.push(`[${new Date().toLocaleTimeString()}] Server stopped`)
    ElMessage.info('服务器已停止')
  }
}

function restartServer() {
  serverRunning.value = false
  consoleLogs.value.push(`[${new Date().toLocaleTimeString()}] Server restarting...`)
  setTimeout(() => {
    serverRunning.value = true
    consoleLogs.value.push(`[${new Date().toLocaleTimeString()}] Server started`)
  }, 2000)
  ElMessage.success('服务器正在重启...')
}

function sendConsoleCommand() {
  if (!consoleCommand.value.trim()) return
  consoleLogs.value.push(`[CMD] ${consoleCommand.value}`)
  consoleCommand.value = ''
}

function kickPlayer(player: any) {
  ElMessageBox.confirm(`确定踢出玩家 "${player.name}" 吗？`, '确认踢出').then(() => {
    players.value = players.value.filter(p => p.steamId !== player.steamId)
    ElMessage.success(`玩家 ${player.name} 已被踢出`)
    consoleLogs.value.push(`[ADMIN] Kicked player: ${player.name} (${player.steamId})`)
  }).catch(() => {})
}

function toggleBan(player: any) {
  player.banned = !player.banned
  ElMessage.success(`玩家 ${player.name} ${player.banned ? '已封禁' : '已解封'}`)
  consoleLogs.value.push(`[ADMIN] ${player.banned ? 'Banned' : 'Unbanned'} player: ${player.name}`)
}

function teleportPlayer(player: any) { ElMessage.info(`传送玩家 ${player.name} (功能开发中)`) }
function giveItem(player: any) { ElMessage.info(`给予物品给 ${player.name} (功能开发中)`) }

function addWhitelist() {
  if (newWhitelist.value) {
    whitelist.value.push(newWhitelist.value)
    newWhitelist.value = ''
  }
}
function removeWhitelist(i: number) { whitelist.value.splice(i, 1) }

function sendAnnouncement() {
  if (!newAnnouncement.value.trim()) return
  announcements.value.unshift({ id: Date.now().toString(), time: new Date().toLocaleTimeString(), text: newAnnouncement.value })
  consoleLogs.value.push(`[ANNOUNCE] ${newAnnouncement.value}`)
  newAnnouncement.value = ''
  ElMessage.success('公告已发送')
}

function saveConfig() { ElMessage.success('配置已保存') }
function applyConfig() {
  ElMessageBox.confirm('应用新配置需要重启服务器,确认应用？', '确认应用').then(() => {
    ElMessage.success('配置已应用,服务器即将重启')
  }).catch(() => {})
}
function backupConfig() { ElMessage.success('配置已备份') }

function addAdmin() { admins.value.push({ steamId: '', level: 0 }) }
function removeAdmin(i: number) { admins.value.splice(i, 1) }

function refreshPlayers() { ElMessage.success('玩家列表已刷新') }

function addShopItem() { shopItems.value.push({ name: '新物品', itemId: 'newItem', price: 0, quantity: 1, category: '其他' }) }
function editShopItem(row: any) { ElMessage.info(`编辑 ${row.name}`) }
function deleteShopItem(row: any) { shopItems.value = shopItems.value.filter(i => i !== row) }

function installMod() { ElMessage.info('MOD 安装 (文件选择器)') }
function toggleMod(mod: any) { mod.enabled = !mod.enabled; ElMessage.success(`MOD ${mod.name} ${mod.enabled ? '已启用' : '已禁用'}`) }
function uninstallMod(mod: any) {
  ElMessageBox.confirm(`确定卸载 MOD "${mod.name}" 吗？`).then(() => {
    installedMods.value = installedMods.value.filter(m => m !== mod)
    ElMessage.success('MOD 已卸载')
  }).catch(() => {})
}

function createBackup() { backups.value.unshift({ name: `${new Date().toISOString().split('T')[0]}_${new Date().toLocaleTimeString().replace(/:/g, '')}_Saves.zip`, size: '1.2GB', date: new Date().toLocaleString() }); ElMessage.success('备份已创建') }
function restoreBackup(row: any) { ElMessage.success(`从备份 ${row.name} 恢复中...`) }
function deleteBackup(row: any) { backups.value = backups.value.filter(b => b !== row); ElMessage.success('备份已删除') }

function initGameCharts() {
  if (!playerChartRef.value) return
  const makeChart = (ref: HTMLElement, color: string, data: number[]) => {
    const c = echarts.init(ref)
    c.setOption({
      grid: { top: 5, right: 5, bottom: 15, left: 35, containLabel: false },
      xAxis: { type: 'category', data: Array.from({ length: 24 }, (_, i) => `${i}h`), show: false },
      yAxis: { type: 'value', splitLine: { lineStyle: { color: '#333' } } },
      series: [{ data: data || Array.from({ length: 24 }, () => Math.floor(Math.random() * 16)), type: 'line', smooth: true, showSymbol: false, lineStyle: { color, width: 2 }, areaStyle: { color: color + '33' } }]
    })
    gameCharts.push(c)
  }
  makeChart(playerChartRef.value, '#409eff', [0, 0, 1, 5, 3, 8, 12, 15, 14, 16, 10, 8, 5, 3, 5, 8, 12, 15, 14, 10, 8, 5, 3, 2])
  makeChart(tpsChartRef.value, '#67c23a', Array.from({ length: 24 }, () => 15 + Math.random() * 5))
  makeChart(memChartRef.value, '#e6a23c', Array.from({ length: 24 }, () => 3000 + Math.random() * 2000))
}

onMounted(() => { setTimeout(initGameCharts, 500) })
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

.console-panel { background: #0d0d0d; border-radius: 8px; overflow: hidden; }
.console-output { height: 400px; overflow-y: auto; padding: 12px; font-family: 'Consolas', monospace; font-size: 13px; }
.console-line { color: #ccc; line-height: 1.6; }
.console-warn { color: #e6a23c; }
.console-error { color: #f56c6c; }
.console-input { padding: 10px; background: #1a1a1a; }

.tab-toolbar { display: flex; justify-content: space-between; align-items: center; padding: 8px 0; margin-bottom: 10px; }
.toolbar-right { display: flex; gap: 8px; }

.sub-section { margin-top: 20px; }
.sub-section h4 { margin-bottom: 10px; font-size: 14px; }

.config-editor { padding: 12px 0; }
.config-toolbar { margin-bottom: 16px; display: flex; gap: 8px; }

.admin-list { display: flex; flex-direction: column; gap: 10px; }
.admin-item { display: flex; align-items: center; }

.announce-item { padding: 6px 0; border-bottom: 1px solid var(--border-color); font-size: 13px; }
.announce-time { color: var(--text-secondary); margin-right: 12px; font-family: monospace; }

.stats-grid-3 { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; }
.stat-card-small { background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 10px; padding: 14px; }
.stat-card-small h4 { margin-bottom: 8px; font-size: 13px; }

.server-not-installed { display: flex; align-items: center; justify-content: center; min-height: 400px; }
.not-installed-content { text-align: center; }
.install-icon { font-size: 72px; margin-bottom: 20px; }
.not-installed-content h3 { margin-bottom: 10px; }
.not-installed-content p { color: var(--text-secondary); margin-bottom: 24px; }
</style>
