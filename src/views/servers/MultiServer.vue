<template>
  <div class="multi-server">
    <div class="page-header">
      <h2>多机管理</h2>
      <div class="header-actions">
        <el-button type="primary" @click="showAddServerDialog = true">
          <el-icon><Plus /></el-icon> 添加服务器
        </el-button>
        <el-button @click="batchCommand" :disabled="selectedServers.length === 0">
          <el-icon><Operation /></el-icon> 批量操作
        </el-button>
      </div>
    </div>

    <div class="servers-grid">
      <div class="server-card card-hover" v-for="server in servers" :key="server.id"
           :class="{ 'server-offline': server.status === 'offline', 'server-selected': selectedServers.includes(server.id) }"
           @click="selectServer(server.id)">
        <div class="server-card-header">
          <div class="server-name">
            <span :class="server.status === 'online' ? 'status-online' : 'status-offline'">●</span>
            {{ server.name }}
          </div>
          <el-dropdown @command="(cmd: string) => handleServerAction(cmd, server)">
            <el-button text circle size="small">
              <el-icon><MoreFilled /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="ssh">SSH 连接</el-dropdown-item>
                <el-dropdown-item command="reboot">重启服务器</el-dropdown-item>
                <el-dropdown-item command="edit">编辑</el-dropdown-item>
                <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>

        <div class="server-body">
          <div class="server-ip">{{ server.ip }}:{{ server.port }}</div>
          <el-tag :type="server.status === 'online' ? 'success' : 'info'" size="small">
            {{ server.status === 'online' ? '在线' : '离线' }}
          </el-tag>
        </div>

        <div class="server-metrics">
          <div class="metric-item">
            <span class="metric-label">CPU</span>
            <div class="metric-bar">
              <div class="metric-fill" :style="{ width: server.cpu + '%', background: getCpuColor(server.cpu) }"></div>
            </div>
            <span class="metric-value">{{ server.cpu }}%</span>
          </div>
          <div class="metric-item">
            <span class="metric-label">内存</span>
            <div class="metric-bar">
              <div class="metric-fill" :style="{ width: server.memory + '%', background: getMemColor(server.memory) }"></div>
            </div>
            <span class="metric-value">{{ server.memory }}%</span>
          </div>
          <div class="metric-item">
            <span class="metric-label">磁盘</span>
            <div class="metric-bar">
              <div class="metric-fill" :style="{ width: server.disk + '%', background: getDiskColor(server.disk) }"></div>
            </div>
            <span class="metric-value">{{ server.disk }}%</span>
          </div>
        </div>

        <div class="server-footer">
          <span>{{ server.os }}</span>
          <span>{{ server.uptime }}</span>
        </div>
      </div>
    </div>

    <div class="manage-sections">
      <div class="manage-card">
        <h4>批量操作</h4>
        <div class="batch-actions">
          <el-input
            v-model="batchCommandText"
            type="textarea"
            :rows="4"
            placeholder="输入要批量执行的命令..."
          />
          <div class="batch-buttons">
            <el-button type="primary" @click="executeBatchCommand" :disabled="selectedServers.length === 0">
              批量执行
            </el-button>
            <el-select v-model="batchTemplate" placeholder="选择预设命令" @change="applyBatchTemplate">
              <el-option label="系统更新" value="apt update && apt upgrade -y" />
              <el-option label="查看磁盘" value="df -h" />
              <el-option label="查看内存" value="free -h" />
              <el-option label="Docker 状态" value="docker ps -a" />
              <el-option label="重启 Nginx" value="systemctl restart nginx" />
              <el-option label="查看所有用户" value="cat /etc/passwd | grep -v nologin" />
            </el-select>
          </div>
        </div>
      </div>

      <div class="manage-card">
        <h4>服务器分组</h4>
        <div class="group-list">
          <div class="group-item" v-for="group in serverGroups" :key="group.name">
            <div class="group-header-row">
              <span class="group-name">{{ group.name }}</span>
              <span class="group-count">{{ group.servers.length }}台</span>
            </div>
            <div class="group-tags">
              <el-tag v-for="s in group.servers" :key="s" size="small" type="info">{{ s }}</el-tag>
            </div>
          </div>
        </div>
      </div>

      <div class="manage-card">
        <h4>告警设置</h4>
        <div class="alert-list">
          <div class="alert-item" v-for="alert in alerts" :key="alert.id">
            <div class="alert-info">
              <div class="alert-name">{{ alert.name }}</div>
              <div class="alert-rule">{{ alert.condition }}</div>
            </div>
            <el-switch v-model="alert.enabled" size="small" />
          </div>
          <el-button size="small" style="width: 100%" @click="addAlert">+ 添加告警规则</el-button>
        </div>
      </div>
    </div>

    <el-dialog v-model="showAddServerDialog" title="添加服务器" width="500px">
      <el-form :model="newServer" label-width="100px">
        <el-form-item label="服务器名称">
          <el-input v-model="newServer.name" placeholder="例: 生产服务器 01" />
        </el-form-item>
        <el-form-item label="IP 地址">
          <el-input v-model="newServer.ip" placeholder="192.168.1.100" />
        </el-form-item>
        <el-form-item label="SSH 端口">
          <el-input-number v-model="newServer.port" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="newServer.username" placeholder="root" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="newServer.password" type="password" show-password />
        </el-form-item>
        <el-form-item label="分组">
          <el-select v-model="newServer.group">
            <el-option label="生产环境" value="生产环境" />
            <el-option label="测试环境" value="测试环境" />
            <el-option label="游戏服务器" value="游戏服务器" />
            <el-option label="默认" value="默认" />
          </el-select>
        </el-form-item>
        <el-form-item label="标签">
          <el-select v-model="newServer.tags" multiple placeholder="选择标签" allow-create />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddServerDialog = false">取消</el-button>
        <el-button type="primary" @click="addServer">添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { v4 as uuid } from 'uuid'

