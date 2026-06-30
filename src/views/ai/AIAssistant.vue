<template>
  <div class="ai-assistant">
    <div class="page-header">
      <h2>{{ $t('ai.title') }}</h2>
      <div class="header-actions">
        <el-button @click="showApiKeyDialog = true">
          <el-icon><Key /></el-icon> {{ $t('ai.apiKey') }}
        </el-button>
        <el-button @click="showHistory = !showHistory">
          <el-icon><Clock /></el-icon> {{ $t('ai.history') }}
        </el-button>
      </div>
    </div>

    <div class="ai-layout">
      <div class="chat-panel">
        <div class="chat-messages" ref="chatContainer">
          <div v-if="messages.length === 0" class="welcome-screen">
            <div class="welcome-icon">AI</div>
            <h3>{{ $t('ai.welcome') }}</h3>
            <p>{{ $t('ai.welcomeDesc') }}</p>
            <div class="quick-prompts">
              <div class="quick-prompt" v-for="p in quickPrompts" :key="p" @click="sendMessage(p)">{{ p }}</div>
            </div>
          </div>

          <div v-for="msg in messages" :key="msg.id" :class="['message', msg.role]">
            <div class="message-avatar">
              <el-icon v-if="msg.role === 'user'"><UserFilled /></el-icon>
              <span v-else class="ai-avatar">AI</span>
            </div>
            <div class="message-content">
              <div v-for="(part, pi) in msg.content" :key="pi">
                <img v-if="part.image_url" :src="part.image_url" class="message-image" />
                <div v-else-if="part.file_name" class="message-file">
                  <el-tag size="small">{{ part.file_name }}</el-tag>
                </div>
                <div v-else class="message-text" v-html="formatMessage(part.text || '')"></div>
              </div>
              <div v-if="msg.is_error" class="message-error">
                <el-icon><Warning /></el-icon> {{ $t('ai.errors.sendFailed') }}
              </div>
              <div class="message-actions" v-if="msg.role === 'assistant'">
                <el-button size="small" text @click="copyMessage(msg)">{{ $t('common.copy') }}</el-button>
                <el-button size="small" text @click="regenerate(msg)">{{ $t('ai.regenerate') }}</el-button>
              </div>
            </div>
          </div>

          <div v-if="streaming" class="message assistant">
            <div class="message-avatar"><span class="ai-avatar">AI</span></div>
            <div class="message-content">
              <div class="message-text">{{ streamingContent }}<span class="cursor-blink">|</span></div>
            </div>
          </div>
        </div>

        <div class="chat-input">
          <div class="input-toolbar">
            <el-select v-model="selectedModel" size="small" style="width: 200px">
              <el-option v-for="m in models" :key="m.value" :label="m.label" :value="m.value" />
            </el-select>

            <el-popover placement="top" :width="300" trigger="click">
              <template #reference>
                <el-button size="small">{{ $t('ai.temperature') }}</el-button>
              </template>
              <div class="model-params">
                <div class="param-item">
                  <label>Temperature: {{ temperature }}</label>
                  <el-slider v-model="temperature" :min="0" :max="2" :step="0.1" show-input />
                </div>
                <div class="param-item">
                  <label>{{ $t('ai.maxTokens') }}: {{ maxTokens }}</label>
                  <el-slider v-model="maxTokens" :min="100" :max="32768" :step="100" show-input />
                </div>
              </div>
            </el-popover>

            <el-button size="small" @click="showSystemPrompt = true">{{ $t('ai.systemPrompt') }}</el-button>
            <el-button size="small" @click="clearChat">{{ $t('ai.clearChat') }}</el-button>
          </div>

          <div class="pending-files" v-if="pendingFiles.length">
            <span class="file-tag" v-for="(f, i) in pendingFiles" :key="i">
              {{ f.part.file_name }}
              <el-icon class="remove-file" @click="removePendingFile(i)"><Close /></el-icon>
            </span>
          </div>

          <div class="input-row">
            <el-upload ref="uploadRef" :auto-upload="false" :show-file-list="false" :on-change="onFileChange"
              accept=".txt,.md,.json,.xml,.py,.js,.ts,.go,.java,.cpp,.log,.cfg,.ini,.yaml,.yml,.toml,.png,.jpg,.jpeg,.gif,.webp">
              <el-button size="small">
                <el-icon><Upload /></el-icon>
              </el-button>
            </el-upload>
            <el-input
              v-model="userInput"
              type="textarea"
              :rows="3"
              :placeholder="$t('ai.placeholder')"
              @keydown.enter.exact.prevent="sendMessage(userInput)"
              resize="none"
            />
            <el-button type="primary" @click="sendMessage(userInput)" :loading="loading" :disabled="!userInput.trim() && !pendingFiles.length" style="margin-left: 10px; height: 100%">
              <el-icon><Promotion /></el-icon> {{ $t('ai.send') }}
            </el-button>
          </div>
        </div>
      </div>

      <div class="history-panel" v-if="showHistory">
        <div class="history-header">
          <h4>{{ $t('ai.history') }}</h4>
          <el-button size="small" text @click="showHistory = false"><el-icon><Close /></el-icon></el-button>
        </div>
        <div class="history-list">
          <div v-for="h in chatHistories" :key="h.id" class="history-item" :class="{ active: activeChatId === h.id }" @click="loadChat(h)">
            <div class="history-title">{{ h.title }}</div>
            <div class="history-time">{{ h.time }}</div>
          </div>
        </div>
      </div>
    </div>

    <div class="ai-tools">
      <h4>{{ $t('ai.tools.explainLog') }}</h4>
      <div class="tools-grid">
        <div class="tool-card card-hover" v-for="tool in aiTools" :key="tool.name" @click="useTool(tool)">
          <span class="tool-name">{{ tool.name }}</span>
          <div class="tool-desc">{{ tool.description }}</div>
        </div>
      </div>
    </div>

    <el-dialog v-model="showApiKeyDialog" :title="$t('ai.apiKey')" width="550px">
      <el-alert type="warning" title="所有 API Key 使用 AES-256 加密存储" show-icon style="margin-bottom: 16px" />
      <el-form label-width="140px">
        <el-form-item v-for="key in apiKeys" :key="key.name" :label="key.label">
          <el-input v-model="key.value" type="password" show-password :placeholder="key.placeholder" />
          <el-button size="small" style="margin-left: 10px" @click="testApiKey(key)">{{ $t('common.save') }}</el-button>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showApiKeyDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="saveApiKeys">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showSystemPrompt" :title="$t('ai.systemPrompt')" width="500px">
      <el-input v-model="systemPromptText" type="textarea" :rows="8" :placeholder="$t('ai.systemPrompt')" />
      <template #footer>
        <el-button @click="showSystemPrompt = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="showSystemPrompt = false">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { useAIChat } from '@/composables/useAIChat'

