<template>
  <div class="ssh-tool">
    <div class="page-header">
      <h2>SSH 远程工具</h2>
      <div class="header-actions">
        <el-button type="primary" @click="showHostDialog = true">
          <el-icon><Plus /></el-icon> 添加主机
        </el-button>
        <el-button type="success" @click="showQuickConnect = true">
          <el-icon><Connection /></el-icon> 快速连接
        </el-button>
      </div>
    </div>

    <div class="ssh-layout">
      <div class="ssh-sidebar">
        <div class="host-search">
          <el-input v-model="hostSearch" placeholder="搜索主机..." clearable size="small" />
        </div>

        <div class="host-groups">
          <div class="host-group" v-for="group in hostGroups" :key="group.name">
            <div class="group-header" @click="toggleGroup(group.name)">
              <el-icon><component :is="group.expanded ? 'ArrowDown' : 'ArrowRight'" /></el-icon>
              <span>{{ group.name }}</span>
              <span class="group-count">{{ group.hosts.length }}</span>
            </div>
            <div class="group-hosts" v-show="group.expanded">
              <div
                v-for="host in filterHosts(group.hosts)"
                :key="host.id"
                class="host-item"
                :class="{ active: activeHost?.id === host.id }"
                @click="connectToHost(host)"
                @contextmenu.prevent="showHostMenu($event, host)"
              >
                <div class="host-status">
                  <span :class="host.connected ? 'status-online' : 'status-offline'">●</span>
                </div>
                <div class="host-info">
                  <div class="host-name">{{ host.name }}</div>
                  <div class="host-addr">{{ host.host }}:{{ host.port }}</div>
                </div>
                <div class="host-tags">
                  <el-tag v-for="tag in host.tags" :key="tag" size="small" type="info">{{ tag }}</el-tag>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="snippet-section">
          <h4>命令片段</h4>
          <div class="snippet-list">
            <div v-for="snip in snippets" :key="snip.name" class="snippet-item" @click="runSnippet(snip)">
              <span class="snippet-name">{{ snip.name }}</span>
              <span class="snippet-cmd">{{ snip.command }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="ssh-main">
        <div class="terminal-tabs" v-if="activeHost || showQuickConnect">
          <el-tabs v-model="activeTerminalTab" type="card" closable @tab-remove="closeTerminal">
            <el-tab-pane
              v-for="tab in terminalTabs"
              :key="tab.id"
              :label="tab.label"
              :name="tab.id"
            />
          </el-tabs>

          <div class="terminal-container" ref="terminalContainer">
            <div class="terminal-toolbar">
              <div class="toolbar-left">
                <el-button size="small" @click="pasteToTerminal">粘贴</el-button>
                <el-button size="small" @click="clearTerminal">清屏</el-button>
                <el-select v-model="terminalFontSize" size="small" style="width: 80px">
                  <el-option label="12px" :value="12" />
                  <el-option label="14px" :value="14" />
                  <el-option label="16px" :value="16" />
                  <el-option label="18px" :value="18" />
                </el-select>
                <el-select v-model="terminalTheme" size="small" style="width: 100px">
                  <el-option label="Dark" value="dark" />
                  <el-option label="Light" value="light" />
                  <el-option label="Monokai" value="monokai" />
                  <el-option label="Solarized" value="solarized" />
                </el-select>
              </div>
              <div class="toolbar-right">
                <span class="connection-status" :class="activeHost?.connected ? 'status-online' : 'status-offline'">
                  {{ activeHost?.connected ? '已连接' : '未连接' }} - {{ activeHost?.name || '无连接' }}
                </span>
              </div>
            </div>
            <div class="terminal-body" ref="terminalBody"></div>
          </div>
        </div>

        <div class="no-connection" v-else>
          <div class="no-connection-content">
            <el-icon :size="64"><Connection /></el-icon>
            <h3>未连接到任何主机</h3>
            <p>选择一个已保存的主机或建立新的 SSH 连接</p>
            <div class="no-connection-actions">
              <el-button type="primary" @click="showQuickConnect = true">快速连接</el-button>
              <el-button @click="showHostDialog = true">添加主机</el-button>
            </div>
          </div>
        </div>

        <div class="sftp-panel" v-if="showSFTP && activeHost?.connected">
          <div class="sftp-header">
            <h4>SFTP 文件管理器</h4>
            <el-button size="small" @click="showSFTP = false">关闭</el-button>
          </div>
          <div class="sftp-body">
            <div class="sftp-pathbar">
              <el-input v-model="sftpPath" size="small" placeholder="/home/user">
                <template #prepend>路径</template>
              </el-input>
              <el-button size="small" @click="loadSFTPFiles">刷新</el-button>
            </div>
            <el-table :data="sftpFiles" size="small" height="300">
              <el-table-column prop="name" label="文件名" min-width="200">
                <template #default="{ row }">
                  <span :style="{ paddingLeft: row.isDir ? '0' : '20px' }">
                    <el-icon v-if="row.isDir"><Folder /></el-icon>
                    <el-icon v-else><Document /></el-icon>
                    {{ row.name }}
                  </span>
                </template>
              </el-table-column>
              <el-table-column prop="size" label="大小" width="100" />
              <el-table-column prop="modified" label="修改时间" width="160" />
              <el-table-column label="操作" width="180">
                <template #default="{ row }">
                  <el-button size="small" text @click="downloadFile(row)">下载</el-button>
                  <el-button size="small" text type="danger" @click="deleteRemoteFile(row)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
            <div class="sftp-upload">
              <el-upload action="#" :auto-upload="false" :show-file-list="false" @change="uploadFile">
                <el-button size="small" type="primary">
                  <el-icon><Upload /></el-icon> 上传文件
                </el-button>
              </el-upload>
            </div>
          </div>
        </div>
      </div>
    </div>

    <el-dialog v-model="showHostDialog" title="添加/编辑 SSH 主机" width="550px">
      <el-form :model="editHost" label-width="100px">
        <el-form-item label="主机名称">
          <el-input v-model="editHost.name" placeholder="例如: 生产服务器" />
        </el-form-item>
        <el-form-item label="主机地址">
          <el-input v-model="editHost.host" placeholder="例: 192.168.1.100" />
        </el-form-item>
        <el-form-item label="端口">
          <el-input-number v-model="editHost.port" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="editHost.username" placeholder="root" />
        </el-form-item>
        <el-form-item label="认证方式">
          <el-radio-group v-model="editHost.authType">
            <el-radio label="password">密码</el-radio>
            <el-radio label="key">密钥</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="editHost.authType === 'password'" label="密码">
          <el-input v-model="editHost.password" type="password" show-password />
        </el-form-item>
        <el-form-item v-else label="私钥路径">
          <el-input v-model="editHost.privateKey" placeholder="/path/to/id_rsa" />
          <el-button @click="browsePrivateKey" style="margin-left: 10px">浏览</el-button>
        </el-form-item>
        <el-form-item label="分组">
          <el-select v-model="editHost.group" placeholder="选择分组" allow-create>
            <el-option v-for="g in ['默认', '生产环境', '测试环境', '游戏服务器']" :key="g" :label="g" :value="g" />
          </el-select>
        </el-form-item>
        <el-form-item label="标签">
          <el-select v-model="editHost.tags" multiple placeholder="选择标签" allow-create>
            <el-option label="Web" value="Web" />
            <el-option label="DB" value="DB" />
            <el-option label="Game" value="Game" />
            <el-option label="Proxy" value="Proxy" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showHostDialog = false">取消</el-button>
        <el-button type="primary" @click="saveHost">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showQuickConnect" title="快速 SSH 连接" width="450px">
      <el-form :model="quickConn" label-width="80px">
        <el-form-item label="主机">
          <el-input v-model="quickConn.host" placeholder="192.168.1.100:22" />
        </el-form-item>
        <el-form-item label="用户">
          <el-input v-model="quickConn.username" placeholder="root" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="quickConn.password" type="password" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showQuickConnect = false">取消</el-button>
        <el-button type="success" @click="quickConnect">连接</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

interface SSHHost {
  id: string
  name: string
  host: string
  port: number
  username: string
  authType: 'password' | 'key'
  password?: string
  privateKey?: string
  group: string
  tags: string[]
  connected: boolean
}

const hostSearch = ref('')
const showHostDialog = ref(false)
const showQuickConnect = ref(false)
const showSFTP = ref(false)
const activeHost = ref<SSHHost | null>(null)
const activeTerminalTab = ref('')
const terminalFontSize = ref(14)
const terminalTheme = ref('dark')
const sftpPath = ref('/home')
const terminalContainer = ref<HTMLElement>()
const terminalBody = ref<HTMLElement>()

const editHost = reactive<SSHHost>({
  id: '', name: '', host: '', port: 22, username: 'root',
  authType: 'password', password: '', privateKey: '', group: '默认', tags: []
})

const quickConn = reactive({ host: '', username: 'root', password: '' })

const hostGroups = ref([
  {
    name: '生产环境', expanded: true, hosts: [
      { id: '1', name: '主 Web 服务器', host: '192.168.1.100', port: 22, username: 'root', authType: 'password' as const, password: '***', group: '生产环境', tags: ['Web', 'Proxy'], connected: false },
      { id: '2', name: '数据库服务器', host: '192.168.1.101', port: 22, username: 'root', authType: 'password' as const, password: '***', group: '生产环境', tags: ['DB'], connected: false }
    ]
  },
  {
    name: '游戏服务器', expanded: true, hosts: [
      { id: '3', name: '七日杀服务器', host: '192.168.2.50', port: 22, username: 'steam', authType: 'key' as const, privateKey: '~/.ssh/id_rsa', group: '游戏服务器', tags: ['Game'], connected: false }
    ]
  },
  {
    name: '测试环境', expanded: false, hosts: [
      { id: '4', name: '测试服务器', host: '10.0.0.10', port: 2222, username: 'dev', authType: 'password' as const, password: '***', group: '测试环境', tags: ['Web'], connected: false }
    ]
  }
])

const terminalTabs = ref<{ id: string; label: string; host: SSHHost }[]>([])

const snippets = ref([
  { name: '系统信息', command: 'uname -a && cat /etc/os-release' },
  { name: '内存使用', command: 'free -h' },
  { name: '磁盘空间', command: 'df -h' },
  { name: 'CPU 信息', command: 'lscpu | head -20' },
  { name: 'Docker 状态', command: 'docker ps -a' },
  { name: '网络连接', command: 'ss -tuln' },
  { name: '进程列表', command: 'ps aux --sort=-%mem | head -20' },
  { name: 'Nginx 重启', command: 'systemctl restart nginx' },
  { name: '查看日志', command: 'tail -f /var/log/syslog' }
])

const sftpFiles = ref([
  { name: 'public_html', isDir: true, size: '-', modified: '2026-06-28 10:30' },
  { name: 'backup', isDir: true, size: '-', modified: '2026-06-25 08:00' },
  { name: 'config.json', isDir: false, size: '2.3KB', modified: '2026-06-29 14:22' },
  { name: 'app.log', isDir: false, size: '15.8KB', modified: '2026-06-29 15:45' }
])

function filterHosts(hosts: SSHHost[]) {
  const q = hostSearch.value.toLowerCase()
  return hosts.filter(h =>
    h.name.toLowerCase().includes(q) || h.host.includes(q)
  )
}

function toggleGroup(name: string) {
  const group = hostGroups.value.find(g => g.name === name)
  if (group) group.expanded = !group.expanded
}

function connectToHost(host: SSHHost) {
  host.connected = true
  activeHost.value = host
  const tabId = 'tab-' + host.id
  const existing = terminalTabs.value.find(t => t.id === tabId)
  if (!existing) {
    terminalTabs.value.push({ id: tabId, label: host.name, host })
  }
  activeTerminalTab.value = tabId
  ElMessage.success(`已连接到 ${host.name}`)
}

function closeTerminal(id: string) {
  const idx = terminalTabs.value.findIndex(t => t.id === id)
  if (idx !== -1) {
    const tab = terminalTabs.value[idx]
    tab.host.connected = false
    terminalTabs.value.splice(idx, 1)
    if (activeTerminalTab.value === id) {
      activeTerminalTab.value = terminalTabs.value[0]?.id || ''
    }
  }
}

function pasteToTerminal() {
  navigator.clipboard.readText().then(t => {
    ElMessage.success('已粘贴')
  })
}

function clearTerminal() { ElMessage.info('终端已清屏 (模拟)') }

function runSnippet(snip: any) {
  if (!activeHost.value?.connected) {
    ElMessage.warning('请先连接到主机')
    return
  }
  ElMessage.info(`执行: ${snip.command}`)
}

function showHostMenu(event: MouseEvent, host: SSHHost) {
  // Context menu for host items
}

function browsePrivateKey() { ElMessage.info('私钥浏览 (系统文件选择器)') }

function saveHost() {
  if (!editHost.name || !editHost.host) {
    ElMessage.warning('请填写主机名称和地址')
    return
  }
  const group = hostGroups.value.find(g => g.name === editHost.group)
  if (group) {
    group.hosts.push({ ...editHost, id: Date.now().toString(), connected: false })
  }
  showHostDialog.value = false
  ElMessage.success('主机已保存')
}

function quickConnect() {
  if (!quickConn.host) {
    ElMessage.warning('请输入主机地址')
    return
  }
  const [host, port] = quickConn.host.includes(':')
    ? quickConn.host.split(':')
    : [quickConn.host, '22']
  const newHost: SSHHost = {
    id: Date.now().toString(), name: `${quickConn.username}@${host}`,
    host, port: parseInt(port), username: quickConn.username,
    authType: 'password', password: quickConn.password, group: '默认', tags: [], connected: true
  }
  activeHost.value = newHost
  terminalTabs.value.push({ id: 'tab-' + newHost.id, label: newHost.name, host: newHost })
  activeTerminalTab.value = 'tab-' + newHost.id
  showQuickConnect.value = false
  ElMessage.success('连接成功')
}

function loadSFTPFiles() { ElMessage.info(`加载 ${sftpPath.value} 目录内容...`) }
function downloadFile(file: any) { ElMessage.info(`下载 ${file.name} (模拟)`) }
function deleteRemoteFile(file: any) {
  ElMessageBox.confirm(`确定要删除 "${file.name}" 吗？`, '确认删除', { type: 'warning' })
    .then(() => { ElMessage.success(`${file.name} 已删除`) }).catch(() => {})
}
function uploadFile() { ElMessage.info('文件上传 (模拟)') }
</script>

<style scoped>
.ssh-tool { padding: 0; height: calc(100vh - var(--header-height) - 60px); }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { font-size: 22px; font-weight: 600; }
.header-actions { display: flex; gap: 10px; }

.ssh-layout { display: flex; height: calc(100% - 56px); gap: 0; }
.ssh-sidebar { width: 260px; min-width: 260px; border-right: 1px solid var(--border-color); background: var(--bg-primary); display: flex; flex-direction: column; overflow-y: auto; border-radius: 8px 0 0 8px; }

.host-search { padding: 12px; border-bottom: 1px solid var(--border-color); }

.host-groups { flex: 1; overflow-y: auto; }
.host-group { border-bottom: 1px solid var(--border-color); }
.group-header { display: flex; align-items: center; gap: 6px; padding: 10px 12px; cursor: pointer; font-weight: 600; font-size: 13px; }
.group-header:hover { background: var(--bg-secondary); }
.group-count { margin-left: auto; font-size: 11px; background: var(--accent-color); color: white; padding: 1px 6px; border-radius: 10px; }

.host-item { display: flex; align-items: center; gap: 8px; padding: 8px 12px 8px 28px; cursor: pointer; transition: all 0.15s; font-size: 13px; }
.host-item:hover { background: var(--bg-secondary); }
.host-item.active { background: rgba(64,158,255,0.1); border-left: 3px solid var(--accent-color); }
.host-status { flex-shrink: 0; font-size: 10px; }
.host-info { flex: 1; min-width: 0; }
.host-name { font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.host-addr { font-size: 11px; color: var(--text-secondary); }
.host-tags { display: flex; gap: 2px; flex-wrap: wrap; }

.snippet-section { border-top: 1px solid var(--border-color); padding: 12px; }
.snippet-section h4 { font-size: 12px; color: var(--text-secondary); margin-bottom: 8px; }
.snippet-list { display: flex; flex-direction: column; gap: 4px; }
.snippet-item { padding: 6px 8px; border-radius: 4px; cursor: pointer; font-size: 12px; }
.snippet-item:hover { background: var(--bg-secondary); }
.snippet-name { display: block; font-weight: 500; }
.snippet-cmd { display: block; color: var(--text-secondary); font-family: 'Consolas', monospace; font-size: 11px; }

.ssh-main { flex: 1; display: flex; flex-direction: column; min-width: 0; overflow: hidden; }

.terminal-container { display: flex; flex-direction: column; height: 100%; background: #0d0d0d; border-radius: 8px; overflow: hidden; }
.terminal-toolbar { display: flex; justify-content: space-between; align-items: center; padding: 8px 12px; background: #1a1a1a; border-bottom: 1px solid #333; }
.toolbar-left { display: flex; gap: 6px; }
.toolbar-right { color: #c0c4cc; font-size: 12px; }
.terminal-body { flex: 1; padding: 12px; font-family: 'Consolas', 'Courier New', monospace; font-size: 14px; color: #00ff00; overflow-y: auto; white-space: pre-wrap; min-height: 300px; }

.terminal-body::before {
  content: "Welcome to OmniPanel SSH Terminal\nuser@omnipanel:~$ _\n\n已连接到: " attr(data-host) "\n就绪。";
  display: block;
}

.no-connection { flex: 1; display: flex; align-items: center; justify-content: center; }
.no-connection-content { text-align: center; color: var(--text-secondary); }
.no-connection-content h3 { margin: 16px 0 8px; }
.no-connection-content p { margin-bottom: 20px; }
.no-connection-actions { display: flex; gap: 12px; justify-content: center; }

.sftp-panel { border-top: 2px solid var(--accent-color); margin-top: 8px; }
.sftp-header { display: flex; justify-content: space-between; align-items: center; padding: 8px 12px; background: var(--bg-primary); border-radius: 8px 8px 0 0; }
.sftp-body { padding: 12px; background: var(--bg-primary); }
.sftp-pathbar { display: flex; gap: 8px; margin-bottom: 12px; }
.sftp-upload { margin-top: 12px; }
</style>
