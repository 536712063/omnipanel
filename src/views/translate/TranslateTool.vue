<template>
  <div class="translate-tool">
    <div class="page-header">
      <h2>{{ $t('translate.title') }}</h2>
      <div class="header-actions">
        <el-button type="primary" @click="showExtractDialog = true">
          <el-icon><Search /></el-icon> {{ $t('translate.extract') }}
        </el-button>
        <el-button v-if="result" @click="showGenerateDialog = true">
          <el-icon><Document /></el-icon> {{ $t('translate.generate') }}
        </el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab">
      <el-tab-pane :label="$t('translate.textTranslate')" name="text">
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
            </el-select>
            <el-button @click="swapLang" circle size="small">
              <el-icon><Sort /></el-icon>
            </el-button>
            <el-select v-model="targetLang" size="small" style="width: 150px">
              <el-option label="简体中文" value="zh-CN" />
              <el-option label="繁体中文" value="zh-TW" />
              <el-option label="英语" value="en" />
              <el-option label="日语" value="ja" />
            </el-select>
            <div style="margin-left: auto; display: flex; gap: 8px">
              <el-select v-model="engine" size="small" style="width: 150px">
                <el-option v-for="e in engines" :key="e.value" :label="$t(`translate.engines.${e.value}`)" :value="e.value" />
              </el-select>
              <el-button type="primary" size="small" @click="doTranslate" :loading="translating">{{ $t('translate.translate') }}</el-button>
            </div>
          </div>

          <div class="translate-area">
            <div class="translate-input">
              <el-input v-model="inputText" type="textarea" :rows="12" placeholder="输入要翻译的文本..." resize="none" />
              <div class="char-count">{{ inputText.length }} / 5000</div>
            </div>
            <div class="translate-output">
              <div class="output-text" v-if="outputText">{{ outputText }}</div>
              <div class="output-placeholder" v-else>{{ $t('translate.translate') }}</div>
              <div class="output-actions" v-if="outputText">
                <el-button size="small" text @click="copyOutput">{{ $t('common.copy') }}</el-button>
              </div>
            </div>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane :label="$t('translate.fileTranslate')" name="file">
        <div class="extract-panel">
          <div class="extract-summary" v-if="result">
            <el-descriptions :column="3" border size="small">
              <el-descriptions-item :label="$t('ai.errors.sendFailed')">{{ result.total_files }}</el-descriptions-item>
              <el-descriptions-item :label="$t('translate.title')">{{ result.total_items }}</el-descriptions-item>
            </el-descriptions>

            <div class="extracted-files" v-for="(items, file) in result.files" :key="file">
              <h4>{{ file }} ({{ items.length }} items)</h4>
              <el-table :data="items" size="small" max-height="300">
                <el-table-column prop="key" label="Key" width="200" />
                <el-table-column prop="original" :label="$t('translate.textTranslate')" min-width="200" />
                <el-table-column prop="translation" :label="$t('translate.textTranslate')" min-width="200">
                  <template #default="{ row }">
                    <el-input v-model="row.translation" size="small" :placeholder="row.original" />
                  </template>
                </el-table-column>
                <el-table-column prop="context" label="Context" width="120" />
              </el-table>
            </div>
          </div>

          <div v-else class="no-result">
            <p>{{ $t('translate.extract') }}</p>
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane :label="$t('translate.memory')" name="memory">
        <div class="memory-panel">
          <div class="memory-toolbar">
            <el-input v-model="memorySearch" :placeholder="$t('common.search')" size="small" clearable style="width: 260px" />
            <el-button size="small" @click="exportMemory">{{ $t('translate.exportMemory') }}</el-button>
            <el-button size="small" @click="importMemory">{{ $t('translate.importMemory') }}</el-button>
          </div>
          <el-table :data="filteredMemory" size="small">
            <el-table-column prop="source" :label="$t('translate.textTranslate')" min-width="200" />
            <el-table-column prop="target" :label="$t('translate.textTranslate')" min-width="200" />
            <el-table-column prop="lang" label="语言对" width="120" />
            <el-table-column prop="date" :label="$t('git.log')" width="140" />
            <el-table-column :label="$t('common.edit')" width="80">
              <template #default="{ row }">
                <el-button size="small" type="danger" text @click="deleteMemory(row)">{{ $t('common.delete') }}</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="showExtractDialog" :title="$t('translate.extract')" width="500px">
      <el-form :model="extractReq" label-width="100px">
        <el-form-item :label="$t('git.localPath')">
          <el-input v-model="extractReq.source_dir" placeholder="/workspace/src" />
        </el-form-item>
        <el-form-item label="Extensions">
          <el-input v-model="extractReq.extensionsStr" placeholder=".xml,.json,.txt (comma separated)" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showExtractDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="doExtract" :loading="loading">{{ $t('translate.extract') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showGenerateDialog" :title="$t('translate.generate')" width="400px">
      <el-form label-width="80px">
        <el-form-item label="Locale">
          <el-input v-model="genLocale" placeholder="zh-CN" />
        </el-form-item>
        <el-form-item :label="$t('git.localPath')">
          <el-input v-model="genOutputPath" placeholder="/tmp/locale.json" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showGenerateDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="doGenerate">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18nTool } from '@/composables/useI18nTool'
import type { I18nExtractReq } from '@/wails/runtime'

const { result, loading, error, extract, generateLocale, preview, apply, batchApply } = useI18nTool()

const activeTab = ref('text')
const sourceLang = ref('auto')
const targetLang = ref('zh-CN')
const engine = ref('baidu')
const inputText = ref('')
const outputText = ref('')
const translating = ref(false)
const memorySearch = ref('')
const showExtractDialog = ref(false)
const showGenerateDialog = ref(false)
const genLocale = ref('zh-CN')
const genOutputPath = ref('')
const extractReq = reactive<I18nExtractReq & { extensionsStr: string }>({ source_dir: '', extensionsStr: '.xml,.json,.txt' })

const engines = [
  { value: 'baidu' },
  { value: 'youdao' },
  { value: 'deepl' },
  { value: 'google' }
]

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

function swapLang() { if (sourceLang.value === 'auto') return; [sourceLang.value, targetLang.value] = [targetLang.value, sourceLang.value]; [inputText.value, outputText.value] = [outputText.value, inputText.value] }

async function doTranslate() {
  if (!inputText.value.trim()) { ElMessage.warning('请输入要翻译的文本'); return }
  translating.value = true
  await new Promise(r => setTimeout(r, 800))
  outputText.value = `[${engine.value}] ${inputText.value} (翻译为${targetLang.value})`
  translationMemory.value.unshift({ source: inputText.value.substring(0, 50), target: outputText.value.substring(0, 50), lang: `${sourceLang.value} -> ${targetLang.value}`, date: new Date().toISOString().split('T')[0] })
  translating.value = false
}

function copyOutput() { navigator.clipboard.writeText(outputText.value); ElMessage.success('已复制') }

async function doExtract() {
  if (!extractReq.source_dir) { ElMessage.warning('请输入源码目录'); return }
  const extensions = extractReq.extensionsStr.split(',').map(s => s.trim()).filter(Boolean)
  await extract({ source_dir: extractReq.source_dir, extensions: extensions.length ? extensions : undefined })
  showExtractDialog.value = false
  activeTab.value = 'file'
}

async function doGenerate() {
  if (!result.value) return
  const allItems = Object.values(result.value.files).flat()
  await generateLocale(allItems, genLocale.value, genOutputPath.value || `/tmp/locale_${genLocale.value}.json`)
  showGenerateDialog.value = false
  ElMessage.success('语言文件已生成')
}

function deleteMemory(row: any) { translationMemory.value = translationMemory.value.filter(m => m !== row) }
function exportMemory() { ElMessage.success('记忆库已导出') }
function importMemory() { ElMessage.success('记忆库已导入') }
</script>

<style scoped>
.translate-tool { padding: 0; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-header h2 { font-size: 22px; font-weight: 600; }
.header-actions { display: flex; gap: 10px; }

.translate-header { display: flex; align-items: center; gap: 10px; margin-bottom: 16px; }
.translate-area { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; min-height: 300px; }
.translate-input, .translate-output { background: var(--bg-primary); border: 1px solid var(--border-color); border-radius: 8px; padding: 16px; position: relative; }
.char-count { position: absolute; bottom: 8px; right: 12px; font-size: 12px; color: var(--text-secondary); }
.output-text { white-space: pre-wrap; line-height: 1.8; }
.output-placeholder { color: var(--text-secondary); display: flex; align-items: center; justify-content: center; height: 200px; }
.output-actions { margin-top: 12px; display: flex; gap: 8px; }

.extract-panel { padding: 12px 0; }
.extracted-files { margin-top: 16px; }
.extracted-files h4 { margin-bottom: 8px; font-size: 14px; }
.no-result { text-align: center; padding: 60px; color: var(--text-secondary); }

.memory-panel { padding: 12px 0; }
.memory-toolbar { display: flex; gap: 10px; margin-bottom: 16px; }
</style>
