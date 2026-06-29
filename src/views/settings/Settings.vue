<template>
  <div class="settings-page">
    <div class="page-header">
      <h2>系统设置</h2>
    </div>

    <el-tabs v-model="activeTab" tab-position="left">
      <el-tab-pane label="主题外观" name="appearance">
        <div class="settings-section">
          <h3>主题外观</h3>
          <div class="setting-item">
            <div>
              <div class="setting-label">主题模式</div>
              <div class="setting-desc">选择亮色、暗色或跟随系统主题</div>
            </div>
            <el-radio-group v-model="settings.theme" @change="settings.setTheme">
              <el-radio-button label="light">亮色</el-radio-button>
              <el-radio-button label="dark">暗色</el-radio-button>
              <el-radio-button label="auto">跟随系统</el-radio-button>
            </el-radio-group>
          </div>

          <div class="setting-item">
            <div>
              <div class="setting-label">侧边栏</div>
              <div class="setting-desc">控制侧边栏默认展开/折叠</div>
            </div>
            <el-switch v-model="settings.sidebarCollapsed" active-text="折叠" inactive-text="展开" />
          </div>

          <div class="setting-item">
            <div>
              <div class="setting-label">语言设置</div>
              <div class="setting-desc">界面显示语言</div>
            </div>
            <el-select v-model="settings.language" size="small">
              <el-option label="简体中文" value="zh-CN" />
              <el-option label="繁體中文" value="zh-TW" />
              <el-option label="English" value="en" />
            </el-select>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="安全设置" name="security">
        <div class="settings-section">
          <h3>安全设置</h3>
          <div class="setting-item">
            <div>
              <div class="setting-label">主密码</div>
              <div class="setting-desc">保护敏感数据 (API Key、SSH 密码等) 加密存储</div>
            </div>
            <el-button size="small" @click="showMasterPassword = true">设置主密码</el-button>
          </div>

          <div class="setting-item">
            <div>
              <div class="setting-label">应用锁</div>
              <div class="setting-desc">启动应用时需要输入密码</div>
            </div>
            <el-switch v-model="appLock" />
          </div>

          <div class="setting-item">
            <div>
              <div class="setting-label">二次验证</div>
              <div class="setting-desc">关键操作需要二次确认</div>
            </div>
            <el-switch v-model="doubleConfirm" active-value />
          </div>

          <div class="setting-item">
            <div>
              <div class="setting-label">加密导出</div>
              <div class="setting-desc">导出配置时使用密码加密</div>
            </div>
            <el-button size="small" @click="exportEncrypted">加密导出配置</el-button>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="本地详情" name="system">
        <div class="settings-section">
          <h3>本地机器详情</h3>

          <div class="system-info-grid">
            <div class="sys-info-card">
              <h4>系统信息</h4>
              <div class="sys-info-list">
                <div class="sys-info-item"><span>操作系统</span><span>{{ sysInfo.os }}</span></div>
                <div class="sys-info-item"><span>版本</span><span>{{ sysInfo.version }}</span></div>
                <div class="sys-info-item"><span>主机名</span><span>{{ sysInfo.hostname }}</span></div>
                <div class="sys-info-item"><span>架构</span><span>{{ sysInfo.arch }}</span></div>
                <div class="sys-info-item"><span>运行时间</span><span>{{ sysInfo.uptime }}</span></div>
              </div>
            </div>

            <div class="sys-info-card">
              <h4>CPU 信息</h4>
              <div class="sys-info-list">
                <div class="sys-info-item"><span>型号</span><span>{{ sysInfo.cpuModel }}</span></div>
                <div class="sys-info-item"><span>核心数</span><span>{{ sysInfo.cpuCores }}</span></div>
                <div class="sys-info-item"><span>频率</span><span>{{ sysInfo.cpuSpeed }}</span></div>
                <div class="sys-info-item"><span>使用率</span><span>{{ sysInfo.cpuUsage }}%</span></div>
              </div>
            </div>

            <div class="sys-info-card">
              <h4>内存信息</h4>
              <div class="sys-info-list">
                <div class="sys-info-item"><span>总内存</span><span>{{ sysInfo.totalMemory }}</span></div>
                <div class="sys-info-item"><span>已用</span><span>{{ sysInfo.usedMemory }}</span></div>
                <div class="sys-info-item"><span>可用</span><span>{{ sysInfo.freeMemory }}</span></div>
                <div class="sys-info-item"><span>使用率</span><span>{{ sysInfo.memUsage }}%</span></div>
              </div>
            </div>

            <div class="sys-info-card">
              <h4>磁盘信息</h4>
              <div class="sys-info-list">
                <div class="sys-info-item"><span>总容量</span><span>{{ sysInfo.diskTotal }}</span></div>
                <div class="sys-info-item"><span>已用</span><span>{{ sysInfo.diskUsed }}</span></div>
                <div class="sys-info-item"><span>可用</span><span>{{ sysInfo.diskFree }}</span></div>
                <div class="sys-info-item"><span>使用率</span><span>{{ sysInfo.diskUsage }}%</span></div>
              </div>
            </div>

            <div class="sys-info-card">
              <h4>网络信息</h4>
              <div class="sys-info-list">
                <div class="sys-info-item"><span>内网 IP</span><span>{{ sysInfo.internalIP }}</span></div>
                <div class="sys-info-item"><span>外网 IP</span><span>{{ sysInfo.externalIP }}</span></div>
                <div class="sys-info-item"><span>MAC 地址</span><span>{{ sysInfo.mac }}</span></div>
              </div>
            </div>
          </div>

          <div class="sub-section" style="margin-top: 24px">
            <h4>进程管理</h4>
            <el-table :data="processList" size="small" max-height="300">
              <el-table-column prop="pid" label="PID" width="80" />
              <el-table-column prop="name" label="进程名" min-width="200" />
              <el-table-column prop="cpu" label="CPU %" width="80" />
              <el-table-column prop="mem" label="内存 %" width="80" />
              <el-table-column label="操作" width="80">
                <template #default="{ row }">
                  <el-button size="small" type="danger" text @click="killProcess(row)">结束</el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>

          <div class="sub-section" style="margin-top: 24px">
            <h4>环境变量</h4>
            <el-table :data="envVars" size="small" max-height="200">
              <el-table-column prop="key" label="变量名" min-width="200" />
              <el-table-column prop="value" label="值" min-width="300" />
            </el-table>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="通知设置" name="notifications">
        <div class="settings-section">
          <h3>通知设置</h3>
          <div class="setting-item">
            <div>
              <div class="setting-label">系统通知</div>
              <div class="setting-desc">允许应用发送系统通知</div>
            </div>
            <el-switch v-model="settings.notifications" />
          </div>
          <div class="setting-item">
            <div>
              <div class="setting-label">声音提示</div>
              <div class="setting-desc">关键事件时播放声音</div>
            </div>
            <el-switch v-model="soundEnabled" />
          </div>
          <div class="setting-item">
            <div>
              <div class="setting-label">邮件通知</div>
              <div class="setting-desc">服务器告警通过邮件通知</div>
            </div>
            <el-switch v-model="emailNotify" />
          </div>
          <div class="setting-item" v-if="emailNotify">
            <div>
              <div class="setting-label">邮箱地址</div>
            </div>
            <el-input v-model="emailAddress" placeholder="your@email.com" size="small" style="width: 300px" />
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="代理设置" name="proxy">
        <div class="settings-section">
          <h3>网络代理</h3>
          <div class="setting-item">
            <div>
              <div class="setting-label">启用代理</div>
            </div>
            <el-switch v-model="proxyEnabled" />
          </div>
          <template v-if="proxyEnabled">
            <div class="setting-item">
              <div class="setting-label">代理类型</div>
              <el-select v-model="proxyType" size="small">
                <el-option label="HTTP" value="http" />
                <el-option label="HTTPS" value="https" />
                <el-option label="SOCKS5" value="socks5" />
              </el-select>
            </div>
            <div class="setting-item">
              <div class="setting-label">代理地址</div>
              <el-input v-model="proxyHost" placeholder="127.0.0.1" size="small" style="width: 200px" />
            </div>
            <div class="setting-item">
              <div class="setting-label">代理端口</div>
              <el-input-number v-model="proxyPort" :min="1" :max="65535" size="small" />
            </div>
          </template>
        </div>
      </el-tab-pane>

      <el-tab-pane label="数据管理" name="data">
        <div class="settings-section">
          <h3>数据备份与恢复</h3>
          <div class="setting-item">
            <div>
              <div class="setting-label">导出所有配置</div>
              <div class="setting-desc">导出配置文件、SSH 连接、API Key 等所有数据</div>
            </div>
            <el-button size="small" @click="exportData">导出</el-button>
          </div>
          <div class="setting-item">
            <div>
              <div class="setting-label">导入配置</div>
              <div class="setting-desc">从备份恢复所有配置数据</div>
            </div>
            <el-button size="small" @click="importData">导入</el-button>
          </div>
          <div class="setting-item">
            <div>
              <div class="setting-label">重置所有设置</div>
              <div class="setting-desc">恢复到出厂默认设置</div>
            </div>
            <el-button size="small" type="danger" @click="resetAll">重置</el-button>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="快捷键" name="shortcuts">
        <div class="settings-section">
          <h3>快捷键自定义</h3>
          <div class="shortcut-list">
            <div class="shortcut-item" v-for="sc in shortcuts" :key="sc.name">
              <span class="shortcut-name">{{ sc.name }}</span>
              <el-input v-model="sc.key" size="small" style="width: 160px" readonly>
                <template #append>
                  <el-button size="small" @click="editShortcut(sc)">编辑</el-button>
                </template>
              </el-input>
            </div>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="关于" name="about">
        <div class="settings-section">
          <div class="about-content">
            <div class="about-logo">⚡ OmniPanel</div>
            <div class="about-version">版本 1.0.0</div>
            <div class="about-desc">一体化服务器&游戏管理桌面应用</div>
            <div class="about-info">
              <div class="about-item"><span>技术栈</span><span>Electron 30 + Vue 3 + TypeScript + Tailwind CSS</span></div>
              <div class="about-item"><span>后端</span><span>Node.js 20+ / Express</span></div>
              <div class="about-item"><span>数据库</span><span>SQLite (better-sqlite3)</span></div>
              <div class="about-item"><span>许可</span><span>MIT License</span></div>
            </div>
            <el-button size="small" @click="checkUpdate">检查更新</el-button>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="showMasterPassword" title="设置主密码" width="400px">
      <el-form>
        <el-form-item label="新密码">
          <el-input v-model="masterPassword" type="password" show-password />
        </el-form-item>
        <el-form-item label="确认密码">
          <el-input v-model="masterPasswordConfirm" type="password" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showMasterPassword = false">取消</el-button>
        <el-button type="primary" @click="setMasterPassword">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { useSettingsStore } from '@/stores/settings'

