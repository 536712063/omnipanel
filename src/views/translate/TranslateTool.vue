<template>
  <div class="translate-tool">
    <div class="page-header">
      <h2>万能汉化工具</h2>
    </div>

    <el-tabs v-model="activeTab">
      <el-tab-pane label="文本翻译" name="text">
        <div class="translate-panel">
          <div class="translate-header">
            <el-select v-model="sourceLang" size="small" style="width: 150px">
              <el-option label="自动检测" value="auto" />
              <el-option label="英语" value="en" />
              <el-option label="日语" value="ja" />
              <el-option label="韩语" value="ko" />
              <el-option label="法语" value="fr" />
              <el-option label="德语" value="de" />
              <el-option label="俄语" value="ru" />
              <el-option label="西班牙语" value="es" />
              <el-option label="葡萄牙语" value="pt" />
              <el-option label="意大利语" value="it" />
            </el-select>
            <el-button @click="swapLang" circle size="small">
              <el-icon><Sort /></el-icon>
            </el-button>
            <el-select v-model="targetLang" size="small" style="width: 150px">
              <el-option label="简体中文" value="zh-CN" />
              <el-option label="繁体中文" value="zh-TW" />
              <el-option label="英语" value="en" />
              <el-option label="日语" value="ja" />
              <el-option label="韩语" value="ko" />
            </el-select>
            <el-select v-model="engine" size="small" style="width: 160px; margin-left: auto">
              <el-option label="百度翻译" value="baidu" />
              <el-option label="有道翻译" value="youdao" />
              <el-option label="DeepL" value="deepl" />
              <el-option label="Google 翻译" value="google" />
            </el-select>
            <el-button type="primary" size="small" @click="translateText" :loading="translating">翻译</el-button>
          </div>

          <div class="translate-area">
            <div class="translate-input">
              <el-input
                v-model="inputText"
                type="textarea"
                :rows="12"
                placeholder="输入要翻译的文本..."
                resize="none"
              />
              <div class="char-count">{{ inputText.length }} / 5000</div>
            </div>
            <div class="translate-output">
              <div class="output-text" v-if="outputText">{{ outputText }}</div>
              <div class="output-placeholder" v-else>翻译结果将显示在这里...</div>
              <div class="output-actions" v-if="outputText">
                <el-button size="small" text @click="copyOutput">复制</el-button>
                <el-button size="small" text @click="addToMemory">添加到记忆库</el-button>
                <el-button size="small" text @click="ttsOutput">朗读</el-button>
              </div>
            </div>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="文件汉化" name="file">
        <div class="file-translate">
          <div class="file-upload-area">
            <el-upload
              drag
              :auto-upload="false"
              :on-change="handleFileSelect"
              accept=".txt,.json,.xml,.properties,.yaml,.yml,.ini,.cfg"
            >
              <el-icon :size="48"><UploadFilled /></el-icon>
              <div>将文件拖到此处,或点击上传</div>
              <div class="upload-hint">支持 .txt / .json / .xml / .properties / .yaml / .ini / .cfg</div>
            </el-upload>
          </div>

          <div class="file-list" v-if="fileQueue.length">
            <h4>待处理文件</h4>
            <div class="file-item" v-for="(file, i) in fileQueue" :key="i">
              <div class="file-info">
                <el-icon><Document /></el-icon>
                <span class="file-name">{{ file.name }}</span>
                <span class="file-size">{{ file.size }}</span>
                <el-tag :type="file.status === 'done' ? 'success' : file.status === 'translating' ? 'warning' : 'info'" size="small">
                  {{ file.status === 'done' ? '完成' : file.status === 'translating' ? '翻译中...' : '等待中' }}
                </el-tag>
              </div>
              <div class="file-actions">
                <el-button size="small" @click="downloadTranslatedFile(file)">下载</el-button>
              </div>
            </div>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="网页翻译" name="web">
        <div class="web-translate">
          <div class="web-input">
            <el-input v-model="webUrl" placeholder="输入网页 URL..." style="width: 400px">
              <template #append>
                <el-button @click="translateWeb" :loading="webTranslating">翻译</el-button>
              </template>
            </el-input>
          </div>

          <div class="web-preview" v-if="webContent">
            <div class="web-translated">
              <h4>翻译结果</h4>
              <div v-html="webContent" class="web-content"></div>
            </div>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="翻译记忆库" name="memory">
        <div class="translation-memory">
          <div class="memory-toolbar">
            <el-input v-model="memorySearch" placeholder="搜索记忆库..." size="small" clearable style="width: 300px" />
            <el-button size="small" @click="exportMemory">导出记忆库</el-button>
            <el-button size="small" @click="importMemory">导入记忆库</el-button>
          </div>

          <el-table :data="filteredMemory" size="small">
            <el-table-column prop="source" label="原文" min-width="200" />
            <el-table-column prop="target" label="译文" min-width="200" />
            <el-table-column prop="lang" label="语言对" width="100" />
            <el-table-column prop="date" label="添加时间" width="160" />
            <el-table-column label="操作" width="80">
              <template #default="{ row }">
                <el-button size="small" type="danger" text @click="deleteMemory(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>

      <el-tab-pane label="OCR 翻译" name="ocr">
        <div class="ocr-translate">
          <div class="ocr-upload">
            <el-upload drag :auto-upload="false" :on-change="handleOCRImage" accept="image/*">
              <el-icon :size="48"><PictureFilled /></el-icon>
              <div>上传截图或图片进行 OCR 识别与翻译</div>
              <div class="upload-hint">支持 PNG / JPG / BMP</div>
            </el-upload>
          </div>

          <div class="ocr-result" v-if="ocrText">
            <h4>识别结果</h4>
            <div class="ocr-original">{{ ocrText }}</div>
            <h4>翻译结果</h4>
            <div class="ocr-translated">{{ ocrTranslated }}</div>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, nextTick } from 'vue'
