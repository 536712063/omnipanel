<template>
  <div class="code-repo">
    <div class="page-header">
      <h2>{{ $t('git.title') }}</h2>
      <div class="header-actions">
        <el-button type="primary" @click="showCloneDialog = true">
          <el-icon><Download /></el-icon> {{ $t('git.clone') }}
        </el-button>
        <el-button @click="showAddLocalDialog = true">
          <el-icon><Plus /></el-icon> {{ $t('common.add') }}
        </el-button>
      </div>
    </div>

    <div class="repo-layout">
      <div class="repo-list-panel">
        <div class="repo-cards">
          <div class="repo-card card-hover" v-for="repo in repos" :key="repo.id"
            :class="{ active: repo.id === selectedRepoId }" @click="selectRepo(repo.id)">
            <div class="repo-card-header">
              <div class="repo-name">
                <el-icon><FolderOpened /></el-icon>
                {{ repo.name }}
              </div>
            </div>
            <div class="repo-path">{{ repo.path }}</div>
            <div class="repo-stats" v-if="stats[repo.id]">
              <span>{{ stats[repo.id] }} commits</span>
            </div>
          </div>
        </div>
      </div>

      <div class="repo-detail" v-if="selectedRepoId">
        <div class="detail-toolbar">
          <el-button size="small" @click="pull">{{ $t('git.pull') }}</el-button>
          <el-button size="small" @click="push">{{ $t('git.push') }}</el-button>
          <el-button size="small" @click="showCommitDialog = true">{{ $t('git.commit') }}</el-button>
          <el-select v-model="currentBranch" size="small" style="width: 150px" @change="checkout">
            <el-option v-for="b in branches" :key="b.name" :label="b.is_current ? `* ${b.name}` : b.name" :value="b.name" />
          </el-select>
          <el-button size="small" @click="showBranchDialog = true">{{ $t('git.createBranch') }}</el-button>
        </div>

        <div class="detail-panels">
          <div class="status-panel">
            <h4>{{ $t('git.fileStatus') }}</h4>
            <div v-if="statusItems.length" class="status-list">
              <div v-for="item in statusItems" :key="item.file" class="status-item">
                <el-tag :type="statusTagType(item.status)" size="small">{{ item.status }}</el-tag>
                <span>{{ item.file }}</span>
              </div>
            </div>
            <div v-else class="no-status">{{ $t('common.loading') }}</div>
          </div>

          <div class="log-panel">
            <h4>{{ $t('git.log') }}</h4>
            <div v-if="commits.length" class="log-list">
              <div v-for="c in commits" :key="c.hash" class="log-item">
                <div class="log-hash">{{ c.short_hash }}</div>
                <div class="log-msg">{{ c.message }}</div>
                <div class="log-meta">{{ c.author }} &middot; {{ c.date?.slice(0, 16) }}</div>
              </div>
            </div>
            <div v-else class="no-status">{{ $t('common.loading') }}</div>
          </div>
        </div>
      </div>

      <div v-else class="no-selection">
        <el-icon :size="48"><FolderOpened /></el-icon>
        <p>{{ $t('git.repoList') }}</p>
      </div>
    </div>

    <el-dialog v-model="showCloneDialog" :title="$t('git.clone')" width="500px">
      <el-form :model="cloneForm" label-width="80px">
        <el-form-item :label="$t('git.cloneUrl')">
          <el-input v-model="cloneForm.url" placeholder="https://github.com/user/repo.git" />
        </el-form-item>
        <el-form-item :label="$t('git.localPath')">
          <el-input v-model="cloneForm.local_path" placeholder="/workspace/repo" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCloneDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="doClone">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showCommitDialog" :title="$t('git.commit')" width="400px">
      <el-input v-model="commitMsg" type="textarea" :rows="4" :placeholder="$t('git.commitMsg')" />
      <template #footer>
        <el-button @click="showCommitDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="doCommit">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showBranchDialog" :title="$t('git.createBranch')" width="400px">
      <el-input v-model="newBranchName" :placeholder="$t('git.branchName')" />
      <template #footer>
        <el-button @click="showBranchDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="doCreateBranch">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showAddLocalDialog" :title="$t('common.add')" width="400px">
      <el-input v-model="addLocalPath" :placeholder="$t('git.localPath')" />
      <template #footer>
        <el-button @click="showAddLocalDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="doAddLocal">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { useGitManager } from '@/composables/useGitManager'

