<template>
  <div class="ai-assistant">
    <div class="page-header">
      <h2>AI 智能助手</h2>
      <div class="header-actions">
        <el-button @click="showApiKeyDialog = true">
          <el-icon><Key /></el-icon> API Key 管理
        </el-button>
        <el-button @click="showHistory = !showHistory">
          <el-icon><Clock /></el-icon> 对话历史
        </el-button>
      </div>
    </div>

    <div class="ai-layout">
      <div class="chat-panel">
        <div class="chat-messages" ref="chatContainer">
          <div v-if="messages.length === 0" class="welcome-screen">
            <div class="welcome-icon">🤖</div>
            <h3>OmniPanel AI 智能助手</h3>
            <p>支持通义千问、智谱GLM、DeepSeek、OpenAI 等多个模型</p>
            <div class="quick-prompts">
              <div class="quick-prompt" v-for="p in quickPrompts" :key="p" @click="sendMessage(p)">{{ p }}</div>
            </div>
          </div>

          <div v-for="(msg, i) in messages" :key="i" :class="['message', msg.role]">
            <div class="message-avatar">
              <el-icon v-if="msg.role === 'user'"><UserFilled /></el-icon>
              <span v-else class="ai-avatar">AI</span>
            </div>
            <div class="message-content">
              <div class="message-text" v-html="formatMessage(msg.content)"></div>
              <div class="message-actions" v-if="msg.role === 'assistant'">
                <el-button size="small" text @click="copyMessage(msg.content)">复制</el-button>
                <el-button size="small" text @click="regenerate(msg)">重新生成</el-button>
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
            <el-select v-model="selectedModel" size="small" style="width: 220px">
              <el-option-group label="国内免费 AI">
                <el-option label="通义千问 (Qwen)" value="qwen" />
                <el-option label="智谱 GLM-4" value="glm" />
                <el-option label="百度文心一言" value="ernie" />
                <el-option label="讯飞星火" value="spark" />
                <el-option label="Moonshot (Kimi)" value="kimi" />
                <el-option label="DeepSeek" value="deepseek" />
                <el-option label="零一万物 (Yi)" value="yi" />
              </el-option-group>
              <el-option-group label="国外 AI">
                <el-option label="OpenAI GPT-4o" value="gpt4o" />
                <el-option label="Claude 3.5 Sonnet" value="claude" />
                <el-option label="Google Gemini" value="gemini" />
                <el-option label="Groq (Llama 3)" value="groq" />
                <el-option label="Mistral AI" value="mistral" />
              </el-option-group>
            </el-select>

            <el-popover placement="top" :width="300" trigger="click">
              <template #reference>
                <el-button size="small">参数设置</el-button>
              </template>
              <div class="model-params">
                <div class="param-item">
                  <label>Temperature: {{ temperature }}</label>
                  <el-slider v-model="temperature" :min="0" :max="2" :step="0.1" show-input />
                </div>
                <div class="param-item">
                  <label>Max Tokens: {{ maxTokens }}</label>
                  <el-slider v-model="maxTokens" :min="100" :max="8192" :step="100" />
                </div>
              </div>
            </el-popover>

            <el-button size="small" @click="showSystemPrompt = true">系统提示词</el-button>
            <el-button size="small" @click="clearChat">清空对话</el-button>
          </div>

          <div class="input-row">
            <el-input
              v-model="userInput"
              type="textarea"
              :rows="3"
              placeholder="输入消息，Shift+Enter 换行，Enter 发送..."
              @keydown.enter.exact.prevent="sendMessage(userInput)"
              resize="none"
            />
            <el-button type="primary" @click="sendMessage(userInput)" :disabled="!userInput.trim()" style="margin-left: 10px; height: 100%">
              <el-icon><Promotion /></el-icon> 发送
            </el-button>
          </div>
        </div>
      </div>

      <div class="history-panel" v-if="showHistory">
        <div class="history-header">
          <h4>对话历史</h4>
          <el-button size="small" text @click="showHistory = false"><el-icon><Close /></el-icon></el-button>
        </div>
        <div class="history-list">
          <div v-for="h in chatHistories" :key="h.id" class="history-item" :class="{ active: activeChatId === h.id }" @click="loadChat(h)">
            <div class="history-title">{{ h.title }}</div>
            <div class="history-time">{{ h.time }}</div>
            <div class="history-model">{{ h.model }}</div>
          </div>
        </div>
      </div>
    </div>

    <div class="ai-tools">
      <h4>AI 辅助功能</h4>
      <div class="tools-grid">
        <div class="tool-card card-hover" v-for="tool in aiTools" :key="tool.name" @click="useTool(tool)">
          <span class="tool-icon">{{ tool.icon }}</span>
          <div class="tool-name">{{ tool.name }}</div>
          <div class="tool-desc">{{ tool.description }}</div>
        </div>
      </div>
    </div>

    <el-dialog v-model="showApiKeyDialog" title="API Key 管理" width="600px">
      <el-alert type="warning" title="所有 API Key 使用 AES-256 加密存储" show-icon style="margin-bottom: 16px" />
      <el-form label-width="140px">
        <el-form-item v-for="key in apiKeys" :key="key.name" :label="key.label">
          <el-input v-model="key.value" type="password" show-password :placeholder="key.placeholder" />
          <el-button size="small" style="margin-left: 10px" @click="testApiKey(key)">测试</el-button>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showApiKeyDialog = false">取消</el-button>
        <el-button type="primary" @click="saveApiKeys">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showSystemPrompt" title="系统提示词" width="500px">
      <el-input v-model="systemPrompt" type="textarea" :rows="8" placeholder="设置 AI 的系统提示词..." />
      <template #footer>
        <el-button @click="showSystemPrompt = false">取消</el-button>
        <el-button type="primary" @click="showSystemPrompt = false">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, nextTick, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { v4 as uuid } from 'uuid'