const settings = useSettingsStore()
const activeTab = ref('appearance')
const showMasterPassword = ref(false)
const masterPassword = ref('')
const masterPasswordConfirm = ref('')
const appLock = ref(false)
const doubleConfirm = ref(true)
const soundEnabled = ref(true)
const emailNotify = ref(false)
const emailAddress = ref('')
const proxyEnabled = ref(false)
const proxyType = ref('http')
const proxyHost = ref('127.0.0.1')
const proxyPort = ref(1080)

const sysInfo = reactive({
  os: 'Linux / Ubuntu 22.04',
  version: '22.04 LTS',
  hostname: 'omnipanel-dev',
  arch: 'x86_64',
  uptime: '30天 12小时 35分钟',
  cpuModel: 'Intel Core i7-13700K @ 5.4GHz',
  cpuCores: '16核 (8P + 8E)',
  cpuSpeed: '3400 MHz',
  cpuUsage: 23,
  totalMemory: '32.0 GB',
  usedMemory: '14.4 GB',
  freeMemory: '17.6 GB',
  memUsage: 45,
  diskTotal: '512 GB SSD',
  diskUsed: '164 GB',
  diskFree: '348 GB',
  diskUsage: 32,
  internalIP: '192.168.1.100',
  externalIP: '123.45.67.89',
  mac: 'AA:BB:CC:DD:EE:FF'
})

