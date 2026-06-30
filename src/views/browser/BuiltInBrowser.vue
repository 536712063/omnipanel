<template>
  <div class="browser-view">
    <div class="page-header">
      <h2>{{ $t('browser.title') }}</h2>
    </div>

    <div class="browser-toolbar">
      <div class="nav-buttons">
        <el-button :disabled="history.length <= 1" @click="goBack" size="small" circle>
          <el-icon><ArrowLeft /></el-icon>
        </el-button>
        <el-button @click="navigate(currentUrl)" size="small" circle>
          <el-icon><Refresh /></el-icon>
        </el-button>
        <el-button @click="navigate('https://www.google.com')" size="small" circle>
          <el-icon><HomeFilled /></el-icon>
        </el-button>
      </div>

      <div class="address-bar">
        <el-input v-model="urlInput" :placeholder="$t('browser.placeholder')" size="small"
          @keydown.enter="navigate(urlInput)">
          <template #prepend>
            <el-icon><Lock /></el-icon>
          </template>
          <template #append>
            <el-button @click="navigate(urlInput)" size="small">{{ $t('browser.go') }}</el-button>
          </template>
        </el-input>
      </div>

      <div class="toolbar-actions">
        <el-dropdown @command="openExternal">
          <el-button size="small">{{ $t('browser.userAgent') }}</el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="https://www.google.com">Google</el-dropdown-item>
              <el-dropdown-item command="https://github.com">GitHub</el-dropdown-item>
              <el-dropdown-item command="https://hub.docker.com">Docker Hub</el-dropdown-item>
              <el-dropdown-item command="https://gitee.com">Gitee</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-switch v-model="adBlock" size="small" :active-text="$t('browser.adBlock')" />
      </div>
    </div>

    <div class="browser-content">
      <div class="web-frame" v-if="browserUrl">
        <div class="browser-placeholder">
          <div class="web-header">
            <div class="web-title">{{ currentPageTitle }}</div>
          </div>
          <div class="web-body" v-html="webPageContent"></div>
        </div>
      </div>
      <div class="no-page" v-else>
        <div class="no-page-content">
          <el-icon :size="48"><ChromeFilled /></el-icon>
          <h3>{{ $t('browser.title') }}</h3>
          <p>{{ $t('browser.noPage') }}</p>
          <div class="quick-links" v-if="info">
            <h4>{{ $t('browser.quickLinks') }}</h4>
            <div class="quick-link-grid">
              <div v-for="link in quickLinks" :key="link.name" class="quick-link" @click="navigate(link.url)">
                <span class="link-name">{{ link.name }}</span>
                <span class="link-url">{{ link.url }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="browser-footer">
      <div class="footer-left">
        <el-button size="small" text @click="showBookmarks = !showBookmarks">
          <el-icon><Star /></el-icon> {{ $t('browser.bookmarks') }}
        </el-button>
        <el-button size="small" text @click="navigate(currentUrl || ''); showHistory = !showHistory">
          <el-icon><Clock /></el-icon> {{ $t('browser.history') }}
        </el-button>
      </div>
      <div class="footer-right">
        <span class="page-status">{{ currentPageTitle || 'ready' }}</span>
      </div>
    </div>

    <div class="bookmarks-panel" v-if="showBookmarks">
      <div class="bookmarks-header">
        <h4>{{ $t('browser.bookmarks') }}</h4>
        <el-button size="small" @click="showBookmarks = false">{{ $t('common.close') }}</el-button>
      </div>
      <div class="bookmark-list">
        <div v-for="bm in bookmarks" :key="bm.id" class="bookmark-item">
          <span class="bookmark-name">{{ bm.name }}</span>
          <span class="bookmark-url">{{ bm.url }}</span>
          <el-button size="small" text @click="navigate(bm.url)">{{ $t('browser.go') }}</el-button>
          <el-button size="small" text type="danger" @click="deleteBookmark(bm)">{{ $t('common.delete') }}</el-button>
        </div>
      </div>
      <el-button size="small" @click="addBookmark" style="margin-top: 10px">+ {{ $t('common.add') }}</el-button>
    </div>

    <div class="history-panel" v-if="showHistory && history.length">
      <div class="bookmarks-header">
        <h4>{{ $t('browser.history') }}</h4>
        <el-button size="small" @click="showHistory = false">{{ $t('common.close') }}</el-button>
      </div>
      <div class="bookmark-list">
        <div v-for="(url, i) in history.slice().reverse()" :key="i" class="bookmark-item">
          <span class="bookmark-url">{{ url }}</span>
          <el-button size="small" text @click="navigate(url)">{{ $t('browser.go') }}</el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { v4 as uuid } from 'uuid'
import { useBrowser } from '@/composables/useBrowser'

const { currentUrl, history, info, loadInfo, navigate, openExternal, goBack } = useBrowser()

const urlInput = ref('')
const browserUrl = ref('')
const adBlock = ref(true)
const showBookmarks = ref(false)
const showHistory = ref(false)
const webPageContent = ref('')

const currentPageTitle = computed(() => browserUrl.value ? browserUrl.value.replace(/https?:\/\//, '') : '')

const bookmarks = ref([
  { id: '1', name: 'Docker Hub', url: 'https://hub.docker.com' },
  { id: '2', name: 'GitHub', url: 'https://github.com' },
  { id: '3', name: '七日杀 Wiki', url: 'https://7daystodie.fandom.com' },
  { id: '4', name: 'FRP 文档', url: 'https://gofrp.org' }
])

const quickLinks = [
  { name: 'Docker Hub', url: 'https://hub.docker.com' },
  { name: 'GitHub', url: 'https://github.com' },
  { name: 'Gitee', url: 'https://gitee.com' },
  { name: '七日杀 Wiki', url: 'https://7daystodie.fandom.com' },
  { name: 'FRP 文档', url: 'https://gofrp.org' },
  { name: 'Stack Overflow', url: 'https://stackoverflow.com' },
  { name: 'MDN 文档', url: 'https://developer.mozilla.org' },
  { name: 'npm Registry', url: 'https://www.npmjs.com' }
]

function handleNavigate(url: string) {
  if (!url) return
  if (!url.startsWith('http')) url = 'https://' + url
  urlInput.value = url
  browserUrl.value = url
  navigate(url)
  webPageContent.value = `
    <div style="padding: 30px; font-family: sans-serif">
      <div style="background: #f0f7ff; border: 1px solid #409eff; border-radius: 8px; padding: 20px; margin-bottom: 20px">
        <h2 style="color: #409eff; margin-top: 0">${url}</h2>
        <p>这是一个模拟的内置浏览器页面。在 Electron 环境中，此处将渲染真实的网页内容。</p>
      </div>
      <div style="display: grid; grid-template-columns: repeat(3, 1fr); gap: 15px">
        <div style="background: #f5f5f5; padding: 15px; border-radius: 6px"><h4>多标签页浏览</h4></div>
        <div style="background: #f5f5f5; padding: 15px; border-radius: 6px"><h4>网页一键翻译</h4></div>
        <div style="background: #f5f5f5; padding: 15px; border-radius: 6px"><h4>广告拦截支持</h4></div>
      </div>
    </div>
  `
}

function addBookmark() {
  if (browserUrl.value) {
    bookmarks.value.push({ id: uuid(), name: currentPageTitle.value, url: browserUrl.value })
    ElMessage.success('已添加书签')
  }
}

function deleteBookmark(bm: any) { bookmarks.value = bookmarks.value.filter(b => b.id !== bm.id) }

loadInfo()
</script>

<style scoped>
.browser-view { padding: 0; height: calc(100vh - var(--header-height) - 60px); display: flex; flex-direction: column; }
.page-header { margin-bottom: 12px; flex-shrink: 0; }
.page-header h2 { font-size: 22px; font-weight: 600; }

.browser-toolbar { display: flex; align-items: center; gap: 8px; padding: 8px 12px; background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 8px; margin-bottom: 8px; flex-shrink: 0; }
.nav-buttons { display: flex; gap: 4px; }
.address-bar { flex: 1; }
.toolbar-actions { display: flex; gap: 6px; align-items: center; }

.browser-content { flex: 1; overflow: auto; min-height: 0; }
.web-frame { height: 100%; }
.browser-placeholder { height: 100%; background: #ffffff; }

.no-page { display: flex; align-items: center; justify-content: center; height: 100%; }
.no-page-content { text-align: center; }
.no-page-content h3 { margin: 12px 0 6px; }
.no-page-content p { color: var(--text-secondary); margin-bottom: 30px; }

.quick-links { max-width: 700px; margin: 0 auto; }
.quick-links h4 { margin-bottom: 12px; }
.quick-link-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 10px; }
.quick-link { padding: 16px 12px; background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 8px; cursor: pointer; text-align: center; transition: all 0.2s; }
.quick-link:hover { border-color: var(--accent-color); }
.link-name { font-size: 13px; font-weight: 500; display: block; }
.link-url { font-size: 11px; color: var(--text-secondary); display: block; margin-top: 2px; overflow: hidden; text-overflow: ellipsis; }

.bookmarks-panel, .history-panel { background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 8px; padding: 16px; margin-bottom: 8px; }
.bookmarks-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.bookmark-item { display: flex; align-items: center; gap: 12px; padding: 8px 0; border-bottom: 1px solid var(--border-color); }
.bookmark-name { font-weight: 500; min-width: 120px; }
.bookmark-url { font-size: 12px; color: var(--text-secondary); flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.browser-footer { display: flex; justify-content: space-between; align-items: center; padding: 6px 12px; background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 0 0 8px 8px; margin-top: 8px; flex-shrink: 0; }
.footer-left { display: flex; gap: 4px; }
.page-status { font-size: 11px; color: var(--text-secondary); }
</style>
