import { ref } from 'vue'
import type { I18nExtractReq, I18nExtractResult, I18nItem } from '@/wails/runtime'
import * as api from '@/wails/runtime'

export function useI18nTool() {
  const result = ref<I18nExtractResult | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function extract(req: I18nExtractReq) {
    loading.value = true
    try { result.value = await api.i18nExtract(req) } catch (e: any) { error.value = e?.message } finally { loading.value = false }
  }

  async function generateLocale(items: I18nItem[], locale: string, outputPath: string) {
    await api.i18nGenerateLocaleFile(items, locale, outputPath)
  }

  async function preview(filePath: string, items: I18nItem[]) {
    return api.i18nPreviewTranslation(filePath, items)
  }

  async function apply(filePath: string, items: I18nItem[]) {
    await api.i18nApplyTranslationFile(filePath, items)
  }

  async function batchApply() {
    if (result.value) await api.i18nBatchApplyTranslation(result.value)
  }

  return { result, loading, error, extract, generateLocale, preview, apply, batchApply }
}