const processList = ref([
  { pid: 1234, name: 'node', cpu: '2.3', mem: '4.5' },
  { pid: 5678, name: 'chrome', cpu: '8.7', mem: '12.3' },
  { pid: 9012, name: 'docker', cpu: '1.2', mem: '3.1' },
  { pid: 3456, name: 'postgres', cpu: '3.4', mem: '8.2' }
])

const envVars = ref([
  { key: 'PATH', value: '/usr/local/bin:/usr/bin:/bin' },
  { key: 'HOME', value: '/home/user' },
  { key: 'SHELL', value: '/bin/bash' },
  { key: 'LANG', value: 'zh_CN.UTF-8' },
  { key: 'NODE_ENV', value: 'development' }
])

const shortcuts = ref([
  { name: '切换主题', key: 'Ctrl+Shift+T' },
  { name: '打开设置', key: 'Ctrl+,' },
  { name: '打开终端', key: 'Ctrl+`' },
  { name: '快速搜索', key: 'Ctrl+P' },
  { name: '全屏切换', key: 'F11' }
])

function setMasterPassword() {
  if (masterPassword.value !== masterPasswordConfirm.value) {
    ElMessage.error('两次密码不一致')
    return
  }
  settings.masterPassword = masterPassword.value
  showMasterPassword.value = false
  ElMessage.success('主密码已设置')
}

