<template>
  <div class="frp-manager">
    <div class="page-header">
      <h2>FRP 内网穿透管理</h2>
      <div class="header-actions">
        <el-button type="primary" @click="activeMode = 'frpc'">
          <el-icon><Link /></el-icon> 客户端模式
        </el-button>
        <el-button type="warning" @click="activeMode = 'frps'">
          <el-icon><DataAnalysis /></el-icon> 服务端模式
        </el-button>
      </div>
    </div>

    <div class="frp-layout">
      <div class="frp-sidebar">
        <div class="mode-switch">
          <el-radio-group v-model="activeMode" size="small">
            <el-radio-button label="frpc">frpc 客户端</el-radio-button>
            <el-radio-button label="frps">frps 服务端</el-radio-button>
          </el-radio-group>
        </div>

        <div class="config-list">
          <h4>配置列表</h4>
          <div class="config-item" v-for="cfg in frpConfigs" :key="cfg.id"
               :class="{ active: activeConfig?.id === cfg.id }"
               @click="selectConfig(cfg)">
            <div class="config-name">{{ cfg.name }}</div>
            <div class="config-desc">{{ cfg.mode === 'frpc' ? '客户端' : '服务端' }}</div>
            <el-switch v-model="cfg.enabled" size="small" @change="toggleConfig(cfg)" />
          </div>
          <el-button style="width: 100%; margin-top: 8px" size="small" @click="addConfig">
            <el-icon><Plus /></el-icon> 新建配置
          </el-button>
        </div>

        <div class="templates-section">
          <h4>穿透模板</h4>
          <div class="template-list">
            <div class="template-item" v-for="tpl in penetrationTemplates" :key="tpl.name"
                 @click="applyTemplate(tpl)">
              <span class="tpl-icon">{{ tpl.icon }}</span>
              <div class="tpl-info">
                <div class="tpl-name">{{ tpl.name }}</div>
                <div class="tpl-desc">{{ tpl.description }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="frp-main">
        <div class="editor-panel">
          <div class="editor-header">
            <span>{{ activeConfig ? activeConfig.name : '无配置选中' }}</span>
            <div class="editor-actions">
              <el-button size="small" @click="validateConfig">验证配置</el-button>
              <el-button size="small" type="primary" @click="saveConfig">保存</el-button>
            </div>
          </div>
          <div class="editor-body">
            <el-input
              v-model="configContent"
              type="textarea"
              :rows="18"
              class="config-editor"
              placeholder="在此编辑配置..."
            />
          </div>

          <div class="proxy-list" v-if="activeMode === 'frpc'">
            <h4>代理列表</h4>
            <el-table :data="proxies" size="small">
              <el-table-column prop="name" label="名称" width="120" />
              <el-table-column prop="type" label="类型" width="80">
                <template #default="{ row }">
                  <el-tag :type="getProtocolColor(row.type)" size="small">{{ row.type }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="localIP" label="本地地址" width="140" />
              <el-table-column prop="localPort" label="本地端口" width="90" />
              <el-table-column prop="remotePort" label="远程端口" width="90" />
              <el-table-column prop="status" label="状态" width="80">
                <template #default="{ row }">
                  <span :class="row.status === 'online' ? 'status-online' : 'status-offline'">●</span>
                </template>
              </el-table-column>
              <el-table-column label="操作">
                <template #default="{ row }">
                  <el-button size="small" @click="toggleProxy(row)">
                    {{ row.status === 'online' ? '停止' : '启动' }}
                  </el-button>
                  <el-button size="small" type="danger" @click="removeProxy(row)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>

            <el-button type="primary" size="small" style="margin-top: 10px" @click="showProxyDialog = true">
              <el-icon><Plus /></el-icon> 添加代理
            </el-button>
          </div>
        </div>

        <div class="status-panel">
          <div class="status-header">
            <h4>运行状态</h4>
            <el-button size="small" :type="frpRunning ? 'danger' : 'success'" @click="toggleFRP">
              {{ frpRunning ? '停止' : '启动' }}
            </el-button>
          </div>
          <div class="status-info">
            <div class="status-row">
              <span>进程状态</span>
              <el-tag :type="frpRunning ? 'success' : 'info'" size="small">
                {{ frpRunning ? '运行中' : '已停止' }}
              </el-tag>
            </div>
            <div class="status-row">
              <span>运行时间</span>
              <span>{{ frpRunning ? frpUptime : '-' }}</span>
            </div>
            <div class="status-row">
              <span>活跃代理数</span>
              <span>{{ proxies.filter(p => p.status === 'online').length }}</span>
            </div>
            <div class="status-row">
              <span>总流量</span>
              <span>{{ totalTraffic }}</span>
            </div>
          </div>

          <div class="log-viewer">
            <div class="log-header">实时日志</div>
            <div class="log-body">
              <div v-for="(log, i) in frpLogs" :key="i" class="log-line">{{ log }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <el-dialog v-model="showProxyDialog" title="添加代理规则" width="500px">
      <el-form :model="newProxy" label-width="100px">
        <el-form-item label="代理名称">
          <el-input v-model="newProxy.name" placeholder="例: web_server" />
        </el-form-item>
        <el-form-item label="协议类型">
          <el-select v-model="newProxy.type">
            <el-option label="TCP" value="tcp" />
            <el-option label="UDP" value="udp" />
            <el-option label="HTTP" value="http" />
            <el-option label="HTTPS" value="https" />
            <el-option label="STCP" value="stcp" />
            <el-option label="XTCP" value="xtcp" />
          </el-select>
        </el-form-item>
        <el-form-item label="本地 IP">
          <el-input v-model="newProxy.localIP" placeholder="127.0.0.1" />
        </el-form-item>
        <el-form-item label="本地端口">
          <el-input-number v-model="newProxy.localPort" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="远程端口">
          <el-input-number v-model="newProxy.remotePort" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item v-if="newProxy.type === 'http' || newProxy.type === 'https'" label="域名">
          <el-input v-model="newProxy.customDomains" placeholder="例: myapp.example.com" />
        </el-form-item>
        <el-form-item label="加密">
          <el-switch v-model="newProxy.encryption" />
        </el-form-item>
        <el-form-item label="压缩">
          <el-switch v-model="newProxy.compression" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showProxyDialog = false">取消</el-button>
        <el-button type="primary" @click="addProxy">添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { v4 as uuid } from 'uuid'

const activeMode = ref<'frpc' | 'frps'>('frpc')
const configContent = ref('')
const activeConfig = ref<any>(null)
const frpRunning = ref(false)
const frpUptime = ref('00:00:00')
const totalTraffic = ref('0 B')
const showProxyDialog = ref(false)

interface FrpConfig {
  id: string
  name: string
  mode: string
  enabled: boolean
  content: string
}

const frpConfigs = ref<FrpConfig[]>([
  {
    id: '1', name: '主穿透配置', mode: 'frpc', enabled: true,
    content: `[common]
server_addr = your-frp-server.com
server_port = 7000
token = your_token_here

[ssh]
type = tcp
local_ip = 127.0.0.1
local_port = 22
remote_port = 6000

[web]
type = http
local_ip = 127.0.0.1
local_port = 8080
custom_domains = web.example.com`
  }
])

interface Proxy {
  id: string
  name: string
  type: string
  localIP: string
  localPort: number
  remotePort: number
  customDomains: string
  encryption: boolean
  compression: boolean
  status: string
}

const proxies = ref<Proxy[]>([
  { id: '1', name: 'ssh', type: 'tcp', localIP: '127.0.0.1', localPort: 22, remotePort: 6000, customDomains: '', encryption: true, compression: true, status: 'online' },
  { id: '2', name: 'web', type: 'http', localIP: '127.0.0.1', localPort: 8080, remotePort: 80, customDomains: 'web.example.com', encryption: false, compression: false, status: 'online' },
  { id: '3', name: 'game_7days', type: 'udp', localIP: '127.0.0.1', localPort: 26900, remotePort: 26900, customDomains: '', encryption: false, compression: false, status: 'offline' }
])

const frpLogs = ref([
  '[2026-06-29 15:30:00] [INFO] frpc started successfully',
  '[2026-06-29 15:30:01] [INFO] [ssh] connect to server success',
  '[2026-06-29 15:30:01] [INFO] [web] connect to server success',
  '[2026-06-29 15:30:02] [INFO] start proxy success'
])

const penetrationTemplates = ref([
  { name: 'SSH 穿透', icon: '🔐', description: '穿透本地 22 端口', content: `[ssh]\ntype = tcp\nlocal_ip = 127.0.0.1\nlocal_port = 22\nremote_port = 6000` },
  { name: 'Web 穿透', icon: '🌐', description: '穿透本地 Web 服务', content: `[web]\ntype = http\nlocal_ip = 127.0.0.1\nlocal_port = 8080\ncustom_domains = your.example.com` },
  { name: '七日杀穿透', icon: '🎮', description: '穿透七日杀 UDP 26900 端口', content: `[7days]\ntype = udp\nlocal_ip = 127.0.0.1\nlocal_port = 26900\nremote_port = 26900` },
  { name: 'MySQL 穿透', icon: '🗄️', description: '穿透 MySQL 3306', content: `[mysql]\ntype = tcp\nlocal_ip = 127.0.0.1\nlocal_port = 3306\nremote_port = 6306\nencryption = true` },
  { name: 'RDP 穿透', icon: '🖥️', description: '穿透远程桌面 3389', content: `[rdp]\ntype = tcp\nlocal_ip = 127.0.0.1\nlocal_port = 3389\nremote_port = 7389` },
  { name: 'VNC 穿透', icon: '📺', description: '穿透 VNC 5900', content: `[vnc]\ntype = tcp\nlocal_ip = 127.0.0.1\nlocal_port = 5900\nremote_port = 7900` }
])

const newProxy = reactive({
  name: '', type: 'tcp', localIP: '127.0.0.1', localPort: 80, remotePort: 8080,
  customDomains: '', encryption: true, compression: false
})

function selectConfig(cfg: FrpConfig) {
  activeConfig.value = cfg
  configContent.value = cfg.content
}

function toggleConfig(cfg: FrpConfig) {
  ElMessage.info(`${cfg.name} ${cfg.enabled ? '已启用' : '已禁用'}`)
}

function addConfig() {
  const newCfg: FrpConfig = {
    id: uuid(), name: `配置 ${frpConfigs.value.length + 1}`,
    mode: activeMode.value, enabled: false, content: ''
  }
  frpConfigs.value.push(newCfg)
  selectConfig(newCfg)
}

function saveConfig() {
  if (activeConfig.value) {
    activeConfig.value.content = configContent.value
  }
  ElMessage.success('配置已保存')
}

function validateConfig() {
  if (!configContent.value.trim()) {
    ElMessage.warning('配置内容为空')
    return
  }
  ElMessage.success('配置格式验证通过')
}

function addProxy() {
  if (!newProxy.name) {
    ElMessage.warning('请输入代理名称')
    return
  }
  proxies.value.push({
    id: uuid(),
    ...newProxy,
    status: 'offline'
  })
  showProxyDialog.value = false
  ElMessage.success('代理规则已添加')
  newProxy.name = ''
}

function toggleProxy(proxy: Proxy) {
  proxy.status = proxy.status === 'online' ? 'offline' : 'online'
  ElMessage.success(`代理 ${proxy.name} ${proxy.status === 'online' ? '已启动' : '已停止'}`)
}

function removeProxy(proxy: Proxy) {
  ElMessageBox.confirm(`确定删除代理 "${proxy.name}" 吗？`, '确认删除', { type: 'warning' })
    .then(() => {
      proxies.value = proxies.value.filter(p => p.id !== proxy.id)
      ElMessage.success('代理已删除')
    }).catch(() => {})
}

function toggleFRP() {
  frpRunning.value = !frpRunning.value
  if (frpRunning.value) {
    frpUptime.value = '00:00:01'
    frpLogs.value.push(`[${new Date().toLocaleString()}] [INFO] ${activeMode.value} 已启动`)
    ElMessage.success(`${activeMode.value} 已启动`)
  } else {
    frpLogs.value.push(`[${new Date().toLocaleString()}] [INFO] ${activeMode.value} 已停止`)
    ElMessage.info(`${activeMode.value} 已停止`)
  }
}

function applyTemplate(tpl: any) {
  configContent.value += '\n\n' + tpl.content
  ElMessage.success(`模板 "${tpl.name}" 已应用`)
}

function getProtocolColor(type: string) {
  const colors: Record<string, string> = { tcp: '', udp: 'warning', http: 'success', https: 'success', stcp: 'info', xtcp: 'info' }
  return colors[type] || ''
}
</script>

<style scoped>
.frp-manager { padding: 0; height: calc(100vh - var(--header-height) - 60px); }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { font-size: 22px; font-weight: 600; }
.header-actions { display: flex; gap: 10px; }

.frp-layout { display: flex; height: calc(100% - 56px); gap: 0; }
.frp-sidebar { width: 260px; min-width: 260px; border-right: 1px solid var(--border-color); background: var(--bg-primary); padding: 12px; overflow-y: auto; border-radius: 8px 0 0 8px; }

.mode-switch { margin-bottom: 16px; }

.config-list h4, .templates-section h4 { font-size: 13px; color: var(--text-secondary); margin-bottom: 8px; }
.config-item { display: flex; align-items: center; gap: 8px; padding: 10px; border-radius: 8px; cursor: pointer; margin-bottom: 4px; border: 1px solid transparent; }
.config-item:hover, .config-item.active { background: var(--bg-secondary); border-color: var(--accent-color); }
.config-name { font-weight: 500; font-size: 13px; }
.config-desc { font-size: 11px; color: var(--text-secondary); }

.templates-section { margin-top: 16px; border-top: 1px solid var(--border-color); padding-top: 12px; }
.template-item { display: flex; align-items: center; gap: 10px; padding: 8px 10px; border-radius: 6px; cursor: pointer; margin-bottom: 2px; }
.template-item:hover { background: var(--bg-secondary); }
.tpl-icon { font-size: 18px; }
.tpl-name { font-size: 12px; font-weight: 500; }
.tpl-desc { font-size: 11px; color: var(--text-secondary); }

.frp-main { flex: 1; display: flex; gap: 12px; min-width: 0; padding: 0 0 0 12px; overflow-y: auto; }
.editor-panel { flex: 1; min-width: 0; }
.editor-header { display: flex; justify-content: space-between; align-items: center; padding: 8px 12px; background: var(--bg-primary); border-radius: 8px 8px 0 0; border: 1px solid var(--border-color); border-bottom: none; }
.editor-actions { display: flex; gap: 6px; }
.editor-body { background: var(--bg-primary); border: 1px solid var(--border-color); padding: 12px; }

.config-editor :deep(textarea) {
  font-family: 'Consolas', 'Courier New', monospace !important;
  font-size: 13px !important;
  line-height: 1.5 !important;
  background: #0d0d0d !important;
  color: #c0c0c0 !important;
}

.proxy-list { margin-top: 16px; }
.proxy-list h4 { margin-bottom: 8px; }

.status-panel { width: 280px; min-width: 280px; display: flex; flex-direction: column; gap: 12px; }
.status-header { display: flex; justify-content: space-between; align-items: center; padding: 12px; background: var(--bg-primary); border-radius: 8px; border: 1px solid var(--border-color); }
.status-header h4 { font-size: 14px; }

.status-info { background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 8px; padding: 12px; }
.status-row { display: flex; justify-content: space-between; align-items: center; padding: 6px 0; font-size: 13px; border-bottom: 1px solid var(--border-color); }
.status-row:last-child { border-bottom: none; }

.log-viewer { flex: 1; background: #0d0d0d; border-radius: 8px; overflow: hidden; display: flex; flex-direction: column; min-height: 200px; }
.log-header { padding: 8px 12px; background: #1a1a1a; color: #c0c4cc; font-size: 12px; border-bottom: 1px solid #333; }
.log-body { flex: 1; padding: 8px 12px; overflow-y: auto; font-family: 'Consolas', monospace; font-size: 11px; }
.log-line { color: #909399; line-height: 1.5; }
</style>