const showAddServerDialog = ref(false)
const selectedServers = ref<string[]>([])
const batchCommandText = ref('')
const batchTemplate = ref('')

interface Server {
  id: string
  name: string
  ip: string
  port: number
  username: string
  password: string
  group: string
  tags: string[]
  status: string
  os: string
  uptime: string
  cpu: number
  memory: number
  disk: number
}

const servers = ref<Server[]>([
  { id: '1', name: '主 Web 服务器', ip: '192.168.1.100', port: 22, username: 'root', password: '***', group: '生产环境', tags: ['Web', 'Nginx'], status: 'online', os: 'Ubuntu 22.04', uptime: '30天', cpu: 23, memory: 45, disk: 32 },
  { id: '2', name: '数据库服务器', ip: '192.168.1.101', port: 22, username: 'root', password: '***', group: '生产环境', tags: ['DB', 'MySQL'], status: 'online', os: 'Ubuntu 22.04', uptime: '30天', cpu: 12, memory: 68, disk: 55 },
  { id: '3', name: '七日杀服务器', ip: '192.168.2.50', port: 22, username: 'steam', password: '***', group: '游戏服务器', tags: ['Game', '7Days'], status: 'online', os: 'Ubuntu 20.04', uptime: '15天', cpu: 45, memory: 72, disk: 38 },
  { id: '4', name: '测试服务器', ip: '10.0.0.10', port: 2222, username: 'dev', password: '***', group: '测试环境', tags: ['Dev'], status: 'offline', os: 'CentOS 8', uptime: '-', cpu: 0, memory: 0, disk: 0 },
  { id: '5', name: '备份服务器', ip: '192.168.3.100', port: 22, username: 'backup', password: '***', group: '生产环境', tags: ['Backup', 'Storage'], status: 'online', os: 'Debian 11', uptime: '60天', cpu: 5, memory: 15, disk: 78 }
])

const serverGroups = computed(() => {
  const map: Record<string, string[]> = {}
  servers.value.forEach(s => {
    if (!map[s.group]) map[s.group] = []
    map[s.group].push(s.name)
  })
  return Object.entries(map).map(([name, servers]) => ({ name, servers }))
})

const alerts = ref([
  { id: '1', name: 'CPU 过载', condition: 'CPU > 90% 持续 5分钟', enabled: true },
  { id: '2', name: '内存不足', condition: '内存 > 85% 持续 5分钟', enabled: true },
  { id: '3', name: '磁盘不足', condition: '磁盘 > 90%', enabled: true },
  { id: '4', name: '服务器离线', condition: '任意服务器离线 > 1分钟', enabled: true }
])

const newServer = reactive({
  name: '', ip: '', port: 22, username: 'root', password: '', group: '默认', tags: [] as string[]
})

function selectServer(id: string) {
  const idx = selectedServers.value.indexOf(id)
  if (idx === -1) selectedServers.value.push(id)
  else selectedServers.value.splice(idx, 1)
}

