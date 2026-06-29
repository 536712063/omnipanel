<template>
  <div class="code-repo">
    <div class="page-header">
      <h2>代码仓库管理</h2>
      <div class="header-actions">
        <el-button type="primary" @click="showCloneDialog = true">
          <el-icon><Download /></el-icon> 克隆仓库
        </el-button>
        <el-button @click="showCreateDialog = true">
          <el-icon><Plus /></el-icon> 新建仓库
        </el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab">
      <el-tab-pane label="仓库列表" name="repos">
        <div class="repo-grid">
          <div class="repo-card card-hover" v-for="repo in repos" :key="repo.id" @click="selectRepo(repo)">
            <div class="repo-card-header">
              <div class="repo-name">
                <el-icon><FolderOpened /></el-icon>
                {{ repo.name }}
              </div>
              <el-tag :type="repo.platform === 'github' ? '' : repo.platform === 'gitlab' ? 'warning' : 'success'" size="small">
                {{ repo.platform }}
              </el-tag>
            </div>
            <div class="repo-description">{{ repo.description }}</div>
            <div class="repo-stats">
              <span>分支: {{ repo.branch }}</span>
              <span>最近提交: {{ repo.lastCommit }}</span>
            </div>
            <div class="repo-actions">
              <el-button size="small" @click.stop="gitAction(repo, 'pull')">
                <el-icon><Download /></el-icon> Pull
              </el-button>
              <el-button size="small" @click.stop="gitAction(repo, 'push')">
                <el-icon><Upload /></el-icon> Push
              </el-button>
              <el-button size="small" @click.stop="showDiffViewer(repo)">差异</el-button>
              <el-button size="small" type="danger" @click.stop="deleteRepo(repo)">删除</el-button>
            </div>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="代码差异" name="diff">
        <div class="diff-viewer" v-if="activeRepo">
          <div class="diff-header">
            <span>{{ activeRepo.name }} - git diff</span>
            <el-select v-model="diffBase" size="small" style="width: 120px">
              <el-option label="HEAD" value="HEAD" />
              <el-option label="main" value="main" />
              <el-option label="develop" value="develop" />
            </el-select>
          </div>
          <div class="diff-content">
            <div v-for="(chunk, i) in diffChunks" :key="i" class="diff-chunk">
              <div class="diff-chunk-header">{{ chunk.header }}</div>
              <div v-for="(line, j) in chunk.lines" :key="j" :class="['diff-line', `diff-${line.type}`]">
                <span class="diff-line-num">{{ line.oldNum }}</span>
                <span class="diff-line-num">{{ line.newNum }}</span>
                <span class="diff-line-text">{{ line.text }}</span>
              </div>
            </div>
          </div>
        </div>
        <div v-else class="no-selection">
          <p>请选择一个仓库查看差异</p>
        </div>
      </el-tab-pane>

      <el-tab-pane label="搜索代码" name="search">
        <div class="code-search">
          <div class="search-bar">
            <el-select v-model="searchRepo" placeholder="选择仓库" style="width: 200px">
              <el-option v-for="r in repos" :key="r.id" :label="r.name" :value="r.id" />
            </el-select>
            <el-input v-model="searchPattern" placeholder="搜索代码内容..." clearable style="width: 300px; margin-left: 10px">
              <template #append>
                <el-button @click="searchCode">搜索</el-button>
              </template>
            </el-input>
          </div>
          <div class="search-results" v-if="searchResults.length">
            <div class="search-result" v-for="(res, i) in searchResults" :key="i">
              <div class="result-file">{{ res.file }}:{{ res.line }}</div>
              <div class="result-code">{{ res.content }}</div>
            </div>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="Webhook" name="webhook">
        <div class="webhook-section">
          <h4>自动部署钩子</h4>
          <el-table :data="webhooks">
            <el-table-column prop="name" label="名称" width="150" />
            <el-table-column prop="url" label="Webhook URL" min-width="250" />
            <el-table-column prop="events" label="触发事件" width="150" />
            <el-table-column prop="status" label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
                  {{ row.status === 'active' ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="100">
              <template #default="{ row }">
                <el-button size="small" type="danger" @click="deleteWebhook(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
          <el-button size="small" type="primary" style="margin-top: 12px" @click="addWebhook">添加 Webhook</el-button>
        </div>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="showCloneDialog" title="克隆仓库" width="500px">
      <el-form :model="cloneForm" label-width="80px">
        <el-form-item label="仓库 URL">
          <el-input v-model="cloneForm.url" placeholder="https://github.com/user/repo.git" />
        </el-form-item>
        <el-form-item label="本地路径">
          <el-input v-model="cloneForm.path" placeholder="/workspace/repo" />
        </el-form-item>
        <el-form-item label="分支">
          <el-input v-model="cloneForm.branch" placeholder="main" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCloneDialog = false">取消</el-button>
        <el-button type="primary" @click="cloneRepo">克隆</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { v4 as uuid } from 'uuid'

const activeTab = ref('repos')
const showCloneDialog = ref(false)
const showCreateDialog = ref(false)
const activeRepo = ref<any>(null)
const diffBase = ref('HEAD')
const searchRepo = ref('')
const searchPattern = ref('')
const searchResults = ref<any[]>([])

const cloneForm = reactive({ url: '', path: '', branch: 'main' })

const repos = ref([
  { id: '1', name: 'omnipanel', description: 'OmniPanel 主仓库', platform: 'github', branch: 'main', lastCommit: '2小时前' },
  { id: '2', name: '7days-server-config', description: '七日杀服务器配置', platform: 'gitlab', branch: 'master', lastCommit: '1天前' },
  { id: '3', name: 'docker-compose-files', description: 'Docker Compose 配置文件集合', platform: 'gitee', branch: 'main', lastCommit: '3天前' },
  { id: '4', name: 'nginx-configs', description: 'Nginx 配置模板', platform: 'github', branch: 'main', lastCommit: '1周前' }
])

const diffChunks = ref([
  {
    header: '@@ -1,7 +1,8 @@',
    lines: [
      { oldNum: '1', newNum: '1', type: 'normal', text: '# OmniPanel' },
      { oldNum: '2', newNum: '2', type: 'normal', text: '' },
      { oldNum: '', newNum: '3', type: 'add', text: '+ ## 新功能' },
      { oldNum: '3', newNum: '4', type: 'normal', text: '- Docker 管理模块' },
      { oldNum: '4', newNum: '5', type: 'delete', text: '- 基础功能' },
      { oldNum: '', newNum: '6', type: 'add', text: '+ 完整 Docker 管理' }
    ]
  }
])

const webhooks = ref([
  { id: '1', name: '自动部署', url: 'http://server:3001/webhook/deploy', events: 'push', status: 'active' },
  { id: '2', name: '通知', url: 'http://server:3001/webhook/notify', events: 'push, merge', status: 'active' }
])

function selectRepo(repo: any) { activeRepo.value = repo; ElMessage.info(`已选中 ${repo.name}`) }

function gitAction(repo: any, action: string) {
  const actions: Record<string, string> = { pull: 'Pull', push: 'Push' }
  ElMessage.success(`${actions[action]} ${repo.name} 成功`)
}

function showDiffViewer(repo: any) {
  activeRepo.value = repo
  activeTab.value = 'diff'
}

function deleteRepo(repo: any) {
  ElMessageBox.confirm(`确定删除仓库 "${repo.name}" 吗？`, '确认删除', { type: 'warning' })
    .then(() => { repos.value = repos.value.filter(r => r.id !== repo.id); ElMessage.success('仓库已删除') }).catch(() => {})
}

function cloneRepo() {
  if (!cloneForm.url) { ElMessage.warning('请输入仓库 URL'); return }
  repos.value.push({
    id: uuid(), name: cloneForm.url.split('/').pop()?.replace('.git', '') || 'unknown',
    description: '新克隆的仓库', platform: 'github', branch: cloneForm.branch, lastCommit: '刚刚'
  })
  showCloneDialog.value = false
  ElMessage.success('仓库克隆成功')
}

function searchCode() {
  searchResults.value = [
    { file: 'src/main.ts', line: 23, content: 'const app = createApp(App)' },
    { file: 'src/router/index.ts', line: 5, content: 'const routes = [' }
  ]
}

function addWebhook() {
  webhooks.value.push({ id: uuid(), name: '新 Webhook', url: 'http://...', events: 'push', status: 'active' })
  ElMessage.success('Webhook 已添加')
}

function deleteWebhook(row: any) {
  webhooks.value = webhooks.value.filter(w => w.id !== row.id)
  ElMessage.success('Webhook 已删除')
}
</script>

<style scoped>
.code-repo { padding: 0; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { font-size: 22px; font-weight: 600; }
.header-actions { display: flex; gap: 10px; }

.repo-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 16px; }
.repo-card { background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 12px; padding: 16px; cursor: pointer; }
.repo-card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.repo-name { font-weight: 600; display: flex; align-items: center; gap: 8px; }
.repo-description { font-size: 13px; color: var(--text-secondary); margin-bottom: 8px; }
.repo-stats { display: flex; justify-content: space-between; font-size: 12px; color: var(--text-secondary); margin-bottom: 12px; }
.repo-actions { display: flex; gap: 6px; }

.diff-viewer { background: #0d0d0d; border-radius: 8px; overflow: hidden; }
.diff-header { padding: 8px 12px; background: #1a1a1a; display: flex; justify-content: space-between; align-items: center; color: #c0c4cc; }
.diff-content { padding: 8px; font-family: 'Consolas', monospace; font-size: 13px; overflow: auto; max-height: 500px; }
.diff-chunk { margin-bottom: 8px; }
.diff-chunk-header { color: #569cd6; padding: 4px 0; }
.diff-line { display: flex; line-height: 1.6; }
.diff-line-num { width: 40px; text-align: right; padding-right: 8px; color: #606366; user-select: none; }
.diff-line-text { flex: 1; white-space: pre; }
.diff-add { background: rgba(64,200,64,0.1); }
.diff-add .diff-line-text { color: #4ec94e; }
.diff-delete { background: rgba(255,80,80,0.1); }
.diff-delete .diff-line-text { color: #f44747; }
.diff-normal .diff-line-text { color: #c0c0c0; }

.code-search { padding: 12px 0; }
.search-bar { display: flex; align-items: center; margin-bottom: 16px; }
.search-result { padding: 8px 12px; background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 6px; margin-bottom: 8px; }
.result-file { font-size: 12px; color: var(--accent-color); margin-bottom: 4px; }
.result-code { font-family: monospace; font-size: 13px; }

.webhook-section { padding: 12px 0; }
.webhook-section h4 { margin-bottom: 12px; }
.no-selection { text-align: center; padding: 60px; color: var(--text-secondary); }
</style>
