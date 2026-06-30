import { ref, onUnmounted } from 'vue'
import type { AIChatMessage, AIChatRequest, AIStreamEvent, AIContentPart } from '@/wails/runtime'
import { aiChatStream, aiUploadFile, aiGetHistory } from '@/wails/runtime'

export interface UploadedFileContext {
  part: AIContentPart
  isImage: boolean
  localPreviewUrl?: string
}

export function useAIChat(sessionId: string) {
  const messages = ref<AIChatMessage[]>([])
  const loading = ref(false)
  const streaming = ref(false)
  const error = ref<string | null>(null)
  const pendingFiles = ref<UploadedFileContext[]>([])

  function addMessage(role: AIChatMessage['role'], content: AIContentPart[]): AIChatMessage {
    const msg: AIChatMessage = {
      id: `${Date.now()}-${Math.random().toString(36).slice(2)}`,
      role,
      content,
      created_at: new Date().toISOString(),
    }
    messages.value.push(msg)
    return msg
  }

  function handleStreamEvent(ev: AIStreamEvent) {
    const last = messages.value[messages.value.length - 1]
    if (!last || last.role !== 'assistant') return

    switch (ev.type) {
      case 'delta':
        if (last.content[0]) {
          last.content[0].text = (last.content[0].text || '') + (ev.delta || '')
        } else {
          last.content.push({ type: 'text', text: ev.delta || '' })
        }
        break
      case 'done':
        streaming.value = false
        break
      case 'error':
        last.is_error = true
        streaming.value = false
        if (last.content[0]) {
          last.content[0].text = (last.content[0].text || '') + `\n\n[Error: ${ev.error}]`
        }
        break
    }
  }

  async function loadHistory() {
    try {
      const history = await aiGetHistory(sessionId)
      messages.value = history
    } catch (e) {
      console.error('loadHistory failed:', e)
    }
  }

  async function sendMessage(text: string) {
    if (!text.trim() && pendingFiles.value.length === 0) return

    const parts: AIContentPart[] = []

    for (const f of pendingFiles.value) {
      parts.push(f.part)
    }
    pendingFiles.value = []

    if (text.trim()) {
      parts.push({ type: 'text', text: text.trim() })
    }

    const userMsg = addMessage('user', parts)
    const assistantMsg = addMessage('assistant', [{ type: 'text', text: '' }])

    loading.value = true
    streaming.value = true
    error.value = null

    const req: AIChatRequest = {
      session_id: sessionId,
      messages: messages.value.slice(0, -1),
    }

    try {
      await aiChatStream(req)
    } catch (e: any) {
      error.value = e?.message || 'AI request failed'
      streaming.value = false
    } finally {
      loading.value = false
    }

    subscribeStream()
  }

  async function uploadFile(file: File): Promise<UploadedFileContext> {
    const data = new Uint8Array(await file.arrayBuffer())
    const part = await aiUploadFile(file.name, Array.from(data))

    const ext = file.name.split('.').pop()?.toLowerCase() || ''
    const isImage = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp'].includes(ext)

    const result: UploadedFileContext = { part, isImage }
    if (isImage) {
      result.localPreviewUrl = URL.createObjectURL(file)
    }
    pendingFiles.value.push(result)
    return result
  }

  function removePendingFile(index: number) {
    const f = pendingFiles.value[index]
    if (f?.localPreviewUrl) URL.revokeObjectURL(f.localPreviewUrl)
    pendingFiles.value.splice(index, 1)
  }

  let unsubStream: (() => void) | null = null
  function subscribeStream() {
    if (unsubStream) unsubStream()
    const handler = (ev: AIStreamEvent) => handleStreamEvent(ev)
    window.go?.main?.App && (window as any)?.EventsOn?.('ai:stream:event', handler)
    unsubStream = () => {
      (window as any)?.EventsOff?.('ai:stream:event', handler)
    }
  }

  onUnmounted(() => {
    if (unsubStream) unsubStream()
  })

  return { messages, loading, streaming, error, pendingFiles, loadHistory, sendMessage, uploadFile, removePendingFile, subscribeStream }
}
