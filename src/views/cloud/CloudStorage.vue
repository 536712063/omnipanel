<template>
  <div class="cloud-storage">
    <div class="page-header">
      <h2>{{ $t('cloud.title') }}</h2>
      <div class="header-actions">
        <el-button type="primary" @click="showProviderDialog = true">
          <el-icon><Plus /></el-icon> {{ $t('cloud.addProvider') }}
        </el-button>
      </div>
    </div>

    <div class="cloud-layout">
      <div class="cloud-sidebar">
        <el-menu class="provider-menu" :default-active="provider" @select="switchProvider">
          <el-menu-item v-for="p in providers" :key="p" :index="p">
            <el-icon><FolderOpened /></el-icon>
            <span>{{ $t(`cloud.providers.${p}`, p) }}</span>
          </el-menu-item>
        </el-menu>

        <div class="quota-info" v-if="preview">
          <div class="quota-item">
            <span>{{ $t('common.detail') }}</span>
            <span class="quota-val">{{ preview.type }} / {{ preview.size }}B</span>
          </div>
        </div>
      </div>

      <div class="cloud-main">
        <div class="path-bar">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item v-for="(seg, i) in pathSegments" :key="i" @click="navigate(seg.path)">
              {{ seg.name || $t(`cloud.providers.${provider}`, provider) }}
            </el-breadcrumb-item>
          </el-breadcrumb>
          <div class="path-actions">
            <el-button size="small" @click="loadFiles" :loading="loading">{{ $t('common.refresh') }}</el-button>
            <el-button size="small" @click="showMkdirPrompt = true">{{ $t('cloud.mkdir') }}</el-button>
          </div>
        </div>

        <el-table v-loading="loading" :data="files" size="small" @row-dblclick="onFileDblClick">
          <el-table-column prop="name" :label="$t('cloud.fileList')" min-width="250">
            <template #default="{ row }">
              <div class="file-cell" @click="onFileDblClick(row)">
                <el-icon v-if="row.is_dir" style="color: #e6a23c"><Folder /></el-icon>
                <el-icon v-else style="color: #909399"><Document /></el-icon>
                <span class="file-name">{{ row.name }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="size" :label="$t('common.detail')" width="100">
            <template #default="{ row }">{{ row.is_dir ? '-' : formatSize(row.size) }}</template>
          </el-table-column>
          <el-table-column prop="modified_at" :label="$t('git.log')" width="160">
            <template #default="{ row }">{{ row.modified_at?.slice(0, 16) }}</template>
          </el-table-column>
          <el-table-column :label="$t('common.edit')" width="200">
            <template #default="{ row }">
              <el-button size="small" text @click="getPreview(row.path)" v-if="!row.is_dir">{{ $t('cloud.preview') }}</el-button>
              <el-button size="small" text @click="downloadFile(row.path, '/tmp/' + row.name)">{{ $t('common.download') }}</el-button>
              <el-button size="small" text type="danger" @click="deleteFile(row.path)">{{ $t('common.delete') }}</el-button>
            </template>
          </el-table-column>
        </el-table>

        <div v-if="!files.length && !loading" class="empty-dir">
          <p>{{ $t('cloud.fileList') }}</p>
        </div>
      </div>
    </div>

    <div class="preview-panel" v-if="preview && previewPath">
      <div class="preview-header">
        <h4>{{ previewPath }}</h4>
        <el-button size="small" text @click="preview = null; previewPath = ''"><el-icon><Close /></el-icon></el-button>
      </div>
      <div class="preview-content">
        <img v-if="isImagePreview" :src="preview.url" class="preview-image" />
        <pre v-else>预览: {{ preview.type }} ({{ preview.size }}B)</pre>
      </div>
    </div>

    <el-dialog v-model="showProviderDialog" :title="$t('cloud.addProvider')" width="500px">
      <el-form :model="newProvider" label-width="100px">
        <el-form-item :label="$t('config.provider')">
          <el-input v-model="newProvider.name" placeholder="alist" />
        </el-form-item>
        <el-form-item :label="$t('config.endpoint')">
          <el-input v-model="newProvider.base_url" placeholder="http://localhost:5244" />
        </el-form-item>
        <el-form-item :label="$t('config.token')">
          <el-input v-model="newProvider.token" type="password" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showProviderDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="saveProvider">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showMkdirPrompt" :title="$t('cloud.mkdir')" width="400px">
      <el-input v-model="newDirName" placeholder="new_folder" @keydown.enter="handleMkdir" />
      <template #footer>
        <el-button @click="showMkdirPrompt = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleMkdir">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useCloudStorage } from '@/composables/useCloudStorage'

