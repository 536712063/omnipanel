<template>
  <div class="dashboard">
    <div class="page-header">
      <h2>仪表盘</h2>
      <span class="text-secondary">{{ currentTime }}</span>
    </div>

    <div class="stats-grid">
      <div class="stat-card card-hover">
        <div class="stat-icon cpu-icon">
          <el-icon :size="28"><Cpu /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-label">CPU 使用率</div>
          <div class="stat-value">{{ systemInfo.cpuUsage }}%</div>
          <el-progress :percentage="systemInfo.cpuUsage" :show-text="false" :stroke-width="4" />
        </div>
      </div>

      <div class="stat-card card-hover">
        <div class="stat-icon mem-icon">
          <el-icon :size="28"><Memo /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-label">内存使用率</div>
          <div class="stat-value">{{ systemInfo.memUsage }}%</div>
          <el-progress :percentage="systemInfo.memUsage" :show-text="false" :stroke-width="4" :color="memColor" />
        </div>
      </div>

      <div class="stat-card card-hover">
        <div class="stat-icon disk-icon">
          <el-icon :size="28"><Coin /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-label">磁盘使用率</div>
          <div class="stat-value">{{ systemInfo.diskUsage }}%</div>
          <el-progress :percentage="systemInfo.diskUsage" :show-text="false" :stroke-width="4" />
        </div>
      </div>

      <div class="stat-card card-hover">
        <div class="stat-icon net-icon">
          <el-icon :size="28"><Connection /></el-icon>
        </div>
        <div class="stat-info">
          <div class="stat-label">网络速度</div>
          <div class="stat-value">{{ systemInfo.netSpeed }}</div>
          <div class="stat-sub">↓ {{ systemInfo.downloadSpeed }} ↑ {{ systemInfo.uploadSpeed }}</div>
        </div>
      </div>
    </div>

    <div class="charts-grid">
      <div class="chart-card">
        <div class="chart-header">
          <h3>CPU 使用率</h3>
        </div>
        <div class="chart-body" ref="cpuChartRef"></div>
      </div>
      <div class="chart-card">
        <div class="chart-header">
          <h3>内存使用率</h3>
        </div>
        <div class="chart-body" ref="memChartRef"></div>
      </div>
      <div class="chart-card">
        <div class="chart-header">
          <h3>网络 IO</h3>
        </div>
        <div class="chart-body" ref="netChartRef"></div>
      </div>
      <div class="chart-card">
        <div class="chart-header">
          <h3>磁盘 IO</h3>
        </div>
        <div class="chart-body" ref="diskChartRef"></div>
      </div>
    </div>

    <div class="info-grid">
      <div class="info-card">
        <h3>系统信息</h3>
        <div class="info-list">
          <div class="info-item">
            <span class="info-label">操作系统</span>
            <span class="info-value">{{ systemInfo.os }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">主机名</span>
            <span class="info-value">{{ systemInfo.hostname }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">运行时间</span>
            <span class="info-value">{{ systemInfo.uptime }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">CPU 型号</span>
            <span class="info-value">{{ systemInfo.cpuModel }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">总内存</span>
            <span class="info-value">{{ systemInfo.totalMemory }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">可用内存</span>
            <span class="info-value">{{ systemInfo.freeMemory }}</span>
          </div>
        </div>
      </div>

      <div class="info-card">
        <h3>模块状态</h3>
        <div class="module-list">
          <div class="module-item" v-for="mod in modules" :key="mod.name">
            <div class="module-left">
              <span class="module-icon">{{ mod.icon }}</span>
              <span class="module-name">{{ mod.name }}</span>
            </div>
            <el-tag :type="mod.status === 'running' ? 'success' : 'info'" size="small">
              {{ mod.status === 'running' ? '运行中' : '待启动' }}
            </el-tag>
          </div>
        </div>
      </div>

      <div class="info-card">
        <h3>快捷操作</h3>
        <div class="quick-actions">
          <el-button type="primary" size="small" @click="$router.push('/docker')">
            <el-icon><Box /></el-icon> Docker 管理
          </el-button>
          <el-button type="success" size="small" @click="$router.push('/ssh')">
            <el-icon><Connection /></el-icon> SSH 连接
          </el-button>
          <el-button type="warning" size="small" @click="$router.push('/ai')">
            <el-icon><Cpu /></el-icon> AI 助手
          </el-button>
          <el-button type="danger" size="small" @click="$router.push('/game')">
            <el-icon><VideoGame /></el-icon> 七日杀
          </el-button>
          <el-button size="small" @click="$router.push('/translate')">
            <el-icon><Document /></el-icon> 汉化工具
          </el-button>
          <el-button size="small" @click="$router.push('/settings')">
            <el-icon><Setting /></el-icon> 系统设置
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, computed } from 'vue'
import * as echarts from 'echarts'

const currentTime = ref('')
let timeTimer: any = null

const systemInfo = reactive({
  cpuUsage: 23,
  memUsage: 45,
  diskUsage: 32,
  netSpeed: '1.2 MB/s',
  downloadSpeed: '850 KB/s',
  uploadSpeed: '380 KB/s',
  os: 'Windows 11 Pro',
  hostname: 'OMNIPANEL-DEV',
  uptime: '3天 12小时',
  cpuModel: 'Intel Core i7-13700K',
  totalMemory: '32.0 GB',
  freeMemory: '17.6 GB'
})

const modules = reactive([
  { name: 'Docker 引擎', icon: '🐳', status: 'running' },
  { name: 'SSH 服务', icon: '🔐', status: 'running' },
  { name: 'FRP 穿透', icon: '🌐', status: 'stopped' },
  { name: 'AI 助手', icon: '🤖', status: 'running' },
  { name: '七日杀服务', icon: '🎮', status: 'stopped' },
  { name: '后端 API', icon: '⚙️', status: 'running' }
])

const memColor = computed(() => {
  if (systemInfo.memUsage > 80) return '#f56c6c'
  if (systemInfo.memUsage > 60) return '#e6a23c'
  return '#67c23a'
})

const cpuChartRef = ref<HTMLElement>()
const memChartRef = ref<HTMLElement>()
const netChartRef = ref<HTMLElement>()
const diskChartRef = ref<HTMLElement>()
let charts: echarts.ECharts[] = []

function createTimeSeriesData(count: number) {
  const now = new Date()
  return Array.from({ length: count }, (_, i) => {
    const t = new Date(now.getTime() - (count - i) * 1000)
    return t.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
  })
}

function createRandomData(count: number, base: number, variance: number) {
  return Array.from({ length: count }, () => Math.max(0, Math.min(100, base + (Math.random() - 0.5) * variance)))
}

function createNetworkData(count: number) {
  return Array.from({ length: count }, () => Math.max(0, Math.random() * 100))
}

function initChart(ref: HTMLElement | undefined, options: any) {
  if (!ref) return
  const chart = echarts.init(ref)
  chart.setOption(options)
  charts.push(chart)
  return chart
}

function initCharts() {
  const timeLabels = createTimeSeriesData(60)
  const darkTheme = document.documentElement.classList.contains('dark')

  const textColor = darkTheme ? '#c0c4cc' : '#606266'
  const gridBg = darkTheme ? '#1a1a2e' : '#ffffff'

  const chartGrid = {
    top: 10,
    right: 10,
    bottom: 20,
    left: 40,
    containLabel: false
  }

  initChart(cpuChartRef.value, {
    grid: chartGrid,
    xAxis: { type: 'category', data: timeLabels, show: false },
    yAxis: { type: 'value', max: 100, min: 0, splitLine: { lineStyle: { color: darkTheme ? '#333' : '#eee' } } },
    series: [{
      data: createRandomData(60, systemInfo.cpuUsage, 20),
      type: 'line',
      smooth: true,
      showSymbol: false,
      lineStyle: { color: '#409eff', width: 2 },
      areaStyle: { color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
        { offset: 0, color: 'rgba(64,158,255,0.4)' },
        { offset: 1, color: 'rgba(64,158,255,0.05)' }
      ]) }
    }]
  })

  initChart(memChartRef.value, {
    grid: chartGrid,
    xAxis: { type: 'category', data: timeLabels, show: false },
    yAxis: { type: 'value', max: 100, min: 0, splitLine: { lineStyle: { color: darkTheme ? '#333' : '#eee' } } },
    series: [{
      data: createRandomData(60, systemInfo.memUsage, 10),
      type: 'line',
      smooth: true,
      showSymbol: false,
      lineStyle: { color: '#67c23a', width: 2 },
      areaStyle: { color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
        { offset: 0, color: 'rgba(103,194,58,0.4)' },
        { offset: 1, color: 'rgba(103,194,58,0.05)' }
      ]) }
    }]
  })

  initChart(netChartRef.value, {
    grid: chartGrid,
    xAxis: { type: 'category', data: timeLabels, show: false },
    yAxis: { type: 'value', splitLine: { lineStyle: { color: darkTheme ? '#333' : '#eee' } } },
    series: [
      {
        data: createNetworkData(60),
        type: 'line',
        smooth: true,
        showSymbol: false,
        lineStyle: { color: '#409eff', width: 2 },
        name: '下载'
      },
      {
        data: createNetworkData(60),
        type: 'line',
        smooth: true,
        showSymbol: false,
        lineStyle: { color: '#e6a23c', width: 2 },
        name: '上传'
      }
    ]
  })

  initChart(diskChartRef.value, {
    grid: chartGrid,
    xAxis: { type: 'category', data: timeLabels, show: false },
    yAxis: { type: 'value', splitLine: { lineStyle: { color: darkTheme ? '#333' : '#eee' } } },
    series: [
      {
        data: createNetworkData(60),
        type: 'line',
        smooth: true,
        showSymbol: false,
        lineStyle: { color: '#e6a23c', width: 2 },
        name: '读取'
      },
      {
        data: createNetworkData(60),
        type: 'line',
        smooth: true,
        showSymbol: false,
        lineStyle: { color: '#f56c6c', width: 2 },
        name: '写入'
      }
    ]
  })
}