import { ElMessage } from 'element-plus'

const activeTab = ref('text')
const sourceLang = ref('auto')
const targetLang = ref('zh-CN')
const engine = ref('baidu')
const inputText = ref('')
const outputText = ref('')
const translating = ref(false)
const webUrl = ref('')
const webContent = ref('')
const webTranslating = ref(false)
const memorySearch = ref('')
const ocrText = ref('')
const ocrTranslated = ref('')

const fileQueue = ref<any[]>([])

const translationMemory = ref([
  { source: 'Hello World', target: '你好世界', lang: 'en -> zh-CN', date: '2026-06-29' },
  { source: 'Server started successfully', target: '服务器启动成功', lang: 'en -> zh-CN', date: '2026-06-29' },
  { source: 'Configuration file not found', target: '配置文件未找到', lang: 'en -> zh-CN', date: '2026-06-28' },
  { source: 'Out of memory', target: '内存不足', lang: 'en -> zh-CN', date: '2026-06-27' },
  { source: 'Docker container running', target: 'Docker 容器运行中', lang: 'en -> zh-CN', date: '2026-06-26' }
])

const filteredMemory = computed(() =>
  translationMemory.value.filter(m =>
    m.source.toLowerCase().includes(memorySearch.value.toLowerCase()) ||
    m.target.includes(memorySearch.value)
  )
)

function swapLang() {
  if (sourceLang.value === 'auto') return
  [sourceLang.value, targetLang.value] = [targetLang.value, sourceLang.value];
  [inputText.value, outputText.value] = [outputText.value, inputText.value]
}

async function translateText() {
  if (!inputText.value.trim()) { ElMessage.warning('请输入要翻译的文本'); return }
  translating.value = true
  await new Promise(r => setTimeout(r, 1000))
  // Mock translation
  const mock: Record<string, string> = {
    'Hello World': '你好世界',
    'hello': '你好',
    'world': '世界'
  }
  outputText.value = mock[inputText.value.trim()] || `[${engine.value}] 翻译结果: ${inputText.value} (翻译为${targetLang.value})`
  translating.value = false
}

