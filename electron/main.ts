import { app, BrowserWindow, ipcMain, shell } from 'electron'
import { join } from 'path'
import { exec } from 'child_process'
import * as os from 'os'

let mainWindow: BrowserWindow | null = null

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1400,
    height: 900,
    minWidth: 1000,
    minHeight: 700,
    title: 'OmniPanel - 全能面板',
    icon: join(__dirname, '../build/icon.png'),
    webPreferences: {
      preload: join(__dirname, 'preload.js'),
      nodeIntegration: false,
      contextIsolation: true,
      webviewTag: true
    },
    frame: true,
    titleBarStyle: 'default',
    backgroundColor: '#ffffff'
  })

  if (process.env.VITE_DEV_SERVER_URL) {
    mainWindow.loadURL(process.env.VITE_DEV_SERVER_URL)
  } else {
    mainWindow.loadFile(join(__dirname, '../renderer/index.html'))
  }

  mainWindow.webContents.setWindowOpenHandler(({ url }) => {
    if (url.startsWith('http')) {
      shell.openExternal(url)
    }
    return { action: 'deny' }
  })
}

ipcMain.handle('get-system-info', async () => {
  const cpus = os.cpus()
  const totalMem = os.totalmem()
  const freeMem = os.freemem()
  return {
    platform: os.platform(),
    arch: os.arch(),
    hostname: os.hostname(),
    release: os.release(),
    uptime: os.uptime(),
    cpuModel: cpus[0]?.model || 'Unknown',
    cpuCores: cpus.length,
    cpuSpeed: cpus[0]?.speed || 0,
    totalMemory: totalMem,
    freeMemory: freeMem,
    usedMemory: totalMem - freeMem,
    homeDir: os.homedir(),
    tempDir: os.tmpdir()
  }
})

ipcMain.handle('get-process-list', async () => {
  return new Promise((resolve) => {
    const cmd = process.platform === 'win32'
      ? 'tasklist /FO CSV /NH'
      : 'ps aux --no-headers'
    exec(cmd, (error, stdout) => {
      if (error) {
        resolve([])
        return
      }
      const lines = stdout.trim().split('\n').slice(0, 200)
      const processes = lines.map(line => {
        const parts = line.trim().split(/\s+/)
        return {
          pid: parseInt(parts[1]) || 0,
          name: parts[0] || parts[parts.length - 1] || 'unknown',
          cpu: parts[2] || '0',
          mem: parts[3] || '0'
        }
      })
      resolve(processes)
    })
  })
})

ipcMain.handle('kill-process', async (_, pid: number) => {
  try {
    process.kill(pid, 'SIGTERM')
    return true
  } catch {
    return false
  }
})

ipcMain.handle('execute-command', async (_, command: string) => {
  return new Promise((resolve) => {
    exec(command, { timeout: 30000 }, (error, stdout, stderr) => {
      resolve(error ? stderr : stdout)
    })
  })
})

ipcMain.handle('open-external', async (_, url: string) => {
  await shell.openExternal(url)
})

ipcMain.handle('get-platform', () => process.platform)

app.whenReady().then(() => {
  createWindow()
  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow()
  })
})

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit()
})