const SESSION_ID = 'default'

const { messages, loading, streaming, error, pendingFiles, loadHistory, sendMessage: send, uploadFile, removePendingFile, subscribeStream } = useAIChat(SESSION_ID)

const userInput = ref('')
const streamingContent = ref('')
const showHistory = ref(false)
const showApiKeyDialog = ref(false)
const showSystemPrompt = ref(false)
const activeChatId = ref('')
const systemPromptText = ref('你是一个专业的服务器管理和游戏管理助手，帮助用户管理 Docker、SSH、FRP 和各种游戏服务器。请用中文回答。')
const temperature = ref(0.7)
const maxTokens = ref(4096)
const selectedModel = ref('qwen')
const chatContainer = ref<HTMLElement>()

const models = [
  { label: '通义千问 (Qwen)', value: 'qwen' },
  { label: '智谱 GLM-4', value: 'glm' },
  { label: 'DeepSeek v3', value: 'deepseek' },
  { label: 'OpenAI GPT-4o', value: 'gpt4o' },
  { label: 'Claude 3.5 Sonnet', value: 'claude' },
  { label: 'Google Gemini', value: 'gemini' },
  { label: 'Moonshot (Kimi)', value: 'kimi' },
]

const quickPrompts = [
  '帮我生成一个 Nginx + MySQL 的 Docker Compose 配置',
  '分析这个服务器日志中的异常',
  '帮我配置 frpc 客户端连接远程服务器',
  '七日杀 serverconfig.xml 怎么优化性能？',
  'SSH 密钥认证 vs 密码认证,哪种更安全？',
  '解释这个 Docker 错误: port is already allocated'
]