const {
  repos, selectedRepoId, branches, commits, statusItems, stats, loading, error,
  loadRepos, addLocalRepo, cloneRepo, selectRepo, pull, push, commit, checkout, createBranch
} = useGitManager()

const showCloneDialog = ref(false)
const showCommitDialog = ref(false)
const showBranchDialog = ref(false)
const showAddLocalDialog = ref(false)
const commitMsg = ref('')
const newBranchName = ref('')
const addLocalPath = ref('')
const currentBranch = ref('')
const cloneForm = reactive({ url: '', local_path: '', username: '', password: '' })

function statusTagType(status: string) {
  if (status === 'modified') return 'warning'
  if (status === 'added') return 'success'
  if (status === 'deleted') return 'danger'
  return 'info'
}

async function doClone() {
  if (!cloneForm.url || !cloneForm.local_path) { ElMessage.warning('请填写完整信息'); return }
  await cloneRepo(cloneForm)
  showCloneDialog.value = false
  ElMessage.success('仓库克隆成功')
}

async function doCommit() {
  if (!commitMsg.value.trim()) { ElMessage.warning('请输入提交信息'); return }
  await commit(commitMsg.value)
  commitMsg.value = ''
  showCommitDialog.value = false
}

async function doCreateBranch() {
  if (!newBranchName.value.trim()) return
  await createBranch(newBranchName.value)
  newBranchName.value = ''
  showBranchDialog.value = false
}

async function doAddLocal() {
  if (!addLocalPath.value) return
  await addLocalRepo(addLocalPath.value)
  addLocalPath.value = ''
  showAddLocalDialog.value = false
}

loadRepos()
</script>

<style scoped>
.code-repo { padding: 0; height: calc(100vh - var(--header-height) - 60px); display: flex; flex-direction: column; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; flex-shrink: 0; }
.page-header h2 { font-size: 22px; font-weight: 600; }
.header-actions { display: flex; gap: 10px; }

.repo-layout { flex: 1; display: flex; gap: 0; min-height: 0; overflow: hidden; }

.repo-list-panel { width: 280px; border-right: 1px solid var(--border-color); background: var(--bg-primary); overflow-y: auto; }
.repo-cards { padding: 8px; }
.repo-card { padding: 12px; border-radius: 8px; cursor: pointer; margin-bottom: 6px; border: 1px solid transparent; }
.repo-card.active { border-color: var(--accent-color); background: rgba(64,158,255,0.05); }
.repo-card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 4px; }
.repo-name { font-weight: 600; display: flex; align-items: center; gap: 8px; font-size: 14px; }
.repo-path { font-size: 12px; color: var(--text-secondary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.repo-stats { font-size: 12px; color: var(--accent-color); margin-top: 4px; }

.repo-detail { flex: 1; display: flex; flex-direction: column; padding: 0 16px; min-width: 0; overflow: hidden; }
.detail-toolbar { display: flex; gap: 6px; padding: 8px 0; flex-shrink: 0; }

.detail-panels { flex: 1; display: grid; grid-template-columns: 1fr 1fr; gap: 16px; min-height: 0; overflow: hidden; }
.status-panel, .log-panel { background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 8px; padding: 12px; overflow-y: auto; }
.status-panel h4, .log-panel h4 { margin-bottom: 10px; font-size: 14px; }
.status-item { display: flex; align-items: center; gap: 8px; padding: 4px 0; }
.status-list { font-size: 13px; }
.log-list { font-size: 13px; }
.log-item { padding: 6px 0; border-bottom: 1px solid var(--border-color); }
.log-hash { font-family: monospace; font-size: 12px; color: var(--accent-color); }
.log-msg { font-weight: 500; margin: 2px 0; }
.log-meta { font-size: 11px; color: var(--text-secondary); }
.no-status { text-align: center; padding: 30px; color: var(--text-secondary); }
.no-selection { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; color: var(--text-secondary); gap: 12px; }
</style>