function exportData() { ElMessage.success('配置数据已导出') }
function importData() { ElMessage.success('配置数据已导入') }
function resetAll() { ElMessage.success('所有设置已重置') }
function checkUpdate() { ElMessage.info('当前已是最新版本 v1.0.0') }
function exportEncrypted() { ElMessage.success('加密配置已导出') }
function killProcess(row: any) { ElMessage.success(`进程 ${row.pid} 已终止`) }
function editShortcut(sc: any) { ElMessage.info(`编辑快捷键: ${sc.name}`) }
</script>

<style scoped>
.settings-page { padding: 0; height: calc(100vh - var(--header-height) - 60px); }
.page-header { margin-bottom: 16px; }
.page-header h2 { font-size: 22px; font-weight: 600; }

.settings-section { padding: 0 20px; }
.settings-section h3 { font-size: 18px; margin-bottom: 20px; padding-bottom: 10px; border-bottom: 1px solid var(--border-color); }

.setting-item { display: flex; justify-content: space-between; align-items: center; padding: 14px 0; border-bottom: 1px solid var(--border-color); }
.setting-label { font-weight: 500; margin-bottom: 4px; }
.setting-desc { font-size: 12px; color: var(--text-secondary); }

.system-info-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 16px; }
.sys-info-card { background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 10px; padding: 16px; }
.sys-info-card h4 { margin-bottom: 12px; font-size: 14px; }
.sys-info-list { display: flex; flex-direction: column; gap: 8px; }
.sys-info-item { display: flex; justify-content: space-between; font-size: 13px; }
.sys-info-item span:first-child { color: var(--text-secondary); }

.sub-section { margin-top: 20px; }
.sub-section h4 { margin-bottom: 10px; }

.shortcut-list { display: flex; flex-direction: column; gap: 12px; max-width: 500px; }
.shortcut-item { display: flex; justify-content: space-between; align-items: center; }
.shortcut-name { font-weight: 500; }

.about-content { text-align: center; padding: 40px 0; }
.about-logo { font-size: 36px; font-weight: bold; margin-bottom: 8px; }
.about-version { font-size: 16px; color: var(--accent-color); margin-bottom: 8px; }
.about-desc { color: var(--text-secondary); margin-bottom: 30px; }
.about-info { max-width: 400px; margin: 0 auto 20px; }
.about-item { display: flex; justify-content: space-between; padding: 8px 0; font-size: 13px; border-bottom: 1px solid var(--border-color); }
.about-item span:first-child { color: var(--text-secondary); }
</style>
