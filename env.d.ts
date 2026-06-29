/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

interface Window {
  electronAPI: {
    getSystemInfo: () => Promise<any>
    getProcessList: () => Promise<any[]>
    killProcess: (pid: number) => Promise<boolean>
    executeCommand: (command: string) => Promise<string>
    openExternal: (url: string) => Promise<void>
    platform: string
    onNotification: (callback: (data: any) => void) => void
    minimize: () => void
    maximize: () => void
    close: () => void
  }
}