const {
  provider, currentPath, files, loading, error, preview, previewPath, providers,
  loadFiles, loadProviders, navigate, goUp, copyFile, moveFile, renameFile, deleteFile, mkdir,
  uploadFile, downloadFile, getPreview, syncFromMachine, addProvider, oauthURL, subscribeProgress
} = useCloudStorage('alist')

const showProviderDialog = ref(false)
const showMkdirPrompt = ref(false)
const newDirName = ref('')
const newProvider = reactive({ name: 'alist', type: 'alist', base_url: '', token: '', username: '', password: '', client_id: '', client_secret: '', redirect_uri: '' })

const pathSegments = computed(() => {
  const parts = currentPath.value.split('/').filter(Boolean)
  const segs: { name: string; path: string }[] = []
  let acc = ''
  for (const p of parts) {
    acc += '/' + p
    segs.push({ name: decodeURIComponent(p), path: acc })
  }
  return segs
})

const isImagePreview = computed(() =>
  preview.value?.mime_type?.startsWith('image/')
)

function switchProvider(p: string) {
  provider.value = p
  loadFiles()
}

function onFileDblClick(row: any) {
  if (row.is_dir) {
    navigate(currentPath.value.replace(/\/+$/, '') + '/' + row.name)
  }
}

function handleMkdir() {
  if (newDirName.value.trim()) {
    mkdir(newDirName.value.trim())
    newDirName.value = ''
    showMkdirPrompt.value = false
  }
}

function saveProvider() {
  if (!newProvider.name || !newProvider.base_url) {
    ElMessage.warning('请填写必要信息')
    return
  }
  addProvider({ ...newProvider })
  showProviderDialog.value = false
  ElMessage.success('存储已添加')
}

function formatSize(bytes: number): string {
  if (!bytes) return '-'
  if (bytes < 1024) return bytes + 'B'
  if (bytes < 1048576) return (bytes / 1024).toFixed(1) + 'KB'
  if (bytes < 1073741824) return (bytes / 1048576).toFixed(1) + 'MB'
  return (bytes / 1073741824).toFixed(1) + 'GB'
}

loadProviders()
loadFiles()
subscribeProgress()
</script>

<style scoped>
.cloud-storage { padding: 0; height: calc(100vh - var(--header-height) - 60px); display: flex; flex-direction: column; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; flex-shrink: 0; }
.page-header h2 { font-size: 22px; font-weight: 600; }
.header-actions { display: flex; gap: 10px; }

.cloud-layout { flex: 1; display: flex; gap: 0; min-height: 0; overflow: hidden; }

.cloud-sidebar { width: 200px; border-right: 1px solid var(--border-color); background: var(--bg-primary); display: flex; flex-direction: column; }
.provider-menu { border-right: none !important; flex: 1; }
.quota-info { padding: 12px; border-top: 1px solid var(--border-color); font-size: 12px; }
.quota-item { display: flex; justify-content: space-between; margin-bottom: 4px; }
.quota-val { color: var(--text-secondary); }

.cloud-main { flex: 1; display: flex; flex-direction: column; min-width: 0; overflow: hidden; padding: 0 16px; }

.path-bar { display: flex; justify-content: space-between; align-items: center; padding: 8px 0; margin-bottom: 8px; flex-shrink: 0; }
.path-actions { display: flex; gap: 6px; }

.file-cell { display: flex; align-items: center; gap: 8px; cursor: pointer; }
.file-name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.empty-dir { text-align: center; padding: 60px; color: var(--text-secondary); }

.preview-panel { border-top: 1px solid var(--border-color); padding: 12px 16px; background: var(--bg-primary); flex-shrink: 0; max-height: 200px; overflow-y: auto; }
.preview-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.preview-content { font-size: 13px; }
.preview-image { max-width: 100%; max-height: 150px; border-radius: 6px; }
</style>
