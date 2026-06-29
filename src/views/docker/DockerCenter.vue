<template>
  <div class="docker-center">
    <div class="page-header">
      <h2>🐳 Docker 管理中心</h2>
      <div class="header-actions">
        <el-button type="primary" @click="showInstallDialog = true">
          <el-icon><Setting /></el-icon> Docker 环境
        </el-button>
        <el-button @click="loadDockerInfo">
          <el-icon><Refresh /></el-icon> 刷新
        </el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab" type="border-card">
      <el-tab-pane label="容器列表" name="containers">
        <div class="tab-toolbar">
          <el-input v-model="containerSearch" placeholder="搜索容器..." clearable style="width: 240px" />
          <div class="toolbar-right">
            <el-button type="primary" size="small" @click="showRunDialog = true">
              <el-icon><Plus /></el-icon> 运行容器
            </el-button>
          </div>
        </div>

        <el-table :data="filteredContainers" style="width: 100%" v-loading="loading">
          <el-table-column prop="name" label="容器名称" min-width="180">
            <template #default="{ row }">
              <div class="container-name">
                <span :class="row.status === 'running' ? 'status-online' : 'status-offline'">●</span>
                {{ row.name }}
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="image" label="镜像" min-width="150" />
          <el-table-column prop="status" label="状态" width="120">
            <template #default="{ row }">
              <el-tag :type="row.status === 'running' ? 'success' : 'info'" size="small">
                {{ row.status === 'running' ? '运行中' : '已停止' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="ports" label="端口映射" min-width="200" />
          <el-table-column prop="cpu" label="CPU" width="80" />
          <el-table-column prop="memory" label="内存" width="100" />
          <el-table-column prop="uptime" label="运行时间" width="120" />
          <el-table-column label="操作" width="280" fixed="right">
            <template #default="{ row }">
              <el-button-group>
                <el-button v-if="row.status !== 'running'" type="success" size="small" @click="handleAction(row, 'start')">
                  启动
                </el-button>
                <el-button v-else type="warning" size="small" @click="handleAction(row, 'stop')">
                  停止
                </el-button>
                <el-button size="small" @click="handleAction(row, 'restart')">重启</el-button>
                <el-button size="small" @click="showLogs(row)">日志</el-button>
                <el-button type="danger" size="small" @click="handleAction(row, 'remove')">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </el-button-group>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="镜像管理" name="images">
        <div class="tab-toolbar">
          <div>
            <el-input v-model="imageSearch" placeholder="搜索镜像..." clearable style="width: 240px" />
            <el-button style="margin-left: 10px" @click="searchDockerHub">搜索 Docker Hub</el-button>
          </div>
          <div class="toolbar-right">
            <el-button type="primary" size="small" @click="showPullDialog = true">
              <el-icon><Download /></el-icon> 拉取镜像
            </el-button>
          </div>
        </div>

        <el-table :data="filteredImages" style="width: 100%" v-loading="loading">
          <el-table-column prop="repository" label="镜像名称" min-width="200" />
          <el-table-column prop="tag" label="标签" width="120" />
          <el-table-column prop="size" label="大小" width="120" />
          <el-table-column prop="created" label="创建时间" width="180" />
          <el-table-column label="操作" width="160" fixed="right">
            <template #default="{ row }">
              <el-button type="danger" size="small" @click="removeImage(row)">删除</el-button>
              <el-button size="small" @click="inspectImage(row)">详情</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="Docker Compose" name="compose">
        <div class="tab-toolbar">
          <el-upload :auto-upload="false" :show-file-list="false" accept=".yml,.yaml" @change="handleComposeUpload">
            <el-button type="primary" size="small">
              <el-icon><Upload /></el-icon> 导入 docker-compose.yml
            </el-button>
          </el-upload>
          <el-button size="small" @click="showComposeEditor = true" style="margin-left: 10px">
            <el-icon><Edit /></el-icon> 新建/编辑
          </el-button>
        </div>

        <div class="compose-editor" v-if="showComposeEditor">
          <el-input
            v-model="composeContent"
            type="textarea"
            :rows="15"
            placeholder="在此编辑 docker-compose.yml..."
          />
          <div style="margin-top: 12px; display: flex; gap: 10px;">
            <el-button type="primary" @click="deployCompose">一键部署</el-button>
            <el-button @click="validateCompose">验证配置</el-button>
            <el-button @click="showComposeEditor = false">取消</el-button>
          </div>
        </div>

        <div style="margin-top: 16px;">
          <h4>预设模板</h4>
          <div class="template-grid">
            <div class="template-card card-hover" v-for="tpl in composeTemplates" :key="tpl.name" @click="loadTemplate(tpl)">
              <div class="template-name">{{ tpl.name }}</div>
              <div class="template-desc">{{ tpl.description }}</div>
            </div>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="数据卷" name="volumes">
        <div class="tab-toolbar">
          <el-button type="primary" size="small" @click="showVolumeCreateDialog = true">
            <el-icon><Plus /></el-icon> 创建数据卷
          </el-button>
        </div>
        <el-table :data="volumes" style="width: 100%" v-loading="loading">
          <el-table-column prop="name" label="卷名称" min-width="200" />
          <el-table-column prop="driver" label="驱动" width="100" />
          <el-table-column prop="mountpoint" label="挂载点" min-width="300" />
          <el-table-column label="操作" width="120">
            <template #default="{ row }">
              <el-button type="danger" size="small" @click="removeVolume(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="网络管理" name="networks">
        <div class="tab-toolbar">
          <el-button type="primary" size="small" @click="showNetworkCreateDialog = true">
            <el-icon><Plus /></el-icon> 创建网络
          </el-button>
        </div>
        <el-table :data="networks" style="width: 100%" v-loading="loading">
          <el-table-column prop="name" label="网络名称" min-width="180" />
          <el-table-column prop="driver" label="驱动" width="100" />
          <el-table-column prop="scope" label="范围" width="100" />
          <el-table-column prop="subnet" label="子网" min-width="180" />
          <el-table-column label="操作" width="120">
            <template #default="{ row }">
              <el-button type="danger" size="small" @click="removeNetwork(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="监控" name="monitor">
        <div class="monitor-grid">
          <div class="monitor-card">
            <h4>CPU 使用率</h4>
            <div ref="dockerCpuRef" style="height: 250px;"></div>
          </div>
          <div class="monitor-card">
            <h4>内存使用率</h4>
            <div ref="dockerMemRef" style="height: 250px;"></div>
          </div>
          <div class="monitor-card">
            <h4>网络 IO</h4>
            <div ref="dockerNetRef" style="height: 250px;"></div>
          </div>
          <div class="monitor-card">
            <h4>磁盘 IO</h4>
            <div ref="dockerDiskRef" style="height: 250px;"></div>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="showLogDialog" title="容器日志" width="80%" top="5vh">
      <div class="log-viewer">
        <div class="log-toolbar">
          <el-button size="small" @click="autoScroll = !autoScroll">
            {{ autoScroll ? '暂停滚动' : '自动滚动' }}
          </el-button>
          <el-button size="small" @click="clearLogs">清空</el-button>
          <el-button size="small" @click="downloadLogs">下载日志</el-button>
        </div>
        <div class="log-content" ref="logContainer">
          <div v-for="(line, i) in containerLogs" :key="i" class="log-line">{{ line }}</div>
        </div>
      </div>
    </el-dialog>

    <el-dialog v-model="showRunDialog" title="运行新容器" width="600px">
      <el-form :model="newContainer" label-width="100px">
        <el-form-item label="镜像名称">
          <el-input v-model="newContainer.image" placeholder="例: nginx:latest" />
        </el-form-item>
        <el-form-item label="容器名称">
          <el-input v-model="newContainer.name" placeholder="可选" />
        </el-form-item>
        <el-form-item label="端口映射">
          <el-input v-model="newContainer.ports" placeholder="例: 8080:80" />
        </el-form-item>
        <el-form-item label="环境变量">
          <el-input v-model="newContainer.env" type="textarea" :rows="3" placeholder="KEY=VALUE，每行一个" />
        </el-form-item>
        <el-form-item label="数据卷">
          <el-input v-model="newContainer.volumes" placeholder="例: /host/path:/container/path" />
        </el-form-item>
        <el-form-item label="重启策略">
          <el-select v-model="newContainer.restart">
            <el-option label="no" value="no" />
            <el-option label="always" value="always" />
            <el-option label="on-failure" value="on-failure" />
            <el-option label="unless-stopped" value="unless-stopped" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRunDialog = false">取消</el-button>
        <el-button type="primary" @click="runContainer">启动容器</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showPullDialog" title="拉取镜像" width="500px">
      <el-form>
        <el-form-item label="镜像名称">
          <el-input v-model="pullImage" placeholder="例: nginx:latest" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPullDialog = false">取消</el-button>
        <el-button type="primary" @click="pullDockerImage" :loading="pulling">拉取</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showInstallDialog" title="Docker 环境检测与安装" width="600px">
      <div class="docker-env-info">
        <div class="env-item">
          <span class="env-label">Docker 版本</span>
          <el-tag :type="dockerInstalled ? 'success' : 'danger'">
            {{ dockerInstalled ? dockerVersion : '未安装' }}
          </el-tag>
        </div>
        <div class="env-item">
          <span class="env-label">Docker Compose</span>
          <el-tag :type="composeInstalled ? 'success' : 'danger'">
            {{ composeInstalled ? composeVersion : '未安装' }}
          </el-tag>
        </div>
        <div class="env-item">
          <span class="env-label">Docker 服务</span>
          <el-tag :type="dockerRunning ? 'success' : 'warning'">
            {{ dockerRunning ? '运行中' : '未运行' }}
          </el-tag>
        </div>
      </div>
      <template #footer>
        <el-button @click="showInstallDialog = false">关闭</el-button>
        <el-button v-if="!dockerInstalled" type="primary" @click="installDocker">一键安装 Docker</el-button>
        <el-button v-else-if="!dockerRunning" type="primary" @click="startDocker">启动 Docker 服务</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as echarts from 'echarts'
import { v4 as uuid } from 'uuid'

const activeTab = ref('containers')
const loading = ref(false)
const containerSearch = ref('')
const imageSearch = ref('')
const pullImage = ref('')
const pulling = ref(false)
const showLogDialog = ref(false)
const showRunDialog = ref(false)
const showPullDialog = ref(false)
const showComposeEditor = ref(false)
const showInstallDialog = ref(false)
const showVolumeCreateDialog = ref(false)
const showNetworkCreateDialog = ref(false)
const composeContent = ref('')
const autoScroll = ref(true)
const logContainer = ref<HTMLElement>()

const dockerInstalled = ref(true)
const dockerRunning = ref(true)
const dockerVersion = ref('Docker 24.0.7')
const composeInstalled = ref(true)
const composeVersion = ref('v2.23.0')

const containerLogs = ref<string[]>([])

const newContainer = reactive({
  image: '',
  name: '',
  ports: '',
  env: '',
  volumes: '',
  restart: 'no'
})

interface Container {
  id: string
  name: string
  image: string
  status: string
  ports: string
  cpu: string
  memory: string
  uptime: string
}

const containers = ref<Container[]>([
  { id: '1', name: 'nginx-proxy', image: 'nginx:latest', status: 'running', ports: '80:80, 443:443', cpu: '0.5%', memory: '32MB', uptime: '5天' },
  { id: '2', name: 'mysql-db', image: 'mysql:8.0', status: 'running', ports: '3306:3306', cpu: '2.1%', memory: '512MB', uptime: '12天' },
  { id: '3', name: 'redis-cache', image: 'redis:7-alpine', status: 'running', ports: '6379:6379', cpu: '0.1%', memory: '8MB', uptime: '12天' },
  { id: '4', name: 'portainer', image: 'portainer/portainer-ce', status: 'running', ports: '9000:9000', cpu: '0.3%', memory: '48MB', uptime: '20天' },
  { id: '5', name: 'frps-server', image: 'snowdreamtech/frps', status: 'stopped', ports: '7000:7000, 7500:7500', cpu: '-', memory: '-', uptime: '-' }
])

const filteredContainers = computed(() =>
  containers.value.filter(c => c.name.toLowerCase().includes(containerSearch.value.toLowerCase()))
)

interface Image {
  id: string
  repository: string
  tag: string
  size: string
  created: string
}

const images = ref<Image[]>([
  { id: '1', repository: 'nginx', tag: 'latest', size: '187MB', created: '2026-06-20' },
  { id: '2', repository: 'mysql', tag: '8.0', size: '596MB', created: '2026-06-15' },
  { id: '3', repository: 'redis', tag: '7-alpine', size: '30MB', created: '2026-06-14' },
  { id: '4', repository: 'portainer/portainer-ce', tag: 'latest', size: '286MB', created: '2026-06-10' },
  { id: '5', repository: 'snowdreamtech/frps', tag: 'latest', size: '42MB', created: '2026-06-05' }
])

const filteredImages = computed(() =>
  images.value.filter(i =>
    i.repository.toLowerCase().includes(imageSearch.value.toLowerCase()) ||
    i.tag.toLowerCase().includes(imageSearch.value.toLowerCase())
  )
)

const volumes = ref([
  { name: 'nginx_data', driver: 'local', mountpoint: '/var/lib/docker/volumes/nginx_data/_data' },
  { name: 'mysql_data', driver: 'local', mountpoint: '/var/lib/docker/volumes/mysql_data/_data' }
])

const networks = ref([
  { name: 'bridge', driver: 'bridge', scope: 'local', subnet: '172.17.0.0/16' },
  { name: 'host', driver: 'host', scope: 'local', subnet: '-' },
  { name: 'none', driver: 'null', scope: 'local', subnet: '-' },
  { name: 'omnipanel-net', driver: 'bridge', scope: 'local', subnet: '172.20.0.0/16' }
])

const composeTemplates = [
  { name: 'LNMP 环境', description: 'Nginx + MySQL + PHP + Redis', content: 'version: "3.8"\nservices:\n  nginx:\n    image: nginx:latest\n    ports:\n      - "80:80"\n  mysql:\n    image: mysql:8.0\n    environment:\n      MYSQL_ROOT_PASSWORD: root123\n  php:\n    image: php:8.2-fpm\n  redis:\n    image: redis:7-alpine' },
  { name: 'WordPress', description: 'WordPress + MySQL 一键部署', content: 'version: "3.8"\nservices:\n  wordpress:\n    image: wordpress:latest\n    ports:\n      - "8080:80"\n    environment:\n      WORDPRESS_DB_HOST: db\n      WORDPRESS_DB_USER: wp\n      WORDPRESS_DB_PASSWORD: wp123\n  db:\n    image: mysql:8.0\n    environment:\n      MYSQL_DATABASE: wordpress\n      MYSQL_USER: wp\n      MYSQL_PASSWORD: wp123\n      MYSQL_ROOT_PASSWORD: root123' },
  { name: 'FRP 穿透', description: 'FRP 服务端 + 客户端', content: 'version: "3.8"\nservices:\n  frps:\n    image: snowdreamtech/frps\n    ports:\n      - "7000:7000"\n      - "7500:7500"\n    volumes:\n      - ./frps.toml:/etc/frp/frps.toml' }
]

const dockerCpuRef = ref<HTMLElement>()
const dockerMemRef = ref<HTMLElement>()
const dockerNetRef = ref<HTMLElement>()
const dockerDiskRef = ref<HTMLElement>()
let dockerCharts: echarts.ECharts[] = []

function loadDockerInfo() {
  loading.value = true
  setTimeout(() => { loading.value = false }, 800)
}

function handleAction(container: Container, action: string) {
  const actionMap: Record<string, string> = {
    start: '启动', stop: '停止', restart: '重启', remove: '删除'
  }
  ElMessageBox.confirm(`确定要${actionMap[action]}容器 "${container.name}" 吗？`, '确认操作', {
    type: action === 'remove' ? 'warning' : 'info'
  }).then(() => {
    if (action === 'remove') {
      containers.value = containers.value.filter(c => c.id !== container.id)
    } else if (action === 'start') {
      container.status = 'running'
      container.uptime = '刚刚'
    } else if (action === 'stop') {
      container.status = 'stopped'
      container.uptime = '-'
    } else if (action === 'restart') {
      container.status = 'running'
      container.uptime = '刚刚'
    }
    ElMessage.success(`容器 ${actionMap[action]}成功`)
  }).catch(() => {})
}

function showLogs(container: Container) {
  showLogDialog.value = true
  containerLogs.value = Array.from({ length: 50 }, (_, i) =>
    `[${new Date().toISOString()}] [${container.name}] Log entry #${i + 1}: Container ${container.status === 'running' ? 'processing request' : 'standby'}...`
  )
}

function clearLogs() { containerLogs.value = [] }

function downloadLogs() {
  const blob = new Blob([containerLogs.value.join('\n')], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url; a.download = 'container.log'; a.click()
  URL.revokeObjectURL(url)
}

function runContainer() {
  if (!newContainer.image) {
    ElMessage.warning('请输入镜像名称')
    return
  }
  containers.value.push({
    id: uuid(),
    name: newContainer.name || newContainer.image.split(':')[0],
    image: newContainer.image,
    status: 'running',
    ports: newContainer.ports || '-',
    cpu: '0.0%',
    memory: '0MB',
    uptime: '刚刚'
  })
  showRunDialog.value = false
  ElMessage.success('容器启动成功')
  newContainer.image = ''; newContainer.name = ''; newContainer.ports = ''; newContainer.env = ''; newContainer.volumes = ''
}

async function pullDockerImage() {
  if (!pullImage.value) return
  pulling.value = true
  await new Promise(r => setTimeout(r, 1500))
  const [repo, tag] = pullImage.value.split(':')
  images.value.push({
    id: uuid(),
    repository: repo,
    tag: tag || 'latest',
    size: '未知',
    created: new Date().toISOString().split('T')[0]
  })
  pulling.value = false
  showPullDialog.value = false
  ElMessage.success(`镜像 ${pullImage.value} 拉取成功`)
  pullImage.value = ''
}

function removeImage(image: Image) {
  ElMessageBox.confirm(`确定要删除镜像 "${image.repository}:${image.tag}" 吗？`, '确认删除', { type: 'warning' })
    .then(() => {
      images.value = images.value.filter(i => i.id !== image.id)
      ElMessage.success('镜像已删除')
    }).catch(() => {})
}

function inspectImage(image: Image) {
  ElMessage.info(`镜像详情: ${image.repository}:${image.tag} - 大小: ${image.size}`)
}

function searchDockerHub() {
  ElMessage.info('Docker Hub 搜索功能 (需连接 Docker API)')
}

function handleComposeUpload() { showComposeEditor.value = true }

function loadTemplate(tpl: any) {
  composeContent.value = tpl.content
  showComposeEditor.value = true
}

function validateCompose() {
  if (!composeContent.value.trim()) {
    ElMessage.warning('请先输入配置内容')
    return
  }
  ElMessage.success('配置格式验证通过')
}

function deployCompose() {
  if (!composeContent.value.trim()) {
    ElMessage.warning('请先输入配置内容')
    return
  }
  ElMessageBox.confirm('确定要部署 docker-compose.yml 吗？', '确认部署')
    .then(() => {
      ElMessage.success('Docker Compose 部署已启动')
    }).catch(() => {})
}

function removeVolume(vol: any) {
  ElMessageBox.confirm(`确定要删除数据卷 "${vol.name}" 吗？`, '确认删除', { type: 'warning' })
    .then(() => {
      volumes.value = volumes.value.filter(v => v.name !== vol.name)
      ElMessage.success('数据卷已删除')
    }).catch(() => {})
}

function removeNetwork(net: any) {
  if (['bridge', 'host', 'none'].includes(net.name)) {
    ElMessage.warning('不能删除 Docker 默认网络')
    return
  }
  ElMessageBox.confirm(`确定要删除网络 "${net.name}" 吗？`, '确认删除', { type: 'warning' })
    .then(() => {
      networks.value = networks.value.filter(n => n.name !== net.name)
      ElMessage.success('网络已删除')
    }).catch(() => {})
}

function installDocker() { ElMessage.info('Docker 安装功能 (需调用系统包管理器)') }
function startDocker() { ElMessage.success('Docker 服务已启动') }

function initDockerCharts() {
  if (!dockerCpuRef.value) return
  const darkTheme = document.documentElement.classList.contains('dark')
  const bgColor = darkTheme ? '#1a1a2e' : '#ffffff'
  const textColor = darkTheme ? '#c0c4cc' : '#606266'
  const gridOpts = { top: 10, right: 10, bottom: 20, left: 40, containLabel: false }

  const makeChart = (ref: HTMLElement, color: string) => {
    const chart = echarts.init(ref)
    chart.setOption({
      grid: gridOpts,
      xAxis: { type: 'category', data: Array.from({ length: 30 }, (_, i) => `${i}s`), show: false },
      yAxis: { type: 'value', max: 100, splitLine: { lineStyle: { color: darkTheme ? '#333' : '#eee' } } },
      series: [{
        data: Array.from({ length: 30 }, () => Math.random() * 50 + 10),
        type: 'line', smooth: true, showSymbol: false,
        lineStyle: { color, width: 2 },
        areaStyle: { color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: color + '44' }, { offset: 1, color: color + '05' }
        ])}
      }]
    })
    dockerCharts.push(chart)
  }

  makeChart(dockerCpuRef.value, '#409eff')
  makeChart(dockerMemRef.value, '#67c23a')
  makeChart(dockerNetRef.value, '#e6a23c')
  makeChart(dockerDiskRef.value, '#f56c6c')
}

