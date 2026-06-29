import { defineStore } from 'pinia'

interface SettingsState {
  theme: 'light' | 'dark' | 'auto'
  language: string
  sidebarCollapsed: boolean
  sidebarWidth: number
  masterPassword: string
  autoStart: boolean
  notifications: boolean
}

export const useSettingsStore = defineStore('settings', {
  state: (): SettingsState => ({
    theme: 'dark',
    language: 'zh-CN',
    sidebarCollapsed: false,
    sidebarWidth: 220,
    masterPassword: '',
    autoStart: false,
    notifications: true
  }),
  getters: {
    isDark: (state) => {
      if (state.theme === 'auto') {
        return window.matchMedia('(prefers-color-scheme: dark)').matches
      }
      return state.theme === 'dark'
    }
  },
  actions: {
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed
    },
    setTheme(theme: 'light' | 'dark' | 'auto') {
      this.theme = theme
      this.applyTheme()
      this.saveSettings()
    },
    applyTheme() {
      const root = document.documentElement
      if (this.isDark) {
        root.classList.add('dark')
        root.style.colorScheme = 'dark'
      } else {
        root.classList.remove('dark')
        root.style.colorScheme = 'light'
      }
    },
    loadSettings() {
      try {
        const saved = localStorage.getItem('omnipanel-settings')
        if (saved) {
          const data = JSON.parse(saved)
          Object.assign(this, data)
        }
      } catch {}
      this.applyTheme()
    },
    saveSettings() {
      const data: Partial<SettingsState> = {
        theme: this.theme,
        language: this.language,
        sidebarCollapsed: this.sidebarCollapsed,
        autoStart: this.autoStart,
        notifications: this.notifications
      }
      localStorage.setItem('omnipanel-settings', JSON.stringify(data))
    }
  }
})
