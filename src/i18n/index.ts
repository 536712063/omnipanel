import { createI18n } from 'vue-i18n'
import zhCN from './locales/zh-CN'

const i18n = createI18n({
  legacy: false,
  locale: 'zhCN',
  fallbackLocale: 'zhCN',
  messages: {
    zhCN
  }
})

export { i18n }