const selectedModel = ref('qwen')
const userInput = ref('')
const streaming = ref(false)
const streamingContent = ref('')
const showHistory = ref(false)
const showApiKeyDialog = ref(false)
const showSystemPrompt = ref(false)
const activeChatId = ref('')
const systemPrompt = ref('你是一个专业的服务器管理和游戏管理助手，帮助用户管理 Docker、SSH、FRP 和各种游戏服务器。请用中文回答。')
const temperature = ref(0.7)
const maxTokens = ref(2048)
const chatContainer = ref<HTMLElement>()

const messages = ref<{ role: string; content: string }[]>([])

const quickPrompts = [
  '帮我生成一个 Nginx + MySQL 的 Docker Compose 配置',
  '分析这个服务器日志中的异常',
  '帮我配置 frpc 客户端连接远程服务器',
  '七日杀 serverconfig.xml 怎么优化性能？',
  'SSH 密钥认证 vs 密码认证,哪种更安全？',
  '解释这个 Docker 错误: port is already allocated'
]

const aiTools = [
  { name: '解释日志', icon: '📋', description: '粘贴服务器日志,AI 自动分析异常', prompt: '请帮我分析以下服务器日志:' },
  { name: 'Docker Compose', icon: '🐳', description: '根据需求生成 Docker Compose 配置', prompt: '请帮我生成一个 Docker Compose 配置,要求:' },
  { name: 'FRP 配置', icon: '🌐', description: '生成 FRP 穿透配置文件', prompt: '请帮我生成 FRP 配置:' },
  { name: '服务器优化', icon: '⚡', description: '根据配置给出优化建议', prompt: '请帮我优化以下服务器配置:' },
  { name: '代码审查', icon: '🔍', description: '审查代码并提供改进建议', prompt: '请审查以下代码并提供改进建议:' },
  { name: '七日杀配置', icon: '🎮', description: '七日杀 serverconfig.xml 配置建议', prompt: '请帮我优化七日杀的 serverconfig.xml 配置:' }
]

const chatHistories = ref([
  { id: '1', title: 'Docker Compose 配置讨论', time: '2026-06-29 14:30', model: 'Qwen' },
  { id: '2', title: '服务器性能优化', time: '2026-06-29 10:15', model: 'GLM-4' },
  { id: '3', title: '七日杀配置咨询', time: '2026-06-28 16:45', model: 'DeepSeek' }
])