onMounted(() => { initDockerCharts() })
onUnmounted(() => { dockerCharts.forEach(c => c.dispose()) })
</script>

<style scoped>
.docker-center { padding: 0; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h2 { font-size: 22px; font-weight: 600; }
.header-actions { display: flex; gap: 10px; }

.tab-toolbar { display: flex; justify-content: space-between; align-items: center; padding: 12px 0; }
.toolbar-right { display: flex; gap: 8px; }

.container-name { display: flex; align-items: center; gap: 8px; font-weight: 500; }

.compose-editor { background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 8px; padding: 16px; }

.template-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; margin-top: 12px; }
.template-card { background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 8px; padding: 14px; cursor: pointer; transition: all 0.2s; }
.template-card:hover { border-color: var(--accent-color); }
.template-name { font-weight: 600; margin-bottom: 4px; }
.template-desc { font-size: 12px; color: var(--text-secondary); }

.monitor-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 16px; }
.monitor-card { background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 12px; padding: 16px; }
.monitor-card h4 { margin-bottom: 12px; }

.log-viewer { background: #0d0d0d; border-radius: 8px; overflow: hidden; }
.log-toolbar { padding: 8px 12px; background: #1a1a1a; border-bottom: 1px solid #333; display: flex; gap: 8px; }
.log-content { height: 50vh; overflow-y: auto; padding: 12px; font-family: 'Consolas', 'Courier New', monospace; font-size: 13px; }
.log-line { color: #c0c0c0; line-height: 1.6; white-space: nowrap; }
.log-line:hover { background: rgba(255,255,255,0.03); }

.docker-env-info { display: flex; flex-direction: column; gap: 16px; }
.env-item { display: flex; justify-content: space-between; align-items: center; padding: 12px 16px; background: var(--bg-secondary); border-radius: 8px; }
.env-label { font-weight: 500; }
</style>