function handleServerAction(cmd: string, server: Server) {
  if (cmd === 'delete') {
    ElMessageBox.confirm(`确定要删除服务器 "${server.name}" 吗？`, '确认删除', { type: 'warning' })
      .then(() => {
        servers.value = servers.value.filter(s => s.id !== server.id)
        ElMessage.success('服务器已删除')
      }).catch(() => {})
  } else if (cmd === 'ssh') {
    ElMessage.info(`打开 SSH 连接到 ${server.name}`)
  } else if (cmd === 'reboot') {
    ElMessageBox.confirm(`确定要重启 "${server.name}" 吗？`, '确认重启', { type: 'warning' })
      .then(() => ElMessage.success('重启指令已发送')).catch(() => {})
  } else if (cmd === 'edit') {
    ElMessage.info(`编辑 ${server.name}`)
  }
}

function addServer() {
  if (!newServer.name || !newServer.ip) {
    ElMessage.warning('请填写服务器名称和 IP 地址')
    return
  }
  servers.value.push({
    id: uuid(),
    ...newServer,
    status: 'offline',
    os: '未知',
    uptime: '-',
    cpu: 0,
    memory: 0,
    disk: 0,
    password: '***'
  })
  showAddServerDialog.value = false
  ElMessage.success('服务器已添加')
  Object.assign(newServer, { name: '', ip: '', port: 22, username: 'root', password: '', group: '默认', tags: [] })
}

function batchCommand() {
  if (selectedServers.value.length === 0) {
    ElMessage.warning('请选择目标服务器')
    return
  }
  ElMessage.info(`批量操作 ${selectedServers.value.length} 台服务器 (模拟)`)
}

function executeBatchCommand() {
  if (!batchCommandText.value.trim()) {
    ElMessage.warning('请输入要执行的命令')
    return
  }
  ElMessage.success(`批量命令已发送到 ${selectedServers.value.length} 台服务器`)
  batchCommandText.value = ''
}

function applyBatchTemplate(value: string) {
  batchCommandText.value = value
}

function addAlert() {
  alerts.value.push({
    id: uuid(),
    name: '新告警规则',
    condition: '条件待配置',
    enabled: false
  })
}

function getCpuColor(val: number) { return val > 80 ? '#f56c6c' : val > 60 ? '#e6a23c' : '#67c23a' }
function getMemColor(val: number) { return val > 80 ? '#f56c6c' : val > 60 ? '#e6a23c' : '#67c23a' }
function getDiskColor(val: number) { return val > 80 ? '#f56c6c' : val > 60 ? '#e6a23c' : '#67c23a' }
</script>

<style scoped>
.multi-server { padding: 0; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h2 { font-size: 22px; font-weight: 600; }
.header-actions { display: flex; gap: 10px; }

.servers-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 16px; margin-bottom: 24px; }
.server-card { background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 12px; padding: 16px; cursor: pointer; }
.server-card.server-offline { opacity: 0.6; }
.server-card.server-selected { border-color: var(--accent-color); box-shadow: 0 0 0 2px rgba(64,158,255,0.2); }
.server-card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.server-name { font-weight: 600; display: flex; align-items: center; gap: 6px; }
.server-body { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.server-ip { font-size: 13px; color: var(--text-secondary); font-family: monospace; }

.server-metrics { display: flex; flex-direction: column; gap: 6px; }
.metric-item { display: flex; align-items: center; gap: 8px; font-size: 12px; }
.metric-label { width: 32px; color: var(--text-secondary); }
.metric-bar { flex: 1; height: 4px; background: var(--bg-secondary); border-radius: 2px; overflow: hidden; }
.metric-fill { height: 100%; border-radius: 2px; transition: width 1s ease; }
.metric-value { width: 32px; text-align: right; font-family: monospace; }
.server-footer { display: flex; justify-content: space-between; margin-top: 12px; font-size: 11px; color: var(--text-secondary); }

.manage-sections { display: grid; grid-template-columns: 2fr 1fr 1fr; gap: 16px; }
.manage-card { background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 12px; padding: 16px; }
.manage-card h4 { margin-bottom: 12px; font-size: 14px; font-weight: 600; }

.batch-buttons { display: flex; gap: 8px; margin-top: 10px; }

.group-list { display: flex; flex-direction: column; gap: 8px; }
.group-item { padding: 8px; background: var(--bg-secondary); border-radius: 6px; }
.group-header-row { display: flex; justify-content: space-between; margin-bottom: 4px; }
.group-name { font-weight: 500; font-size: 13px; }
.group-count { font-size: 11px; color: var(--text-secondary); }
.group-tags { display: flex; gap: 4px; flex-wrap: wrap; }

.alert-list { display: flex; flex-direction: column; gap: 8px; }
.alert-item { display: flex; justify-content: space-between; align-items: center; padding: 8px; background: var(--bg-secondary); border-radius: 6px; }
.alert-name { font-size: 13px; font-weight: 500; }
.alert-rule { font-size: 11px; color: var(--text-secondary); }
</style>