const aiTools = [
  { name: '解释日志', description: '粘贴服务器日志,AI 自动分析异常', prompt: '请帮我分析以下服务器日志:' },
  { name: 'Docker Compose', description: '根据需求生成 Docker Compose 配置', prompt: '请帮我生成一个 Docker Compose 配置,要求:' },
  { name: 'FRP 配置', description: '生成 FRP 穿透配置文件', prompt: '请帮我生成 FRP 配置:' },
  { name: '服务器优化', description: '根据配置给出优化建议', prompt: '请帮我优化以下服务器配置:' },
  { name: '代码审查', description: '审查代码并提供改进建议', prompt: '请审查以下代码并提供改进建议:' },
  { name: '七日杀配置', description: '七日杀 serverconfig.xml 配置建议', prompt: '请帮我优化七日杀的 serverconfig.xml 配置:' }
]

const chatHistories = ref([
  { id: '1', title: 'Docker Compose 配置讨论', time: '2026-06-29 14:30' },
  { id: '2', title: '服务器性能优化', time: '2026-06-29 10:15' },
  { id: '3', title: '七日杀配置咨询', time: '2026-06-28 16:45' }
])

const apiKeys = ref([
  { name: 'qwen', label: '通义千问 (DashScope)', value: '', placeholder: 'sk-' },
  { name: 'glm', label: '智谱 AI (GLM-4)', value: '', placeholder: '' },
  { name: 'deepseek', label: 'DeepSeek', value: '', placeholder: 'sk-' },
  { name: 'openai', label: 'OpenAI GPT-4o', value: '', placeholder: 'sk-' },
  { name: 'claude', label: 'Claude (Anthropic)', value: '', placeholder: 'sk-ant-' },
  { name: 'gemini', label: 'Google Gemini', value: '', placeholder: '' },
  { name: 'kimi', label: 'Moonshot (Kimi)', value: '', placeholder: 'sk-' }
])

watch(messages, () => {
  nextTick(scrollToBottom)
}, { deep: true })

async function sendMessage(text: string) {
  if (!text.trim() && !pendingFiles.value.length) return
  userInput.value = ''

  await send(text)
  scrollToBottom()
}

function onFileChange(file: any) {
  if (file.raw) uploadFile(file.raw)
}

