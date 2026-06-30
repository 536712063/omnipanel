import { ref, reactive, onUnmounted } from 'vue'
import type { CloudFileInfo, CloudProgressEvent, CloudPreviewInfo, CloudProviderConfig, CloudMachineSyncReq } from '@/wails/runtime'
import * as api from '@/wails/runtime'

export function useCloudStorage(defaultProvider = 'alist') {
  const provider = ref(defaultProvider)
  const currentPath = ref('/')
  const files = ref<CloudFileInfo[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const progress = reactive<Record<string, CloudProgressEvent>>({})
  const preview = ref<CloudPreviewInfo | null>(null)
  const previewPath = ref('')
  const providers = ref<string[]>([])

  async function loadFiles() {
    loading.value = true
    try {
      files.value = await api.cloudListFiles(provider.value, currentPath.value)
    } catch (e: any) {
      error.value = e?.message || 'Load failed'
    } finally {
      loading.value = false
    }
  }

  async function loadProviders() {
    providers.value = await api.cloudListProviders()
  }

  function navigate(path: string) { currentPath.value = path; loadFiles() }
  function goUp() {
    const parts = currentPath.value.split('/').filter(Boolean)
    parts.pop()
    currentPath.value = '/' + parts.join('/')
    loadFiles()
  }

  async function copyFile(src: string, dst: string) { await api.cloudCopyFile(provider.value, src, dst); await loadFiles() }
  async function moveFile(src: string, dst: string) { await api.cloudMoveFile(provider.value, src, dst); await loadFiles() }
  async function renameFile(path: string, name: string) { await api.cloudRenameFile(provider.value, path, name); await loadFiles() }
  async function deleteFile(path: string) { await api.cloudDeleteFile(provider.value, path); await loadFiles() }
  async function mkdir(name: string) { await api.cloudMkdir(provider.value, currentPath.value + '/' + name); await loadFiles() }
  async function uploadFile(localPath: string, remoteName: string) { return api.cloudUploadFile(provider.value, localPath, remoteName) }
  async function downloadFile(filePath: string, localPath: string) { return api.cloudDownloadFile(provider.value, filePath, localPath) }

  async function getPreview(filePath: string) {
    preview.value = await api.cloudGetPreview(provider.value, filePath)
    previewPath.value = filePath
  }

  async function syncFromMachine(req: CloudMachineSyncReq) { return api.cloudSyncFromMachine(req) }
  async function addProvider(cfg: CloudProviderConfig) { await api.cloudAddProvider(cfg); await loadProviders() }
  async function oauthURL(p: string) { return api.cloudOAuthURL(p) }

  let unsubProgress: (() => void) | null = null
  function subscribeProgress() {
    const handler = (ev: CloudProgressEvent) => { progress[ev.context_id] = ev }
    ;(window as any)?.EventsOn?.('cloud:progress:event', handler)
    unsubProgress = () => { (window as any)?.EventsOff?.('cloud:progress:event', handler) }
  }

  onUnmounted(() => { if (unsubProgress) unsubProgress() })

  return { provider, currentPath, files, loading, error, progress, preview, previewPath, providers,
    loadFiles, loadProviders, navigate, goUp, copyFile, moveFile, renameFile, deleteFile, mkdir,
    uploadFile, downloadFile, getPreview, syncFromMachine, addProvider, oauthURL, subscribeProgress }
}
