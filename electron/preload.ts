import { contextBridge, ipcRenderer } from 'electron'

contextBridge.exposeInMainWorld('electronAPI', {
  getSystemInfo: () => ipcRenderer.invoke('get-system-info'),
  getProcessList: () => ipcRenderer.invoke('get-process-list'),
  killProcess: (pid: number) => ipcRenderer.invoke('kill-process', pid),
  executeCommand: (command: string) => ipcRenderer.invoke('execute-command', command),
  openExternal: (url: string) => ipcRenderer.invoke('open-external', url),
  platform: process.platform,
  onNotification: (callback: (data: any) => void) => {
    ipcRenderer.on('notification', (_, data) => callback(data))
  },
  minimize: () => ipcRenderer.invoke('window-minimize'),
  maximize: () => ipcRenderer.invoke('window-maximize'),
  close: () => ipcRenderer.invoke('window-close')
})