function copyOutput() {
  navigator.clipboard.writeText(outputText.value)
  ElMessage.success('已复制到剪贴板')
}

function addToMemory() {
  translationMemory.value.unshift({
    source: inputText.value.substring(0, 50),
    target: outputText.value.substring(0, 50),
    lang: `${sourceLang.value} -> ${targetLang.value}`,
    date: new Date().toISOString().split('T')[0]
  })
  ElMessage.success('已添加到翻译记忆库')
}

function ttsOutput() { ElMessage.info('语音朗读 (功能开发中)') }

function handleFileSelect(file: any) {
  fileQueue.value.push({
    name: file.name,
    size: formatSize(file.size || 0),
    status: 'pending',
    raw: file.raw
  })
}

function downloadTranslatedFile(file: any) { ElMessage.success(`下载 ${file.name} 的翻译结果`) }

async function translateWeb() {
  if (!webUrl.value) { ElMessage.warning('请输入网页 URL'); return }
  webTranslating.value = true
  await new Promise(r => setTimeout(r, 1500))
  webContent.value = `<div style="padding: 20px"><h3>${webUrl.value} - 汉化版本</h3><p>这是一个模拟的网页翻译结果...</p></div>`
  webTranslating.value = false
}

function handleOCRImage(file: any) {
  ocrText.value = 'This is a sample text extracted from the image.'
  ocrTranslated.value = '这是从图片中提取的示例文本。'
}

function deleteMemory(row: any) {
  translationMemory.value = translationMemory.value.filter(m => m !== row)
}

function exportMemory() { ElMessage.success('记忆库已导出') }
function importMemory() { ElMessage.success('记忆库已导入') }

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + 'B'
  if (bytes < 1048576) return (bytes / 1024).toFixed(1) + 'KB'
  return (bytes / 1048576).toFixed(1) + 'MB'
}
</script>

<style scoped>
.translate-tool { padding: 0; }
.page-header { margin-bottom: 16px; }
.page-header h2 { font-size: 22px; font-weight: 600; }

.translate-header { display: flex; align-items: center; gap: 10px; margin-bottom: 16px; }

.translate-area { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; min-height: 400px; }
.translate-input, .translate-output { background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 8px; padding: 16px; position: relative; }
.char-count { position: absolute; bottom: 8px; right: 12px; font-size: 12px; color: var(--text-secondary); }
.output-text { white-space: pre-wrap; line-height: 1.8; }
.output-placeholder { color: var(--text-secondary); display: flex; align-items: center; justify-content: center; height: 200px; }
.output-actions { margin-top: 12px; display: flex; gap: 8px; }

.file-upload-area { margin-bottom: 20px; }
.upload-hint { font-size: 12px; color: var(--text-secondary); margin-top: 8px; }

.file-item { display: flex; justify-content: space-between; align-items: center; padding: 10px 14px; background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 8px; margin-bottom: 8px; }
.file-info { display: flex; align-items: center; gap: 10px; }
.file-name { font-weight: 500; }
.file-size { color: var(--text-secondary); font-size: 12px; }

.web-translate { padding: 16px 0; }
.web-input { margin-bottom: 16px; }
.web-content { background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 8px; padding: 20px; }

.memory-toolbar { display: flex; gap: 10px; margin-bottom: 16px; }

.ocr-upload { margin-bottom: 20px; }
.ocr-result { background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 8px; padding: 20px; }
.ocr-original { padding: 12px; background: var(--bg-secondary); border-radius: 6px; margin-bottom: 16px; white-space: pre-wrap; }
.ocr-translated { padding: 12px; background: rgba(64,158,255,0.05); border-radius: 6px; }
.ocr-result h4 { margin-bottom: 8px; }
</style>