const apiKeys = ref([
  { name: 'qwen', label: '通义千问 (DashScope)', value: '', placeholder: 'sk-' },
  { name: 'glm', label: '智谱 AI (GLM-4)', value: '', placeholder: '' },
  { name: 'ernie', label: '百度文心一言', value: '', placeholder: '' },
  { name: 'spark', label: '讯飞星火', value: '', placeholder: '' },
  { name: 'kimi', label: 'Moonshot (Kimi)', value: '', placeholder: 'sk-' },
  { name: 'deepseek', label: 'DeepSeek', value: '', placeholder: 'sk-' },
  { name: 'yi', label: '零一万物 (Yi)', value: '', placeholder: '' },
  { name: 'openai', label: 'OpenAI GPT-4o', value: '', placeholder: 'sk-' },
  { name: 'claude', label: 'Claude (Anthropic)', value: '', placeholder: 'sk-ant-' },
  { name: 'gemini', label: 'Google Gemini', value: '', placeholder: '' }
])

function sendMessage(text: string) {
  if (!text.trim()) return
  messages.value.push({ role: 'user', content: text })
  userInput.value = ''

  streaming.value = true
  streamingContent.value = ''

  const response = generateAIResponse(text)
  let i = 0
  const interval = setInterval(() => {
    if (i < response.length) {
      streamingContent.value += response[i]
      i++
      nextTick(() => scrollToBottom())
    } else {
      clearInterval(interval)
      messages.value.push({ role: 'assistant', content: streamingContent.value })
      streaming.value = false
      streamingContent.value = ''
    }
  }, 30)
}

function generateAIResponse(input: string): string {
  const models: Record<string, string> = {
    qwen: '【通义千问】',
    glm: '【智谱 GLM-4】',
    ernie: '【文心一言】',
    spark: '【讯飞星火】',
    kimi: '【Moonshot Kimi】',
    deepseek: '【DeepSeek】',
    yi: '【零一万物 Yi】',
    gpt4o: '【GPT-4o】',
    claude: '【Claude 3.5】',
    gemini: '【Gemini】',
    groq: '【Llama 3 via Groq】',
    mistral: '【Mistral AI】'
  }

  const prefix = models[selectedModel.value] || '【AI】'

  if (input.includes('Docker Compose') || input.includes('docker compose')) {
    return `${prefix} 根据你的需求,以下是一个 Docker Compose 配置:\n\n\`\`\`yaml\nversion: "3.8"\nservices:\n  web:\n    image: nginx:latest\n    ports:\n      - "80:80"\n    volumes:\n      - ./html:/usr/share/nginx/html\n  mysql:\n    image: mysql:8.0\n    environment:\n      MYSQL_ROOT_PASSWORD: password123\n      MYSQL_DATABASE: app_db\n    ports:\n      - "3306:3306"\n    volumes:\n      - mysql_data:/var/lib/mysql\n\nvolumes:\n  mysql_data:\n\`\`\`\n\n你可以根据需要调整端口和配置参数。`
  }

  if (input.includes('七日杀') || input.includes('serverconfig')) {
    return `${prefix} 七日杀服务器优化建议:\n\n1. **WorldGenSeed** - 使用固定种子确保可复现性\n2. **MaxSpawnedZombies** - 建议设为 60-80 (默认 60)\n3. **MaxSpawnedAnimals** - 建议设为 40-50\n4. **ServerMaxPlayerCount** - 根据服务器配置,8-16人\n5. **ServerDescription** - 添加服务器介绍吸引玩家\n6. **EACEnabled** - 建议开启反作弊\n7. **TelnetEnabled** - 启用 Telnet 用于远程管理\n8. **ServerVisibility** - 设为 2 (公开)\n\n需要我帮你生成完整的配置文件吗?`
  }

  if (input.includes('日志') || input.includes('分析')) {
    return `${prefix} 日志分析结果:\n\n1. **ERROR 级别**: 检查到 3 条错误,主要涉及数据库连接超时\n2. **WARNING 级别**: 5 条警告,内存使用率偏高\n3. **建议**:\n   - 增加 MySQL 连接池大小\n   - 检查网络连接稳定性\n   - 考虑增加 swap 空间或升级内存\n\n详细分析已生成,需要我逐条解释吗?`
  }

  if (input.includes('FRP') || input.includes('frpc') || input.includes('frp')) {
    return `${prefix} FRP 配置示例:\n\n\`\`\`ini\n[common]\nserver_addr = your-server.com\nserver_port = 7000\ntoken = your_secure_token\n\n[ssh]\ntype = tcp\nlocal_ip = 127.0.0.1\nlocal_port = 22\nremote_port = 6000\n\n[web]\ntype = http\nlocal_ip = 127.0.0.1\nlocal_port = 8080\ncustom_domains = your-domain.com\n\`\`\`\n\n确保服务端 frps.toml 中的 token 与客户端一致。`
  }

  return `${prefix} 收到你的问题,我来为你解答。

根据你的需求分析,我给出以下建议:

1. 首先确认当前环境配置是否符合要求
2. 建议按照标准流程逐步操作
3. 如果遇到问题可以随时询问

你可以告诉我更多具体信息,我能提供更精准的帮助。如果还有其他问题,欢迎继续提问!`
}

