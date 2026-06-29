<template>
  <div class="browser-view">
    <div class="page-header">
      <h2>内置浏览器</h2>
    </div>

    <div class="browser-toolbar">
      <div class="nav-buttons">
        <el-button :disabled="!canGoBack" @click="goBack" size="small" circle>
          <el-icon><ArrowLeft /></el-icon>
        </el-button>
        <el-button :disabled="!canGoForward" @click="goForward" size="small" circle>
          <el-icon><ArrowRight /></el-icon>
        </el-button>
        <el-button @click="refreshPage" size="small" circle>
          <el-icon><Refresh /></el-icon>
        </el-button>
        <el-button @click="goHome" size="small" circle>
          <el-icon><HomeFilled /></el-icon>
        </el-button>
      </div>

      <div class="address-bar">
        <el-input v-model="urlInput" placeholder="输入网址..." size="small"
                  @keydown.enter="navigateTo(urlInput)">
          <template #prepend>
            <el-icon v-if="pageSecure" style="color: #67c23a"><Lock /></el-icon>
            <el-icon v-else><Warning /></el-icon>
          </template>
          <template #append>
            <el-button @click="navigateTo(urlInput)" size="small">转到</el-button>
          </template>
        </el-input>
      </div>

      <div class="toolbar-actions">
        <el-button size="small" @click="translatePage">翻译</el-button>
        <el-dropdown @command="handleUAChange">
          <el-button size="small">UA 切换</el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="default">默认 (Chrome)</el-dropdown-item>
              <el-dropdown-item command="mobile">移动端 (iPhone)</el-dropdown-item>
              <el-dropdown-item command="android">Android</el-dropdown-item>
              <el-dropdown-item command="edge">Edge</el-dropdown-item>
              <el-dropdown-item command="firefox">Firefox</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-switch v-model="adBlock" size="small" active-text="广告拦截" />
        <el-button size="small" @click="openDevTools" circle>
          <el-icon><Tools /></el-icon>
        </el-button>
      </div>
    </div>

    <div class="browser-tabs" v-if="tabs.length > 1">
      <div class="tab-item" v-for="tab in tabs" :key="tab.id"
           :class="{ active: tab.id === activeTabId }"
           @click="switchTab(tab.id)">
        <span class="tab-title">{{ tab.title }}</span>
        <el-button text size="small" @click.stop="closeTab(tab.id)">
          <el-icon><Close /></el-icon>
        </el-button>
      </div>
      <el-button size="small" text @click="newTab" style="margin-left: 4px">
        <el-icon><Plus /></el-icon>
      </el-button>
    </div>

    <div class="browser-content" v-if="!showBookmarks">
      <div class="web-frame" v-if="browserUrl">
        <div class="browser-placeholder">
          <div class="web-header">
            <div v-if="currentPage" class="web-title">{{ currentPage.title }}</div>
          </div>
          <div class="web-body" v-html="webPageContent"></div>
        </div>
      </div>
      <div class="no-page" v-else>
        <div class="no-page-content">
          <el-icon :size="48"><ChromeFilled /></el-icon>
          <h3>内置浏览器</h3>
          <p>在地址栏输入 URL 开始浏览</p>
          <div class="quick-links">
            <h4>快捷入口</h4>
            <div class="quick-link-grid">
              <div v-for="link in quickLinks" :key="link.name" class="quick-link" @click="navigateTo(link.url)">
                <span class="link-icon">{{ link.icon }}</span>
                <span class="link-name">{{ link.name }}</span>
                <span class="link-url">{{ link.url }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="bookmarks-panel" v-if="showBookmarks">
      <div class="bookmarks-header">
        <h4>书签管理</h4>
        <el-button size="small" @click="showBookmarks = false">关闭</el-button>
      </div>
      <div class="bookmark-list">
        <div v-for="bm in bookmarks" :key="bm.id" class="bookmark-item">
          <span class="bookmark-name">{{ bm.name }}</span>
          <span class="bookmark-url">{{ bm.url }}</span>
          <el-button size="small" text @click="navigateTo(bm.url)">打开</el-button>
          <el-button size="small" text type="danger" @click="deleteBookmark(bm)">删除</el-button>
        </div>
      </div>
      <el-button size="small" @click="addBookmark" style="margin-top: 10px">+ 添加书签</el-button>
    </div>

    <div class="browser-footer">
      <div class="footer-left">
        <el-button size="small" text @click="showBookmarks = !showBookmarks">
          <el-icon><Star /></el-icon> 书签
        </el-button>
        <el-button size="small" text @click="showHistory = !showHistory">
          <el-icon><Clock /></el-icon> 历史
        </el-button>
        <el-button size="small" text @click="showDownloadManager">
          <el-icon><Download /></el-icon> 下载
        </el-button>
      </div>
      <div class="footer-right">
        <span class="page-status">{{ currentPage ? currentPage.url : '就绪' }}</span>
      </div>
    </div>

    <el-dialog v-model="showDownloadManager" title="下载管理" width="600px">
      <el-table :data="downloads" size="small">
        <el-table-column prop="name" label="文件名" min-width="200" />
        <el-table-column prop="size" label="大小" width="100" />
        <el-table-column prop="progress" label="进度" width="150">
          <template #default="{ row }">
            <el-progress :percentage="row.progress" :status="row.status === 'completed' ? 'success' : ''" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button size="small" v-if="row.status === 'completed'" @click="openDownload(row)">打开</el-button>
            <el-button size="small" type="danger" @click="cancelDownload(row)">取消</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { v4 as uuid } from 'uuid'

const urlInput = ref('')
const browserUrl = ref('')
const canGoBack = ref(false)
const canGoForward = ref(false)
const pageSecure = ref(true)
const adBlock = ref(true)
const showBookmarks = ref(false)
const showHistory = ref(false)
const showDownloadManager = ref(false)
const activeTabId = ref('1')
const webPageContent = ref('')

const currentPage = ref<{ title: string; url: string } | null>(null)

const tabs = ref([
  { id: '1', title: '新标签页', url: '' }
])

const bookmarks = ref([
  { id: '1', name: 'Docker Hub', url: 'https://hub.docker.com' },
  { id: '2', name: 'GitHub', url: 'https://github.com' },
  { id: '3', name: '七日杀 Wiki', url: 'https://7daystodie.fandom.com' },
  { id: '4', name: 'FRP 文档', url: 'https://gofrp.org' }
])

const downloads = ref([
  { id: '1', name: 'nginx.conf', size: '2.3KB', progress: 100, status: 'completed' },
  { id: '2', name: 'Dockerfile', size: '512B', progress: 45, status: 'downloading' }
])

const quickLinks = [
  { name: 'Docker Hub', icon: '🐳', url: 'https://hub.docker.com' },
  { name: 'GitHub', icon: '🐙', url: 'https://github.com' },
  { name: 'Gitee', icon: '🇨🇳', url: 'https://gitee.com' },
  { name: '七日杀 Wiki', icon: '🎮', url: 'https://7daystodie.fandom.com' },
  { name: 'FRP 文档', icon: '🌐', url: 'https://gofrp.org' },
  { name: 'Stack Overflow', icon: '📚', url: 'https://stackoverflow.com' },
  { name: 'MDN 文档', icon: '📖', url: 'https://developer.mozilla.org' },
  { name: 'npm Registry', icon: '📦', url: 'https://www.npmjs.com' }
]

function navigateTo(url: string) {
  if (!url) return
  if (!url.startsWith('http://') && !url.startsWith('https://')) {
    url = 'https://' + url
  }
  urlInput.value = url
  browserUrl.value = url
  currentPage.value = { title: url.replace(/https?:\/\//, ''), url }
  canGoBack.value = true

  webPageContent.value = `
    <div style="padding: 30px; font-family: sans-serif">
      <div style="background: #f0f7ff; border: 1px solid #409eff; border-radius: 8px; padding: 20px; margin-bottom: 20px">
        <h2 style="color: #409eff; margin-top: 0">${url}</h2>
        <p>这是一个模拟的内置浏览器页面。在 Electron 环境中，此处将渲染真实的网页内容。</p>
        <p>您可以在此浏览网页、调试应用和访问在线资源。</p>
      </div>
      <div style="display: grid; grid-template-columns: repeat(3, 1fr); gap: 15px">
        <div style="background: #f5f5f5; padding: 15px; border-radius: 6px">
          <h4>功能 1</h4><p>多标签页浏览</p>
        </div>
        <div style="background: #f5f5f5; padding: 15px; border-radius: 6px">
          <h4>功能 2</h4><p>网页一键翻译</p>
        </div>
        <div style="background: #f5f5f5; padding: 15px; border-radius: 6px">
          <h4>功能 3</h4><p>广告拦截支持</p>
        </div>
      </div>
    </div>
  `
}

function goBack() { ElMessage.info('后退 (浏览器历史记录)') }
function goForward() { ElMessage.info('前进 (浏览器历史记录)') }
function refreshPage() { if (browserUrl.value) navigateTo(browserUrl.value); ElMessage.success('页面已刷新') }
function goHome() { browserUrl.value = ''; urlInput.value = ''; currentPage.value = null; webPageContent.value = '' }

function translatePage() { ElMessage.success('页面翻译已启用 (调用汉化工具模块)') }
function handleUAChange(ua: string) { ElMessage.info(`User Agent 切换为: ${ua}`) }
function openDevTools() { ElMessage.info('开发者工具 (F12)') }

function newTab() {
  const id = uuid()
  tabs.value.push({ id, title: '新标签页', url: '' })
  switchTab(id)
}

function switchTab(id: string) { activeTabId.value = id }
function closeTab(id: string) {
  if (tabs.value.length <= 1) return
  tabs.value = tabs.value.filter(t => t.id !== id)
  if (activeTabId.value === id) activeTabId.value = tabs.value[0].id
}

function addBookmark() {
  if (currentPage.value) {
    bookmarks.value.push({ id: uuid(), name: currentPage.value.title, url: currentPage.value.url })
    ElMessage.success('已添加书签')
  }
}

function deleteBookmark(bm: any) { bookmarks.value = bookmarks.value.filter(b => b.id !== bm.id) }

function openDownload(row: any) { ElMessage.success(`打开 ${row.name}`) }
function cancelDownload(row: any) { downloads.value = downloads.value.filter(d => d.id !== row.id) }
</script>

<style scoped>
.browser-view { padding: 0; height: calc(100vh - var(--header-height) - 60px); display: flex; flex-direction: column; }
.page-header { margin-bottom: 12px; flex-shrink: 0; }
.page-header h2 { font-size: 22px; font-weight: 600; }

.browser-toolbar { display: flex; align-items: center; gap: 8px; padding: 8px 12px; background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 8px; margin-bottom: 8px; flex-shrink: 0; }
.nav-buttons { display: flex; gap: 4px; }
.address-bar { flex: 1; }
.toolbar-actions { display: flex; gap: 6px; align-items: center; }

.browser-tabs { display: flex; align-items: center; gap: 2px; padding: 4px 8px 0; background: var(--bg-secondary); border-radius: 6px 6px 0 0; overflow-x: auto; flex-shrink: 0; }
.tab-item { display: flex; align-items: center; gap: 6px; padding: 6px 12px; background: var(--bg-primary); border-radius: 6px 6px 0 0; cursor: pointer; font-size: 12px; white-space: nowrap; max-width: 200px; }
.tab-item.active { background: var(--accent-color); color: white; }
.tab-title { overflow: hidden; text-overflow: ellipsis; }

.browser-content { flex: 1; overflow: auto; min-height: 0; }
.web-frame { height: 100%; }
.browser-placeholder { height: 100%; background: #ffffff; }
.web-body { padding: 0; }

.no-page { display: flex; align-items: center; justify-content: center; height: 100%; }
.no-page-content { text-align: center; }
.no-page-content h3 { margin: 12px 0 6px; }
.no-page-content p { color: var(--text-secondary); margin-bottom: 30px; }

.quick-links { max-width: 700px; margin: 0 auto; }
.quick-links h4 { margin-bottom: 12px; }
.quick-link-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 10px; }
.quick-link { padding: 16px 12px; background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 8px; cursor: pointer; text-align: center; transition: all 0.2s; }
.quick-link:hover { border-color: var(--accent-color); }
.link-icon { display: block; font-size: 28px; margin-bottom: 6px; }
.link-name { font-size: 13px; font-weight: 500; display: block; }
.link-url { font-size: 11px; color: var(--text-secondary); display: block; margin-top: 2px; overflow: hidden; text-overflow: ellipsis; }

.bookmarks-panel { background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 8px; padding: 16px; margin-bottom: 8px; }
.bookmarks-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.bookmark-item { display: flex; align-items: center; gap: 12px; padding: 8px 0; border-bottom: 1px solid var(--border-color); }
.bookmark-name { font-weight: 500; min-width: 120px; }
.bookmark-url { font-size: 12px; color: var(--text-secondary); flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.browser-footer { display: flex; justify-content: space-between; align-items: center; padding: 6px 12px; background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 0 0 8px 8px; margin-top: 8px; flex-shrink: 0; }
.footer-left { display: flex; gap: 4px; }
.page-status { font-size: 11px; color: var(--text-secondary); }
</style>