function updateTime() {
  currentTime.value = new Date().toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    weekday: 'long'
  })
}

// Simulate updating metrics
let metricTimer: any = null

onMounted(async () => {
  updateTime()
  timeTimer = setInterval(updateTime, 1000)
  initCharts()

  metricTimer = setInterval(() => {
    systemInfo.cpuUsage = Math.max(0, Math.min(100, systemInfo.cpuUsage + (Math.random() - 0.5) * 10))
    systemInfo.memUsage = Math.max(0, Math.min(100, systemInfo.memUsage + (Math.random() - 0.5) * 6))
    systemInfo.diskUsage = Math.max(0, Math.min(100, systemInfo.diskUsage + (Math.random() - 0.5) * 2))
  }, 3000)
})

onUnmounted(() => {
  clearInterval(timeTimer)
  clearInterval(metricTimer)
  charts.forEach(c => c.dispose())
})
</script>

<style scoped>
.dashboard { padding: 0; }

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-header h2 {
  font-size: 22px;
  font-weight: 600;
  color: var(--text-primary);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.cpu-icon { background: rgba(64,158,255,0.15); color: #409eff; }
.mem-icon { background: rgba(103,194,58,0.15); color: #67c23a; }
.disk-icon { background: rgba(230,162,60,0.15); color: #e6a23c; }
.net-icon { background: rgba(245,108,108,0.15); color: #f56c6c; }

.stat-info { flex: 1; min-width: 0; }
.stat-label { font-size: 12px; color: var(--text-secondary); margin-bottom: 4px; }
.stat-value { font-size: 24px; font-weight: 700; color: var(--text-primary); margin-bottom: 8px; }
.stat-sub { font-size: 11px; color: var(--text-secondary); margin-top: 4px; }

.charts-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.chart-card {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 16px;
}

.chart-header {
  margin-bottom: 8px;
}

.chart-header h3 {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.chart-body {
  height: 180px;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.info-card {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 20px;
}

.info-card h3 {
  font-size: 15px;
  font-weight: 600;
  margin-bottom: 16px;
  color: var(--text-primary);
}

.info-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-label {
  font-size: 13px;
  color: var(--text-secondary);
}

.info-value {
  font-size: 13px;
  color: var(--text-primary);
  font-weight: 500;
}

.module-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.module-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.module-left {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.module-icon { font-size: 16px; }

.quick-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

@media (max-width: 1200px) {
  .stats-grid { grid-template-columns: repeat(2, 1fr); }
  .charts-grid { grid-template-columns: 1fr; }
  .info-grid { grid-template-columns: 1fr 1fr; }
}
</style>
