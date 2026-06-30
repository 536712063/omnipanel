import { ref } from 'vue'
import type { BrowserInfo } from '@/wails/runtime'
import * as api from '@/wails/runtime'

export function useBrowser() {
  const currentUrl = ref('https://www.google.com')
  const history = ref<string[]>([])
  const info = ref<BrowserInfo | null>(null)

  async function loadInfo() { info.value = await api.browserGetInfo() }

  async function navigate(url: string) {
    if (!url.startsWith('http')) url = 'https://' + url
    currentUrl.value = url
    history.value.push(url)
  }

  async function openExternal(url: string) {
    await api.browserOpenExternalURL(url)
  }

  function goBack() {
    if (history.value.length > 1) {
      history.value.pop()
      currentUrl.value = history.value[history.value.length - 1]
    }
  }

  return { currentUrl, history, info, loadInfo, navigate, openExternal, goBack }
}