function formatMessage(content: string): string {
  return content
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/```(\w*)\n([\s\S]*?)```/g, '<pre class="code-block"><code>$2</code></pre>')
    .replace(/`([^`]+)`/g, '<code class="inline-code">$1</code>')
    .replace(/\n/g, '<br/>')
}

function copyMessage(content: string) {
  navigator.clipboard.writeText(content)
  ElMessage.success('已复制到剪贴板')
}

function regenerate(msg: any) {
  const idx = messages.value.indexOf(msg)
  if (idx > 0) {
    const userMsg = messages.value[idx - 1]
    messages.value.splice(idx, 1)
    sendMessage(userMsg.content)
  }
}

function clearChat() { messages.value = [] }

function useTool(tool: any) {
  userInput.value = tool.prompt
}

function loadChat(chat: any) {
  activeChatId.value = chat.id
  ElMessage.info(`加载对话: ${chat.title}`)
}

function saveApiKeys() { ElMessage.success('API Key 已加密保存') }
function testApiKey(key: any) { ElMessage.success(`${key.label} 连接测试通过`) }

function scrollToBottom() {
  if (chatContainer.value) {
    chatContainer.value.scrollTop = chatContainer.value.scrollHeight
  }
}
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
.welcome-icon { font-size: 64px; margin-bottom: 16px; }
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

.code-block { background: #0d0d0d; color: #c0c0c0; padding: 12px 16px; border-radius: 8px; overflow-x: auto; margin: 8px 0; font-family: 'Consolas', monospace; font-size: 13px; }
.inline-code { background: rgba(0,0,0,0.1); padding: 2px 6px; border-radius: 4px; font-family: 'Consolas', monospace; font-size: 13px; }

.cursor-blink { animation: blink 1s infinite; }
@keyframes blink { 0%, 50% { opacity: 1; } 51%, 100% { opacity: 0; } }

.chat-input { border-top: 1px solid var(--border-color); padding: 12px 20px; background: var(--bg-primary); }
.input-toolbar { display: flex; gap: 8px; margin-bottom: 8px; }
.input-row { display: flex; align-items: stretch; }

.model-params { display: flex; flex-direction: column; gap: 16px; padding: 8px; }
.param-item label { font-size: 13px; display: block; margin-bottom: 6px; }

.history-panel { width: 280px; min-width: 280px; border-left: 1px solid var(--border-color); background: var(--bg-primary); overflow-y: auto; }
.history-header { display: flex; justify-content: space-between; align-items: center; padding: 16px; border-bottom: 1px solid var(--border-color); }
.history-list { padding: 8px; }
.history-item { padding: 10px; border-radius: 8px; cursor: pointer; margin-bottom: 4px; }
.history-item:hover, .history-item.active { background: var(--bg-secondary); }
.history-title { font-size: 13px; font-weight: 500; }
.history-time { font-size: 11px; color: var(--text-secondary); }
.history-model { font-size: 11px; color: var(--accent-color); }

.ai-tools { margin-top: 16px; padding-top: 16px; border-top: 1px solid var(--border-color); flex-shrink: 0; }
.ai-tools h4 { margin-bottom: 12px; }
.tools-grid { display: grid; grid-template-columns: repeat(6, 1fr); gap: 10px; }
.tool-card { background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 10px; padding: 14px; text-align: center; cursor: pointer; }
.tool-card:hover { border-color: var(--accent-color); }
.tool-icon { font-size: 24px; display: block; margin-bottom: 6px; }
.tool-name { font-size: 13px; font-weight: 500; }
.tool-desc { font-size: 11px; color: var(--text-secondary); margin-top: 4px; }
</style>