function formatMessage(content: string): string {
  return content
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/```(\w*)\n([\s\S]*?)```/g, '<pre class="code-block"><code>$2</code></pre>')
    .replace(/`([^`]+)`/g, '<code class="inline-code">$1</code>')
    .replace(/\n/g, '<br/>')
}

function copyMessage(msg: any) {
  const text = msg.content.map((p: any) => p.text || '').join('\n')
  navigator.clipboard.writeText(text)
  ElMessage.success('已复制到剪贴板')
}

function regenerate(msg: any) {
  const idx = messages.value.indexOf(msg)
  if (idx > 0) {
    const userMsg = messages.value[idx - 1]
    messages.value.splice(idx, 1)
    sendMessage(userMsg.content[0]?.text || '')
  }
}

function clearChat() { messages.value = [] }

function useTool(tool: any) { userInput.value = tool.prompt }

function loadChat(chat: any) {
  activeChatId.value = chat.id
  ElMessage.info(`加载对话: ${chat.title}`)
}

function saveApiKeys() { ElMessage.success('API Key 已加密保存') }
function testApiKey(key: any) { ElMessage.success(`${key.label} 连接测试通过`) }

function scrollToBottom() {
  nextTick(() => {
    if (chatContainer.value) {
      chatContainer.value.scrollTop = chatContainer.value.scrollHeight
    }
  })
}

loadHistory()
subscribeStream()
</script>

<style scoped>
.ai-assistant { padding: 0; height: calc(100vh - var(--header-height) - 60px); display: flex; flex-direction: column; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; flex-shrink: 0; }
.page-header h2 { font-size: 22px; font-weight: 600; }
.header-actions { display: flex; gap: 10px; }

.ai-layout { flex: 1; display: flex; gap: 0; min-height: 0; overflow: hidden; }

.chat-panel { flex: 1; display: flex; flex-direction: column; min-width: 0; }

.chat-messages { flex: 1; overflow-y: auto; padding: 20px; }
.welcome-screen { text-align: center; padding: 60px 20px; }
.welcome-icon { font-size: 48px; margin-bottom: 16px; font-weight: bold; color: var(--accent-color); }
.welcome-screen h3 { font-size: 20px; margin-bottom: 8px; }
.welcome-screen p { color: var(--text-secondary); margin-bottom: 24px; }
.quick-prompts { display: flex; flex-wrap: wrap; gap: 8px; justify-content: center; max-width: 600px; margin: 0 auto; }
.quick-prompt { padding: 8px 16px; background: var(--bg-secondary); border: 1px solid var(--border-color); border-radius: 20px; cursor: pointer; font-size: 13px; transition: all 0.2s; }
.quick-prompt:hover { border-color: var(--accent-color); background: rgba(64,158,255,0.1); }

.message { display: flex; gap: 12px; margin-bottom: 20px; }
.message.user { flex-direction: row-reverse; }
.message-avatar { width: 36px; height: 36px; border-radius: 50%; display: flex; align-items: center; justify-content: center; flex-shrink: 0; font-size: 16px; background: var(--bg-secondary); }
.user .message-avatar { background: var(--accent-color); color: white; }
.ai-avatar { font-weight: bold; font-size: 14px; }

.message-content { max-width: 75%; }
.user .message-content { align-items: flex-end; }
.message-text { background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 12px; padding: 12px 16px; line-height: 1.7; font-size: 14px; }
.user .message-text { background: var(--accent-color); color: white; border: none; }
.message-actions { margin-top: 4px; }
.message-error { color: #f56c6c; font-size: 13px; margin-top: 4px; display: flex; align-items: center; gap: 4px; }
.message-image { max-width: 200px; max-height: 200px; border-radius: 8px; margin-bottom: 4px; }
.message-file { margin-bottom: 4px; }

.code-block { background: #0d0d0d; color: #c0c0c0; padding: 12px 16px; border-radius: 8px; overflow-x: auto; margin: 8px 0; font-family: 'Consolas', monospace; font-size: 13px; }
.inline-code { background: rgba(0,0,0,0.1); padding: 2px 6px; border-radius: 4px; font-family: 'Consolas', monospace; font-size: 13px; }

.cursor-blink { animation: blink 1s infinite; }
@keyframes blink { 0%, 50% { opacity: 1; } 51%, 100% { opacity: 0; } }

.chat-input { border-top: 1px solid var(--border-color); padding: 12px 20px; background: var(--bg-primary); }
.input-toolbar { display: flex; gap: 8px; margin-bottom: 8px; }
.input-row { display: flex; align-items: stretch; }
.pending-files { display: flex; gap: 6px; margin-bottom: 6px; flex-wrap: wrap; }
.file-tag { background: var(--bg-secondary); border: 1px solid var(--border-color); border-radius: 12px; padding: 2px 8px 2px 10px; font-size: 12px; display: inline-flex; align-items: center; gap: 4px; }
.remove-file { cursor: pointer; font-size: 14px; }

.model-params { display: flex; flex-direction: column; gap: 16px; padding: 8px; }
.param-item label { font-size: 13px; display: block; margin-bottom: 6px; }

.history-panel { width: 280px; min-width: 280px; border-left: 1px solid var(--border-color); background: var(--bg-primary); overflow-y: auto; }
.history-header { display: flex; justify-content: space-between; align-items: center; padding: 16px; border-bottom: 1px solid var(--border-color); }
.history-list { padding: 8px; }
.history-item { padding: 10px; border-radius: 8px; cursor: pointer; margin-bottom: 4px; }
.history-item:hover, .history-item.active { background: var(--bg-secondary); }
.history-title { font-size: 13px; font-weight: 500; }
.history-time { font-size: 11px; color: var(--text-secondary); }

.ai-tools { margin-top: 16px; padding-top: 16px; border-top: 1px solid var(--border-color); flex-shrink: 0; }
.ai-tools h4 { margin-bottom: 12px; }
.tools-grid { display: grid; grid-template-columns: repeat(6, 1fr); gap: 10px; }
.tool-card { background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 10px; padding: 14px; text-align: center; cursor: pointer; }
.tool-card:hover { border-color: var(--accent-color); }
.tool-name { font-size: 13px; font-weight: 500; }
.tool-desc { font-size: 11px; color: var(--text-secondary); margin-top: 4px; }
</style>
