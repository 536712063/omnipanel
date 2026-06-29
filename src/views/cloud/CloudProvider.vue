<template>
  <div class="cloud-provider">
    <el-alert
      v-if="!configured"
      :title="`${providerNames[type]} 未配置`"
      type="warning"
      :description="`请先配置 ${providerNames[type]} 的访问密钥`"
      show-icon
      closable
    >
      <template #default>
        <el-button size="small" type="primary" @click="$emit('configure')">立即配置</el-button>
      </template>
    </el-alert>

    <div class="provider-toolbar" v-if="configured">
      <div class="toolbar-left">
        <el-button type="primary" size="small" @click="showUpload = true">
          <el-icon><Upload /></el-icon> 上传文件
        </el-button>
        <el-button size="small" @click="showFolderCreate = true">
          <el-icon><FolderAdd /></el-icon> 新建文件夹
        </el-button>
        <el-button size="small" @click="refreshFiles">
          <el-icon><Refresh /></el-icon> 刷新
        </el-button>
        <el-input v-model="searchQuery" placeholder="搜索文件..." size="small" clearable style="width: 200px; margin-left: 10px" />
      </div>
      <div class="toolbar-right">
        <span class="storage-usage">
          已使用 {{ usedSpace }} / {{ totalSpace }}
        </span>
        <el-progress :percentage="usagePercent" :stroke-width="6" style="width: 150px; margin-left: 10px" />
      </div>
    </div>

    <el-table :data="filteredFiles" size="small" v-if="configured">
      <el-table-column label="文件名" min-width="250">
        <template #default="{ row }">
          <div class="file-name-cell" @click="row.isDir && navigateTo(row)" style="cursor: pointer">
            <el-icon :size="18">
              <Folder v-if="row.isDir" />
              <Document v-else />
            </el-icon>
            <span>{{ row.name }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="size" label="大小" width="120" />
      <el-table-column prop="modified" label="修改时间" width="180" />
      <el-table-column label="操作" width="240" fixed="right">
        <template #default="{ row }">
          <el-button size="small" text @click="downloadFile(row)">下载</el-button>
          <el-button size="small" text @click="shareFile(row)">分享</el-button>
          <el-button size="small" text @click="renameFile(row)">重命名</el-button>
          <el-button size="small" text type="danger" @click="deleteFile(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="showShareDialog" title="文件分享" width="450px">
      <div class="share-info" v-if="shareLink">
        <p>分享链接 (24小时有效):</p>
        <el-input v-model="shareLink" readonly>
          <template #append>
            <el-button @click="copyShareLink">复制</el-button>
          </template>
        </el-input>
      </div>
      <div v-else>
        <el-form label-width="80px">
          <el-form-item label="有效期">
            <el-select v-model="shareExpiry">
              <el-option label="1小时" :value="3600" />
              <el-option label="24小时" :value="86400" />
              <el-option label="7天" :value="604800" />
              <el-option label="永久" :value="0" />
            </el-select>
          </el-form-item>
        </el-form>
        <el-button type="primary" @click="generateShareLink">生成链接</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

const props = defineProps<{ type: string; files: any[] }>()

const configured = ref(true)
const showUpload = ref(false)
const showFolderCreate = ref(false)
const showShareDialog = ref(false)
const shareLink = ref('')
const shareExpiry = ref(86400)
const searchQuery = ref('')
const usedSpace = ref('12.5GB')
const totalSpace = ref('100GB')

const providerNames: Record<string, string> = {
  oss: '阿里云 OSS', cos: '腾讯云 COS', qiniu: '七牛云', minio: 'MinIO', s3: 'AWS S3'
}

const usagePercent = computed(() => Math.round((12.5 / 100) * 100))

const filteredFiles = computed(() =>
  props.files.filter(f => f.name.toLowerCase().includes(searchQuery.value.toLowerCase()))
)

function navigateTo(row: any) { ElMessage.info(`进入目录: ${row.name}`) }
function refreshFiles() { ElMessage.success('文件列表已刷新') }
function downloadFile(row: any) { ElMessage.success(`下载 ${row.name} (模拟)`) }
function renameFile(row: any) { ElMessage.info(`重命名 ${row.name} (模拟)`) }
function deleteFile(row: any) {
  ElMessageBox.confirm(`确定删除 "${row.name}" 吗？`, '确认删除', { type: 'warning' })
    .then(() => ElMessage.success(`${row.name} 已删除`)).catch(() => {})
}
function shareFile(row: any) { showShareDialog.value = true }
function generateShareLink() { shareLink.value = `https://${props.type}.omnipanel.local/share/${Date.now()}` }
function copyShareLink() { navigator.clipboard.writeText(shareLink.value); ElMessage.success('链接已复制') }
</script>

<style scoped>
.cloud-provider { padding: 12px 0; }
.provider-toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; flex-wrap: wrap; gap: 8px; }
.toolbar-left { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }
.toolbar-right { display: flex; align-items: center; }
.storage-usage { font-size: 12px; color: var(--text-secondary); white-space: nowrap; }
.file-name-cell { display: flex; align-items: center; gap: 8px; }
.share-info p { margin-bottom: 10px; }
</style>
