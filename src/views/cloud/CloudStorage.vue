<template>
  <div class="cloud-storage">
    <div class="page-header">
      <h2>云存储管理</h2>
      <div class="header-actions">
        <el-button type="primary" @click="showProviderDialog = true">
          <el-icon><Plus /></el-icon> 添加存储
        </el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab">
      <el-tab-pane label="阿里云 OSS" name="oss">
        <CloudProvider type="oss" :files="ossFiles" />
      </el-tab-pane>
      <el-tab-pane label="腾讯云 COS" name="cos">
        <CloudProvider type="cos" :files="cosFiles" />
      </el-tab-pane>
      <el-tab-pane label="七牛云" name="qiniu">
        <CloudProvider type="qiniu" :files="qiniuFiles" />
      </el-tab-pane>
      <el-tab-pane label="MinIO" name="minio">
        <CloudProvider type="minio" :files="minioFiles" />
      </el-tab-pane>
      <el-tab-pane label="AWS S3" name="s3">
        <CloudProvider type="s3" :files="s3Files" />
      </el-tab-pane>
      <el-tab-pane label="备份任务" name="backup">
        <div class="backup-section">
          <div class="backup-header">
            <h4>定时备份任务</h4>
            <el-button type="primary" size="small" @click="showBackupDialog = true">
              <el-icon><Plus /></el-icon> 新建任务
            </el-button>
          </div>
          <el-table :data="backupTasks">
            <el-table-column prop="name" label="任务名称" min-width="150" />
            <el-table-column prop="source" label="源路径" min-width="200" />
            <el-table-column prop="target" label="目标存储" width="150" />
            <el-table-column prop="schedule" label="执行计划" width="130" />
            <el-table-column prop="lastRun" label="上次执行" width="160" />
            <el-table-column prop="status" label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.status === 'success' ? 'success' : 'info'" size="small">
                  {{ row.status === 'success' ? '成功' : '等待中' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="160">
              <template #default="{ row }">
                <el-button size="small" @click="runBackup(row)">立即执行</el-button>
                <el-button size="small" type="danger" @click="deleteBackup(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="showBackupDialog" title="新建备份任务" width="500px">
      <el-form :model="newBackup" label-width="100px">
        <el-form-item label="任务名称">
          <el-input v-model="newBackup.name" placeholder="例: 每日数据库备份" />
        </el-form-item>
        <el-form-item label="源路径">
          <el-input v-model="newBackup.source" placeholder="/data/database" />
        </el-form-item>
        <el-form-item label="目标存储">
          <el-select v-model="newBackup.target">
            <el-option label="阿里云 OSS" value="阿里云 OSS" />
            <el-option label="腾讯云 COS" value="腾讯云 COS" />
            <el-option label="MinIO" value="MinIO" />
          </el-select>
        </el-form-item>
        <el-form-item label="执行计划">
          <el-select v-model="newBackup.schedule">
            <el-option label="每天 03:00" value="0 3 * * *" />
            <el-option label="每周一 03:00" value="0 3 * * 1" />
            <el-option label="每小时" value="0 * * * *" />
            <el-option label="每6小时" value="0 */6 * * *" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showBackupDialog = false">取消</el-button>
        <el-button type="primary" @click="addBackupTask">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { v4 as uuid } from 'uuid'
import CloudProvider from './CloudProvider.vue'

const activeTab = ref('oss')
const showBackupDialog = ref(false)
const showProviderDialog = ref(false)

const ossFiles = ref([
  { name: 'backups/', isDir: true, size: '-', modified: '2026-06-29' },
  { name: 'images/', isDir: true, size: '-', modified: '2026-06-28' },
  { name: 'database.sql.gz', isDir: false, size: '256MB', modified: '2026-06-29' },
  { name: 'config.json', isDir: false, size: '2.3KB', modified: '2026-06-25' }
])

const cosFiles = ref([
  { name: 'website/', isDir: true, size: '-', modified: '2026-06-29' },
  { name: 'index.html', isDir: false, size: '15KB', modified: '2026-06-29' }
])

const qiniuFiles = ref([
  { name: 'media/', isDir: true, size: '-', modified: '2026-06-20' }
])

const minioFiles = ref([
  { name: 'docker-data/', isDir: true, size: '-', modified: '2026-06-28' },
  { name: 'backup.tar.gz', isDir: false, size: '1.2GB', modified: '2026-06-29' }
])

const s3Files = ref([
  { name: 'static/', isDir: true, size: '-', modified: '2026-06-15' },
  { name: 'reports/', isDir: true, size: '-', modified: '2026-06-10' }
])

const backupTasks = ref([
  { id: '1', name: '每日数据库备份', source: '/data/mysql/backups', target: '阿里云 OSS', schedule: '每天 03:00', lastRun: '2026-06-29 03:00', status: 'success' },
  { id: '2', name: '每周配置备份', source: '/etc/nginx', target: 'MinIO', schedule: '每周一 05:00', lastRun: '2026-06-23 05:00', status: 'success' },
  { id: '3', name: '七日杀存档备份', source: '/home/steam/7days/Saves', target: '腾讯云 COS', schedule: '每6小时', lastRun: '2026-06-29 12:00', status: 'success' }
])

const newBackup = reactive({ name: '', source: '', target: '', schedule: '0 3 * * *' })

function addBackupTask() {
  if (!newBackup.name || !newBackup.source) {
    ElMessage.warning('请填写完整信息')
    return
  }
  backupTasks.value.push({
    id: uuid(), name: newBackup.name, source: newBackup.source,
    target: newBackup.target, schedule: newBackup.schedule,
    lastRun: '-', status: 'pending'
  })
  showBackupDialog.value = false
  ElMessage.success('备份任务创建成功')
}

function runBackup(task: any) { ElMessage.success(`备份任务 "${task.name}" 已启动`) }
function deleteBackup(task: any) { backupTasks.value = backupTasks.value.filter(t => t.id !== task.id); ElMessage.success('备份任务已删除') }
</script>

<style scoped>
.cloud-storage { padding: 0; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { font-size: 22px; font-weight: 600; }
.backup-section { padding: 16px 0; }
.backup-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.backup-header h4 { font-size: 15px; }
</style>
